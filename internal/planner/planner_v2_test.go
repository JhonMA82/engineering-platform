package planner_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/foundationconfig"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// planForIntent resolves, composes and plans one routing intent.
func planForIntent(t *testing.T, intentName string) planner.MaterializationPlan {
	t.Helper()
	root := repoRoot(t)
	raw := routingIntent(t, root, intentName)
	_, _, plan, err := app.PlanProject(raw, "")
	if err != nil {
		t.Fatalf("plan %s: %v", intentName, err)
	}
	return *plan
}

func findComponent(plan planner.MaterializationPlan, surface string) planner.PlanComponent {
	for _, c := range plan.Components {
		if c.Surface == surface {
			return c
		}
	}
	return planner.PlanComponent{}
}

// TestPlanV2ContainsGenerationStrategy requires schema 2 with resolved
// per-component strategies: the api surface generates, the admin copies.
func TestPlanV2ContainsGenerationStrategy(t *testing.T) {
	plan := planForIntent(t, "admin-mobile.json")
	if plan.SchemaVersion != 2 {
		t.Fatalf("schema_version = %d, want 2", plan.SchemaVersion)
	}
	api := findComponent(plan, "api")
	if api.EffectiveStrategy() != domain.StrategyGenerate {
		t.Fatalf("api strategy = %q, want generate", api.EffectiveStrategy())
	}
	admin := findComponent(plan, "web-admin")
	if admin.EffectiveStrategy() != domain.StrategyCopy {
		t.Fatalf("admin strategy = %q, want copy", admin.EffectiveStrategy())
	}
}

// TestPlanContainsResolvedProfile requires the resolved profile,
// non-runtime arguments and the adapter fingerprint on generated
// components.
func TestPlanContainsResolvedProfile(t *testing.T) {
	plan := planForIntent(t, "admin-mobile.json")
	api := findComponent(plan, "api")
	if api.Materialization.Profile == "" {
		t.Fatal("generated component must record a profile")
	}
	if len(api.Materialization.Arguments) == 0 {
		t.Fatal("generated component must record resolved arguments")
	}
	if api.Materialization.AdapterFingerprint == "" {
		t.Fatal("generated component must record the adapter fingerprint")
	}
	if api.Materialization.Name == "" {
		t.Fatal("generated component must record the logical name")
	}
}

// matsForPlan rebuilds the fingerprint input from plan components.
func matsForPlan(plan planner.MaterializationPlan, comp composer.Composition) map[domain.SurfaceID]foundationconfig.MaterializationConfig {
	mats := map[domain.SurfaceID]foundationconfig.MaterializationConfig{}
	for i, c := range plan.Components {
		mats[comp.Components[i].Surface] = c.Materialization
		_ = i
	}
	return mats
}

// TestPlanFingerprintChangesWhenProfileChanges pins drift sensitivity.
func TestPlanFingerprintChangesWhenProfileChanges(t *testing.T) {
	plan := planForIntent(t, "admin-mobile.json")
	root := repoRoot(t)
	raw := routingIntent(t, root, "admin-mobile.json")
	_, comp, _, err := app.PlanProject(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	base := planner.FingerprintPlan("fp", "cat", *comp, matsForPlan(plan, *comp))
	edited := plan
	for i, c := range edited.Components {
		if c.Surface == "api" {
			edited.Components[i].Materialization.Profile = "platform"
		}
	}
	other := planner.FingerprintPlan("fp", "cat", *comp, matsForPlan(edited, *comp))
	if base == other {
		t.Fatal("fingerprint must change when the profile changes")
	}
}

// TestPlanFingerprintChangesWhenArgumentsChange pins drift sensitivity.
func TestPlanFingerprintChangesWhenArgumentsChange(t *testing.T) {
	plan := planForIntent(t, "admin-mobile.json")
	root := repoRoot(t)
	raw := routingIntent(t, root, "admin-mobile.json")
	_, comp, _, err := app.PlanProject(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	base := planner.FingerprintPlan("fp", "cat", *comp, matsForPlan(plan, *comp))
	edited := plan
	for i, c := range edited.Components {
		if c.Surface == "api" {
			edited.Components[i].Materialization.Arguments = append(
				append([]string{}, c.Materialization.Arguments...), "--extra")
		}
	}
	other := planner.FingerprintPlan("fp", "cat", *comp, matsForPlan(edited, *comp))
	if base == other {
		t.Fatal("fingerprint must change when arguments change")
	}
}

// TestPlanDoesNotContainTemporaryPaths serializes the plan and rejects
// machine-specific runtime locations.
func TestPlanDoesNotContainTemporaryPaths(t *testing.T) {
	plan := planForIntent(t, "admin-mobile.json")
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	for _, marker := range []string{"/tmp/", "/var/folders", ".staging-", "C:\\", "/sandbox-"} {
		if strings.Contains(doc, marker) {
			t.Fatalf("plan serializes a temporary path %q", marker)
		}
	}
}

// TestCopyComponentStillPlansCorrectly keeps static providers on copy
// with a fingerprint but no generation metadata.
func TestCopyComponentStillPlansCorrectly(t *testing.T) {
	plan := planForIntent(t, "admin-only.json")
	if len(plan.Components) != 1 {
		t.Fatalf("components = %d, want 1", len(plan.Components))
	}
	c := plan.Components[0]
	if c.EffectiveStrategy() != domain.StrategyCopy {
		t.Fatalf("strategy = %q, want copy", c.EffectiveStrategy())
	}
	if c.Materialization.Profile != "" || len(c.Materialization.Arguments) != 0 {
		t.Fatalf("copy component must not carry generation metadata: %+v", c.Materialization)
	}
	if c.Materialization.AdapterFingerprint == "" {
		t.Fatal("copy component must carry the adapter fingerprint")
	}
	if plan.Fingerprint == "" {
		t.Fatal("plan must carry a fingerprint")
	}
}

// TestV1PlanReadsAsCopy unmarshals a schema v1 component and treats it
// as copy semantics.
func TestV1PlanReadsAsCopy(t *testing.T) {
	var c planner.PlanComponent
	raw := `{"boilerplate":"b","pin":"p","destination":"apps/x","surface":"api"}`
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.EffectiveStrategy() != domain.StrategyCopy {
		t.Fatalf("v1 strategy = %q, want copy", c.EffectiveStrategy())
	}
}

// TestPlanGoldenFilesExist keeps the golden helper honest about paths.
func TestPlanGoldenFilesExist(t *testing.T) {
	root := repoRoot(t)
	for golden := range goldenPlans {
		if _, err := os.Stat(filepath.Join(root, "testdata", "plans", golden)); err != nil {
			t.Fatalf("golden %s: %v", golden, err)
		}
	}
}
