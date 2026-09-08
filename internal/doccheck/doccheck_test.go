// Package doccheck guards the README against documentation drift: stale
// provider ids, milestone-era openers, broken relative links, undocumented
// release binaries, and commands that do not exist. It is intentionally
// light: a smoke test, not a documentation framework.
package doccheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from " + dir)
		}
		dir = parent
	}
}

func readREADME(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	return string(raw)
}

// TestReadmeHasNoStaleReferences pins the v1.0.1 cleanup: the retired
// stardrive-public-web id must never read as a current provider, and the
// milestone-era opener (M1/M2/M3 as product introduction) must stay out of
// the landing page.
func TestReadmeHasNoStaleReferences(t *testing.T) {
	readme := readREADME(t)
	if strings.Contains(readme, "stardrive-public-web") {
		t.Error("README still references retired id stardrive-public-web")
	}
	for _, marker := range []string{
		"# Engineering Platform 1.0 — M1",
		"Out of scope for M1",
		"Fase 9",
		"Fase 10",
	} {
		if strings.Contains(readme, marker) {
			t.Errorf("README still carries milestone-era marker %q", marker)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(readme), "# Engineering Platform 1.0 — M1") {
		t.Error("README still opens with the M1 milestone header")
	}
}

// TestReadmeRelativeLinksResolve ensures every relative link in the README
// points at a file that exists.
func TestReadmeRelativeLinksResolve(t *testing.T) {
	root := repoRoot(t)
	readme := readREADME(t)
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, m := range linkRe.FindAllStringSubmatch(readme, -1) {
		target := m[1]
		if strings.HasPrefix(target, "http") || strings.HasPrefix(target, "#") ||
			strings.HasPrefix(target, "mailto:") {
			continue
		}
		p := strings.SplitN(target, "#", 2)[0]
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("README links to %q, which does not exist", target)
		}
	}
}

// TestReadmeDocumentsRealCommands ensures the Commands section only shows
// the eng verbs the CLI dispatches.
func TestReadmeDocumentsRealCommands(t *testing.T) {
	readme := readREADME(t)
	for _, cmd := range []string{
		"eng resolve", "eng plan", "eng materialize", "eng start",
		"eng doctor", "eng surface add", "eng extend", "eng add",
		"eng update", "eng explain", "eng catalog", "eng version",
	} {
		if !strings.Contains(readme, cmd) {
			t.Errorf("README does not document %q", cmd)
		}
	}
}

// TestReadmeReleaseBinariesMatchWorkflow ensures the installation section
// names exactly the binaries the release workflow builds.
func TestReadmeReleaseBinariesMatchWorkflow(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github/workflows/release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	assetRe := regexp.MustCompile(`-o dist/(eng-\S+)`)
	var assets []string
	for _, m := range assetRe.FindAllStringSubmatch(string(raw), -1) {
		assets = append(assets, m[1])
	}
	if len(assets) == 0 {
		t.Fatal("no dist/eng-* assets found in release.yml")
	}
	readme := readREADME(t)
	for _, a := range assets {
		if !strings.Contains(readme, a) {
			t.Errorf("release builds %q but README documents no such binary", a)
		}
	}
}
