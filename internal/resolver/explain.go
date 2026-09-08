package resolver

import (
	"fmt"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Explain renders a human Why/Rejected view of a decision for `eng explain`.
func Explain(d domain.ArchitectureDecision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "status: %s\n", d.Status)
	if d.Selected != nil {
		fmt.Fprintf(&b, "selected: %s (score %d)\n", d.Selected.Recipe, d.Selected.Score)
	}
	fmt.Fprintf(&b, "confidence: %s (margin %d)\n", d.Confidence.Level, d.Confidence.Margin)
	b.WriteString("\nwhy:\n")
	if len(d.Reasons) == 0 {
		b.WriteString("  (no reasons recorded)\n")
	}
	for _, r := range d.Reasons {
		fmt.Fprintf(&b, "  - %s\n", r)
	}
	b.WriteString("\nrejected:\n")
	any := false
	for _, c := range d.Candidates {
		if d.Selected != nil && c.Recipe == d.Selected.Recipe {
			continue
		}
		any = true
		fmt.Fprintf(&b, "  %s eligible=%v score=%d\n", c.Recipe, c.Eligible, c.Score)
		for _, n := range c.NegativeReasons {
			fmt.Fprintf(&b, "    - %s\n", n)
		}
	}
	if !any {
		b.WriteString("  (no other candidates)\n")
	}
	if len(d.DerivedRequirements) > 0 {
		b.WriteString("\nderived requirements:\n")
		for _, dr := range d.DerivedRequirements {
			fmt.Fprintf(&b, "  - %s (%s)\n", dr.ID, dr.Source)
		}
	}
	if len(d.UnresolvedDimensions) > 0 {
		b.WriteString("\nquestions for discovery:\n")
		for _, u := range d.UnresolvedDimensions {
			fmt.Fprintf(&b, "  - %s: %s [options: %s]\n", u.Dimension, u.Reason, strings.Join(u.Options, ", "))
		}
	}
	if len(d.MissingArchitecture) > 0 {
		b.WriteString("\nmissing architecture:\n")
		for _, m := range d.MissingArchitecture {
			fmt.Fprintf(&b, "  - %s=%s\n", m.Kind, m.Ref)
			for _, rc := range m.ResearchCriteria {
				fmt.Fprintf(&b, "      research: %s\n", rc)
			}
		}
	}
	return b.String()
}
