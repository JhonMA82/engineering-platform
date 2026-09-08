// Package domain holds the M1 core contracts.
//
// It must import zero infrastructure: no os, os/exec, net/http, cobra.
package domain

import (
	"fmt"
	"sort"
	"strings"
)

// ScopeStatus distinguishes work required now from future or excluded work.
type ScopeStatus string

const (
	ScopeRequiredNow  ScopeStatus = "required_now"
	ScopePlannedLater ScopeStatus = "planned_later"
	ScopeExcluded     ScopeStatus = "explicitly_excluded"
)

// SurfaceID and CapabilityID are nominal string types validated by the
// catalog, never closed enums compiled into the core.
type SurfaceID string

// CapabilityID is a nominal string type validated by the catalog.
type CapabilityID string

// SurfaceIntent expresses one required interface/client.
type SurfaceIntent struct {
	Kind   SurfaceID   `json:"kind"`
	Access string      `json:"access,omitempty"`
	Scope  ScopeStatus `json:"scope,omitempty"`
}

// EffectiveScope defaults an empty scope to required_now.
func (s SurfaceIntent) EffectiveScope() ScopeStatus {
	if s.Scope == "" {
		return ScopeRequiredNow
	}
	return s.Scope
}

// DataIntent captures data needs before brand decisions.
type DataIntent struct {
	Persistence   string `json:"persistence,omitempty"`
	MultiUser     bool   `json:"multi_user,omitempty"`
	Relational    string `json:"relational,omitempty"`
	OfflineSync   bool   `json:"offline_sync,omitempty"`
	ExpectedScale string `json:"expected_scale,omitempty"`
	PublicAccess  string `json:"public_access,omitempty"`
}

// OperationalIntent captures runtime needs; only promoted to architecture
// requirements when they change foundation or composition.
type OperationalIntent struct {
	BackgroundJobs   bool `json:"background_jobs,omitempty"`
	Realtime         bool `json:"realtime,omitempty"`
	OfflineOperation bool `json:"offline_operation,omitempty"`
	FileProcessing   bool `json:"file_processing,omitempty"`
}

// ProductRequirement describes what the product must do. Product features
// must never become hard routing constraints.
type ProductRequirement struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

// ArchitectureRequirement describes a property that conditions the
// foundation. Ref points at a surface or capability id.
type ArchitectureRequirement struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Ref         string `json:"ref,omitempty"`
	Strength    string `json:"strength,omitempty"`
}

// EffectiveStrength defaults an empty strength to required.
func (r ArchitectureRequirement) EffectiveStrength() string {
	if r.Strength == "" {
		return "required"
	}
	return r.Strength
}

// TechnicalConstraint is an explicit user decision (obligatory).
type TechnicalConstraint struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
}

// Preference influences ranking but never eliminates candidates.
type Preference struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
}

// ScopeIntent tracks lifecycle placement beyond per-surface scope.
type ScopeIntent struct {
	PlannedLater       []string `json:"planned_later,omitempty"`
	ExplicitlyExcluded []string `json:"explicitly_excluded,omitempty"`
}

// ProjectIntent represents what the product needs, not how it is built.
// There is intentionally no project_type field: routing derives the recipe.
type ProjectIntent struct {
	SchemaVersion            int                       `json:"schema_version"`
	Name                     string                    `json:"name"`
	Problem                  string                    `json:"problem,omitempty"`
	ProductRequirements      []ProductRequirement      `json:"product_requirements,omitempty"`
	ArchitectureRequirements []ArchitectureRequirement `json:"architecture_requirements,omitempty"`
	Surfaces                 []SurfaceIntent           `json:"surfaces,omitempty"`
	Data                     DataIntent                `json:"data,omitempty"`
	Ops                      OperationalIntent         `json:"ops,omitempty"`
	TechnicalConstraints     []TechnicalConstraint     `json:"technical_constraints,omitempty"`
	Preferences              []Preference              `json:"preferences,omitempty"`
	Scope                    ScopeIntent               `json:"scope,omitempty"`
	Notes                    []string                  `json:"notes,omitempty"`
}

// RequiredSurfaces returns required_now surface kinds in stable order.
func (p ProjectIntent) RequiredSurfaces() []SurfaceID {
	var out []SurfaceID
	for _, s := range p.Surfaces {
		if s.EffectiveScope() == ScopeRequiredNow {
			out = append(out, SurfaceID(strings.ToLower(string(s.Kind))))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

var validConstraintKinds = map[string]bool{
	"must-use": true, "must-not-use": true,
	"must-run": true, "must-support": true, "must-share": true,
}

var validPreferenceKinds = map[string]bool{
	"prefer": true, "avoid": true,
}

// Validate enforces R1 structural rules. Unknown architectural vocabulary is
// NOT invalid here; uncovered architecture becomes a catalog-gap later.
func (p ProjectIntent) Validate() error {
	if p.SchemaVersion < 1 {
		return Validation("schema_version must be >= 1")
	}
	if strings.TrimSpace(p.Name) == "" {
		return Validation("name is required")
	}
	required := map[SurfaceID]bool{}
	planned := map[SurfaceID]bool{}
	excluded := map[SurfaceID]bool{}
	for _, s := range p.Surfaces {
		kind := SurfaceID(strings.ToLower(strings.TrimSpace(string(s.Kind))))
		if kind == "" {
			return Validation("surface kind must not be empty")
		}
		switch s.EffectiveScope() {
		case ScopeRequiredNow:
			if required[kind] {
				return Validation(fmt.Sprintf("duplicate required surface %q", kind))
			}
			required[kind] = true
		case ScopePlannedLater:
			planned[kind] = true
		case ScopeExcluded:
			excluded[kind] = true
		default:
			return Validation(fmt.Sprintf("unknown scope %q", s.Scope))
		}
	}
	for _, e := range p.Scope.ExplicitlyExcluded {
		excluded[SurfaceID(strings.ToLower(e))] = true
	}
	for k := range required {
		if excluded[k] {
			return Validation(fmt.Sprintf("surface %q is both required and excluded", k))
		}
	}
	for _, r := range p.ProductRequirements {
		if strings.TrimSpace(r.ID) == "" {
			return Validation("product requirement id is required")
		}
	}
	for _, r := range p.ArchitectureRequirements {
		if strings.TrimSpace(r.ID) == "" {
			return Validation("architecture requirement id is required")
		}
		switch r.EffectiveStrength() {
		case "required", "possible":
		default:
			return Validation(fmt.Sprintf("unknown strength %q", r.Strength))
		}
	}
	mustUse := map[string]bool{}
	mustNotUse := map[string]bool{}
	for _, c := range p.TechnicalConstraints {
		if !validConstraintKinds[c.Kind] {
			return Validation(fmt.Sprintf("unknown constraint kind %q", c.Kind))
		}
		v := strings.ToLower(c.Value)
		if c.Kind == "must-use" {
			mustUse[v] = true
		}
		if c.Kind == "must-not-use" {
			mustNotUse[v] = true
		}
	}
	for v := range mustUse {
		if mustNotUse[v] {
			return Validation(fmt.Sprintf("contradictory constraints on %q", v))
		}
	}
	for _, pr := range p.Preferences {
		if !validPreferenceKinds[pr.Kind] {
			return Validation(fmt.Sprintf("unknown preference kind %q", pr.Kind))
		}
	}
	_ = planned
	if len(required) == 0 && len(p.ArchitectureRequirements) == 0 {
		return Validation("intent carries no architectural signal")
	}
	return nil
}
