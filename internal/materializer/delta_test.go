package materializer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/planner"
)

func TestMaterializeDeltaHappyPath(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("seed materialize: %v", err)
	}
	extra := plan.Components[0]
	extra.Destination = "apps/extra"
	extra.Surface = "public-intake"
	if err := MaterializeDelta(DeltaRequest{
		Components:           []planner.PlanComponent{extra},
		Catalog:              cat,
		ProjectDir:           out,
		ExistingDestinations: []string{"apps/web", "services/api"},
	}); err != nil {
		t.Fatalf("delta: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "apps", "extra", "index.html")); err != nil {
		t.Errorf("expected delta file: %v", err)
	}
}

func TestMaterializeDeltaRejectsExistingDestination(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("seed materialize: %v", err)
	}
	err := MaterializeDelta(DeltaRequest{
		Components:           plan.Components[:1],
		Catalog:              cat,
		ProjectDir:           out,
		ExistingDestinations: []string{"services/api"},
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected destination-collision error, got %v", err)
	}
}

func TestMaterializeDeltaRequiresProjectDir(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	err := MaterializeDelta(DeltaRequest{
		Components: plan.Components[:1],
		Catalog:    cat,
		ProjectDir: filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil {
		t.Fatalf("expected missing-dir error, got nil")
	}
}
