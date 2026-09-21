package cli

import (
	"io"
	"os"
	"strings"
	"testing"
)

// TestInitAlreadyUpToDateNamesTheAgent proves the re-run report names the
// integration that is actually current: OpenCode workspaces must not claim
// "PI integration is up to date".
func TestInitAlreadyUpToDateNamesTheAgent(t *testing.T) {
	cases := map[string]string{
		"pi":       "PI integration is up to date.",
		"opencode": "OpenCode integration is up to date.",
		"none":     "Workspace state is up to date.",
	}
	for agent, want := range cases {
		dir := t.TempDir()
		if code := runInitIn(t, dir, "--agent", agent); code != 0 {
			t.Fatalf("init --agent %s: exit %d", agent, code)
		}
		out := runInitInCapture(t, dir, "--agent", agent)
		if !strings.Contains(out, want) {
			t.Fatalf("second init --agent %s lacks %q:\n%s", agent, want, out)
		}
		if agent == "opencode" && strings.Contains(out, "PI integration") {
			t.Fatalf("opencode re-run must not mention PI integration:\n%s", out)
		}
	}
}

// runInitIn runs runInit with cwd pointed at dir (runInit resolves the
// workspace from os.Getwd) and restores the caller's directory.
func runInitIn(t *testing.T, dir string, args ...string) int {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()
	return runInit(args)
}

func runInitInCapture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := runInitIn(t, dir, args...)
	_ = w.Close()
	os.Stdout = old
	if code != 0 {
		t.Fatalf("runInit exit = %d", code)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
