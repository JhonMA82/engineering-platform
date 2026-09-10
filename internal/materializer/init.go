package materializer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	opencodebootstrap "github.com/jhonma82/engineering-platform/integrations/opencode"
	pibootstrap "github.com/jhonma82/engineering-platform/integrations/pi"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// BootstrapVersion versions the eng-owned bootstrap resources installed by
// InitWorkspace. Bump it when the embedded PI files change so a second
// `eng init` can detect outdated integrations and repair them.
const BootstrapVersion = "1.0.2"

// managedFiles maps eng-owned project-local paths (slash-separated,
// relative to the workspace root) to their embedded content for one agent.
// InitWorkspace and CleanupBootstrap manage exactly these paths and never
// anything else inside the agent directories.
func managedFiles(agent string) [][2]string {
	switch agent {
	case "pi":
		return [][2]string{
			{".pi/prompts/newproject.md", pibootstrap.NewprojectMD},
			{".pi/skills/project-discovery/SKILL.md", pibootstrap.DiscoverySkillMD},
		}
	case "opencode":
		return [][2]string{
			{".opencode/commands/newproject.md", opencodebootstrap.NewprojectMD},
			{".opencode/skills/engineering-project-discovery/SKILL.md", opencodebootstrap.DiscoverySkillMD},
		}
	default:
		return nil
	}
}

// InitOptions configures InitWorkspace.
type InitOptions struct {
	// Agent is "pi" or "opencode" (install local agent integration) or
	// "none" (state only).
	Agent string
	// Force repairs eng-owned files even when they were customized.
	Force bool
	// EngVersion stamps bootstrap.json; callers pass the binary release line.
	EngVersion string
}

// InitReport describes what InitWorkspace did.
type InitReport struct {
	Workspace        string
	Agent            string
	BootstrapVersion string
	EngVersion       string
	AlreadyUpToDate  bool
	Created          []string
	Updated          []string
	Kept             []string
}

// initBootstrapState is the JSON document stored in .engineering/bootstrap.json.
type initBootstrapState struct {
	Agent            string `json:"agent"`
	BootstrapVersion string `json:"bootstrapVersion"`
	EngVersion       string `json:"engVersion"`
	InitializedAt    string `json:"initializedAt"`
}

// InitWorkspace initializes dir as an Engineering Platform agent bootstrap
// workspace: it writes .engineering/bootstrap.json and, for agent "pi",
// installs the embedded PI integration into project-local .pi/ paths.
//
// It is idempotent: a second run with the same agent and current resources
// reports AlreadyUpToDate and writes nothing. Files outside the managed set
// are never touched; managed files customized by the user are preserved
// unless Force is set. All writes are atomic (temp file + rename).
func InitWorkspace(dir string, opts InitOptions) (*InitReport, error) {
	if opts.Agent != "pi" && opts.Agent != "opencode" && opts.Agent != "none" {
		return nil, domain.Validation(fmt.Sprintf("unknown agent %q (want pi, opencode or none)", opts.Agent))
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, domain.Filesystem("workspace: " + err.Error())
	}
	if !info.IsDir() {
		return nil, domain.Filesystem("workspace is not a directory: " + dir)
	}

	report := &InitReport{
		Workspace:        dir,
		Agent:            opts.Agent,
		BootstrapVersion: BootstrapVersion,
		EngVersion:       opts.EngVersion,
	}

	bootstrapPath := filepath.Join(dir, ".engineering", "bootstrap.json")
	existing, err := readBootstrapState(bootstrapPath)
	if err != nil {
		return nil, err
	}

	if existing != nil && !opts.Force &&
		existing.Agent == opts.Agent &&
		existing.BootstrapVersion == BootstrapVersion &&
		managedFilesCurrent(dir, opts.Agent) {
		report.AlreadyUpToDate = true
		return report, nil
	}

	if err := os.MkdirAll(filepath.Join(dir, ".engineering"), 0o755); err != nil {
		return nil, domain.Filesystem("create .engineering: " + err.Error())
	}
	state := initBootstrapState{
		Agent:            opts.Agent,
		BootstrapVersion: BootstrapVersion,
		EngVersion:       opts.EngVersion,
		InitializedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, domain.Filesystem("marshal bootstrap state: " + err.Error())
	}
	raw = append(raw, '\n')
	if err := writeFileAtomic(dir, filepath.FromSlash(".engineering/bootstrap.json"), string(raw)); err != nil {
		return nil, err
	}
	if existing == nil {
		report.Created = append(report.Created, ".engineering/bootstrap.json")
	} else {
		report.Updated = append(report.Updated, ".engineering/bootstrap.json")
	}

	if opts.Agent == "pi" || opts.Agent == "opencode" {
		for _, mf := range managedFiles(opts.Agent) {
			rel := filepath.FromSlash(mf[0])
			current, readErr := os.ReadFile(filepath.Join(dir, rel))
			switch {
			case readErr != nil && os.IsNotExist(readErr):
				if err := writeFileAtomic(dir, rel, mf[1]); err != nil {
					return nil, err
				}
				report.Created = append(report.Created, mf[0])
			case readErr != nil:
				return nil, domain.Filesystem("read " + mf[0] + ": " + readErr.Error())
			case string(current) == mf[1]:
				// Already current, touch nothing.
			case opts.Force:
				if err := writeFileAtomic(dir, rel, mf[1]); err != nil {
					return nil, err
				}
				report.Updated = append(report.Updated, mf[0])
			default:
				report.Kept = append(report.Kept, mf[0])
			}
		}
	}

	return report, nil
}

// managedFilesCurrent reports whether the managed agent files match the
// embedded resources.
func managedFilesCurrent(dir, agent string) bool {
	for _, mf := range managedFiles(agent) {
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(mf[0])))
		if err != nil || string(raw) != mf[1] {
			return false
		}
	}
	return true
}

// readBootstrapState returns nil when bootstrap.json does not exist.
func readBootstrapState(path string) (*initBootstrapState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, domain.Filesystem("read bootstrap state: " + err.Error())
	}
	var state initBootstrapState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, domain.Filesystem("parse bootstrap state: " + err.Error())
	}
	return &state, nil
}

// writeFileAtomic writes content to rel (relative to dir) via a temp file
// in the same directory followed by a rename.
func writeFileAtomic(dir, rel, content string) error {
	target := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return domain.Filesystem("create parent dir: " + err.Error())
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".eng-tmp-*")
	if err != nil {
		return domain.Filesystem("create temp file: " + err.Error())
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return domain.Filesystem("write temp file: " + err.Error())
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return domain.Filesystem("close temp file: " + err.Error())
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		os.Remove(tmpName)
		return domain.Filesystem("chmod temp file: " + err.Error())
	}
	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return domain.Filesystem("publish file: " + err.Error())
	}
	return nil
}
