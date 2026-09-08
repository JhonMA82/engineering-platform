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

// typedLabel renders a typed constraint as target=value for reason
// strings, e.g. framework=tanstack.
func typedLabel(c domain.TechnicalConstraint) string {
	return fmt.Sprintf("%s=%s", strings.ToLower(strings.TrimSpace(c.Target)), strings.ToLower(strings.TrimSpace(c.Value)))
}

// typedTechMatch evaluates an explicit-target constraint against the
// catalog technology metadata for that target (see
// catalog.Index.RecipeTechnology). No technology names are hardcoded: the
// signals come from boilerplate technology entries and recipe tech_tags.
func typedTechMatch(r domain.Recipe, c domain.TechnicalConstraint, idx catalog.Index) bool {
	return techMatch(idx.RecipeTechnology(r, strings.ToLower(strings.TrimSpace(c.Target))), c.Value)
}

// recipeSatisfiesDatabase reports whether a recipe policy offers any
// profile identifying value (id, engine or provider match against the
// curated profiles). It answers from catalog data, never from a brand
// allow-list in Go.
func recipeSatisfiesDatabase(r domain.Recipe, value string, idx catalog.Index) bool {
	for _, id := range idx.RecipeDatabaseProfiles(r) {
		if p, ok := idx.DatabaseProfile(id); ok && p.MatchesTechnology(value) {
			return true
		}
	}
	return false
}

// appendDeploymentNotes records deployment-target constraints on eligible
// candidates. The catalog curates no deployment metadata yet, so per §2.3
// these constraints are not evaluated — the note keeps the explicit user
// decision visible instead of silently dropping it.
func appendDeploymentNotes(e *Eligibility, n Normalized) {
	seen := map[string]bool{}
	for _, c := range n.Intent.TechnicalConstraints {
		if strings.ToLower(strings.TrimSpace(c.Target)) != domain.ConstraintTargetDeployment {
			continue
		}
		note := fmt.Sprintf("deployment constraint %s=%s not evaluated (no catalog deployment metadata)",
			strings.ToLower(strings.TrimSpace(c.Kind)), strings.ToLower(strings.TrimSpace(c.Value)))
		if !seen[note] {
			seen[note] = true
			e.Positive = append(e.Positive, note)
		}
	}
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
		for _, mu := range sp.TypedMustUse {
			if !typedTechMatch(r, mu, idx) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("incompatible with must-use %s", typedLabel(mu)))
			}
		}
		for _, mn := range sp.TypedMustNotUse {
			if typedTechMatch(r, mn, idx) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("excluded by must-not-use %s", typedLabel(mn)))
			}
		}
		for _, db := range sp.DatabaseMustUse {
			// A database value with no curated profile anywhere is a
			// catalog gap, not a per-recipe verdict: every candidate
			// stays untouched on this axis so R7 can name the missing
			// database-profile foundation.
			if _, ok := idx.MatchDatabaseProfile(db); !ok {
				continue
			}
			if !recipeSatisfiesDatabase(r, db, idx) {
				e.Eligible = false
				e.Negative = append(e.Negative, fmt.Sprintf("no allowed database profile satisfies must-use database=%s", db))
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
			for _, mu := range sp.TypedMustUse {
				e.Positive = append(e.Positive, fmt.Sprintf("compatible with must-use %s", typedLabel(mu)))
			}
			for _, db := range sp.DatabaseMustUse {
				if _, ok := idx.MatchDatabaseProfile(db); ok {
					e.Positive = append(e.Positive, fmt.Sprintf("compatible with must-use database=%s", db))
				}
			}
			for _, p := range sp.DatabasePrefer {
				e.Positive = append(e.Positive, fmt.Sprintf("database preference: %s=%s", p.Kind, strings.ToLower(strings.TrimSpace(p.Value))))
			}
			for _, db := range sp.DatabaseMustNotUse {
				e.Positive = append(e.Positive, fmt.Sprintf("database avoidance noted: must-not-use database=%s", db))
			}
			appendDeploymentNotes(&e, n)
		}
		out = append(out, e)
	}
	return out
}
