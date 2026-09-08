package materializer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// pinMarkerFile states the expected pin at the root of a local source: the
// offline analog of the git tag match. It ships with the copied tree like
// any other source file.
const pinMarkerFile = "PIN"

// gitTimeout bounds every git invocation.
const gitTimeout = 2 * time.Minute

// FetchSource resolves the boilerplate source and stages the pinned content
// into a fresh directory under workRoot, returning its path. Local sources
// verify the pin against the PIN marker file; git sources clone
// --depth 1 --branch <pin> for named refs (and verify HEAD resolves to the
// pin), or fetch exactly one commit for full SHA pins, so a drifting
// default branch can never satisfy a pinned fetch.
func FetchSource(ctx context.Context, bp domain.Boilerplate, pin, workRoot string) (string, error) {
	switch bp.Source.Kind() {
	case "local":
		return fetchLocal(bp, pin)
	default:
		return fetchGit(ctx, bp, pin, workRoot)
	}
}

func fetchLocal(bp domain.Boilerplate, pin string) (string, error) {
	src := bp.Source.Path
	st, err := os.Stat(src)
	if err != nil || !st.IsDir() {
		return "", domain.Materialization(fmt.Sprintf("local source for %q is not a directory: %s", bp.ID, src))
	}
	raw, err := os.ReadFile(filepath.Join(src, pinMarkerFile))
	if err != nil {
		return "", domain.Materialization(fmt.Sprintf(
			"local source for %q has no %s marker: cannot verify pin %q", bp.ID, pinMarkerFile, pin))
	}
	if strings.TrimSpace(string(raw)) != pin {
		return "", domain.Materialization(fmt.Sprintf(
			"pin mismatch for %q: plan pins %q but the local source declares %q",
			bp.ID, pin, strings.TrimSpace(string(raw))))
	}
	return src, nil
}

// isSHAPin reports whether pin is a full immutable commit SHA (40 hex
// digits). The catalog pin policy (H6) freezes reviewed snapshots as SHAs;
// the git fetch path must honor them instead of treating every pin as a
// branch/tag ref (`git clone --branch <sha>` never resolves).
func isSHAPin(pin string) bool {
	if len(pin) != 40 {
		return false
	}
	for _, r := range pin {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func fetchGit(ctx context.Context, bp domain.Boilerplate, pin, workRoot string) (string, error) {
	repo := bp.EffectiveRepo()
	if strings.TrimSpace(repo) == "" {
		return "", domain.Materialization(fmt.Sprintf("boilerplate %q declares no git repository", bp.ID))
	}
	if strings.TrimSpace(pin) == "" {
		return "", domain.Materialization(fmt.Sprintf("boilerplate %q declares no pin", bp.ID))
	}
	if _, err := exec.LookPath("git"); err != nil {
		return "", domain.ExternalCommand("git is not available on PATH")
	}
	dest, err := os.MkdirTemp(workRoot, "fetch-*")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create fetch dir: %v", err))
	}
	cloneCtx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	if isSHAPin(pin) {
		// A SHA names no branch, so clone --branch cannot resolve it:
		// fetch the single commit into an empty repo and detach HEAD at
		// it. Verification below (HEAD == rev-list pin) is unchanged.
		if err := fetchSHA(cloneCtx, repo, pin, dest); err != nil {
			_ = os.RemoveAll(dest)
			return "", err
		}
	} else {
		clone := exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", "--branch", pin, repo, dest)
		clone.Env = allowlistEnv()
		if out, err := clone.CombinedOutput(); err != nil {
			_ = os.RemoveAll(dest)
			return "", domain.Materialization(fmt.Sprintf("clone %s@%s: %v\n%s", repo, pin, err, tailLines(string(out))))
		}
	}
	verifyCtx, vcancel := context.WithTimeout(ctx, gitTimeout)
	defer vcancel()
	head, err := gitOutput(verifyCtx, dest, "rev-parse", "HEAD")
	if err != nil {
		_ = os.RemoveAll(dest)
		return "", domain.Materialization(fmt.Sprintf("verify %s@%s: %v", repo, pin, err))
	}
	want, err := gitOutput(verifyCtx, dest, "rev-list", "-n", "1", pin)
	if err != nil || strings.TrimSpace(want) != strings.TrimSpace(head) {
		_ = os.RemoveAll(dest)
		return "", domain.Materialization(fmt.Sprintf(
			"pin verification failed for %s: HEAD %q does not resolve to pin %q", repo, shortOut(head), pin))
	}
	return dest, nil
}

// fetchSHA stages exactly one commit into dest (created by MkdirTemp) and
// leaves HEAD detached at it. Each step runs under the shared clone
// context so one timeout bounds the whole fetch.
func fetchSHA(ctx context.Context, repo, pin, dest string) error {
	run := func(args ...string) error {
		c := exec.CommandContext(ctx, "git", args...)
		c.Dir = dest
		c.Env = allowlistEnv()
		if out, err := c.CombinedOutput(); err != nil {
			return domain.Materialization(fmt.Sprintf("fetch %s@%s (%s): %v\n%s",
				repo, pin, strings.Join(args, " "), err, tailLines(string(out))))
		}
		return nil
	}
	if err := run("init", "-q"); err != nil {
		return err
	}
	if err := run("remote", "add", "origin", repo); err != nil {
		return err
	}
	if err := run("fetch", "--depth", "1", "origin", pin); err != nil {
		return err
	}
	return run("checkout", "--detach", "-q", pin)
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, "git", args...)
	c.Dir = dir
	c.Env = allowlistEnv()
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func shortOut(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
