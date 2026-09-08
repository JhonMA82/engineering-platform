package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Finding severities.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// planCopyFile is the persisted materialization plan inside EngineeringDir.
const planCopyFile = "materialization-plan.json"

// Finding is one typed doctor observation.
type Finding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// HasErrors reports whether any finding has error severity.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

func errorFinding(code, format string, args ...any) Finding {
	return Finding{Severity: SeverityError, Code: code, Message: fmt.Sprintf(format, args...)}
}

func warnFinding(code, format string, args ...any) Finding {
	return Finding{Severity: SeverityWarning, Code: code, Message: fmt.Sprintf(format, args...)}
}

// Doctor checks project consistency: manifest↔filesystem, project-map↔dirs
// and provenance↔manifest fingerprint. Content problems come back as
// error-severity findings with a nil Go error; only an unreadable project
// directory itself returns a Go error. A non-empty error-severity set must
// drive a non-zero CLI exit.
func Doctor(projectDir string) ([]Finding, error) {
	st, err := os.Stat(projectDir)
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("doctor: cannot stat project dir: %v", err))
	}
	if !st.IsDir() {
		return nil, domain.Filesystem(fmt.Sprintf("doctor: not a directory: %s", projectDir))
	}
	var findings []Finding
	manifest, err := ReadManifest(projectDir)
	if err != nil {
		return []Finding{errorFinding("manifest-unreadable", "manifest unreadable: %v", err)}, nil
	}
	for _, c := range manifest.Components {
		dest := filepath.Join(projectDir, filepath.FromSlash(c.Destination))
		dst, err := os.Stat(dest)
		if err != nil || !dst.IsDir() {
			findings = append(findings, errorFinding("destination-missing",
				"destination %q (surface %s, provider %s) is missing", c.Destination, c.Surface, c.Boilerplate))
		}
	}
	for _, f := range manifest.Files {
		if _, err := os.Lstat(filepath.Join(projectDir, filepath.FromSlash(f))); err != nil {
			findings = append(findings, errorFinding("file-missing",
				"manifest-tracked file %q is missing", f))
		}
	}
	pmap, err := ReadProjectMap(projectDir)
	if err != nil {
		findings = append(findings, errorFinding("project-map-unreadable", "project-map unreadable: %v", err))
	} else {
		want := map[string]string{}
		for _, c := range manifest.Components {
			want[c.Surface] = c.Destination
		}
		got := map[string]string{}
		for surface, entry := range pmap.Surfaces {
			got[surface] = entry.Path
		}
		if !equalStringMap(want, got) {
			findings = append(findings, errorFinding("project-map-drift",
				"project-map surfaces %s do not match manifest destinations %s", describeMap(got), describeMap(want)))
		}
	}
	prov, err := ReadProvenance(projectDir)
	if err != nil {
		findings = append(findings, errorFinding("provenance-unreadable", "provenance unreadable: %v", err))
	} else {
		if prov.PlanFingerprint != manifest.PlanFingerprint {
			findings = append(findings, errorFinding("provenance-mismatch",
				"provenance plan_fingerprint %q does not match manifest %q", shortHash(prov.PlanFingerprint), shortHash(manifest.PlanFingerprint)))
		}
		findings = append(findings, checkEvolutionEvents(prov, manifest)...)
	}
	planRaw, err := os.ReadFile(engineeringPath(projectDir, planCopyFile))
	if err != nil {
		findings = append(findings, errorFinding("materialization-plan-missing", "materialization-plan.json is missing"))
	} else if !strings.Contains(string(planRaw), manifest.PlanFingerprint) {
		findings = append(findings, errorFinding("materialization-plan-drift",
			"materialization-plan.json does not reference manifest fingerprint %q", shortHash(manifest.PlanFingerprint)))
	}
	for _, extra := range unexpectedEntries(projectDir, manifest) {
		findings = append(findings, warnFinding("unexpected-entry", "unexpected top-level entry %q is not tracked by the manifest", extra))
	}
	return findings, nil
}

// unexpectedEntries lists top-level entries that are neither generated
// agent-context files nor manifest-tracked destination roots.
func unexpectedEntries(projectDir string, m Manifest) []string {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil
	}
	roots := map[string]bool{EngineeringDir: true, "AGENTS.md": true, "ARCHITECTURE.md": true, "GENTLE.md": true}
	for _, c := range m.Components {
		seg := c.Destination
		if i := strings.Index(seg, "/"); i >= 0 {
			seg = seg[:i]
		}
		roots[seg] = true
	}
	var out []string
	for _, e := range entries {
		if !roots[e.Name()] {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// checkEvolutionEvents validates provenance evolution events against the
// manifest. An event naming a surface that is neither materialized nor
// mapped is a warning, never an error: the checker stays
// forward-compatible with event types and surfaces minted by newer cores.
func checkEvolutionEvents(prov Provenance, m Manifest) []Finding {
	if len(prov.Events) == 0 {
		return nil
	}
	known := map[string]bool{}
	for _, c := range m.Components {
		known[c.Surface] = true
	}
	var findings []Finding
	for i, ev := range prov.Events {
		if ev.Type == "" {
			findings = append(findings, warnFinding("evolution-event-untyped",
				"provenance event %d has no type", i))
			continue
		}
		if ev.Surface == "" {
			continue
		}
		if !known[ev.Surface] {
			findings = append(findings, warnFinding("evolution-event-unknown-surface",
				"provenance event %q names surface %q, which is not materialized (forward-compatible: kept as warning)", ev.Type, ev.Surface))
		}
	}
	return findings
}

func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

func describeMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+m[k])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}
