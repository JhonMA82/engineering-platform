package materializer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// CleanupReport describes what CleanupBootstrap did.
type CleanupReport struct {
	Workspace string
	Removed   []string
	// Cleaned is false when dir was never prepared by eng init: cleanup is
	// then a no-op and nothing is touched.
	Cleaned bool
}

// CleanupBootstrap removes exactly the eng-owned bootstrap resources of
// the workspace's agent (bootstrap RFC sections 11-13): the managed
// integration files plus .engineering/bootstrap.json.
//
// Empty parent directories below the agent directory are pruned only when
// left empty; user content is never deleted (pruning stops at the first
// non-empty directory). Product code, generated agent tooling and the
// minimal .engineering/ provenance are kept. Cleanup runs only after
// successful materialization and validation; on failure paths the bootstrap
// state is retained for diagnosis and retry.
func CleanupBootstrap(dir string) (*CleanupReport, error) {
	report := &CleanupReport{Workspace: dir}
	state, err := readBootstrapState(filepath.Join(dir, ".engineering", "bootstrap.json"))
	if err != nil {
		return nil, err
	}
	if state == nil {
		return report, nil
	}
	if state.Agent != "pi" && state.Agent != "opencode" && state.Agent != "none" {
		return nil, domain.Validation(fmt.Sprintf("unknown agent %q in bootstrap state", state.Agent))
	}
	targets := append(managedPaths(state.Agent), ".engineering/bootstrap.json")
	for _, target := range targets {
		p := filepath.Join(dir, filepath.FromSlash(target))
		if err := os.Remove(p); err == nil {
			report.Removed = append(report.Removed, target)
		} else if !os.IsNotExist(err) {
			return nil, domain.Filesystem(fmt.Sprintf("remove %s: %v", target, err))
		}
	}
	for _, parent := range pruneParents(state.Agent) {
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(parent))); err != nil {
			break
		}
	}
	report.Cleaned = true
	return report, nil
}

// managedPaths returns the slash-separated eng-owned integration paths for
// one agent plus nothing else.
func managedPaths(agent string) []string {
	out := make([]string, 0, 2)
	for _, mf := range managedFiles(agent) {
		out = append(out, mf[0])
	}
	return out
}

// pruneParents returns the candidate empty-parent chain for one agent,
// innermost first. Cleanup removes them while empty and stops at the
// first one that is not.
func pruneParents(agent string) []string {
	switch agent {
	case "opencode":
		return []string{
			".opencode/commands",
			".opencode/skills/engineering-project-discovery",
			".opencode/skills",
			".opencode",
		}
	default:
		return []string{
			".pi/prompts",
			".pi/skills/project-discovery",
			".pi/skills",
			".pi",
		}
	}
}
