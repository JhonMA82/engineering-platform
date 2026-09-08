package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
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
