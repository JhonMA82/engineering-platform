package catalog

import (
	"encoding/json"
	"testing"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

func TestBaseCatalogAdaptersAreObjects(t *testing.T) {
	cat, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := Validate(cat); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
	if len(cat.Boilerplates) == 0 {
		t.Fatal("no boilerplates in base catalog")
	}
	for _, b := range cat.Boilerplates {
		if b.DeliveryStatus == "catalog-only" {
			// Catalog knowledge, not a materializable foundation: no
			// adapter is declared and none may be invented (§5.2-5.3).
			continue
		}
		if b.AdapterSpec == nil {
			t.Errorf("%s: expected adapter object, got legacy string", b.ID)
		}
		if b.Adapter == "" {
			t.Errorf("%s: adapter name must survive the object form", b.ID)
		}
	}
}

func TestBoilerplateAdapterForms(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
		check   func(t *testing.T, b domain.Boilerplate)
	}{
		{
			name: "legacy string adapter",
			raw:  `{"id":"s","repo":"https://example/s","pin":"v1","adapter":"hono"}`,
			check: func(t *testing.T, b domain.Boilerplate) {
				t.Helper()
				if b.Adapter != "hono" || b.AdapterSpec != nil {
					t.Errorf("string form parsed as name=%q spec=%v", b.Adapter, b.AdapterSpec)
				}
			},
		},
		{
			name: "object adapter with name",
			raw:  `{"id":"o","repo":"https://example/o","pin":"v1","adapter":{"name":"hono","operations":["fetch","copy"],"managed_files":["AGENTS.md"]}}`,
			check: func(t *testing.T, b domain.Boilerplate) {
				t.Helper()
				if b.AdapterSpec == nil || b.Adapter != "hono" {
					t.Fatalf("object form lost name/spec: %+v", b)
				}
				if len(b.AdapterSpec.Operations) != 2 {
					t.Errorf("operations = %v", b.AdapterSpec.Operations)
				}
			},
		},
		{
			name: "object adapter defaults name to id",
			raw:  `{"id":"n","repo":"https://example/n","pin":"v1","adapter":{"operations":["fetch","copy"]}}`,
			check: func(t *testing.T, b domain.Boilerplate) {
				t.Helper()
				if b.Adapter != "n" {
					t.Errorf("adapter name fallback = %q, want boilerplate id", b.Adapter)
				}
			},
		},
		{
			name:    "numeric adapter rejected",
			raw:     `{"id":"x","repo":"https://example/x","pin":"v1","adapter":42}`,
			wantErr: true,
		},
		{
			name:    "unknown operation rejected",
			raw:     `{"id":"x","repo":"https://example/x","pin":"v1","adapter":{"operations":["teleport"]}}`,
			wantErr: true,
		},
		{
			name:    "prune escape rejected",
			raw:     `{"id":"x","repo":"https://example/x","pin":"v1","adapter":{"prune_paths":["../evil"]}}`,
			wantErr: true,
		},
		{
			name:    "unknown source type rejected",
			raw:     `{"id":"x","repo":"https://example/x","pin":"v1","adapter":"hono","source":{"type":"s3"}}`,
			wantErr: true,
		},
		{
			name:    "local source without path rejected",
			raw:     `{"id":"x","pin":"v1","adapter":"hono","source":{"type":"local"}}`,
			wantErr: true,
		},
		{
			name: "local source without repo accepted",
			raw:  `{"id":"x","pin":"v1","adapter":"hono","source":{"type":"local","path":"/tmp/fixture"}}`,
			check: func(t *testing.T, b domain.Boilerplate) {
				t.Helper()
				if b.Source.Kind() != "local" {
					t.Errorf("source kind = %q", b.Source.Kind())
				}
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var b domain.Boilerplate
			if err := json.Unmarshal([]byte(tt.raw), &b); err != nil {
				if !tt.wantErr {
					t.Fatalf("unmarshal: %v", err)
				}
				return
			}
			err := b.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
			if err == nil && tt.check != nil {
				tt.check(t, b)
			}
		})
	}
}
