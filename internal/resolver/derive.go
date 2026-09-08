package resolver

import (
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Derive runs R4: visible rules turning intent signals into canonical
// architectural requirements.
func Derive(n Normalized, sp Split) []domain.DerivedRequirement {
	var out []domain.DerivedRequirement
	seen := map[string]bool{}
	add := func(id, source string) {
		if !seen[id] {
			seen[id] = true
			out = append(out, domain.DerivedRequirement{ID: id, Source: source})
		}
	}
	hasSurface := func(id domain.SurfaceID) bool {
		for _, s := range n.RequiredSurfaces {
			if s == id {
				return true
			}
		}
		return false
	}
	for _, s := range n.Intent.Surfaces {
		if s.EffectiveScope() != domain.ScopeRequiredNow {
			continue
		}
		switch s.Access {
		case "anonymous", "mixed", "public":
			add("anonymous-public-access", "anonymous surface access requires public-access")
		}
	}
	if n.Intent.Data.PublicAccess == "yes" {
		add("anonymous-public-access", "data declares public access")
	}
	sharedData := n.Intent.Data.MultiUser || n.Intent.Data.Persistence == "shared"
	if hasSurface("web-admin") && hasSurface("mobile-native") && sharedData {
		add("shared-backend", "admin web plus mobile-native share business data")
	}
	if hasSurface("api") && len(n.RequiredSurfaces) >= 2 {
		add("shared-backend", "multiple clients consume the same api")
	}
	if n.Intent.Ops.OfflineOperation || n.Intent.Data.OfflineSync {
		add("offline-operation", "offline operation changes the foundation")
	}
	if n.Intent.Ops.BackgroundJobs {
		add("background-processing", "background jobs need a worker topology, not the request path")
	}
	if n.Intent.Ops.Realtime {
		add("realtime", "realtime operation changes the foundation")
	}
	return out
}

// RequiredCapabilities merges direct required refs that are capabilities
// with derived capability ids, in stable order.
func RequiredCapabilities(n Normalized, derived []domain.DerivedRequirement) []domain.CapabilityID {
	seen := map[domain.CapabilityID]bool{}
	for _, d := range derived {
		seen[domain.CapabilityID(d.ID)] = true
	}
	for _, r := range n.RequiredRefs {
		seen[domain.CapabilityID(r)] = true
	}
	var out []domain.CapabilityID
	for c := range seen {
		out = append(out, c)
	}
	sortCapabilities(out)
	return out
}

func sortCapabilities(out []domain.CapabilityID) {
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
}
