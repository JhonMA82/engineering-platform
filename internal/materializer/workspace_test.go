package materializer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/project"
)

// TestMaterializeIntoInitWorkspace covers bootstrap RFC section 10: a
// workspace prepared with eng init (bootstrap state + local .pi files plus
// user discovery files) accepts materialization. Pre-existing files survive
// the atomic commit and the result still passes doctor.
func TestMaterializeIntoInitWorkspace(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	out := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(out, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	intentPath := filepath.Join(out, "project-intent.json")
	if err := os.WriteFile(intentPath, []byte(`{"project":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("materialize into init workspace: %v", err)
	}

	// User and bootstrap files survive.
	for _, p := range []string{
		"project-intent.json",
		".engineering/bootstrap.json",
		".pi/prompts/newproject.md",
		".pi/skills/project-discovery/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(p))); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
	// Product files land too.
	if _, err := os.Stat(filepath.Join(out, "AGENTS.md")); err != nil {
		t.Fatalf("stat AGENTS.md: %v", err)
	}
	if findings, err := project.Doctor(out); err != nil || project.HasErrors(findings) {
		t.Fatalf("doctor: findings=%v err=%v", findings, err)
	}
}

// TestMaterializeRefusesCollisionInInitWorkspace guards bootstrap RFC
// section 24: when a user file in an init workspace collides with a path
// the plan would generate, materialization refuses before touching
// anything. No user data is lost and no partial project is committed.
func TestMaterializeRefusesCollisionInInitWorkspace(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()

	// Calibrate: find a root-level path the plan generates.
	probe := filepath.Join(t.TempDir(), "probe")
	if _, err := Materialize(happyRequest(t, cat, plan, probe)); err != nil {
		t.Fatalf("probe materialize: %v", err)
	}
	staged, err := ListFiles(probe)
	if err != nil {
		t.Fatal(err)
	}
	var victim string
	for _, f := range staged {
		if !strings.Contains(f, "/") {
			victim = f
			break
		}
	}
	if victim == "" {
		t.Fatal("probe project has no root-level file to collide with")
	}

	out := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(out, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	userData := []byte("precious user content")
	if err := os.WriteFile(filepath.Join(out, filepath.FromSlash(victim)), userData, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Materialize(happyRequest(t, cat, plan, out)); err == nil {
		t.Fatal("expected overwrite refusal")
	} else if !strings.Contains(err.Error(), "overwrite") {
		t.Fatalf("expected overwrite complaint, got %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(victim)))
	if err != nil || string(raw) != string(userData) {
		t.Fatal("user file must survive refused materialization")
	}
	if _, err := os.Stat(filepath.Join(out, ".engineering", "bootstrap.json")); err != nil {
		t.Fatalf("bootstrap must survive refused materialization: %v", err)
	}
}

// TestCleanupBootstrap covers bootstrap RFC sections 11-13: after success,
// only eng-owned bootstrap resources disappear. Product code, agent
// tooling and minimal provenance stay.
func TestCleanupBootstrap(t *testing.T) {
	cat := twoComponentCatalog(t)
	plan := twoComponentPlan()
	out := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := InitWorkspace(out, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}

	report, err := CleanupBootstrap(out)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if !report.Cleaned {
		t.Fatal("expected cleanup to run in an init workspace")
	}
	if len(report.Removed) != 3 {
		t.Fatalf("want 3 removed files, got %v", report.Removed)
	}

	for _, p := range []string{
		".engineering/bootstrap.json",
		".pi/prompts/newproject.md",
		".pi/skills/project-discovery/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(p))); !os.IsNotExist(err) {
			t.Fatalf("%s must be gone", p)
		}
	}
	// Minimal provenance and product stay.
	for _, p := range []string{
		".engineering/project.json",
		".engineering/provenance.json",
		"AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(p))); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, ".pi")); !os.IsNotExist(err) {
		t.Fatal(".pi must be pruned when it holds only bootstrap files")
	}

	second, err := CleanupBootstrap(out)
	if err != nil {
		t.Fatalf("second cleanup: %v", err)
	}
	if second.Cleaned {
		t.Fatal("second cleanup must be a no-op")
	}
}

// TestCleanupKeepsUserPiFiles proves cleanup is ownership-based: user files
// inside .pi/ survive and their parent directories are preserved.
func TestCleanupKeepsUserPiFiles(t *testing.T) {
	out := t.TempDir()
	if _, err := InitWorkspace(out, InitOptions{Agent: "pi", EngVersion: "test"}); err != nil {
		t.Fatalf("init: %v", err)
	}
	own := filepath.Join(out, ".pi", "prompts", "my-own.md")
	if err := os.WriteFile(own, []byte("# mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := CleanupBootstrap(out); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if _, err := os.Stat(own); err != nil {
		t.Fatalf("user .pi file must survive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, ".pi", "prompts")); err != nil {
		t.Fatalf("non-empty user dir must survive: %v", err)
	}
}

// TestCleanupNoopWithoutBootstrap proves cleanup never touches a workspace
// that eng init did not prepare.
func TestCleanupNoopWithoutBootstrap(t *testing.T) {
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(out, "notes.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := CleanupBootstrap(out)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if report.Cleaned || len(report.Removed) != 0 {
		t.Fatalf("want no-op, got %+v", report)
	}
}
