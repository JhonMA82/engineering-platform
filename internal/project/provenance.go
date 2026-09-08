package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// ProvenanceFile is the provenance filename inside EngineeringDir.
const ProvenanceFile = "provenance.json"

// EvolutionEvent records one project evolution step. History is
// append-only: events are added, never rewritten, so provenance keeps the
// full lineage from bootstrap through every surface-add, scope extension
// and requirement addition.
type EvolutionEvent struct {
	Type        string `json:"type"`
	At          string `json:"at"`
	Surface     string `json:"surface,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Recipe      string `json:"recipe,omitempty"`
	Requirement string `json:"requirement,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

// Evolution event types emitted by the eng evolution commands.
const (
	EventSurfaceAdd     = "surface-add"
	EventScopeExtend    = "scope-extend"
	EventRequirementAdd = "requirement-add"
)

// Provenance records how a project came to be: core and catalog versions,
// fingerprints, pins and the materialization timestamp. It is the ONLY
// document allowed to carry timestamps; decisions and plans stay
// deterministic and timestamp-free.
type Provenance struct {
	SchemaVersion     string            `json:"schema_version"`
	CoreVersion       string            `json:"core_version"`
	CatalogVersion    string            `json:"catalog_version"`
	IntentFingerprint string            `json:"intent_fingerprint,omitempty"`
	PlanFingerprint   string            `json:"plan_fingerprint"`
	Pins              map[string]string `json:"pins"`
	MaterializedAt    string            `json:"materialized_at"`
	Events            []EvolutionEvent  `json:"events,omitempty"`
}

// BuildProvenance assembles the provenance record. now is injected so tests
// stay deterministic; production passes time.Now().UTC().
func BuildProvenance(coreVersion, catalogVersion, intentFingerprint, planFingerprint string, pins map[string]string, now time.Time) Provenance {
	cp := map[string]string{}
	for k, v := range pins {
		cp[k] = v
	}
	return Provenance{
		SchemaVersion:     "1",
		CoreVersion:       coreVersion,
		CatalogVersion:    catalogVersion,
		IntentFingerprint: intentFingerprint,
		PlanFingerprint:   planFingerprint,
		Pins:              cp,
		MaterializedAt:    now.UTC().Format(time.RFC3339),
	}
}

// ProvenanceRelPath is the project-relative slash path of the record.
const ProvenanceRelPath = EngineeringDir + "/" + ProvenanceFile

// Marshal returns the canonical .engineering/provenance.json bytes.
func (p Provenance) Marshal() ([]byte, error) {
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, domain.Filesystem(fmt.Sprintf("encode provenance: %v", err))
	}
	return append(raw, '\n'), nil
}

// ReadProvenance loads .engineering/provenance.json.
func ReadProvenance(projectDir string) (Provenance, error) {
	var p Provenance
	if err := readEngineeringJSON(projectDir, ProvenanceFile, &p); err != nil {
		return Provenance{}, err
	}
	return p, nil
}

// AppendEvent appends one evolution event to the project provenance and
// persists the canonical bytes. History is never rewritten: earlier events
// are preserved verbatim. An empty At is stamped with now (UTC, RFC3339).
// Only provenance.json is touched; manifest and map updates are the
// caller's responsibility.
func AppendEvent(projectDir string, ev EvolutionEvent, now time.Time) error {
	p, err := ReadProvenance(projectDir)
	if err != nil {
		return err
	}
	if ev.Type == "" {
		return domain.Filesystem("evolution event type is required")
	}
	if ev.At == "" {
		ev.At = now.UTC().Format(time.RFC3339)
	}
	p.Events = append(p.Events, ev)
	raw, err := p.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(engineeringPath(projectDir, ProvenanceFile), raw, 0o644); err != nil {
		return domain.Filesystem(fmt.Sprintf("write %s: %v", ProvenanceFile, err))
	}
	return nil
}

// WriteProvenance persists the canonical provenance bytes. It overwrites
// only the rolling pointers the caller already updated (fingerprints,
// pins); MaterializedAt and Events lineage must be preserved by the caller.
func WriteProvenance(projectDir string, p Provenance) error {
	raw, err := p.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(engineeringPath(projectDir, ProvenanceFile), raw, 0o644); err != nil {
		return domain.Filesystem(fmt.Sprintf("write %s: %v", ProvenanceFile, err))
	}
	return nil
}

// IntentFingerprintOf extracts intent_fingerprint from a decision document
// when present; it returns "" for missing or unreadable input.
func IntentFingerprintOf(decisionJSON []byte) string {
	if len(strings.TrimSpace(string(decisionJSON))) == 0 {
		return ""
	}
	var doc struct {
		IntentFingerprint string `json:"intent_fingerprint"`
	}
	if err := json.Unmarshal(decisionJSON, &doc); err != nil {
		return ""
	}
	return doc.IntentFingerprint
}
