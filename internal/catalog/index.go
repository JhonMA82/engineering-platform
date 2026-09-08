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
