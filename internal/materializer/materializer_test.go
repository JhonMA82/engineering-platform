package materializer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
)

const (
	fixtureWebPin = "v0.0.0-fixture-web"
	fixtureAPIPin = "v0.0.0-fixture-api"
)

func fixtureSource(t *testing.T, name string) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", name))
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		t.Fatalf("fixture source %s missing: %v", name, err)
	}
	return abs
}

func localBoilerplate(id, pin, src string) domain.Boilerplate {
	return domain.Boilerplate{
		ID:             id,
		Pin:            pin,
		Adapter:        "fixture",
		AdapterSpec:    &domain.AdapterSpec{Name: "fixture", Operations: []string{"fetch", "copy"}},
		Source:         domain.SourceSpec{Type: "local", Path: src},
		DeliveryStatus: "stable",
		DecisionStatus: "curated",
	}
}

func twoComponentCatalog(t *testing.T) catalog.Catalog {
	t.Helper()
	return catalog.Catalog{
		CatalogVersion: "0.0.0-test",
		MinCoreVersion: "1.0.0",
		SchemaVersion:  1,
		Boilerplates: []domain.Boilerplate{
			localBoilerplate("fixture-web", fixtureWebPin, fixtureSource(t, "fixture-web")),
			localBoilerplate("fixture-api", fixtureAPIPin, fixtureSource(t, "fixture-api")),
		},
	}
}

func twoComponentPlan() planner.MaterializationPlan {
	return planner.MaterializationPlan{
		SchemaVersion:   1,
		Project:         "test-proj",
		Recipe:          "TEST",
		RecipeVersion:   "1.0.0",
		DatabaseProfile: "sqlite-local",
		Fingerprint:     "test-fingerprint-1",
		Components: []planner.PlanComponent{
			{Boilerplate: "fixture-web", Pin: fixtureWebPin, Destination: "apps/web", Surface: "public-web"},
			{Boilerplate: "fixture-api", Pin: fixtureAPIPin, Destination: "services/api", Surface: "api"},
		},
	}
}

func happyRequest(t *testing.T, cat catalog.Catalog, plan planner.MaterializationPlan, output string) Request {
	t.Helper()
	return Request{
		Plan:         plan,
		Catalog:      cat,
		OutputDir:    output,
		IntentJSON:   []byte(`{"schema_version":1,"name":"test-proj","problem":"offline test","product_requirements":[{"id":"records","description":"manage records"}]}`),
		DecisionJSON: []byte(`{"schema_version":1,"status":"resolved","intent_fingerprint":"intent-abc"}`),
		CoreVersion:  "0.0.0-test",
	}
}

func TestMaterializeHappyPath(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	out := filepath.Join(t.TempDir(), "proj")
	res, err := Materialize(happyRequest(t, cat, plan, out))
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if res.ProjectDir != out {
		t.Fatalf("project dir = %q, want %q", res.ProjectDir, out)
	}
	for _, want := range []string{
		"apps/web/index.html",
		"services/api/server.js",
		".engineering/project.json",
		".engineering/provenance.json",
		".engineering/project-map.json",
		".engineering/project-intent.json",
		".engineering/architecture-decision.json",
		".engineering/materialization-plan.json",
		".engineering/implementation-brief.md",
		".engineering/handoff.json",
		"AGENTS.md",
		"ARCHITECTURE.md",
		"GENTLE.md",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected file %s: %v", want, err)
		}
	}
	// Foundation-owned surface instructions are preserved, not overwritten.
	raw, err := os.ReadFile(filepath.Join(out, "apps", "web", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Fixture Web") {
		t.Errorf("surface AGENTS.md was not preserved:\n%s", raw)
	}
	router, err := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"| api | services/api |", "| public-web | apps/web |"} {
		if !strings.Contains(string(router), want) {
			t.Errorf("router table missing %q:\n%s", want, router)
		}
	}
	prov, err := project.ReadProvenance(out)
	if err != nil {
		t.Fatal(err)
	}
	if prov.CoreVersion != "0.0.0-test" || prov.IntentFingerprint != "intent-abc" || prov.MaterializedAt == "" {
		t.Errorf("unexpected provenance: %+v", prov)
	}
	if findings, err := project.Doctor(out); err != nil || project.HasErrors(findings) {
		t.Errorf("doctor after happy path: findings=%v err=%v", findings, err)
	}
}

