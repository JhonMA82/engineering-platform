package resolver

import (
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Split is the R3 output. Product features are carried along but never
// become hard constraints by construction.
type Split struct {
	Product      []domain.ProductRequirement
	ArchRequired []domain.ArchitectureRequirement
	ArchPossible []domain.ArchitectureRequirement
	// MustUse/MustNotUse carry legacy untyped values, matched against
	// recipe tech_tags exactly as before H2 (see the backward-compat rule
	// on domain.TechnicalConstraint).
	MustUse    []string
	MustNotUse []string
	// TypedMustUse/TypedMustNotUse carry explicit-target constraints for
	// framework, language, runtime and provider, matched against the
	// catalog technology metadata for that target (see
	// catalog.Index.RecipeTechnology). Database targets flow through
	// DatabaseMustUse/DatabasePrefer instead; deployment targets have no
	// catalog metadata yet and are not evaluated (recorded in reasons).
	TypedMustUse    []domain.TechnicalConstraint
	TypedMustNotUse []domain.TechnicalConstraint
	// DatabaseMustUse holds must-use database values resolved against
	// curated profiles at composition time. DatabasePrefer holds
	// prefer/avoid database values, which never affect eligibility.
	// DatabaseMustNotUse holds must-not-use database values, which never
	// eliminate recipes; the composer avoids those profiles when the
	// recipe policy offers an alternative.
	DatabaseMustUse    []string
	DatabasePrefer     []domain.Preference
	DatabaseMustNotUse []string
	// Preferences holds ranking-only signals: the preferences[] array plus
	// prefer/avoid technical constraints (any non-database target), which
	// behave identically by construction.
	Preferences []domain.Preference
}

// SplitRequirements runs R3 classification.
func SplitRequirements(n Normalized) Split {
	sp := Split{Product: n.Intent.ProductRequirements, Preferences: n.Intent.Preferences}
	for _, r := range n.Intent.ArchitectureRequirements {
		if r.EffectiveStrength() == "possible" {
			sp.ArchPossible = append(sp.ArchPossible, r)
		} else {
			sp.ArchRequired = append(sp.ArchRequired, r)
		}
	}
	seenUse, seenNot := map[string]bool{}, map[string]bool{}
	seenTypedUse, seenTypedNot := map[string]bool{}, map[string]bool{}
	seenDBUse, seenDBNot, seenDBPref := map[string]bool{}, map[string]bool{}, map[string]bool{}
	addTyped := func(dst *[]domain.TechnicalConstraint, seen map[string]bool, c domain.TechnicalConstraint, target, value string) {
		key := target + "\x00" + value
		if seen[key] {
			return
		}
		seen[key] = true
		*dst = append(*dst, domain.TechnicalConstraint{Target: target, Kind: c.Kind, Value: value})
	}
	for _, c := range n.Intent.TechnicalConstraints {
		target := strings.ToLower(strings.TrimSpace(c.Target))
		v := strings.ToLower(strings.TrimSpace(c.Value))
		switch c.Kind {
		case "must-use":
			switch {
			case target == domain.ConstraintTargetDatabase:
				if !seenDBUse[v] {
					seenDBUse[v] = true
					sp.DatabaseMustUse = append(sp.DatabaseMustUse, v)
				}
			case target == domain.ConstraintTargetDeployment:
				// No catalog metadata: recorded in reasons, not evaluated.
			case target == "":
				if !seenUse[v] {
					seenUse[v] = true
					sp.MustUse = append(sp.MustUse, v)
				}
			default:
				addTyped(&sp.TypedMustUse, seenTypedUse, c, target, v)
			}
		case "must-not-use":
			switch {
			case target == domain.ConstraintTargetDatabase:
				if !seenDBNot[v] {
					seenDBNot[v] = true
					sp.DatabaseMustNotUse = append(sp.DatabaseMustNotUse, v)
				}
			case target == domain.ConstraintTargetDeployment:
				// No catalog metadata: recorded in reasons, not evaluated.
			case target == "":
				if !seenNot[v] {
					seenNot[v] = true
					sp.MustNotUse = append(sp.MustNotUse, v)
				}
			default:
				addTyped(&sp.TypedMustNotUse, seenTypedNot, c, target, v)
			}
		case "prefer", "avoid":
			if target == domain.ConstraintTargetDatabase {
				key := c.Kind + "\x00" + v
				if !seenDBPref[key] {
					seenDBPref[key] = true
					sp.DatabasePrefer = append(sp.DatabasePrefer, domain.Preference{Kind: c.Kind, Value: v})
				}
				continue
			}
			sp.Preferences = append(sp.Preferences, domain.Preference{Kind: c.Kind, Value: v})
		}
	}
	sort.Strings(sp.MustUse)
	sort.Strings(sp.MustNotUse)
	sort.Slice(sp.TypedMustUse, func(i, j int) bool {
		if sp.TypedMustUse[i].Target != sp.TypedMustUse[j].Target {
			return sp.TypedMustUse[i].Target < sp.TypedMustUse[j].Target
		}
		return sp.TypedMustUse[i].Value < sp.TypedMustUse[j].Value
	})
	sort.Slice(sp.TypedMustNotUse, func(i, j int) bool {
		if sp.TypedMustNotUse[i].Target != sp.TypedMustNotUse[j].Target {
			return sp.TypedMustNotUse[i].Target < sp.TypedMustNotUse[j].Target
		}
		return sp.TypedMustNotUse[i].Value < sp.TypedMustNotUse[j].Value
	})
	sort.Strings(sp.DatabaseMustUse)
	sort.Strings(sp.DatabaseMustNotUse)
	sort.Slice(sp.DatabasePrefer, func(i, j int) bool {
		if sp.DatabasePrefer[i].Kind != sp.DatabasePrefer[j].Kind {
			return sp.DatabasePrefer[i].Kind < sp.DatabasePrefer[j].Kind
		}
		return sp.DatabasePrefer[i].Value < sp.DatabasePrefer[j].Value
	})
	return sp
}
