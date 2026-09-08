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
	// Typed reason channels carry target=value pairs, e.g.
	// "compatible with must-use framework=tanstack". The legacy
	// tech-only prefixes above are matched first so untyped constraints
	// keep their exact pre-H2 recovery.
	prefixMustUseTyped    = "compatible with must-use "
	prefixMustNotUseTyped = "excluded by must-not-use "
	// Database reason channels feed profile selection: must-use values
	// constrain the choice, preferences explain fallbacks.
	prefixDatabaseMustUse = "compatible with must-use database="
	prefixDatabasePrefer  = "database preference: "
	prefixDatabaseAvoid   = "database avoidance noted: must-not-use database="
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

// TechConstraint is a recovered technology constraint: an untyped legacy
// value (Target "") or an explicit target=value pair. Recovery parses the
// decision reason strings so the decision stays the single source of truth
// for composition.
type TechConstraint struct {
	Target string
	Value  string
}

// MustUseTech recovers explicit must-use technology values from winner
// reasons, in stable order. Both legacy ("tech V") and typed ("T=V")
// forms contribute their values.
func MustUseTech(decision domain.ArchitectureDecision) []string {
	out := collectPrefixed(decisionReasons(decision), prefixMustUse)
	for _, c := range MustUseConstraints(decision) {
		if c.Target != "" {
			out = append(out, c.Value)
		}
	}
	sort.Strings(out)
	return dedupStrings(out)
}

// MustNotUseTech recovers must-not-use values recorded anywhere in the
// candidate reasons, in stable order (legacy and typed forms).
func MustNotUseTech(decision domain.ArchitectureDecision) []string {
	out := collectPrefixed(allCandidateReasons(decision), prefixMustNotUse)
	for _, c := range MustNotUseConstraints(decision) {
		if c.Target != "" {
			out = append(out, c.Value)
		}
	}
	sort.Strings(out)
	return dedupStrings(out)
}

func dedupStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// MustUseConstraints recovers typed must-use constraints (target=value)
// plus legacy untyped values (Target "") from winner reasons.
func MustUseConstraints(decision domain.ArchitectureDecision) []TechConstraint {
	return collectTechConstraints(decisionReasons(decision), prefixMustUseTyped, prefixMustUse)
}

// MustNotUseConstraints recovers typed and legacy must-not-use
// constraints recorded anywhere in the candidate reasons.
func MustNotUseConstraints(decision domain.ArchitectureDecision) []TechConstraint {
	return collectTechConstraints(allCandidateReasons(decision), prefixMustNotUseTyped, prefixMustNotUse)
}

// collectTechConstraints parses legacy "tech V" values first (Target "")
// and then "T=V" pairs from the typed prefix. A typed-looking value
// without "=" is ignored: only the resolver mints these strings.
func collectTechConstraints(reasons []string, typedPrefix, legacyPrefix string) []TechConstraint {
	var out []TechConstraint
	seen := map[string]bool{}
	add := func(target, value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		target = strings.ToLower(strings.TrimSpace(target))
		if value == "" || seen[target+"\x00"+value] {
			return
		}
		seen[target+"\x00"+value] = true
		out = append(out, TechConstraint{Target: target, Value: value})
	}
	for _, r := range reasons {
		if rest, ok := strings.CutPrefix(r, legacyPrefix); ok {
			add("", rest)
			continue
		}
		if rest, ok := strings.CutPrefix(r, typedPrefix); ok {
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "tech ") {
				continue
			}
			if target, value, ok := strings.Cut(rest, "="); ok {
				add(target, value)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Target != out[j].Target {
			return out[i].Target < out[j].Target
		}
		return out[i].Value < out[j].Value
	})
	return out
}

// DatabaseMustUse recovers must-use database values from winner reasons.
func DatabaseMustUse(decision domain.ArchitectureDecision) []string {
	return collectPrefixed(decisionReasons(decision), prefixDatabaseMustUse)
}

// DatabasePreferences recovers prefer/avoid database hints ("kind=value")
// from winner reasons, in stable order.
func DatabasePreferences(decision domain.ArchitectureDecision) []domain.Preference {
	var out []domain.Preference
	seen := map[string]bool{}
	for _, r := range decisionReasons(decision) {
		rest, ok := strings.CutPrefix(r, prefixDatabasePrefer)
		if !ok {
			continue
		}
		kind, value, ok := strings.Cut(strings.TrimSpace(rest), "=")
		if !ok {
			continue
		}
		kind = strings.ToLower(strings.TrimSpace(kind))
		value = strings.ToLower(strings.TrimSpace(value))
		if (kind != "prefer" && kind != "avoid") || value == "" || seen[kind+"\x00"+value] {
			continue
		}
		seen[kind+"\x00"+value] = true
		out = append(out, domain.Preference{Kind: kind, Value: value})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Value < out[j].Value
	})
	return out
}

// DatabaseAvoidance recovers must-not-use database values from winner
// reasons, in stable order.
func DatabaseAvoidance(decision domain.ArchitectureDecision) []string {
	return collectPrefixed(decisionReasons(decision), prefixDatabaseAvoid)
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
// tech tags plus its adapter name (legacy untyped signal set).
func providerTech(b domain.Boilerplate) []string {
	return append(append([]string{}, b.TechTags...), b.Adapter)
}

// providerSignals returns the catalog technology signals of a boilerplate
// under one constraint target: the boilerplate identity (id, adapter) for
// provider, the technology[target] metadata plus tech_tags fallback for
// framework/language/runtime, and the legacy tech_tags plus adapter for
// untyped constraints (exact pre-H2 behavior).
func providerSignals(b domain.Boilerplate, target string) []string {
	switch target {
	case "provider":
		return append(append([]string{b.ID, b.Adapter}, b.TechTags...), b.Technology.Values(target)...)
	case "framework", "language", "runtime":
		return append(append([]string{}, b.Technology.Values(target)...), b.TechTags...)
	default:
		return providerTech(b)
	}
}

// constraintMatches reports whether a recovered tech constraint is
// satisfied by a boilerplate, using catalog signals only.
func constraintMatches(b domain.Boilerplate, c TechConstraint) bool {
	return techMatch(providerSignals(b, c.Target), c.Value)
}

// techEnforcedTarget reports whether a recovered constraint is enforced
// against providers. Database targets resolve through profile selection
// (SelectDatabaseProfileFor) and deployment targets have no catalog
// metadata; neither may fail or steer provider compatibility.
func techEnforcedTarget(target string) bool {
	switch target {
	case "", "framework", "language", "runtime", "provider":
		return true
	}
	return false
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
	mustUse := MustUseConstraints(decision)
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
		mi := countConstraintMatches(ranked[i], mustUse)
		mj := countConstraintMatches(ranked[j], mustUse)
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

func countConstraintMatches(b domain.Boilerplate, constraints []TechConstraint) int {
	n := 0
	for _, c := range constraints {
		if !techEnforcedTarget(c.Target) {
			continue
		}
		if constraintMatches(b, c) {
			n++
		}
	}
	return n
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
