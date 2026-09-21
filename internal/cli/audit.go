package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/project"
	"github.com/jhonma82/engineering-platform/internal/version"
)

// auditOptions carries the dev-only audit flags shared by every command.
type auditOptions struct {
	enabled   bool
	dir       string
	sessionID string
}

func addAuditFlags(fs *flag.FlagSet) *auditOptions {
	o := &auditOptions{}
	fs.BoolVar(&o.enabled, "audit", false, "write dev-only audit trace (requires ENG_AUDIT=1)")
	fs.StringVar(&o.dir, "audit-dir", "", "audit base directory (default: workspace .engineering/audit)")
	fs.StringVar(&o.sessionID, "audit-session", "", "audit session id (default: timestamp)")
	return o
}

// auditSession resolves the session id and base directory, enforcing the
// ENG_AUDIT=1 dev-only gate. ok=false means auditing is off for this run
// (flag absent or gate closed); the caller continues without writes.
func auditSession(o *auditOptions, fallbackOutput string, now time.Time) (baseDir, sessionID string, ok bool) {
	if o == nil || !o.enabled {
		return "", "", false
	}
	if !project.AuditEnabled() {
		fmt.Fprintln(os.Stderr, "audit: disabled — set ENG_AUDIT=1 to enable dev-only audit writes")
		return "", "", false
	}
	base := o.dir
	if base == "" {
		base = reportBaseDir(fallbackOutput)
	}
	id := o.sessionID
	if id == "" {
		id = now.UTC().Format("20060102-150405")
	}
	return base, project.SanitizeSessionID(id), true
}

// auditBase resolves only the base directory for skill-stitched sessions.
func auditBase(o *auditOptions, fallbackOutput string) string {
	if o != nil && o.dir != "" {
		return o.dir
	}
	return reportBaseDir(fallbackOutput)
}

// auditRecord appends one event best-effort: failures warn only and never
// mask the run result.
func auditRecord(baseDir, sessionID string, ev project.AuditEvent) {
	if baseDir == "" || sessionID == "" {
		return
	}
	if !project.AuditEnabled() {
		return
	}
	if ev.At == "" {
		ev.At = time.Now().UTC().Format(time.RFC3339)
	}
	if _, err := materializer.AppendAuditEvent(baseDir, sessionID, ev); err != nil {
		fmt.Fprintf(os.Stderr, "audit: could not append event: %v\n", err)
	}
}

// auditFinalize stitches events.jsonl with the tail session and writes
// audit.json + AUDIT.md best-effort. It warns only, never fails the run.
func auditFinalize(baseDir string, tail project.AuditSession) {
	if baseDir == "" || tail.SessionID == "" {
		return
	}
	if !project.AuditEnabled() {
		return
	}
	now := time.Now().UTC()
	if tail.StartedAt == "" {
		tail.StartedAt = now.Format(time.RFC3339)
	}
	if tail.CoreVersion == "" {
		tail.CoreVersion = version.CoreVersion
	}
	if tail.CatalogVersion == "" {
		if cat, err := catalog.Load(""); err == nil {
			tail.CatalogVersion = cat.CatalogVersion
		}
	}
	dir, err := materializer.FinalizeAuditSession(baseDir, tail, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit: could not finalize report: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "audit: %s/%s and %s\n", dir, materializer.AuditMDFile, materializer.AuditJSONFile)
}
