// Package handoff audit rendering.
//
// RenderAuditReport is pure: it turns an AuditSession into deterministic
// AUDIT.md bytes. The materializer performs every write.
package handoff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/project"
)

// RenderAuditReport returns the human-readable dev-only audit document.
// It never redacts further: project.AddEvent already redacted secrets.
func RenderAuditReport(sess project.AuditSession) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Audit — session %s\n\n", sess.SessionID))
	sb.WriteString(fmt.Sprintf("Schema %s · started %s", sess.SchemaVersion, sess.StartedAt))
	if sess.CoreVersion != "" || sess.CatalogVersion != "" {
		sb.WriteString(fmt.Sprintf(" · core %s · catalog %s", sess.CoreVersion, sess.CatalogVersion))
	}
	sb.WriteString("\n\n")
	sb.WriteString("> Dev-only trace: idea → discovery Q&A → intent → resolve → plan → materialize → doctor.\n")
	sb.WriteString("> Generated with ENG_AUDIT=1 and --audit; never written without the gate.\n\n")

	if len(sess.Events) == 0 {
		sb.WriteString("No events recorded.\n")
		return []byte(sb.String())
	}

	sb.WriteString("## Timeline\n\n")
	sb.WriteString("| # | at | phase | kind | summary |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, ev := range sess.Events {
		summary := strings.ReplaceAll(ev.Summary, "|", "\\|")
		summary = strings.ReplaceAll(summary, "\n", " ")
		if len(summary) > 160 {
			summary = summary[:157] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			ev.Seq, ev.At, ev.Phase, ev.Kind, summary))
	}
	sb.WriteString("\n")

	byPhase := map[string][]project.AuditEvent{}
	for _, ev := range sess.Events {
		byPhase[ev.Phase] = append(byPhase[ev.Phase], ev)
	}
	order := []string{
		project.AuditPhaseIdea,
		project.AuditPhaseDiscovery,
		project.AuditPhaseIntent,
		project.AuditPhaseResolve,
		project.AuditPhasePlan,
		project.AuditPhaseMaterial,
		project.AuditPhaseDoctor,
		project.AuditPhaseDone,
	}
	seen := map[string]bool{}
	for _, ph := range order {
		evs, ok := byPhase[ph]
		if !ok {
			continue
		}
		seen[ph] = true
		sb.WriteString(fmt.Sprintf("## %s\n\n", titlePhase(ph)))
		for _, ev := range evs {
			sb.WriteString(fmt.Sprintf("### #%d %s (%s)\n\n", ev.Seq, ev.Kind, ev.At))
			if ev.Summary != "" {
				sb.WriteString(ev.Summary + "\n\n")
			}
			if ev.Command != "" {
				sb.WriteString(fmt.Sprintf("command: `%s`\n\n", ev.Command))
			}
			if len(ev.Argv) > 0 {
				sb.WriteString("argv:\n\n```text\n" + strings.Join(ev.Argv, " ") + "\n```\n\n")
			}
			if ev.Status != "" {
				sb.WriteString(fmt.Sprintf("status: `%s`\n\n", ev.Status))
			}
			if len(ev.Data) > 0 {
				keys := make([]string, 0, len(ev.Data))
				for k := range ev.Data {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				sb.WriteString("facts:\n\n")
				for _, k := range keys {
					v := ev.Data[k]
					if strings.Contains(v, "\n") {
						sb.WriteString(fmt.Sprintf("- %s:\n\n  ```text\n  %s\n  ```\n", k, indentBlock(v)))
					} else {
						sb.WriteString(fmt.Sprintf("- %s: `%s`\n", k, v))
					}
				}
				sb.WriteString("\n")
			}
		}
	}
	// Any custom phases the skill introduced render last, sorted.
	var rest []string
	for ph := range byPhase {
		if !seen[ph] {
			rest = append(rest, ph)
		}
	}
	sort.Strings(rest)
	for _, ph := range rest {
		sb.WriteString(fmt.Sprintf("## %s\n\n", titlePhase(ph)))
		for _, ev := range byPhase[ph] {
			sb.WriteString(fmt.Sprintf("### #%d %s (%s)\n\n%s\n\n", ev.Seq, ev.Kind, ev.At, ev.Summary))
		}
	}

	sb.WriteString("## How to reproduce\n\n")
	sb.WriteString("Re-run the recorded `argv` lines with the same intent/decision/plan files.\n")
	sb.WriteString("Audit never affects fingerprints: decisions and plans stay deterministic.\n")
	return []byte(sb.String())
}

func titlePhase(ph string) string {
	switch ph {
	case project.AuditPhaseIdea:
		return "Idea"
	case project.AuditPhaseDiscovery:
		return "Discovery Q&A"
	case project.AuditPhaseIntent:
		return "Intent versions"
	case project.AuditPhaseResolve:
		return "Resolve"
	case project.AuditPhasePlan:
		return "Plan"
	case project.AuditPhaseMaterial:
		return "Materialize"
	case project.AuditPhaseDoctor:
		return "Doctor"
	case project.AuditPhaseDone:
		return "Done"
	default:
		if ph == "" {
			return "Notes"
		}
		return ph
	}
}

func indentBlock(v string) string {
	lines := strings.Split(v, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}
