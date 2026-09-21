// Package selfupdate holds the pure (no I/O) logic behind eng's remote
// installation and update channel.
//
// The channel is the GitHub release line: every v* tag publishes the
// release-matrix binaries plus checksums.txt (see
// .github/workflows/release.yml). This package only computes names, URLs
// and integrity decisions; all network and filesystem effects live in
// internal/materializer, and the thin CLI surface lives in internal/cli.
//
// Only the standard library is used.
package selfupdate

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/version"
)

// DefaultRepo is the GitHub repository (owner/name) eng installs from.
// Forks override it via ENG_GITHUB_REPO / --repo; validation keeps the
// value a plain owner/name slug so it can never become a URL or a shell
// fragment.
const DefaultRepo = "JhonMA82/engineering-platform"

// ChecksumsFile is the release asset carrying sha256 lines for every
// binary, written by the release workflow with `sha256sum eng-*`.
const ChecksumsFile = "checksums.txt"

// AssetName maps a GOOS/GOARCH pair to the exact release asset published
// by the release matrix. Pairs outside the matrix (notably darwin/amd64
// and windows/arm64) are an explicit error, never a guessed name.
func AssetName(goos, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "linux/amd64":
		return "eng-linux-amd64", nil
	case "linux/arm64":
		return "eng-linux-arm64", nil
	case "darwin/arm64":
		return "eng-darwin-arm64", nil
	case "windows/amd64":
		return "eng-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("unsupported platform %s/%s (releases publish linux/amd64, linux/arm64, darwin/arm64, windows/amd64)", goos, goarch)
	}
}

// ValidateRepo rejects anything that is not a plain "owner/name" slug:
// empty parts, extra segments, whitespace, URL schemes and path escapes
// all fail so the value is safe to interpolate into release URLs.
func ValidateRepo(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("invalid repo %q (want OWNER/NAME)", repo)
	}
	for _, p := range parts {
		if p != strings.TrimSpace(p) || strings.ContainsAny(p, " \t\n\r\\\"'`<>()|&;$:?#@!") || p == "." || p == ".." || strings.HasPrefix(p, ".") || strings.HasPrefix(p, "-") {
			return fmt.Errorf("invalid repo %q (want OWNER/NAME)", repo)
		}
		for _, r := range p {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' {
				return fmt.Errorf("invalid repo %q (want OWNER/NAME)", repo)
			}
		}
	}
	if strings.Contains(repo, "..") {
		return fmt.Errorf("invalid repo %q (want OWNER/NAME)", repo)
	}
	return nil
}

// NormalizeTag accepts "1.2.3" or "v1.2.3" and returns the release tag
// form "v1.2.3". Anything else (notably "latest" or "dev") is an error:
// callers resolve aliases to a concrete tag first.
func NormalizeTag(tag string) (string, error) {
	t := strings.TrimSpace(tag)
	t = strings.TrimPrefix(t, "v")
	t = strings.TrimPrefix(t, "V")
	if _, err := version.ParseVersion(t); err != nil {
		return "", fmt.Errorf("invalid version %q (want X.Y.Z or vX.Y.Z)", tag)
	}
	return "v" + t, nil
}

// StripTag drops the "v" prefix (and any pre-release/build suffix via the
// version parser) so tags compare with version.CompareVersions.
func StripTag(tag string) string {
	t := strings.TrimSpace(tag)
	t = strings.TrimPrefix(t, "v")
	t = strings.TrimPrefix(t, "V")
	return t
}

// ReleaseBaseURL is the download root for one tag:
// https://github.com/<repo>/releases/download/<tag>.
func ReleaseBaseURL(repo, tag string) string {
	return "https://github.com/" + repo + "/releases/download/" + tag
}

// AssetURL is the direct download URL of one release asset.
func AssetURL(repo, tag, asset string) string {
	return ReleaseBaseURL(repo, tag) + "/" + asset
}

// ChecksumsURL is the direct download URL of the release checksums file.
func ChecksumsURL(repo, tag string) string {
	return ReleaseBaseURL(repo, tag) + "/" + ChecksumsFile
}

// LatestReleaseAPIURL is the GitHub API endpoint reporting the newest
// release; its tag_name resolves the "latest" alias.
func LatestReleaseAPIURL(repo string) string {
	return "https://api.github.com/repos/" + repo + "/releases/latest"
}

// ParseLatestTag extracts tag_name from a GitHub releases/latest API
// payload. Unknown fields are ignored; a missing tag is an error.
func ParseLatestTag(body []byte) (string, error) {
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("parse latest release: %v", err)
	}
	if strings.TrimSpace(payload.TagName) == "" {
		return "", fmt.Errorf("parse latest release: missing tag_name")
	}
	return payload.TagName, nil
}

// ParseChecksums parses a sha256sum-format checksums file
// ("<hex>  <name>" per line, tolerating the "*" binary marker and
// CRLF) into name -> lowercase hex digest. Malformed lines fail closed.
func ParseChecksums(text string) (map[string]string, error) {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed checksums line %q (want \"<sha256>  <name>\")", line)
		}
		digest := strings.ToLower(fields[0])
		name := strings.TrimPrefix(fields[1], "*")
		if len(digest) != 64 {
			return nil, fmt.Errorf("malformed checksums line %q (want 64 hex digits)", line)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return nil, fmt.Errorf("malformed checksums line %q (not hex)", line)
		}
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "/\\") {
			return nil, fmt.Errorf("malformed checksums line %q (bad file name)", line)
		}
		out[name] = digest
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("checksums file carries no entries")
	}
	return out, nil
}

// ExpectedChecksum returns the recorded digest for asset, or an error
// when the checksums file does not cover it.
func ExpectedChecksum(checksums map[string]string, asset string) (string, error) {
	digest, ok := checksums[asset]
	if !ok {
		return "", fmt.Errorf("checksums file does not cover %q", asset)
	}
	return digest, nil
}

// VerifySHA256 reports whether data hashes to the expected lowercase hex
// digest, using a constant-time comparison so verification has no
// early-exit oracle.
func VerifySHA256(data []byte, wantHex string) error {
	sum := sha256.Sum256(data)
	want, err := hex.DecodeString(strings.ToLower(strings.TrimSpace(wantHex)))
	if err != nil || len(want) != sha256.Size {
		return fmt.Errorf("checksum mismatch: bad recorded digest")
	}
	if subtle.ConstantTimeCompare(sum[:], want) != 1 {
		return fmt.Errorf("checksum mismatch: downloaded file failed sha256 verification")
	}
	return nil
}

// NeedsUpdate compares the running core line against the target release
// tag: true when the target is newer. The "dev" working-tree marker is
// always stale (any release is newer); equal or newer running lines need
// no update, so self-update never downgrades.
func NeedsUpdate(current, latestTag string) (bool, error) {
	if strings.TrimSpace(current) == "dev" {
		return true, nil
	}
	cmp, err := version.CompareVersions(StripTag(latestTag), StripTag(current))
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}
