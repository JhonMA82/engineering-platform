package materializer

import (
	"path"
	"regexp"
	"strings"
	"testing"

	opencodebootstrap "github.com/jhonma82/engineering-platform/integrations/opencode"
)

var validSkillID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// TestOpenCodeV2NativeLayout locks the OpenCode v1+v2 compatible bootstrap
// layout: the v2 preferred paths (.opencode/commands/<name>.md and
// .opencode/skills/<skill-id>/SKILL.md) work unchanged on v1, so `eng init
// --agent opencode` needs no version switch. It also guards the v2 native
// frontmatter rules: no legacy `subtask`/`variant` keys (v2 uses `subagent`
// and `model#variant`), a discoverable skill description, and `$ARGUMENTS`
// injection on the command.
func TestOpenCodeV2NativeLayout(t *testing.T) {
	files := managedFiles("opencode")
	want := []string{
		".opencode/commands/newproject.md",
		".opencode/skills/engineering-project-discovery/SKILL.md",
	}
	if len(files) != len(want) {
		t.Fatalf("managed opencode files = %v, want %v", files, want)
	}
	for i, w := range want {
		if files[i][0] != w {
			t.Fatalf("managed opencode file %d = %q, want %q", i, files[i][0], w)
		}
	}

	// Legacy singular directories (command/, skill/) must never come back:
	// v2 discovers both but prefers the plural form we ship.
	for _, mf := range files {
		if strings.Contains(mf[0], ".opencode/command/") ||
			strings.Contains(mf[0], ".opencode/skill/") {
			t.Fatalf("opencode managed path %q uses legacy singular dir", mf[0])
		}
	}

	// Skill ID is path-derived in v2: directory name must be a portable
	// kebab-case ID and match the frontmatter name.
	skillPath := files[1][0]
	skillID := path.Base(path.Dir(skillPath))
	if !validSkillID.MatchString(skillID) {
		t.Fatalf("skill id %q must match ^[a-z0-9]+(-[a-z0-9]+)*$", skillID)
	}
	if len(skillID) > 64 {
		t.Fatalf("skill id %q exceeds 64 chars", skillID)
	}
	front := opencodebootstrap.DiscoverySkillMD
	if !strings.HasPrefix(front, "---\n") || !strings.Contains(front, "\n---") {
		t.Fatal("skill must carry YAML frontmatter")
	}
	for _, key := range []string{"name:", "description:"} {
		if !strings.Contains(front[:strings.Index(front, "\n---")], key) {
			t.Fatalf("skill frontmatter must contain %q", key)
		}
	}
	if !strings.Contains(front, "name: "+skillID) {
		t.Fatalf("skill frontmatter name must match directory id %q", skillID)
	}

	// Command must inject arguments and stay on v2 native fields.
	cmd := opencodebootstrap.NewprojectMD
	if !strings.Contains(cmd, "$ARGUMENTS") {
		t.Fatal("opencode command must reference $ARGUMENTS")
	}
	if !strings.Contains(cmd, "description:") {
		t.Fatal("opencode command must declare description frontmatter")
	}
	for _, legacy := range []string{"\nsubtask:", "\nvariant:"} {
		if strings.Contains(cmd, legacy) || strings.Contains(front, legacy) {
			t.Fatalf("opencode bootstrap must not use legacy v1-only key %q (v2: subagent / model#variant)", strings.TrimSpace(legacy))
		}
	}

	// Cleanup must prune exactly the v2 directory chain.
	parents := pruneParents("opencode")
	for _, p := range []string{".opencode/commands", ".opencode/skills/engineering-project-discovery", ".opencode/skills", ".opencode"} {
		found := false
		for _, got := range parents {
			if got == p {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("pruneParents(opencode) misses %q (got %v)", p, parents)
		}
	}
}
