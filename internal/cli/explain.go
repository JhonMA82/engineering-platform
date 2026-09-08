package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

func runExplain(args []string) int {
	fs := newFlagSet("explain")
	input := fs.String("input", "", "path to decision JSON file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(os.Stderr, "explain: --input is required")
		return 2
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explain: %v\n", err)
		return 1
	}
	var decision domain.ArchitectureDecision
	if err := json.Unmarshal(raw, &decision); err != nil {
		fmt.Fprintf(os.Stderr, "explain: %v\n", err)
		return 1
	}
	fmt.Print(resolver.Explain(decision))
	return 0
}
