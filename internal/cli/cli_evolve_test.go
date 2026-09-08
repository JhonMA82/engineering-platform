package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/app"
)

// TestEvolveCommandRouting smokes the Fase 10 command surface: eng surface
// add vs eng add dispatch, extend/update wiring and usage exits. Behavior is
// asserted through exit codes on a real fixture project.
func TestEvolveCommandRouting(t *testing.T) {
	webSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-web"))
	if err != nil {
		t.Fatal(err)
	}
	apiSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-api"))
	if err != nil {
		t.Fatal(err)
	}
	webPin := strings.TrimSpace(string(mustRead(t, filepath.Join(webSrc, "PIN"))))
	apiPin := strings.TrimSpace(string(mustRead(t, filepath.Join(apiSrc, "PIN"))))
	overlay := t.TempDir()
	mustWrite(t, filepath.Join(overlay, "metadata.json"),
		`{"catalog_version":"0.0.0-cli","min_core_version":"1.0.0","schema_version":1}`)
	bpDir := filepath.Join(overlay, "boilerplates")
	if err := os.MkdirAll(bpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overlaySpecs := map[string][3]string{
		"tanstack-admin":             {webPin, webSrc, `["web-admin"]|`},
		"hono-api":                   {apiPin, apiSrc, `["api"]|["shared-backend"]`},
		"tanstack-transactional-pwa": {webPin, webSrc, `["public-web", "public-intake", "mobile-native"]|`},
	}
	for id, spec := range overlaySpecs {
		pin, src := spec[0], spec[1]
		parts := strings.SplitN(spec[2], "|", 2)
		surfaces, caps := parts[0], "[]"
		if len(parts) == 2 && parts[1] != "" {
			caps = parts[1]
		}
		mustWrite(t, filepath.Join(bpDir, id+".json"), fmt.Sprintf(`{
          "id": %q, "repo": "https://example.invalid/%s", "pin": %q,
          "adapter": {"name": "fixture", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
          "source": {"type": "local", "path": %q},
          "delivery_status": "stable", "decision_status": "curated",
          "provides": {"surfaces": %s, "capabilities": %s},
          "tech_tags": ["fixture"], "included_features": []
        }`, id, id, pin, src, surfaces, caps))
	}
	intent := `{"schema_version":1,"name":"cli-probe","problem":"p",` +
		`"surfaces":[{"kind":"web-admin","scope":"required_now"},{"kind":"api","scope":"required_now"}]}`
	decision, _, plan, err := app.PlanProject([]byte(intent), overlay)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	planJSON, _ := json.Marshal(plan)
	decisionJSON, _ := json.Marshal(decision)
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := app.MaterializeProject(planJSON, []byte(intent), decisionJSON, overlay, out); err != nil {
		t.Fatalf("materialize: %v", err)
	}

	if code := Main([]string{"surface", "add", "--project", out, "--surface", "public-intake", "--catalog-dir", overlay}); code != 0 {
		t.Errorf("surface add exit = %d, want 0", code)
	}
	if code := Main([]string{"add", "--project", out, "--requirement", "cli smoke requirement", "--catalog-dir", overlay}); code != 0 {
		t.Errorf("add exit = %d, want 0", code)
	}
	if code := Main([]string{"update", "--project", out, "--catalog-dir", overlay}); code != 0 {
		t.Errorf("update exit = %d, want 0", code)
	}
	if code := Main([]string{"update", "--project", out, "--catalog-dir", overlay, "--json"}); code != 0 {
		t.Errorf("update --json exit = %d, want 0", code)
	}
	if code := Main([]string{"extend", "--project", out, "--surface", "web-admin", "--catalog-dir", overlay}); code == 0 {
		t.Errorf("extend of a required surface should fail")
	}
	if code := Main([]string{"surface"}); code != 2 {
		t.Errorf("bare surface exit = %d, want 2", code)
	}
	if code := Main([]string{"surface", "remove", "--project", out}); code != 2 {
		t.Errorf("unknown surface subcommand exit = %d, want 2", code)
	}
	if code := Main([]string{"add", "--project", out}); code != 2 {
		t.Errorf("add without --requirement exit = %d, want 2", code)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
