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
		criteria := []string{
			fmt.Sprintf("curate or register a foundation providing %s=%s", kind, ref),
			fmt.Sprintf("define the provider contract for %s=%s", kind, ref),
			"re-run resolve after the catalog update; do not invent an uncurated provider",
		}
		if kind == "database-profile" {
			criteria = []string{
				fmt.Sprintf("curate a database profile for %s (engine/provider/supports per the catalog contract)", ref),
				"do not invent an uncurated provider: contract first, profiles later",
				"re-run resolve after the catalog update",
			}
		}
		missing = append(missing, domain.MissingPiece{
			Kind:             kind,
			Ref:              ref,
			ResearchCriteria: criteria,
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
	// A must-use database value with no curated profile anywhere is an
	// architectural gap, not an intent error: the catalog owns the missing
	// foundation (kind database-profile), and curation — not the user —
	// must close it. Values that do match a profile but no recipe policy
	// stay on the unsupported path (covered-but-eliminated).
	for _, db := range sp.DatabaseMustUse {
		if _, ok := idx.MatchDatabaseProfile(db); !ok {
			add("database-profile", db)
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

// DiagnoseDatabaseGap runs the database leg of R7: every must-use
// database value with no curated catalog profile becomes a
// database-profile missing piece, regardless of recipe eligibility. The
// resolver owns this verdict (not the composer) so the outcome is a typed
// catalog-gap decision instead of a late composition error.
func DiagnoseDatabaseGap(sp Split, idx catalog.Index) []domain.MissingPiece {
	var missing []domain.MissingPiece
	seen := map[string]bool{}
	for _, db := range sp.DatabaseMustUse {
		if seen[db] {
			continue
		}
		seen[db] = true
		if _, ok := idx.MatchDatabaseProfile(db); !ok {
			missing = append(missing, domain.MissingPiece{
				Kind: "database-profile",
				Ref:  db,
				ResearchCriteria: []string{
					fmt.Sprintf("curate a database profile for %s (engine/provider/supports per the catalog contract)", db),
					"do not invent an uncurated provider: contract first, profiles later",
					"re-run resolve after the catalog update",
				},
			})
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].Ref < missing[j].Ref })
	return missing
}
