package materializer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/project"
)

func TestAuditAppendAndFinalize(t *testing.T) {
	base := t.TempDir()
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	session := "demo-session"

	// Skill-side discovery event first.
	if _, err := AppendAuditEvent(base, session, project.AuditEvent{
		At: "2026-09-21T12:00:00Z", Phase: project.AuditPhaseDiscovery,
		Kind: project.AuditKindQuestion, Summary: "users?",
	}); err != nil {
		t.Fatal(err)
	}
	tail := project.NewAuditSession(session, "dev", "test-catalog", now)
	tail.AddEvent(project.AuditEvent{
		Phase: project.AuditPhaseResolve, Kind: project.AuditKindResult,
		Summary: "resolved", Status: "resolved",
		Data: map[string]string{"recipe": "GP-07"},
	}, now)
	dir, err := FinalizeAuditSession(base, tail, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{AuditJSONFile, AuditMDFile, AuditEventsFile} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	events, err := ReadAuditEvents(base, session)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 stitched events, got %d", len(events))
	}
	if events[0].Seq != 1 || events[1].Seq != 2 {
		t.Fatalf("seq not dense: %+v", events)
	}
}
