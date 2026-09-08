package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

func runAdd(args []string) int {
	fs := newFlagSet("add")
	projectDir := fs.String("project", ".", "materialized project directory")
	requirement := fs.String("requirement", "", "product requirement text to record")
	scope := fs.String("scope", "required_now", "requirement scope: required_now|planned_later")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *requirement == "" {
		fmt.Fprintln(os.Stderr, "add: --requirement is required")
		return 2
	}
	res, err := app.AddProjectRequirement(*projectDir, *requirement, *scope, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "add: %v\n", err)
		return 1
	}
	fmt.Printf("requirement recorded: %s\n", res.ID)
	if res.Warning != "" {
		fmt.Fprintf(os.Stderr, "warning: %s\n", res.Warning)
	}
	return 0
}
