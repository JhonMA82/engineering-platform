// Package evolve plans and applies Fase 10 project evolution.
//
// Boundary: the resolver decides, the composer maps, the planner binds and
// the materializer writes component trees. This package orchestrates those
// pure stages over a materialized project directory and performs only
// project-file writes (the .engineering copies, agent-context refresh,
// manifest, map and provenance). It never changes resolver, composer,
// planner or materializer-core semantics: every stage is reused as-is.
//
// Provenance is append-only: evolution adds events, never rewrites history.
package evolve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/handoff"
	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

// Project is the loaded evolution input: the stored intent plus the
// machine-readable state derived from it.
type Project struct {
	Dir        string
	Intent     domain.ProjectIntent
	Decision   domain.ArchitectureDecision
	Plan       planner.MaterializationPlan
	HasPlan    bool
	Manifest   project.Manifest
	Map        project.ProjectMap
	Provenance project.Provenance
}

// LoadProject reads the stored intent, decision, plan, manifest, map and
// provenance of a materialized project. A project materialized without an
// intent copy cannot evolve: there is no product truth to extend.
func LoadProject(dir string) (Project, error) {
	var p Project
	p.Dir = dir
	intentRaw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(handoff.IntentCopyRelPath)))
	if err != nil {
		return Project{}, domain.Filesystem(fmt.Sprintf("evolve: project has no stored intent (%s): %v", handoff.IntentCopyRelPath, err))
	}
	if err := json.Unmarshal(intentRaw, &p.Intent); err != nil {
		return Project{}, domain.Validation(fmt.Sprintf("evolve: parse stored intent: %v", err))
	}
	decisionRaw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(handoff.DecisionCopyRelPath)))
	if err != nil {
		return Project{}, domain.Filesystem(fmt.Sprintf("evolve: project has no stored decision (%s): %v", handoff.DecisionCopyRelPath, err))
	}
	if err := json.Unmarshal(decisionRaw, &p.Decision); err != nil {
		return Project{}, domain.Validation(fmt.Sprintf("evolve: parse stored decision: %v", err))
	}
	if planRaw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(handoff.PlanCopyRelPath))); err == nil {
		var plan planner.MaterializationPlan
		if err := json.Unmarshal(planRaw, &plan); err != nil {
			return Project{}, domain.Validation(fmt.Sprintf("evolve: parse stored plan: %v", err))
		}
		p.Plan = plan
		p.HasPlan = true
	}
	manifest, err := project.ReadManifest(dir)
	if err != nil {
		return Project{}, err
	}
	p.Manifest = manifest
	pmap, err := project.ReadProjectMap(dir)
	if err != nil {
		return Project{}, err
	}
	p.Map = pmap
	prov, err := project.ReadProvenance(dir)
	if err != nil {
		return Project{}, err
	}
	p.Provenance = prov
	return p, nil
}

// canonicalSurface resolves a user-supplied surface term through the catalog
// alias table and rejects anything the catalog does not register.
func canonicalSurface(term string, idx catalog.Index) (domain.SurfaceID, error) {
	t := strings.ToLower(strings.TrimSpace(term))
	if t == "" {
		return "", domain.Validation("evolve: surface must not be empty")
	}
	if c, ok := idx.CanonicalAlias(t); ok {
		t = strings.ToLower(strings.TrimSpace(c))
	}
	id := domain.SurfaceID(t)
	if !idx.SurfaceKnown(id) {
		return "", domain.Validation(fmt.Sprintf("evolve: unknown surface %q (not in the catalog surfaces or aliases)", term))
	}
	return id, nil
}

// intentSurfaceScopes maps canonical surface kinds to their stored scope.
func intentSurfaceScopes(intent domain.ProjectIntent) map[domain.SurfaceID]domain.ScopeStatus {
	out := map[domain.SurfaceID]domain.ScopeStatus{}
	for _, s := range intent.Surfaces {
		kind := domain.SurfaceID(strings.ToLower(strings.TrimSpace(string(s.Kind))))
		if _, ok := out[kind]; !ok {
			out[kind] = s.EffectiveScope()
		}
	}
	return out
}

// manifestSurfaces maps materialized surface names to destinations.
func manifestSurfaces(m project.Manifest) map[string]string {
	out := map[string]string{}
	for _, c := range m.Components {
		out[c.Surface] = c.Destination
	}
	return out
}

