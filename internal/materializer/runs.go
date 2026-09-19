package materializer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// WriteRunReport persists one run report under <baseDir>/.engineering/runs/
// and returns the file path. baseDir is joined with static path segments
// only; the file name is a timestamp plus a sanitized command token, so no
// user input reaches the filesystem as a path. The CLI calls it best-effort
// and never lets a report failure mask the original run error.
func WriteRunReport(baseDir string, report project.RunReport) (string, error) {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}
	dir := filepath.Join(baseDir, project.EngineeringDir, project.RunsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create runs dir: %v", err))
	}
	name := fmt.Sprintf("%s-%s.json",
		time.Now().UTC().Format("20060102-150405.000000000"),
		sanitizeCommand(report.Command))
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("encode run report: %v", err))
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("write run report: %v", err))
	}
	return path, nil
}

func sanitizeCommand(cmd string) string {
	out := make([]byte, 0, len(cmd))
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			out = append(out, c)
		} else {
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "run"
	}
	return string(out)
}
