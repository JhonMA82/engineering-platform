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
	MustUse      []string
	MustNotUse   []string
	Preferences  []domain.Preference
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
	for _, c := range n.Intent.TechnicalConstraints {
		v := strings.ToLower(strings.TrimSpace(c.Value))
		switch c.Kind {
		case "must-use":
			if !seenUse[v] {
				seenUse[v] = true
				sp.MustUse = append(sp.MustUse, v)
			}
		case "must-not-use":
			if !seenNot[v] {
				seenNot[v] = true
				sp.MustNotUse = append(sp.MustNotUse, v)
			}
		}
	}
	sort.Strings(sp.MustUse)
	sort.Strings(sp.MustNotUse)
	return sp
}
