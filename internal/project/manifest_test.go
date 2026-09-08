package project

import (
	"os"
	"path/filepath"
	"testing"
)

func writeManifest(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, EngineeringDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(engineeringPath(dir, ManifestFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestReadManifestLegacyStardriveAlias maps manifests written before the
// legacy-identity repair (§18): the retired stardrive-public-web id reads as
// the canonical stardrive provider. The alias is reader-side only — the
// catalog never lists it and new manifests record the canonical id.
func TestReadManifestLegacyStardriveAlias(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{"schema_version":1,"project":"legacy-site","recipe":"GP-01",
"catalog_version":"1.0.0","plan_fingerprint":"abc",
"components":[{"surface":"public-web","boilerplate":"stardrive-public-web","pin":"v1.0.0","destination":"apps/web"}],
"files":[]}`)
	m, err := ReadManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if len(m.Components) != 1 {
		t.Fatalf("components = %+v, want one", m.Components)
	}
	if m.Components[0].Boilerplate != "stardrive" {
		t.Fatalf("boilerplate = %q, want canonical stardrive", m.Components[0].Boilerplate)
	}
}

// TestReadManifestKeepsCanonicalIDs proves the alias rewrites nothing else:
// canonical ids and unknown ids pass through untouched.
func TestReadManifestKeepsCanonicalIDs(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{"schema_version":1,"project":"sites","recipe":"GP-06",
"catalog_version":"1.0.0","plan_fingerprint":"abc",
"components":[
{"surface":"public-web","boilerplate":"stardrive","pin":"sha","destination":"apps/web"},
{"surface":"web-admin","boilerplate":"tanstack-admin","pin":"sha","destination":"apps/admin"}],
"files":[]}`)
	m, err := ReadManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	got := map[string]string{}
	for _, c := range m.Components {
		got[c.Surface] = c.Boilerplate
	}
	if got["public-web"] != "stardrive" || got["web-admin"] != "tanstack-admin" {
		t.Fatalf("components rewritten: %+v", got)
	}
}
