// Package project owns the machine-readable state of a materialized
// project: manifest, provenance, project map and the doctor check.
//
// It performs reads and writes under a given project directory but never
// fetches sources or executes commands; those side effects belong to the
// materializer. It imports stdlib plus the pure domain and planner types.
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// EngineeringDir is the project metadata directory.
const EngineeringDir = ".engineering"

// ManifestFile is the project manifest filename inside EngineeringDir.
const ManifestFile = "project.json"

// ManifestComponent records one materialized surface provider.
// Strategy, Name, Profile, Arguments and AdapterFingerprint carry the
// generation configuration for generated foundations; schema v1 records
// omit them and read as copy.
type ManifestComponent struct {
	Surface            string   `json:"surface"`
	Boilerplate        string   `json:"boilerplate"`
	Pin                string   `json:"pin"`
	Destination        string   `json:"destination"`
	Strategy           string   `json:"strategy,omitempty"`
	Name               string   `json:"name,omitempty"`
	Profile            string   `json:"profile,omitempty"`
	Arguments          []string `json:"arguments,omitempty"`
	AdapterFingerprint string   `json:"adapter_fingerprint,omitempty"`
}

// Manifest is the reproducible record of what was materialized: plan
// fingerprint, catalog version, pins and the resulting file list. It
// carries no timestamps; see Provenance for the dated record.
type Manifest struct {
	SchemaVersion   int                 `json:"schema_version"`
	Project         string              `json:"project"`
	Recipe          string              `json:"recipe"`
	RecipeVersion   string              `json:"recipe_version,omitempty"`
	CatalogVersion  string              `json:"catalog_version"`
	DatabaseProfile string              `json:"database_profile,omitempty"`
	PlanFingerprint string              `json:"plan_fingerprint"`
	Components      []ManifestComponent `json:"components"`
	Files           []string            `json:"files"`
}

// BuildManifest derives the manifest from the executed plan, the catalog
// version and the sorted slash-separated relative file list.
func BuildManifest(plan planner.MaterializationPlan, catalogVersion string, files []string) Manifest {
	components := make([]ManifestComponent, 0, len(plan.Components))
	for _, c := range plan.Components {
		mc := ManifestComponent{
			Surface:     c.Surface,
			Boilerplate: c.Boilerplate,
			Pin:         c.Pin,
			Destination: c.Destination,
		}
		if c.Materialization.Strategy != "" {
			mc.Strategy = c.Materialization.Strategy
			mc.Name = c.Materialization.Name
			mc.Profile = c.Materialization.Profile
			mc.Arguments = append([]string{}, c.Materialization.Arguments...)
			mc.AdapterFingerprint = c.Materialization.AdapterFingerprint
		}
		components = append(components, mc)
	}
	sort.Slice(components, func(i, j int) bool {
		if components[i].Destination != components[j].Destination {
			return components[i].Destination < components[j].Destination
		}
		return components[i].Surface < components[j].Surface
	})
	sorted := append([]string{}, files...)
	sort.Strings(sorted)
	return Manifest{
		SchemaVersion:   2,
		Project:         plan.Project,
		Recipe:          plan.Recipe,
		RecipeVersion:   plan.RecipeVersion,
		CatalogVersion:  catalogVersion,
		DatabaseProfile: plan.DatabaseProfile,
		PlanFingerprint: plan.Fingerprint,
		Components:      components,
		Files:           sorted,
	}
}

// ManifestRelPath is the project-relative slash path of the manifest.
const ManifestRelPath = EngineeringDir + "/" + ManifestFile

// Marshal returns the canonical .engineering/project.json bytes.
func (m Manifest) Marshal() ([]byte, error) {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("encode manifest: %v", err))
	}
	return append(raw, '\n'), nil
}

// legacyBoilerplateAliases maps retired catalog ids to their canonical
// successor for READING manifests written before the rename (§18). It is
// applied only by ReadManifest; the catalog never lists an alias as an
// active provider, and new manifests always record the canonical id.
var legacyBoilerplateAliases = map[string]string{
	"stardrive-public-web": "stardrive",
}

// ReadManifest loads .engineering/project.json.
func ReadManifest(projectDir string) (Manifest, error) {
	var m Manifest
	if err := readEngineeringJSON(projectDir, ManifestFile, &m); err != nil {
		return Manifest{}, err
	}
	if m.SchemaVersion < 1 {
		return Manifest{}, domain.Filesystem("project manifest schema_version must be >= 1")
	}
	for i := range m.Components {
		if canonical, ok := legacyBoilerplateAliases[m.Components[i].Boilerplate]; ok {
			m.Components[i].Boilerplate = canonical
		}
	}
	return m, nil
}

func engineeringPath(projectDir, name string) string {
	return filepath.Join(projectDir, EngineeringDir, name)
}

func readEngineeringJSON(projectDir, name string, v any) error {
	raw, err := os.ReadFile(engineeringPath(projectDir, name))
	if err != nil {
		return domain.Filesystem(fmt.Sprintf("read %s: %v", name, err))
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return domain.Filesystem(fmt.Sprintf("parse %s: %v", name, err))
	}
	return nil
}
