package composer

import (
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Reason prefixes emitted by the resolver. The composer recovers the
// required surfaces, capabilities and technical constraints from these
// reason strings so the decision stays the single source of truth:
// required work is never invented, and product-feature mentions
// ("includes product feature: ...") are deliberately ignored.
const (
	prefixRequiredSurface = "covers required surface "
	prefixCapability      = "covers capability "
	prefixDerived         = "covers derived requirement "
	prefixMustUse         = "compatible with must-use tech "
	prefixMustNotUse      = "excluded by must-not-use tech "
)

// decisionReasons returns the winner reasons plus the selected candidate
// positives, deduplicated, so hand-built decisions that only fill one side
// still parse.
func decisionReasons(decision domain.ArchitectureDecision) []string {
	seen := map[string]bool{}
	var out []string
	add := func(reasons []string) {
		for _, r := range reasons {
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	add(decision.Reasons)
	if decision.Selected != nil {
		for _, c := range decision.Candidates {
			if c.Recipe == decision.Selected.Recipe {
				add(c.PositiveReasons)
			}
		}
	}
	return out
}

// allCandidateReasons scans every candidate, used for must-not-use values
// that only losing candidates record.
func allCandidateReasons(decision domain.ArchitectureDecision) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range decision.Candidates {
		for _, r := range append(append([]string{}, c.PositiveReasons...), c.NegativeReasons...) {
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// RequiredSurfaces recovers the required_now surfaces from the decision.
// It parses "covers required surface X" reasons and "covers capability X"
// reasons whose ref is a known surface, in stable order.
func RequiredSurfaces(decision domain.ArchitectureDecision, idx catalog.Index) []domain.SurfaceID {
	seen := map[domain.SurfaceID]bool{}
	for _, r := range decisionReasons(decision) {
		if rest, ok := strings.CutPrefix(r, prefixRequiredSurface); ok {
			id := domain.SurfaceID(strings.ToLower(strings.TrimSpace(rest)))
			if id != "" {
				seen[id] = true
			}
			continue
		}
		if rest, ok := strings.CutPrefix(r, prefixCapability); ok {
			id := domain.SurfaceID(strings.ToLower(strings.TrimSpace(rest)))
			if id != "" && idx.SurfaceKnown(id) {
				seen[id] = true
			}
		}
	}
	out := make([]domain.SurfaceID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// RequiredCapabilities recovers the architectural capabilities the
// composition must cover: derived requirements plus required refs that
// name a known capability. "possible" refs never appear in winner reasons
// and therefore never constrain composition.
func RequiredCapabilities(decision domain.ArchitectureDecision, idx catalog.Index) []domain.CapabilityID {
	seen := map[domain.CapabilityID]bool{}
	for _, d := range decision.DerivedRequirements {
		id := domain.CapabilityID(strings.ToLower(strings.TrimSpace(d.ID)))
		if id != "" {
			seen[id] = true
		}
	}
	for _, r := range decisionReasons(decision) {
		if rest, ok := strings.CutPrefix(r, prefixDerived); ok {
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				seen[domain.CapabilityID(strings.ToLower(fields[0]))] = true
			}
			continue
		}
		if rest, ok := strings.CutPrefix(r, prefixCapability); ok {
			id := domain.CapabilityID(strings.ToLower(strings.TrimSpace(rest)))
			if id != "" && idx.CapabilityKnown(id) {
				seen[id] = true
			}
		}
	}
	out := make([]domain.CapabilityID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// MustUseTech recovers explicit must-use technology values from winner
// reasons, in stable order.
func MustUseTech(decision domain.ArchitectureDecision) []string {
	return collectPrefixed(decisionReasons(decision), prefixMustUse)
}

// MustNotUseTech recovers must-not-use values recorded anywhere in the
// candidate reasons, in stable order.
func MustNotUseTech(decision domain.ArchitectureDecision) []string {
	return collectPrefixed(allCandidateReasons(decision), prefixMustNotUse)
}

func collectPrefixed(reasons []string, prefix string) []string {
	seen := map[string]bool{}
	for _, r := range reasons {
		if rest, ok := strings.CutPrefix(r, prefix); ok {
			v := strings.ToLower(strings.TrimSpace(rest))
			if v != "" {
				seen[v] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// techMatch mirrors the resolver semantics: a value matches when either
// side contains the other, case-insensitively.
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

// providerTech returns the technology signals of a boilerplate: its
// tech tags plus its adapter name.
func providerTech(b domain.Boilerplate) []string {
	return append(append([]string{}, b.TechTags...), b.Adapter)
}

// capabilityCoverage counts how many of the required capabilities a
// boilerplate provides.
func capabilityCoverage(b domain.Boilerplate, caps []domain.CapabilityID) int {
	have := map[domain.CapabilityID]bool{}
	for _, c := range b.Provides.Capabilities {
		have[c] = true
	}
	n := 0
	for _, c := range caps {
		if have[c] {
			n++
		}
	}
	return n
}

// SelectProvider picks the best eligible boilerplate for a surface.
// Ranking is deterministic: recipe primary boilerplates first (catalog
// order), then must-use technology fit, then required-capability coverage,
// then lexicographically smallest id.
func SelectProvider(surface domain.SurfaceID, recipe domain.Recipe, decision domain.ArchitectureDecision, cat catalog.Catalog, caps []domain.CapabilityID) (domain.Boilerplate, error) {
	candidates := EligibleProviders(cat, surface)
	if len(candidates) == 0 {
		return domain.Boilerplate{}, domain.Composition(
			"no eligible provider for surface " + string(surface) + " in recipe " + recipe.ID)
	}
	primary := map[string]int{}
	for i, id := range recipe.PrimaryBoilerplates {
		if _, ok := primary[id]; !ok {
			primary[id] = i
		}
	}
	mustUse := MustUseTech(decision)
	ranked := append([]domain.Boilerplate{}, candidates...)
	sort.SliceStable(ranked, func(i, j int) bool {
		pi, iPrimary := primary[ranked[i].ID]
		pj, jPrimary := primary[ranked[j].ID]
		if iPrimary != jPrimary {
			return iPrimary
		}
		if iPrimary && jPrimary && pi != pj {
			return pi < pj
		}
		mi := countTechMatches(providerTech(ranked[i]), mustUse)
		mj := countTechMatches(providerTech(ranked[j]), mustUse)
		if mi != mj {
			return mi > mj
		}
		ci := capabilityCoverage(ranked[i], caps)
		cj := capabilityCoverage(ranked[j], caps)
		if ci != cj {
			return ci > cj
		}
		return ranked[i].ID < ranked[j].ID
	})
	return ranked[0], nil
}

func countTechMatches(tags, values []string) int {
	n := 0
	for _, v := range values {
		if techMatch(tags, v) {
			n++
		}
	}
	return n
}