func TestMaterializeRejectsUnsafeDestination(t *testing.T) {
	cases := []struct {
		name string
		dest string
	}{
		{"parent traversal", "../evil"},
		{"nested traversal", "apps/../../evil"},
		{"absolute path", "/tmp/evil"},
		{"empty destination", ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cat := twoComponentCatalog(t)
			plan := twoComponentPlan()
			plan.Components[0].Destination = tt.dest
			out := filepath.Join(t.TempDir(), "proj")
			if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
				t.Fatalf("expected rejection of destination %q", tt.dest)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Errorf("output dir must stay untouched, stat=%v", statErr)
			}
		})
	}
}

func TestMaterializeRejectsCollision(t *testing.T) {
	cases := []struct {
		name   string
		second string
	}{
		{"duplicate destination", "apps/web"},
		{"nested destination", "apps/web/nested"},
		{"parent destination", "apps"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cat := twoComponentCatalog(t)
			plan := twoComponentPlan()
			plan.Components[1].Boilerplate = "fixture-web"
			plan.Components[1].Pin = fixtureWebPin
			plan.Components[1].Surface = "public-web-extra"
			plan.Components[1].Destination = tt.second
			out := filepath.Join(t.TempDir(), "proj")
			if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
				t.Fatalf("expected collision rejection for %q", tt.second)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Errorf("output dir must stay untouched, stat=%v", statErr)
			}
		})
	}
}

func TestMaterializeRejectsSymlinkEscape(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "PIN"), []byte(fixtureWebPin+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "ok.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(src, "evil-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test",
		Boilerplates:   []domain.Boilerplate{localBoilerplate("fixture-web", fixtureWebPin, src)},
	}
	plan := twoComponentPlan()
	plan.Components = plan.Components[:1]
	out := filepath.Join(t.TempDir(), "proj")
	_, err := Materialize(happyRequest(t, cat, plan, out))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink-escape rejection, got %v", err)
	}
}

func TestMaterializeRejectsMaliciousAdapter(t *testing.T) {
	cases := []struct {
		name string
		cmd  domain.AdapterCommand
	}{
		{"shell interpreter", domain.AdapterCommand{Run: []string{"sh", "-c", "rm -rf /"}}},
		{"metachar argument", domain.AdapterCommand{Run: []string{"npm", "install; rm -rf /"}}},
		{"path command", domain.AdapterCommand{Run: []string{"/tmp/evil", "go"}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cat := twoComponentCatalog(t)
			bp := cat.Boilerplates[0]
			spec := *bp.AdapterSpec
			spec.Setup = []domain.AdapterCommand{tt.cmd}
			cat.Boilerplates[0].AdapterSpec = &spec
			plan := twoComponentPlan()
			plan.Components = plan.Components[:1]
			out := filepath.Join(t.TempDir(), "proj")
			if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
				t.Fatalf("expected command rejection for %v", tt.cmd.Run)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Errorf("output dir must stay untouched, stat=%v", statErr)
			}
		})
	}
}

func TestMaterializeRejectsPinMismatch(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	plan.Components[0].Pin = "v9.9.9-drifted"
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
		t.Fatal("expected pin mismatch rejection")
	}
}

func TestMaterializeRequiresEmptyOutputDir(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	parent := t.TempDir()
	out := filepath.Join(parent, "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("first materialize: %v", err)
	}
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
		t.Fatal("expected re-run on the same directory to fail")
	} else if !strings.Contains(err.Error(), "non-empty") {
		t.Fatalf("expected non-empty complaint, got %v", err)
	}
	// The committed project still passes doctor after the refused re-run.
	if findings, err := project.Doctor(out); err != nil || project.HasErrors(findings) {
		t.Errorf("doctor after refused re-run: findings=%v err=%v", findings, err)
	}
}

func TestValidateCommands(t *testing.T) {
	cases := []struct {
		name    string
		cmds    []domain.AdapterCommand
		wantErr bool
	}{
		{"honest argv", []domain.AdapterCommand{{Run: []string{"npm", "run", "build"}}}, false},
		{"no commands", nil, false},
		{"empty run", []domain.AdapterCommand{{Run: nil}}, true},
		{"shell", []domain.AdapterCommand{{Run: []string{"bash", "setup.sh"}}}, true},
		{"substitution", []domain.AdapterCommand{{Run: []string{"go", "generate", "$(evil)"}}}, true},
		{"redirection", []domain.AdapterCommand{{Run: []string{"npm", "test", ">", "out"}}}, true},
		{"path binary", []domain.AdapterCommand{{Run: []string{"./setup.sh"}}}, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCommands(tt.cmds); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommands(%v) err=%v wantErr=%v", tt.cmds, err, tt.wantErr)
			}
		})
	}
}
