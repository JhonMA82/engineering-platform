// Package pi exposes the embedded Engineering Platform PI bootstrap
// resources. The markdown files in this directory are the single source of
// truth for what `eng init` installs into project-local .pi/ paths; they
// are embedded into the eng binary so initialization works from a release
// artifact without cloning the repository.
package pi

import _ "embed"

// NewprojectMD is the /newproject prompt installed to .pi/prompts/newproject.md.
//
//go:embed prompts/newproject.md
var NewprojectMD string

// DiscoverySkillMD is the project-discovery skill installed to
// .pi/skills/project-discovery/SKILL.md.
//
//go:embed skills/project-discovery/SKILL.md
var DiscoverySkillMD string
