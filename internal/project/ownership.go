package project

import (
	"path/filepath"
	"strings"
)

// AiContext-owned files inside EngineeringDir. Engineering operations must
// never delete, overwrite, or require them; the manifest file list must
// never absorb them (see FilterManifestFiles). Unknown files are preserved
// by the same rule: only Engineering-owned paths are ever written.
//
// Ownership table (authoritative):
//
//	Engineering: project.json, provenance.json, project-map.json,
//	  project-intent.json, architecture-decision.json,
//	  materialization-plan.json, implementation-brief.md, handoff.json,
//	  runs/*.json, bootstrap.json, ARCHITECTURE.md, GENTLE.md,
//	  root AGENTS.md (outside aicontext blocks),
//	  <surface>/AGENTS.md (outside aicontext blocks)
//	AiContext: aicontext.toml, PROJECT_STATE.md, PATTERNS.md,
//	  consistency.yml, subprojects.yml, rules/** (all under .engineering/),
//	  <!-- aicontext:* --> blocks inside AGENTS.md (root and surfaces)
var aiContextOwnedFiles = map[string]bool{
	"aicontext.toml":   true,
	"PROJECT_STATE.md": true,
	"PATTERNS.md":      true,
	"consistency.yml":  true,
	"subprojects.yml":  true,
}

// IsAiContextOwned reports whether a project-relative slash path belongs to
// AiContext and must be left untouched by Engineering operations.
func IsAiContextOwned(rel string) bool {
	rel = filepath.ToSlash(rel)
	if rel == EngineeringDir+"/rules" || strings.HasPrefix(rel, EngineeringDir+"/rules/") {
		return true
	}
	if !strings.HasPrefix(rel, EngineeringDir+"/") {
		return false
	}
	base := strings.TrimPrefix(rel, EngineeringDir+"/")
	if strings.Contains(base, "/") {
		return false
	}
	return aiContextOwnedFiles[base]
}

// FilterManifestFiles drops AiContext-owned paths from a manifest file list.
// The manifest stays Engineering's reproducible record; AiContext state is
// validated by `aicontext check`, never by `eng doctor`.
func FilterManifestFiles(files []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if !IsAiContextOwned(f) {
			out = append(out, f)
		}
	}
	return out
}

// AiContext block markers shared inside AGENTS.md. Bytes outside these
// markers are owned by Engineering (or the user) and must never be modified.
var aiContextBlockMarkers = [][2]string{
	{"<!-- aicontext:routing:start -->", "<!-- aicontext:routing:end -->"},
	{"<!-- aicontext:context:start -->", "<!-- aicontext:context:end -->"},
}

// extractMarkedBlock returns the marker-inclusive slice for one block, or "".
func extractMarkedBlock(text, start, end string) string {
	s := strings.Index(text, start)
	if s < 0 {
		return ""
	}
	rest := text[s:]
	e := strings.Index(rest, end)
	if e < 0 {
		return ""
	}
	return rest[:e+len(end)]
}

// MergeAgentsPreservingAicontext carries AiContext-owned marked blocks from
// an existing root AGENTS.md into freshly rendered Engineering content.
// Blocks present in the existing file survive byte-identically; all other
// bytes come from the regenerated file. When no blocks exist the generated
// content is returned unchanged.
func MergeAgentsPreservingAicontext(existing, generated string) string {
	var carried []string
	for _, m := range aiContextBlockMarkers {
		if b := extractMarkedBlock(existing, m[0], m[1]); b != "" {
			carried = append(carried, b)
		}
	}
	if len(carried) == 0 {
		return generated
	}
	out := strings.TrimRight(generated, "\n") + "\n"
	for _, b := range carried {
		if !strings.Contains(out, b) {
			out += "\n" + b + "\n"
		}
	}
	return out
}
