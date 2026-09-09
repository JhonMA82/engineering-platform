// Package handoff generates the agent-context and ownership-transfer
// artifacts of a materialized project. Generation is pure: Render returns
// project-relative files and the materializer performs every write. There
// is intentionally no separate `eng handoff` command: the materializer
// invokes generation as part of every materialization.
package handoff

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// Handoff status, ownership and implementation-mode constants. The platform
// never sets implementation_mode=sdd: Gentle decides after reading the
// handoff (see GENTLE.md).
const (
	StatusReadyForImplementation    = "ready_for_implementation"
	NextOwnerGentle                 = "gentle-ai"
	ImplementationModeGentleDecides = "gentle-decides"
)

// LockedDecisions are the decisions Gentle normally does not rediscover.
// database-profile is locked: the composer already bound the recipe policy
// to a curated profile, so Gentle must not swap storage silently.
var LockedDecisions = []string{"architecture", "surface-topology", "selected-foundations", "database-profile"}

// Input carries everything generation may render. Intent and decision
// documents are optional: when absent, intent/decision copies are skipped
// and the brief degrades to plan-plus-catalog facts.
type Input struct {
	Plan         planner.MaterializationPlan
	Catalog      catalog.Catalog
	IntentJSON   []byte
	DecisionJSON []byte
}

// Requirement is one pending product requirement transferred to Gentle.
type Requirement struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

// Handoff declares the formal ownership transfer to Gentle AI.
// Requirements/OpenQuestions are the historic field names (product
// requirements and routing-ambiguity questions); ProductRequirements and
// OpenProductQuestions are the §5.3 contract names carrying the same product
// content. ImplementationMode is always "gentle-decides": the platform
// never prescribes direct-build vs SDD.
type Handoff struct {
	SchemaVersion        int             `json:"schema_version"`
	Status               string          `json:"status"`
	NextOwner            string          `json:"next_owner"`
	Locked               []string        `json:"locked"`
	Requirements         []Requirement   `json:"requirements"`
	OpenQuestions        []string        `json:"open_questions"`
	ProductRequirements  []Requirement   `json:"product_requirements"`
	OpenProductQuestions []string        `json:"open_product_questions"`
	ImplementationMode   string          `json:"implementation_mode"`
	Foundations          []foundationRef `json:"foundations"`
}

// File is one rendered artifact: a project-relative slash path, its
// bytes, and whether it overwrites an existing file. Surface stubs never
// overwrite: a foundation-shipped AGENTS.md is preserved.
type File struct {
	Path      string
	Data      []byte
	Overwrite bool
}

// Project-relative paths of the handoff-owned files.
const (
	IntentCopyRelPath   = ".engineering/project-intent.json"
	DecisionCopyRelPath = ".engineering/architecture-decision.json"
	PlanCopyRelPath     = ".engineering/materialization-plan.json"
	BriefRelPath        = ".engineering/implementation-brief.md"
	HandoffRelPath      = ".engineering/handoff.json"
	AgentsRelPath       = "AGENTS.md"
	ArchitectureRelPath = "ARCHITECTURE.md"
	GentleRelPath       = "GENTLE.md"
)

