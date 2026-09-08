// Package materializer executes a MaterializationPlan with filesystem and
// process side effects. It is the FIRST and ONLY package allowed those
// effects: resolver, composer and planner stay pure and are only read for
// types. The materializer never re-resolves architecture; the plan is the
// decision.
//
// Pipeline: validate output dir (empty-or-new) → staging temp dir → per
// component fetch source → verify pin → copy+prune with path-safety
// checks → collision check → write manifest, provenance, project map and
// agent-context files into staging → post-materialize checks → atomic move
// into place. Staging is removed on any failure; the output directory is
// only touched by the final rename.
package materializer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/handoff"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// Request is the full materialization input.
type Request struct {
	Plan         planner.MaterializationPlan
	Catalog      catalog.Catalog
	OutputDir    string
	IntentJSON   []byte
	DecisionJSON []byte
	CoreVersion  string
	// CommandTimeout bounds curated adapter commands; <=0 selects the default.
	CommandTimeout time.Duration
}

// Result describes the committed project.
type Result struct {
	ProjectDir string
	Manifest   project.Manifest
}

// Materialize executes the request pipeline and returns the committed
// manifest. Any failure before the final rename leaves the output
// directory untouched.
func Materialize(req Request) (Result, error) {
	if err := preValidate(req); err != nil {
		return Result{}, err
	}
	if err := ensureEmptyOrNew(req.OutputDir); err != nil {
		return Result{}, err
	}
	parent := filepath.Dir(req.OutputDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Result{}, domain.Filesystem(fmt.Sprintf("create output parent: %v", err))
	}
	staging, err := os.MkdirTemp(parent, ".staging-*")
	if err != nil {
		return Result{}, domain.Filesystem(fmt.Sprintf("create staging dir: %v", err))
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(staging)
		}
	}()
	fetchRoot, err := os.MkdirTemp("", "eng-fetch-*")
	if err != nil {
		return Result{}, domain.Filesystem(fmt.Sprintf("create fetch dir: %v", err))
	}
	defer os.RemoveAll(fetchRoot)

	ordered := orderedComponents(req.Plan)
	idx := catalog.NewIndex(req.Catalog)
	ctx := context.Background()
	for _, c := range ordered {
		bp, ok := idx.Boilerplate(c.Boilerplate)
		if !ok {
			return Result{}, domain.Catalog(fmt.Sprintf("boilerplate %q is not in the catalog", c.Boilerplate))
		}
		if bp.Pin != c.Pin {
			return Result{}, domain.Materialization(fmt.Sprintf(
				"catalog drift: plan pins %s@%s but the catalog pins %s", c.Boilerplate, c.Pin, bp.Pin))
		}
		spec := bp.EffectiveSpec()
		srcDir, err := FetchSource(ctx, bp, c.Pin, fetchRoot)
		if err != nil {
			return Result{}, err
		}
		dest := projectSubdir(staging, c.Destination)
		if _, err := os.Stat(dest); err == nil {
			return Result{}, domain.Materialization(fmt.Sprintf("destination collision: %q already staged", c.Destination))
		}
		if err := CopyTree(srcDir, dest, spec.PrunePaths); err != nil {
			return Result{}, err
		}
		if len(spec.Setup) > 0 {
			if _, err := RunCommands(ctx, spec.Setup, dest, req.CommandTimeout); err != nil {
				return Result{}, err
			}
		}
		if len(spec.Checks) > 0 {
			if _, err := RunCommands(ctx, spec.Checks, dest, req.CommandTimeout); err != nil {
				return Result{}, err
			}
		}
	}
	files, err := ListFiles(staging)
	if err != nil {
		return Result{}, err
	}
	manifest := project.BuildManifest(req.Plan, req.Catalog.CatalogVersion, files)
	manifestBytes, err := manifest.Marshal()
	if err != nil {
		return Result{}, err
	}
	pins := map[string]string{}
	for _, c := range ordered {
		pins[c.Boilerplate] = c.Pin
	}
	provenance := project.BuildProvenance(
		req.CoreVersion, req.Catalog.CatalogVersion,
		project.IntentFingerprintOf(req.DecisionJSON), req.Plan.Fingerprint,
		pins, time.Now().UTC())
	provenanceBytes, err := provenance.Marshal()
	if err != nil {
		return Result{}, err
	}
	projectMapBytes, err := project.BuildProjectMap(req.Plan).Marshal()
	if err != nil {
		return Result{}, err
	}
	agentFiles, err := handoff.Render(handoff.Input{
		Plan:         req.Plan,
		Catalog:      req.Catalog,
		IntentJSON:   req.IntentJSON,
		DecisionJSON: req.DecisionJSON,
	})
	if err != nil {
		return Result{}, err
	}
	state := []handoff.File{
		{Path: project.ManifestRelPath, Data: manifestBytes, Overwrite: true},
		{Path: project.ProvenanceRelPath, Data: provenanceBytes, Overwrite: true},
		{Path: project.ProjectMapRelPath, Data: projectMapBytes, Overwrite: true},
	}
	if err := writeStagingFiles(staging, append(state, agentFiles...)); err != nil {
		return Result{}, err
	}
	// Re-list: manifest and agent-context files are part of the record.
	files, err = ListFiles(staging)
	if err != nil {
		return Result{}, err
	}
	manifest = project.BuildManifest(req.Plan, req.Catalog.CatalogVersion, files)
	manifestBytes, err = manifest.Marshal()
	if err != nil {
		return Result{}, err
	}
	if err := writeStagingFiles(staging, []handoff.File{
		{Path: project.ManifestRelPath, Data: manifestBytes, Overwrite: true},
	}); err != nil {
		return Result{}, err
	}
	if err := Verify(staging, req.Plan, manifest); err != nil {
		return Result{}, err
	}
	if err := commitStaging(staging, req.OutputDir); err != nil {
		return Result{}, err
	}
	committed = true
	return Result{ProjectDir: req.OutputDir, Manifest: manifest}, nil
}

