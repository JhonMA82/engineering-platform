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

// namePlaceholder is the legacy single placeholder kept for error
// messages; resolution now uses the full closed vocabulary from domain.
const namePlaceholder = "{name}"

// validAppName keeps generated directory names portable across platforms:
// the placeholder can never inject separators, dots-only escapes or spaces.
var validAppName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// GenerationValues carries the resolved placeholder values for one
// generated component. Output is the controlled sandbox location the
// generator must write to; it is resolved at materialization time and
// never serialized in plans.
type GenerationValues struct {
	Name    string
	Project string
	Surface string
	Profile string
	Output  string
}

// generationTimeoutError marks timeout failures so callers can attribute
// the stage without parsing strings.
type generationStageError struct {
	stage string
	err   error
}

func (e *generationStageError) Error() string { return e.err.Error() }
func (e *generationStageError) Unwrap() error { return e.err }

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
// It stays for the legacy single-command path; multi-placeholder
// resolution uses domain.SubstituteGeneratorPlaceholders.
func substituteName(argv []string, name string) []string {
	out := make([]string, len(argv))
	for i, arg := range argv {
		out[i] = strings.ReplaceAll(arg, namePlaceholder, name)
	}
	return out
}

// valuesMap flattens GenerationValues for placeholder substitution.
func (v GenerationValues) valuesMap() map[string]string {
	return map[string]string{
		"name": v.Name, "project": v.Project, "surface": v.Surface,
		"profile": v.Profile, "output": v.Output,
	}
}

// RunGenerate executes a generic generator declaration with the legacy
// single-placeholder contract: {name} derives from the destination and
// the command runs in a fresh work directory under workRoot. It exists
// for schema v1 plans and direct unit-test use; the pipeline prefers
// RunGenerateWithValues.
func RunGenerate(ctx context.Context, spec *domain.GenerateSpec, destination, workRoot string, timeout time.Duration) (string, error) {
	if spec == nil {
		return "", domain.Materialization("generate: adapter declares no generate spec")
	}
	name, err := generateAppName(destination)
	if err != nil {
		return "", err
	}
	work, err := os.MkdirTemp(workRoot, "generate-*")
	if err != nil {
		return "", domain.Filesystem(fmt.Sprintf("create generate dir: %v", err))
	}
	values := GenerationValues{Name: name, Project: "project", Surface: name, Output: filepath.Join(work, "output")}
	return runGenerateIn(ctx, spec, values, work, work, timeout, nil)
}

// RunGenerateWithValues executes a generator with fully resolved
// placeholders. factoryDir is the acquired pinned generator source (or ""
// when the generator is an external pinned tool and needs no factory);
// sandbox is the per-component isolated workspace owning the output.
// extraArgs carries the resolved component Materialization.Arguments (the
// curated profile arguments serialized in the plan); they are substituted
// with the same values and appended to the base Run argv, order preserved.
// The curated prepare steps run inside the factory (or sandbox when no
// factory was acquired); the run command executes with the same working
// directory and must produce the declared output inside the sandbox.
// Only the output directory is returned: the factory itself never flows
// into the project.
func RunGenerateWithValues(ctx context.Context, spec *domain.GenerateSpec, values GenerationValues, factoryDir, sandbox string, timeout time.Duration, extraArgs []string) (string, error) {
	if spec == nil {
		return "", domain.Materialization("generate: adapter declares no generate spec")
	}
	if err := spec.Validate(); err != nil {
		return "", err
	}
	cwd := factoryDir
	if cwd == "" {
		cwd = sandbox
	}
	return runGenerateIn(ctx, spec, values, cwd, sandbox, timeout, extraArgs)
}

// runGenerateIn runs prepare plus the generator command with cwd as the
// working directory and confines the declared output to sandboxRoot.
// extraArgs are substituted element-wise with the same values as the base
// Run argv (never shell strings) and appended after it; the FULL effective
// argv is validated once with ValidateCommands so profile arguments cannot
// bypass the argv-only gate.
func runGenerateIn(ctx context.Context, spec *domain.GenerateSpec, values GenerationValues, cwd, sandboxRoot string, timeout time.Duration, extraArgs []string) (string, error) {
	if err := spec.Validate(); err != nil {
		return "", err
	}
	base, err := domain.SubstituteGeneratorPlaceholders(spec.Run.Run, values.valuesMap())
	if err != nil {
		return "", err
	}
	extra, err := domain.SubstituteGeneratorPlaceholders(extraArgs, values.valuesMap())
	if err != nil {
		return "", err
	}
	argv := append(append([]string{}, base...), extra...)
	if err := ValidateCommands([]domain.AdapterCommand{{Run: argv}}); err != nil {
		return "", err
	}
	outputRaw, err := domain.SubstituteGeneratorPlaceholder(spec.Output, values.valuesMap())
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(outputRaw) == "" {
		return "", domain.Materialization("generate: output must not be empty")
	}
	failed := true
	defer func() {
		if failed {
			_ = os.RemoveAll(sandboxRoot)
		}
	}()
	if len(spec.Prepare) > 0 {
		prepared := make([]domain.AdapterCommand, 0, len(spec.Prepare))
		for _, cmd := range spec.Prepare {
			resolved, err := domain.SubstituteGeneratorPlaceholders(cmd.Run, values.valuesMap())
			if err != nil {
				return "", err
			}
			prepared = append(prepared, domain.AdapterCommand{Run: resolved})
		}
		if _, err := RunCommands(ctx, prepared, cwd, timeout); err != nil {
			return "", &generationStageError{stage: "prepare", err: domain.ExternalCommand(fmt.Sprintf(
				"generator preparation failed: %v", err))}
		}
	}
	if _, err := RunCommands(ctx, []domain.AdapterCommand{{Run: argv}}, cwd, timeout); err != nil {
		return "", &generationStageError{stage: "run", err: domain.ExternalCommand(fmt.Sprintf(
			"generator execution failed: %v", err))}
	}
	outDir := outputRaw
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(cwd, filepath.FromSlash(outputRaw))
	}
	st, err := os.Stat(outDir)
	if err != nil || !st.IsDir() {
		return "", domain.Materialization(fmt.Sprintf(
			"generator output missing: command produced no output directory %q", outputRaw))
	}
	resolved, err := filepath.EvalSymlinks(outDir)
	if err != nil {
		return "", domain.Materialization(fmt.Sprintf("generator output %q: %v", outputRaw, err))
	}
	sandboxResolved, err := filepath.EvalSymlinks(sandboxRoot)
	if err != nil {
		sandboxResolved = sandboxRoot
	}
	if resolved != sandboxResolved && !strings.HasPrefix(resolved, sandboxResolved+string(os.PathSeparator)) {
		return "", domain.Materialization(fmt.Sprintf(
			"generator output escaped sandbox: %q is outside its working directory", outputRaw))
	}
	cwdResolved, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		cwdResolved = cwd
	}
	if resolved == cwdResolved {
		return "", domain.Materialization(fmt.Sprintf(
			"generator output validation failed: output %q is the factory root itself; only generated output may flow into the project", outputRaw))
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return "", domain.Materialization(fmt.Sprintf("generator output %q: %v", outputRaw, err))
	}
	if len(entries) == 0 {
		return "", domain.Materialization(fmt.Sprintf(
			"generator output validation failed: output directory %q is empty", outputRaw))
	}
	failed = false
	return resolved, nil
}
