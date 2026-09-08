package catalog

import (
	"fmt"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Validate enforces catalog invariants: unique ids, dangling refs,
// pin/adapter presence and min_core_version.
func Validate(c Catalog) error {
	var problems []string
	if strings.TrimSpace(c.CatalogVersion) == "" {
		problems = append(problems, "catalog_version is required")
	}
	if !validSemver(c.MinCoreVersion) {
		problems = append(problems, "min_core_version must look like X.Y.Z")
	}
	if strings.TrimSpace(c.MaxCoreVersion) != "" && !validSemver(c.MaxCoreVersion) {
		problems = append(problems, "max_core_version must look like X.Y.Z")
	}
	seen := map[string]bool{}
	uniq := func(kind, id string) {
		if id == "" || seen[kind+"\x00"+id] {
			problems = append(problems, fmt.Sprintf("duplicate %s id %q", kind, id))
			return
		}
		seen[kind+"\x00"+id] = true
	}
	surfaces := map[domain.SurfaceID]bool{}
	for _, s := range c.Surfaces {
		uniq("surface", string(s.ID))
		if err := s.Validate(); err != nil {
			problems = append(problems, err.Error())
		}
		surfaces[s.ID] = true
	}
	caps := map[domain.CapabilityID]bool{}
	for _, cp := range c.Capabilities {
		uniq("capability", string(cp.ID))
		if err := cp.Validate(); err != nil {
			problems = append(problems, err.Error())
		}
		caps[cp.ID] = true
	}
	boilerplates := map[string]bool{}
	for _, b := range c.Boilerplates {
		uniq("boilerplate", b.ID)
		if err := b.Validate(); err != nil {
			problems = append(problems, err.Error())
		}
		boilerplates[b.ID] = true
	}
	profiles := map[string]bool{}
	for _, p := range c.DatabaseProfiles {
		uniq("database-profile", p.ID)
		if err := p.Validate(); err != nil {
			problems = append(problems, err.Error())
		}
		profiles[p.ID] = true
	}
	if len(c.Recipes) == 0 {
		problems = append(problems, "catalog must declare at least one recipe")
	}
	for _, r := range c.Recipes {
		uniq("recipe", r.ID)
		if err := r.Validate(); err != nil {
			problems = append(problems, err.Error())
		}
		for _, s := range r.Provides.Surfaces {
			if !surfaces[s] {
				problems = append(problems, fmt.Sprintf("recipe %s references unknown surface %q", r.ID, s))
			}
		}
		for _, cp := range r.Provides.Capabilities {
			if !caps[cp] {
				problems = append(problems, fmt.Sprintf("recipe %s references unknown capability %q", r.ID, cp))
			}
		}
		for _, b := range r.PrimaryBoilerplates {
			if !boilerplates[b] {
				problems = append(problems, fmt.Sprintf("recipe %s references unknown boilerplate %q", r.ID, b))
			}
		}
		for _, combo := range r.AllowedSurfaceComposition {
			for _, s := range combo {
				if !surfaces[s] {
					problems = append(problems, fmt.Sprintf("recipe %s composes unknown surface %q", r.ID, s))
				}
			}
		}
		if dp := r.DatabasePolicy.DefaultProfile; dp != "" && !profiles[dp] {
			problems = append(problems, fmt.Sprintf("recipe %s references unknown database profile %q", r.ID, dp))
		}
		for _, p := range r.DatabasePolicy.AllowedProfiles {
			if !profiles[p] {
				problems = append(problems, fmt.Sprintf("recipe %s allows unknown database profile %q", r.ID, p))
			}
		}
	}
	if len(problems) > 0 {
		return domain.Catalog(strings.Join(problems, "; "))
	}
	return nil
}

func validSemver(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}
