package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/project"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

func runResolve(args []string) int {
	fs := newFlagSet("resolve")
	input := fs.String("input", "", "path to intent JSON file")
	asJSON := fs.Bool("json", false, "print the full decision as JSON")
	verbose := fs.Bool("verbose", false, "include fingerprint and candidate detail")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	audit := addAuditFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "resolve: --input is required")
		return 2
	}
	now := time.Now().UTC()
	baseDir, sessionID, auditing := auditSession(audit, ".", now)
	argv := append([]string{"eng", "resolve"}, args...)
	if auditing {
		auditRecord(baseDir, sessionID, project.AuditEvent{
			Phase:   project.AuditPhaseResolve,
			Kind:    project.AuditKindCommand,
			Summary: fmt.Sprintf("resolve --input %s", *input),
			Command: "resolve",
			Argv:    argv,
			Data:    map[string]string{"input": *input},
		})
		intentRaw, _ := os.ReadFile(*input)
		auditRecord(baseDir, sessionID, project.AuditEvent{
			Phase:   project.AuditPhaseIntent,
			Kind:    project.AuditKindIntent,
			Summary: fmt.Sprintf("intent snapshot for resolve (%d bytes)", len(intentRaw)),
			Data: map[string]string{
				"intent_bytes": fmt.Sprintf("%d", len(intentRaw)),
				"summary_name": project.SummarizeIntent(intentRaw, "").Name,
			},
		})
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve: %v\n", err)
		return 1
	}
	decision, err := app.ResolveProject(raw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve: %v\n", err)
		if auditing {
			tail := project.NewAuditSession(sessionID, "", "", now)
			tail.AddEvent(project.AuditEvent{
				Phase:   project.AuditPhaseResolve,
				Kind:    project.AuditKindResult,
				Summary: fmt.Sprintf("resolve failed: %v", err),
				Status:  "error",
				Data:    map[string]string{"error": err.Error()},
			}, now)
			auditFinalize(baseDir, tail)
		}
		return 1
	}
	if auditing {
		status := "resolved"
		recipe := ""
		if decision.Selected != nil {
			recipe = decision.Selected.Recipe
		}
		if decision.Status != "" {
			status = string(decision.Status)
		}
		tail := project.NewAuditSession(sessionID, "", "", now)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseResolve,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("resolve %s: recipe %s fingerprint %s", status, recipe, decision.IntentFingerprint),
			Status:  status,
			Data: map[string]string{
				"decision_status":    status,
				"recipe":             recipe,
				"intent_fingerprint": decision.IntentFingerprint,
				"unresolved_count":   fmt.Sprintf("%d", len(decision.UnresolvedDimensions)),
				"candidate_count":    fmt.Sprintf("%d", len(decision.Candidates)),
			},
		}, now)
		auditFinalize(baseDir, tail)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(decision, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	fmt.Print(resolver.Explain(*decision))
	if *verbose {
		fmt.Printf("fingerprint: %s\n", decision.IntentFingerprint)
	}
	return 0
}
