// Architecture regression tests (§7): the PRD fundamentals that must hold
// at v1.0. Fixture-driven scenarios already live in testdata/routing and run
// inside TestRoutingScenarios; this file pins the invariants that fixtures
// alone cannot express: routing invariance under product questions, the
// absence of project_type, the tui catalog-gap, open-ended surfaces via a
// synthetic overlay, and resolver-core import purity.
package resolver

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// TestOpenProductQuestionsDoNotAffectRouting proves §5.2: product questions
// travel intent→handoff but never change the routing decision.
func TestOpenProductQuestionsDoNotAffectRouting(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	base := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "backoffice",
		Problem:       "operators manage records",
		Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
	}
	withQuestions := base
	withQuestions.OpenProductQuestions = []string{
		"define approval roles for record publishing",
		"define cancellation behavior for in-flight edits",
	}
	a := Resolve(base, cat)
	b := Resolve(withQuestions, cat)
	if string(a.Status) != string(b.Status) {
		t.Fatalf("status changed with product questions: %q vs %q", a.Status, b.Status)
	}
	if (a.Selected == nil) != (b.Selected == nil) ||
		(a.Selected != nil && a.Selected.Recipe != b.Selected.Recipe) {
		t.Fatalf("selection changed with product questions: %+v vs %+v", a.Selected, b.Selected)
	}
	ra, _ := json.Marshal(a.Reasons)
	rb, _ := json.Marshal(b.Reasons)
	if string(ra) != string(rb) {
		t.Fatalf("reasons changed with product questions:\n%s\n%s", ra, rb)
	}
	if string(b.Status) != string(domain.StatusResolved) {
		t.Fatalf("expected resolved, got %q", b.Status)
	}
}

// TestIntentHasNoProjectType pins §7.1 at compile-time level: ProjectIntent
// carries no authorizing project_type field or JSON tag, and routing resolves
// with zero project_type signal.
func TestIntentHasNoProjectType(t *testing.T) {
	rt := reflect.TypeOf(domain.ProjectIntent{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if strings.Contains(strings.ToLower(f.Name), "project_type") ||
			strings.Contains(strings.ToLower(f.Tag.Get("json")), "project_type") {
			t.Fatalf("ProjectIntent must not carry project_type: field %s tag %q", f.Name, f.Tag.Get("json"))
		}
	}
	var raw map[string]any
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "backoffice",
		Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
	}
	encoded, _ := json.Marshal(intent)
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["project_type"]; ok {
		t.Fatal("serialized intent must not contain project_type")
	}
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if got := Resolve(intent, cat); string(got.Status) != string(domain.StatusResolved) {
		t.Fatalf("routing without project_type = %q, want resolved", got.Status)
	}
}

// TestTuiWithoutProviderIsCatalogGap pins §7.3: a required tui surface with
// no curated provider in the catalog is an architectural catalog-gap (while
// missing product features such as PDF/Excel never are — see
// pdf-excel-no-gap.json, asserted in TestRoutingScenarios).
func TestTuiWithoutProviderIsCatalogGap(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if got := len(catalog.NewIndex(cat).ProvidersForSurface("tui")); got != 0 {
		t.Fatalf("precondition: expected zero tui providers, have %d", got)
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "terminal tool",
		Problem:       "operators need a terminal app",
		Surfaces:      []domain.SurfaceIntent{{Kind: "tui", Scope: domain.ScopeRequiredNow}},
	}
	got := Resolve(intent, cat)
	if string(got.Status) != string(domain.StatusCatalogGap) {
		t.Fatalf("status = %q, want catalog-gap\n%s", got.Status, Explain(got))
	}
	if !strings.Contains(strings.Join(got.Reasons, "\n"), "surface=tui") {
		t.Fatalf("gap reasons lack surface=tui\n%s", Explain(got))
	}
}

// TestSyntheticSurfaceResolvesWithoutCoreChanges pins §7.4: a fixture-only
// overlay (new surface + provider + recipe) loads and resolves, proving no
// closed surface enum hides in the core.
func TestSyntheticSurfaceResolvesWithoutCoreChanges(t *testing.T) {
	base, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	dir := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("metadata.json", `{"catalog_version":"test","min_core_version":"0.0.0","schema_version":1}`)
	write("surfaces/test-surface.json", `{"id":"test-surface","description":"Synthetic regression surface."}`)
	write("boilerplates/test-provider.json", `{
      "id": "test-provider",
      "repo": "https://example.com/test-provider",
      "pin": "abc123",
      "adapter": {"name": "test-provider", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
      "provides": {"surfaces": ["test-surface"], "capabilities": []},
      "tech_tags": ["test"]
    }`)
	write("recipes/GP-99-test.json", `{
      "id": "GP-99",
      "version": "0.0.1",
      "status": "stable",
      "description": "Synthetic regression recipe.",
      "provides": {"surfaces": ["test-surface"], "capabilities": []},
      "primary_boilerplates": ["test-provider"],
      "allowed_surface_composition": [["test-surface"]]
    }`)
	overlay, err := catalog.LoadDir(dir)
	if err != nil {
		t.Fatalf("load overlay: %v", err)
	}
	merged := catalog.MergeOverlay(base, overlay)
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "synthetic",
		Problem:       "prove the surface vocabulary is open",
		Surfaces:      []domain.SurfaceIntent{{Kind: "test-surface", Scope: domain.ScopeRequiredNow}},
	}
	got := Resolve(intent, merged)
	if string(got.Status) != string(domain.StatusResolved) {
		t.Fatalf("status = %q, want resolved\n%s", got.Status, Explain(got))
	}
	if got.Selected == nil || got.Selected.Recipe != "GP-99" {
		t.Fatalf("selected = %+v, want GP-99", got.Selected)
	}
}

// TestCoreImportPurity pins §7.7: the pure pipeline packages (resolver,
// composer, planner, domain) never import filesystem, process execution or
// network directly. Transitive stdlib deps (e.g. via path/filepath) are out
// of scope; only direct imports are asserted, parsed with stdlib go/parser
// so the test itself needs no external tool.
func TestCoreImportPurity(t *testing.T) {
	forbidden := map[string]bool{
		"os": true, "os/exec": true,
		"net": true, "net/http": true, "net/url": true,
		"syscall": true, "os/signal": true,
	}
	pkgs := []string{"resolver", "composer", "planner", "domain"}
	fset := token.NewFileSet()
	for _, pkg := range pkgs {
		dir := filepath.Join("..", pkg)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s/%s: %v", pkg, name, err)
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if forbidden[path] {
					t.Errorf("%s/%s directly imports %q: pure core must not", pkg, name, path)
				}
			}
		}
	}
}
