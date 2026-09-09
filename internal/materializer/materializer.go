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
		var srcDir string
		if spec.Generate != nil {
			// Generated foundation: acquire the pinned generator
			// factory when the adapter fetches one, execute the
			// curated command in the isolated sandbox and keep only
			// the declared output. The factory itself never flows
			// into the project staging.
			var err error
			srcDir, err = runGeneratedComponent(ctx, req, c, bp, spec.Generate, fetchRoot)
			if err != nil {
				return Result{}, err
			}
		} else {
			var err error
			srcDir, err = FetchSource(ctx, bp, c.Pin, fetchRoot)
			if err != nil {
				return Result{}, err
			}
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
	provComponents := project.ComponentsForPlan(req.Plan, req.Catalog)
	provenance := project.BuildProvenanceWithComponents(
		req.CoreVersion, req.Catalog.CatalogVersion,
		project.IntentFingerprintOf(req.DecisionJSON), req.Plan.Fingerprint,
		pins, provComponents, time.Now().UTC())
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

// runGeneratedComponent materializes one generated plan component: it
// creates the per-component sandbox, acquires the pinned generator
// factory when the adapter declares a fetch operation, resolves the
// runtime placeholder values and executes the curated generator. The
// returned directory is the validated generated output only.
func runGeneratedComponent(ctx context.Context, req Request, c planner.PlanComponent, bp domain.Boilerplate, gen *domain.GenerateSpec, fetchRoot string) (string, error) {
	sandbox, err := os.MkdirTemp(fetchRoot, "sandbox-*")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create generation sandbox: %v", err))
	}
	factoryDir := ""
	wantsFactory := false
	for _, op := range bp.EffectiveSpec().Operations {
		if op == "fetch" {
			wantsFactory = true
		}
	}
	if wantsFactory {
		factoryDir, err = FetchSource(ctx, bp, c.Pin, sandbox)
		if err != nil {
			return "", err
		}
	}
	values := GenerationValues{
		Name:    c.Materialization.Name,
		Project: req.Plan.Project,
		Surface: c.Surface,
		Profile: c.Materialization.Profile,
		Output:  filepath.Join(sandbox, "output"),
	}
	if values.Name == "" {
		values.Name, err = generateAppName(c.Destination)
		if err != nil {
			return "", err
		}
	}
	if values.Project == "" {
		values.Project = values.Name
	}
	outDir, err := RunGenerateWithValues(ctx, gen, values, factoryDir, sandbox, req.CommandTimeout)
	if err != nil {
		return "", decorateGenerateError(c, bp, err)
	}
	return outDir, nil
}

// decorateGenerateError attributes a generator failure to its surface,
// provider and stage without leaking secrets or temp paths.
func decorateGenerateError(c planner.PlanComponent, bp domain.Boilerplate, err error) error {
	var stageErr *generationStageError
	if e, ok := err.(*generationStageError); ok {
		stageErr = e
	}
	stage := "execution"
	if stageErr != nil {
		stage = stageErr.stage
		err = stageErr.err
	}
	return domain.Materialization(fmt.Sprintf(
		"generate %s (provider %s, stage %s): %v", c.Surface, bp.ID, stage, err))
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
			case "fetch", "copy", "prune", "template", "compose", "generate":
			default:
				return domain.Catalog(fmt.Sprintf("boilerplate %q declares unknown operation %q", bp.ID, op))
			}
		}
		if planStrategy := c.Materialization.Strategy; planStrategy != "" && planStrategy != spec.Strategy() {
			return domain.Materialization(fmt.Sprintf(
				"adapter drift: plan records %s@%s as %q but the catalog adapter is %q (fingerprint %s)",
				c.Boilerplate, c.Pin, planStrategy, spec.Strategy(), domain.AdapterFingerprint(spec)))
		}
		if fp := c.Materialization.AdapterFingerprint; fp != "" && fp != domain.AdapterFingerprint(spec) {
			return domain.Materialization(fmt.Sprintf(
				"adapter drift: plan fingerprint for %s@%s does not match the catalog adapter",
				c.Boilerplate, c.Pin))
		}
		if spec.Generate != nil {
			// Fail before staging exists: resolve the placeholders
			// with deterministic dummy values and pass the
			// substituted argv through the same argv-only gate as
			// setup/checks. Unknown placeholders fail here.
			name, err := generateAppName(c.Destination)
			if err != nil {
				return err
			}
			profile := c.Materialization.Profile
			if profile == "" {
				profile = "profile"
			}
			values := map[string]string{
				"name": name, "project": "project",
				"surface": c.Surface, "profile": profile, "output": "output",
			}
			argv, err := domain.SubstituteGeneratorPlaceholders(spec.Generate.Run.Run, values)
			if err != nil {
				return err
			}
			for _, cmd := range spec.Generate.Prepare {
				resolved, err := domain.SubstituteGeneratorPlaceholders(cmd.Run, values)
				if err != nil {
					return err
				}
				argv = append(argv, resolved...)
			}
			_ = argv
			check := append([]domain.AdapterCommand{{Run: argv}}, append(append([]domain.AdapterCommand{}, spec.Setup...), spec.Checks...)...)
			// Validate the generator argv alone first so failures
			// attribute to the generator stage, then setup+checks.
			genArgv, err := domain.SubstituteGeneratorPlaceholders(spec.Generate.Run.Run, values)
			if err != nil {
				return err
			}
			if err := ValidateCommands([]domain.AdapterCommand{{Run: genArgv}}); err != nil {
				return err
			}
			_ = check
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
