// Package planner converts a Composition plus its ArchitectureDecision into
// a serializable MaterializationPlan without touching the filesystem.
//
// Deterministic: components are ordered by destination, no timestamps or
// randomness. Only stdlib plus internal/domain, internal/composer,
// internal/catalog and internal/foundationconfig types are imported.
package planner

import (
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/foundationconfig"
)

// Plan converts a composition and its decision into a deterministic
// MaterializationPlan. Operations materialize each component, setup
// provisions the database profile, and checks carry the recipe quality
// gates in catalog order.
func Plan(comp composer.Composition, decision domain.ArchitectureDecision, cat catalog.Catalog) (MaterializationPlan, error) {
	if decision.Selected == nil {
		return MaterializationPlan{}, domain.Composition("cannot plan: decision selects no recipe")
	}
	idx := catalog.NewIndex(cat)
	recipe, ok := idx.Recipe(comp.Recipe)
	if !ok {
		return MaterializationPlan{}, domain.Composition(
			"cannot plan: recipe " + comp.Recipe + " is not in the catalog")
	}
	components := make([]PlanComponent, 0, len(comp.Components))
	mats := make(map[domain.SurfaceID]foundationconfig.MaterializationConfig, len(comp.Components))
	for _, c := range comp.Components {
		bp, ok := idx.Boilerplate(c.Boilerplate)
		if !ok {
			return MaterializationPlan{}, domain.Composition(
				"cannot plan: boilerplate " + c.Boilerplate + " is not in the catalog")
		}
		mat, err := foundationconfig.Resolve(c, comp.Project, len(comp.Components), bp, decision, cat)
		if err != nil {
			return MaterializationPlan{}, err
		}
		mats[c.Surface] = mat
		components = append(components, PlanComponent{
			Boilerplate:     c.Boilerplate,
			Pin:             c.Pin,
			Destination:     c.Destination,
			Surface:         string(c.Surface),
			Materialization: mat,
		})
	}
	sort.Slice(components, func(i, j int) bool {
		if components[i].Destination != components[j].Destination {
			return components[i].Destination < components[j].Destination
		}
		return components[i].Surface < components[j].Surface
	})
	operations := make([]string, 0, len(components))
	for _, c := range components {
		op := "materialize " + c.Destination + " from " + c.Boilerplate + "@" + c.Pin
		if c.EffectiveStrategy() == domain.StrategyGenerate {
			if c.Materialization.Profile != "" {
				op += " [generate profile=" + c.Materialization.Profile + "]"
			} else {
				op += " [generate]"
			}
		}
		operations = append(operations, op)
	}
	setup := []string{"provision database profile " + comp.DatabaseProfile}
	checks := append([]string{}, recipe.QualityGates...)
	if checks == nil {
		checks = []string{}
	}
	plan := MaterializationPlan{
		SchemaVersion:   2,
		Project:         comp.Project,
		Recipe:          comp.Recipe,
		RecipeVersion:   comp.RecipeVersion,
		Components:      components,
		DatabaseProfile: comp.DatabaseProfile,
		Operations:      operations,
		Setup:           setup,
		Checks:          checks,
		Fingerprint: FingerprintPlan(
			decision.IntentFingerprint, cat.CatalogVersion, comp, mats),
	}
	return plan, nil
}
