package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/project"
)

func runMaterialize(args []string) int {
	fs := newFlagSet("materialize")
	planPath := fs.String("plan", "", "path to materialization plan JSON file")
	output := fs.String("output", "", "project output directory (new, empty, or eng-init workspace; use . inside init workspace)")
	intentPath := fs.String("intent", "", "optional path to project intent JSON file")
	decisionPath := fs.String("decision", "", "optional path to architecture decision JSON file")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	audit := addAuditFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *planPath == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "materialize: --plan and --output are required")
		return 2
	}
	now := time.Now().UTC()
	baseDir, sessionID, auditing := auditSession(audit, *output, now)
	if auditing {
		auditRecord(baseDir, sessionID, project.AuditEvent{
			Phase:   project.AuditPhaseMaterial,
			Kind:    project.AuditKindCommand,
			Summary: fmt.Sprintf("materialize --plan %s --output %s", *planPath, *output),
			Command: "materialize",
			Argv:    append([]string{"eng", "materialize"}, args...),
			Data:    map[string]string{"plan": *planPath, "output": *output, "intent": *intentPath, "decision": *decisionPath},
		})
	}
	planRaw, err := os.ReadFile(*planPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
		persistReport(reportBaseDir(*output), project.ErrorReport("materialize", project.SummarizeIntent(nil, *output), err, nil, ""))
		return 1
	}
	var intentRaw, decisionRaw []byte
	if *intentPath != "" {
		intentRaw, err = os.ReadFile(*intentPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
			persistReport(reportBaseDir(*output), project.ErrorReport("materialize", project.SummarizeIntent(nil, *output), err, nil, ""))
			return 1
		}
	}
	if *decisionPath != "" {
		decisionRaw, err = os.ReadFile(*decisionPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
			persistReport(reportBaseDir(*output), project.ErrorReport("materialize", project.SummarizeIntent(intentRaw, *output), err, nil, ""))
			return 1
		}
	}
	manifest, err := app.MaterializeProject(planRaw, intentRaw, decisionRaw, *catalogDir, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
		persistReport(reportBaseDir(*output), project.ErrorReport("materialize", project.SummarizeIntent(intentRaw, *output), err, nil, ""))
		if auditing {
			tail := project.NewAuditSession(sessionID, "", "", now)
			tail.AddEvent(project.AuditEvent{
				Phase: project.AuditPhaseMaterial, Kind: project.AuditKindResult,
				Summary: fmt.Sprintf("materialize failed: %v", err), Status: "error",
				Data: map[string]string{"error": err.Error(), "output": *output},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	fmt.Printf("materialized %s — recipe %s (%d components, %d files)\n",
		manifest.Project, manifest.Recipe, len(manifest.Components), len(manifest.Files))
	for _, c := range manifest.Components {
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
	}
	fmt.Printf("fingerprint: %s\n", manifest.PlanFingerprint)
	if auditing {
		tail := project.NewAuditSession(sessionID, "", "", now)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseMaterial,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("materialized %s recipe %s (%d components, %d files) fingerprint %s", manifest.Project, manifest.Recipe, len(manifest.Components), len(manifest.Files), manifest.PlanFingerprint),
			Status:  "ok",
			Data: map[string]string{
				"project": manifest.Project, "recipe": manifest.Recipe,
				"plan_fingerprint": manifest.PlanFingerprint,
				"components":       fmt.Sprintf("%d", len(manifest.Components)),
				"files":            fmt.Sprintf("%d", len(manifest.Files)),
				"output":           *output,
			},
		}, now)
		auditFinalize(baseDir, tail)
	}
	return 0
}
