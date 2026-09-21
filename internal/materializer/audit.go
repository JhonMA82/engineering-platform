package materializer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/handoff"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// Audit file names inside .engineering/audit/<session>/.
const (
	AuditEventsFile = "events.jsonl"
	AuditJSONFile   = "audit.json"
	AuditMDFile     = "AUDIT.md"
)

// AuditSessionDir returns <baseDir>/.engineering/audit/<session>. baseDir
// joins static segments only; the session token is sanitized so no user
// input reaches the filesystem as a path.
func AuditSessionDir(baseDir, sessionID string) string {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}
	return filepath.Join(baseDir, project.EngineeringDir, project.AuditDirName, project.SanitizeSessionID(sessionID))
}

// EnsureAuditDir creates the session directory. It is the only audit
// helper that touches the disk without writing a report.
func EnsureAuditDir(baseDir, sessionID string) (string, error) {
	dir := AuditSessionDir(baseDir, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create audit dir: %v", err))
	}
	return dir, nil
}

// AppendAuditEvent appends one event to events.jsonl, creating the session
// directory on demand. The skill (agent side) and the CLI share this
// format so discovery Q&A and Go command traces stitch into one session.
// Seq and At default when empty so hand-written skill JSON stays dense.
// It never rewrites history and never fails loudly: callers treat errors
// as best-effort warnings.
func AppendAuditEvent(baseDir, sessionID string, ev project.AuditEvent) (string, error) {
	dir, err := EnsureAuditDir(baseDir, sessionID)
	if err != nil {
		return "", err
	}
	// Redact once more at the boundary in case the caller built the event
	// by hand (the skill writes JSON directly).
	if ev.Data != nil {
		red := make(map[string]string, len(ev.Data))
		for k, v := range ev.Data {
			red[k] = project.RedactValue(k, v)
		}
		ev.Data = red
	}
	if ev.At == "" {
		ev.At = time.Now().UTC().Format(time.RFC3339)
	}
	if ev.Seq == 0 {
		existing, _ := ReadAuditEvents(baseDir, sessionID)
		ev.Seq = len(existing) + 1
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("encode audit event: %v", err))
	}
	path := filepath.Join(dir, AuditEventsFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("append audit event: %v", err))
	}
	defer f.Close()
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("append audit event: %v", err))
	}
	return path, nil
}

// ReadAuditEvents loads events.jsonl best-effort: corrupt lines are
// skipped so one bad skill write never destroys the session.
func ReadAuditEvents(baseDir, sessionID string) ([]project.AuditEvent, error) {
	path := filepath.Join(AuditSessionDir(baseDir, sessionID), AuditEventsFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, domain.Filesystem(fmt.Sprintf("read audit events: %v", err))
	}
	var out []project.AuditEvent
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev project.AuditEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		out = append(out, ev)
	}
	return out, nil
}

// WriteAuditReport persists audit.json (machine) plus AUDIT.md (human) for
// a complete session. dir is created on demand. The markdown render is
// pure (handoff); this function performs the only writes.
func WriteAuditReport(baseDir string, sess project.AuditSession) (string, error) {
	dir, err := EnsureAuditDir(baseDir, sess.SessionID)
	if err != nil {
		return "", err
	}
	raw, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("encode audit session: %v", err))
	}
	if err := os.WriteFile(filepath.Join(dir, AuditJSONFile), append(raw, '\n'), 0o644); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("write audit json: %v", err))
	}
	md := handoff.RenderAuditReport(sess)
	if err := os.WriteFile(filepath.Join(dir, AuditMDFile), md, 0o644); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("write audit markdown: %v", err))
	}
	return dir, nil
}

// FinalizeAuditSession stitches existing events.jsonl entries with the
// caller-supplied tail session and writes the final report. It returns the
// session directory. now is injected so tests stay deterministic. The
// events file is rewritten canonically so Seq stays dense even when the
// skill wrote hand-made JSON without sequence numbers.
func FinalizeAuditSession(baseDir string, tail project.AuditSession, now time.Time) (string, error) {
	existing, err := ReadAuditEvents(baseDir, tail.SessionID)
	if err != nil {
		return "", err
	}
	merged := project.AuditSession{
		SchemaVersion:  project.AuditSchema,
		SessionID:      tail.SessionID,
		StartedAt:      tail.StartedAt,
		CoreVersion:    tail.CoreVersion,
		CatalogVersion: tail.CatalogVersion,
		Events:         []project.AuditEvent{},
	}
	if merged.StartedAt == "" {
		merged.StartedAt = now.UTC().Format(time.RFC3339)
	}
	// Existing skill/CLI events keep their order; tail events are
	// re-sequenced after them so Seq stays dense.
	seq := 0
	for _, ev := range existing {
		seq++
		ev.Seq = seq
		if ev.At == "" {
			ev.At = now.UTC().Format(time.RFC3339)
		}
		merged.Events = append(merged.Events, ev)
	}
	for _, ev := range tail.Events {
		seq++
		ev.Seq = seq
		if ev.At == "" {
			ev.At = now.UTC().Format(time.RFC3339)
		}
		merged.Events = append(merged.Events, ev)
	}
	dir, err := EnsureAuditDir(baseDir, tail.SessionID)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, ev := range merged.Events {
		raw, merr := json.Marshal(ev)
		if merr != nil {
			return "", domain.Filesystem(fmt.Sprintf("encode audit event: %v", merr))
		}
		sb.Write(raw)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(filepath.Join(dir, AuditEventsFile), []byte(sb.String()), 0o644); err != nil {
		return "", domain.Filesystem(fmt.Sprintf("rewrite audit events: %v", err))
	}
	return WriteAuditReport(baseDir, merged)
}
