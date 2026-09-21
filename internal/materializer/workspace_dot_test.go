package materializer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/project"
)

// TestMaterializeIntoInitWorkspaceWithDotCwd is the --output . regression:
// CWD is the init workspace and the output is the relative ".". Before the
// fix the stash temp dir landed inside the workspace and moveTopEntries
// attempted rename .eng-stash-X into .eng-stash-X/.eng-stash-X (EINVAL).
func TestMaterializeIntoInitWorkspaceWithDotCwd(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	ws := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(ws, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	intentPath := filepath.Join(ws, "project-intent.json")
	if err := os.WriteFile(intentPath, []byte(`{"project":"test"}`), 0o644); err != nil {
		t.Fatal(err)
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

	if _, err := Materialize(happyRequest(t, cat, plan, ".")); err != nil {
		t.Fatalf("materialize with --output . : %v", err)
	}

	// Workspace files survive alongside the product.
	for _, p := range []string{
		"project-intent.json",
		".engineering/bootstrap.json",
		"AGENTS.md",
		".engineering/project.json",
		".engineering/provenance.json",
	} {
		if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(p))); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
	if findings, err := project.Doctor(ws); err != nil || project.HasErrors(findings) {
		t.Fatalf("doctor: findings=%v err=%v", findings, err)
	}
}
