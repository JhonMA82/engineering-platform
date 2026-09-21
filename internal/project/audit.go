// Package project audit session types.
//
// The dev-only audit trail records the full newproject journey
// (idea → discovery Q&A → intent versions → resolve/plan/materialize/doctor)
// as append-only events. Builders here are pure: only the materializer
// performs filesystem writes. Timestamps live only in audit documents;
// decisions and plans stay deterministic and timestamp-free.
package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// AuditSchema is the stable schema version for audit sessions.
const AuditSchema = "1"

// AuditEnv gates every audit write. Audit files are created only when
// ENG_AUDIT=1 (or "true"); without it --audit warns and writes nothing.
// This keeps the capability strictly dev-only and off by default.
const AuditEnv = "ENG_AUDIT"

// AuditDirName is the diagnostics directory inside EngineeringDir where
// audit sessions live: .engineering/audit/<session>/.
const AuditDirName = "audit"

// Audit event phases covering the whole pipeline.
const (
	AuditPhaseIdea      = "idea"
	AuditPhaseDiscovery = "discovery"
	AuditPhaseIntent    = "intent"
	AuditPhaseResolve   = "resolve"
	AuditPhasePlan      = "plan"
	AuditPhaseMaterial  = "materialize"
	AuditPhaseDoctor    = "doctor"
	AuditPhaseDone      = "done"
)

// Audit event kinds.
const (
	AuditKindIdea     = "idea"
	AuditKindQuestion = "question"
	AuditKindAnswer   = "answer"
	AuditKindIntent   = "intent"
	AuditKindCommand  = "command"
	AuditKindArtifact = "artifact"
	AuditKindResult   = "result"
	AuditKindNote     = "note"
)

// AuditEvent is one append-only trace entry. At is RFC3339 UTC and is the
// only wall-clock field allowed outside provenance/run-reports. Argv
// records the exact CLI invocation; Summary is human-readable; Data
// carries fingerprints, recipe, status and other machine facts. Secrets
// must never reach Data: callers pass values through RedactValue.
type AuditEvent struct {
	Seq     int               `json:"seq"`
	At      string            `json:"at"`
	Phase   string            `json:"phase"`
	Kind    string            `json:"kind"`
	Summary string            `json:"summary,omitempty"`
	Command string            `json:"command,omitempty"`
	Argv    []string          `json:"argv,omitempty"`
	Status  string            `json:"status,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}

// AuditSession is one stitched journey from idea to materialization.
// Events are append-only; the session file is rewritten on finalize.
type AuditSession struct {
	SchemaVersion  string       `json:"schema_version"`
	SessionID      string       `json:"session_id"`
	StartedAt      string       `json:"started_at"`
	CoreVersion    string       `json:"core_version,omitempty"`
	CatalogVersion string       `json:"catalog_version,omitempty"`
	Events         []AuditEvent `json:"events"`
}

// AuditEnabled reports whether the dev-only gate is open.
func AuditEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(AuditEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// NewAuditSession starts a session with an explicit start time so tests
// stay deterministic; production passes time.Now().UTC().
func NewAuditSession(sessionID, coreVersion, catalogVersion string, now time.Time) AuditSession {
	if strings.TrimSpace(sessionID) == "" {
		sessionID = now.UTC().Format("20060102-150405")
	}
	return AuditSession{
		SchemaVersion:  AuditSchema,
		SessionID:      sessionID,
		StartedAt:      now.UTC().Format(time.RFC3339),
		CoreVersion:    coreVersion,
		CatalogVersion: catalogVersion,
		Events:         []AuditEvent{},
	}
}

// AddEvent appends one event, stamping Seq and At when empty. Values in
// data are redacted before storage so a stray token never persists.
func (s *AuditSession) AddEvent(ev AuditEvent, now time.Time) AuditEvent {
	if ev.At == "" {
		ev.At = now.UTC().Format(time.RFC3339)
	}
	ev.Seq = len(s.Events) + 1
	if ev.Data != nil {
		red := make(map[string]string, len(ev.Data))
		for k, v := range ev.Data {
			red[k] = RedactValue(k, v)
		}
		ev.Data = red
	}
	if ev.Argv != nil {
		cp := make([]string, len(ev.Argv))
		copy(cp, ev.Argv)
		ev.Argv = cp
	}
	s.Events = append(s.Events, ev)
	return ev
}

var secretKeyPattern = regexp.MustCompile(`(?i)(token|secret|password|passwd|api[_-]?key|credentials|authorization|cookie|session)`)

// RedactValue replaces secret-looking values with [redacted]. A key whose
// name matches the secret pattern is always redacted; values that look
// like bearer tokens or long opaque secrets are redacted too. File paths,
// argv strings and fingerprints are never redacted.
func RedactValue(key, value string) string {
	if value == "" {
		return ""
	}
	if secretKeyPattern.MatchString(key) {
		return "[redacted]"
	}
	v := strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(v), "bearer ") {
		return "[redacted]"
	}
	// Paths and multi-word text are never secrets.
	if strings.Contains(v, "/") || strings.Contains(v, "\\") || strings.Contains(v, " ") || strings.Contains(v, "\n") {
		return value
	}
	// Long opaque single-token strings are almost certainly secrets.
	if len(v) >= 32 && !strings.Contains(v, " ") && !strings.Contains(v, "\n") {
		// Fingerprints and pins are hex/short dotted versions; keep them.
		if isFingerprintLike(v) {
			return value
		}
		return "[redacted]"
	}
	return value
}

func isFingerprintLike(v string) bool {
	// Plan/intent fingerprints are short hex (typically <= 64 chars with
	// hex alphabet) and pins look like versions or short shas.
	if len(v) > 64 {
		return false
	}
	for _, c := range v {
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' || c == '.' || c == '-' || c == '_' || c == 'v' {
			continue
		}
		return false
	}
	return true
}

// SanitizeSessionID keeps audit directory names safe: only
// alphanumerics, dash and underscore survive; anything else becomes _.
func SanitizeSessionID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "session"
	}
	var sb strings.Builder
	for _, r := range id {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	out := sb.String()
	if out == "" {
		return "session"
	}
	return out
}

// AuditSummaryLine renders one event for logs and tests.
func AuditSummaryLine(ev AuditEvent) string {
	return fmt.Sprintf("#%d [%s/%s] %s", ev.Seq, ev.Phase, ev.Kind, ev.Summary)
}
