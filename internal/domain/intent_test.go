package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleIntent() ProjectIntent {
	return ProjectIntent{
		SchemaVersion: 1,
		Name:          "field ops",
		Problem:       "track field visits",
		ProductRequirements: []ProductRequirement{
			{ID: "export-pdf", Description: "export visit report as PDF"},
		},
		ArchitectureRequirements: []ArchitectureRequirement{
			{ID: "arch-offline", Ref: "offline-operation"},
		},
		Surfaces: []SurfaceIntent{
			{Kind: "web-admin", Access: "authenticated", Scope: ScopeRequiredNow},
			{Kind: "mobile-native", Scope: ScopePlannedLater},
		},
		Data:                 DataIntent{Persistence: "shared", MultiUser: true},
		Ops:                  OperationalIntent{OfflineOperation: true},
		TechnicalConstraints: []TechnicalConstraint{{Kind: "must-use", Value: "tanstack"}},
		Preferences:          []Preference{{Kind: "prefer", Value: "low-ops"}},
		Scope:                ScopeIntent{PlannedLater: []string{"mobile-native"}},
		Notes:                []string{"phase 1 is admin only"},
	}
}

func TestProjectIntentRoundtrip(t *testing.T) {
	want := sampleIntent()
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ProjectIntent
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	back, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("remarshal: %v", err)
	}
	if string(raw) != string(back) {
		t.Fatalf("roundtrip mismatch:\n%s\n%s", raw, back)
	}
	if got.Surfaces[1].EffectiveScope() != ScopePlannedLater {
		t.Fatalf("scope not preserved: %+v", got.Surfaces[1])
	}
}

func TestProjectIntentValidation(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*ProjectIntent)
		wantErr string
	}{
		{"valid", func(*ProjectIntent) {}, ""},
		{"missing schema", func(p *ProjectIntent) { p.SchemaVersion = 0 }, "schema_version"},
		{"missing name", func(p *ProjectIntent) { p.Name = " " }, "name is required"},
		{"duplicate surface", func(p *ProjectIntent) {
			p.Surfaces = append(p.Surfaces, SurfaceIntent{Kind: "web-admin"})
		}, "duplicate required surface"},
		{"required meets excluded", func(p *ProjectIntent) {
			p.Scope.ExplicitlyExcluded = []string{"web-admin"}
		}, "both required and excluded"},
		{"empty product id", func(p *ProjectIntent) {
			p.ProductRequirements[0].ID = ""
		}, "product requirement id"},
		{"bad strength", func(p *ProjectIntent) {
			p.ArchitectureRequirements[0].Strength = "maybe"
		}, "unknown strength"},
		{"contradictory tech", func(p *ProjectIntent) {
			p.TechnicalConstraints = append(p.TechnicalConstraints,
				TechnicalConstraint{Kind: "must-not-use", Value: "tanstack"})
		}, "contradictory constraints"},
		{"unknown constraint kind", func(p *ProjectIntent) {
			p.TechnicalConstraints[0].Kind = "should-use"
		}, "unknown constraint kind"},
		{"no signal", func(p *ProjectIntent) {
			p.Surfaces = nil
			p.ArchitectureRequirements = nil
		}, "no architectural signal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := sampleIntent()
			tc.mutate(&in)
			err := in.Validate()
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				if _, ok := err.(*Error); !ok {
					t.Fatalf("expected typed *domain.Error, got %T", err)
				}
			}
		})
	}
}

func TestDefaultScopeIsRequiredNow(t *testing.T) {
	in := ProjectIntent{
		SchemaVersion: 1, Name: "x",
		Surfaces: []SurfaceIntent{{Kind: "Public-Web"}},
	}
	if err := in.Validate(); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
	got := in.RequiredSurfaces()
	if len(got) != 1 || got[0] != "public-web" {
		t.Fatalf("unexpected required surfaces: %v", got)
	}
}
