package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/project"
)

func runDoctor(args []string) int {
	fs := newFlagSet("doctor")
	projectDir := fs.String("project", ".", "materialized project directory")
	asJSON := fs.Bool("json", false, "print findings as JSON")
	audit := addAuditFlags(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	now := time.Now().UTC()
	baseDir, sessionID, auditing := auditSession(audit, *projectDir, now)
	if auditing {
		auditRecord(baseDir, sessionID, project.AuditEvent{
			Phase:   project.AuditPhaseDoctor,
			Kind:    project.AuditKindCommand,
			Summary: fmt.Sprintf("doctor --project %s", *projectDir),
			Command: "doctor",
			Argv:    append([]string{"eng", "doctor"}, args...),
			Data:    map[string]string{"project": *projectDir},
		})
	}
	findings, err := app.DoctorProject(*projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor: %v\n", err)
		persistReport(reportBaseDir(*projectDir), project.ErrorReport("doctor", project.IntentSummary{Output: *projectDir}, err, nil, ""))
		if auditing {
			tail := project.NewAuditSession(sessionID, "", "", now)
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
		out, _ := json.MarshalIndent(doctorJSON(findings), "", "  ")
		fmt.Println(string(out))
	} else {
		printDoctorHuman(*projectDir, findings)
	}
	if auditing {
		status := "ok"
		if project.HasErrors(findings) {
			status = "error"
		}
		tail := project.NewAuditSession(sessionID, "", "", now)
		tail.AddEvent(project.AuditEvent{
			Phase:   project.AuditPhaseDoctor,
			Kind:    project.AuditKindResult,
			Summary: fmt.Sprintf("doctor %s: %d finding(s)", status, len(findings)),
			Status:  status,
			Data:    map[string]string{"findings": fmt.Sprintf("%d", len(findings)), "project": *projectDir},
		}, now)
		auditFinalize(baseDir, tail)
	}
	if project.HasErrors(findings) {
		persistReport(reportBaseDir(*projectDir), project.DoctorReport(project.IntentSummary{Output: *projectDir}, findings))
		return 1
	}
	return 0
}

// doctorJSON normalizes nil findings to an empty array so --json output is
// stable regardless of how many findings doctor reports.
func doctorJSON(findings []project.Finding) []project.Finding {
	if findings == nil {
		return []project.Finding{}
	}
	return findings
}

// printDoctorHuman renders findings the same way for doctor and the start
// chain so both commands describe the same project state identically.
func printDoctorHuman(projectDir string, findings []project.Finding) {
	if len(findings) == 0 {
		fmt.Printf("doctor: %s is consistent\n", projectDir)
		return
	}
	for _, f := range findings {
		fmt.Printf("doctor [%s] %s: %s\n", f.Severity, f.Code, f.Message)
	}
}
