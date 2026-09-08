package composer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

type compositionFixture struct {
	Name         string                `json:"name"`
	IntentFile   string                `json:"intent_file,omitempty"`
	Intent       *domain.ProjectIntent `json:"intent,omitempty"`
	Destinations map[string]string     `json:"destinations,omitempty"`
	Expect       struct {
		Recipe          string `json:"recipe"`
		DatabaseProfile string `json:"database_profile,omitempty"`
		Components      []struct {
			Surface     string `json:"surface"`
			Provider    string `json:"provider,omitempty"`
			Destination string `json:"destination,omitempty"`
		} `json:"components"`
		MustFail bool `json:"must_fail,omitempty"`
	} `json:"expect"`
}

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

func loadCompositionFixtures(t *testing.T) []compositionFixture {
	t.Helper()
	root := repoRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "testdata", "composition", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no composition fixtures: %v", err)
	}
	var out []compositionFixture
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var fx compositionFixture
		if err := json.Unmarshal(raw, &fx); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		out = append(out, fx)
	}
	return out
}

// fixtureIntent resolves the inline intent or the referenced intent file.
// Referenced files may be raw intents or routing wrappers carrying .intent.
func fixtureIntent(t *testing.T, fx compositionFixture) domain.ProjectIntent {
	t.Helper()
	if fx.Intent != nil {
		return *fx.Intent
	}
	if fx.IntentFile == "" {
		t.Fatalf("fixture %s: neither intent nor intent_file", fx.Name)
	}
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), fx.IntentFile))
	if err != nil {
		t.Fatalf("fixture %s: read %s: %v", fx.Name, fx.IntentFile, err)
	}
	var wrapper struct {
		Intent *domain.ProjectIntent `json:"intent"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("fixture %s: parse %s: %v", fx.Name, fx.IntentFile, err)
	}
	if wrapper.Intent != nil {
		return *wrapper.Intent
	}
	var intent domain.ProjectIntent
	if err := json.Unmarshal(raw, &intent); err != nil {
		t.Fatalf("fixture %s: parse raw intent %s: %v", fx.Name, fx.IntentFile, err)
	}
	return intent
}

func TestCompositionFixtures(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	for _, fx := range loadCompositionFixtures(t) {
		t.Run(fx.Name, func(t *testing.T) {
			intent := fixtureIntent(t, fx)
			decision := resolver.Resolve(intent, cat)
			if decision.Status != domain.StatusResolved {
				if fx.Expect.MustFail {
					t.Skip("intent does not resolve; negative path covered at destination level")
				}
				t.Fatalf("intent did not resolve: %s", resolver.Explain(decision))
			}
			overrides := map[domain.SurfaceID]string{}
			for s, d := range fx.Destinations {
				overrides[domain.SurfaceID(s)] = d
			}
			comp, err := ComposeWithDestinations(decision, cat, overrides)
			if fx.Expect.MustFail {
				if err == nil {
					t.Fatalf("expected composition failure, got %+v", comp)
				}
				var derr *domain.Error
				if !asDomainError(err, &derr) || derr.Class != domain.ClassComposition {
					t.Fatalf("expected typed composition error, got %T: %v", err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("compose: %v", err)
			}
			if comp.Recipe != fx.Expect.Recipe {
				t.Fatalf("recipe = %q, want %q", comp.Recipe, fx.Expect.Recipe)
			}
			if fx.Expect.DatabaseProfile != "" && comp.DatabaseProfile != fx.Expect.DatabaseProfile {
				t.Fatalf("database_profile = %q, want %q", comp.DatabaseProfile, fx.Expect.DatabaseProfile)
			}
			if len(comp.Components) != len(fx.Expect.Components) {
				t.Fatalf("components = %+v, want %d components", comp.Components, len(fx.Expect.Components))
			}
			bySurface := map[string]Component{}
			for _, c := range comp.Components {
				bySurface[string(c.Surface)] = c
			}
			for _, want := range fx.Expect.Components {
				got, ok := bySurface[want.Surface]
				if !ok {
					t.Fatalf("missing component for surface %q in %+v", want.Surface, comp.Components)
				}
				if want.Provider != "" && got.Boilerplate != want.Provider {
					t.Fatalf("surface %q provider = %q, want %q", want.Surface, got.Boilerplate, want.Provider)
				}
				if want.Destination != "" && got.Destination != want.Destination {
					t.Fatalf("surface %q destination = %q, want %q", want.Surface, got.Destination, want.Destination)
				}
				if got.Pin == "" {
					t.Fatalf("surface %q component lacks pin", want.Surface)
				}
			}
			// Product features must never add components: recompose must be
			// byte-identical (no hidden feature-driven expansion).
			again, err := ComposeWithDestinations(decision, cat, overrides)
			if err != nil {
				t.Fatalf("recompose: %v", err)
			}
			a, _ := json.Marshal(comp)
			b, _ := json.Marshal(again)
			if string(a) != string(b) {
				t.Fatalf("non-deterministic composition:\n%s\n%s", a, b)
			}
		})
	}
}

func asDomainError(err error, target **domain.Error) bool {
	if err == nil {
		return false
	}
	if de, ok := err.(*domain.Error); ok {
		*target = de
		return true
	}
	return false
}

func TestDestinationSafety(t *testing.T) {
	cases := []struct {
		name    string
		surface domain.SurfaceID
		dest    string
		wantErr string
	}{
		{"traversal parent", "web-admin", "../evil", "traversal"},
		{"traversal nested", "web-admin", "apps/../evil", "traversal"},
		{"absolute unix", "web-admin", "/apps/admin", "absolute"},
		{"absolute drive", "web-admin", `C:\apps\admin`, "absolute"},
		{"empty", "web-admin", "", "empty"},
		{"dot", "web-admin", ".", "clean relative"},
		{"blank segment", "web-admin", "apps//admin", "clean relative"},
		{"ok default style", "web-admin", "apps/admin", ""},
		{"ok nested depth", "web-admin", "apps/backoffice/v2", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateDestination(tc.surface, tc.dest)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == "" {
					t.Fatal("expected cleaned destination")
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got destination %q", tc.wantErr, got)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestDestinationCollisions(t *testing.T) {
	surfaces := []domain.SurfaceID{"web-admin", "mobile-native", "api"}
	t.Run("duplicate", func(t *testing.T) {
		_, err := AssignDestinations(surfaces, map[domain.SurfaceID]string{
			"web-admin": "apps/main", "mobile-native": "apps/main",
		})
		if err == nil || !strings.Contains(err.Error(), "collision") {
			t.Fatalf("expected collision error, got %v", err)
		}
	})
	t.Run("nested overlap", func(t *testing.T) {
		_, err := AssignDestinations(surfaces, map[domain.SurfaceID]string{
			"web-admin": "apps/main", "mobile-native": "apps/main/inner",
		})
		if err == nil || !strings.Contains(err.Error(), "nested") {
			t.Fatalf("expected nested error, got %v", err)
		}
	})
	t.Run("parent swallows child default", func(t *testing.T) {
		_, err := AssignDestinations(surfaces, map[domain.SurfaceID]string{
			"web-admin": "apps",
		})
		if err == nil || !strings.Contains(err.Error(), "nested") {
			t.Fatalf("expected nested error, got %v", err)
		}
	})
	t.Run("distinct ok", func(t *testing.T) {
		got, err := AssignDestinations(surfaces, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 || got["api"] != "services/api" {
			t.Fatalf("unexpected destinations: %+v", got)
		}
	})
}

func TestProviderEligibility(t *testing.T) {
	eligible := domain.Boilerplate{
		ID: "ok", Repo: "https://example/ok", Pin: "v1.0.0", Adapter: "tanstack",
		DecisionStatus: "curated", DeliveryStatus: "stable",
	}
	if !Eligible(eligible) {
		t.Fatal("curated+stable provider should be eligible")
	}
	for name, mutate := range map[string]func(*domain.Boilerplate){
		"deprecated decision":   func(b *domain.Boilerplate) { b.DecisionStatus = "deprecated" },
		"rejected decision":     func(b *domain.Boilerplate) { b.DecisionStatus = "rejected" },
		"experimental decision": func(b *domain.Boilerplate) { b.DecisionStatus = "experimental" },
		"missing pin":           func(b *domain.Boilerplate) { b.Pin = "" },
		"missing adapter":       func(b *domain.Boilerplate) { b.Adapter = "" },
	} {
		t.Run(name, func(t *testing.T) {
			b := eligible
			mutate(&b)
			if Eligible(b) {
				t.Fatalf("provider %+v should be ineligible", b)
			}
		})
	}
}
