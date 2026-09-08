package catalog_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

// repoRoot walks up from the working directory to the repo root (go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root (go.mod) not found")
		}
		dir = parent
	}
}

// phase9Boilerplates pins the Fase 9 migration contract: every new foundation
// carries a real repo URL, a verified upstream pin, an adapter object,
// provided surfaces and technology tags, plus a curation evidence stub.
var phase9Boilerplates = []struct {
	id   string
	repo string
	pin  string
}{
	{"ignite", "https://github.com/infinitered/ignite", "e829d2f922c5568a59a77bfb6232aeb500be3f13"},
	{"tauri-ui", "https://github.com/agmmnn/tauri-ui", "8eb86d894c19b6df04ff883ab28b412b1e5f23ea"},
	{"speedpy", "https://github.com/speedpy/speedpy", "3fbf725d8e9cf6b8aadb3aeaf1db2822522282b9"},
	{"react-starter-kit", "https://github.com/kriasoft/react-starter-kit", "0aa7603435f16159ad0b8fef68fb7f6280be7ca1"},
}

func TestPhase9BoilerplateContracts(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	idx := catalog.NewIndex(cat)
	root := repoRoot(t)
	for _, tt := range phase9Boilerplates {
		t.Run(tt.id, func(t *testing.T) {
			bp, ok := idx.Boilerplate(tt.id)
			if !ok {
				t.Fatalf("%s missing from catalog", tt.id)
			}
			if bp.EffectiveRepo() != tt.repo {
				t.Fatalf("repo = %q, want %q", bp.EffectiveRepo(), tt.repo)
			}
			if bp.Pin != tt.pin {
				t.Fatalf("pin = %q, want verified pin %q", bp.Pin, tt.pin)
			}
			if err := bp.Validate(); err != nil {
				t.Fatalf("boilerplate invalid: %v", err)
			}
			if bp.Adapter == "" {
				t.Fatal("boilerplate declares no adapter")
			}
			if len(bp.Provides.Surfaces) == 0 {
				t.Fatal("boilerplate provides no surface")
			}
			if len(bp.TechTags) == 0 {
				t.Fatal("boilerplate declares no tech tags")
			}
			if !composer.Eligible(bp) {
				t.Fatalf("boilerplate %s is not an eligible composition provider (decision=%q delivery=%q)",
					tt.id, bp.DecisionStatus, bp.DeliveryStatus)
			}
			stub := filepath.Join(root, "catalog", "curation", tt.id+".md")
			raw, err := os.ReadFile(stub)
			if err != nil {
				t.Fatalf("curation evidence stub missing: %v", err)
			}
			if len(raw) == 0 {
				t.Fatal("curation evidence stub is empty")
			}
		})
	}
}

// phase9Participation ties every new foundation to the decisions it serves:
// the linked routing fixture must resolve and the linked composition fixture
// must place the boilerplate as a provider.
var phase9Participation = []struct {
	boilerplate string
	routing     string
	composition string
}{
	{"ignite", "mobile-field-app.json", "mobile-only.json"},
	{"tauri-ui", "desktop-local-tool.json", "desktop-only.json"},
	{"speedpy", "python-reporting-tool.json", "python-job.json"},
	{"react-starter-kit", "saas-plus-marketing.json", "saas-marketing.json"},
}

func TestPhase9FixtureParticipation(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	root := repoRoot(t)
	for _, tt := range phase9Participation {
		t.Run(tt.boilerplate, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, "testdata", "routing", tt.routing))
			if err != nil {
				t.Fatalf("read routing %s: %v", tt.routing, err)
			}
			var wrapper struct {
				Intent domain.ProjectIntent `json:"intent"`
			}
			if err := json.Unmarshal(raw, &wrapper); err != nil {
				t.Fatalf("parse routing %s: %v", tt.routing, err)
			}
			decision := resolver.Resolve(wrapper.Intent, cat)
			if decision.Status != domain.StatusResolved {
				t.Fatalf("routing %s did not resolve: %s", tt.routing, resolver.Explain(decision))
			}
			comp, err := composer.Compose(decision, cat)
			if err != nil {
				t.Fatalf("compose: %v", err)
			}
			for _, c := range comp.Components {
				if c.Boilerplate == tt.boilerplate {
					return
				}
			}
			t.Fatalf("boilerplate %s serves no component in %+v (fixture %s)",
				tt.boilerplate, comp.Components, tt.composition)
		})
	}
}

// TestPhase9RecipesActive guards the four migrated recipes: active status,
// primary boilerplates present in the catalog, at least one allowed surface
// composition, and a database policy with a known default profile.
func TestPhase9RecipesActive(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	idx := catalog.NewIndex(cat)
	for _, id := range []string{"GP-03", "GP-04", "GP-05", "GP-07"} {
		t.Run(id, func(t *testing.T) {
			r, ok := idx.Recipe(id)
			if !ok {
				t.Fatalf("recipe %s missing from catalog", id)
			}
			if !r.Active() {
				t.Fatalf("recipe %s status %q is not active", id, r.Status)
			}
			if len(r.PrimaryBoilerplates) == 0 {
				t.Fatalf("recipe %s declares no primary boilerplates", id)
			}
			for _, b := range r.PrimaryBoilerplates {
				if _, ok := idx.Boilerplate(b); !ok {
					t.Fatalf("recipe %s references unknown boilerplate %q", id, b)
				}
			}
			if len(r.AllowedSurfaceComposition) == 0 {
				t.Fatalf("recipe %s allows no surface composition", id)
			}
			if r.DatabasePolicy.DefaultProfile == "" {
				t.Fatalf("recipe %s declares no default database profile", id)
			}
		})
	}
}