// preValidate runs every check that needs no side effects so malformed
// plans, unknown providers, drifted pins, unsafe destinations, colliding
// layouts and malicious adapter commands fail before staging exists.
func preValidate(req Request) error {
	plan := req.Plan
	if len(plan.Components) == 0 {
		return domain.Materialization("cannot materialize: plan has no components")
	}
	if plan.Fingerprint == "" {
		return domain.Materialization("cannot materialize: plan has no fingerprint")
	}
	idx := catalog.NewIndex(req.Catalog)
	dests := make([]string, 0, len(plan.Components))
	for _, c := range plan.Components {
		if err := ValidateDestination(c.Destination); err != nil {
			return err
		}
		dests = append(dests, c.Destination)
		bp, ok := idx.Boilerplate(c.Boilerplate)
		if !ok {
			return domain.Catalog(fmt.Sprintf("boilerplate %q is not in the catalog", c.Boilerplate))
		}
		if err := bp.Validate(); err != nil {
			return err
		}
		if bp.Pin != c.Pin {
			return domain.Materialization(fmt.Sprintf(
				"catalog drift: plan pins %s@%s but the catalog pins %s", c.Boilerplate, c.Pin, bp.Pin))
		}
		spec := bp.EffectiveSpec()
		for _, op := range spec.Operations {
			switch op {
			case "fetch", "copy", "prune", "template", "compose":
			default:
				return domain.Catalog(fmt.Sprintf("boilerplate %q declares unknown operation %q", bp.ID, op))
			}
		}
		if err := ValidateCommands(append(append([]domain.AdapterCommand{}, spec.Setup...), spec.Checks...)); err != nil {
			return err
		}
		if bp.Source.Kind() == "local" && !filepath.IsAbs(bp.Source.Path) {
			return domain.Materialization(fmt.Sprintf(
				"local source for %q must be an absolute path (got %q)", bp.ID, bp.Source.Path))
		}
	}
	if err := CheckCollisions(dests); err != nil {
		return err
	}
	if strings.TrimSpace(req.OutputDir) == "" {
		return domain.Filesystem("output directory must not be empty")
	}
	return nil
}

func orderedComponents(plan planner.MaterializationPlan) []planner.PlanComponent {
	out := append([]planner.PlanComponent{}, plan.Components...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Destination != out[j].Destination {
			return out[i].Destination < out[j].Destination
		}
		return out[i].Surface < out[j].Surface
	})
	return out
}

// writeStagingFiles is the single project-write path: every generated
// byte passes through here. Non-overwriting files (surface stubs) are
// skipped when the foundation already shipped them.
func writeStagingFiles(staging string, files []handoff.File) error {
	for _, f := range files {
		target := projectSubdir(staging, f.Path)
		if !insideDir(staging, target) {
			return domain.Filesystem(fmt.Sprintf("generated path %q escapes staging", f.Path))
		}
		if !f.Overwrite {
			if _, err := os.Stat(target); err == nil {
				continue
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return domain.Filesystem(fmt.Sprintf("create dir for %q: %v", f.Path, err))
		}
		if err := os.WriteFile(target, f.Data, 0o644); err != nil {
			return domain.Filesystem(fmt.Sprintf("write %q: %v", f.Path, err))
		}
	}
	return nil
}

// empty output dir is removed first so the rename is the commit point.
func commitStaging(staging, output string) error {
	if _, err := os.Stat(output); err == nil {
		if err := os.Remove(output); err != nil {
			return domain.Filesystem(fmt.Sprintf("clear empty output dir: %v", err))
		}
	}
	if err := os.Rename(staging, output); err != nil {
		return domain.Filesystem(fmt.Sprintf("commit staging to %q: %v", output, err))
	}
	return nil
}

// projectSubdir joins a project-relative slash destination under root.
func projectSubdir(root, dest string) string {
	return filepath.Join(root, filepath.FromSlash(dest))
}