// resolveEvolved validates and resolves an evolved intent. A non-resolved
// outcome aborts with a typed error before any project mutation.
func resolveEvolved(intent domain.ProjectIntent, cat catalog.Catalog) (domain.ArchitectureDecision, error) {
	if err := intent.Validate(); err != nil {
		return domain.ArchitectureDecision{}, err
	}
	decision := resolver.Resolve(intent, cat)
	if decision.Status != domain.StatusResolved {
		return decision, domain.Resolution(
			fmt.Sprintf("cannot evolve: evolved intent is %q, not resolved (%s)",
				decision.Status, firstReason(decision)))
	}
	if decision.Selected == nil {
		return decision, domain.Resolution("cannot evolve: resolution selected no recipe")
	}
	return decision, nil
}

func firstReason(d domain.ArchitectureDecision) string {
	if len(d.Reasons) > 0 {
		return d.Reasons[0]
	}
	return "no reason recorded"
}

// requireSameRecipe aborts when the evolved intent resolves to a different
// recipe than the project was materialized with. Recipe migration is out of
// v1 scope; the message says so.
func requireSameRecipe(decision domain.ArchitectureDecision, m project.Manifest) error {
	if decision.Selected.Recipe != m.Recipe {
		return domain.Resolution(fmt.Sprintf(
			"cannot evolve: evolved intent resolves to recipe %q but the project was materialized with %q; "+
				"recipe migration is out of v1 scope — materialize a new project for the new architecture instead",
			decision.Selected.Recipe, m.Recipe))
	}
	return nil
}

// applyProviderOverride swaps the composed provider for one surface with an
// explicitly requested boilerplate. The override is validated against the
// active catalog; destinations still come from composition.
func applyProviderOverride(comp *composer.Composition, surface domain.SurfaceID, provider string, cat catalog.Catalog, decision domain.ArchitectureDecision) error {
	if strings.TrimSpace(provider) == "" {
		return nil
	}
	idx := catalog.NewIndex(cat)
	bp, ok := idx.Boilerplate(provider)
	if !ok {
		return domain.Catalog(fmt.Sprintf("evolve: provider %q is not in the catalog", provider))
	}
	provides := false
	for _, s := range bp.Provides.Surfaces {
		if s == surface {
			provides = true
			break
		}
	}
	if !provides {
		return domain.Composition(fmt.Sprintf(
			"evolve: provider %q does not serve surface %q", provider, string(surface)))
	}
	if !composer.Eligible(bp) {
		return domain.Composition(fmt.Sprintf(
			"evolve: provider %q is not eligible (delivery/decision state, pin or adapter)", provider))
	}
	matched := false
	for i := range comp.Components {
		if comp.Components[i].Surface == surface {
			comp.Components[i].Boilerplate = bp.ID
			comp.Components[i].Pin = bp.Pin
			matched = true
		}
	}
	if !matched {
		return domain.Composition(fmt.Sprintf(
			"evolve: composed project has no %q surface to assign provider %q to", string(surface), provider))
	}
	return nil
}

// computeDelta returns the evolved-plan components no materialized surface
// covers yet. Existing surfaces must keep their destinations: evolution adds
// directories, never moves them.
func computeDelta(m project.Manifest, plan planner.MaterializationPlan) ([]planner.PlanComponent, error) {
	known := manifestSurfaces(m)
	want := map[string]string{}
	for _, c := range m.Components {
		want[c.Surface] = c.Destination
	}
	var delta []planner.PlanComponent
	for _, c := range plan.Components {
		dest, ok := want[c.Surface]
		if !ok {
			delta = append(delta, c)
			continue
		}
		if dest != c.Destination {
			return nil, domain.Composition(fmt.Sprintf(
				"cannot evolve: surface %q would move from %q to %q; evolution never moves materialized surfaces",
				c.Surface, dest, c.Destination))
		}
	}
	_ = known
	if len(delta) == 0 {
		return nil, domain.Composition("cannot evolve: the evolved plan adds no new components")
	}
	sort.Slice(delta, func(i, j int) bool {
		if delta[i].Destination != delta[j].Destination {
			return delta[i].Destination < delta[j].Destination
		}
		return delta[i].Surface < delta[j].Surface
	})
	return delta, nil
}

// AddedComponent describes one materialized delta component.
type AddedComponent struct {
	Surface     string `json:"surface"`
	Boilerplate string `json:"boilerplate"`
	Pin         string `json:"pin"`
	Destination string `json:"destination"`
}

// EvolutionResult summarizes an applied architecture evolution.
type EvolutionResult struct {
	Recipe          string           `json:"recipe"`
	Added           []AddedComponent `json:"added"`
	PlanFingerprint string           `json:"plan_fingerprint"`
}

// SurfaceAddOptions carries the surface-add inputs. Catalog is the active
// catalog; Now stamps the provenance event.
type SurfaceAddOptions struct {
	Surface  string
	Provider string
	Catalog  catalog.Catalog
	Now      time.Time
}

