// Package handoff golden behavior: the Gentle ownership contract (§5).
//
// These tests pin the Direct-implementation vs SDD-session decision rule,
// the handoff.json product-question flow, the locked set (including
// database-profile), root AGENTS routing and per-surface stub coverage.
package handoff

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

func testInput(t *testing.T) Input {
	t.Helper()
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	plan := planner.MaterializationPlan{
		SchemaVersion:   1,
		Project:         "demo",
		Recipe:          "GP-06",
		RecipeVersion:   "1.0.0",
		DatabaseProfile: "postgresql-managed",
		Checks:          []string{"typecheck"},
		Fingerprint:     "fp-test",
		Components: []planner.PlanComponent{
			{Surface: "api", Destination: "services/api", Boilerplate: "hono-api", Pin: "v1.0.0"},
			{Surface: "web-admin", Destination: "apps/admin", Boilerplate: "tanstack-admin", Pin: "v1.0.0"},
		},
	}
	intent := domain.ProjectIntent{
		SchemaVersion: 1,
		Name:          "demo",
		Problem:       "operators manage records",
		ProductRequirements: []domain.ProductRequirement{
			{ID: "approve-record", Description: "approve records before publishing"},
		},
		OpenProductQuestions: []string{
			"define approval roles for record publishing",
			"define cancellation behavior for in-flight edits",
		},
		Surfaces: []domain.SurfaceIntent{{Kind: "web-admin", Scope: domain.ScopeRequiredNow}},
		TechnicalConstraints: []domain.TechnicalConstraint{
			{Target: "framework", Kind: "must-use", Value: "tanstack"},
		},
	}
	intentJSON, err := json.Marshal(intent)
	if err != nil {
		t.Fatalf("marshal intent: %v", err)
	}
	decision := domain.ArchitectureDecision{
		SchemaVersion: 1,
		Status:        domain.StatusResolved,
		Selected:      &domain.SelectedRecipe{Recipe: "GP-06", RecipeVersion: "1.0.0"},
		Confidence:    domain.Confidence{Level: "high", Margin: 30},
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		t.Fatalf("marshal decision: %v", err)
	}
	return Input{Plan: plan, Catalog: cat, IntentJSON: intentJSON, DecisionJSON: decisionJSON}
}

func renderFiles(t *testing.T, in Input) map[string]string {
	t.Helper()
	files, err := Render(in)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	out := map[string]string{}
	for _, f := range files {
		out[f.Path] = string(f.Data)
	}
	return out
}

// TestGentleDirectVsSDDRule pins the §5.1 decision contract: Gentle reads the
// brief/router/architecture/map, then decides direct implementation vs an
// SDD session limited to product/domain — never re-asking recorded facts,
// never reopening locked foundations without a contradiction.
func TestGentleDirectVsSDDRule(t *testing.T) {
	gentle := renderFiles(t, testInput(t))[GentleRelPath]
	for _, want := range []string{
		"Direct implementation",
		"SDD session",
		"product behavior, workflows, rules, permissions,",
		"Do not ask the user to repeat information already present in the handoff",
		"Do not reopen framework, database, boilerplate or topology choices",
		"gentle-decides",
		".engineering/implementation-brief.md",
		".engineering/project-map.json",
	} {
		if !strings.Contains(gentle, want) {
			t.Errorf("GENTLE.md lacks %q\n%s", want, gentle)
		}
	}
}

// TestHandoffJSONProductQuestions pins the §5.3 contract: product questions
// flow intent→handoff untouched by routing, the mode stays gentle-decides
// (the platform never prescribes sdd), and database-profile is locked.
func TestHandoffJSONProductQuestions(t *testing.T) {
	raw := renderFiles(t, testInput(t))[HandoffRelPath]
	var h Handoff
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("parse handoff: %v", err)
	}
	if h.Status != StatusReadyForImplementation {
		t.Errorf("status = %q", h.Status)
	}
	if h.NextOwner != NextOwnerGentle {
		t.Errorf("next_owner = %q", h.NextOwner)
	}
	if h.ImplementationMode != ImplementationModeGentleDecides {
		t.Errorf("implementation_mode = %q, want gentle-decides (platform never sets sdd)", h.ImplementationMode)
	}
	for _, want := range []string{"architecture", "surface-topology", "selected-foundations", "database-profile"} {
		found := false
		for _, l := range h.Locked {
			if l == want {
				found = true
			}
		}
		if !found {
			t.Errorf("locked lacks %q: %v", want, h.Locked)
		}
	}
	if len(h.ProductRequirements) != 1 || h.ProductRequirements[0].ID != "approve-record" {
		t.Errorf("product_requirements = %+v", h.ProductRequirements)
	}
	if len(h.OpenProductQuestions) != 2 ||
		h.OpenProductQuestions[0] != "define approval roles for record publishing" ||
		h.OpenProductQuestions[1] != "define cancellation behavior for in-flight edits" {
		t.Errorf("open_product_questions = %v", h.OpenProductQuestions)
	}
	// Legacy names stay populated with the same product content.
	if len(h.Requirements) != 1 || len(h.OpenQuestions) != 0 {
		t.Errorf("legacy requirements=%+v open_questions=%v", h.Requirements, h.OpenQuestions)
	}
}

// TestBriefSeparatesProductFromRouting pins §5.4: the brief distinguishes
// pending implementation, already-provided foundation features, open product
// questions and locked decisions.
func TestBriefSeparatesProductFromRouting(t *testing.T) {
	brief := renderFiles(t, testInput(t))[BriefRelPath]
	for _, want := range []string{
		"## Pending product implementation",
		"approve-record",
		"## Already provided by foundation",
		"## Open product questions",
		"define approval roles for record publishing",
		"## Locked architecture decisions",
	} {
		if !strings.Contains(brief, want) {
			t.Errorf("brief lacks %q\n%s", want, brief)
		}
	}
}

// TestRootAgentsRoutes pins the router: every materialized surface appears
// with its destination and per-surface instructions, plus one AGENTS.md stub
// per component so project-map coverage has a file counterpart.
func TestRootAgentsRoutes(t *testing.T) {
	files := renderFiles(t, testInput(t))
	agents := files[AgentsRelPath]
	for _, want := range []string{
		"services/api", "apps/admin",
		"services/api/AGENTS.md", "apps/admin/AGENTS.md",
	} {
		if !strings.Contains(agents, want) {
			t.Errorf("root AGENTS.md lacks %q\n%s", want, agents)
		}
	}
	for _, stub := range []string{"services/api/AGENTS.md", "apps/admin/AGENTS.md"} {
		if _, ok := files[stub]; !ok {
			t.Errorf("missing per-surface stub %s", stub)
		}
	}
}
