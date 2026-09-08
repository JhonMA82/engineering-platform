// Package composer maps a resolved ArchitectureDecision onto provider
// boilerplates without touching the filesystem, network or processes.
//
// Boundary: the resolver decides the recipe family and foundation fit;
// the composer decides how the required surfaces are satisfied with
// compatible providers, destinations and database profile. Product
// features never add components: the composer only reads surfaces and
// architectural capabilities out of the decision, never product
// requirements.
//
// Deterministic: stable sorts, no time/rand/map iteration in outputs.
// Only stdlib plus internal/domain and internal/catalog are imported.
package composer

import (
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Component is one materializable unit: a required surface served by a
// pinned provider boilerplate at a relative destination.
type Component struct {
	Surface     domain.SurfaceID `json:"surface"`
	Boilerplate string           `json:"boilerplate"`
	Pin         string           `json:"pin"`
	Destination string           `json:"destination"`
}

// Composition is the composer output consumed by the planner.
type Composition struct {
	// Project is the intent name. A decision carries no name, so Compose
	// leaves it empty and app.PlanProject fills it from the intent.
	Project         string      `json:"project,omitempty"`
	Recipe          string      `json:"recipe"`
	RecipeVersion   string      `json:"recipe_version,omitempty"`
	Components      []Component `json:"components"`
	DatabaseProfile string      `json:"database_profile"`
}

// Compose maps a resolved decision onto providers with default destinations.
func Compose(decision domain.ArchitectureDecision, cat catalog.Catalog) (Composition, error) {
	return ComposeWithDestinations(decision, cat, nil)
}

// ComposeWithDestinations maps a resolved decision onto providers, applying
// explicit per-surface destination overrides on top of the defaults.
// Overrides exist so callers and tests can pin destinations; collisions and
// unsafe paths are still rejected with typed errors.
func ComposeWithDestinations(decision domain.ArchitectureDecision, cat catalog.Catalog, overrides map[domain.SurfaceID]string) (Composition, error) {
	if decision.Status != domain.StatusResolved {
		return Composition{}, domain.Composition(
			"cannot compose: decision status is " + string(decision.Status) + " (need resolved)")
	}
	if decision.Selected == nil {
		return Composition{}, domain.Composition("cannot compose: decision selects no recipe")
	}
	idx := catalog.NewIndex(cat)
	recipe, ok := idx.Recipe(decision.Selected.Recipe)
	if !ok {
		return Composition{}, domain.Composition(
			"cannot compose: recipe " + decision.Selected.Recipe + " is not in the catalog")
	}

	surfaces := RequiredSurfaces(decision, idx)
	caps := RequiredCapabilities(decision, idx)
	mustUse := MustUseTech(decision)
	mustNotUse := MustNotUseTech(decision)

	var components []Component
	bySurface := map[domain.SurfaceID]domain.Boilerplate{}
	for _, surface := range surfaces {
		provider, err := SelectProvider(surface, recipe, decision, cat, caps)
		if err != nil {
			return Composition{}, err
		}
		bySurface[surface] = provider
		components = append(components, Component{
			Surface:     surface,
			Boilerplate: provider.ID,
			Pin:         provider.Pin,
		})
	}
	// A recipe with a shared backend needs the backend component even when
	// the intent did not name the api surface explicitly: multiple clients
	// consume the same api over shared business data.
	if recipe.DatabasePolicy.SharedBackend {
		if _, ok := bySurface[domain.SurfaceID("api")]; !ok {
			provider, err := SelectProvider(domain.SurfaceID("api"), recipe, decision, cat, caps)
			if err != nil {
				return Composition{}, err
			}
			bySurface[domain.SurfaceID("api")] = provider
			components = append(components, Component{
				Surface:     domain.SurfaceID("api"),
				Boilerplate: provider.ID,
				Pin:         provider.Pin,
			})
		}
	}

	surfaceList := make([]domain.SurfaceID, 0, len(components))
	for _, c := range components {
		surfaceList = append(surfaceList, c.Surface)
	}
	destinations, err := AssignDestinations(surfaceList, overrides)
	if err != nil {
		return Composition{}, err
	}
	for i := range components {
		components[i].Destination = destinations[components[i].Surface]
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Surface < components[j].Surface })

	profile, err := SelectDatabaseProfile(recipe, cat)
	if err != nil {
		return Composition{}, err
	}
	out := Composition{
		Recipe:          recipe.ID,
		RecipeVersion:   recipe.Version,
		Components:      components,
		DatabaseProfile: profile,
	}
	if err := CheckCompatibility(decision, recipe, out, caps, mustUse, mustNotUse, cat); err != nil {
		return Composition{}, err
	}
	return out, nil
}