// SurfaceAdd evolves the project architecture with one new required surface:
// resolve the evolved intent, compose the full project, materialize only the
// delta, refresh map/manifest/provenance/docs. Any pre-write failure aborts
// with a typed error and no partial mutation.
func SurfaceAdd(projectDir string, opts SurfaceAddOptions) (*EvolutionResult, error) {
	p, err := LoadProject(projectDir)
	if err != nil {
		return nil, err
	}
	idx := catalog.NewIndex(opts.Catalog)
	surface, err := canonicalSurface(opts.Surface, idx)
	if err != nil {
		return nil, err
	}
	scopes := intentSurfaceScopes(p.Intent)
	if scope, ok := scopes[surface]; ok {
		return nil, domain.Validation(fmt.Sprintf(
			"evolve: surface %q is already present in the project intent (scope %q)", string(surface), string(scope)))
	}
	if _, ok := manifestSurfaces(p.Manifest)[string(surface)]; ok {
		return nil, domain.Validation(fmt.Sprintf(
			"evolve: surface %q is already materialized in the project", string(surface)))
	}
	evolved := p.Intent
	evolved.Surfaces = append(append([]domain.SurfaceIntent{}, p.Intent.Surfaces...),
		domain.SurfaceIntent{Kind: surface, Scope: domain.ScopeRequiredNow})
	return applyArchitectureEvolution(p, evolved, opts.Provider, opts.Catalog, opts.Now,
		project.EvolutionEvent{Type: project.EventSurfaceAdd, Surface: string(surface), Recipe: p.Manifest.Recipe})
}

// ExtendOptions carries the scope-extension inputs.
type ExtendOptions struct {
	Surface string
	Catalog catalog.Catalog
	Now     time.Time
}

// ExtendScope promotes a planned_later surface to required_now and ships it
// through the same pipeline as surface-add. Anything else is a usage error:
// an unknown or already-required surface names the command that fits.
func ExtendScope(projectDir string, opts ExtendOptions) (*EvolutionResult, error) {
	p, err := LoadProject(projectDir)
	if err != nil {
		return nil, err
	}
	idx := catalog.NewIndex(opts.Catalog)
	surface, err := canonicalSurface(opts.Surface, idx)
	if err != nil {
		return nil, err
	}
	scopes := intentSurfaceScopes(p.Intent)
	scope, ok := scopes[surface]
	if !ok {
		return nil, domain.Validation(fmt.Sprintf(
			"evolve: surface %q is not in the project intent; use 'eng surface add' to introduce a new surface", string(surface)))
	}
	if scope != domain.ScopePlannedLater {
		return nil, domain.Validation(fmt.Sprintf(
			"evolve: surface %q has scope %q, not planned_later; nothing to extend", string(surface), string(scope)))
	}
	evolved := p.Intent
	surfaces := append([]domain.SurfaceIntent{}, p.Intent.Surfaces...)
	for i := range surfaces {
		if domain.SurfaceID(strings.ToLower(strings.TrimSpace(string(surfaces[i].Kind)))) == surface {
			surfaces[i].Scope = domain.ScopeRequiredNow
		}
	}
	evolved.Surfaces = surfaces
	return applyArchitectureEvolution(p, evolved, "", opts.Catalog, opts.Now,
		project.EvolutionEvent{Type: project.EventScopeExtend, Surface: string(surface), Recipe: p.Manifest.Recipe})
}