// Render returns every handoff-owned file without touching the disk: the
// .engineering copies (intent, decision, plan, brief, handoff), the root
// AGENTS.md router, ARCHITECTURE.md, GENTLE.md and one per-surface
// AGENTS.md stub (non-overwriting, so foundation-shipped instructions
// survive). The materializer writes the result into staging.
func Render(in Input) ([]File, error) {
	if len(in.Plan.Components) == 0 {
		return nil, domain.Materialization("cannot generate handoff: plan has no components")
	}
	var intent *domain.ProjectIntent
	if len(bytesTrimSpace(in.IntentJSON)) > 0 {
		var parsed domain.ProjectIntent
		if err := json.Unmarshal(in.IntentJSON, &parsed); err != nil {
			return nil, domain.Validation(fmt.Sprintf("handoff: parse intent: %v", err))
		}
		intent = &parsed
	}
	var decision *domain.ArchitectureDecision
	if len(bytesTrimSpace(in.DecisionJSON)) > 0 {
		var parsed domain.ArchitectureDecision
		if err := json.Unmarshal(in.DecisionJSON, &parsed); err != nil {
			return nil, domain.Validation(fmt.Sprintf("handoff: parse decision: %v", err))
		}
		decision = &parsed
	}
	var out []File
	if intent != nil {
		data, err := canonicalJSON(in.IntentJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: IntentCopyRelPath, Data: data, Overwrite: true})
	}
	if decision != nil {
		data, err := canonicalJSON(in.DecisionJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: DecisionCopyRelPath, Data: data, Overwrite: true})
	}
	planBytes, err := encodeJSON(in.Plan)
	if err != nil {
		return nil, err
	}
	out = append(out, File{Path: PlanCopyRelPath, Data: planBytes, Overwrite: true})
	view := buildView(in, intent, decision)
	for _, doc := range []struct {
		path string
		name string
		data any
	}{
		{BriefRelPath, "brief", view},
		{AgentsRelPath, "agents", view},
		{ArchitectureRelPath, "architecture", view},
		{GentleRelPath, "gentle", view},
	} {
		data, err := renderDoc(doc.name, doc.data)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: doc.path, Data: data, Overwrite: true})
	}
	handoffBytes, err := encodeJSON(view.Handoff)
	if err != nil {
		return nil, err
	}
	out = append(out, File{Path: HandoffRelPath, Data: handoffBytes, Overwrite: true})
	for _, c := range view.Components {
		data, err := renderDoc("surface", c)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: c.Destination + "/AGENTS.md", Data: data, Overwrite: false})
	}
	return out, nil
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func encodeJSON(v any) ([]byte, error) {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("handoff: encode: %v", err))
	}
	return append(raw, '\n'), nil
}

func canonicalJSON(raw []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, domain.Validation(fmt.Sprintf("handoff: invalid JSON: %v", err))
	}
	return encodeJSON(v)
}

// componentView is one plan component enriched with provider metadata.
// Strategy and Profile name the generation configuration that produced
// the surface so Gentle understands which foundation variant it owns
// without rediscovering why the profile was chosen.
type componentView struct {
	Surface          string
	Destination      string
	Boilerplate      string
	Pin              string
	Strategy         string
	Profile          string
	Responsibility   string
	Instructions     string
	ProvidedFeatures []string
}

// foundationRef is the per-surface generation record transferred to
// Gentle: which foundation, pin, strategy and profile produced it.
type foundationRef struct {
	Surface     string `json:"surface"`
	Destination string `json:"destination"`
	Provider    string `json:"provider"`
	Pin         string `json:"pin"`
	Strategy    string `json:"strategy"`
	Profile     string `json:"profile,omitempty"`
}

// briefView is the template model shared by every markdown document.
type briefView struct {
	Project          string
	Recipe           string
	RecipeVersion    string
	DatabaseProfile  string
	PlanFingerprint  string
	CatalogVersion   string
	Components       []componentView
	Relationships    []string
	Problem          string
	ProductReqs      []Requirement
	ArchReqs         []Requirement
	Constraints      []string
	PlannedLater     []string
	Excluded         []string
	OpenQuestions    []string
	ProductQuestions []string
	RoutingQuestions []string
	Checks           []string
	HasIntent        bool
	SingleSurface    bool
	SelectedRecipe   string
	Confidence       string
	Handoff          Handoff
}

