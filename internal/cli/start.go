package cli

import (
	"encoding/json"
	"fmt"
	"os"

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
func runStart(args []string) int {
	fs := newFlagSet("start")
	intentPath := fs.String("intent", "", "path to project intent JSON file")
	output := fs.String("output", "", "project output directory (new or empty)")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	dryRun := fs.Bool("dry-run", false, "print the plan without writing anything")
	asJSON := fs.Bool("json", false, "print machine-readable JSON")
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
	intentRaw, err := os.ReadFile(*intentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		return 1
	}
	decision, _, plan, err := app.PlanProject(intentRaw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		return 1
	}
	if *dryRun {
		if *asJSON {
			out, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(out))
			return 0
		}
		printPlanHuman(*plan)
		fmt.Println("(dry-run: no filesystem writes)")
		return 0
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		return 1
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		return 1
	}
	manifest, err := app.MaterializeProject(planJSON, intentRaw, decisionJSON, *catalogDir, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		return 1
	}
	findings, err := app.DoctorProject(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
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
	if project.HasErrors(findings) {
		return 1
	}
	// Bootstrap lifecycle (bootstrap RFC sections 11-13): a workspace
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
