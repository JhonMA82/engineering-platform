package evolve

import (
	"fmt"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// UpdateEntry compares one materialized component pin against the active
// catalog pin. It is report-only: v1 never auto-applies updates.
type UpdateEntry struct {
	Component       string `json:"component"`
	Destination     string `json:"destination"`
	Current         string `json:"current"`
	Catalog         string `json:"catalog"`
	Strategy        string `json:"strategy"`
	Action          string `json:"action"`
	UpdateAvailable bool   `json:"update_available"`
}

// UpdateReport is the eng update output: one entry per component plus a
// human-readable summary. Current is true when every pin matches.
type UpdateReport struct {
	Entries []UpdateEntry `json:"entries"`
	Current bool          `json:"current"`
	Summary string        `json:"summary"`
}

// BuildUpdateReport compares each manifest component pin with the active
// catalog pin. It performs zero mutation: the project directory is only
// read. An unknown provider is reported with a manual strategy so the next
// step is human review, not a silent skip.
func BuildUpdateReport(projectDir string, cat catalog.Catalog) (*UpdateReport, error) {
	manifest, err := project.ReadManifest(projectDir)
	if err != nil {
		return nil, err
	}
	idx := catalog.NewIndex(cat)
	report := &UpdateReport{Current: true}
	for _, c := range manifest.Components {
		entry := UpdateEntry{
			Component:   c.Boilerplate,
			Destination: c.Destination,
			Current:     c.Pin,
		}
		bp, ok := idx.Boilerplate(c.Boilerplate)
		if !ok {
			entry.Catalog = "(removed from catalog)"
			entry.Strategy = "manual"
			entry.Action = fmt.Sprintf("provider %q left the catalog; manually review %s", c.Boilerplate, c.Destination)
			entry.UpdateAvailable = true
			report.Entries = append(report.Entries, entry)
			continue
		}
		entry.Catalog = bp.Pin
		entry.Strategy = bp.EffectiveUpdateStrategy()
		entry.UpdateAvailable = c.Pin != bp.Pin
		entry.Action = updateAction(entry.Strategy, bp.ID, c.Destination, c.Pin, bp.Pin, entry.UpdateAvailable)
		report.Entries = append(report.Entries, entry)
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		if report.Entries[i].Component != report.Entries[j].Component {
			return report.Entries[i].Component < report.Entries[j].Component
		}
		return report.Entries[i].Destination < report.Entries[j].Destination
	})
	pending := 0
	for _, e := range report.Entries {
		if e.UpdateAvailable {
			pending++
		}
	}
	report.Current = pending == 0
	if report.Current {
		report.Summary = fmt.Sprintf("all current: %d component(s) match the active catalog", len(report.Entries))
	} else {
		report.Summary = fmt.Sprintf("%d of %d component(s) differ from the active catalog (report only — v1 never auto-applies updates)", pending, len(report.Entries))
	}
	return report, nil
}

// updateAction phrases the human next step for a strategy. Current pins
// still name their strategy so the table stays informative when clean.
func updateAction(strategy, boilerplate, dest, current, catalogPin string, available bool) string {
	if !available {
		return fmt.Sprintf("up to date (%s)", strategy)
	}
	switch strategy {
	case "replace":
		return fmt.Sprintf("re-materialize %s from %s@%s", dest, boilerplate, catalogPin)
	case "merge-seed":
		return fmt.Sprintf("merge seed changes from %s@%s into %s", boilerplate, catalogPin, dest)
	case "fork-track":
		return fmt.Sprintf("rebase tracked fork at %s onto %s@%s", dest, boilerplate, catalogPin)
	default:
		return fmt.Sprintf("manually review %s (foundation %s moved %s -> %s)", dest, boilerplate, current, catalogPin)
	}
}
