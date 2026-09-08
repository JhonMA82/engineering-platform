package resolver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

type routingCase struct {
	Name   string               `json:"name"`
	Intent domain.ProjectIntent `json:"intent"`
	Expect struct {
		Status                   string   `json:"status"`
		SelectedRecipe           string   `json:"selected_recipe,omitempty"`
		MustIncludeReasons       []string `json:"must_include_reasons,omitempty"`
		MustNotSelect            []string `json:"must_not_select,omitempty"`
		DiscriminatingDimensions []string `json:"discriminating_dimensions,omitempty"`
	} `json:"expect"`
}

func loadRoutingCases(t *testing.T) []routingCase {
	t.Helper()
	files, err := filepath.Glob("../../testdata/routing/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no routing fixtures: %v", err)
	}
	var out []routingCase
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var tc routingCase
		if err := json.Unmarshal(raw, &tc); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		out = append(out, tc)
	}
	return out
}

func TestRoutingScenarios(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	cases := loadRoutingCases(t)
	if len(cases) < 15 {
		t.Fatalf("expected at least 15 routing scenarios, got %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := Resolve(tc.Intent, cat)
			if string(got.Status) != tc.Expect.Status {
				t.Fatalf("status = %q, want %q\n%s", got.Status, tc.Expect.Status, Explain(got))
			}
			if tc.Expect.SelectedRecipe != "" {
				if got.Selected == nil || got.Selected.Recipe != tc.Expect.SelectedRecipe {
					t.Fatalf("selected = %+v, want %q", got.Selected, tc.Expect.SelectedRecipe)
				}
			}
			haystack := strings.Join(got.Reasons, "\n")
			for _, c := range got.Candidates {
				haystack += "\n" + strings.Join(c.PositiveReasons, "\n") + "\n" + strings.Join(c.NegativeReasons, "\n")
			}
			for _, want := range tc.Expect.MustIncludeReasons {
				if !strings.Contains(haystack, want) {
					t.Fatalf("reasons lack %q\n%s", want, Explain(got))
				}
			}
			for _, banned := range tc.Expect.MustNotSelect {
				if got.Selected != nil && got.Selected.Recipe == banned {
					t.Fatalf("must not select %q", banned)
				}
			}
			if len(tc.Expect.DiscriminatingDimensions) > 0 {
				have := map[string]bool{}
				for _, u := range got.UnresolvedDimensions {
					have[u.Dimension] = true
				}
				for _, d := range tc.Expect.DiscriminatingDimensions {
					if !have[d] {
						t.Fatalf("missing discriminating dimension %q, have %+v", d, got.UnresolvedDimensions)
					}
				}
			}
			if got.IntentFingerprint == "" {
				t.Fatal("decision lacks intent fingerprint")
			}
		})
	}
}

func TestResolveDeterministic(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	cases := loadRoutingCases(t)
	for _, tc := range cases {
		a := Resolve(tc.Intent, cat)
		b := Resolve(tc.Intent, cat)
		ra, _ := json.Marshal(a)
		rb, _ := json.Marshal(b)
		if string(ra) != string(rb) {
			t.Fatalf("non-deterministic decision for %s", tc.Name)
		}
	}
}
