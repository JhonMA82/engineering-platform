package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

func runUpdate(args []string) int {
	fs := newFlagSet("update")
	projectDir := fs.String("project", ".", "materialized project directory")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	asJSON := fs.Bool("json", false, "print the report as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	report, err := app.UpdateProjectReport(*projectDir, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		return 1
	}
	if *asJSON {
		out, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	fmt.Printf("%-24s %-22s %-22s %-10s %s\n", "COMPONENT", "CURRENT", "CATALOG", "STRATEGY", "ACTION")
	for _, e := range report.Entries {
		fmt.Printf("%-24s %-22s %-22s %-10s %s\n",
			e.Component+" ("+e.Destination+")", e.Current, e.Catalog, e.Strategy, e.Action)
	}
	fmt.Println(report.Summary)
	return 0
}
