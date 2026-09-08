package composer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// SelectDatabaseProfile resolves the recipe database policy against the
// catalog profiles. The default profile wins; explicit per-constraint
// database selection is future work documented in composition.md.
func SelectDatabaseProfile(recipe domain.Recipe, cat catalog.Catalog) (string, error) {
	profile, _, err := SelectDatabaseProfileFor(recipe, cat, nil, nil, nil)
	return profile, err
}

// SelectDatabaseProfileFor resolves the recipe database policy against the
// catalog profiles while honoring database-target constraints recovered
// from the decision (§2.4):
//
//   - must-use database=X selects the allowed profile identifying X (id,
//     engine or provider match against curated profiles), even when it is
//     not the recipe default. No allowed profile identifies X: typed
//     composition error (a catalog-wide miss is a resolver catalog-gap
//     before composition is ever reached).
//   - must-not-use database=X avoids that profile when an alternative
//     allowed profile exists; otherwise the default stands with a note.
//   - prefer database=X selects the allowed profile identifying X when one
//     exists; otherwise the default stands and the returned note explains
//     the deviation. avoid database=X prefers any other allowed profile.
//     Preferences never fail selection.
//
// The returned note is empty when the default applies without deviation.
// Matching runs against catalog profile data only — no brands in Go.
func SelectDatabaseProfileFor(recipe domain.Recipe, cat catalog.Catalog, mustUse []string, mustNotUse []string, prefer []domain.Preference) (profile, note string, err error) {
	idx := catalog.NewIndex(cat)
	known := map[string]bool{}
	for _, p := range cat.DatabaseProfiles {
		known[p.ID] = true
	}
	policy := recipe.DatabasePolicy
	if policy.DefaultProfile == "" {
		return "", "", domain.Composition("recipe " + recipe.ID + " declares no default database profile")
	}
	if !known[policy.DefaultProfile] {
		return "", "", domain.Composition(
			"recipe " + recipe.ID + " defaults to unknown database profile " + policy.DefaultProfile)
	}
	for _, p := range policy.AllowedProfiles {
		if !known[p] {
			return "", "", domain.Composition(
				"recipe " + recipe.ID + " allows unknown database profile " + p)
		}
	}
	if len(policy.AllowedProfiles) > 0 {
		allowed := false
		for _, p := range policy.AllowedProfiles {
			if p == policy.DefaultProfile {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", "", domain.Composition(
				"recipe " + recipe.ID + " default profile " + policy.DefaultProfile + " is not allowed")
		}
	}
	allowed := idx.RecipeDatabaseProfiles(recipe)
	profileFor := func(value string) (domain.DatabaseProfile, bool) {
		for _, id := range allowed {
			if p, ok := idx.DatabaseProfile(id); ok && p.MatchesTechnology(value) {
				return p, true
			}
		}
		return domain.DatabaseProfile{}, false
	}
	for _, v := range mustUse {
		p, ok := profileFor(v)
		if !ok {
			return "", "", domain.Composition(fmt.Sprintf(
				"recipe %s offers no allowed database profile for must-use database=%s",
				recipe.ID, v))
		}
		if p.ID == policy.DefaultProfile {
			return p.ID, "", nil
		}
		return p.ID, fmt.Sprintf("selected database profile %s per must-use database=%s", p.ID, v), nil
	}
	avoided := map[string]bool{}
	for _, v := range mustNotUse {
		if p, ok := profileFor(v); ok {
			avoided[p.ID] = true
		}
	}
	for _, p := range prefer {
		v := strings.ToLower(strings.TrimSpace(p.Value))
		switch strings.ToLower(strings.TrimSpace(p.Kind)) {
		case "prefer":
			if match, ok := profileFor(v); ok {
				if match.ID == policy.DefaultProfile {
					return match.ID, "", nil
				}
				return match.ID, fmt.Sprintf(
					"selected database profile %s per prefer database=%s", match.ID, v), nil
			}
			return policy.DefaultProfile, fmt.Sprintf(
				"prefer database=%s has no curated profile in recipe %s; using default %s",
				v, recipe.ID, policy.DefaultProfile), nil
		case "avoid":
			for _, id := range allowed {
				if id == policy.DefaultProfile {
					continue
				}
				if alt, ok := idx.DatabaseProfile(id); ok && !avoided[id] && !alt.MatchesTechnology(v) {
					return id, fmt.Sprintf(
						"selected database profile %s per avoid database=%s", id, v), nil
				}
			}
		}
	}
	if !avoided[policy.DefaultProfile] {
		return policy.DefaultProfile, "", nil
	}
	for _, id := range allowed {
		if !avoided[id] {
			return id, fmt.Sprintf(
				"selected database profile %s per must-not-use database (default %s avoided)",
				id, policy.DefaultProfile), nil
		}
	}
	return policy.DefaultProfile, fmt.Sprintf(
		"must-not-use database leaves recipe %s no alternative; keeping default %s",
		recipe.ID, policy.DefaultProfile), nil
}

