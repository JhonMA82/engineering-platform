package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/evolve"
	"github.com/jhonma82/engineering-platform/internal/project"
)

const (
	evolveWebAdminAPIIntent = `{"schema_version":1,"name":"evolve-base","problem":"ops share data over an api",` +
		`"surfaces":[{"kind":"web-admin","access":"authenticated","scope":"required_now"},` +
		`{"kind":"api","scope":"required_now"}]}`
	evolveWebAdminAPIPlannedIntakeIntent = `{"schema_version":1,"name":"evolve-base","problem":"ops share data over an api",` +
		`"surfaces":[{"kind":"web-admin","access":"authenticated","scope":"required_now"},` +
		`{"kind":"api","scope":"required_now"},` +
		`{"kind":"public-intake","access":"anonymous","scope":"planned_later"}]}`
	evolveAdminOnlyIntent = `{"schema_version":1,"name":"backoffice","problem":"operators manage records",` +
		`"surfaces":[{"kind":"web-admin","access":"authenticated","scope":"required_now"}]}`
)

// evolveOverlay builds an offline overlay catalog: the three GP-06 providers
// needed for the web-admin/api/public-intake growth path serve local fixture
// sources. hono-api declares a replace strategy so the update report has a
// non-default strategy to surface.
func evolveOverlay(t *testing.T) string {
	t.Helper()
	webSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-web"))
	if err != nil {
		t.Fatal(err)
	}
	apiSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-api"))
	if err != nil {
		t.Fatal(err)
	}
	webPin := readPIN(t, webSrc)
	apiPin := readPIN(t, apiSrc)
	overlay := t.TempDir()
	if err := os.WriteFile(filepath.Join(overlay, "metadata.json"),
		[]byte(`{"catalog_version":"0.0.0-evolve","min_core_version":"1.0.0","schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	bpDir := filepath.Join(overlay, "boilerplates")
	if err := os.MkdirAll(bpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	entries := map[string]string{
		"tanstack-admin": fmt.Sprintf(`{
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
        }`, webPin, webSrc),
		"hono-api": fmt.Sprintf(`{
          "id": "hono-api",
          "repo": "https://github.com/JhonMA82/api-starter",
          "pin": %q,
          "adapter": {"name": "hono", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
          "source": {"type": "local", "path": %q},
          "delivery_status": "stable",
          "decision_status": "curated",
          "provides": {"surfaces": ["api"], "capabilities": ["shared-backend"]},
          "tech_tags": ["hono"],
          "included_features": ["rest", "openapi"],
          "update_strategy": "replace"
        }`, apiPin, apiSrc),
		"tanstack-transactional-pwa": fmt.Sprintf(`{
          "id": "tanstack-transactional-pwa",
          "repo": "https://github.com/JhonMA82/tanstack-transactional-pwa",
          "pin": %q,
          "adapter": {"name": "tanstack", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
          "source": {"type": "local", "path": %q},
          "delivery_status": "stable",
          "decision_status": "curated",
          "provides": {"surfaces": ["public-web", "public-intake", "mobile-native"], "capabilities": ["offline-operation", "anonymous-public-access"]},
          "tech_tags": ["tanstack"],
          "included_features": ["offline-sync", "installable-pwa"]
        }`, webPin, webSrc),
	}
	for id, body := range entries {
		if err := os.WriteFile(filepath.Join(bpDir, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return overlay
}

func readPIN(t *testing.T, src string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(src, "PIN"))
	if err != nil {
		t.Fatalf("read fixture PIN: %v", err)
	}
	return strings.TrimSpace(string(raw))
}

// materializeEvolveProject plans and materializes intentJSON with overlay,
// returning the project dir and the resolved recipe.
func materializeEvolveProject(t *testing.T, overlay, intentJSON string) (string, string) {
	t.Helper()
	decision, comp, plan, err := PlanProject([]byte(intentJSON), overlay)
	if err != nil {
		t.Fatalf("plan: %v", err)
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
	if _, err := MaterializeProject(planJSON, []byte(intentJSON), decisionJSON, overlay, out); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	return out, comp.Recipe
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func storedIntent(t *testing.T, projectDir string) domain.ProjectIntent {
	t.Helper()
	var intent domain.ProjectIntent
	if err := json.Unmarshal([]byte(readFile(t, filepath.Join(projectDir, ".engineering", "project-intent.json"))), &intent); err != nil {
		t.Fatalf("parse stored intent: %v", err)
	}
	return intent
}

func assertDoctorGreen(t *testing.T, projectDir string) {
	t.Helper()
	findings, err := DoctorProject(projectDir)
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	for _, f := range findings {
		if f.Severity == project.SeverityError {
			t.Errorf("doctor error: %s: %s", f.Code, f.Message)
		}
	}
}

func manifestSurfaces(m project.Manifest) map[string]project.ManifestComponent {
	out := map[string]project.ManifestComponent{}
	for _, c := range m.Components {
		out[c.Surface] = c
	}
	return out
}

func TestSurfaceAddHappyPath(t *testing.T) {
	overlay := evolveOverlay(t)
	out, recipe := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	if recipe != "GP-06" {
		t.Fatalf("base recipe = %q, want GP-06", recipe)
	}
	adminBefore := readFile(t, filepath.Join(out, "apps", "admin", "index.html"))
	manifestBefore, err := project.ReadManifest(out)
	if err != nil {
		t.Fatal(err)
	}

	res, err := SurfaceAddProject(out, "public-intake", "", overlay)
	if err != nil {
		t.Fatalf("surface add: %v", err)
	}
	if res.Recipe != "GP-06" {
		t.Errorf("evolved recipe = %q, want GP-06", res.Recipe)
	}
	if len(res.Added) != 1 || res.Added[0].Surface != "public-intake" {
		t.Fatalf("added = %+v, want exactly the public-intake delta", res.Added)
	}
	if res.Added[0].Destination != "apps/intake" {
		t.Errorf("delta destination = %q, want apps/intake", res.Added[0].Destination)
	}

	// Delta-only: the new tree exists, existing content is byte-identical.
	if _, err := os.Stat(filepath.Join(out, "apps", "intake", "index.html")); err != nil {
		t.Errorf("expected delta file apps/intake/index.html: %v", err)
	}
	if got := readFile(t, filepath.Join(out, "apps", "admin", "index.html")); got != adminBefore {
		t.Errorf("existing surface file changed by surface-add")
	}

	manifest, err := project.ReadManifest(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Components) != len(manifestBefore.Components)+1 {
		t.Errorf("manifest components = %d, want %d", len(manifest.Components), len(manifestBefore.Components)+1)
	}
	if manifest.PlanFingerprint == manifestBefore.PlanFingerprint {
		t.Errorf("manifest fingerprint did not advance")
	}
	if _, ok := manifestSurfaces(manifest)["public-intake"]; !ok {
		t.Errorf("manifest misses the public-intake component")
	}

	intent := storedIntent(t, out)
	found := false
	for _, s := range intent.Surfaces {
		if strings.ToLower(string(s.Kind)) == "public-intake" && s.EffectiveScope() == domain.ScopeRequiredNow {
			found = true
		}
	}
	if !found {
		t.Errorf("stored intent misses required_now public-intake: %+v", intent.Surfaces)
	}

	pmap, err := project.ReadProjectMap(out)
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := pmap.Surfaces["public-intake"]
	if !ok || entry.Path != "apps/intake" {
		t.Errorf("project map misses public-intake at apps/intake: %+v", pmap.Surfaces)
	}

	prov, err := project.ReadProvenance(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(prov.Events) != 1 {
		t.Fatalf("provenance events = %d, want 1", len(prov.Events))
	}
	ev := prov.Events[0]
	if ev.Type != project.EventSurfaceAdd || ev.Surface != "public-intake" || ev.Recipe != "GP-06" {
		t.Errorf("unexpected evolution event: %+v", ev)
	}
	if ev.Provider == "" || ev.At == "" {
		t.Errorf("evolution event misses provider/at: %+v", ev)
	}
	if ev.Provider != manifestSurfaces(manifest)["public-intake"].Boilerplate {
		t.Errorf("event provider %q != materialized provider", ev.Provider)
	}
	if prov.PlanFingerprint != manifest.PlanFingerprint {
		t.Errorf("provenance fingerprint %q != manifest %q", prov.PlanFingerprint, manifest.PlanFingerprint)
	}

	assertDoctorGreen(t, out)
}

func TestSurfaceAddAbortCases(t *testing.T) {
	overlay := evolveOverlay(t)
	tests := []struct {
		name    string
		intent  string
		surface string
		wantErr string
	}{
		{
			name:    "recipe change aborts with migration policy",
			intent:  evolveAdminOnlyIntent,
			surface: "mobile-native",
			wantErr: "recipe migration is out of v1 scope",
		},
		{
			name:    "duplicate surface rejected",
			intent:  evolveWebAdminAPIIntent,
			surface: "web-admin",
			wantErr: "already present",
		},
		{
			name:    "unknown surface rejected",
			intent:  evolveWebAdminAPIIntent,
			surface: "starship",
			wantErr: "unknown surface",
		},
		{
			name:    "alias resolves before recipe check",
			intent:  evolveAdminOnlyIntent,
			surface: "mobile app",
			wantErr: "recipe migration is out of v1 scope",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _ := materializeEvolveProject(t, overlay, tt.intent)
			manifestBefore, err := project.ReadManifest(out)
			if err != nil {
				t.Fatal(err)
			}
			_, err = SurfaceAddProject(out, tt.surface, "", overlay)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
			// Clean abort: no partial mutation.
			if _, statErr := os.Stat(filepath.Join(out, "apps", "mobile")); !os.IsNotExist(statErr) && tt.surface != "web-admin" && tt.surface != "starship" {
				t.Errorf("aborted evolution left a mobile dir behind")
			}
			manifest, err := project.ReadManifest(out)
			if err != nil {
				t.Fatal(err)
			}
			if manifest.PlanFingerprint != manifestBefore.PlanFingerprint {
				t.Errorf("aborted evolution changed the manifest fingerprint")
			}
			assertDoctorGreen(t, out)
		})
	}
}

func TestSurfaceAddUnknownProvider(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	_, err := SurfaceAddProject(out, "public-intake", "no-such-boilerplate", overlay)
	if err == nil || !strings.Contains(err.Error(), "not in the catalog") {
		t.Fatalf("expected unknown-provider error, got %v", err)
	}
	assertDoctorGreen(t, out)
}

func TestExtendScope(t *testing.T) {
	overlay := evolveOverlay(t)
	t.Run("planned later promotes and ships", func(t *testing.T) {
		out, recipe := materializeEvolveProject(t, overlay, evolveWebAdminAPIPlannedIntakeIntent)
		if recipe != "GP-06" {
			t.Fatalf("base recipe = %q, want GP-06", recipe)
		}
		res, err := ExtendProjectScope(out, "public-intake", overlay)
		if err != nil {
			t.Fatalf("extend: %v", err)
		}
		if len(res.Added) != 1 || res.Added[0].Surface != "public-intake" {
			t.Fatalf("added = %+v, want the public-intake delta", res.Added)
		}
		intent := storedIntent(t, out)
		for _, s := range intent.Surfaces {
			if strings.ToLower(string(s.Kind)) == "public-intake" && s.EffectiveScope() != domain.ScopeRequiredNow {
				t.Fatalf("stored scope = %q, want required_now", s.EffectiveScope())
			}
		}
		prov, err := project.ReadProvenance(out)
		if err != nil {
			t.Fatal(err)
		}
		if len(prov.Events) != 1 || prov.Events[0].Type != project.EventScopeExtend {
			t.Fatalf("provenance events = %+v, want one scope-extend", prov.Events)
		}
		assertDoctorGreen(t, out)
	})

	t.Run("errors name the fitting command", func(t *testing.T) {
		tests := []struct {
			name    string
			surface string
			wantErr string
		}{
			{"absent surface suggests surface add", "mobile-native", "use 'eng surface add'"},
			{"already required has nothing to extend", "web-admin", "not planned_later"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIPlannedIntakeIntent)
				_, err := ExtendProjectScope(out, tt.surface, overlay)
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				assertDoctorGreen(t, out)
			})
		}
	})
}

func TestAddRequirement(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	manifestBefore, err := project.ReadManifest(out)
	if err != nil {
		t.Fatal(err)
	}
	briefBefore := readFile(t, filepath.Join(out, ".engineering", "implementation-brief.md"))

	res, err := AddProjectRequirement(out, "monthly PDF export for auditors", "", overlay)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if res.ID != "REQ-001" {
		t.Errorf("requirement id = %q, want REQ-001", res.ID)
	}
	if res.Warning != "" {
		t.Errorf("unexpected architecture warning: %q", res.Warning)
	}

	// Brief and handoff carry the pending requirement.
	if got := readFile(t, filepath.Join(out, ".engineering", "implementation-brief.md")); !strings.Contains(got, "monthly PDF export for auditors") || got == briefBefore {
		t.Errorf("brief did not refresh with the new requirement")
	}
	var handoffDoc struct {
		Requirements []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal([]byte(readFile(t, filepath.Join(out, ".engineering", "handoff.json"))), &handoffDoc); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range handoffDoc.Requirements {
		if r.ID == "REQ-001" && strings.Contains(r.Description, "monthly PDF export") {
			found = true
		}
	}
	if !found {
		t.Errorf("handoff requirements miss REQ-001: %+v", handoffDoc.Requirements)
	}

	// Architecture untouched: same components, same fingerprint, same recipe.
	manifest, err := project.ReadManifest(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Components) != len(manifestBefore.Components) {
		t.Errorf("components changed by eng add")
	}
	if manifest.PlanFingerprint != manifestBefore.PlanFingerprint {
		t.Errorf("plan fingerprint changed by eng add")
	}
	decision, err := ResolveProject([]byte(readFile(t, filepath.Join(out, ".engineering", "project-intent.json"))), overlay)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Selected == nil || decision.Selected.Recipe != "GP-06" {
		t.Errorf("re-resolution moved: %+v", decision.Selected)
	}
	assertDoctorGreen(t, out)
}

func TestAddRequirementVocabularyWarning(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	res, err := AddProjectRequirement(out, "native mobile app for field visits", "", overlay)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if res.Warning == "" || !strings.Contains(res.Warning, "eng surface add") {
		t.Errorf("expected architecture-vocabulary warning, got %q", res.Warning)
	}
	// Still recorded as a product requirement.
	intent := storedIntent(t, out)
	found := false
	for _, r := range intent.ProductRequirements {
		if r.ID == res.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("requirement %s was not recorded", res.ID)
	}
	assertDoctorGreen(t, out)
}

func TestAddRequirementPlannedLaterScope(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	res, err := AddProjectRequirement(out, "kiosk self-check-in next quarter", "planned_later", overlay)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	intent := storedIntent(t, out)
	planned := false
	for _, id := range intent.Scope.PlannedLater {
		if id == res.ID {
			planned = true
		}
	}
	if !planned {
		t.Errorf("planned_later scope did not track %s: %+v", res.ID, intent.Scope.PlannedLater)
	}
	if _, err := AddProjectRequirement(out, "another", "someday", overlay); err == nil {
		t.Errorf("expected unknown-scope error, got nil")
	}
	assertDoctorGreen(t, out)
}

func TestUpdateReport(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	if _, err := SurfaceAddProject(out, "public-intake", "", overlay); err != nil {
		t.Fatalf("surface add: %v", err)
	}
	manifestBefore := readFile(t, filepath.Join(out, ".engineering", "project.json"))

	t.Run("clean when pins match", func(t *testing.T) {
		report, err := UpdateProjectReport(out, overlay)
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if !report.Current {
			t.Errorf("report should be current: %+v", report.Entries)
		}
		if !strings.Contains(report.Summary, "all current") {
			t.Errorf("summary = %q, want all-current", report.Summary)
		}
		if len(report.Entries) != 3 {
			t.Errorf("entries = %d, want 3", len(report.Entries))
		}
		for _, e := range report.Entries {
			if e.UpdateAvailable {
				t.Errorf("entry %+v flagged without drift", e)
			}
		}
	})

	t.Run("stale pin listed with strategy, project untouched", func(t *testing.T) {
		stale := evolveOverlay(t)
		staleEntry := readFile(t, filepath.Join(stale, "boilerplates", "tanstack-admin.json"))
		staleEntry = strings.Replace(staleEntry, readPIN(t, mustAbs(t, "fixture-web")), "v9.9.9-stale", 1)
		if err := os.WriteFile(filepath.Join(stale, "boilerplates", "tanstack-admin.json"), []byte(staleEntry), 0o644); err != nil {
			t.Fatal(err)
		}
		report, err := UpdateProjectReport(out, stale)
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if report.Current {
			t.Fatalf("report should list the stale pin")
		}
		var admin *evolve.UpdateEntry
		for i := range report.Entries {
			if report.Entries[i].Component == "tanstack-admin" {
				admin = &report.Entries[i]
			}
		}
		if admin == nil {
			t.Fatalf("entries miss tanstack-admin: %+v", report.Entries)
		}
		if !admin.UpdateAvailable || admin.Catalog != "v9.9.9-stale" {
			t.Errorf("stale entry = %+v", *admin)
		}
		if admin.Strategy != "manual" {
			t.Errorf("default strategy = %q, want manual", admin.Strategy)
		}
		// Declared replace strategy on hono-api is reported even when clean.
		for _, e := range report.Entries {
			if e.Component == "hono-api" && e.Strategy != "replace" {
				t.Errorf("hono-api strategy = %q, want replace", e.Strategy)
			}
		}
		if got := readFile(t, filepath.Join(out, ".engineering", "project.json")); got != manifestBefore {
			t.Errorf("update report mutated the manifest")
		}
		assertDoctorGreen(t, out)
	})
}

func mustAbs(t *testing.T, fixture string) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", fixture))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestDoctorWarnsOnUnknownEvolutionSurface(t *testing.T) {
	overlay := evolveOverlay(t)
	out, _ := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	if err := project.AppendEvent(out, project.EvolutionEvent{
		Type:    project.EventSurfaceAdd,
		Surface: "warp-drive",
	}, time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	findings, err := DoctorProject(out)
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	if project.HasErrors(findings) {
		t.Fatalf("forward-compatible event must not error: %+v", findings)
	}
	found := false
	for _, f := range findings {
		if f.Code == "evolution-event-unknown-surface" && f.Severity == project.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Errorf("expected evolution-event-unknown-surface warning: %+v", findings)
	}
}
