// Package opencode exposes the embedded Engineering Platform OpenCode
// bootstrap resources. The markdown files in this directory are the single
// source of truth for what `eng init --agent opencode` installs into
// project-local .opencode/ paths; they are embedded into the eng binary so
// initialization works from a release artifact without cloning the
// repository.
//
// The layout uses the OpenCode v2 preferred paths
// (.opencode/commands/<name>.md and
// .opencode/skills/<skill-id>/SKILL.md with `$ARGUMENTS` argument
// injection), which remain compatible with OpenCode v1 per the v1 to v2
// migration guide: supported v1 commands, skills and .opencode/ files
// continue to work in v2 without changes.
package opencode

import _ "embed"

// NewprojectMD is the /newproject command installed to
// .opencode/commands/newproject.md (preferred path in OpenCode v1 and v2).
//
//go:embed commands/newproject.md
var NewprojectMD string

// DiscoverySkillMD is the discovery skill installed to
// .opencode/skills/engineering-project-discovery/SKILL.md (preferred
// layout in OpenCode v1 and v2; the directory name is the skill ID).
//
//go:embed skills/engineering-project-discovery/SKILL.md
var DiscoverySkillMD string
