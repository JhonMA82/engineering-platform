package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAdapterAcceptsCopyStrategy keeps legacy string adapters on copy.
func TestAdapterAcceptsCopyStrategy(t *testing.T) {
	var b Boilerplate
	raw := `{"id":"x","repo":"https://example.com/r","pin":"p","adapter":"plain","provides":{},"delivery_status":"pilot-ready","decision_status":"default"}`
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	if got := b.MaterializationStrategy(); got != StrategyCopy {
		t.Fatalf("strategy = %q, want copy", got)
	}
}

// TestAdapterAcceptsGenerateStrategy accepts the declarative generate form.
func TestAdapterAcceptsGenerateStrategy(t *testing.T) {
	var b Boilerplate
	raw := `{"id":"x","repo":"https://example.com/r","pin":"p",` +
		`"adapter":{"name":"g","operations":["fetch","generate"],` +
		`"generate":{"run":["fake","--out={output}"],"output":"{output}"}},` +
		`"provides":{},"delivery_status":"pilot-ready","decision_status":"default"}`
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	if err := b.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got := b.MaterializationStrategy(); got != StrategyGenerate {
		t.Fatalf("strategy = %q, want generate", got)
	}
}

// TestGenerateRequiresRunCommand rejects a generate spec without argv.
func TestGenerateRequiresRunCommand(t *testing.T) {
	g := GenerateSpec{Output: "{name}"}
	if err := g.Validate(); err == nil {
		t.Fatal("expected run-required failure, got nil")
	}
}

// TestGenerateRequiresSafeOutputDeclaration rejects escapes and absolutes.
func TestGenerateRequiresSafeOutputDeclaration(t *testing.T) {
	for _, output := range []string{"", "../out", "/abs", "a//b", "{name}/../../x"} {
		g := GenerateSpec{Run: AdapterCommand{Run: []string{"fake"}}, Output: output}
		if err := g.Validate(); err == nil {
			t.Fatalf("output %q: expected failure, got nil", output)
		}
	}
}

// TestGenerateRejectsUnknownPlaceholder rejects undeclared {tokens}.
func TestGenerateRejectsUnknownPlaceholder(t *testing.T) {
	g := GenerateSpec{
		Run:    AdapterCommand{Run: []string{"fake", "--prompt={prompt}"}},
		Output: "{name}",
	}
	if err := g.Validate(); err == nil {
		t.Fatal("expected unknown-placeholder failure, got nil")
	} else if !strings.Contains(err.Error(), "{prompt}") {
		t.Fatalf("error should name the token: %v", err)
	}
	out := GenerateSpec{
		Run:    AdapterCommand{Run: []string{"fake"}},
		Output: "{destination}",
	}
	if err := out.Validate(); err == nil {
		t.Fatal("expected unknown output placeholder failure, got nil")
	}
}

// TestGenerateRejectsInteractiveOnlyContract documents that interactive
// generators have no representation: a template that needs a TTY answer
// cannot be expressed without a placeholder, and unknown placeholders
// fail. Curation must only select non-interactive contracts.
func TestGenerateRejectsInteractiveOnlyContract(t *testing.T) {
	g := GenerateSpec{
		Run:    AdapterCommand{Run: []string{"fake", "--answer={tty_answer}"}},
		Output: "{name}",
	}
	if err := g.Validate(); err == nil {
		t.Fatal("expected interactive-placeholder failure, got nil")
	}
}

// TestStaticAdapterRemainsBackwardCompatible keeps v1 adapters valid.
func TestStaticAdapterRemainsBackwardCompatible(t *testing.T) {
	var b Boilerplate
	raw := `{"id":"s","repo":"https://example.com/r","pin":"p",` +
		`"adapter":{"name":"s","operations":["fetch","copy"]},` +
		`"provides":{},"delivery_status":"pilot-ready","decision_status":"default"}`
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	if err := b.Validate(); err != nil {
		t.Fatalf("legacy adapter must stay valid: %v", err)
	}
	if fp := AdapterFingerprint(b.EffectiveSpec()); fp == "" {
		t.Fatal("fingerprint must not be empty")
	}
}

// TestGeneratorProfilesHaveUniqueIDs rejects duplicate profile ids.
func TestGeneratorProfilesHaveUniqueIDs(t *testing.T) {
	g := GenerateSpec{
		Run:            AdapterCommand{Run: []string{"fake"}},
		Output:         "{name}",
		DefaultProfile: "a",
		Profiles: []GeneratorProfile{
			{ID: "a", Arguments: []string{"--p=a"}},
			{ID: "a", Arguments: []string{"--p=a2"}},
		},
	}
	if err := g.Validate(); err == nil {
		t.Fatal("expected duplicate-profile failure, got nil")
	}
}

// TestDefaultGeneratorProfileExists requires the default to be declared.
func TestDefaultGeneratorProfileExists(t *testing.T) {
	missing := GenerateSpec{
		Run:            AdapterCommand{Run: []string{"fake"}},
		Output:         "{name}",
		DefaultProfile: "ghost",
		Profiles:       []GeneratorProfile{{ID: "a"}},
	}
	if err := missing.Validate(); err == nil {
		t.Fatal("expected missing-default failure, got nil")
	}
	empty := GenerateSpec{
		Run:      AdapterCommand{Run: []string{"fake"}},
		Output:   "{name}",
		Profiles: []GeneratorProfile{{ID: "a"}},
	}
	if err := empty.Validate(); err == nil {
		t.Fatal("expected empty-default failure, got nil")
	}
}

// TestGeneratorCommandsRejectShellExecution pins the boundary: domain
// validation stays structural (argv shape), while shell refusal lives in
// the materializer process boundary. A shell-looking argv must parse
// here and fail there.
func TestGeneratorCommandsRejectShellExecution(t *testing.T) {
	cmd := AdapterCommand{Run: []string{"sh", "-c", "echo hi"}}
	if err := cmd.Validate(); err != nil {
		t.Fatalf("domain stays structural, must accept shape: %v", err)
	}
}

// TestAdapterFingerprintChangesOnContractEdit pins drift sensitivity.
func TestAdapterFingerprintChangesOnContractEdit(t *testing.T) {
	base := AdapterSpec{
		Name: "g", Operations: []string{"fetch", "generate"},
		Generate: &GenerateSpec{
			Run: AdapterCommand{Run: []string{"fake", "--out={output}"}}, Output: "{output}",
			DefaultProfile: "minimal",
			Profiles:       []GeneratorProfile{{ID: "minimal", Arguments: []string{"--profile=minimal"}}},
		},
	}
	edited := base
	edited.Generate = &GenerateSpec{
		Run: base.Generate.Run, Output: base.Generate.Output,
		DefaultProfile: base.Generate.DefaultProfile,
		Profiles: []GeneratorProfile{
			{ID: "minimal", Arguments: []string{"--profile=minimal", "--extra"}},
		},
	}
	if AdapterFingerprint(base) == AdapterFingerprint(edited) {
		t.Fatal("fingerprint must change when profile arguments change")
	}
}
