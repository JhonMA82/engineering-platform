package materializer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/foundationconfig"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// argvRecorderDir installs an executable fixture script mirroring the
// hono-api contract: Run=[<script> --output {output}] carries NO profile in
// Run; the profile travels as Materialization.Arguments (extraArgs). The
// script records the exact argv it received into argv.txt under the output
// dir so tests can assert the effective command line.
func argvRecorderDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "argv-recorder")
	body := "#!/bin/sh\n" +
		"set -u\n" +
		"OUTPUT=\"\"\n" +
		"prev=\"\"\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$prev\" = \"--output\" ]; then OUTPUT=\"$a\"; fi\n" +
		"  prev=\"$a\"\n" +
		"done\n" +
		"[ -n \"$OUTPUT\" ] || { echo \"output is required\" >&2; exit 2; }\n" +
		"mkdir -p \"$OUTPUT\"\n" +
		"printf '%s\\n' \"$(basename \"$0\")\" \"$@\" > \"$OUTPUT/argv.txt\"\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

// argsRecorderSpec declares the generator base command with NO profile baked
// in: baking --profile into Run would hide the dropped-arguments bug.
func argsRecorderSpec() *domain.GenerateSpec {
	return &domain.GenerateSpec{
		Run:    domain.AdapterCommand{Run: []string{"argv-recorder", "--output", "{output}"}},
		Output: "{output}",
	}
}

func argsRecorderValues(sandbox string) GenerationValues {
	return GenerationValues{
		Name: "demo", Project: "demo", Surface: "api",
		Profile: "authenticated", Output: filepath.Join(sandbox, "output"),
	}
}

// normArg resolves existing paths so recorded argv compares equal to the
// expected argv even when the temp dir passes through a symlink.
func normArg(s string) string {
	if r, err := filepath.EvalSymlinks(s); err == nil {
		return r
	}
	return s
}

func assertArgv(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argv = %q, want %q", got, want)
	}
	for i := range want {
		if normArg(got[i]) != normArg(want[i]) {
			t.Fatalf("argv = %q, want %q", got, want)
		}
	}
}

