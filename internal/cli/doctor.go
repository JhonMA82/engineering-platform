package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/project"
)

func runDoctor(args []string) int {
	fs := newFlagSet("doctor")
	projectDir := fs.String("project", ".", "materialized project directory")
	asJSON := fs.Bool("json", false, "print findings as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	findings, err := app.DoctorProject(*projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor: %v\n", err)
		return 1
	}
	if *asJSON {
		out, _ := json.MarshalIndent(doctorJSON(findings), "", "  ")
		fmt.Println(string(out))
	} else {
		printDoctorHuman(*projectDir, findings)
	}
	if project.HasErrors(findings) {
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
