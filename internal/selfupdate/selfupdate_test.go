package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestAssetNameCoversReleaseMatrix(t *testing.T) {
	cases := map[[2]string]string{
		{"linux", "amd64"}:   "eng-linux-amd64",
		{"linux", "arm64"}:   "eng-linux-arm64",
		{"darwin", "arm64"}:  "eng-darwin-arm64",
		{"windows", "amd64"}: "eng-windows-amd64.exe",
	}
	for platform, want := range cases {
		got, err := AssetName(platform[0], platform[1])
		if err != nil {
			t.Fatalf("AssetName(%s/%s): %v", platform[0], platform[1], err)
		}
		if got != want {
			t.Fatalf("AssetName(%s/%s) = %q, want %q", platform[0], platform[1], got, want)
		}
	}
}

func TestAssetNameRejectsUnpublishedPairs(t *testing.T) {
	for _, platform := range [][2]string{{"darwin", "amd64"}, {"windows", "arm64"}, {"freebsd", "amd64"}} {
		if _, err := AssetName(platform[0], platform[1]); err == nil {
			t.Fatalf("AssetName(%s/%s) should fail: no such release asset", platform[0], platform[1])
		}
	}
}

func TestValidateRepoAcceptsSlugOnly(t *testing.T) {
	for _, bad := range []string{"", "owner", "a/b/c", "https://github.com/a/b", "a/b c", "../b", "a/..", "a/b?x=1", "a/b:123"} {
		if err := ValidateRepo(bad); err == nil {
			t.Fatalf("ValidateRepo(%q) should fail", bad)
		}
	}
	if err := ValidateRepo(DefaultRepo); err != nil {
		t.Fatalf("ValidateRepo(default): %v", err)
	}
	for _, good := range []string{"owner/repo.js", "my-org/my_repo", "A1/b2-c3"} {
		if err := ValidateRepo(good); err != nil {
			t.Fatalf("ValidateRepo(%q): %v", good, err)
		}
	}
}

func TestNormalizeTagAndURLs(t *testing.T) {
	tag, err := NormalizeTag("1.4.0")
	if err != nil || tag != "v1.4.0" {
		t.Fatalf("NormalizeTag(1.4.0) = %q, %v", tag, err)
	}
	if tag, err := NormalizeTag("v1.4.0"); err != nil || tag != "v1.4.0" {
		t.Fatalf("NormalizeTag(v1.4.0) = %q, %v", tag, err)
	}
	for _, bad := range []string{"latest", "dev", "", "abc"} {
		if _, err := NormalizeTag(bad); err == nil {
			t.Fatalf("NormalizeTag(%q) should fail", bad)
		}
	}
	if got := AssetURL("O/R", "v1.0.0", "eng-linux-amd64"); got != "https://github.com/O/R/releases/download/v1.0.0/eng-linux-amd64" {
		t.Fatalf("AssetURL = %q", got)
	}
	if got := ChecksumsURL("O/R", "v1.0.0"); got != "https://github.com/O/R/releases/download/v1.0.0/checksums.txt" {
		t.Fatalf("ChecksumsURL = %q", got)
	}
	if got := LatestReleaseAPIURL("O/R"); got != "https://api.github.com/repos/O/R/releases/latest" {
		t.Fatalf("LatestReleaseAPIURL = %q", got)
	}
}

func TestParseLatestTag(t *testing.T) {
	tag, err := ParseLatestTag([]byte(`{"tag_name":"v1.4.0","other":1}`))
	if err != nil || tag != "v1.4.0" {
		t.Fatalf("ParseLatestTag = %q, %v", tag, err)
	}
	for _, bad := range []string{`{}`, `{"tag_name":""}`, `not json`} {
		if _, err := ParseLatestTag([]byte(bad)); err == nil {
			t.Fatalf("ParseLatestTag(%q) should fail", bad)
		}
	}
}

func TestParseChecksumsRoundTrip(t *testing.T) {
	sum := sha256.Sum256([]byte("fake-binary"))
	digest := hex.EncodeToString(sum[:])
	text := digest + "  eng-linux-amd64\n" + digest + " *eng-windows-amd64.exe\r\n"
	parsed, err := ParseChecksums(text)
	if err != nil {
		t.Fatalf("ParseChecksums: %v", err)
	}
	if parsed["eng-linux-amd64"] != digest || parsed["eng-windows-amd64.exe"] != digest {
		t.Fatalf("ParseChecksums lost entries: %v", parsed)
	}
	for _, bad := range []string{"", "oops", digest + "  a/b", "zz  name", digest + "  "} {
		if _, err := ParseChecksums(bad); err == nil {
			t.Fatalf("ParseChecksums(%q) should fail", bad)
		}
	}
}

func TestVerifySHA256(t *testing.T) {
	data := []byte("eng binary bytes")
	sum := sha256.Sum256(data)
	if err := VerifySHA256(data, hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("VerifySHA256(valid): %v", err)
	}
	if err := VerifySHA256(data, hex.EncodeToString(sum[:])+"00"); err == nil {
		t.Fatal("VerifySHA256 should fail on tampered digest")
	}
	if err := VerifySHA256(append(data, 'x'), hex.EncodeToString(sum[:])); err == nil {
		t.Fatal("VerifySHA256 should fail on tampered bytes")
	}
}

func TestNeedsUpdateNeverDowngrades(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"dev", "v9.9.9", true},
		{"1.3.0", "v1.4.0", true},
		{"v1.4.0", "v1.4.0", false},
		{"2.0.0", "v1.4.0", false},
		{"1.4", "v1.4.0", false},
	}
	for _, c := range cases {
		got, err := NeedsUpdate(c.current, c.latest)
		if err != nil {
			t.Fatalf("NeedsUpdate(%q,%q): %v", c.current, c.latest, err)
		}
		if got != c.want {
			t.Fatalf("NeedsUpdate(%q,%q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
	if _, err := NeedsUpdate("1.4.0", "latest"); err == nil {
		t.Fatal("NeedsUpdate with a non-version tag should fail")
	}
	if !strings.Contains(StripTag("v1.4.0"), "1.4.0") {
		t.Fatal("StripTag should drop the v prefix")
	}
}