func buildView(in Input, intent *domain.ProjectIntent, decision *domain.ArchitectureDecision) briefView {
	idx := catalog.NewIndex(in.Catalog)
	components := make([]componentView, 0, len(in.Plan.Components))
	for _, c := range in.Plan.Components {
		cv := componentView{
			Surface:          c.Surface,
			Destination:      c.Destination,
			Boilerplate:      c.Boilerplate,
			Pin:              c.Pin,
			Strategy:         c.EffectiveStrategy(),
			Profile:          c.Materialization.Profile,
			Responsibility:   surfaceResponsibility(c.Surface),
			Instructions:     c.Destination + "/AGENTS.md",
			ProvidedFeatures: []string{},
		}
		if bp, ok := idx.Boilerplate(c.Boilerplate); ok && len(bp.IncludedFeatures) > 0 {
			cv.ProvidedFeatures = append([]string{}, bp.IncludedFeatures...)
			sort.Strings(cv.ProvidedFeatures)
		}
		components = append(components, cv)
	}
	view := briefView{
		Project:         in.Plan.Project,
		Recipe:          in.Plan.Recipe,
		RecipeVersion:   in.Plan.RecipeVersion,
		DatabaseProfile: in.Plan.DatabaseProfile,
		PlanFingerprint: in.Plan.Fingerprint,
		CatalogVersion:  in.Catalog.CatalogVersion,
		Components:      components,
		Checks:          append([]string{}, in.Plan.Checks...),
		SingleSurface:   len(components) == 1,
	}
	view.Relationships = relationshipLines(in.Plan)
	if intent != nil {
		view.HasIntent = true
		view.Problem = intent.Problem
		for _, r := range intent.ProductRequirements {
			view.ProductReqs = append(view.ProductReqs, Requirement{ID: r.ID, Description: r.Description})
		}
		for _, r := range intent.ArchitectureRequirements {
			view.ArchReqs = append(view.ArchReqs, Requirement{ID: r.ID, Description: r.Description})
		}
		for _, c := range intent.TechnicalConstraints {
			if c.Value == "" {
				view.Constraints = append(view.Constraints, c.Kind)
			} else {
				view.Constraints = append(view.Constraints, c.Kind+": "+c.Value)
			}
		}
		view.PlannedLater = append([]string{}, intent.Scope.PlannedLater...)
		view.Excluded = append([]string{}, intent.Scope.ExplicitlyExcluded...)
		view.ProductQuestions = append([]string{}, intent.OpenProductQuestions...)
	}
	if decision != nil && decision.Selected != nil {
		view.SelectedRecipe = decision.Selected.Recipe
		if decision.Selected.RecipeVersion != "" {
			view.SelectedRecipe += " " + decision.Selected.RecipeVersion
		}
	}
	if decision != nil && decision.Confidence.Level != "" {
		view.Confidence = fmt.Sprintf("%s (margin %d)", decision.Confidence.Level, decision.Confidence.Margin)
		for _, d := range decision.UnresolvedDimensions {
			q := d.Dimension
			if d.Reason != "" {
				q += ": " + d.Reason
			}
			view.RoutingQuestions = append(view.RoutingQuestions, q)
		}
	}
	// OpenQuestions keeps the historic routing-ambiguity content; product
	// questions travel separately and never affect routing.
	foundations := make([]foundationRef, 0, len(components))
	for _, cv := range components {
		foundations = append(foundations, foundationRef{
			Surface:     cv.Surface,
			Destination: cv.Destination,
			Provider:    cv.Boilerplate,
			Pin:         cv.Pin,
			Strategy:    cv.Strategy,
			Profile:     cv.Profile,
		})
	}
	handoff := Handoff{
		SchemaVersion:        1,
		Status:               StatusReadyForImplementation,
		NextOwner:            NextOwnerGentle,
		Locked:               append([]string{}, LockedDecisions...),
		Foundations:          foundations,
		Requirements:         append([]Requirement{}, view.ProductReqs...),
		OpenQuestions:        append([]string{}, view.RoutingQuestions...),
		ProductRequirements:  append([]Requirement{}, view.ProductReqs...),
		OpenProductQuestions: append([]string{}, view.ProductQuestions...),
		ImplementationMode:   ImplementationModeGentleDecides,
	}
	if handoff.Requirements == nil {
		handoff.Requirements = []Requirement{}
	}
	if handoff.OpenQuestions == nil {
		handoff.OpenQuestions = []string{}
	}
	if handoff.ProductRequirements == nil {
		handoff.ProductRequirements = []Requirement{}
	}
	if handoff.OpenProductQuestions == nil {
		handoff.OpenProductQuestions = []string{}
	}
	if handoff.Foundations == nil {
		handoff.Foundations = []foundationRef{}
	}
	view.OpenQuestions = append([]string{}, view.RoutingQuestions...)
	view.Handoff = handoff
	return view
}

