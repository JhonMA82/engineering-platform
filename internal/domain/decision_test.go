package domain

import (
	"encoding/json"
	"testing"
)

func sampleDecision() ArchitectureDecision {
	return ArchitectureDecision{
		SchemaVersion:     1,
		Status:            StatusResolved,
		IntentFingerprint: "abc123",
		Selected:          &SelectedRecipe{Recipe: "GP-02", RecipeVersion: "1.0.0", Score: 90},
		Confidence:        Confidence{Level: "high", Margin: 21},
		Reasons:           []string{"covers required surface web-admin"},
		Candidates: []Candidate{
			{Recipe: "GP-02", Eligible: true, Score: 90, PositiveReasons: []string{"covers required surface web-admin"}},
			{Recipe: "GP-06", Eligible: true, Score: 69, PositiveReasons: []string{"covers required surface web-admin"}},
		},
		DerivedRequirements:  []DerivedRequirement{{ID: "shared-backend", Source: "test"}},
		UnresolvedDimensions: nil,
	}
}

func TestArchitectureDecisionRoundtrip(t *testing.T) {
	want := sampleDecision()
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ArchitectureDecision
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != StatusResolved || got.Selected.Recipe != "GP-02" {
		t.Fatalf("roundtrip lost data: %+v", got)
	}
	back, _ := json.Marshal(got)
	if string(raw) != string(back) {
		t.Fatalf("roundtrip mismatch:\n%s\n%s", raw, back)
	}
}

func TestRecipeBoilerplateValidation(t *testing.T) {
	if err := (Recipe{ID: "GP-02", Version: "1.0.0"}).Validate(); err != nil {
		t.Fatalf("recipe should be valid: %v", err)
	}
	if err := (Recipe{}).Validate(); err == nil {
		t.Fatal("empty recipe should be invalid")
	}
	bp := Boilerplate{ID: "x", Repo: "https://example/x", Pin: "v1.0.0", Adapter: "tanstack"}
	if err := bp.Validate(); err != nil {
		t.Fatalf("boilerplate should be valid: %v", err)
	}
	bp.Pin = ""
	if err := bp.Validate(); err == nil {
		t.Fatal("boilerplate without pin should be invalid")
	}
	if err := (Surface{ID: "api"}).Validate(); err != nil {
		t.Fatalf("surface should be valid: %v", err)
	}
	if err := (Capability{ID: "realtime"}).Validate(); err != nil {
		t.Fatalf("capability should be valid: %v", err)
	}
	if err := (DatabaseProfile{ID: "sqlite-local"}).Validate(); err != nil {
		t.Fatalf("database profile should be valid: %v", err)
	}
}
