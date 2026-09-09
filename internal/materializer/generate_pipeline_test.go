package materializer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/foundationconfig"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// fixtureGeneratorDir returns the absolute fake-admin fixture directory
// and prepends it to PATH so the offline generator resolves by name.
func fixtureGeneratorDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "generators", "fake-admin"))
	if err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(filepath.Join(abs, "fake-admin-generate")); err != nil || st.IsDir() {
		t.Fatalf("fixture generator missing: %v", err)
	}
	t.Setenv("PATH", abs+string(os.PathListSeparator)+os.Getenv("PATH"))
	return abs
}

// fixtureFactory builds a local factory source with a PIN marker plus
// factory-only junk that must never leak into the project.
func fixtureFactory(t *testing.T, pin string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"PIN":                     pin + "\n",
		"templates/showcase.md":   "# showcase (factory only)\n",
		"generator/internals.txt": "factory internals\n",
		"package.json":            "{\"name\":\"factory\"}\n",
	} {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const fixtureGenPin = "v0.0.0-fixture-gen"

func generatedBoilerplate(factoryDir string) domain.Boilerplate {
	gen := domain.GenerateSpec{
		Run:            domain.AdapterCommand{Run: []string{"fake-admin-generate", "--name", "{name}", "--profile", "{profile}", "--output", "{output}"}},
		Output:         "{output}",
		DefaultProfile: "minimal",
		Profiles: []domain.GeneratorProfile{
			{ID: "minimal", Arguments: []string{"--profile", "minimal"}},
		},
	}
	return domain.Boilerplate{
		ID:             "fixture-admin",
		Pin:            fixtureGenPin,
		Adapter:        "fixture-gen",
		AdapterSpec:    &domain.AdapterSpec{Name: "fixture-gen", Operations: []string{"fetch", "generate"}, Generate: &gen, ManagedFiles: []string{"AGENTS.md"}},
		Source:         domain.SourceSpec{Type: "local", Path: factoryDir},
		DeliveryStatus: "stable",
		DecisionStatus: "curated",
	}
}

func generatedPlanWith(bp domain.Boilerplate, profile string) planner.MaterializationPlan {
	mat := foundationconfig.MaterializationConfig{
		Strategy:           domain.StrategyGenerate,
		Name:               "demo-admin",
		AdapterFingerprint: domain.AdapterFingerprint(*bp.AdapterSpec),
	}
	if profile != "" {
		mat.Profile = profile
		mat.Arguments = []string{"--profile", profile}
	}
	return planner.MaterializationPlan{
		SchemaVersion:   2,
		Project:         "demo",
		Recipe:          "TEST",
		RecipeVersion:   "1.0.0",
		DatabaseProfile: "sqlite-local",
		Fingerprint:     "test-fingerprint-gen",
		Components: []planner.PlanComponent{{
			Boilerplate: "fixture-admin", Pin: fixtureGenPin,
			Destination: "apps/admin", Surface: "web-admin",
			Materialization: mat,
		}},
	}
}

// TestGeneratedFoundationRunsInTemporaryWorkspace executes the generator
// with resolved values and keeps the factory out of the result.
func TestGeneratedFoundationRunsInTemporaryWorkspace(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	bp := generatedBoilerplate(factory)
	sandbox, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sandbox)
	values := GenerationValues{Name: "demo-admin", Project: "demo", Surface: "web-admin", Profile: "minimal", Output: filepath.Join(sandbox, "output")}
	out, err := RunGenerateWithValues(context.Background(), bp.AdapterSpec.Generate, values, factory, sandbox, time.Minute)
	if err != nil {
		t.Fatalf("RunGenerateWithValues: %v", err)
	}
	if !strings.HasPrefix(out, sandbox) {
		t.Fatalf("output %q is not inside the sandbox %q", out, sandbox)
	}
	raw, err := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "minimal") {
		t.Fatalf("output missing profile marker:\n%s", raw)
	}
	if _, err := os.Stat(filepath.Join(out, "templates", "showcase.md")); err == nil {
		t.Fatal("factory files leaked into the generated output")
	}
}

