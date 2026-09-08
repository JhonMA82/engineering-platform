package cli

import (
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/version"
)

// TestVersionCommandDisplaysCoreAndCatalog proves the §1.3 report: the core
// release line (single canonical source) and the catalog line are distinct
// fields, plus commit/build-date provenance. Offline: it reads the in-repo
// catalog only.
func TestVersionCommandDisplaysCoreAndCatalog(t *testing.T) {
	out, code := VersionString()
	if code != 0 {
		t.Fatalf("version exit = %d, output:\n%s", code, out)
	}
	for _, want := range []string{
		"Core: " + version.CoreVersion,
		"Catalog: ",
		"Catalog schema: ",
		"Commit: ",
		"Build date: ",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("version output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Catalog: unknown") {
		t.Fatalf("base catalog should load in version report:\n%s", out)
	}
}
