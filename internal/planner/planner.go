// Package planner converts a Composition plus its ArchitectureDecision into
// a serializable MaterializationPlan without touching the filesystem.
//
// Deterministic: components are ordered by destination, no timestamps or
// randomness. Only stdlib plus internal/domain and internal/composer types
// are imported.
package planner

import (
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
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
	for _, c := range comp.Components {
		components = append(components, PlanComponent{
			Boilerplate: c.Boilerplate,
			Pin:         c.Pin,
			Destination: c.Destination,
			Surface:     string(c.Surface),
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
		operations = append(operations,
			"materialize "+c.Destination+" from "+c.Boilerplate+"@"+c.Pin)
	}
	setup := []string{"provision database profile " + comp.DatabaseProfile}
	checks := append([]string{}, recipe.QualityGates...)
	if checks == nil {
		checks = []string{}
	}
	plan := MaterializationPlan{
		SchemaVersion:   1,
		Project:         comp.Project,
		Recipe:          comp.Recipe,
		RecipeVersion:   comp.RecipeVersion,
		Components:      components,
		DatabaseProfile: comp.DatabaseProfile,
		Operations:      operations,
		Setup:           setup,
		Checks:          checks,
		Fingerprint: FingerprintPlan(
			decision.IntentFingerprint, cat.CatalogVersion, comp),
	}
	return plan, nil
}
