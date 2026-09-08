package materializer

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// namePlaceholder is substituted with the destination basename in generate
// argv and output paths. Every other byte of the declaration is literal.
const namePlaceholder = "{name}"

// validAppName keeps generated directory names portable across platforms:
// the placeholder can never inject separators, dots-only escapes or spaces.
var validAppName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// generateAppName derives the generator placeholder value from a validated
// project destination (e.g. "apps/mobile" -> "mobile").
func generateAppName(destination string) (string, error) {
	base := path.Base(filepath.ToSlash(destination))
	if !validAppName.MatchString(base) {
		return "", domain.Materialization(fmt.Sprintf(
			"generate placeholder: destination %q has no portable app name (base %q)", destination, base))
	}
	return base, nil
}

// substituteName replaces every {name} placeholder with the app name.
func substituteName(argv []string, name string) []string {
	out := make([]string, len(argv))
	for i, arg := range argv {
		out[i] = strings.ReplaceAll(arg, namePlaceholder, name)
	}
	return out
}

// RunGenerate executes a generic generator declaration (H4): an argv-only
// curated command that scaffolds its own output directory, as opposed to
// fetch+copy of a checked-in tree. The command runs with cwd set to a fresh
// work directory under workRoot, allowlisted environment and a timeout; the
// declared output must appear as a directory confined to that work dir.
// Callers copy the returned directory into staging; workRoot cleanup stays
// with the caller (Materialize passes its fetch root, removed by defer).
func RunGenerate(ctx context.Context, spec *domain.GenerateSpec, destination, workRoot string, timeout time.Duration) (string, error) {
	if spec == nil {
		return "", domain.Materialization("generate: adapter declares no generate spec")
	}
	if err := spec.Validate(); err != nil {
		return "", err
	}
	name, err := generateAppName(destination)
	if err != nil {
		return "", err
	}
	argv := substituteName(spec.Run.Run, name)
	if err := ValidateCommands([]domain.AdapterCommand{{Run: argv}}); err != nil {
		return "", err
	}
	outputRel := strings.ReplaceAll(spec.Output, namePlaceholder, name)
	if strings.TrimSpace(outputRel) == "" {
		return "", domain.Materialization("generate: output must not be empty")
	}
	work, err := os.MkdirTemp(workRoot, "generate-*")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create generate dir: %v", err))
	}
	failed := true
	defer func() {
		if failed {
			_ = os.RemoveAll(work)
		}
	}()
	if _, err := RunCommands(ctx, []domain.AdapterCommand{{Run: argv}}, work, timeout); err != nil {
		return "", err
	}
	outDir := filepath.Join(work, filepath.FromSlash(outputRel))
	st, err := os.Stat(outDir)
	if err != nil || !st.IsDir() {
		_ = os.RemoveAll(work)
		return "", domain.Materialization(fmt.Sprintf(
			"generate: command produced no output directory %q", outputRel))
	}
	resolved, err := filepath.EvalSymlinks(outDir)
	if err != nil {
		return "", domain.Materialization(fmt.Sprintf("generate: output %q: %v", outputRel, err))
	}
	workResolved, err := filepath.EvalSymlinks(work)
	if err != nil {
		workResolved = work
	}
	if resolved != workResolved && !strings.HasPrefix(resolved, workResolved+string(os.PathSeparator)) {
		return "", domain.Materialization(fmt.Sprintf(
			"generate: output %q escapes its working directory", outputRel))
	}
	failed = false
	return outDir, nil
}
