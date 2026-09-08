package catalog_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Legacy migration integrity baseline (§10-11). The fixture below preserves
// the historical identity minimum (id → repo → pin-or-null) for the 11
// legacy boilerplates. It is NEVER loaded by the runtime: catalog.Load only
// reads catalog/ subdirectories. This suite fails closed if a future rewrite
// infers a repository URL from the catalog id, invents a release tag, drops
// an entry silently, or promotes a non-selectable entry.

type legacyBaselineEntry struct {
	ID   string  `json:"id"`
	Repo string  `json:"repo"`
	Pin  *string `json:"pin"`
}

type legacyBaseline struct {
	Entries    []legacyBaselineEntry `json:"entries"`
	Tombstones []struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	} `json:"tombstones"`
}

func loadLegacyBaseline(t *testing.T) legacyBaseline {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "testdata", "catalog", "legacy-boilerplate-baseline.json"))
	if err != nil {
		t.Fatalf("read legacy baseline: %v", err)
	}
	var base legacyBaseline
	if err := json.Unmarshal(raw, &base); err != nil {
		t.Fatalf("parse legacy baseline: %v", err)
	}
	if len(base.Entries) != 11 {
		t.Fatalf("baseline entries = %d, want the 11 historical ids", len(base.Entries))
	}
	return base
}

func loadCatalogForLegacy(t *testing.T) (catalog.Catalog, catalog.Index) {
	t.Helper()
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	return cat, catalog.NewIndex(cat)
}

// TestLegacyCatalogCanonicalRepositories fails if any restored mapping
// regresses to an id-derived repository (e.g. hono-api → JhonMA82/hono-api
// instead of the historical JhonMA82/api-starter).
func TestLegacyCatalogCanonicalRepositories(t *testing.T) {
	base := loadLegacyBaseline(t)
	_, idx := loadCatalogForLegacy(t)
	for _, e := range base.Entries {
		t.Run(e.ID, func(t *testing.T) {
			bp, ok := idx.Boilerplate(e.ID)
			if !ok {
				t.Fatalf("%s missing from catalog", e.ID)
			}
			if bp.EffectiveRepo() != e.Repo {
				t.Fatalf("repo = %q, want canonical %q", bp.EffectiveRepo(), e.Repo)
			}
		})
	}
}

// TestLegacyCatalogCanonicalPins pins the known historical commits. Entries
// without a historical pin are covered by TestNoInventedLegacyPins.
func TestLegacyCatalogCanonicalPins(t *testing.T) {
	base := loadLegacyBaseline(t)
	_, idx := loadCatalogForLegacy(t)
	for _, e := range base.Entries {
		if e.Pin == nil {
			continue
		}
		t.Run(e.ID, func(t *testing.T) {
			bp, ok := idx.Boilerplate(e.ID)
			if !ok {
				t.Fatalf("%s missing from catalog", e.ID)
			}
			if bp.Pin != *e.Pin {
				t.Fatalf("pin = %q, want historical %q", bp.Pin, *e.Pin)
			}
		})
	}
}

// TestLegacyCatalogIDsPreserved restores the historical id: stardrive is the
// canonical provider and the invented stardrive-public-web id must not exist
// as an active provider (reader-side manifest alias only).
func TestLegacyCatalogIDsPreserved(t *testing.T) {
	_, idx := loadCatalogForLegacy(t)
	if _, ok := idx.Boilerplate("stardrive"); !ok {
		t.Fatal("canonical id stardrive missing from catalog")
	}
	if _, ok := idx.Boilerplate("stardrive-public-web"); ok {
		t.Fatal("stardrive-public-web must not exist as an active provider (alias stardrive)")
	}
}

// TestNoInventedRepositoryFromCatalogID forbids deriving the repository from
// the catalog id by textual similarity.
func TestNoInventedRepositoryFromCatalogID(t *testing.T) {
	_, idx := loadCatalogForLegacy(t)
	forbidden := map[string][]string{
		"hono-api":       {"https://github.com/JhonMA82/hono-api"},
		"tanstack-admin": {"https://github.com/JhonMA82/tanstack-admin"},
		"stardrive": {
			"https://github.com/JhonMA82/stardrive-public-web",
			"https://github.com/JhonMA82/stardrive",
		},
	}
	for id, repos := range forbidden {
		t.Run(id, func(t *testing.T) {
			bp, ok := idx.Boilerplate(id)
			if !ok {
				t.Fatalf("%s missing from catalog", id)
			}
			for _, bad := range repos {
				if bp.EffectiveRepo() == bad {
					t.Fatalf("repo %q was inferred from the catalog id", bad)
				}
			}
		})
	}
}

