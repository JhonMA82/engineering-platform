package evolve

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/handoff"
	"github.com/jhonma82/engineering-platform/internal/project"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

// Requirement scopes accepted by eng add.
const (
	RequirementScopeNow   = "required_now"
	RequirementScopeLater = "planned_later"
)

// RequirementResult reports the recorded product requirement.
type RequirementResult struct {
	ID      string `json:"id"`
	Warning string `json:"warning,omitempty"`
}

// AddRequirementOptions carries the requirement-add inputs.
type AddRequirementOptions struct {
	Text    string
	Scope   string
	Catalog catalog.Catalog
	Now     time.Time
}

// AddRequirement records a product requirement without touching the
// architecture: the intent gains a ProductRequirement, the handoff
// requirements and brief refresh as pending implementation, and provenance
// gains an event. Resolution before and after must select the same recipe;
// any drift aborts with a typed error and no mutation.
func AddRequirement(projectDir string, opts AddRequirementOptions) (*RequirementResult, error) {
	text := strings.TrimSpace(opts.Text)
	if text == "" {
		return nil, domain.Validation("evolve: requirement text must not be empty")
	}
	scope := opts.Scope
	if scope == "" {
		scope = RequirementScopeNow
	}
	if scope != RequirementScopeNow && scope != RequirementScopeLater {
		return nil, domain.Validation(fmt.Sprintf("evolve: unknown requirement scope %q (want required_now|planned_later)", opts.Scope))
	}
	p, err := LoadProject(projectDir)
	if err != nil {
		return nil, err
	}
	if !p.HasPlan {
		return nil, domain.Filesystem(fmt.Sprintf("evolve: project has no stored plan (%s)", handoff.PlanCopyRelPath))
	}
	before := resolver.Resolve(p.Intent, opts.Catalog)
	if before.Status != domain.StatusResolved || before.Selected == nil {
		return nil, domain.Resolution("evolve: stored intent no longer resolves; refusing to evolve an inconsistent project")
	}
	evolved := p.Intent
	id := nextRequirementID(evolved.ProductRequirements)
	evolved.ProductRequirements = append(append([]domain.ProductRequirement{}, evolved.ProductRequirements...),
		domain.ProductRequirement{ID: id, Description: text})
	if scope == RequirementScopeLater && !containsString(evolved.Scope.PlannedLater, id) {
		evolved.Scope.PlannedLater = append(append([]string{}, evolved.Scope.PlannedLater...), id)
	}
	if err := evolved.Validate(); err != nil {
		return nil, err
	}
	after := resolver.Resolve(evolved, opts.Catalog)
	if after.Status != domain.StatusResolved || after.Selected == nil {
		return nil, domain.Resolution(fmt.Sprintf(
			"cannot evolve: requirement recording changed resolution to %q; refusing an architecture change via 'eng add'", after.Status))
	}
	if after.Selected.Recipe != p.Manifest.Recipe {
		return nil, domain.Resolution(fmt.Sprintf(
			"cannot evolve: requirement recording would move recipe %q to %q; 'eng add' never changes architecture",
			p.Manifest.Recipe, after.Selected.Recipe))
	}
	intentJSON, err := json.Marshal(evolved)
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("evolve: encode intent: %v", err))
	}
	decisionJSON, err := json.Marshal(p.Decision)
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("evolve: encode decision: %v", err))
	}
	if err := refreshGeneratedFiles(projectDir, handoff.Input{
		Plan:         p.Plan,
		Catalog:      opts.Catalog,
		IntentJSON:   intentJSON,
		DecisionJSON: decisionJSON,
	}); err != nil {
		return nil, err
	}
	if err := project.AppendEvent(projectDir, project.EvolutionEvent{
		Type:        project.EventRequirementAdd,
		Requirement: id,
		Recipe:      p.Manifest.Recipe,
	}, opts.Now); err != nil {
		return nil, err
	}
	if err := requireDoctorGreen(projectDir); err != nil {
		return nil, err
	}
	return &RequirementResult{ID: id, Warning: architectureVocabularyWarning(text, opts.Catalog)}, nil
}

// nextRequirementID mints the next deterministic REQ-NNN id, skipping
// collisions with hand-written ids.
func nextRequirementID(existing []domain.ProductRequirement) string {
	used := map[string]bool{}
	for _, r := range existing {
		used[r.ID] = true
	}
	for n := len(existing) + 1; ; n++ {
		id := fmt.Sprintf("REQ-%03d", n)
		if !used[id] {
			return id
		}
	}
}

func containsString(items []string, want string) bool {
	for _, it := range items {
		if it == want {
			return true
		}
	}
	return false
}

// architectureVocabularyWarning flags requirement text that names catalog
// architecture vocabulary (a surface, capability or alias). The requirement
// is still recorded as a product requirement; the warning points at the
// architecture commands in case the author meant a scope change.
func architectureVocabularyWarning(text string, cat catalog.Catalog) string {
	lowered := strings.ToLower(text)
	tokens := map[string]bool{}
	for _, tok := range strings.FieldsFunc(lowered, func(r rune) bool {
		return r != '-' && r != '_' && (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if tok = strings.Trim(tok, "-_"); tok != "" {
			tokens[tok] = true
		}
	}
	var singles []string
	for _, s := range cat.Surfaces {
		singles = append(singles, strings.ToLower(string(s.ID)))
	}
	for _, c := range cat.Capabilities {
		singles = append(singles, strings.ToLower(string(c.ID)))
	}
	for term, canonical := range cat.Aliases {
		t := strings.ToLower(strings.TrimSpace(term))
		if strings.Contains(t, " ") {
			if strings.Contains(lowered, t) {
				return vocabularyWarning(t)
			}
			continue
		}
		singles = append(singles, t)
		if c := strings.ToLower(strings.TrimSpace(canonical)); c != "" {
			singles = append(singles, c)
		}
	}
	sort.Strings(singles)
	for _, term := range singles {
		if term == "" {
			continue
		}
		if strings.Contains(term, " ") {
			if strings.Contains(lowered, term) {
				return vocabularyWarning(term)
			}
			continue
		}
		if tokens[term] {
			return vocabularyWarning(term)
		}
	}
	return ""
}

func vocabularyWarning(term string) string {
	return fmt.Sprintf("requirement mentions architecture vocabulary %q; recorded as a product requirement (pending implementation) — "+
		"if you meant to change architecture, use 'eng surface add' or 'eng extend' instead", term)
}
