// Package app evolution services: thin boundary over internal/evolve.
//
// The CLI stays thin and the resolver decides; these functions only load the
// active catalog, stamp time and delegate to the evolution planner.
package app

import (
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/evolve"
)

// SurfaceAddProject adds a new required surface to a materialized project.
// provider optionally pins the foundation; empty selects the composed
// default. catalogDir is the overlay catalog directory (empty = base).
func SurfaceAddProject(projectDir, surface, provider, catalogDir string) (*evolve.EvolutionResult, error) {
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	return evolve.SurfaceAdd(projectDir, evolve.SurfaceAddOptions{
		Surface:  surface,
		Provider: provider,
		Catalog:  cat,
		Now:      time.Now().UTC(),
	})
}

// ExtendProjectScope promotes a planned_later surface to required_now.
func ExtendProjectScope(projectDir, surface, catalogDir string) (*evolve.EvolutionResult, error) {
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	return evolve.ExtendScope(projectDir, evolve.ExtendOptions{
		Surface: surface,
		Catalog: cat,
		Now:     time.Now().UTC(),
	})
}

// AddProjectRequirement records a product requirement. scope is
// required_now (default) or planned_later.
func AddProjectRequirement(projectDir, text, scope, catalogDir string) (*evolve.RequirementResult, error) {
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	return evolve.AddRequirement(projectDir, evolve.AddRequirementOptions{
		Text:    text,
		Scope:   scope,
		Catalog: cat,
		Now:     time.Now().UTC(),
	})
}

// UpdateProjectReport compares materialized pins with the active catalog.
// It never mutates the project.
func UpdateProjectReport(projectDir, catalogDir string) (*evolve.UpdateReport, error) {
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	return evolve.BuildUpdateReport(projectDir, cat)
}
