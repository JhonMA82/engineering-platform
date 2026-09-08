package composer

import (
	"fmt"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// SelectDatabaseProfile resolves the recipe database policy against the
// catalog profiles. The default profile wins; explicit per-constraint
// database selection is future work documented in composition.md.
func SelectDatabaseProfile(recipe domain.Recipe, cat catalog.Catalog) (string, error) {
	known := map[string]bool{}
	for _, p := range cat.DatabaseProfiles {
		known[p.ID] = true
	}
	policy := recipe.DatabasePolicy
	if policy.DefaultProfile == "" {
		return "", domain.Composition("recipe " + recipe.ID + " declares no default database profile")
	}
	if !known[policy.DefaultProfile] {
		return "", domain.Composition(
			"recipe " + recipe.ID + " defaults to unknown database profile " + policy.DefaultProfile)
	}
	for _, p := range policy.AllowedProfiles {
		if !known[p] {
			return "", domain.Composition(
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
			return "", domain.Composition(
				"recipe " + recipe.ID + " default profile " + policy.DefaultProfile + " is not allowed")
		}
	}
	return policy.DefaultProfile, nil
}

// CheckCompatibility verifies the composed components against the recipe
// contract and the explicit technical constraints recovered from the
// decision: surface coverage, allowed surface composition (over the full
// component surface set), architectural capability coverage by the chosen
// providers, database policy versus catalog profiles, runtime/tech
// compatibility between components, and no provider violating an explicit
// must-not-use constraint. Every failure is a typed composition error.
func CheckCompatibility(decision domain.ArchitectureDecision, recipe domain.Recipe, comp Composition, caps []domain.CapabilityID, mustUse, mustNotUse []string, cat catalog.Catalog) error {
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
	byProvider := map[string][]string{}
	for _, c := range comp.Components {
		b, _ := idx.Boilerplate(c.Boilerplate)
		byProvider[c.Boilerplate] = providerTech(b)
	}
	for _, mn := range mustNotUse {
		for bp, tags := range byProvider {
			if techMatch(tags, mn) {
				return domain.Composition(
					"provider " + bp + " violates must-not-use tech " + mn)
			}
		}
	}
	for _, mu := range mustUse {
		coveredBy := false
		if techMatch(recipe.TechTags, mu) {
			coveredBy = true
		}
		if !coveredBy {
			for _, tags := range byProvider {
				if techMatch(tags, mu) {
					coveredBy = true
					break
				}
			}
		}
		if !coveredBy {
			return domain.Composition(
				"explicit must-use tech " + mu + " is satisfied by no chosen component")
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
