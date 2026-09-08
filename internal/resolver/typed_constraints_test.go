package resolver

import (
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// typedTestCatalog builds a two-foundation catalog: a TanStack admin and a
// Next admin over the same surface, plus a Python and a Go api. All names
// live in test data; the resolver matches them through catalog metadata.
func typedTestCatalog() catalog.Catalog {
	bp := func(id, adapter string, tech domain.Technology, tags ...string) domain.Boilerplate {
		return domain.Boilerplate{
			ID: id, Repo: "https://example.com/" + id, Pin: "v1",
			Adapter: adapter, DeliveryStatus: "stable", DecisionStatus: "curated",
			Provides:   domain.Provides{Surfaces: []domain.SurfaceID{"web-admin"}},
			TechTags:   tags,
			Technology: tech,
		}
	}
	api := func(id string, tags ...string) domain.Boilerplate {
		b := bp(id, id, domain.Technology{}, tags...)
		b.Provides = domain.Provides{Surfaces: []domain.SurfaceID{"api"}}
		return b
	}
	recipe := func(id string, surfaces []domain.SurfaceID, bps []string, tags ...string) domain.Recipe {
		return domain.Recipe{
			ID: id, Version: "1.0.0", Status: "stable",
			Provides:                  domain.Provides{Surfaces: surfaces},
			PrimaryBoilerplates:       bps,
			AllowedSurfaceComposition: [][]domain.SurfaceID{surfaces},
			DatabasePolicy: domain.DatabasePolicy{
				DefaultProfile:  "sqlite-local",
				AllowedProfiles: []string{"sqlite-local"},
			},
			TechTags: tags,
		}
	}
	webAdmin := []domain.SurfaceID{"web-admin"}
	apiOnly := []domain.SurfaceID{"api"}
	return catalog.Catalog{
		CatalogVersion: "test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Surfaces: []domain.Surface{
			{ID: "web-admin"}, {ID: "api"},
		},
		Boilerplates: []domain.Boilerplate{
			bp("tanstack-admin", "tanstack",
				domain.Technology{Framework: []string{"tanstack"}, Language: []string{"typescript"}}, "tanstack"),
			bp("next-admin", "next",
				domain.Technology{Framework: []string{"next"}, Language: []string{"typescript"}}, "next"),
			api("py-api", "python"),
			api("go-api", "go"),
		},
		DatabaseProfiles: []domain.DatabaseProfile{{ID: "sqlite-local", Engine: "sqlite"}},
		Recipes: []domain.Recipe{
			recipe("R-TAN", webAdmin, []string{"tanstack-admin"}, "tanstack"),
			recipe("R-NEXT", webAdmin, []string{"next-admin"}, "next"),
			recipe("R-PY", apiOnly, []string{"py-api"}, "python"),
			recipe("R-GO", apiOnly, []string{"go-api"}, "go"),
		},
	}
}

func resolveOnTypedCatalog(t *testing.T, intent domain.ProjectIntent) domain.ArchitectureDecision {
	t.Helper()
	cat := typedTestCatalog()
	if err := catalog.Validate(cat); err != nil {
		t.Fatalf("test catalog invalid: %v", err)
	}
	return Resolve(intent, cat)
}

// TestTypedMustUseSelectsTanStack proves §2.3: must-use framework=tanstack
// keeps the TanStack candidate and rejects the Next candidate through
// catalog technology metadata.
func TestTypedMustUseSelectsTanStack(t *testing.T) {
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "tanstack backoffice",
		Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "framework", Kind: "must-use", Value: "tanstack"},
		},
	}
	got := resolveOnTypedCatalog(t, intent)
	if got.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want resolved\n%s", got.Status, Explain(got))
	}
	if got.Selected == nil || got.Selected.Recipe != "R-TAN" {
		t.Fatalf("selected = %+v, want R-TAN", got.Selected)
	}
	for _, c := range got.Candidates {
		if c.Recipe != "R-NEXT" {
			continue
		}
		if c.Eligible {
			t.Fatal("R-NEXT must be ineligible under must-use framework=tanstack")
		}
		if !strings.Contains(strings.Join(c.NegativeReasons, "\n"), "must-use framework=tanstack") {
			t.Fatalf("R-NEXT lacks the typed rejection: %v", c.NegativeReasons)
		}
	}
}

