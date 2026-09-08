package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// Offline materialization pilots (§53): the Golden Path slices below run
// exclusively against local fixture sources via the shared evolveOverlay, so
// CI needs no network and no real upstreams. Each pilot proves the full
// spine — intent → decision → plan → materialize → doctor — for one path.

const (
	// pilotAdminMobileIntent mirrors testdata/routing/admin-mobile.json.
	pilotAdminMobileIntent = `{"schema_version":1,"name":"field ops","problem":"office and field share visits",` +
		`"surfaces":[{"kind":"web-admin","access":"authenticated","scope":"required_now"},` +
		`{"kind":"mobile-native","access":"authenticated","scope":"required_now"}],` +
		`"data":{"persistence":"shared","multi_user":true}}`
	// pilotMobileIntent mirrors testdata/routing/mobile-field-app.json.
	pilotMobileIntent = `{"schema_version":1,"name":"field inspections","problem":"inspectors capture photo evidence with gps",` +
		`"product_requirements":[{"id":"camera","description":"capture photo evidence per visit"}],` +
		`"surfaces":[{"kind":"mobile-native","access":"authenticated","scope":"required_now"}],` +
		`"data":{"persistence":"shared","multi_user":true}}`
)

// TestPilotMaterialization materializes one Golden Path slice per case with
// fixture sources only and requires a doctor-clean project carrying the
// Gentle handoff artifacts.
func TestPilotMaterialization(t *testing.T) {
	cases := []struct {
		name           string
		intent         string
		recipe         string
		wantComponents int
	}{
		{name: "GP-02-single-surface", intent: evolveAdminOnlyIntent, recipe: "GP-02", wantComponents: 1},
		{name: "GP-06-multi-surface", intent: pilotAdminMobileIntent, recipe: "GP-06", wantComponents: 3},
		{name: "GP-04-mobile-backend", intent: pilotMobileIntent, recipe: "GP-04", wantComponents: 2},
	}
	// The GP-04 recipe prefers its curated ignite primary, which points at a
	// real upstream. The overlay below repoints that id at a local fixture
	// source so the mobile pilot stays offline; recipe selection is untouched
	// (the resolver never reads boilerplate sources).
	overrideIgniteForOffline := func(t *testing.T, overlay string) {
		t.Helper()
		webSrc, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "boilerplates", "fixture-web"))
		if err != nil {
			t.Fatal(err)
		}
		webPin := readPIN(t, webSrc)
		body := fmt.Sprintf(`{
          "id": "ignite",
          "repo": "https://github.com/infinitered/ignite",
          "pin": %q,
          "adapter": {"name": "ignite", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
          "source": {"type": "local", "path": %q},
          "delivery_status": "stable",
          "decision_status": "curated",
          "provides": {"surfaces": ["mobile-native"], "capabilities": []},
          "tech_tags": ["react-native", "expo"],
          "included_features": []
        }`, webPin, webSrc)
		if err := os.WriteFile(filepath.Join(overlay, "boilerplates", "ignite.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			overlay := evolveOverlay(t)
			if tc.recipe == "GP-04" {
				overrideIgniteForOffline(t, overlay)
			}
			out, recipe := materializeEvolveProject(t, overlay, tc.intent)
			if recipe != tc.recipe {
				t.Fatalf("recipe = %q, want %q", recipe, tc.recipe)
			}
			assertDoctorGreen(t, out)
			assertPilotArtifacts(t, out)
		})
	}
}

// TestPilotEvolutionSurfaceAdd grows a piloted GP-06 project by one surface
// and requires the evolved tree to stay doctor-clean.
func TestPilotEvolutionSurfaceAdd(t *testing.T) {
	overlay := evolveOverlay(t)
	out, recipe := materializeEvolveProject(t, overlay, evolveWebAdminAPIIntent)
	if recipe != "GP-06" {
		t.Fatalf("base recipe = %q, want GP-06", recipe)
	}
	res, err := SurfaceAddProject(out, "public-intake", "", overlay)
	if err != nil {
		t.Fatalf("surface add: %v", err)
	}
	if res.Recipe != "GP-06" {
		t.Errorf("evolved recipe = %q, want GP-06", res.Recipe)
	}
	if len(res.Added) != 1 || res.Added[0].Surface != "public-intake" {
		t.Fatalf("added = %+v, want exactly the public-intake delta", res.Added)
	}
	assertDoctorGreen(t, out)
	assertPilotArtifacts(t, out)
}

// assertPilotArtifacts requires the agent-navigation files every piloted
// project must carry: root router, handoff trio and stable intent copy.
func assertPilotArtifacts(t *testing.T, projectDir string) {
	t.Helper()
	for _, want := range []string{
		"AGENTS.md",
		".engineering/project-intent.json",
		".engineering/project-map.json",
		".engineering/implementation-brief.md",
		".engineering/handoff.json",
	} {
		if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected %s: %v", want, err)
		}
	}
}

// TestPilotIgniteGeneration is the H4 upstream pilot: it materializes the
// real ignite adapter (generic generate op) through npx + network. Normal
// suites never run it: besides -short and the npx/network probes below, it
// requires ENG_UPSTREAM_PILOTS=1, so `go test ./...` stays offline and
// deterministic. Explicit run:
//
//	ENG_UPSTREAM_PILOTS=1 go test ./internal/app/ -run TestPilotIgniteGeneration -timeout 30m -v
//
// Follow-up recorded in docs/decisions/ignite-materialization.md:
// confirming or extending setup/checks from the generated tree.
func TestPilotIgniteGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode: upstream pilot skipped")
	}
	if os.Getenv("ENG_UPSTREAM_PILOTS") != "1" {
		t.Skip("ENG_UPSTREAM_PILOTS != 1: upstream pilot skipped")
	}
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not on PATH: upstream pilot skipped")
	}
	probeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(probeCtx, "git", "ls-remote",
		"https://github.com/infinitered/ignite", "HEAD").CombinedOutput(); err != nil {
		t.Skipf("ignite upstream unreachable: upstream pilot skipped (%v: %s)", err, string(out))
	}
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	idx := catalog.NewIndex(cat)
	bp, ok := idx.Boilerplate("ignite")
	if !ok {
		t.Fatal("ignite missing from catalog")
	}
	if bp.EffectiveSpec().Generate == nil {
		t.Fatal("ignite adapter declares no generate spec")
	}
	plan := planner.MaterializationPlan{
		SchemaVersion: 1,
		Project:       "ignite-pilot",
		Recipe:        "GP-04",
		Components: []planner.PlanComponent{{
			Surface: "mobile-native", Boilerplate: "ignite", Pin: bp.Pin, Destination: "apps/mobile",
		}},
		Fingerprint: "ignite-network-pilot",
	}
	out := filepath.Join(t.TempDir(), "proj")
	res, err := materializer.Materialize(materializer.Request{
		Plan:           plan,
		Catalog:        cat,
		OutputDir:      out,
		CoreVersion:    CoreVersion,
		CommandTimeout: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("ignite materialization: %v", err)
	}
	if res.ProjectDir != out {
		t.Fatalf("project dir = %q, want %q", res.ProjectDir, out)
	}
	// A generated Ignite app always ships its package manifest and the
	// Expo config at the scaffold root.
	for _, want := range []string{"apps/mobile/package.json", "apps/mobile/app.json"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Errorf("expected generated %s: %v", want, err)
		}
	}
	assertPilotArtifacts(t, out)
}