// TestNoInventedLegacyPins forbids release-tag pins: no migrated entry may
// carry v1.0.0, and entries with no historical pin must stay pin-less
// (catalog-only) instead of receiving an invented one.
func TestNoInventedLegacyPins(t *testing.T) {
	base := loadLegacyBaseline(t)
	_, idx := loadCatalogForLegacy(t)
	for _, e := range base.Entries {
		t.Run(e.ID, func(t *testing.T) {
			bp, ok := idx.Boilerplate(e.ID)
			if !ok {
				t.Fatalf("%s missing from catalog", e.ID)
			}
			if bp.Pin == "v1.0.0" {
				t.Fatal("invented v1.0.0 pin: use the historical commit or stay pin-less")
			}
			if e.Pin == nil && bp.Pin != "" {
				t.Fatalf("pin %q invented: legacy records no pin for %s", bp.Pin, e.ID)
			}
		})
	}
}

// TestAllLegacyEntriesRepresented requires every historical id to exist in
// the catalog or to carry an explicit migration tombstone with a reason.
// Every represented entry must record its legacy provenance.
func TestAllLegacyEntriesRepresented(t *testing.T) {
	base := loadLegacyBaseline(t)
	_, idx := loadCatalogForLegacy(t)
	tombstoned := map[string]string{}
	for _, ts := range base.Tombstones {
		if strings.TrimSpace(ts.Reason) == "" {
			t.Fatalf("tombstone %s carries no reason", ts.ID)
		}
		tombstoned[ts.ID] = ts.Reason
	}
	for _, e := range base.Entries {
		t.Run(e.ID, func(t *testing.T) {
			bp, ok := idx.Boilerplate(e.ID)
			if !ok {
				if _, retired := tombstoned[e.ID]; retired {
					return
				}
				t.Fatalf("%s missing from catalog and has no migration tombstone", e.ID)
			}
			if strings.TrimSpace(bp.Provenance.Source) == "" {
				t.Fatalf("%s records no provenance source", e.ID)
			}
			if bp.Provenance.Source != "engineering-platform-legacy" {
				t.Fatalf("%s provenance source = %q, want engineering-platform-legacy", e.ID, bp.Provenance.Source)
			}
			raw, err := json.Marshal(bp)
			if err != nil {
				t.Fatalf("marshal %s: %v", e.ID, err)
			}
			var wire struct {
				Provenance *domain.Provenance `json:"provenance"`
			}
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatalf("remarshal %s: %v", e.ID, err)
			}
			if wire.Provenance == nil || wire.Provenance.Source != "engineering-platform-legacy" {
				t.Fatalf("%s provenance does not survive a JSON round-trip", e.ID)
			}
		})
	}
}

// TestLegacyExperimentalEntriesRemainNonSelectable keeps catalog knowledge
// out of the candidate pool: goship (experimental/catalog-only, no pin, no
// adapter) and fastapi (unpinned catalog-only) must never serve a surface.
func TestLegacyExperimentalEntriesRemainNonSelectable(t *testing.T) {
	cat, idx := loadCatalogForLegacy(t)
	for _, id := range []string{"goship", "fastapi"} {
		t.Run(id, func(t *testing.T) {
			bp, ok := idx.Boilerplate(id)
			if !ok {
				t.Fatalf("%s missing from catalog", id)
			}
			if composer.Eligible(bp) {
				t.Fatalf("%s is eligible: non-selectable entries must stay out of the pool", id)
			}
		})
	}
	for _, surface := range []domain.SurfaceID{"api", "public-web", "web-admin", "mobile-native", "desktop"} {
		for _, bp := range composer.EligibleProviders(cat, surface) {
			if bp.ID == "goship" || bp.ID == "fastapi" {
				t.Fatalf("%s serves surface %s: must remain non-selectable", bp.ID, surface)
			}
		}
	}
	if bp, ok := idx.Boilerplate("goship"); ok && bp.DecisionStatus != "experimental" {
		t.Fatalf("goship decision = %q, want experimental", bp.DecisionStatus)
	}
}

// TestLegacyAlternativeEntriesDoNotBecomeDefault keeps the recovered
// alternatives honest: next-admin stays alternative (eligible, but never the
// default — the admin default remains tanstack-admin) and fastapi is never
// promoted by migration alone.
func TestLegacyAlternativeEntriesDoNotBecomeDefault(t *testing.T) {
	_, idx := loadCatalogForLegacy(t)
	for _, id := range []string{"next-admin", "fastapi"} {
		t.Run(id, func(t *testing.T) {
			bp, ok := idx.Boilerplate(id)
			if !ok {
				t.Fatalf("%s missing from catalog", id)
			}
			if bp.DecisionStatus == "default" {
				t.Fatalf("%s became default: alternatives must be promoted by curation, not migration", id)
			}
			if bp.DecisionStatus != "alternative" {
				t.Fatalf("%s decision = %q, want alternative", id, bp.DecisionStatus)
			}
		})
	}
	next, ok := idx.Boilerplate("next-admin")
	if !ok {
		t.Fatal("next-admin missing from catalog")
	}
	if !composer.Eligible(next) {
		t.Fatal("next-admin should be eligible (pin+adapter present): only its decision keeps it from default")
	}
	def, ok := idx.Boilerplate("tanstack-admin")
	if !ok {
		t.Fatal("tanstack-admin missing from catalog")
	}
	if def.DecisionStatus != "default" {
		t.Fatalf("tanstack-admin decision = %q, want default (the admin default must not move)", def.DecisionStatus)
	}
}
