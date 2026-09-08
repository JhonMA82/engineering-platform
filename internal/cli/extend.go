package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

func runExtend(args []string) int {
	fs := newFlagSet("extend")
	projectDir := fs.String("project", ".", "materialized project directory")
	surface := fs.String("surface", "", "planned_later surface to promote to required_now")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *surface == "" {
		fmt.Fprintln(os.Stderr, "extend: --surface is required")
		return 2
	}
	res, err := app.ExtendProjectScope(*projectDir, *surface, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extend: %v\n", err)
		return 1
	}
	fmt.Printf("scope extended: recipe %s\n", res.Recipe)
	for _, c := range res.Added {
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
	}
	fmt.Printf("fingerprint: %s\n", res.PlanFingerprint)
	return 0
}
