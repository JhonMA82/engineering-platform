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
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// DeltaRequest materializes additional components into an existing project
// directory. Unlike Request, the project dir must already exist and must NOT
// be empty: only brand-new destinations are written, existing content is
// never touched. The top-level Materialize empty-dir guard is intentionally
// kept; evolution flows use this entrypoint instead.
type DeltaRequest struct {
	Components []planner.PlanComponent
	Catalog    catalog.Catalog
	ProjectDir string
	// ExistingDestinations guards against nesting into materialized surfaces.
	ExistingDestinations []string
	// CommandTimeout bounds curated adapter commands; <=0 selects the default.
	CommandTimeout time.Duration
}

// MaterializeDelta fetches, verifies and copies each requested component
// into the existing project directory. Every destination must be absent; a
// present destination aborts with a typed error before anything is written
// for that component. Components are applied in destination order; a failure
// leaves already-applied earlier components in place, so callers must
// validate everything (pins, collisions, resolution) before invoking it.
func MaterializeDelta(req DeltaRequest) error {
	if len(req.Components) == 0 {
		return domain.Materialization("cannot materialize delta: no new components")
	}
	if strings.TrimSpace(req.ProjectDir) == "" {
		return domain.Filesystem("project directory must not be empty")
	}
	st, err := os.Stat(req.ProjectDir)
	if err != nil || !st.IsDir() {
		return domain.Filesystem(fmt.Sprintf("project directory %q is not a readable dir: %v", req.ProjectDir, err))
	}
	idx := catalog.NewIndex(req.Catalog)
	all := append(append([]string{}, req.ExistingDestinations...), deltaDests(req.Components)...)
	if err := CheckCollisions(all); err != nil {
		return err
	}
	for _, c := range req.Components {
		if err := ValidateDestination(c.Destination); err != nil {
			return err
		}
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
		if err := ValidateCommands(append(append([]domain.AdapterCommand{}, spec.Setup...), spec.Checks...)); err != nil {
			return err
		}
		if bp.Source.Kind() == "local" && !filepath.IsAbs(bp.Source.Path) {
			return domain.Materialization(fmt.Sprintf(
				"local source for %q must be an absolute path (got %q)", bp.ID, bp.Source.Path))
		}
		dest := projectSubdir(req.ProjectDir, c.Destination)
		if _, err := os.Stat(dest); err == nil {
			return domain.Materialization(fmt.Sprintf("destination collision: %q already exists in the project", c.Destination))
		}
	}
	ordered := append([]planner.PlanComponent{}, req.Components...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Destination != ordered[j].Destination {
			return ordered[i].Destination < ordered[j].Destination
		}
		return ordered[i].Surface < ordered[j].Surface
	})
	fetchRoot, err := os.MkdirTemp("", "eng-delta-fetch-*")
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("create fetch dir: %v", err))
	}
	defer os.RemoveAll(fetchRoot)
	ctx := context.Background()
	for _, c := range ordered {
		bp, _ := idx.Boilerplate(c.Boilerplate)
		spec := bp.EffectiveSpec()
		srcDir, err := FetchSource(ctx, bp, c.Pin, fetchRoot)
		if err != nil {
			return err
		}
		dest := projectSubdir(req.ProjectDir, c.Destination)
		if _, err := os.Stat(dest); err == nil {
			return domain.Materialization(fmt.Sprintf("destination collision: %q already exists in the project", c.Destination))
		}
		if err := CopyTree(srcDir, dest, spec.PrunePaths); err != nil {
			return err
		}
		if len(spec.Setup) > 0 {
			if _, err := RunCommands(ctx, spec.Setup, dest, req.CommandTimeout); err != nil {
				return err
			}
		}
		if len(spec.Checks) > 0 {
			if _, err := RunCommands(ctx, spec.Checks, dest, req.CommandTimeout); err != nil {
				return err
			}
		}
	}
	return nil
}

func deltaDests(components []planner.PlanComponent) []string {
	out := make([]string, 0, len(components))
	for _, c := range components {
		out = append(out, c.Destination)
	}
	return out
}
