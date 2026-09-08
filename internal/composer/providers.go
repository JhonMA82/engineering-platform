package composer

import (
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Decision states on the PRD decision axis that may serve a surface.
var eligibleDecisionStates = map[string]bool{
	"curated": true, "default": true, "alternative": true,
	"specialized": true, "reference": true,
}

// Delivery states on the PRD delivery axis that count as available.
var eligibleDeliveryStates = map[string]bool{
	"stable": true, "curated": true, "pilot-ready": true, "released": true,
}

// Eligible reports whether a boilerplate may serve a surface: curated
// decision state, available delivery state, pin and adapter present.
// Deprecated, rejected and experimental foundations are never eligible.
func Eligible(b domain.Boilerplate) bool {
	if !eligibleDecisionStates[b.DecisionStatus] {
		return false
	}
	if !eligibleDeliveryStates[b.DeliveryStatus] {
		return false
	}
	return b.Pin != "" && b.Adapter != ""
}

// EligibleProviders returns the eligible boilerplates providing a surface,
// sorted by id for determinism.
func EligibleProviders(cat catalog.Catalog, surface domain.SurfaceID) []domain.Boilerplate {
	var out []domain.Boilerplate
	for _, b := range cat.Boilerplates {
		if !Eligible(b) {
			continue
		}
		for _, s := range b.Provides.Surfaces {
			if s == surface {
				out = append(out, b)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
