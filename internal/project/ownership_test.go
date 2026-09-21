package project

import (
	"os"
	"path/filepath"
	"testing"
)

// AiContext-owned .engineering files must never enter the manifest and
// `eng doctor` must never require them.
func TestFilterManifestFilesDropsAicontextOwned(t *testing.T) {
	files := []string{
		"apps/web/index.html",
		".engineering/project.json",
		".engineering/aicontext.toml",
		".engineering/PROJECT_STATE.md",
		".engineering/PATTERNS.md",
		".engineering/consistency.yml",
		".engineering/subprojects.yml",
		".engineering/rules/ast-grep/no-console.yml",
		"AGENTS.md",
	}
	got := FilterManifestFiles(files)
	want := []string{
		"apps/web/index.html",
		".engineering/project.json",
		"AGENTS.md",
	}
	if len(got) != len(want) {
		t.Fatalf("filtered = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("filtered = %v, want %v", got, want)
		}
	}
}

func TestIsAiContextOwnedTable(t *testing.T) {
	owned := []string{
		".engineering/aicontext.toml",
		".engineering/PROJECT_STATE.md",
		".engineering/PATTERNS.md",
		".engineering/consistency.yml",
		".engineering/subprojects.yml",
		".engineering/rules/ast-grep/x.yml",
	}
	for _, f := range owned {
		if !IsAiContextOwned(f) {
			t.Errorf("%q should be AiContext-owned", f)
		}
	}
	engineering := []string{
		".engineering/project.json",
		".engineering/provenance.json",
		".engineering/project-map.json",
		".engineering/project-intent.json",
		".engineering/architecture-decision.json",
		".engineering/materialization-plan.json",
		".engineering/implementation-brief.md",
		".engineering/handoff.json",
		".engineering/runs/2026-01-01-init.json",
		".engineering/bootstrap.json",
		"AGENTS.md",
		"ARCHITECTURE.md",
	}
	for _, f := range engineering {
		if IsAiContextOwned(f) {
			t.Errorf("%q should NOT be AiContext-owned", f)
		}
	}
}

// The root AGENTS.md refresh must preserve AiContext marked blocks.
func TestMergeAgentsPreservingAicontext(t *testing.T) {
	existing := "# demo — agent router\n\nSome engineering content.\n\n" +
		"<!-- aicontext:routing:start -->\nOLD ROUTING\n<!-- aicontext:routing:end -->\n\n" +
		"<!-- aicontext:context:start -->\nOLD CONTEXT\n<!-- aicontext:context:end -->\n"
	generated := "# demo — agent router\n\nNew engineering content.\n"
	merged := MergeAgentsPreservingAicontext(existing, generated)
	for _, want := range []string{
		"New engineering content.",
		"<!-- aicontext:routing:start -->\nOLD ROUTING\n<!-- aicontext:routing:end -->",
		"<!-- aicontext:context:start -->\nOLD CONTEXT\n<!-- aicontext:context:end -->",
	} {
		if !contains(merged, want) {
			t.Errorf("merged AGENTS.md lacks %q\n%s", want, merged)
		}
	}
	if contains(merged, "Some engineering content.") {
		t.Errorf("stale engineering content should be replaced\n%s", merged)
	}
	if got := MergeAgentsPreservingAicontext("", generated); got != generated {
		t.Errorf("no existing blocks should return generated unchanged")
	}
}

// Doctor must stay green when AiContext state exists alongside the project.
func TestDoctorIgnoresAicontextState(t *testing.T) {
	dir := seedProject(t)
	for _, f := range []string{
		".engineering/aicontext.toml",
		".engineering/PROJECT_STATE.md",
		".engineering/PATTERNS.md",
		".engineering/consistency.yml",
		".engineering/subprojects.yml",
	} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(f)), []byte("owned"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, ".engineering", "rules", "ast-grep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".engineering", "rules", "ast-grep", "x.yml"), []byte("rules"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := Doctor(dir)
	if err != nil {
		t.Fatal(err)
	}
	if HasErrors(findings) {
		t.Fatalf("doctor should ignore aicontext state, got %+v", findings)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
