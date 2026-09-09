package foundationconfig

import (
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// fixtureCatalog builds a minimal catalog with one generated boilerplate
// carrying three ordered profiles and the surfaces/capabilities the
// selection signals need.
func fixtureCatalog() catalog.Catalog {
	gen := domain.GenerateSpec{
		Run:            domain.AdapterCommand{Run: []string{"fake-gen", "--out={output}"}},
		Output:         "{output}",
		DefaultProfile: "minimal",
		Profiles: []domain.GeneratorProfile{
			{ID: "minimal", Arguments: []string{"--profile=minimal"}},
			{ID: "data", Arguments: []string{"--profile=data"},
				Requires: &domain.ProfileRequirements{Capabilities: []string{"shared-backend"}}},
			{ID: "auth", Arguments: []string{"--profile=auth"},
				Requires: &domain.ProfileRequirements{
					Capabilities:       []string{"shared-backend"},
					UnlessCapabilities: []string{"anonymous-public-access"},
				}},
		},
	}
	var bp domain.Boilerplate
	bp.ID = "gen-api"
	bp.Pin = "pin1"
	bp.Adapter = "gen"
	bp.AdapterSpec = &domain.AdapterSpec{
		Name: "gen", Operations: []string{"fetch", "generate"}, Generate: &gen,
	}
	return catalog.Catalog{
		CatalogVersion: "test",
		Recipes:        []domain.Recipe{{ID: "R", Provides: domain.Provides{}}},
		Boilerplates:   []domain.Boilerplate{bp},
		Surfaces:       []domain.Surface{{ID: "api"}, {ID: "web-admin"}},
		Capabilities: []domain.Capability{
			{ID: "shared-backend"}, {ID: "anonymous-public-access"},
		},
	}
}

func decisionWith(reasons []string, derived ...string) domain.ArchitectureDecision {
	d := domain.ArchitectureDecision{Status: domain.StatusResolved}
	d.Reasons = reasons
	for _, id := range derived {
		d.DerivedRequirements = append(d.DerivedRequirements, domain.DerivedRequirement{ID: id})
	}
	return d
}

func comp() composer.Component {
	return composer.Component{Surface: "api", Boilerplate: "gen-api", Pin: "pin1", Destination: "services/api"}
}

// TestGeneratedFoundationUsesDefaultMinimalProfile pins the smallest-valid
// default: no signals, no richer profile.
func TestGeneratedFoundationUsesDefaultMinimalProfile(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	got, err := SelectProfile(bp, decisionWith(nil), cat)
	if err != nil {
		t.Fatal(err)
	}
	if got != "minimal" {
		t.Fatalf("profile = %q, want minimal", got)
	}
}

// TestProfileSelectionIsDeterministic runs selection twice over the same
// decision and requires the same answer.
func TestProfileSelectionIsDeterministic(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	d := decisionWith(nil, "shared-backend")
	first, err := SelectProfile(bp, d, cat)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		got, err := SelectProfile(bp, d, cat)
		if err != nil {
			t.Fatal(err)
		}
		if got != first {
			t.Fatalf("run %d: profile = %q, want %q", i, got, first)
		}
	}
}

// TestSmallestCompatibleProfileWins covers the escalation: shared-backend
// selects data, and a non-public backend selects auth (richest match).
func TestSmallestCompatibleProfileWins(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	got, err := SelectProfile(bp, decisionWith(nil, "shared-backend"), cat)
	if err != nil {
		t.Fatal(err)
	}
	if got != "auth" {
		t.Fatalf("shared-backend without public access: profile = %q, want auth", got)
	}
	public := decisionWith(nil, "shared-backend", "anonymous-public-access")
	got, err = SelectProfile(bp, public, cat)
	if err != nil {
		t.Fatal(err)
	}
	if got != "data" {
		t.Fatalf("public shared backend: profile = %q, want data", got)
	}
}

// TestProductFeatureDoesNotDisqualifyFoundation proves an unmatched
// product requirement never fails selection: the foundation stays
// eligible on the smallest profile.
func TestProductFeatureDoesNotDisqualifyFoundation(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	d := decisionWith([]string{"includes product feature: pdf-export (+1 tie-break)"})
	got, err := SelectProfile(bp, d, cat)
	if err != nil {
		t.Fatalf("unknown product feature must not fail selection: %v", err)
	}
	if got != "minimal" {
		t.Fatalf("profile = %q, want minimal fallback", got)
	}
	if _, err := Resolve(comp(), "demo", 1, bp, d, cat); err != nil {
		t.Fatalf("resolve must succeed: %v", err)
	}
}

// TestUnknownProductFeatureRemainsForGentle documents that selection
// ignores unknown features instead of consuming them: the feature id
// appears nowhere in the resolved configuration.
func TestUnknownProductFeatureRemainsForGentle(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	d := decisionWith([]string{"includes product feature: holograms (+1 tie-break)"})
	mat, err := Resolve(comp(), "demo", 1, bp, d, cat)
	if err != nil {
		t.Fatal(err)
	}
	for _, arg := range mat.Arguments {
		if arg == "holograms" {
			t.Fatal("unknown feature leaked into generation arguments")
		}
	}
}

// TestTechnicalConstraintIsRespectedDuringGenerationConfig fails closed
// when a hard must-not-use constraint excludes the foundation.
func TestTechnicalConstraintIsRespectedDuringGenerationConfig(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	bp.Technology = domain.Technology{Framework: []string{"gen-fw"}}
	d := decisionWith([]string{"covers required surface api"})
	d.Candidates = []domain.Candidate{{Recipe: "R", NegativeReasons: []string{"excluded by must-not-use tech gen-fw"}}}
	if _, err := Resolve(comp(), "demo", 1, bp, d, cat); err == nil {
		t.Fatal("expected hard-constraint failure, got nil")
	}
}

// TestResolveCopyStaysMinimal keeps copy components free of generation
// metadata beyond the fingerprint.
func TestResolveCopyStaysMinimal(t *testing.T) {
	cat := fixtureCatalog()
	bp, _ := catalog.NewIndex(cat).Boilerplate("gen-api")
	bp.AdapterSpec = nil
	bp.Adapter = "plain"
	mat, err := Resolve(comp(), "demo", 1, bp, decisionWith(nil), cat)
	if err != nil {
		t.Fatal(err)
	}
	if mat.Strategy != domain.StrategyCopy {
		t.Fatalf("strategy = %q, want copy", mat.Strategy)
	}
	if mat.Profile != "" || len(mat.Arguments) != 0 || mat.Name != "" {
		t.Fatalf("copy config must not carry generation metadata: %+v", mat)
	}
	if mat.AdapterFingerprint == "" {
		t.Fatal("copy config must carry the adapter fingerprint")
	}
}

// TestLogicalNameSeparatesSingleAndMultiSurface pins the naming rule.
func TestLogicalNameSeparatesSingleAndMultiSurface(t *testing.T) {
	if got := LogicalName("inventory", "admin", 1); got != "inventory" {
		t.Fatalf("single-surface name = %q", got)
	}
	if got := LogicalName("inventory", "admin", 2); got != "inventory-admin" {
		t.Fatalf("multi-surface name = %q", got)
	}
	if got := LogicalName("Field Ops", "web-admin", 2); got != "field-ops-web-admin" {
		t.Fatalf("slugified name = %q", got)
	}
}
