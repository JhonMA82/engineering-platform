package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
)

func runPlan(args []string) int {
	fs := newFlagSet("plan")
	input := fs.String("input", "", "path to intent JSON file")
	asJSON := fs.Bool("json", false, "print the full materialization plan as JSON")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	audit := addAuditFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "plan: --input is required")
		return 2
	}
	now := time.Now().UTC()
	baseDir, sessionID, auditing := auditSession(audit, ".", now)
	if auditing {
		auditRecord(baseDir, sessionID, project.AuditEvent{
			Phase:   project.AuditPhasePlan,
			Kind:    project.AuditKindCommand,
			Summary: fmt.Sprintf("plan --input %s", *input),
			Command: "plan",
			Argv:    append([]string{"eng", "plan"}, args...),
			Data:    map[string]string{"input": *input},
		})
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "plan: %v\n", err)
		if auditing {
			tail := project.NewAuditSession(sessionID, "", "", now)
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhasePlan, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("plan failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	_, _, plan, err := app.PlanProject(raw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "plan: %v\n", err)
		if auditing {
			tail := project.NewAuditSession(sessionID, "", "", now)
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhasePlan, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("plan failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if auditing {
		tail := project.NewAuditSession(sessionID, "", "", now)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhasePlan,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("plan recipe %s (%d components) fingerprint %s", plan.Recipe, len(plan.Components), plan.Fingerprint),
			Status:  "ok",
			Data: map[string]string{
				"recipe":           plan.Recipe,
				"recipe_version":   plan.RecipeVersion,
				"plan_fingerprint": plan.Fingerprint,
				"components":       fmt.Sprintf("%d", len(plan.Components)),
				"database_profile": plan.DatabaseProfile,
			},
		}, now)
		auditFinalize(baseDir, tail)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	printPlanHuman(*plan)
	return 0
}

// printPlanHuman renders the deterministic plan summary shared by plan and
// the start dry-run so both commands describe the same plan identically.
func printPlanHuman(plan planner.MaterializationPlan) {
	fmt.Printf("plan: %s — recipe %s %s\n", plan.Project, plan.Recipe, plan.RecipeVersion)
	fmt.Printf("database_profile: %s\n", plan.DatabaseProfile)
	fmt.Println("components:")
	for _, c := range plan.Components {
		detail := c.Surface
		if strategy := c.EffectiveStrategy(); strategy != "copy" {
			detail += ", " + strategy
			if c.Materialization.Profile != "" {
				detail += " profile=" + c.Materialization.Profile
			}
		}
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, detail)
	}
	fmt.Printf("fingerprint: %s\n", plan.Fingerprint)
}
