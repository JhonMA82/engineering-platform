package handoff

import (
	"strings"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/project"
)

func TestRenderAuditReportCoversPhases(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	s := project.NewAuditSession("demo-1", "dev", "v1", now)
	s.AddEvent(project.AuditEvent{Phase: project.AuditPhaseIdea, Kind: project.AuditKindIdea, Summary: "quiero una tienda"}, now)
	s.AddEvent(project.AuditEvent{Phase: project.AuditPhaseDiscovery, Kind: project.AuditKindQuestion, Summary: "users?"}, now)
	s.AddEvent(project.AuditEvent{Phase: project.AuditPhaseResolve, Kind: project.AuditKindResult, Summary: "resolved GP-07", Status: "resolved"}, now)
	md := string(RenderAuditReport(s))
	for _, want := range []string{"demo-1", "Idea", "Discovery", "Resolve", "Timeline", "How to reproduce"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}

func TestRenderAuditEmpty(t *testing.T) {
	now := time.Now().UTC()
	s := project.NewAuditSession("empty", "", "", now)
	if got := string(RenderAuditReport(s)); !strings.Contains(got, "No events") {
		t.Fatalf("empty session must note no events: %s", got)
	}
}
