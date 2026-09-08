package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestGenerateSpecShapes proves the generic generate declaration accepts the
// argv forms adapters already use for setup/checks, and rejects malformed
// declarations before anything executes.
func TestGenerateSpecShapes(t *testing.T) {
	t.Run("argv array run", func(t *testing.T) {
		var g GenerateSpec
		raw := `{"run": ["npx", "ignite-cli@11.5.0", "new", "{name}", "--yes"], "output": "{name}"}`
		if err := json.Unmarshal([]byte(raw), &g); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if err := g.Validate(); err != nil {
			t.Fatalf("validate: %v", err)
		}
		if g.Output != "{name}" || len(g.Run.Run) != 5 {
			t.Fatalf("unexpected spec: %+v", g)
		}
	})
	t.Run("command plus args object run", func(t *testing.T) {
		var g GenerateSpec
		raw := `{"run": {"command": "npx", "args": ["new", "{name}"]}, "output": "app"}`
		if err := json.Unmarshal([]byte(raw), &g); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if err := g.Validate(); err != nil {
			t.Fatalf("validate: %v", err)
		}
	})
	for _, tt := range []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"missing run rejected", `{"output": "app"}`, "requires run"},
		{"empty run rejected", `{"run": [], "output": "app"}`, "must not be empty"},
		{"empty output rejected", `{"run": ["gen", "{name}"]}`, "must not be empty"},
		{"absolute output rejected", `{"run": ["gen"], "output": "/tmp/app"}`, "relative slash path"},
		{"traversal output rejected", `{"run": ["gen"], "output": "../app"}`, "escapes"},
		{"not an object rejected", `["gen"]`, "must be an object"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var g GenerateSpec
			if err := json.Unmarshal([]byte(tt.raw), &g); err != nil {
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unmarshal error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err := g.Validate(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validate error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

// TestAdapterGenerateConsistency keeps the operation vocabulary honest: the
// generate word and the generate spec must appear together, never alone.
func TestAdapterGenerateConsistency(t *testing.T) {
	gen := &GenerateSpec{Run: AdapterCommand{Run: []string{"gen", "{name}"}}, Output: "{name}"}
	base := AdapterSpec{Name: "fixture", Operations: []string{"fetch", "copy"}}
	for _, tt := range []struct {
		name    string
		mutate  func(*AdapterSpec)
		wantErr string
	}{
		{"no generate is valid", func(s *AdapterSpec) {}, ""},
		{
			"generate op without spec rejected",
			func(s *AdapterSpec) { s.Operations = []string{"generate"} },
			"no generate spec",
		},
		{
			"generate spec without op rejected",
			func(s *AdapterSpec) { s.Generate = gen },
			"lacks generate",
		},
		{
			"generate op with spec is valid",
			func(s *AdapterSpec) { s.Operations = []string{"generate"}; s.Generate = gen },
			"",
		},
		{
			"unknown operation still rejected",
			func(s *AdapterSpec) { s.Operations = []string{"teleport"} },
			"unknown adapter operation",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			spec := base
			tt.mutate(&spec)
			err := spec.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

// TestBoilerplateCurationRoundTrip proves the curation link survives JSON
// round-trips and stays absent when unset (no `"curation":{}` pollution).
func TestBoilerplateCurationRoundTrip(t *testing.T) {
	raw := `{"id":"b","repo":"https://example.com/b","pin":"v1",` +
		`"adapter":{"name":"b","operations":["fetch","copy"]},` +
		`"delivery_status":"pilot-ready","decision_status":"curated",` +
		`"curation":{"status":"pilot-ready","evidence":"curation/b.md"}}`
	var bp Boilerplate
	if err := json.Unmarshal([]byte(raw), &bp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bp.Curation.Status != "pilot-ready" || bp.Curation.Evidence != "curation/b.md" {
		t.Fatalf("curation = %+v", bp.Curation)
	}
	out, err := json.Marshal(bp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Boilerplate
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if back.Curation != bp.Curation {
		t.Fatalf("round trip curation = %+v, want %+v", back.Curation, bp.Curation)
	}
	var plain Boilerplate
	if err := json.Unmarshal([]byte(`{"id":"p","repo":"r","pin":"v","adapter":"p"}`), &plain); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	plainOut, err := json.Marshal(plain)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(plainOut), "curation") {
		t.Fatalf("unset curation must not marshal, got %s", plainOut)
	}
}
