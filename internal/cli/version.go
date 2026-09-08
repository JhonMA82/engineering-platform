package cli

import (
	"fmt"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/version"
)

// VersionString renders the §1.3 version report: the core release line comes
// from the single canonical source (internal/version, stamped via ldflags on
// release builds) and the catalog line from the loaded catalog. Separated
// from runVersion so tests can assert the format without capturing stdout.
func VersionString() (string, int) {
	cat, err := catalog.Load("")
	if err != nil {
		return fmt.Sprintf("Core: %s\nCatalog: unknown: %v\nCommit: %s\nBuild date: %s\n",
			version.CoreVersion, err, version.Commit, version.BuildDate), 1
	}
	return fmt.Sprintf("Core: %s\nCatalog: %s\nCatalog schema: %d\nCommit: %s\nBuild date: %s\n",
		version.CoreVersion, cat.CatalogVersion, cat.SchemaVersion, version.Commit, version.BuildDate), 0
}

func runVersion(args []string) int {
	if len(args) != 0 {
		fmt.Println("usage: eng version")
		return 2
	}
	out, code := VersionString()
	fmt.Print(out)
	return code
}
