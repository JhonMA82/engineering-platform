package resolver

import (
	"fmt"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// DiagnoseGap runs R7: every required surface, capability ref and derived
// requirement not provided by ANY active recipe becomes missing
// architecture with research criteria. Product misses never land here.
func DiagnoseGap(n Normalized, sp Split, derived []domain.DerivedRequirement, idx catalog.Index) []domain.MissingPiece {
	coveredSurface := map[domain.SurfaceID]bool{}
	coveredCap := map[domain.CapabilityID]bool{}
	for _, r := range idx.ActiveRecipes() {
		for _, s := range r.Provides.Surfaces {
			coveredSurface[s] = true
		}
		for _, c := range r.Provides.Capabilities {
			coveredCap[c] = true
		}
	}
	var missing []domain.MissingPiece
	seen := map[string]bool{}
	add := func(kind, ref string) {
		key := kind + "|" + ref
		if seen[key] {
			return
		}
		seen[key] = true
		missing = append(missing, domain.MissingPiece{
			Kind: kind,
			Ref:  ref,
			ResearchCriteria: []string{
				fmt.Sprintf("curate or register a foundation providing %s=%s", kind, ref),
				fmt.Sprintf("define the provider contract for %s=%s", kind, ref),
				"re-run resolve after the catalog update; do not invent an uncurated provider",
			},
		})
	}
	for _, s := range n.RequiredSurfaces {
		if !coveredSurface[s] {
			add("surface", string(s))
		}
	}
	for _, ref := range n.RequiredRefs {
		if idx.CapabilityKnown(domain.CapabilityID(ref)) {
			if !coveredCap[domain.CapabilityID(ref)] {
				add("capability", ref)
			}
			continue
		}
		if idx.SurfaceKnown(domain.SurfaceID(ref)) {
			if !coveredSurface[domain.SurfaceID(ref)] {
				add("surface", ref)
			}
			continue
		}
		add("capability", ref)
	}
	for _, d := range derived {
		if !coveredCap[domain.CapabilityID(d.ID)] {
			add("capability", d.ID)
		}
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Kind != missing[j].Kind {
			return missing[i].Kind < missing[j].Kind
		}
		return missing[i].Ref < missing[j].Ref
	})
	return missing
}
