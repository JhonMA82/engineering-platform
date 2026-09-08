package catalog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// TestLoadRejectsUnknownSchema proves the §30 gate: a catalog tree declaring
// a schema revision the core does not understand fails fast with a typed
// catalog error instead of feeding unknown contracts into the engine.
func TestLoadRejectsUnknownSchema(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"),
		[]byte(`{"catalog_version":"9.9.9","min_core_version":"1.0.0","schema_version":999}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := catalog.LoadDir(dir)
	if err == nil {
		t.Fatal("expected schema rejection, got nil")
	}
	derr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected typed domain error, got %T: %v", err, err)
	}
	if derr.Class != domain.ClassCatalog {
		t.Fatalf("expected catalog-class error, got %q", derr.Class)
	}
}

// TestLoadAcceptsSupportedSchema keeps the gate honest: the in-repo base
// catalog and schema 1 overlays still load.
func TestLoadAcceptsSupportedSchema(t *testing.T) {
	if _, err := catalog.Load(""); err != nil {
		t.Fatalf("base catalog must load: %v", err)
	}
}