// TestTypedMustNotUseExcludesFramework proves must-not-use with a target
// excludes only the matching foundation.
func TestTypedMustNotUseExcludesFramework(t *testing.T) {
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "non-tanstack backoffice",
		Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "framework", Kind: "must-not-use", Value: "tanstack"},
		},
	}
	got := resolveOnTypedCatalog(t, intent)
	if got.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want resolved\n%s", got.Status, Explain(got))
	}
	if got.Selected == nil || got.Selected.Recipe != "R-NEXT" {
		t.Fatalf("selected = %+v, want R-NEXT", got.Selected)
	}
}

// TestTypedPreferInfluencesRanking proves prefer with a target affects the
// score only: both api candidates stay eligible, the Python one wins.
func TestTypedPreferInfluencesRanking(t *testing.T) {
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "python-flavored api",
		Surfaces:      []domain.SurfaceIntent{{Kind: "api", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "language", Kind: "prefer", Value: "python"},
		},
	}
	got := resolveOnTypedCatalog(t, intent)
	if got.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want resolved\n%s", got.Status, Explain(got))
	}
	if got.Selected == nil || got.Selected.Recipe != "R-PY" {
		t.Fatalf("selected = %+v, want R-PY", got.Selected)
	}
	for _, c := range got.Candidates {
		if c.Recipe == "R-GO" && !c.Eligible {
			t.Fatalf("prefer must never eliminate: R-GO ineligible: %v", c.NegativeReasons)
		}
	}
	if !strings.Contains(strings.Join(got.Reasons, "\n"), "preference match: python") {
		t.Fatalf("winner lacks the preference reason: %v", got.Reasons)
	}
}

// TestHardDeploymentConstraintIsInvalid proves the v1.0.1 contract: a hard
// deployment constraint is rejected at validation (invalid intent) instead
// of being accepted and silently ignored.
func TestHardDeploymentConstraintIsInvalid(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	for _, kind := range []string{"must-use", "must-not-use"} {
		t.Run(kind, func(t *testing.T) {
			intent := domain.ProjectIntent{
				SchemaVersion: 1,
				Name:          "edge admin",
				Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
				TechnicalConstraints: []domain.TechnicalConstraint{
					{Target: "deployment", Kind: kind, Value: "edge"},
				},
			}
			got := Resolve(intent, cat)
			if got.Status != domain.StatusInvalid {
				t.Fatalf("status = %q, want invalid\n%s", got.Status, Explain(got))
			}
			if !strings.Contains(strings.Join(got.Reasons, "\n"), "unsupported technical constraint target: deployment") {
				t.Fatalf("reasons lack the unsupported-target error: %v", got.Reasons)
			}
		})
	}
}

// TestMustUseDatabaseMissingProfileIsCatalogGap proves §2.4: a mandatory
// database value with no curated profile is a database-profile catalog-gap,
// even though recipes cover the surface.
func TestMustUseDatabaseMissingProfileIsCatalogGap(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "turso backoffice",
		Surfaces:      []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "database", Kind: "must-use", Value: "turso"},
		},
	}
	got := Resolve(intent, cat)
	if got.Status != domain.StatusCatalogGap {
		t.Fatalf("status = %q, want catalog-gap\n%s", got.Status, Explain(got))
	}
	found := false
	for _, m := range got.MissingArchitecture {
		if m.Kind == "database-profile" && m.Ref == "turso" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing database-profile=turso gap: %+v", got.MissingArchitecture)
	}
	if !strings.Contains(strings.Join(got.Reasons, "\n"), "database-profile=turso") {
		t.Fatalf("reasons lack the gap: %v", got.Reasons)
	}
}

// TestPreferUnavailableDatabaseNeverGaps proves a prefer database value with
// no curated profile still resolves; the composer owns the fallback.
func TestPreferUnavailableDatabaseNeverGaps(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "turso-curious api",
		Surfaces:      []domain.SurfaceIntent{{Kind: "api", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "database", Kind: "prefer", Value: "turso"},
		},
	}
	got := Resolve(intent, cat)
	if got.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want resolved\n%s", got.Status, Explain(got))
	}
}
