package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

func runSurfaceAdd(args []string) int {
	fs := newFlagSet("surface add")
	projectDir := fs.String("project", ".", "materialized project directory")
	surface := fs.String("surface", "", "surface id to add (must be known to the catalog)")
	provider := fs.String("provider", "", "optional boilerplate id serving the surface")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *surface == "" {
		fmt.Fprintln(os.Stderr, "surface add: --surface is required")
		return 2
	}
	res, err := app.SurfaceAddProject(*projectDir, *surface, *provider, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "surface add: %v\n", err)
		return 1
	}
	fmt.Printf("surface added: recipe %s\n", res.Recipe)
	for _, c := range res.Added {
		fmt.Printf("  - %s <- %s@%s (%s)\n", c.Destination, c.Boilerplate, c.Pin, c.Surface)
	}
	fmt.Printf("fingerprint: %s\n", res.PlanFingerprint)
	return 0
}
