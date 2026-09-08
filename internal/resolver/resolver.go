// Package resolver implements the pure in-memory R1-R10 pipeline.
//
// Deterministic: stable sorts, sha256 fingerprint of the normalized intent,
// no time/rand/net/LLM/filesystem access.
package resolver

import (
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// ValidateIntent runs R1 structural validation.
func ValidateIntent(intent domain.ProjectIntent) error {
	return intent.Validate()
}

// Candidates runs R5: every active recipe is a candidate. Routing never
// starts from a project_type assumption.
func Candidates(idx catalog.Index) []domain.Recipe {
	return idx.ActiveRecipes()
}

// Resolve orchestrates R1-R10 and returns an ArchitectureDecision.
func Resolve(intent domain.ProjectIntent, cat catalog.Catalog) domain.ArchitectureDecision {
	idx := catalog.NewIndex(cat)
	if err := ValidateIntent(intent); err != nil {
		return domain.ArchitectureDecision{
			SchemaVersion:     1,
			Status:            domain.StatusInvalid,
			IntentFingerprint: FingerprintRaw(intent),
			Reasons:           []string{err.Error()},
			Confidence:        domain.Confidence{Level: "low"},
		}
	}
	n := Normalize(intent, idx)
	sp := SplitRequirements(n)
	derived := Derive(n, sp)
	recs := Candidates(idx)
	elig := ApplyConstraints(n, sp, derived, recs, idx)
	scored := ScoreAll(n, sp, derived, elig, idx)
	eligible := filterEligible(scored)
	if len(eligible) == 0 {
		return decideBlocked(n, sp, derived, scored, idx)
	}
	winner := eligible[0]
	runner := 0
	if len(eligible) > 1 {
		runner = eligible[1].Score
	}
	margin := winner.Score - runner
	dims := Unresolved(n, sp, scored, margin)
	status := domain.StatusResolved
	level := ConfidenceLevel(winner.Score, margin, len(dims))
	if len(dims) > 0 && margin < ambiguousMargin {
		status = domain.StatusAmbiguous
		level = "low"
	}
	decision := domain.ArchitectureDecision{
		SchemaVersion:        1,
		Status:               status,
		IntentFingerprint:    FingerprintNormalized(n),
		Selected:             &domain.SelectedRecipe{Recipe: winner.Recipe, RecipeVersion: winner.Version, Score: winner.Score},
		Confidence:           domain.Confidence{Level: level, Margin: margin},
		Reasons:              winner.Positive,
		Candidates:           toCandidates(scored),
		DerivedRequirements:  derived,
		UnresolvedDimensions: dims,
	}
	return decision
}

func filterEligible(scored []Scored) []Scored {
	var out []Scored
	for _, s := range scored {
		if s.Eligible {
			out = append(out, s)
		}
	}
	return out
}

// blockedHeadline names the dominant rejection signal instead of always
// blaming technical constraints: tech rejections mention must-use / use tech
// constraint vocabulary, anything else is surface-composition coverage.
func blockedHeadline(scored []Scored) string {
	for _, s := range scored {
		for _, neg := range s.Negative {
			if strings.Contains(neg, "must-use") || strings.Contains(neg, "must-not-use") {
				return "architecture is catalog-covered but hard technical constraints eliminate every candidate"
			}
		}
	}
	return "architecture is catalog-covered but no recipe composes the required surfaces"
}

func toCandidates(scored []Scored) []domain.Candidate {
	out := make([]domain.Candidate, 0, len(scored))
	for _, s := range scored {
		out = append(out, domain.Candidate{
			Recipe:          s.Recipe,
			Eligible:        s.Eligible,
			Score:           s.Score,
			PositiveReasons: s.Positive,
			NegativeReasons: s.Negative,
		})
	}
	return out
}

func decideBlocked(n Normalized, sp Split, derived []domain.DerivedRequirement, scored []Scored, idx catalog.Index) domain.ArchitectureDecision {
	missing := DiagnoseGap(n, sp, derived, idx)
	status := domain.StatusUnsupported
	if len(missing) > 0 {
		status = domain.StatusCatalogGap
	}
	reasons := []string{}
	if status == domain.StatusCatalogGap {
		for _, m := range missing {
			reasons = append(reasons, "missing architectural foundation for "+m.Kind+"="+m.Ref)
		}
	} else {
		reasons = append(reasons, blockedHeadline(scored))
		seen := map[string]bool{}
		for _, s := range scored {
			for _, neg := range s.Negative {
				if !seen[neg] {
					seen[neg] = true
					reasons = append(reasons, neg)
				}
			}
		}
	}
	return domain.ArchitectureDecision{
		SchemaVersion:       1,
		Status:              status,
		IntentFingerprint:   FingerprintNormalized(n),
		Confidence:          domain.Confidence{Level: "low"},
		Reasons:             reasons,
		Candidates:          toCandidates(scored),
		DerivedRequirements: derived,
		MissingArchitecture: missing,
	}
}
