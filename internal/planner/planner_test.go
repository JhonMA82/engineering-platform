package planner_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

var update = flag.Bool("update", false, "regenerate golden files")

// goldenPlans maps a golden file to the routing wrapper holding its intent.
var goldenPlans = map[string]string{
	"admin-only.golden.json":   "admin-only.json",
	"admin-mobile.golden.json": "admin-mobile.json",
}

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

func routingIntent(t *testing.T, root, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "testdata", "routing", name))
	if err != nil {
		t.Fatalf("read routing %s: %v", name, err)
	}
	var wrapper struct {
		Intent json.RawMessage `json:"intent"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil || wrapper.Intent == nil {
		t.Fatalf("routing %s lacks intent: %v", name, err)
	}
	return wrapper.Intent
}

func TestPlannerGolden(t *testing.T) {
	root := repoRoot(t)
	for golden, routing := range goldenPlans {
		t.Run(golden, func(t *testing.T) {
			intentJSON := routingIntent(t, root, routing)
			_, _, plan, err := app.PlanProject(intentJSON, "")
			if err != nil {
				t.Fatalf("plan %s: %v", routing, err)
			}
			raw, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			raw = append(raw, '\n')
			path := filepath.Join(root, "testdata", "plans", golden)
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(path, raw, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s (rerun with -update): %v", golden, err)
			}
			if string(raw) != string(want) {
				t.Fatalf("golden %s mismatch:\n got: %s\nwant: %s", golden, raw, want)
			}
		})
	}
}

func TestPlanDeterministic(t *testing.T) {
	root := repoRoot(t)
	intentJSON := routingIntent(t, root, "admin-mobile.json")
	_, _, first, err := app.PlanProject(intentJSON, "")
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	_, _, second, err := app.PlanProject(intentJSON, "")
	if err != nil {
		t.Fatalf("replan: %v", err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) {
		t.Fatalf("non-deterministic plan:\n%s\n%s", a, b)
	}
	if first.Fingerprint == "" {
		t.Fatal("plan lacks fingerprint")
	}
}

func TestPlanProjectRejectsUnresolved(t *testing.T) {
	root := repoRoot(t)
	for _, routing := range []string{"ambiguous-admin-multiapp.json", "tui-gap.json", "invalid-empty.json"} {
		t.Run(routing, func(t *testing.T) {
			intentJSON := routingIntent(t, root, routing)
			decision, comp, plan, err := app.PlanProject(intentJSON, "")
			if err == nil {
				t.Fatalf("expected error, got plan %+v", plan)
			}
			if plan != nil || comp != nil {
				t.Fatal("unresolved decision must yield no composition and no plan")
			}
			if decision == nil {
				t.Fatal("expected the decision back alongside the error")
			}
			var derr *domain.Error
			asErr, ok := err.(*domain.Error)
			if !ok {
				t.Fatalf("expected typed domain error, got %T: %v", err, err)
			}
			derr = asErr
			if derr.Class != domain.ClassComposition {
				t.Fatalf("expected composition-class error, got %q", derr.Class)
			}
		})
	}
}
