// Package opencode exposes the embedded Engineering Platform OpenCode
// bootstrap resources. The markdown files in this directory are the single
// source of truth for what `eng init --agent opencode` installs into
// project-local .opencode/ paths; they are embedded into the eng binary so
// initialization works from a release artifact without cloning the
// repository.
package opencode

import _ "embed"

// NewprojectMD is the /newproject command installed to
// .opencode/commands/newproject.md.
//
//go:embed commands/newproject.md
var NewprojectMD string

// DiscoverySkillMD is the discovery skill installed to
// .opencode/skills/engineering-project-discovery/SKILL.md.
//
//go:embed skills/engineering-project-discovery/SKILL.md
var DiscoverySkillMD string
