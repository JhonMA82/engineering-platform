package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// reportBaseDir picks where a run report lands: the first candidate that is
// an existing directory (project/output dir when available), otherwise the
// workspace the command ran in. It never creates the output dir itself, so
// a failed materialization that never committed leaves no stray project.
func reportBaseDir(candidates ...string) string {
	for _, d := range candidates {
		if d == "" {
			continue
		}
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d
		}
	}
	return "."
}

// persistReport writes rep best-effort and names the file plus the hint on
// stderr. A report failure warns only; it never masks the run error.
func persistReport(baseDir string, rep project.RunReport) {
	path, err := materializer.WriteRunReport(baseDir, rep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "report: could not persist run report: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "report: %s\n", path)
	if rep.Result.Hint != "" {
		fmt.Fprintf(os.Stderr, "hint: %s\n", rep.Result.Hint)
	}
}