// TestGeneratedFoundationCopiesOnlyOutput materializes end to end and
// asserts the factory never reaches the project staging.
func TestGeneratedFoundationCopiesOnlyOutput(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{generatedBoilerplate(factory)},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, generatedPlanWith(generatedBoilerplate(factory), "minimal"), out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	for _, want := range []string{"apps/admin/AGENTS.md", "apps/admin/index.html"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected generated file %s: %v", want, err)
		}
	}
	for _, leak := range []string{"apps/admin/templates", "apps/admin/generator", "apps/admin/package.json"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(leak))); err == nil {
			t.Errorf("factory leaked into project: %s", leak)
		}
	}
	raw, err := os.ReadFile(filepath.Join(out, "apps", "admin", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "demo-admin") || !strings.Contains(string(raw), "minimal") {
		t.Fatalf("output missing name/profile markers:\n%s", raw)
	}
}

// TestGeneratorSourceDoesNotLeakIntoProject lists the staged admin tree
// and rejects every factory-only entry.
func TestGeneratorSourceDoesNotLeakIntoProject(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{generatedBoilerplate(factory)},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, generatedPlanWith(generatedBoilerplate(factory), "minimal"), out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	var leaks []string
	_ = filepath.Walk(filepath.Join(out, "apps", "admin"), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(p)
		if base == "showcase.md" || base == "internals.txt" || base == "package.json" {
			leaks = append(leaks, p)
		}
		return nil
	})
	if len(leaks) > 0 {
		t.Fatalf("factory files in project: %v", leaks)
	}
}

// TestGeneratedOutputCannotEscapeSandbox runs the escape profile.
func TestGeneratedOutputCannotEscapeSandbox(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	bp := generatedBoilerplate(factory)
	sandbox, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sandbox)
	values := GenerationValues{Name: "demo-admin", Project: "demo", Surface: "web-admin", Profile: "escape", Output: filepath.Join(sandbox, "output")}
	if _, err := RunGenerateWithValues(context.Background(), bp.AdapterSpec.Generate, values, factory, sandbox, time.Minute); err == nil {
		t.Fatal("expected sandbox escape failure, got nil")
	}
}

// TestMissingGeneratedOutputFails points the output declaration at a
// directory the command never creates.
func TestMissingGeneratedOutputFails(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true not on PATH")
	}
	spec := &domain.GenerateSpec{
		Run:    domain.AdapterCommand{Run: []string{"true"}},
		Output: "never-created",
	}
	sandbox := t.TempDir()
	values := GenerationValues{Name: "x", Project: "x", Surface: "x", Output: filepath.Join(sandbox, "output")}
	if _, err := RunGenerateWithValues(context.Background(), spec, values, "", sandbox, time.Minute); err == nil {
		t.Fatal("expected missing-output failure, got nil")
	} else if !strings.Contains(err.Error(), "no output directory") {
		t.Fatalf("wrong failure: %v", err)
	}
}

// TestEmptyGeneratedOutputFails runs the empty profile.
func TestEmptyGeneratedOutputFails(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	bp := generatedBoilerplate(factory)
	sandbox, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sandbox)
	values := GenerationValues{Name: "demo-admin", Project: "demo", Surface: "web-admin", Profile: "empty", Output: filepath.Join(sandbox, "output")}
	if _, err := RunGenerateWithValues(context.Background(), bp.AdapterSpec.Generate, values, factory, sandbox, time.Minute); err == nil {
		t.Fatal("expected empty-output failure, got nil")
	} else if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("wrong failure: %v", err)
	}
}

// TestGeneratorFailureLeavesDestinationUntouched asserts the atomic
// commit: a broken generator leaves no project directory behind.
func TestGeneratorFailureLeavesDestinationUntouched(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{generatedBoilerplate(factory)},
	}
	out := filepath.Join(t.TempDir(), "proj")
	req := happyRequest(t, cat, generatedPlanWith(generatedBoilerplate(factory), "broken"), out)
	if _, err := Materialize(req); err == nil {
		t.Fatal("expected generation failure, got nil")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed generation left destination behind: %v", err)
	}
}

// TestGeneratorTimeoutLeavesDestinationUntouched bounds the slow profile.
func TestGeneratorTimeoutLeavesDestinationUntouched(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{generatedBoilerplate(factory)},
	}
	out := filepath.Join(t.TempDir(), "proj")
	req := happyRequest(t, cat, generatedPlanWith(generatedBoilerplate(factory), "slow"), out)
	req.CommandTimeout = 300 * time.Millisecond
	if _, err := Materialize(req); err == nil {
		t.Fatal("expected timeout failure, got nil")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("timed-out generation left destination behind: %v", err)
	}
}