// applyArchitectureEvolution runs the shared resolve→compose→plan→delta→write
// pipeline. Resolution, recipe-identity and delta validation all complete
// before the first project write, so aborts leave the project untouched.
func applyArchitectureEvolution(p Project, evolved domain.ProjectIntent, provider string, cat catalog.Catalog, now time.Time, event project.EvolutionEvent) (*EvolutionResult, error) {
	decision, err := resolveEvolved(evolved, cat)
	if err != nil {
		return nil, err
	}
	if err := requireSameRecipe(decision, p.Manifest); err != nil {
		return nil, err
	}
	comp, err := composer.Compose(decision, cat)
	if err != nil {
		return nil, err
	}
	comp.Project = evolved.Name
	newSurface := domain.SurfaceID(event.Surface)
	if err := applyProviderOverride(&comp, newSurface, provider, cat, decision); err != nil {
		return nil, err
	}
	plan, err := planner.Plan(comp, decision, cat)
	if err != nil {
		return nil, err
	}
	delta, err := computeDelta(p.Manifest, plan)
	if err != nil {
		return nil, err
	}
	existing := make([]string, 0, len(p.Manifest.Components))
	for _, c := range p.Manifest.Components {
		existing = append(existing, c.Destination)
	}
	if err := materializer.CheckCollisions(append(existing, deltaDests(delta)...)); err != nil {
		return nil, err
	}
	intentJSON, err := json.Marshal(evolved)
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("evolve: encode intent: %v", err))
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("evolve: encode decision: %v", err))
	}
	if err := materializer.MaterializeDelta(materializer.DeltaRequest{
		Components:           delta,
		Catalog:              cat,
		ProjectDir:           p.Dir,
		ExistingDestinations: existing,
	}); err != nil {
		return nil, err
	}
	if err := refreshGeneratedFiles(p.Dir, handoff.Input{
		Plan:         plan,
		Catalog:      cat,
		IntentJSON:   intentJSON,
		DecisionJSON: decisionJSON,
	}); err != nil {
		return nil, err
	}
	if err := writeProjectMap(p.Dir, plan); err != nil {
		return nil, err
	}
	if err := rebuildManifest(p.Dir, plan, cat.CatalogVersion); err != nil {
		return nil, err
	}
	pins := map[string]string{}
	for _, c := range plan.Components {
		pins[c.Boilerplate] = c.Pin
	}
	prov, err := project.ReadProvenance(p.Dir)
	if err != nil {
		return nil, err
	}
	prov.IntentFingerprint = decision.IntentFingerprint
	prov.PlanFingerprint = plan.Fingerprint
	prov.Pins = pins
	if err := project.WriteProvenance(p.Dir, prov); err != nil {
		return nil, err
	}
	for _, c := range delta {
		if c.Surface == event.Surface {
			event.Provider = c.Boilerplate
		}
	}
	if err := project.AppendEvent(p.Dir, event, now); err != nil {
		return nil, err
	}
	if err := requireDoctorGreen(p.Dir); err != nil {
		return nil, err
	}
	res := &EvolutionResult{Recipe: plan.Recipe, PlanFingerprint: plan.Fingerprint}
	for _, c := range delta {
		res.Added = append(res.Added, AddedComponent{
			Surface: c.Surface, Boilerplate: c.Boilerplate, Pin: c.Pin, Destination: c.Destination,
		})
	}
	return res, nil
}

func deltaDests(delta []planner.PlanComponent) []string {
	out := make([]string, 0, len(delta))
	for _, c := range delta {
		out = append(out, c.Destination)
	}
	return out
}

// refreshGeneratedFiles writes the handoff-rendered agent-context and
// .engineering copies. Overwriting files are refreshed; surface stubs never
// overwrite, so foundation-shipped instructions survive evolution.
func refreshGeneratedFiles(projectDir string, in handoff.Input) error {
	files, err := handoff.Render(in)
	if err != nil {
		return err
	}
	for _, f := range files {
		target := filepath.Join(projectDir, filepath.FromSlash(f.Path))
		if !f.Overwrite {
			if _, err := os.Stat(target); err == nil {
				continue
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return domain.Filesystem(fmt.Sprintf("evolve: create dir for %q: %v", f.Path, err))
		}
		if err := os.WriteFile(target, f.Data, 0o644); err != nil {
			return domain.Filesystem(fmt.Sprintf("evolve: write %q: %v", f.Path, err))
		}
	}
	return nil
}

func writeProjectMap(projectDir string, plan planner.MaterializationPlan) error {
	raw, err := project.BuildProjectMap(plan).Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(projectDir, filepath.FromSlash(project.ProjectMapRelPath)), raw, 0o644); err != nil {
		return domain.Filesystem(fmt.Sprintf("evolve: write %s: %v", project.ProjectMapRelPath, err))
	}
	return nil
}

// rebuildManifest re-lists the project directory and records the evolved
// plan fingerprint, pins and file list.
func rebuildManifest(projectDir string, plan planner.MaterializationPlan, catalogVersion string) error {
	files, err := materializer.ListFiles(projectDir)
	if err != nil {
		return err
	}
	manifest := project.BuildManifest(plan, catalogVersion, files)
	raw, err := manifest.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(projectDir, filepath.FromSlash(project.ManifestRelPath)), raw, 0o644); err != nil {
		return domain.Filesystem(fmt.Sprintf("evolve: write %s: %v", project.ManifestRelPath, err))
	}
	return nil
}

// requireDoctorGreen keeps evolution honest: a mutation that does not leave
// a consistent project is reported as a materialization failure.
func requireDoctorGreen(projectDir string) error {
	findings, err := project.Doctor(projectDir)
	if err != nil {
		return err
	}
	if project.HasErrors(findings) {
		msgs := make([]string, 0, len(findings))
		for _, f := range findings {
			if f.Severity == project.SeverityError {
				msgs = append(msgs, f.Code+": "+f.Message)
			}
		}
		sort.Strings(msgs)
		return domain.Materialization(fmt.Sprintf("evolution left the project inconsistent: %s", strings.Join(msgs, "; ")))
	}
	return nil
}
