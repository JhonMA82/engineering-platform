package materializer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitWorkspaceOpenCodeFresh(t *testing.T) {
	dir := t.TempDir()
	report, err := InitWorkspace(dir, InitOptions{Agent: "opencode", EngVersion: "test"})
	if err != nil {
		t.Fatalf("InitWorkspace: %v", err)
	}
	if report.AlreadyUpToDate {
		t.Fatal("fresh init must not report AlreadyUpToDate")
	}
	if len(report.Created) != 3 {
		t.Fatalf("want 3 created files, got %v", report.Created)
	}
	for _, p := range []string{
		".opencode/commands/newproject.md",
		".opencode/skills/engineering-project-discovery/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".pi")); !os.IsNotExist(err) {
		t.Fatal(".pi must not be created for agent opencode")
	}
}

func TestInitWorkspaceOpenCodeIdempotent(t *testing.T) {
	dir := t.TempDir()
	if _, err := InitWorkspace(dir, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("first init: %v", err)
	}
	second, err := InitWorkspace(dir, InitOptions{Agent: "opencode", EngVersion: "test"})
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if !second.AlreadyUpToDate {
		t.Fatalf("second init must report AlreadyUpToDate, got %+v", second)
	}
}

func TestCleanupOpenCode(t *testing.T) {
	dir := t.TempDir()
	if _, err := InitWorkspace(dir, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	report, err := CleanupBootstrap(dir)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if !report.Cleaned || len(report.Removed) != 3 {
		t.Fatalf("want 3 removed files, got %+v", report)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(err) {
		t.Fatal(".opencode must be pruned when it holds only bootstrap files")
	}
	if _, err := os.Stat(filepath.Join(dir, ".engineering", "bootstrap.json")); !os.IsNotExist(err) {
		t.Fatal("bootstrap.json must be gone")
	}
}
