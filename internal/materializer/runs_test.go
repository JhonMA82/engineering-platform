package materializer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/project"
)

func TestWriteRunReportRoundTrip(t *testing.T) {
	base := t.TempDir()
	rep := project.ErrorReport("doctor", project.IntentSummary{Output: base}, nil,
		[]project.Finding{{Severity: project.SeverityError, Code: "file-missing", Message: "gone"}}, "")
	path, err := WriteRunReport(base, rep)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var back project.RunReport
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.Result.Status != project.RunError || back.Result.Hint == "" {
		t.Fatalf("report not persisted with hint: %+v", back.Result)
	}
	if got := filepath.Dir(path); filepath.Base(got) != project.RunsDir {
		t.Fatalf("report outside runs dir: %s", path)
	}
}
