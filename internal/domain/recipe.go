package domain

import "strings"

// Provides declares what a recipe or boilerplate covers.
type Provides struct {
	Surfaces     []SurfaceID    `json:"surfaces,omitempty"`
	Capabilities []CapabilityID `json:"capabilities,omitempty"`
}

// DatabasePolicy selects profiles without hardcoding brands in the core.
type DatabasePolicy struct {
	DefaultProfile  string   `json:"default_profile,omitempty"`
	AllowedProfiles []string `json:"allowed_profiles,omitempty"`
	SharedBackend   bool     `json:"shared_backend,omitempty"`
}

// Recipe is a candidate architecture profile. It deliberately carries no
// project_types field; matching derives from provides and composition.
type Recipe struct {
	ID                        string         `json:"id"`
	Version                   string         `json:"version,omitempty"`
	Status                    string         `json:"status,omitempty"`
	Description               string         `json:"description,omitempty"`
	Provides                  Provides       `json:"provides,omitempty"`
	PrimaryBoilerplates       []string       `json:"primary_boilerplates,omitempty"`
	AllowedSurfaceComposition [][]SurfaceID  `json:"allowed_surface_composition,omitempty"`
	DatabasePolicy            DatabasePolicy `json:"database_policy,omitempty"`
	TechTags                  []string       `json:"tech_tags,omitempty"`
	QualityGates              []string       `json:"quality_gates,omitempty"`
}

// Active reports whether the recipe participates in candidate generation.
func (r Recipe) Active() bool {
	return r.Status == "stable" || r.Status == "active"
}

// Validate checks structural presence of a recipe record.
func (r Recipe) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return Validation("recipe id is required")
	}
	if strings.TrimSpace(r.Version) == "" {
		return Validation("recipe version is required")
	}
	return nil
}
