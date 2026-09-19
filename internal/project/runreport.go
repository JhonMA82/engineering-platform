package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// RunReportSchema is the stable schema version for run reports.
const RunReportSchema = "1"

// RunsDir is the diagnostics directory inside EngineeringDir where the CLI
// persists one JSON report per failed (or doctor-dirty) run.
const RunsDir = "runs"

// RunStatus distinguishes a clean run from a failed one.
const (
	RunOK    = "ok"
	RunError = "error"
)

// IntentSummary records what was asked: the product intention that drove
// the run. It is a small human-readable projection, never the raw intent.
type IntentSummary struct {
	Name     string   `json:"name,omitempty"`
	Problem  string   `json:"problem,omitempty"`
	Surfaces []string `json:"surfaces,omitempty"`
	Recipe   string   `json:"recipe,omitempty"`
	Plan     string   `json:"plan_fingerprint,omitempty"`
	Output   string   `json:"output,omitempty"`
}

// ResultSummary records what happened: status, typed error, actionable
// hint and doctor findings when the post-check ran.
type ResultSummary struct {
	Status       string    `json:"status"`
	ErrorClass   string    `json:"error_class,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Hint         string    `json:"hint,omitempty"`
	Findings     []Finding `json:"findings,omitempty"`
	Manifest     string    `json:"manifest_fingerprint,omitempty"`
}

// RunReport is one persisted execution trace: intention vs result. It is
// the ONLY diagnostics document allowed to carry a wall-clock timestamp;
// decisions and plans stay deterministic and timestamp-free.
type RunReport struct {
	SchemaVersion string        `json:"schema_version"`
	At            string        `json:"at"`
	Command       string        `json:"command"`
	Intent        IntentSummary `json:"intent"`
	Result        ResultSummary `json:"result"`
}

// SummarizeIntent derives a report-ready projection from raw intent bytes.
// Unparseable input yields a best-effort summary instead of an error so a
// broken intent file still leaves a trace naming the output it targeted.
func SummarizeIntent(intentRaw []byte, output string) IntentSummary {
	sum := IntentSummary{Output: output}
	if len(intentRaw) == 0 {
		return sum
	}
	var intent domain.ProjectIntent
	if err := json.Unmarshal(intentRaw, &intent); err != nil {
		return sum
	}
	sum.Name = intent.Name
	sum.Problem = intent.Problem
	for _, s := range intent.Surfaces {
		scope := string(s.EffectiveScope())
		sum.Surfaces = append(sum.Surfaces, fmt.Sprintf("%s:%s", strings.ToLower(string(s.Kind)), scope))
	}
	return sum
}

// SummarizePlan completes the intent projection once resolve/plan ran.
func (s IntentSummary) SummarizePlan(plan *planner.MaterializationPlan) IntentSummary {
	if plan == nil {
		return s
	}
	s.Recipe = plan.Recipe
	s.Plan = plan.Fingerprint
	if s.Name == "" {
		s.Name = plan.Project
	}
	return s
}

// ErrorReport builds a failure trace for command with the given intent and
// Go error. Findings and manifest fingerprint are attached when the run
// reached the doctor or materialize stage.
func ErrorReport(command string, intent IntentSummary, runErr error, findings []Finding, manifestFP string) RunReport {
	class, msg := ErrorClassOf(runErr)
	if len(findings) > 0 && class == "" {
		class = "doctor"
		msg = fmt.Sprintf("%d finding(s), %d error(s)", len(findings), countErrors(findings))
	}
	return RunReport{
		SchemaVersion: RunReportSchema,
		At:            time.Now().UTC().Format(time.RFC3339),
		Command:       command,
		Intent:        intent,
		Result: ResultSummary{
			Status:       RunError,
			ErrorClass:   class,
			ErrorMessage: msg,
			Hint:         HintFor(class, msg, findings),
			Findings:     findings,
			Manifest:     manifestFP,
		},
	}
}

// DoctorReport builds a trace for a doctor run that completed: status is
// ok when no error-severity findings remain.
func DoctorReport(intent IntentSummary, findings []Finding) RunReport {
	status := RunOK
	class, msg := "", ""
	if HasErrors(findings) {
		status = RunError
		class = "doctor"
		msg = fmt.Sprintf("%d finding(s), %d error(s)", len(findings), countErrors(findings))
	}
	var hint string
	if status == RunError {
		hint = HintFor(class, msg, findings)
	}
	return RunReport{
		SchemaVersion: RunReportSchema,
		At:            time.Now().UTC().Format(time.RFC3339),
		Command:       "doctor",
		Intent:        intent,
		Result: ResultSummary{
			Status:       status,
			ErrorClass:   class,
			ErrorMessage: msg,
			Hint:         hint,
			Findings:     findings,
		},
	}
}

// ErrorClassOf extracts the typed domain class from a Go error. Unknown
// errors report an empty class so callers fall back to the raw message.
func ErrorClassOf(err error) (string, string) {
	if err == nil {
		return "", ""
	}
	var derr *domain.Error
	if errors.As(err, &derr) {
		return string(derr.Class), derr.Message
	}
	// domain.Error travels as *Error; errors.As unwraps it. Anything else
	// falls back to the "class: message" shape or the raw message.
	msg := err.Error()
	if i := strings.Index(msg, ": "); i > 0 {
		class := msg[:i]
		switch domain.Class(class) {
		case domain.ClassValidation, domain.ClassCatalog, domain.ClassResolution,
			domain.ClassComposition, domain.ClassMaterialization,
			domain.ClassFilesystem, domain.ClassExternalCommand:
			return class, strings.TrimSpace(msg[i+2:])
		}
	}
	return "", msg
}

func countErrors(findings []Finding) int {
	n := 0
	for _, f := range findings {
		if f.Severity == SeverityError {
			n++
		}
	}
	return n
}

// HintFor maps a failure to the single most actionable fix. It stays
// heuristic on purpose: the hint names where to look, the report carries
// the evidence.
func HintFor(class, msg string, findings []Finding) string {
	if len(findings) > 0 {
		if h := hintForFinding(findings[0].Code); h != "" {
			return h
		}
	}
	switch class {
	case "validation":
		return "fix intent.json (schema_version, name, surfaces) and re-run; hint: eng resolve --input intent.json"
	case "catalog":
		return "run eng catalog validate to name the drifted entry, then pin or overlay --catalog-dir"
	case "resolution":
		return "no recipe matches the intent: relax surfaces/constraints or add a recipe covering them"
	case "composition":
		return "resolve picked a recipe the composer cannot lay out: check surface destinations for collisions"
	case "materialization":
		if strings.Contains(msg, "collision") {
			return "two components target the same destination: change the plan destinations and re-run into an empty dir"
		}
		if strings.Contains(msg, "drift") || strings.Contains(msg, "pin") {
			return "plan pins differ from the catalog: re-run eng plan to refresh the fingerprint"
		}
		return "materialization failed before commit: fix the named cause and re-run into a new or empty --output"
	case "filesystem":
		return "check paths and permissions (output new or empty, intent readable) and re-run"
	case "external-command":
		return "a curated generator/setup command failed: re-run with the fixture locally and check its output"
	case "doctor":
		return "run eng doctor --project <dir> --json for the full finding list; fix the first error-severity code"
	default:
		return "re-run with --json for the machine-readable trace and fix the first named cause"
	}
}

func hintForFinding(code string) string {
	switch code {
	case "manifest-unreadable":
		return "restore .engineering/project.json from the materialization-plan.json copy or re-materialize"
	case "destination-missing", "file-missing":
		return "a tracked path was deleted by hand: restore it or re-materialize into a clean dir"
	case "project-map-drift":
		return "surfaces moved without eng surface add: move them back or extend via the evolution commands"
	case "provenance-mismatch", "provenance-component-mismatch", "generation-config-drift":
		return "provenance no longer matches the manifest: the project was hand-edited after materialization"
	case "materialization-plan-missing", "materialization-plan-drift":
		return "the embedded plan copy is missing or stale: re-materialize to restore a consistent fingerprint"
	default:
		return ""
	}
}
