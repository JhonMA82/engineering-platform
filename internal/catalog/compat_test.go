package catalog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
)

func writeMeta(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestCatalogRejectsNewerRequiredCore proves the §1.2 gate: a catalog that
// needs a newer core than the running binary fails fast with the exact
// release-hardening vocabulary.
func TestCatalogRejectsNewerRequiredCore(t *testing.T) {
	cases := []struct {
		name string
		min  string
		core string
	}{
		{"major ahead", "9.9.9", "1.0.0"},
		{"minor ahead", "1.1.0", "1.0.0"},
		{"patch ahead", "1.0.1", "1.0.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeMeta(t, `{"catalog_version":"9.9.9","min_core_version":"`+tc.min+`","schema_version":1}`)
			_, err := catalog.LoadDirWithCore(dir, tc.core)
			if err == nil {
				t.Fatal("expected core rejection, got nil")
			}
			want := "catalog requires core >= " + tc.min + ", running core is " + tc.core
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), want)
			}
		})
	}
}

// TestCatalogAcceptsCompatibleCore proves equal and greater core lines load,
// across the accepted version forms (plain, v-prefixed, short).
func TestCatalogAcceptsCompatibleCore(t *testing.T) {
	cases := []struct {
		name string
		min  string
		core string
	}{
		{"equal", "1.0.0", "1.0.0"},
		{"greater minor", "1.0.0", "1.2.0"},
		{"greater major", "1.0.0", "2.0.0"},
		{"v-prefixed min", "v1.0.0", "1.0.0"},
		{"short min", "1.0", "1.0.0"},
		{"dev bypass", "9.9.9", "dev"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeMeta(t, `{"catalog_version":"9.9.9","min_core_version":"`+tc.min+`","schema_version":1}`)
			if _, err := catalog.LoadDirWithCore(dir, tc.core); err != nil {
				t.Fatalf("compatible core rejected: %v", err)
			}
		})
	}
}

// TestCatalogRejectsUnsupportedSchema proves the schema gate keeps the exact
// §1.2 vocabulary: "unsupported catalog schema version: N".
func TestCatalogRejectsUnsupportedSchema(t *testing.T) {
	dir := writeMeta(t, `{"catalog_version":"9.9.9","min_core_version":"1.0.0","schema_version":999}`)
	_, err := catalog.LoadDirWithCore(dir, "1.0.0")
	if err == nil {
		t.Fatal("expected schema rejection, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported catalog schema version: 999") {
		t.Fatalf("error = %q, want the §1.2 schema vocabulary", err.Error())
	}
}

// TestCatalogRejectsNewerMaxCore proves the optional max_core_version bound:
// a binary newer than the catalog ceiling is rejected instead of silently
// running against a catalog it outgrew.
func TestCatalogRejectsNewerMaxCore(t *testing.T) {
	dir := writeMeta(t, `{"catalog_version":"1.0.0","min_core_version":"1.0.0","max_core_version":"1.9.9","schema_version":1}`)
	_, err := catalog.LoadDirWithCore(dir, "2.0.0")
	if err == nil {
		t.Fatal("expected max-core rejection, got nil")
	}
	want := "catalog requires core <= 1.9.9, running core is 2.0.0"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err.Error(), want)
	}
}
