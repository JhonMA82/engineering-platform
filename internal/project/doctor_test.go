package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/planner"
)

func seedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	plan := planner.MaterializationPlan{
		SchemaVersion:   1,
		Project:         "doc-proj",
		Recipe:          "TEST",
		DatabaseProfile: "sqlite-local",
		Fingerprint:     "fp-1",
		Components: []planner.PlanComponent{
			{Boilerplate: "fixture-web", Pin: "v1", Destination: "apps/web", Surface: "public-web"},
		},
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps", "web"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, EngineeringDir), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(rel, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("apps/web/index.html", "<h1>hi</h1>")
	manifest := BuildManifest(plan, "0.0.0-test", []string{
		"apps/web/index.html",
		".engineering/project.json",
		".engineering/provenance.json",
		".engineering/project-map.json",
		".engineering/materialization-plan.json",
	})
	raw, err := manifest.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	write(ManifestRelPath, string(raw))
	prov, err := BuildProvenance("0.0.0-test", "0.0.0-test", "", "fp-1",
		map[string]string{"fixture-web": "v1"}, time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	write(ProvenanceRelPath, string(prov))
	pmap, err := BuildProjectMap(plan).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	write(ProjectMapRelPath, string(pmap))
	write(".engineering/materialization-plan.json", `{"fingerprint":"fp-1"}`)
	return dir
}

func TestDoctorConsistentProject(t *testing.T) {
	dir := seedProject(t)
	findings, err := Doctor(dir)
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	if HasErrors(findings) {
		t.Errorf("expected clean doctor, got %v", findings)
	}
}

func TestDoctorFindings(t *testing.T) {
	cases := []struct {
		name     string
		breakIt  func(t *testing.T, dir string)
		wantCode string
	}{
		{"missing file", func(t *testing.T, dir string) {
			t.Helper()
			_ = os.Remove(filepath.Join(dir, "apps", "web", "index.html"))
		}, "file-missing"},
		{"missing destination", func(t *testing.T, dir string) {
			t.Helper()
			_ = os.RemoveAll(filepath.Join(dir, "apps"))
		}, "destination-missing"},
		{"provenance drift", func(t *testing.T, dir string) {
			t.Helper()
			raw, _ := os.ReadFile(filepath.Join(dir, EngineeringDir, "provenance.json"))
			_ = raw
			_ = os.WriteFile(filepath.Join(dir, EngineeringDir, "provenance.json"),
				[]byte(`{"schema_version":"1","plan_fingerprint":"other"}`), 0o644)
		}, "provenance-mismatch"},
		{"map drift", func(t *testing.T, dir string) {
			t.Helper()
			_ = os.WriteFile(filepath.Join(dir, EngineeringDir, "project-map.json"),
				[]byte(`{"schema_version":1,"surfaces":{"public-web":{"path":"apps/elsewhere","provider":"fixture-web","instructions":"apps/elsewhere/AGENTS.md"}},"relationships":[]}`), 0o644)
		}, "project-map-drift"},
		{"unexpected entry warns", func(t *testing.T, dir string) {
			t.Helper()
			_ = os.WriteFile(filepath.Join(dir, "stray.txt"), []byte("x"), 0o644)
		}, "unexpected-entry"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := seedProject(t)
			tt.breakIt(t, dir)
			findings, err := Doctor(dir)
			if err != nil {
				t.Fatalf("doctor: %v", err)
			}
			for _, f := range findings {
				if f.Code == tt.wantCode {
					if tt.wantCode == "unexpected-entry" && f.Severity != SeverityWarning {
						t.Errorf("unexpected-entry should warn, got %v", f)
					}
					return
				}
			}
			t.Errorf("want finding %q, got %v", tt.wantCode, findings)
		})
	}
}
