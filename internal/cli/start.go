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

// runStart chains the P7 bootstrap without merging responsibilities:
// PlanProject resolves+composes+plans, MaterializeProject executes the plan,
// DoctorProject verifies the result. The CLI only wires bytes between the
// thin app services. --dry-run prints the plan and performs zero filesystem
// writes. An unresolved decision surfaces the typed composition error from
// PlanProject and exits non-zero.
//
// With --audit (and ENG_AUDIT=1) the whole chain is stitched into one audit
// session: intent snapshot, resolve result, plan, materialize and doctor
// land in .engineering/audit/<session>/AUDIT.md + audit.json alongside the
// skill-written discovery events.
func runStart(args []string) int {
	fs := newFlagSet("start")
	intentPath := fs.String("intent", "", "path to project intent JSON file")
	output := fs.String("output", "", "project output directory (new, empty, or eng-init workspace; use . inside init workspace)")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	dryRun := fs.Bool("dry-run", false, "print the plan without writing anything")
	asJSON := fs.Bool("json", false, "print machine-readable JSON")
	audit := addAuditFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *intentPath == "" {
		fmt.Fprintln(os.Stderr, "start: --intent is required")
		return 2
	}
	if *output == "" && !*dryRun {
		fmt.Fprintln(os.Stderr, "start: --output is required (omit only with --dry-run)")
		return 2
	}
	now := time.Now().UTC()
	auditOutput := *output
	if *dryRun && auditOutput == "" {
		auditOutput = "."
	}
	baseDir, sessionID, auditing := auditSession(audit, auditOutput, now)
	var tail project.AuditSession
	if auditing {
		tail = project.NewAuditSession(sessionID, "", "", now)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseIntent,
			Kind:    project.AuditKindCommand,
			Summary: fmt.Sprintf("start --intent %s --output %s", *intentPath, *output),
			Command: "start",
			Argv:    append([]string{"eng", "start"}, args...),
			Data:    map[string]string{"intent": *intentPath, "output": *output},
		}, now)
	}
	intentRaw, err := os.ReadFile(*intentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", project.SummarizeIntent(nil, *output), err, nil, ""))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseIntent, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("read intent failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if auditing {
		sum := project.SummarizeIntent(intentRaw, *output)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseIntent,
			Kind:    project.AuditKindIntent,
			Summary: fmt.Sprintf("intent %q: %s (%d surfaces)", sum.Name, sum.Problem, len(sum.Surfaces)),
			Data: map[string]string{
				"name": sum.Name, "problem": sum.Problem,
				"surfaces": fmt.Sprintf("%d", len(sum.Surfaces)),
				"output":   *output,
			},
		}, now)
	}
	decision, _, plan, err := app.PlanProject(intentRaw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", project.SummarizeIntent(intentRaw, *output), err, nil, ""))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseResolve, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("resolve/plan failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if auditing && decision != nil {
		status := string(decision.Status)
		recipe := ""
		if decision.Selected != nil {
			recipe = decision.Selected.Recipe
		}
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseResolve,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("resolve %s: recipe %s fingerprint %s (%d unresolved)", status, recipe, decision.IntentFingerprint, len(decision.UnresolvedDimensions)),
			Status:  status,
			Data: map[string]string{
				"decision_status": status, "recipe": recipe,
				"intent_fingerprint": decision.IntentFingerprint,
				"unresolved_count":   fmt.Sprintf("%d", len(decision.UnresolvedDimensions)),
			},
		}, now)
	}
	if auditing && plan != nil {
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhasePlan,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("plan recipe %s (%d components) fingerprint %s", plan.Recipe, len(plan.Components), plan.Fingerprint),
			Status:  "ok",
			Data: map[string]string{
				"recipe": plan.Recipe, "recipe_version": plan.RecipeVersion,
				"plan_fingerprint": plan.Fingerprint,
				"components":       fmt.Sprintf("%d", len(plan.Components)),
			},
		}, now)
	}
	if *dryRun {
		if *asJSON {
			out, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(out))
			if auditing {
				tail.AddEvent(project.AuditEvent{
					Phase: project.AuditPhaseDone, Kind: project.AuditKindResult,
					Summary: "dry-run: plan printed, zero filesystem writes", Status: "ok",
				}, now)
				auditFinalize(baseDir, tail)
			}
			return 0
		}
		printPlanHuman(*plan)
		fmt.Println("(dry-run: no filesystem writes)")
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseDone, Kind: project.AuditKindResult,
				Summary: "dry-run: plan printed, zero filesystem writes", Status: "ok",
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 0
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		intent := project.SummarizeIntent(intentRaw, *output).SummarizePlan(plan)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", intent, err, nil, ""))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhasePlan, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("encode plan failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		intent := project.SummarizeIntent(intentRaw, *output).SummarizePlan(plan)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", intent, err, nil, ""))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseResolve, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("encode decision failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	manifest, err := app.MaterializeProject(planJSON, intentRaw, decisionJSON, *catalogDir, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		intent := project.SummarizeIntent(intentRaw, *output).SummarizePlan(plan)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", intent, err, nil, ""))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseMaterial, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("materialize failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error(), "output": *output},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if auditing && manifest != nil {
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseMaterial,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("materialized %s recipe %s (%d components, %d files)", manifest.Project, manifest.Recipe, len(manifest.Components), len(manifest.Files)),
			Status:  "ok",
			Data: map[string]string{
				"project": manifest.Project, "recipe": manifest.Recipe,
				"plan_fingerprint": manifest.PlanFingerprint,
				"components":       fmt.Sprintf("%d", len(manifest.Components)),
				"output":           *output,
			},
		}, now)
	}
	findings, err := app.DoctorProject(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		intent := project.SummarizeIntent(intentRaw, *output).SummarizePlan(plan)
		manifestFP := ""
		if manifest != nil {
			manifestFP = manifest.PlanFingerprint
		}
		persistReport(reportBaseDir(*output), project.ErrorReport("start", intent, err, nil, manifestFP))
		if auditing {
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseDoctor, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("doctor failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if *asJSON {
		out, _ := json.MarshalIndent(struct {
			Plan     *planner.MaterializationPlan `json:"plan"`
			Manifest *project.Manifest            `json:"manifest"`
			Doctor   []project.Finding            `json:"doctor"`
		}{Plan: plan, Manifest: manifest, Doctor: doctorJSON(findings)}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Printf("started %s — recipe %s (%d components, %d files)\n",
			manifest.Project, manifest.Recipe, len(manifest.Components), len(manifest.Files))
		for _, c := range manifest.Components {
			fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
		}
		printDoctorHuman(*output, findings)
	}
	if auditing {
		status := "ok"
		if project.HasErrors(findings) {
			status = "error"
		}
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseDoctor,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("doctor %s: %d finding(s)", status, len(findings)),
			Status:  status,
			Data:    map[string]string{"findings": fmt.Sprintf("%d", len(findings))},
		}, now)
		tail.AddEvent(project.AuditEvent{
			Phase: project.AuditPhaseDone, Kind: project.AuditKindResult,
			Summary: fmt.Sprintf("start %s: %s recipe %s", status, manifest.Project, manifest.Recipe),
			Status:  status,
			Data:    map[string]string{"project": manifest.Project, "recipe": manifest.Recipe},
		}, now)
		auditFinalize(baseDir, tail)
	}
	if project.HasErrors(findings) {
		intent := project.SummarizeIntent(intentRaw, *output).SummarizePlan(plan)
		persistReport(reportBaseDir(*output), project.ErrorReport("start", intent, nil, findings, manifest.PlanFingerprint))
		return 1
	}
	// Bootstrap lifecycle: a workspace
	// prepared by eng init sheds its disposable bootstrap resources only
	// after successful materialization and validation. Cleanup is
	// ownership-based (never pattern-based) and a no-op for directories
	// eng init never prepared. A cleanup failure warns but does not fail
	// the otherwise valid project.
	if crep, err := app.CleanupBootstrap(*output); err != nil {
		fmt.Fprintf(os.Stderr, "warning: bootstrap cleanup: %v\n", err)
	} else if crep.Cleaned {
		fmt.Printf("bootstrap cleanup: removed %d file(s)\n", len(crep.Removed))
	}
	return 0
}
