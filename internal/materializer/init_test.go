package materializer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	opencodebootstrap "github.com/jhonma82/engineering-platform/integrations/opencode"
	pibootstrap "github.com/jhonma82/engineering-platform/integrations/pi"
)

func TestInitWorkspaceFresh(t *testing.T) {
	dir := t.TempDir()
	report, err := InitWorkspace(dir, InitOptions{Agent: "pi", EngVersion: "test"})
	if err != nil {
		t.Fatalf("InitWorkspace: %v", err)
	}
	if report.AlreadyUpToDate {
		t.Fatal("fresh init must not report AlreadyUpToDate")
	}
	if len(report.Created) != 3 {
		t.Fatalf("want 3 created files, got %v", report.Created)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".engineering", "bootstrap.json"))
	if err != nil {
		t.Fatalf("read bootstrap.json: %v", err)
	}
	var state map[string]string
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("parse bootstrap.json: %v", err)
	}
	if state["agent"] != "pi" {
		t.Fatalf("agent = %q, want pi", state["agent"])
	}
	if state["bootstrapVersion"] != BootstrapVersion {
		t.Fatalf("bootstrapVersion = %q, want %q", state["bootstrapVersion"], BootstrapVersion)
	}
	if state["engVersion"] != "test" {
		t.Fatalf("engVersion = %q, want test", state["engVersion"])
	}
	if state["initializedAt"] == "" {
		t.Fatal("initializedAt must be set")
	}

	for _, p := range []string{
		".pi/prompts/newproject.md",
		".pi/skills/project-discovery/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
}

func TestInitWorkspaceIdempotent(t *testing.T) {
	dir := t.TempDir()
	if _, err := InitWorkspace(dir, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("first init: %v", err)
	}
	second, err := InitWorkspace(dir, InitOptions{Agent: "pi", EngVersion: "test"})
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if !second.AlreadyUpToDate {
		t.Fatalf("second init must report AlreadyUpToDate, got %+v", second)
	}
	if len(second.Created) != 0 || len(second.Updated) != 0 {
		t.Fatalf("idempotent run must not write, got %+v", second)
	}
}

func TestInitWorkspacePreservesUserFiles(t *testing.T) {
	dir := t.TempDir()
	custom := "# my custom prompt\n"
	target := filepath.Join(dir, ".pi", "prompts", "newproject.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without force the user file must survive.
	report, err := InitWorkspace(dir, InitOptions{Agent: "pi", EngVersion: "test"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	raw, _ := os.ReadFile(target)
	if string(raw) != custom {
		t.Fatal("user file was overwritten without --force")
	}
	if len(report.Kept) != 1 {
		t.Fatalf("want 1 kept file, got %+v", report)
	}

	// With force the managed file is repaired.
	forced, err := InitWorkspace(dir, InitOptions{Agent: "pi", EngVersion: "test", Force: true})
	if err != nil {
		t.Fatalf("forced init: %v", err)
	}
	raw, _ = os.ReadFile(target)
	if string(raw) == custom {
		t.Fatal("forced init must repair managed files")
	}
	if len(forced.Updated) == 0 {
		t.Fatalf("forced run must report updates, got %+v", forced)
	}
}

func TestInitWorkspaceNoAgent(t *testing.T) {
	dir := t.TempDir()
	report, err := InitWorkspace(dir, InitOptions{Agent: "none", EngVersion: "test"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pi")); !os.IsNotExist(err) {
		t.Fatal(".pi must not be created for agent none")
	}
	if len(report.Created) != 1 {
		t.Fatalf("want only bootstrap.json created, got %v", report.Created)
	}
}

func TestInitWorkspaceBadAgent(t *testing.T) {
	if _, err := InitWorkspace(t.TempDir(), InitOptions{Agent: "clippy"}); err == nil {
		t.Fatal("want error for unknown agent")
	}
}

// TestEmbeddedSkillsHaveFrontmatter is a regression test: agent harnesses
// refuse skills without name/description frontmatter (pi reports
// "[Skill conflicts] ... description is required"), so every embedded
// skill must carry it. Prompts need no frontmatter.
func TestEmbeddedSkillsHaveFrontmatter(t *testing.T) {
	for _, content := range []string{
		pibootstrap.DiscoverySkillMD,
		opencodebootstrap.DiscoverySkillMD,
	} {
		if !strings.HasPrefix(content, "---\n") {
			t.Fatal("embedded skill must start with YAML frontmatter")
		}
		end := strings.Index(content[len("---\n"):], "\n---")
		if end < 0 {
			t.Fatal("embedded skill frontmatter must close")
		}
		front := content[:end]
		for _, key := range []string{"name:", "description:"} {
			if !strings.Contains(front, key) {
				t.Fatalf("embedded skill frontmatter must contain %q", key)
			}
		}
	}
}
