package resolver

import (
	"testing"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Background jobs condition the foundation (worker topology), so R4 must
// promote them to a canonical architectural requirement carried by data.
func TestDeriveBackgroundJobs(t *testing.T) {
	cases := []struct {
		name string
		ops  domain.OperationalIntent
		want bool
	}{
		{"background jobs derive the capability", domain.OperationalIntent{BackgroundJobs: true}, true},
		{"no ops signal derives nothing new", domain.OperationalIntent{}, false},
		{"realtime alone does not imply workers", domain.OperationalIntent{Realtime: true}, false},
		{"file processing alone does not imply workers", domain.OperationalIntent{FileProcessing: true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cat, err := catalog.Load("")
			if err != nil {
				t.Fatalf("load catalog: %v", err)
			}
			idx := catalog.NewIndex(cat)
			intent := domain.ProjectIntent{
				SchemaVersion: 1,
				Name:          "jobs",
				Surfaces:      []domain.SurfaceIntent{{Kind: "api", Scope: domain.ScopeRequiredNow}},
				Ops:           tc.ops,
			}
			derived := Derive(Normalize(intent, idx), SplitRequirements(Normalize(intent, idx)))
			found := false
			for _, d := range derived {
				if d.ID == "background-processing" {
					found = true
				}
			}
			if found != tc.want {
				t.Fatalf("background-processing derived = %v, want %v (derived: %+v)", found, tc.want, derived)
			}
		})
	}
}
