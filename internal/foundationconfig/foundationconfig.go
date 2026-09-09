// Package foundationconfig resolves the deterministic generation
// configuration for each selected foundation, after composition and
// before planning. It is pure: stdlib plus domain, composer and catalog
// types only — no filesystem, network or process effects.
//
// Boundary: the platform decides which foundation to use and with which
// curated configuration; the materializer only executes the resolved
// plan. Profile ids (minimal, authenticated, ...) are foundation
// concepts the core never interprets; selection matches catalog-declared
// requirements against structured decision signals (surfaces,
// capabilities, stable product-requirement ids), never free text and
// never an LLM. A missing product feature never disqualifies a
// foundation: unmet requirements fall back to a smaller profile while
// the feature stays pending for Gentle.
package foundationconfig

import (
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// MaterializationConfig is the per-component generation decision
// recorded in the plan. Strategy is copy or generate; Name is the
// logical generated name (not the final destination); Profile and
// Arguments carry the resolved non-runtime configuration; temporary
// paths are never recorded here. Copy components carry only the
// strategy and the adapter fingerprint.
type MaterializationConfig struct {
	Strategy           string   `json:"strategy"`
	Name               string   `json:"name,omitempty"`
	Profile            string   `json:"profile,omitempty"`
	Arguments          []string `json:"arguments,omitempty"`
	AdapterFingerprint string   `json:"adapter_fingerprint"`
}

// LogicalName derives the deterministic generator name: the project name
// for single-surface projects, {project}-{surface} otherwise. Names are
// slugified (lowercase, separator runs collapse to one dash) so the value
// stays portable when interpolated into argv and output paths. The
// generator name never decides the monorepo topology: the composer owns
// destinations and the generator output is relocated into staging.
func LogicalName(project string, surface domain.SurfaceID, componentCount int) string {
	base := project
	if componentCount > 1 {
		base = project + "-" + string(surface)
	}
	return slugifyName(base)
}

// slugifyName normalizes a logical generator name: lowercase, every run
// of characters outside [a-z0-9._-] collapses to one dash, leading and
// trailing dashes and dots are trimmed.
func slugifyName(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	dash := false
	for i := 0; i < len(lower); i++ {
		c := lower[i]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-' {
			if c == '-' {
				if dash {
					continue
				}
				dash = true
			} else {
				dash = false
			}
			b.WriteByte(c)
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-.")
}

// validLogicalName keeps generated names portable: the name is
// interpolated into argv, so separators, escapes and blanks fail here,
// before any plan is written.
func validLogicalName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if c == '.' || c == '_' || c == '-' {
			continue
		}
		return false
	}
	if name == "." || name == ".." {
		return false
	}
	return true
}

// Resolve produces the materialization decision for one composed
// component. Copy foundations resolve to strategy plus fingerprint.
// Generated foundations resolve the smallest valid profile, its declared
// arguments and the adapter fingerprint. Hard must-not-use technology
// constraints fail closed here (no fallback): a configuration that would
// violate them is never produced.
func Resolve(comp composer.Component, project string, componentCount int, bp domain.Boilerplate, decision domain.ArchitectureDecision, cat catalog.Catalog) (MaterializationConfig, error) {
	spec := bp.EffectiveSpec()
	fingerprint := domain.AdapterFingerprint(spec)
	if spec.Generate == nil {
		return MaterializationConfig{
			Strategy:           domain.StrategyCopy,
			AdapterFingerprint: fingerprint,
		}, nil
	}
	if len(project) == 0 {
		project = "project"
	}
	name := LogicalName(project, comp.Surface, componentCount)
	if !validLogicalName(name) {
		return MaterializationConfig{}, domain.Composition(
			"generator configuration unresolved: logical name " + name + " is not portable")
	}
	profile, err := SelectProfile(bp, decision, cat)
	if err != nil {
		return MaterializationConfig{}, err
	}
	args := profileArguments(spec.Generate, profile)
	values := map[string]string{
		"name": name, "project": project,
		"surface": string(comp.Surface), "profile": profile, "output": "output",
	}
	if _, err := domain.SubstituteGeneratorPlaceholders(append(append([]string{}, spec.Generate.Run.Run...), args...), values); err != nil {
		return MaterializationConfig{}, domain.Composition(
			"generator configuration unresolved: " + err.Error())
	}
	if _, err := domain.SubstituteGeneratorPlaceholder(spec.Generate.Output, values); err != nil {
		return MaterializationConfig{}, domain.Composition(
			"generator configuration unresolved: " + err.Error())
	}
	if err := checkHardConstraints(bp, decision, cat); err != nil {
		return MaterializationConfig{}, err
	}
	out := MaterializationConfig{
		Strategy:           domain.StrategyGenerate,
		Name:               name,
		AdapterFingerprint: fingerprint,
	}
	if len(profile) != 0 {
		out.Profile = profile
	}
	if len(args) != 0 {
		out.Arguments = append([]string{}, args...)
	}
	return out, nil
}

