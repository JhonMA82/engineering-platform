package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndValidateBaseCatalog(t *testing.T) {
	cat, err := Load("")
	if err != nil {
		t.Fatalf("load base catalog: %v", err)
	}
	if err := Validate(cat); err != nil {
		t.Fatalf("base catalog invalid: %v", err)
	}
	if len(cat.Recipes) != 7 {
		t.Fatalf("expected 7 recipes, got %d", len(cat.Recipes))
	}
	idx := NewIndex(cat)
	if got, ok := idx.CanonicalAlias("panel"); !ok || got != "web-admin" {
		t.Fatalf("alias panel -> %q, %v", got, ok)
	}
	if got, ok := idx.CanonicalAlias("landing"); !ok || got != "public-web" {
		t.Fatalf("alias landing -> %q, %v", got, ok)
	}
	if len(idx.ProvidersForSurface("tui")) != 0 {
		t.Fatal("expected no provider for tui (gap fixture)")
	}
}

func TestOverlayMergeReplacesByID(t *testing.T) {
	base, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	dir := t.TempDir()
	meta := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(meta, []byte(`{"catalog_version":"9.9.9","min_core_version":"1.0.0","schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	overlay, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("load overlay: %v", err)
	}
	merged := MergeOverlay(base, overlay)
	if merged.CatalogVersion != "9.9.9" {
		t.Fatalf("overlay version not applied: %q", merged.CatalogVersion)
	}
	if len(merged.Recipes) != len(base.Recipes) {
		t.Fatalf("recipe count changed by empty overlay: %d", len(merged.Recipes))
	}
}

func TestTransactionalPwaBoarded(t *testing.T) {
	cat, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	idx := NewIndex(cat)
	bp, ok := idx.Boilerplate("tanstack-transactional-pwa")
	if !ok {
		t.Fatal("tanstack-transactional-pwa missing from catalog")
	}
	if err := bp.Validate(); err != nil {
		t.Fatalf("boilerplate invalid: %v", err)
	}
	for _, want := range []string{"public-intake", "mobile-native"} {
		found := false
		for _, s := range bp.Provides.Surfaces {
			if string(s) == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("transactional-pwa does not provide surface %q", want)
		}
	}
	recipe, ok := idx.Recipe("GP-06")
	if !ok {
		t.Fatal("GP-06 missing from catalog")
	}
	primary := false
	for _, id := range recipe.PrimaryBoilerplates {
		if id == "tanstack-transactional-pwa" {
			primary = true
			break
		}
	}
	if !primary {
		t.Fatal("GP-06 does not list tanstack-transactional-pwa as a primary boilerplate")
	}
}
