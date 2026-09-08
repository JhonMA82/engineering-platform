package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

func runMaterialize(args []string) int {
	fs := newFlagSet("materialize")
	planPath := fs.String("plan", "", "path to materialization plan JSON file")
	output := fs.String("output", "", "project output directory (new or empty)")
	intentPath := fs.String("intent", "", "optional path to project intent JSON file")
	decisionPath := fs.String("decision", "", "optional path to architecture decision JSON file")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *planPath == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "materialize: --plan and --output are required")
		return 2
	}
	planRaw, err := os.ReadFile(*planPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
		return 1
	}
	var intentRaw, decisionRaw []byte
	if *intentPath != "" {
		intentRaw, err = os.ReadFile(*intentPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
			return 1
		}
	}
	if *decisionPath != "" {
		decisionRaw, err = os.ReadFile(*decisionPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
			return 1
		}
	}
	manifest, err := app.MaterializeProject(planRaw, intentRaw, decisionRaw, *catalogDir, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "materialize: %v\n", err)
		return 1
	}
	fmt.Printf("materialized %s — recipe %s (%d components, %d files)\n",
		manifest.Project, manifest.Recipe, len(manifest.Components), len(manifest.Files))
	for _, c := range manifest.Components {
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
	}
	fmt.Printf("fingerprint: %s\n", manifest.PlanFingerprint)
	return 0
}
