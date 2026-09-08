package catalog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// writeEvidence creates dir/curation/<name> with body, returning dir.
func writeEvidence(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "curation"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "curation", name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func curationCatalog(bp domain.Boilerplate) catalog.Catalog {
	return catalog.Catalog{
		CatalogVersion: "test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{bp},
	}
}

func validPilotReadyEvidence() string {
	return "# Evidence\n\n- License: MIT (checked).\n- Repo and pin verified on review.\n"
}

// TestCurationEnforcement is the §3.5 contract: released without evidence is
// invalid, catalog-only without evidence is valid, broken references are
// invalid, and curated entries with evidence are valid.
func TestCurationEnforcement(t *testing.T) {
	base := domain.Boilerplate{
		ID: "fixture", Repo: "https://example.com/fixture", Pin: "v1",
		Adapter: "fixture",
	}
	tests := []struct {
		name     string
		mutate   func(*domain.Boilerplate)
		evidence map[string]string // name -> body, written under curation/
		link     string            // extra symlink name -> target (escape probe)
		wantErr  string
	}{
		{
			name:    "released without evidence is invalid",
			mutate:  func(b *domain.Boilerplate) { b.DeliveryStatus = "released" },
			wantErr: "requires curation evidence",
		},
		{
			name:    "curated without evidence is invalid",
			mutate:  func(b *domain.Boilerplate) { b.DeliveryStatus = "curated" },
			wantErr: "requires curation evidence",
		},
		{
			name:   "catalog-only without evidence is valid",
			mutate: func(b *domain.Boilerplate) { b.DeliveryStatus = "catalog-only" },
		},
		{
			name: "pilot-ready with basic evidence is valid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": validPilotReadyEvidence()},
		},
		{
			name: "curated with evidence and pilot record is valid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "curated"
				b.Curation = domain.Curation{Status: "curated", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": validPilotReadyEvidence() +
				"- Pilot: 2026-09-08 offline pilot passed (fetch, copy, checks green).\n"},
		},
		{
			name: "curated without pilot record is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "curated"
				b.Curation = domain.Curation{Status: "curated", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": validPilotReadyEvidence()},
			wantErr:  "Pilot",
		},
		{
			name: "released with denied pilot record is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "released"
				b.Curation = domain.Curation{Status: "released", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": validPilotReadyEvidence() +
				"- Pilot: not run — pilot scheduled, not executed.\n"},
			wantErr: "Pilot",
		},
		{
			name: "pilot-ready without license mention is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": "# Evidence\n\n- Repo verified.\n"},
			wantErr:  "license",
		},
		{
			name: "missing evidence file is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "curation/gone.md"}
			},
			wantErr: "does not exist",
		},
		{
			name: "path traversal is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "../secret.md"}
			},
			wantErr: "escapes the catalog",
		},
		{
			name: "absolute path is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "/etc/passwd"}
			},
			wantErr: "relative slash path",
		},
		{
			name: "symlink escape is invalid",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "pilot-ready", Evidence: "curation/link.md"}
			},
			link:    "link.md",
			wantErr: "escapes the catalog",
		},
		{
			name: "curation status must match delivery status",
			mutate: func(b *domain.Boilerplate) {
				b.DeliveryStatus = "pilot-ready"
				b.Curation = domain.Curation{Status: "curated", Evidence: "curation/fixture.md"}
			},
			evidence: map[string]string{"fixture.md": validPilotReadyEvidence()},
			wantErr:  "single status axis",
		},
		{
			name:    "unknown delivery status is invalid",
			mutate:  func(b *domain.Boilerplate) { b.DeliveryStatus = "beta" },
			wantErr: "unknown delivery_status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, body := range tt.evidence {
				writeEvidence(t, dir, name, body)
			}
			if tt.link != "" {
				outside := t.TempDir()
				target := filepath.Join(outside, "real.md")
				if err := os.WriteFile(target, []byte(validPilotReadyEvidence()), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(dir, "curation"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, filepath.Join(dir, "curation", tt.link)); err != nil {
					t.Fatal(err)
				}
			}
			bp := base
			tt.mutate(&bp)
			err := catalog.ValidateCuration(curationCatalog(bp), []string{dir})
			if tt.wantErr == "" && err != nil {
				t.Fatalf("ValidateCuration: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

// TestCurationOverlayEvidence proves adding evidence is a data-only
// operation: a synthetic overlay entry carrying its own curation link and
// stub loads and validates with no core changes.
func TestCurationOverlayEvidence(t *testing.T) {
	overlay := t.TempDir()
	if err := os.WriteFile(filepath.Join(overlay, "metadata.json"),
		[]byte(`{"catalog_version":"0.0.0-curation","min_core_version":"1.0.0","schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(overlay, "boilerplates"), 0o755); err != nil {
		t.Fatal(err)
	}
	entry := `{
      "id": "overlay-widget",
      "repo": "https://example.com/overlay-widget",
      "pin": "v9.9.9",
      "adapter": {"name": "overlay-widget", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
      "delivery_status": "pilot-ready",
      "decision_status": "curated",
      "curation": {"status": "pilot-ready", "evidence": "curation/overlay-widget.md"},
      "provides": {"surfaces": [], "capabilities": []},
      "tech_tags": ["widget"]
    }`
	if err := os.WriteFile(filepath.Join(overlay, "boilerplates", "overlay-widget.json"), []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	writeEvidence(t, overlay, "overlay-widget.md", validPilotReadyEvidence())
	cat, err := catalog.Load(overlay)
	if err != nil {
		t.Fatalf("load merged catalog: %v", err)
	}
	found := false
	for _, b := range cat.Boilerplates {
		if b.ID == "overlay-widget" {
			found = true
		}
	}
	if !found {
		t.Fatal("overlay entry did not merge into the catalog")
	}
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("structural validation: %v", err)
	}
	if err := catalog.ValidateCuration(cat, []string{overlay, catalog.BaseDir()}); err != nil {
		t.Fatalf("curation validation: %v", err)
	}
}

// TestBaseCatalogCurationValid keeps the shipped catalog honest: every
// entry above catalog-only links resolvable evidence, so `eng catalog
// validate` stays green without exceptions.
func TestBaseCatalogCurationValid(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if err := catalog.ValidateCuration(cat, []string{catalog.BaseDir()}); err != nil {
		t.Fatalf("base catalog curation: %v", err)
	}
}
