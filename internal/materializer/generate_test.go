package materializer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// TestGenerateAppName keeps the {name} placeholder portable: destination
// basenames that could inject separators or escapes are rejected before
// any command runs.
func TestGenerateAppName(t *testing.T) {
	for _, tt := range []struct {
		dest string
		want string
		ok   bool
	}{
		{"apps/mobile", "mobile", true},
		{"mobile", "mobile", true},
		{"apps/my-app_v2", "my-app_v2", true},
		{"apps/My.App-1_2", "My.App-1_2", true},
		{"apps/my app", "", false},
		{"apps/a;b", "", false},
		{"apps/..", "", false},
		{"apps/.", "", false},
	} {
		t.Run(tt.dest, func(t *testing.T) {
			got, err := generateAppName(tt.dest)
			if tt.ok && err != nil {
				t.Fatalf("generateAppName: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("generateAppName(%q) = %q, want error", tt.dest, got)
			}
			if tt.ok && got != tt.want {
				t.Fatalf("generateAppName(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

// requireGit skips command-execution tests gracefully when git is absent
// (or -short is set): generation tests must never fail for environment
// reasons, mirroring the network-gated pilot policy.
func requireGit(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("short mode: skipping command-execution test")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH: skipping command-execution test")
	}
}

// initFixtureRepo builds a committable fixture source in a temp dir.
func initFixtureRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("add", "-A")
	run("commit", "-qm", "fixture")
	return dir
}

// TestRunGenerateOffline exercises the generic generate operation without
// network: git clones a local fixture repo into the placeholder directory,
// the output is validated, and failures leave no residue.
func TestRunGenerateOffline(t *testing.T) {
	requireGit(t)
	ctx := context.Background()
	src := initFixtureRepo(t, map[string]string{"app.json": `{"name":"fixture"}` + "\n"})

	t.Run("generated output validates and carries files", func(t *testing.T) {
		spec := &domain.GenerateSpec{
			Run:    domain.AdapterCommand{Run: []string{"git", "clone", "-q", src, "{name}"}},
			Output: "{name}",
		}
		workRoot := t.TempDir()
		out, err := RunGenerate(ctx, spec, "apps/mobile", workRoot, time.Minute)
		if err != nil {
			t.Fatalf("RunGenerate: %v", err)
		}
		raw, err := os.ReadFile(filepath.Join(out, "app.json"))
		if err != nil {
			t.Fatalf("generated output missing app.json: %v", err)
		}
		if !strings.Contains(string(raw), "fixture") {
			t.Fatalf("unexpected generated content: %s", raw)
		}
	})

	t.Run("failing command cleans the work dir", func(t *testing.T) {
		spec := &domain.GenerateSpec{
			Run:    domain.AdapterCommand{Run: []string{"git", "clone", "-q", filepath.Join(src, "nope"), "{name}"}},
			Output: "{name}",
		}
		workRoot := t.TempDir()
		if _, err := RunGenerate(ctx, spec, "apps/mobile", workRoot, time.Minute); err == nil {
			t.Fatal("expected generation failure, got nil")
		}
		entries, err := os.ReadDir(workRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("work root keeps residue: %v", entries)
		}
	})

	t.Run("missing output directory fails", func(t *testing.T) {
		spec := &domain.GenerateSpec{
			Run:    domain.AdapterCommand{Run: []string{"git", "clone", "-q", "--no-checkout", src, "other"}},
			Output: "{name}",
		}
		if _, err := RunGenerate(ctx, spec, "apps/mobile", t.TempDir(), time.Minute); err == nil {
			t.Fatal("expected missing-output failure, got nil")
		}
	})

	t.Run("shell commands are refused", func(t *testing.T) {
		spec := &domain.GenerateSpec{
			Run:    domain.AdapterCommand{Run: []string{"sh", "-c", "echo hi"}},
			Output: "{name}",
		}
		if _, err := RunGenerate(ctx, spec, "apps/mobile", t.TempDir(), time.Minute); err == nil {
			t.Fatal("expected shell refusal, got nil")
		}
	})
}

// TestNoBoilerplateSpecialCases guards the H4 acceptance: no core package
// selects behavior by boilerplate identity (no `if boilerplate == ignite`).
// Catalog data may name foundations; core code must not.
func TestNoBoilerplateSpecialCases(t *testing.T) {
	root := repoRoot(t)
	pkgs := []string{"materializer", "composer", "resolver", "catalog", "domain", "planner"}
	var hits []string
	for _, pkg := range pkgs {
		dir := filepath.Join(root, "internal", pkg)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(raw), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") {
					continue
				}
				if strings.Contains(strings.ToLower(line), "ignite") {
					hits = append(hits, pkg+"/"+e.Name()+": "+trimmed)
				}
			}
		}
	}
	if len(hits) > 0 {
		t.Fatalf("core references a boilerplate identity:\n%s", strings.Join(hits, "\n"))
	}
}

// repoRoot walks up from the working directory to the repo root (go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root (go.mod) not found")
		}
		dir = parent
	}
}