// CheckCompatibility verifies the composed components against the recipe
// contract and the explicit technical constraints recovered from the
// decision: surface coverage, allowed surface composition (over the full
// component surface set), architectural capability coverage by the chosen
// providers, database policy versus catalog profiles, runtime/tech
// compatibility between components, and no provider violating an explicit
// must-not-use constraint. Every failure is a typed composition error.
func CheckCompatibility(decision domain.ArchitectureDecision, recipe domain.Recipe, comp Composition, caps []domain.CapabilityID, mustUse, mustNotUse []TechConstraint, cat catalog.Catalog) error {
	if decision.Selected != nil && decision.Selected.RecipeVersion != "" &&
		recipe.Version != "" && decision.Selected.RecipeVersion != recipe.Version {
		return domain.Composition(fmt.Sprintf(
			"decision targets %s %s but catalog holds %s",
			recipe.ID, decision.Selected.RecipeVersion, recipe.Version))
	}
	provides := map[domain.SurfaceID]bool{}
	for _, s := range recipe.Provides.Surfaces {
		provides[s] = true
	}
	set := map[domain.SurfaceID]bool{}
	for _, c := range comp.Components {
		set[c.Surface] = true
		if !provides[c.Surface] {
			return domain.Composition(
				"surface " + string(c.Surface) + " is not covered by recipe " + recipe.ID)
		}
	}
	if len(recipe.AllowedSurfaceComposition) > 0 && !compositionAllowed(recipe, set) {
		surfaces := make([]string, 0, len(set))
		for s := range set {
			surfaces = append(surfaces, string(s))
		}
		sort.Strings(surfaces)
		return domain.Composition(
			"recipe " + recipe.ID + " does not allow composing surfaces [" +
				joinComma(surfaces) + "]")
	}
	covered := map[domain.CapabilityID]bool{}
	idx := catalog.NewIndex(cat)
	for _, c := range comp.Components {
		b, ok := idx.Boilerplate(c.Boilerplate)
		if !ok {
			return domain.Composition("composed unknown boilerplate " + c.Boilerplate)
		}
		for _, cp := range b.Provides.Capabilities {
			covered[cp] = true
		}
	}
	for _, cp := range caps {
		if !covered[cp] {
			return domain.Composition(
				"architectural capability " + string(cp) + " is covered by no chosen provider")
		}
	}
	byProvider := map[string]domain.Boilerplate{}
	for _, c := range comp.Components {
		b, _ := idx.Boilerplate(c.Boilerplate)
		byProvider[c.Boilerplate] = b
	}
	for _, mn := range mustNotUse {
		if !techEnforcedTarget(mn.Target) {
			continue
		}
		for bp, b := range byProvider {
			if constraintMatches(b, mn) {
				label := mn.Value
				if mn.Target != "" {
					label = mn.Target + "=" + mn.Value
				}
				return domain.Composition(
					"provider " + bp + " violates must-not-use tech " + label)
			}
		}
	}
	for _, mu := range mustUse {
		if !techEnforcedTarget(mu.Target) {
			continue
		}
		coveredBy := false
		if mu.Target == "" && techMatch(recipe.TechTags, mu.Value) {
			coveredBy = true
		}
		if !coveredBy && mu.Target != "" {
			for _, id := range recipe.PrimaryBoilerplates {
				if b, ok := idx.Boilerplate(id); ok && constraintMatches(b, mu) {
					coveredBy = true
					break
				}
			}
			if !coveredBy {
				for _, b := range byProvider {
					if constraintMatches(b, mu) {
						coveredBy = true
						break
					}
				}
			}
		}
		if !coveredBy {
			for _, b := range byProvider {
				if constraintMatches(b, mu) {
					coveredBy = true
					break
				}
			}
		}
		if !coveredBy {
			label := mu.Value
			if mu.Target != "" {
				label = mu.Target + "=" + mu.Value
			}
			return domain.Composition(
				"explicit must-use tech " + label + " is satisfied by no chosen component")
		}
	}
	return nil
}

// compositionAllowed mirrors the resolver subset semantics: some allowed
// combo must contain every composed surface.
func compositionAllowed(recipe domain.Recipe, set map[domain.SurfaceID]bool) bool {
	for _, combo := range recipe.AllowedSurfaceComposition {
		in := map[domain.SurfaceID]bool{}
		for _, s := range combo {
			in[s] = true
		}
		ok := true
		for s := range set {
			if !in[s] {
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

func joinComma(items []string) string {
	out := ""
	for i, it := range items {
		if i > 0 {
			out += ", "
		}
		out += it
	}
	return out
}