func relationshipLines(plan planner.MaterializationPlan) []string {
	hasAPI := false
	var others []string
	for _, c := range plan.Components {
		if c.Surface == "api" {
			hasAPI = true
		} else {
			others = append(others, c.Surface+" ("+c.Destination+") consumes api")
		}
	}
	if !hasAPI {
		return []string{}
	}
	sort.Strings(others)
	return others
}

// surfaceResponsibility names the job of a surface for the router table.
func surfaceResponsibility(surface string) string {
	switch surface {
	case "api":
		return "Shared backend"
	case "web-admin":
		return "Admin UI"
	case "public-web":
		return "Public website"
	case "public-intake":
		return "Public intake"
	case "mobile-native":
		return "Native client"
	default:
		return titleize(surface)
	}
}

func titleize(s string) string {
	parts := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if len(parts) == 0 {
		return s
	}
	return strings.Join(parts, " ")
}

var documents = map[string]string{
	"brief": `# Implementation brief — {{.Project}}

Recipe {{.Recipe}}{{if .RecipeVersion}} {{.RecipeVersion}}{{end}} · database profile {{.DatabaseProfile}} · catalog {{.CatalogVersion}}.
{{if .HasIntent}}
## Problem

{{if .Problem}}{{.Problem}}{{else}}No problem statement recorded.{{end}}
{{end}}
## Surfaces

{{range .Components}}
- {{.Surface}} → {{.Destination}} ({{.Boilerplate}}@{{.Pin}})
{{end}}
{{if .ProductReqs}}
## Pending product implementation

{{range .ProductReqs}}
- {{.ID}}{{if .Description}}: {{.Description}}{{end}}
{{end}}
{{end}}
{{if .ArchReqs}}
## Architecture requirements

{{range .ArchReqs}}
- {{.ID}}{{if .Description}}: {{.Description}}{{end}}
{{end}}
{{end}}
## Selected foundations

{{range .Components}}
- {{.Destination}} serves {{.Surface}} via {{.Boilerplate}}@{{.Pin}} ({{.Strategy}}{{if .Profile}}, profile {{.Profile}}{{end}})
{{end}}

Remaining product work is Gentle's: foundations cover structure, not features.
{{range .Components}}{{if .ProvidedFeatures}}
## Already provided by foundation — {{.Destination}}

{{range .ProvidedFeatures}}
- {{.}}
{{end}}
{{end}}{{end}}
{{if .Constraints}}
## Locked architecture decisions

Architecture, surface topology, selected foundations, database profile and the
technical constraints below are locked (see .engineering/handoff.json). Reopen
them only on a concrete contradiction.

{{range .Constraints}}
- {{.}}
{{end}}
{{end}}
{{if .PlannedLater}}
## Planned later (out of scope now)

{{range .PlannedLater}}
- {{.}}
{{end}}
{{end}}
{{if .Excluded}}
## Explicitly excluded

{{range .Excluded}}
- {{.}}
{{end}}
{{end}}
{{if .Checks}}
## Quality gates carried from the recipe

{{range .Checks}}
- {{.}}
{{end}}
{{end}}
{{if .ProductQuestions}}
## Open product questions

These are product/domain unknowns. They never affected routing and must not
reopen architecture: answer them by direct implementation or an SDD session
focused on product behavior (see GENTLE.md).

{{range .ProductQuestions}}
- {{.}}
{{end}}
{{end}}
{{if .RoutingQuestions}}
## Unresolved routing questions

{{range .RoutingQuestions}}
- {{.}}
{{end}}
{{end}}
Plan fingerprint {{.PlanFingerprint}}.
`,
	"agents": `# {{.Project}} — agent router

Recipe {{.Recipe}}{{if .RecipeVersion}} {{.RecipeVersion}}{{end}} · database profile {{.DatabaseProfile}}.

This file routes work. It does not repeat foundation internals; read the
linked surface instructions before modifying a surface.

| Surface | Path | Responsibility | Instructions |
|---|---|---|---|
{{range .Components}}| {{.Surface}} | {{.Destination}} | {{.Responsibility}} | {{.Instructions}} |
{{end}}
## Cross-surface rules

{{if .SingleSurface}}- Single-surface project: no cross-surface calls.
{{else}}{{range .Relationships}}- {{.}}
{{end}}{{end}}- Respect each foundation's managed files; never move a surface directory.
- Architecture, surface topology and selected foundations are locked (see .engineering/handoff.json).

## Sources of truth

- .engineering/project-intent.json — what the product needs
- .engineering/architecture-decision.json — why this architecture won
- .engineering/materialization-plan.json — what was materialized
- .engineering/project-map.json — machine-readable surface routing
- ARCHITECTURE.md — how the parts relate
- .engineering/implementation-brief.md — what to build next
`,
	"architecture": `# Architecture — {{.Project}}

Recipe {{.Recipe}}{{if .RecipeVersion}} {{.RecipeVersion}}{{end}} on database profile {{.DatabaseProfile}}.

## Parts

{{range .Components}}
- {{.Destination}} — {{.Surface}} ({{.Responsibility}}), provider {{.Boilerplate}}@{{.Pin}}
{{end}}
## Relations

{{if .Relationships}}{{range .Relationships}}- {{.}}
{{end}}{{else}}Single part: no inter-surface relations.
{{end}}
## Decisions

Locked: architecture, surface-topology, selected-foundations.
Full rationale: .engineering/architecture-decision.json.
Plan fingerprint {{.PlanFingerprint}}.
`,
	"gentle": `# GENTLE.md — taking ownership of {{.Project}}

You are taking ownership of a project generated by Engineering Platform.

Do not rediscover or replace the selected architecture unless a concrete
contradiction is found.

Read:
1. .engineering/implementation-brief.md
2. AGENTS.md
3. ARCHITECTURE.md
4. .engineering/project-map.json
5. the AGENTS.md of the Surface you will modify

Then decide:

A. Direct implementation
   Use this path when the product requirements are sufficiently defined.

B. SDD session
   Use this path only when important product/domain rules remain undefined.

An SDD session must focus on product behavior, workflows, rules, permissions,
edge cases and domain decisions.

Do not ask the user to repeat information already present in the handoff.
Do not reopen framework, database, boilerplate or topology choices unless a
real contradiction is discovered.

Handoff state: status {{.Handoff.Status}}, next owner {{.Handoff.NextOwner}},
implementation mode {{.Handoff.ImplementationMode}} (you decide: direct
implementation or SDD session — the platform never prescribes it).
`,
	"surface": `# {{.Surface}} surface — {{.Destination}}

Provider {{.Boilerplate}}@{{.Pin}} ({{.Responsibility}}), materialized via {{.Strategy}}{{if .Profile}} with profile {{.Profile}}{{end}}.
{{if .ProvidedFeatures}}
Already provided by the foundation: {{join .ProvidedFeatures ", "}}.
{{end}}
Work here following the foundation conventions shipped in this directory.
See the repository root AGENTS.md for cross-surface routing; do not move
this directory.
`,
}

var blankRuns = regexp.MustCompile(`\n{3,}`)

func renderDoc(name string, data any) ([]byte, error) {
	tpl, err := template.New(name).Funcs(template.FuncMap{"join": strings.Join}).Parse(documents[name])
	if err != nil {
		return nil, domain.Materialization(fmt.Sprintf("handoff: bad %s template: %v", name, err))
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return nil, domain.Materialization(fmt.Sprintf("handoff: render %s: %v", name, err))
	}
	// Range/if actions leave blank runs; collapse them so every document
	// renders tight deterministic markdown.
	return []byte(blankRuns.ReplaceAllString(sb.String(), "\n\n")), nil
}
