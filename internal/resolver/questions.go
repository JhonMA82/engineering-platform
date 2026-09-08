package resolver

import (
	"fmt"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// ambiguousMargin is the winner-vs-#2 gap below which the resolver asks
// instead of guessing, provided a discriminating dimension exists.
const ambiguousMargin = 12

var dimensionOptions = map[string][]string{
	"public_access":     {"authenticated-only", "anonymous-public-access"},
	"shared_backend":    {"single-client", "shared-backend"},
	"offline":           {"online-only", "offline-operation"},
	"offline_operation": {"online-only", "offline-operation"},
}

// dimensionForRef maps an architectural ref to a neutral dimension name.
func dimensionForRef(ref string) string {
	switch strings.ToLower(ref) {
	case "anonymous-public-access":
		return "public_access"
	case "shared-backend":
		return "shared_backend"
	case "offline-operation":
		return "offline"
	default:
		return strings.ReplaceAll(strings.ToLower(ref), "-", "_")
	}
}

func optionsFor(dimension string) []string {
	if o, ok := dimensionOptions[dimension]; ok {
		return o
	}
	return []string{"required", "not-required"}
}

// Unresolved runs R10: possible-strength requirements always surface as
// neutral dimensions; explicit data unknowns join only on close calls.
func Unresolved(n Normalized, sp Split, scored []Scored, margin int) []domain.UnresolvedDimension {
	var dims []domain.UnresolvedDimension
	seen := map[string]bool{}
	add := func(dimension, reason string) {
		if seen[dimension] {
			return
		}
		seen[dimension] = true
		dims = append(dims, domain.UnresolvedDimension{
			Dimension: dimension,
			Reason:    reason,
			Options:   optionsFor(dimension),
		})
	}
	top := ""
	second := ""
	for i, s := range scored {
		if !s.Eligible {
			continue
		}
		if top == "" {
			top = s.Recipe
		} else if second == "" {
			second = s.Recipe
			break
		}
		_ = i
	}
	reason := fmt.Sprintf("distinguishes %s from %s", top, second)
	if second == "" {
		reason = fmt.Sprintf("unconfirmed scope for %s", top)
	}
	for _, r := range sp.ArchPossible {
		add(dimensionForRef(r.Ref), reason)
	}
	if margin < ambiguousMargin && strings.ToLower(n.Intent.Data.PublicAccess) == "unknown" {
		add("public_access", reason)
	}
	return dims
}
