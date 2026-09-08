package domain

import (
	"strings"
	"testing"
)

func TestBoilerplateUpdateStrategy(t *testing.T) {
	base := Boilerplate{
		ID:             "fixture",
		Pin:            "v1",
		Adapter:        "fixture",
		Source:         SourceSpec{Type: "local", Path: "/tmp/x"},
		DeliveryStatus: "stable",
		DecisionStatus: "curated",
	}
	tests := []struct {
		name     string
		strategy string
		wantErr  string
		wantEff  string
	}{
		{"absent defaults to manual", "", "", "manual"},
		{"replace accepted", "replace", "", "replace"},
		{"merge-seed accepted", "merge-seed", "", "merge-seed"},
		{"fork-track accepted", "fork-track", "", "fork-track"},
		{"manual accepted", "manual", "", "manual"},
		{"unknown rejected", "auto", "unknown update_strategy", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := base
			bp.UpdateStrategy = tt.strategy
			err := bp.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if got := bp.EffectiveUpdateStrategy(); got != tt.wantEff {
				t.Errorf("EffectiveUpdateStrategy = %q, want %q", got, tt.wantEff)
			}
		})
	}
}
