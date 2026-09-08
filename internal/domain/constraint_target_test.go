package domain

import (
	"strings"
	"testing"
)

func targetIntent() ProjectIntent {
	return ProjectIntent{
		SchemaVersion: 1,
		Name:          "typed",
		Surfaces:      []SurfaceIntent{{Kind: "web-admin", Scope: ScopeRequiredNow}},
	}
}

// TestConstraintTargetValidation proves the H2 target vocabulary: the six
// documented targets validate, unknown targets fail with the list, and the
// legacy target-less shape still parses.
func TestConstraintTargetValidation(t *testing.T) {
	for _, target := range ValidConstraintTargets {
		t.Run("accept-"+target, func(t *testing.T) {
			in := targetIntent()
			in.TechnicalConstraints = []TechnicalConstraint{
				{Target: target, Kind: "must-use", Value: "x"},
			}
			if err := in.Validate(); err != nil {
				t.Fatalf("target %q rejected: %v", target, err)
			}
		})
	}
	t.Run("accept-empty-target-legacy", func(t *testing.T) {
		in := targetIntent()
		in.TechnicalConstraints = []TechnicalConstraint{{Kind: "must-use", Value: "tanstack"}}
		if err := in.Validate(); err != nil {
			t.Fatalf("legacy untyped constraint rejected: %v", err)
		}
	})
	t.Run("reject-unknown-target-with-list", func(t *testing.T) {
		in := targetIntent()
		in.TechnicalConstraints = []TechnicalConstraint{
			{Target: "brand", Kind: "must-use", Value: "x"},
		}
		err := in.Validate()
		if err == nil {
			t.Fatal("expected unknown-target error, got nil")
		}
		for _, want := range []string{`unknown constraint target "brand"`, "framework", "database", "provider"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error lacks %q: %v", want, err)
			}
		}
	})
}

// TestConstraintPreferAvoidKinds proves prefer/avoid are accepted inside
// technical_constraints (ranking-only; eligibility is decided downstream).
func TestConstraintPreferAvoidKinds(t *testing.T) {
	for _, kind := range []string{"prefer", "avoid"} {
		t.Run(kind, func(t *testing.T) {
			in := targetIntent()
			in.TechnicalConstraints = []TechnicalConstraint{
				{Target: "language", Kind: kind, Value: "python"},
			}
			if err := in.Validate(); err != nil {
				t.Fatalf("kind %q rejected: %v", kind, err)
			}
		})
	}
}

// TestConstraintContradictionIsTargetAware proves same-value constraints on
// different explicit targets are not contradictory, while an untyped value
// still collides with every typed same-value constraint.
func TestConstraintContradictionIsTargetAware(t *testing.T) {
	t.Run("different-targets-compatible", func(t *testing.T) {
		in := targetIntent()
		in.TechnicalConstraints = []TechnicalConstraint{
			{Target: "framework", Kind: "must-use", Value: "tanstack"},
			{Target: "language", Kind: "must-not-use", Value: "tanstack"},
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("distinct targets should not contradict: %v", err)
		}
	})
	t.Run("untyped-collides-with-typed", func(t *testing.T) {
		in := targetIntent()
		in.TechnicalConstraints = []TechnicalConstraint{
			{Kind: "must-use", Value: "tanstack"},
			{Target: "provider", Kind: "must-not-use", Value: "tanstack"},
		}
		err := in.Validate()
		if err == nil || !strings.Contains(err.Error(), "contradictory constraints") {
			t.Fatalf("expected contradiction, got %v", err)
		}
	})
}
