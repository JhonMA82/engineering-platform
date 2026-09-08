package materializer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// DefaultCommandTimeout bounds curated adapter commands.
const DefaultCommandTimeout = 5 * time.Minute

// shellBinaries are never executable as adapter commands: execution is
// argv-only with no shell, so any shell indirection fails closed.
var shellBinaries = map[string]bool{
	"sh": true, "bash": true, "dash": true, "ash": true, "zsh": true,
	"fish": true, "ksh": true, "csh": true, "tcsh": true,
	"cmd": true, "cmd.exe": true, "powershell": true, "powershell.exe": true, "pwsh": true,
}

// allowedEnvKeys is the environment allowlist for adapter commands. Only
// locale, path, temp and CI-signal variables pass; everything else —
// notably any secret — is dropped and never logged.
var allowedEnvKeys = map[string]bool{
	"PATH": true, "PATHEXT": true,
	"LANG": true, "LC_ALL": true, "LC_CTYPE": true, "LANGUAGE": true,
	"HOME": true, "USER": true,
	"TMPDIR": true, "TEMP": true, "TMP": true,
	"TZ": true, "NO_COLOR": true, "CLICOLOR": true, "TERM": true, "CI": true,
}

// ValidateCommands rejects commands that could escape argv-only execution:
// empty argv, shell interpreters, shell metacharacters in any element and
// path-like command names (binaries resolve via PATH, never via caller
// supplied paths). Catalog data is trusted but validated: a malicious or
// malformed adapter fails safe here, before anything executes.
func ValidateCommands(cmds []domain.AdapterCommand) error {
	for _, cmd := range cmds {
		if err := cmd.Validate(); err != nil {
			return domain.ExternalCommand(err.Error())
		}
		name := cmd.Run[0]
		base := name
		if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
			base = base[i+1:]
		}
		lower := strings.ToLower(name)
		lowerBase := strings.ToLower(base)
		if shellBinaries[lower] || shellBinaries[lowerBase] {
			return domain.ExternalCommand(fmt.Sprintf("refusing shell interpreter %q: adapter commands run without a shell", name))
		}
		if strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
			return domain.ExternalCommand(fmt.Sprintf("refusing path-like command %q: binaries resolve via PATH", name))
		}
		for _, arg := range cmd.Run {
			// No shell interprets these arguments, so spaces and glob
			// characters are literal and safe; the rejected set covers
			// sequencing, substitution, redirection and control bytes.
			if strings.ContainsAny(arg, ";&|$`()<>\n\r\x00") {
				return domain.ExternalCommand(fmt.Sprintf("refusing command argument %q: shell metacharacters are not allowed", arg))
			}
		}
	}
	return nil
}

// RunCommands executes validated curated commands with dir as working
// directory, allowlisted environment and a timeout. Output is captured;
// the environment is never logged. Validation runs again so direct
// callers cannot bypass it.
func RunCommands(ctx context.Context, cmds []domain.AdapterCommand, dir string, timeout time.Duration) (string, error) {
	if err := ValidateCommands(cmds); err != nil {
		return "", err
	}
	if timeout <= 0 {
		timeout = DefaultCommandTimeout
	}
	var combined strings.Builder
	for _, cmd := range cmds {
		out, err := runOne(ctx, cmd.Run, dir, timeout)
		combined.WriteString(out)
		if err != nil {
			return combined.String(), err
		}
	}
	return combined.String(), nil
}

func runOne(ctx context.Context, argv []string, dir string, timeout time.Duration) (string, error) {
	name := argv[0]
	if _, err := exec.LookPath(name); err != nil {
		return "", domain.ExternalCommand(fmt.Sprintf("command %q not found on PATH", name))
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c := exec.CommandContext(ctx, name, argv[1:]...)
	c.Dir = dir
	c.Env = allowlistEnv()
	var out strings.Builder
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Run(); err != nil {
		return out.String(), domain.ExternalCommand(fmt.Sprintf("command %q failed: %v\n%s", displayArgv(argv), err, tailLines(out.String())))
	}
	return out.String(), nil
}

// allowlistEnv filters the process environment down to the allowlist.
func allowlistEnv() []string {
	var out []string
	for _, kv := range os.Environ() {
		if i := strings.Index(kv, "="); i > 0 && allowedEnvKeys[kv[:i]] {
			out = append(out, kv)
		}
	}
	return out
}

func displayArgv(argv []string) string {
	return strings.Join(argv, " ")
}

// tailLines keeps failure reports bounded without dropping the ending
// where diagnostics usually live.
func tailLines(s string) string {
	const max = 2000
	if len(s) <= max {
		return s
	}
	return "…(truncated)…\n" + s[len(s)-max:]
}