// profileArguments returns the declared arguments of one profile.
func profileArguments(gen *domain.GenerateSpec, profile string) []string {
	for _, p := range gen.Profiles {
		if p.ID == profile {
			return append([]string{}, p.Arguments...)
		}
	}
	return nil
}

// SelectProfile picks the smallest valid generation profile. Profiles
// are declared smallest-first; the richest profile whose catalog-declared
// requirements are all satisfied by the decision signals wins. When no
// profile declares requirements beyond the default, or no profile fully
// matches, the default (smallest valid) profile wins and the remaining
// product work stays pending for Gentle. Foundations without profiles
// (legacy generators with a complete run template) resolve to "".
func SelectProfile(bp domain.Boilerplate, decision domain.ArchitectureDecision, cat catalog.Catalog) (string, error) {
	spec := bp.EffectiveSpec()
	if spec.Generate == nil || len(spec.Generate.Profiles) == 0 {
		return "", nil
	}
	idx := catalog.NewIndex(cat)
	caps := map[string]bool{}
	for _, c := range composer.RequiredCapabilities(decision, idx) {
		caps[strings.ToLower(string(c))] = true
	}
	surfaces := map[string]bool{}
	for _, s := range composer.RequiredSurfaces(decision, idx) {
		surfaces[strings.ToLower(string(s))] = true
	}
	features := productFeatureIDs(decision)
	winner := ""
	for _, p := range spec.Generate.Profiles {
		if profileMatches(p, caps, surfaces, features) {
			winner = p.ID
		}
	}
	if winner != "" {
		return winner, nil
	}
	return spec.Generate.DefaultProfile, nil
}

// profileMatches reports whether every catalog-declared requirement of a
// profile is satisfied by the decision signals.
func profileMatches(p domain.GeneratorProfile, caps, surfaces, features map[string]bool) bool {
	if p.Requires == nil {
		return true
	}
	for _, c := range p.Requires.Capabilities {
		if !caps[strings.ToLower(c)] {
			return false
		}
	}
	for _, s := range p.Requires.Surfaces {
		if !surfaces[strings.ToLower(s)] {
			return false
		}
	}
	for _, id := range p.Requires.RequirementIDs {
		if !features[strings.ToLower(id)] {
			return false
		}
	}
	for _, c := range p.Requires.UnlessCapabilities {
		if caps[strings.ToLower(c)] {
			return false
		}
	}
	return true
}

// featurePrefix is the stable scoring reason carrying a product feature
// id. Only exact-id matches count as signals; narrative text is never
// inspected.
const featurePrefix = "includes product feature: "

// productFeatureIDs recovers the stable product-requirement ids recorded
// in the decision reasons (winner and candidates), in lowercase.
func productFeatureIDs(decision domain.ArchitectureDecision) map[string]bool {
	out := map[string]bool{}
	collect := func(reasons []string) {
		for _, r := range reasons {
			rest, ok := strings.CutPrefix(r, featurePrefix)
			if !ok {
				continue
			}
			id := rest
			if i := strings.Index(rest, " ("); i >= 0 {
				id = rest[:i]
			}
			id = strings.ToLower(strings.TrimSpace(id))
			if id != "" {
				out[id] = true
			}
		}
	}
	collect(decision.Reasons)
	for _, c := range decision.Candidates {
		collect(c.PositiveReasons)
		collect(c.NegativeReasons)
	}
	return out
}

// checkHardConstraints fails closed when a must-not-use technology
// constraint excludes the selected foundation technology. Composition
// normally rejects such layouts first; this is the configuration-stage
// fail-safe so a generated profile can never silently violate a hard
// constraint.
func checkHardConstraints(bp domain.Boilerplate, decision domain.ArchitectureDecision, cat catalog.Catalog) error {
	_ = cat
	for _, c := range composer.MustNotUseConstraints(decision) {
		target := strings.ToLower(c.Target)
		value := strings.ToLower(c.Value)
		if value == "" {
			continue
		}
		switch target {
		case domain.ConstraintTargetFramework, domain.ConstraintTargetLanguage, domain.ConstraintTargetRuntime:
			for _, v := range bp.Technology.Values(target) {
				if strings.ToLower(v) == value {
					return domain.Composition(
						"generator configuration unresolved: hard constraint must-not-use " +
							target + "=" + c.Value + " excludes foundation " + bp.ID)
				}
			}
		case "":
			// Untyped legacy constraints match broadly: tech tags plus
			// every technology value, mirroring intent validation.
			for _, tag := range bp.TechTags {
				if strings.ToLower(tag) == value {
					return domain.Composition(
						"generator configuration unresolved: hard constraint must-not-use " +
							c.Value + " excludes foundation " + bp.ID)
				}
			}
			for _, target := range []string{
				domain.ConstraintTargetFramework,
				domain.ConstraintTargetLanguage,
				domain.ConstraintTargetRuntime,
			} {
				for _, v := range bp.Technology.Values(target) {
					if strings.ToLower(v) == value {
						return domain.Composition(
							"generator configuration unresolved: hard constraint must-not-use " +
								c.Value + " excludes foundation " + bp.ID)
					}
				}
			}
		}
	}
	return nil
}
