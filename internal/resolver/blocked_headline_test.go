package resolver

import (
	"strings"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// The blocked-decision headline must name the dominant rejection signal:
// technical constraints only when tech vocabulary actually eliminated
// candidates, composition coverage otherwise. Aggregated reasons must not
// repeat.
func TestBlockedHeadline(t *testing.T) {
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	surface := func(kind string) domain.SurfaceIntent {
		return domain.SurfaceIntent{Kind: domain.SurfaceID(kind), Scope: domain.ScopeRequiredNow}
	}
	cases := []struct {
		name         string
		intent       domain.ProjectIntent
		wantStatus   domain.ResolutionStatus
		wantHeadline string
		wantAbsent   string
	}{
		{
			name: "composition coverage without tech constraints",
			intent: domain.ProjectIntent{
				SchemaVersion: 1, Name: "field desk",
				Surfaces: []domain.SurfaceIntent{surface("web-admin"), surface("public-intake"), surface("mobile-native")},
			},
			wantStatus:   domain.StatusUnsupported,
			wantHeadline: "architecture is catalog-covered but no recipe composes the required surfaces",
			wantAbsent:   "technical constraints",
		},
		{
			name: "technical constraints blamed only when they eliminate",
			intent: domain.ProjectIntent{
				SchemaVersion: 1, Name: "ledger admin",
				Surfaces: []domain.SurfaceIntent{surface("web-admin")},
				TechnicalConstraints: []domain.TechnicalConstraint{
					{Kind: "must-use", Value: "blockchain"},
				},
			},
			wantStatus:   domain.StatusUnsupported,
			wantHeadline: "architecture is catalog-covered but hard technical constraints eliminate every candidate",
			wantAbsent:   "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.intent, cat)
			if got.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", got.Status, tc.wantStatus)
			}
			if len(got.Reasons) == 0 || got.Reasons[0] != tc.wantHeadline {
				t.Fatalf("headline = %q, want %q", first(got.Reasons), tc.wantHeadline)
			}
			if tc.wantAbsent != "" && strings.Contains(strings.Join(got.Reasons, "\n"), tc.wantAbsent) {
				t.Fatalf("reasons mention %q:\n%s", tc.wantAbsent, strings.Join(got.Reasons, "\n"))
			}
			seen := map[string]bool{}
			for _, r := range got.Reasons {
				if seen[r] {
					t.Fatalf("duplicate reason %q", r)
				}
				seen[r] = true
			}
		})
	}
}

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}
