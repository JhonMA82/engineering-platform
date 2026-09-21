package materializer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureMaterializableRefusesNestedNewDirInsideInitWorkspace is the
// reloj_checador regression: an init workspace must not gain a
// project-named subdirectory (reloj_checador/reloj_checador_escolar).
// The agent must use --output . there; the core refuses the nesting with
// an actionable message.
func TestEnsureMaterializableRefusesNestedNewDirInsideInitWorkspace(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "reloj_checador")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	nested := filepath.Join(ws, "reloj_checador_escolar")
	if _, err := ensureMaterializable(nested); err == nil {
		t.Fatal("expected refusal for nested new dir inside init workspace")
	} else if !strings.Contains(err.Error(), "inside eng-init workspace") {
		t.Fatalf("expected inside-workspace complaint, got %v", err)
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("refused output must not be created, stat = %v", err)
	}
}

// TestEnsureMaterializableRefusesEmptyNestedDirInsideInitWorkspace covers
// the mkdir-first variant: an existing empty subdirectory of a workspace
// is not a valid output either.
func TestEnsureMaterializableRefusesEmptyNestedDirInsideInitWorkspace(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "reloj_checador")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	nested := filepath.Join(ws, "reloj_checador_escolar")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureMaterializable(nested); err == nil {
		t.Fatal("expected refusal for empty nested dir inside init workspace")
	} else if !strings.Contains(err.Error(), "inside eng-init workspace") {
		t.Fatalf("expected inside-workspace complaint, got %v", err)
	}
}

// TestEnsureMaterializableNestedRelativeFromWorkspace covers the exact
// agent failure mode: CWD is the init workspace and --output is a relative
// project name. The absolute resolution still detects the workspace.
func TestEnsureMaterializableNestedRelativeFromWorkspace(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "reloj_checador")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(ws); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if _, err := ensureMaterializable("reloj_checador_escolar"); err == nil {
		t.Fatal("expected refusal for relative nested output inside init workspace")
	} else if !strings.Contains(err.Error(), "inside eng-init workspace") {
		t.Fatalf("expected inside-workspace complaint, got %v", err)
	}
}

// TestEnsureMaterializableAllowsWorkspaceAndOutsideSibling guards the
// valid cases: the workspace itself stays initMode, and an unrelated
// sibling directory is still accepted.
func TestEnsureMaterializableAllowsWorkspaceAndOutsideSibling(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "reloj_checador")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if initMode, err := ensureMaterializable(ws); err != nil || !initMode {
		t.Fatalf("workspace itself must stay initMode, initMode=%v err=%v", initMode, err)
	}
	sibling := filepath.Join(root, "otro_proyecto")
	if initMode, err := ensureMaterializable(sibling); err != nil || initMode {
		t.Fatalf("outside sibling must be allowed as new dir, initMode=%v err=%v", initMode, err)
	}
}

// TestMaterializeRefusesNestedInsideInitWorkspace proves the pipeline
// fails before staging when the agent passes a nested output.
func TestMaterializeRefusesNestedInsideInitWorkspace(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	ws := filepath.Join(t.TempDir(), "reloj_checador")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "opencode", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	nested := filepath.Join(ws, "reloj_checador_escolar")
	if _, err := Materialize(happyRequest(t, cat, plan, nested)); err == nil {
		t.Fatal("expected materialize refusal for nested output")
	} else if !strings.Contains(err.Error(), "inside eng-init workspace") {
		t.Fatalf("expected inside-workspace complaint, got %v", err)
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("refused nested output must not exist, stat = %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws, ".engineering", "bootstrap.json")); err != nil {
		t.Fatalf("workspace bootstrap must survive refused materialization: %v", err)
	}
}
