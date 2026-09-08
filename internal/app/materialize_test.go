package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEndToEndAdminOnlyOffline runs the M3 spine without network: the
// admin-only routing intent resolves against a fixture-source overlay
// catalog, the resulting plan materializes into a temp dir, and doctor
// reports a consistent project.
func TestEndToEndAdminOnlyOffline(t *testing.T) {
	fixtureRaw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "routing", "admin-only.json"))
	if err != nil {
		t.Fatalf("read routing fixture: %v", err)
	}
	var wrapper struct {
		Intent json.RawMessage `json:"intent"`
	}
	if err := json.Unmarshal(fixtureRaw, &wrapper); err != nil {
		t.Fatalf("parse routing fixture: %v", err)
	}
	intentJSON := []byte(wrapper.Intent)

	webSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-web"))
	if err != nil {
		t.Fatal(err)
	}
	pinRaw, err := os.ReadFile(filepath.Join(webSrc, "PIN"))
	if err != nil {
		t.Fatalf("read fixture PIN: %v", err)
	}
	pin := strings.TrimSpace(string(pinRaw))

	overlay := t.TempDir()
	if err := os.WriteFile(filepath.Join(overlay, "metadata.json"),
		[]byte(`{"catalog_version":"0.0.0-e2e","min_core_version":"1.0.0","schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	bpDir := filepath.Join(overlay, "boilerplates")
	if err := os.MkdirAll(bpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overlayBoilerplate := fmt.Sprintf(`{
      "id": "tanstack-admin",
      "repo": "https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard",
      "pin": %q,
      "adapter": {"name": "tanstack", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
      "source": {"type": "local", "path": %q},
      "delivery_status": "stable",
      "decision_status": "curated",
      "provides": {"surfaces": ["web-admin"], "capabilities": []},
      "tech_tags": ["tanstack"],
      "included_features": ["crud", "auth"]
    }`, pin, webSrc)
	if err := os.WriteFile(filepath.Join(bpDir, "tanstack-admin.json"), []byte(overlayBoilerplate), 0o644); err != nil {
		t.Fatal(err)
	}

	decision, comp, plan, err := PlanProject(intentJSON, overlay)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if comp.Recipe != "GP-02" {
		t.Fatalf("recipe = %q, want GP-02", comp.Recipe)
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	manifest, err := MaterializeProject(planJSON, intentJSON, decisionJSON, overlay, out)
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if len(manifest.Components) != 1 || manifest.Components[0].Destination != "apps/admin" {
		t.Fatalf("unexpected manifest components: %+v", manifest.Components)
	}

	findings, err := DoctorProject(out)
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	for _, f := range findings {
		if f.Severity == "error" {
			t.Errorf("doctor error: %s: %s", f.Code, f.Message)
		}
	}
	for _, want := range []string{
		"apps/admin/index.html",
		".engineering/project-intent.json",
		".engineering/architecture-decision.json",
		".engineering/materialization-plan.json",
		".engineering/project-map.json",
		".engineering/implementation-brief.md",
		".engineering/handoff.json",
		".engineering/project.json",
		".engineering/provenance.json",
		"AGENTS.md",
		"ARCHITECTURE.md",
		"GENTLE.md",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected %s: %v", want, err)
		}
	}
}
