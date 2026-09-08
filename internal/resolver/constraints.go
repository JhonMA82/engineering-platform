package resolver

import (
	"fmt"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Eligibility is the R6 per-candidate verdict with reasons.
type Eligibility struct {
	Recipe   domain.Recipe
	Eligible bool
	Positive []string
	Negative []string
}

func techMatch(tags []string, value string) bool {
	v := strings.ToLower(value)
	for _, t := range tags {
		tag := strings.ToLower(t)
		if tag == v || strings.Contains(tag, v) || strings.Contains(v, tag) {
			return true
		}
	}
	return false
}

func providesSurface(r domain.Recipe, id domain.SurfaceID) bool {
	for _, s := range r.Provides.Surfaces {
		if s == id {
			return true
		}
	}
	return false
}

func providesCapability(r domain.Recipe, id domain.CapabilityID) bool {
	for _, c := range r.Provides.Capabilities {
		if c == id {
			return true
		}
	}
	return false
}

func compositionAllows(r domain.Recipe, required []domain.SurfaceID) bool {
	if len(r.AllowedSurfaceComposition) == 0 {
		return true
	}
	for _, combo := range r.AllowedSurfaceComposition {
		set := map[domain.SurfaceID]bool{}
		for _, s := range combo {
			set[s] = true
		}
		ok := true
		for _, req := range required {
			if !set[req] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// ApplyConstraints runs R6 hard architectural constraints. Product features
// are never consulted here, so a boilerplate missing PDF stays eligible.
func ApplyConstraints(n Normalized, sp Split, derived []domain.DerivedRequirement, recs []domain.Recipe, idx catalog.Index) []Eligibility {
	derivedByID := map[string]string{}
	for _, d := range derived {
		derivedByID[d.ID] = d.Source
	}
	out := make([]Eligibility, 0, len(recs))
	for _, r := range recs {
		e := Eligibility{Recipe: r, Eligible: true}
		for _, s := range n.RequiredSurfaces {
			if !providesSurface(r, s) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("missing required surface %s", s))
			}
		}
		if e.Eligible && !compositionAllows(r, n.RequiredSurfaces) {
			e.Eligible = false
			e.Negative = append(e.Negative, "required surfaces cannot be composed by this recipe")
		}
		for _, ref := range n.RequiredRefs {
			cap := domain.CapabilityID(ref)
			surf := domain.SurfaceID(ref)
			switch {
			case idx.CapabilityKnown(cap):
				if !providesCapability(r, cap) {
					e.Eligible = false
					e.Negative = append(e.Negative, fmt.Sprintf("missing capability %s", cap))
				}
			case idx.SurfaceKnown(surf):
				if !providesSurface(r, surf) {
					e.Eligible = false
					e.Negative = append(e.Negative, fmt.Sprintf("missing required surface %s", surf))
				}
			default:
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("no catalog foundation covers %s", ref))
			}
		}
		for _, d := range derived {
			cap := domain.CapabilityID(d.ID)
			if !providesCapability(r, cap) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("missing derived requirement %s", d.ID))
			}
		}
		for _, mu := range sp.MustUse {
			if !techMatch(r.TechTags, mu) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("incompatible with must-use tech %s", mu))
			}
		}
		for _, mn := range sp.MustNotUse {
			if techMatch(r.TechTags, mn) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("excluded by must-not-use tech %s", mn))
			}
		}
		if e.Eligible {
			for _, s := range n.RequiredSurfaces {
				e.Positive = append(e.Positive, fmt.Sprintf("covers required surface %s", s))
			}
			for _, ref := range n.RequiredRefs {
				e.Positive = append(e.Positive, fmt.Sprintf("covers capability %s", ref))
			}
			for _, d := range derived {
				e.Positive = append(e.Positive, fmt.Sprintf("covers derived requirement %s (%s)", d.ID, d.Source))
			}
			for _, mu := range sp.MustUse {
				e.Positive = append(e.Positive, fmt.Sprintf("compatible with must-use tech %s", mu))
			}
		}
		out = append(out, e)
	}
	return out
}
