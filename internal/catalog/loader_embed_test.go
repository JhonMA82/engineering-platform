package catalog

import (
	"os"
	"reflect"
	"testing"
)

// TestEmbeddedBaseCatalogFallback proves a globally installed eng works
// without a checkout: with no catalog tree resolvable from the working
// directory, Load falls back to the binary-embedded base catalog.
func TestEmbeddedBaseCatalogFallback(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()
	if err := os.Chdir(empty); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()
	got, err := LoadWithCore("", "1.1.0")
	if err != nil {
		t.Fatalf("embedded fallback: %v", err)
	}
	if got.CatalogVersion != "1.1.0" {
		t.Fatalf("CatalogVersion = %q, want 1.1.0", got.CatalogVersion)
	}
	found := false
	for _, bp := range got.Boilerplates {
		if bp.ID == "hono-api" {
			found = true
			for _, p := range bp.EffectiveSpec().Generate.Profiles {
				if p.ID == "authenticated" && len(p.Arguments) == 1 && p.Arguments[0] == "--profile=authenticated" {
					return
				}
			}
			t.Fatal("embedded hono-api lacks the authenticated profile arguments")
		}
	}
	if !found {
		t.Fatal("embedded catalog lacks hono-api")
	}
}

// TestEmbeddedBaseMatchesTree proves the embedded catalog is byte-identical
// in content to the in-repo tree: embedding must never silently fork.
func TestEmbeddedBaseMatchesTree(t *testing.T) {
	fromTree, err := LoadDirWithCore("../../catalog", "1.1.0")
	if err != nil {
		t.Fatalf("tree load: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()
	fromEmbedded, err := LoadWithCore("", "1.1.0")
	if err != nil {
		t.Fatalf("embedded load: %v", err)
	}
	if !reflect.DeepEqual(fromTree, fromEmbedded) {
		t.Fatal("embedded catalog differs from the in-repo tree")
	}
}
