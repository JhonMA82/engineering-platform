package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// startCatalogOverlay builds the minimal offline overlay for the start tests:
// tanstack-admin serves web-admin from the local fixture source.
func startCatalogOverlay(t *testing.T) string {
	t.Helper()
	webSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-web"))
	if err != nil {
		t.Fatal(err)
	}
	webPin := strings.TrimSpace(string(mustRead(t, filepath.Join(webSrc, "PIN"))))
	overlay := t.TempDir()
	mustWrite(t, filepath.Join(overlay, "metadata.json"),
		`{"catalog_version":"0.0.0-start","min_core_version":"1.0.0","schema_version":1}`)
	if err := os.MkdirAll(filepath.Join(overlay, "boilerplates"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(overlay, "boilerplates", "tanstack-admin.json"), fmt.Sprintf(`{
      "id": "tanstack-admin", "repo": "https://example.invalid/tanstack-admin", "pin": %q,
      "adapter": {"name": "tanstack", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
      "source": {"type": "local", "path": %q},
      "delivery_status": "stable", "decision_status": "curated",
      "provides": {"surfaces": ["web-admin"], "capabilities": []},
      "tech_tags": ["tanstack"], "included_features": []
    }`, webPin, webSrc))
	return overlay
}

func writeIntent(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "intent.json")
	mustWrite(t, path, body)
	return path
}

const startAdminIntent = `{"schema_version":1,"name":"backoffice","problem":"operators manage records",` +
	`"surfaces":[{"kind":"web-admin","access":"authenticated","scope":"required_now"}]}`

// TestStartDryRunWritesNothing proves --dry-run resolves and plans with zero
// filesystem writes: the output directory must not appear.
func TestStartDryRunWritesNothing(t *testing.T) {
	overlay := startCatalogOverlay(t)
	intent := writeIntent(t, startAdminIntent)
	out := filepath.Join(t.TempDir(), "proj")
	if code := Main([]string{"start", "--intent", intent, "--output", out, "--dry-run", "--catalog-dir", overlay}); code != 0 {
		t.Fatalf("start --dry-run exit = %d, want 0", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote %s (or stat failed): %v", out, err)
	}
}

// TestStartFullChain proves the P7 convenience chain: plan → materialize →
// doctor through the CLI, ending in a navigable project.
func TestStartFullChain(t *testing.T) {
	overlay := startCatalogOverlay(t)
	intent := writeIntent(t, startAdminIntent)
	out := filepath.Join(t.TempDir(), "proj")
	if code := Main([]string{"start", "--intent", intent, "--output", out, "--catalog-dir", overlay}); code != 0 {
		t.Fatalf("start exit = %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(out, "AGENTS.md")); err != nil {
		t.Errorf("expected AGENTS.md: %v", err)
	}
	if code := Main([]string{"doctor", "--project", out}); code != 0 {
		t.Errorf("doctor after start exit = %d, want 0", code)
	}
}

// TestStartUnresolvedFails proves an ambiguous intent exits non-zero with no
// project directory created.
func TestStartUnresolvedFails(t *testing.T) {
	overlay := startCatalogOverlay(t)
	intent := writeIntent(t, `{"schema_version":1,"name":"vague","problem":"something"}`)
	out := filepath.Join(t.TempDir(), "proj")
	if code := Main([]string{"start", "--intent", intent, "--output", out, "--catalog-dir", overlay}); code == 0 {
		t.Fatal("start on unresolved decision should fail")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed start wrote %s (or stat failed): %v", out, err)
	}
}

// TestCatalogListAndShow smokes the catalog read surface: bare list default,
// explicit list, show hits across kinds, and a miss exit.
func TestCatalogListAndShow(t *testing.T) {
	if code := Main([]string{"catalog"}); code != 0 {
		t.Errorf("bare catalog exit = %d, want 0", code)
	}
	if code := Main([]string{"catalog", "list"}); code != 0 {
		t.Errorf("catalog list exit = %d, want 0", code)
	}
	for _, id := range []string{"GP-06", "tanstack-admin", "web-admin", "shared-backend", "postgresql-managed"} {
		if code := Main([]string{"catalog", "show", id}); code != 0 {
			t.Errorf("catalog show %s exit = %d, want 0", id, code)
		}
	}
	if code := Main([]string{"catalog", "show", "GP-06", "--json"}); code != 0 {
		t.Errorf("catalog show --json exit = %d, want 0", code)
	}
	if code := Main([]string{"catalog", "show", "no-such-id"}); code == 0 {
		t.Error("catalog show on unknown id should fail")
	}
	if code := Main([]string{"catalog", "validate"}); code != 0 {
		t.Errorf("catalog validate exit = %d, want 0", code)
	}
}
