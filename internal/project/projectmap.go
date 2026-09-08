package project

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// ProjectMapFile is the project map filename inside EngineeringDir.
const ProjectMapFile = "project-map.json"

// SurfaceEntry is the machine-readable routing record for one surface.
type SurfaceEntry struct {
	Path         string `json:"path"`
	Provider     string `json:"provider"`
	Instructions string `json:"instructions"`
}

// Relationship links two surfaces; v1 only emits consumes edges toward
// the shared backend.
type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// ProjectMap is the machine-readable counterpart of the root AGENTS.md
// router table: where each surface lives, which provider serves it and
// which instructions govern it.
type ProjectMap struct {
	SchemaVersion int                     `json:"schema_version"`
	Surfaces      map[string]SurfaceEntry `json:"surfaces"`
	Relationships []Relationship          `json:"relationships"`
}

// BuildProjectMap derives the map from the executed plan. Instructions
// always point at the per-surface AGENTS.md. Every non-api surface gains
// a consumes edge toward api when a shared backend is present; otherwise
// no relationships are emitted.
func BuildProjectMap(plan planner.MaterializationPlan) ProjectMap {
	surfaces := map[string]SurfaceEntry{}
	hasAPI := false
	for _, c := range plan.Components {
		surfaces[c.Surface] = SurfaceEntry{
			Path:         c.Destination,
			Provider:     c.Boilerplate,
			Instructions: c.Destination + "/AGENTS.md",
		}
		if c.Surface == "api" {
			hasAPI = true
		}
	}
	relationships := []Relationship{}
	if hasAPI {
		names := make([]string, 0, len(surfaces))
		for name := range surfaces {
			if name != "api" {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			relationships = append(relationships, Relationship{From: name, To: "api", Type: "consumes"})
		}
	}
	return ProjectMap{SchemaVersion: 1, Surfaces: surfaces, Relationships: relationships}
}

// ProjectMapRelPath is the project-relative slash path of the map.
const ProjectMapRelPath = EngineeringDir + "/" + ProjectMapFile

// Marshal returns the canonical .engineering/project-map.json bytes.
func (m ProjectMap) Marshal() ([]byte, error) {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("encode project map: %v", err))
	}
	return append(raw, '\n'), nil
}

// ReadProjectMap loads .engineering/project-map.json.
func ReadProjectMap(projectDir string) (ProjectMap, error) {
	var m ProjectMap
	if err := readEngineeringJSON(projectDir, ProjectMapFile, &m); err != nil {
		return ProjectMap{}, err
	}
	if m.Surfaces == nil {
		m.Surfaces = map[string]SurfaceEntry{}
	}
	if m.Relationships == nil {
		m.Relationships = []Relationship{}
	}
	return m, nil
}