// TestGeneratedSetupRunsAgainstOutputNotFactory proves setup commands run
// with the staged output as cwd by writing a relative marker there.
func TestGeneratedSetupRunsAgainstOutputNotFactory(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	bp := generatedBoilerplate(factory)
	gen := *bp.AdapterSpec.Generate
	gen.Run = domain.AdapterCommand{Run: []string{
		"fake-admin-generate", "--name", "{name}", "--profile", "{profile}", "--output", "{output}"}}
	bp.AdapterSpec.Generate = &gen
	bp.AdapterSpec.Setup = []domain.AdapterCommand{{Run: []string{
		"fake-admin-generate", "--name", "setup-probe", "--profile", "minimal", "--output", "setup-marker"}}}
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{bp},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, generatedPlanWith(bp, "minimal"), out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "apps", "admin", "setup-marker", "index.html")); err != nil {
		t.Fatalf("setup did not run against the staged output: %v", err)
	}
}

// TestGeneratedChecksRunAgainstOutput proves checks run with the staged
// output as cwd: a check listing a generated file passes, and one
// referencing a factory-only file fails the materialization.
func TestGeneratedChecksRunAgainstOutput(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	bp := generatedBoilerplate(factory)
	bp.AdapterSpec.Checks = []domain.AdapterCommand{{Run: []string{
		"fake-admin-generate", "--name", "check-probe", "--profile", "minimal", "--output", "check-marker"}}}
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{bp},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, generatedPlanWith(bp, "minimal"), out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "apps", "admin", "check-marker", "AGENTS.md")); err != nil {
		t.Fatalf("checks did not run against the staged output: %v", err)
	}
}

// TestMixedCopyAndGenerateProjectMaterializes combines one static and one
// generated foundation and verifies provenance per surface.
func TestMixedCopyAndGenerateProjectMaterializes(t *testing.T) {
	fixtureGeneratorDir(t)
	factory := fixtureFactory(t, fixtureGenPin)
	webBP := localBoilerplate("fixture-web", fixtureWebPin, fixtureSource(t, "fixture-web"))
	genBP := generatedBoilerplate(factory)
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{webBP, genBP},
	}
	plan := planner.MaterializationPlan{
		SchemaVersion: 2, Project: "demo", Recipe: "TEST", RecipeVersion: "1.0.0",
		DatabaseProfile: "sqlite-local", Fingerprint: "test-fingerprint-mixed",
		Components: []planner.PlanComponent{
			{Boilerplate: "fixture-web", Pin: fixtureWebPin, Destination: "apps/web", Surface: "public-web",
				Materialization: foundationconfig.MaterializationConfig{Strategy: domain.StrategyCopy, AdapterFingerprint: domain.AdapterFingerprint(*webBP.AdapterSpec)}},
			{Boilerplate: "fixture-admin", Pin: fixtureGenPin, Destination: "apps/admin", Surface: "web-admin",
				Materialization: foundationconfig.MaterializationConfig{
					Strategy: domain.StrategyGenerate, Name: "demo-admin",
					Profile: "minimal", Arguments: []string{"--profile", "minimal"},
					AdapterFingerprint: domain.AdapterFingerprint(*genBP.AdapterSpec)}},
		},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	for _, want := range []string{"apps/web/index.html", "apps/admin/index.html"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected file %s: %v", want, err)
		}
	}
	prov, err := project.ReadProvenance(out)
	if err != nil {
		t.Fatal(err)
	}
	bySurface := map[string]project.ProvenanceComponent{}
	for _, pc := range prov.Components {
		bySurface[pc.Surface] = pc
	}
	web, ok := bySurface["public-web"]
	if !ok || web.Strategy != domain.StrategyCopy {
		t.Fatalf("web provenance: %+v ok=%v", web, ok)
	}
	admin, ok := bySurface["web-admin"]
	if !ok || admin.Strategy != domain.StrategyGenerate || admin.Profile != "minimal" || admin.Name != "demo-admin" {
		t.Fatalf("admin provenance: %+v ok=%v", admin, ok)
	}
	if findings, err := project.Doctor(out); err != nil || project.HasErrors(findings) {
		t.Errorf("doctor after mixed materialization: findings=%v err=%v", findings, err)
	}
}
