package composer

import (
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

func dbTestCatalog() catalog.Catalog {
	return catalog.Catalog{
		CatalogVersion: "test",
		MinCoreVersion: "1.0.0",
		SchemaVersion:  1,
		DatabaseProfiles: []domain.DatabaseProfile{
			{ID: "postgresql-managed", Engine: "postgres", Provider: "managed-postgres"},
			{ID: "sqlite-local", Engine: "sqlite", Provider: "sqlite"},
		},
	}
}

func dbTestRecipe() domain.Recipe {
	return domain.Recipe{
		ID:      "GP-DB",
		Version: "1.0.0",
		Status:  "stable",
		DatabasePolicy: domain.DatabasePolicy{
			DefaultProfile:  "postgresql-managed",
			AllowedProfiles: []string{"postgresql-managed", "sqlite-local"},
		},
	}
}

// TestSelectDatabaseProfileHonorsMustUse proves §2.4: a must-use database
// value selects the allowed profile identifying it even when it is not the
// recipe default, with an explanatory note.
func TestSelectDatabaseProfileHonorsMustUse(t *testing.T) {
	cat := dbTestCatalog()
	profile, note, err := SelectDatabaseProfileFor(dbTestRecipe(), cat,
		[]string{"sqlite-local"}, nil, nil)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if profile != "sqlite-local" {
		t.Fatalf("profile = %q, want sqlite-local", profile)
	}
	if !strings.Contains(note, "must-use database=sqlite-local") {
		t.Fatalf("note lacks the must-use explanation: %q", note)
	}
}

// TestSelectDatabaseProfileMatchesEngine proves matching runs against
// profile engine data, not the profile id alone.
func TestSelectDatabaseProfileMatchesEngine(t *testing.T) {
	cat := dbTestCatalog()
	profile, _, err := SelectDatabaseProfileFor(dbTestRecipe(), cat,
		[]string{"postgres"}, nil, nil)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if profile != "postgresql-managed" {
		t.Fatalf("profile = %q, want postgresql-managed", profile)
	}
}

// TestSelectDatabaseProfileRejectsUnsatisfiableMustUse proves a must-use
// value the recipe policy cannot honor is a typed composition error. (A
// catalog-wide miss never reaches composition: the resolver reports a
// database-profile catalog-gap first.)
func TestSelectDatabaseProfileRejectsUnsatisfiableMustUse(t *testing.T) {
	cat := dbTestCatalog()
	recipe := dbTestRecipe()
	recipe.DatabasePolicy.AllowedProfiles = []string{"postgresql-managed"}
	_, _, err := SelectDatabaseProfileFor(recipe, cat, []string{"sqlite-local"}, nil, nil)
	if err == nil {
		t.Fatal("expected composition error, got nil")
	}
	derr, ok := err.(*domain.Error)
	if !ok || derr.Class != domain.ClassComposition {
		t.Fatalf("expected composition-class error, got %T: %v", err, err)
	}
}

// TestSelectDatabaseProfilePreferFallback proves a prefer value with no
// curated profile falls back to the default with an explained deviation
// reason instead of failing.
func TestSelectDatabaseProfilePreferFallback(t *testing.T) {
	cat := dbTestCatalog()
	profile, note, err := SelectDatabaseProfileFor(dbTestRecipe(), cat, nil, nil,
		[]domain.Preference{{Kind: "prefer", Value: "turso"}})
	if err != nil {
		t.Fatalf("prefer must never fail selection: %v", err)
	}
	if profile != "postgresql-managed" {
		t.Fatalf("profile = %q, want default postgresql-managed", profile)
	}
	if !strings.Contains(note, "turso") || !strings.Contains(note, "default") {
		t.Fatalf("note must explain the fallback: %q", note)
	}
}

// TestSelectDatabaseProfileAvoidsMustNotUse proves must-not-use steers
// selection to an alternative when the policy offers one.
func TestSelectDatabaseProfileAvoidsMustNotUse(t *testing.T) {
	cat := dbTestCatalog()
	profile, note, err := SelectDatabaseProfileFor(dbTestRecipe(), cat,
		nil, []string{"postgresql-managed"}, nil)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if profile != "sqlite-local" {
		t.Fatalf("profile = %q, want sqlite-local", profile)
	}
	if !strings.Contains(note, "must-not-use") {
		t.Fatalf("note lacks the avoidance explanation: %q", note)
	}
}

// TestComposePreferUnavailableDatabaseFallback is the end-to-end §2.4
// prefer path: a prefer-database intent resolves, composes onto the recipe
// default, and carries the deviation explanation on the composition.
func TestComposePreferUnavailableDatabaseFallback(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "prefer turso api",
		Surfaces:      []domain.SurfaceIntent{{Kind: "api", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "database", Kind: "prefer", Value: "turso"},
		},
	}
	decision := resolver.Resolve(intent, cat)
	if decision.Status != domain.StatusResolved {
		t.Fatalf("intent did not resolve: %s", resolver.Explain(decision))
	}
	comp, err := Compose(decision, cat)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if comp.DatabaseProfile == "" {
		t.Fatal("composition lacks a database profile")
	}
	if !strings.Contains(comp.DatabaseNote, "turso") {
		t.Fatalf("composition lacks the fallback explanation: %+v", comp)
	}
}

// TestComposeMustUseDatabaseSelectsProfile is the end-to-end §2.4
// must-use path: sqlite-local is selected over the postgres default.
func TestComposeMustUseDatabaseSelectsProfile(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "local sqlite api",
		Surfaces:      []domain.SurfaceIntent{{Kind: "api", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "database", Kind: "must-use", Value: "sqlite-local"},
		},
	}
	decision := resolver.Resolve(intent, cat)
	if decision.Status != domain.StatusResolved {
		t.Fatalf("intent did not resolve: %s", resolver.Explain(decision))
	}
	comp, err := Compose(decision, cat)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if comp.DatabaseProfile != "sqlite-local" {
		t.Fatalf("profile = %q, want sqlite-local (note %q)", comp.DatabaseProfile, comp.DatabaseNote)
	}
}
