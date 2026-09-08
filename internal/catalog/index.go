package catalog

import (
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Index offers lookup helpers over a loaded catalog.
type Index struct {
	recipes      map[string]domain.Recipe
	boilerplates map[string]domain.Boilerplate
	surfaces     map[domain.SurfaceID]domain.Surface
	capabilities map[domain.CapabilityID]domain.Capability
	profiles     map[string]domain.DatabaseProfile
	aliases      map[string]string
}

// NewIndex builds lookup helpers for cat.
func NewIndex(cat Catalog) Index {
	idx := Index{
		recipes:      map[string]domain.Recipe{},
		boilerplates: map[string]domain.Boilerplate{},
		surfaces:     map[domain.SurfaceID]domain.Surface{},
		capabilities: map[domain.CapabilityID]domain.Capability{},
		profiles:     map[string]domain.DatabaseProfile{},
		aliases:      map[string]string{},
	}
	for _, r := range cat.Recipes {
		idx.recipes[r.ID] = r
	}
	for _, b := range cat.Boilerplates {
		idx.boilerplates[b.ID] = b
	}
	for _, s := range cat.Surfaces {
		idx.surfaces[s.ID] = s
	}
	for _, c := range cat.Capabilities {
		idx.capabilities[c.ID] = c
	}
	for _, p := range cat.DatabaseProfiles {
		idx.profiles[p.ID] = p
	}
	for k, v := range cat.Aliases {
		idx.aliases[strings.ToLower(k)] = v
	}
	return idx
}

// Recipe returns the recipe with id.
func (x Index) Recipe(id string) (domain.Recipe, bool) {
	r, ok := x.recipes[id]
	return r, ok
}

// Boilerplate returns the boilerplate with id.
func (x Index) Boilerplate(id string) (domain.Boilerplate, bool) {
	b, ok := x.boilerplates[id]
	return b, ok
}

// DatabaseProfile returns the database profile with id.
func (x Index) DatabaseProfile(id string) (domain.DatabaseProfile, bool) {
	p, ok := x.profiles[id]
	return p, ok
}

// SurfaceKnown reports whether id is a registered surface.
func (x Index) SurfaceKnown(id domain.SurfaceID) bool {
	_, ok := x.surfaces[id]
	return ok
}

// CapabilityKnown reports whether id is a registered capability.
func (x Index) CapabilityKnown(id domain.CapabilityID) bool {
	_, ok := x.capabilities[id]
	return ok
}

// CanonicalAlias maps a free term to its canonical surface/capability id.
func (x Index) CanonicalAlias(term string) (string, bool) {
	v, ok := x.aliases[strings.ToLower(strings.TrimSpace(term))]
	return v, ok
}

// ActiveRecipes returns active recipes sorted by id for determinism.
func (x Index) ActiveRecipes() []domain.Recipe {
	var out []domain.Recipe
	for _, r := range x.recipes {
		if r.Active() {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ProvidersForSurface returns active recipes providing surface, sorted by id.
func (x Index) ProvidersForSurface(id domain.SurfaceID) []domain.Recipe {
	var out []domain.Recipe
	for _, r := range x.ActiveRecipes() {
		for _, s := range r.Provides.Surfaces {
			if s == id {
				out = append(out, r)
			}
		}
	}
	return out
}

// ProvidersForCapability returns active recipes providing cap, sorted by id.
func (x Index) ProvidersForCapability(id domain.CapabilityID) []domain.Recipe {
	var out []domain.Recipe
	for _, r := range x.ActiveRecipes() {
		for _, c := range r.Provides.Capabilities {
			if c == id {
				out = append(out, r)
			}
		}
	}
	return out
}

// RecipeTechnology returns the catalog-known technology signals of a recipe
// under one constraint target, deduplicated and sorted for determinism.
// Sources are catalog data only: the recipe tech_tags (legacy untyped
// signals, kept as a documented fallback) plus the technology metadata of
// the recipe primary boilerplates. The provider target additionally covers
// the boilerplate identity (id and adapter name). Database and deployment
// targets have no recipe-level metadata: database resolves through profiles
// (see MatchDatabaseProfile), deployment is not yet curated and yields
// nothing.
func (x Index) RecipeTechnology(r domain.Recipe, target string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(values ...string) {
		for _, v := range values {
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			out = append(out, v)
		}
	}
	switch target {
	case domain.ConstraintTargetFramework, domain.ConstraintTargetLanguage, domain.ConstraintTargetRuntime:
		for _, id := range r.PrimaryBoilerplates {
			if b, ok := x.boilerplates[id]; ok {
				add(b.Technology.Values(target)...)
			}
		}
		add(r.TechTags...)
	case domain.ConstraintTargetProvider:
		for _, id := range r.PrimaryBoilerplates {
			add(id)
			if b, ok := x.boilerplates[id]; ok {
				add(b.Adapter)
				add(b.TechTags...)
			}
		}
		add(r.TechTags...)
	}
	sort.Strings(out)
	return out
}

// MatchDatabaseProfile reports whether value identifies any curated
// database profile via its id, engine or provider (see
// DatabaseProfile.MatchesTechnology). No brand names are hardcoded here:
// the comparison runs against catalog data.
func (x Index) MatchDatabaseProfile(value string) (domain.DatabaseProfile, bool) {
	for _, p := range x.profiles {
		if p.MatchesTechnology(value) {
			return p, true
		}
	}
	return domain.DatabaseProfile{}, false
}

// RecipeDatabaseProfiles returns the profile ids a recipe may use: its
// allowed profiles, or the default alone when no allow-list is declared.
func (x Index) RecipeDatabaseProfiles(r domain.Recipe) []string {
	if len(r.DatabasePolicy.AllowedProfiles) > 0 {
		return append([]string{}, r.DatabasePolicy.AllowedProfiles...)
	}
	if r.DatabasePolicy.DefaultProfile != "" {
		return []string{r.DatabasePolicy.DefaultProfile}
	}
	return nil
}
