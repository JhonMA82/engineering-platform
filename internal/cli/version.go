package cli

import (
	"fmt"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/catalog"
)

// BinaryVersion is the eng binary release line (single source: app.CoreVersion).
const BinaryVersion = app.CoreVersion

// BuildCommit and BuildDate carry reproducible-build provenance. Release
// builds stamp them via ldflags:
//
//	go build -X .../internal/cli.BuildCommit=<sha> -X .../internal/cli.BuildDate=<date> ./cmd/eng
//
// Unset values report as "unknown" per the §30 "cuando estén disponibles" rule.
var (
	BuildCommit = "unknown"
	BuildDate   = "unknown"
)

func runVersion(args []string) int {
	if len(args) != 0 {
		fmt.Println("usage: eng version")
		return 2
	}
	cat, err := catalog.Load("")
	if err != nil {
		fmt.Printf("eng %s (catalog: unknown: %v) commit %s date %s\n", BinaryVersion, err, BuildCommit, BuildDate)
		return 1
	}
	fmt.Printf("eng %s (catalog %s, min-core %s, schema %d) commit %s date %s\n",
		BinaryVersion, cat.CatalogVersion, cat.MinCoreVersion, cat.SchemaVersion, BuildCommit, BuildDate)
	return 0
}
