package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Delivery vocabulary for the single status axis (§3.2). "stable" is the
// legacy release-grade value kept for compatibility: it enforces the same
// evidence bar as "released". New entries use catalog-only, pilot-ready,
// curated or released.
var validDeliveryStates = map[string]bool{
	"catalog-only": true, "pilot-ready": true, "curated": true,
	"released": true, "stable": true,
}

// pilotLine matches the structured pilot record every curated+ evidence stub
// must carry (see docs/guides/curate-boilerplate.md): a "Pilot:" line whose
// value describes the successful pilot.
var pilotLine = regexp.MustCompile(`(?im)^\s*(?:[-*]\s*)?pilot\s*:\s*(.+?)\s*$`)

// pilotDenials marks Pilot: values that admit no successful pilot ran. The
// check is deliberately literal: curators write "Pilot: not run" and the
// entry stays pilot-ready until a real pilot lands.
var pilotDenials = regexp.MustCompile(`(?i)\b(none|pending|todo|missing|tbd)\b|not\s+run|not\s+executed|no\s+pilot|n/a`)

// ValidateCuration enforces the H3 curation contract (§3): every boilerplate
// above catalog-only must link verifiable curation evidence, and curated+
// entries must record a successful pilot. Evidence paths resolve against
// catalogDirs in order (overlay first, then the base catalog dir), so
// overlay entries can ship their own evidence with no core changes.
//
// Rules by effective delivery status (curation.status defaults to
// delivery_status; a set curation.status must match it — one axis):
//
//	catalog-only: evidence optional (a declared link must still resolve).
//	pilot-ready:  link required; file must exist, be non-empty and address
//	              the license; repo/pin/adapter presence is structural and
//	              already enforced by domain validation.
//	curated, released, stable: all of the above plus a Pilot: success record.
func ValidateCuration(c Catalog, catalogDirs []string) error {
	var problems []string
	for _, b := range c.Boilerplates {
		status := strings.TrimSpace(b.DeliveryStatus)
		if status == "" && strings.TrimSpace(b.Curation.Status) != "" {
			status = strings.TrimSpace(b.Curation.Status)
		}
		if strings.TrimSpace(b.Curation.Status) != "" &&
			strings.TrimSpace(b.Curation.Status) != strings.TrimSpace(b.DeliveryStatus) {
			problems = append(problems, fmt.Sprintf(
				"boilerplate %s: curation.status %q must match delivery_status %q (single status axis)",
				b.ID, b.Curation.Status, b.DeliveryStatus))
			continue
		}
		if status == "" {
			// No status anywhere: pre-H3 entries stay structurally valid;
			// the CLI reports them so curators can backfill.
			continue
		}
		if !validDeliveryStates[status] {
			problems = append(problems, fmt.Sprintf(
				"boilerplate %s: unknown delivery_status %q (want catalog-only|pilot-ready|curated|released)", b.ID, status))
			continue
		}
		evidence := strings.TrimSpace(b.Curation.Evidence)
		if status == "catalog-only" && evidence == "" {
			continue
		}
		if evidence == "" {
			problems = append(problems, fmt.Sprintf(
				"boilerplate %s: delivery_status %q requires curation evidence", b.ID, status))
			continue
		}
		body, err := readEvidence(catalogDirs, evidence)
		if err != nil {
			problems = append(problems, fmt.Sprintf("boilerplate %s: %v", b.ID, err))
			continue
		}
		if len(strings.TrimSpace(string(body))) == 0 {
			problems = append(problems, fmt.Sprintf(
				"boilerplate %s: curation evidence %q is empty", b.ID, evidence))
			continue
		}
		if !mentionsLicense(string(body)) {
			problems = append(problems, fmt.Sprintf(
				"boilerplate %s: curation evidence %q does not address the license", b.ID, evidence))
			continue
		}
		if status == "curated" || status == "released" || status == "stable" {
			if pilot := pilotRecord(string(body)); pilot == "" {
				problems = append(problems, fmt.Sprintf(
					"boilerplate %s: delivery_status %q requires a Pilot: success record in %q", b.ID, status, evidence))
				continue
			}
		}
	}
	if len(problems) > 0 {
		return domain.Catalog(strings.Join(problems, "; "))
	}
	return nil
}

// readEvidence resolves a catalog-relative evidence path against catalogDirs
// in order and returns the file body. Paths stay confined to the catalog:
// absolute paths, ".." escapes and backslashes are rejected, and a symlink
// whose target leaves the catalog directory aborts the read.
func readEvidence(catalogDirs []string, evidence string) ([]byte, error) {
	if strings.TrimSpace(evidence) == "" {
		return nil, fmt.Errorf("curation evidence path is empty")
	}
	if filepath.IsAbs(evidence) || strings.Contains(evidence, "\\") {
		return nil, fmt.Errorf("curation evidence %q must be a relative slash path inside the catalog", evidence)
	}
	for _, seg := range strings.Split(filepath.ToSlash(evidence), "/") {
		if seg == ".." {
			return nil, fmt.Errorf("curation evidence %q escapes the catalog", evidence)
		}
		if strings.TrimSpace(seg) == "" {
			return nil, fmt.Errorf("curation evidence %q is not clean", evidence)
		}
	}
	var tried []string
	for _, dir := range catalogDirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		candidate := filepath.Join(dir, filepath.FromSlash(evidence))
		raw, err := os.ReadFile(candidate)
		if os.IsNotExist(err) {
			tried = append(tried, dir)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("curation evidence %q: %v", evidence, err)
		}
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return nil, fmt.Errorf("curation evidence %q: %v", evidence, err)
		}
		base, err := filepath.EvalSymlinks(dir)
		if err != nil {
			base = dir
		}
		if resolved != base && !strings.HasPrefix(resolved, base+string(os.PathSeparator)) {
			return nil, fmt.Errorf("curation evidence %q escapes the catalog via symlink", evidence)
		}
		if st, err := os.Stat(candidate); err != nil || st.IsDir() {
			return nil, fmt.Errorf("curation evidence %q is not a file", evidence)
		}
		return raw, nil
	}
	if len(tried) == 0 {
		return nil, fmt.Errorf("curation evidence %q: no catalog directory to resolve against", evidence)
	}
	return nil, fmt.Errorf("curation evidence %q does not exist in the catalog", evidence)
}

// mentionsLicense is the basic-evidence tripwire: the stub must address
// licensing (even if only to record "unverified" as an explicit gap).
func mentionsLicense(body string) bool {
	return strings.Contains(strings.ToLower(body), "license")
}

// pilotRecord returns the Pilot: value when the stub records a successful
// pilot, or "" when no pilot ran (or the value denies one).
func pilotRecord(body string) string {
	m := pilotLine.FindStringSubmatch(body)
	if m == nil {
		return ""
	}
	if pilotDenials.MatchString(m[1]) {
		return ""
	}
	return strings.TrimSpace(m[1])
}
