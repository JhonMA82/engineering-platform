package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

func runPlan(args []string) int {
	fs := newFlagSet("plan")
	input := fs.String("input", "", "path to intent JSON file")
	asJSON := fs.Bool("json", false, "print the full materialization plan as JSON")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "plan: --input is required")
		return 2
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "plan: %v\n", err)
		return 1
	}
	_, _, plan, err := app.PlanProject(raw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "plan: %v\n", err)
		return 1
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
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
	}
	fmt.Printf("fingerprint: %s\n", plan.Fingerprint)
}