// runRecorderWithArgs executes the recorder spec and returns the sandbox,
// the resolved output dir and the recorded argv lines.
func runRecorderWithArgs(t *testing.T, extraArgs []string) (string, string, []string) {
	t.Helper()
	argvRecorderDir(t)
	sandbox := t.TempDir()
	out, err := RunGenerateWithValues(context.Background(),
		argsRecorderSpec(), argsRecorderValues(sandbox), "", sandbox, time.Minute, extraArgs)
	if err != nil {
		t.Fatalf("RunGenerateWithValues: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "argv.txt"))
	if err != nil {
		t.Fatalf("read argv.txt: %v", err)
	}
	return sandbox, out, strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
}

// TestGenerateNoArgumentsKeepsBaseArgv pins the legacy path: with zero
// component arguments the effective argv is exactly the substituted base.
func TestGenerateNoArgumentsKeepsBaseArgv(t *testing.T) {
	sandbox, _, lines := runRecorderWithArgs(t, nil)
	assertArgv(t, lines, []string{"argv-recorder", "--output", filepath.Join(sandbox, "output")})
}

// TestGenerateSingleArgumentThreaded proves the resolved component argument
// reaches the generator: the old code dropped Materialization.Arguments.
func TestGenerateSingleArgumentThreaded(t *testing.T) {
	sandbox, _, lines := runRecorderWithArgs(t, []string{"--profile=authenticated"})
	assertArgv(t, lines, []string{
		"argv-recorder", "--output", filepath.Join(sandbox, "output"),
		"--profile=authenticated",
	})
}

// TestGenerateMultipleArgumentsPreserveOrder pins base++args ordering.
func TestGenerateMultipleArgumentsPreserveOrder(t *testing.T) {
	sandbox, _, lines := runRecorderWithArgs(t, []string{"--profile=authenticated", "--verbose"})
	assertArgv(t, lines, []string{
		"argv-recorder", "--output", filepath.Join(sandbox, "output"),
		"--profile=authenticated", "--verbose",
	})
}

// TestGenerateArgumentsResolvePlaceholders proves extra args go through the
// same placeholder substitution as the base Run.
func TestGenerateArgumentsResolvePlaceholders(t *testing.T) {
	sandbox, _, lines := runRecorderWithArgs(t, []string{"--out={output}", "--name={name}"})
	assertArgv(t, lines, []string{
		"argv-recorder", "--output", filepath.Join(sandbox, "output"),
		"--out=" + filepath.Join(sandbox, "output"), "--name=demo",
	})
}

// TestGenerateArgumentsRejectedByValidation proves the FULL effective argv
// (base + args) passes through ValidateCommands: a metacharacter in an
// argument fails before anything executes.
func TestGenerateArgumentsRejectedByValidation(t *testing.T) {
	argvRecorderDir(t)
	sandbox := t.TempDir()
	_, err := RunGenerateWithValues(context.Background(),
		argsRecorderSpec(), argsRecorderValues(sandbox), "", sandbox, time.Minute,
		[]string{"--profile=authenticated;evil"})
	if err == nil {
		t.Fatal("expected validation failure for metacharacter argument, got nil")
	}
	if !strings.Contains(err.Error(), "metacharacters") {
		t.Fatalf("wrong failure: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(sandbox, "output", "argv.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("rejected arguments must not execute: stat err = %v", statErr)
	}
}

// TestMaterializeThreadsPlanArgumentsToGenerator proves the pipeline
// wiring end to end: the plan's Materialization.Arguments reach the executed
// generator argv through a full Materialize, not just the unit path. The
// recorded --output value is sandbox-internal, so the test pins the shape
// (base ++ single profile argument) instead of the temp path.
func TestMaterializeThreadsPlanArgumentsToGenerator(t *testing.T) {
	argvRecorderDir(t)
	bp := domain.Boilerplate{
		ID:      "fixture-recorder",
		Pin:     "v0.0.0-recorder",
		Adapter: "fixture-rec",
		AdapterSpec: &domain.AdapterSpec{
			Name:       "fixture-rec",
			Operations: []string{"generate"},
			Generate: &domain.GenerateSpec{
				Run:            domain.AdapterCommand{Run: []string{"argv-recorder", "--output", "{output}"}},
				Output:         "{output}",
				DefaultProfile: "authenticated",
				Profiles: []domain.GeneratorProfile{
					{ID: "authenticated", Arguments: []string{"--profile=authenticated"}},
				},
			},
			ManagedFiles: []string{"AGENTS.md"},
		},
		Source:         domain.SourceSpec{Type: "local", Path: t.TempDir()},
		DeliveryStatus: "stable",
		DecisionStatus: "curated",
	}
	cat := catalog.Catalog{
		CatalogVersion: "0.0.0-test", MinCoreVersion: "1.0.0", SchemaVersion: 1,
		Boilerplates: []domain.Boilerplate{bp},
	}
	plan := planner.MaterializationPlan{
		SchemaVersion: 2, Project: "demo", Recipe: "TEST",
		RecipeVersion: "1.0.0", DatabaseProfile: "sqlite-local",
		Fingerprint: "test-fingerprint-args",
		Components: []planner.PlanComponent{{
			Boilerplate: "fixture-recorder", Pin: "v0.0.0-recorder",
			Destination: "apps/api", Surface: "api",
			Materialization: foundationconfig.MaterializationConfig{
				Strategy:           domain.StrategyGenerate,
				Name:               "demo-api",
				Profile:            "authenticated",
				Arguments:          []string{"--profile=authenticated"},
				AdapterFingerprint: domain.AdapterFingerprint(*bp.AdapterSpec),
			},
		}},
	}
	out := filepath.Join(t.TempDir(), "proj")
	if _, err := Materialize(happyRequest(t, cat, plan, out)); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "apps", "api", "argv.txt"))
	if err != nil {
		t.Fatalf("read recorded argv: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != 4 || lines[0] != "argv-recorder" || lines[1] != "--output" ||
		lines[3] != "--profile=authenticated" {
		t.Fatalf("effective argv = %q, want base ++ [--profile=authenticated]", lines)
	}
	if st, err := os.Stat(filepath.Join(out, "apps", "api", "argv.txt")); err != nil || st.IsDir() {
		t.Fatalf("recorded argv.txt missing in project: %v", err)
	}
	if !filepath.IsAbs(lines[2]) || filepath.Base(lines[2]) != "output" {
		t.Fatalf("recorded output %q is not the sandbox output dir", lines[2])
	}
}
