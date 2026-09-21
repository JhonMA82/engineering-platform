package project

import (
	"testing"
	"time"
)

func TestRedactSecrets(t *testing.T) {
	if got := RedactValue("api_token", "abc123"); got != "[redacted]" {
		t.Fatalf("token key not redacted: %q", got)
	}
	if got := RedactValue("recipe", "GP-01"); got != "GP-01" {
		t.Fatalf("recipe must survive: %q", got)
	}
	if got := RedactValue("plan_fingerprint", "abc123def456"); got != "abc123def456" {
		t.Fatalf("fingerprint must survive: %q", got)
	}
	if got := RedactValue("note", "bearer secret-value-here"); got != "[redacted]" {
		t.Fatalf("bearer must redact: %q", got)
	}
	if got := RedactValue("input", "/tmp/opencode/audit-smoke/intent.json"); got != "/tmp/opencode/audit-smoke/intent.json" {
		t.Fatalf("paths must survive: %q", got)
	}
}

func TestAuditSessionSequencing(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	s := NewAuditSession("demo", "dev", "v1", now)
	s.AddEvent(AuditEvent{Phase: AuditPhaseIdea, Kind: AuditKindIdea, Summary: "idea text", Data: map[string]string{"api_token": "x"}}, now)
	if len(s.Events) != 1 || s.Events[0].Seq != 1 {
		t.Fatalf("seq not assigned: %+v", s.Events)
	}
	if s.Events[0].Data["api_token"] != "[redacted]" {
		t.Fatalf("secret leaked: %+v", s.Events[0].Data)
	}
}

func TestSanitizeSessionID(t *testing.T) {
	if got := SanitizeSessionID("2026/09/21 12:00"); got != "2026_09_21_12_00" {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeSessionID(""); got != "session" {
		t.Fatalf("empty must default: %q", got)
	}
}
