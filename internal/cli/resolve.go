package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

func runResolve(args []string) int {
	fs := newFlagSet("resolve")
	input := fs.String("input", "", "path to intent JSON file")
	asJSON := fs.Bool("json", false, "print the full decision as JSON")
	verbose := fs.Bool("verbose", false, "include fingerprint and candidate detail")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "resolve: --input is required")
		return 2
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve: %v\n", err)
		return 1
	}
	decision, err := app.ResolveProject(raw, *catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve: %v\n", err)
		return 1
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
