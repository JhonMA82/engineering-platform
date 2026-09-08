// Package version is the single canonical source for the eng binary
// release line (§1.3 of the v1.0 release-hardening plan).
//
// CoreVersion defaults to "dev" for working-tree builds. Release builds stamp
// the real line via ldflags (see the Makefile release target):
//
//	go build -ldflags "-X github.com/jhonma82/engineering-platform/internal/version.CoreVersion=1.0.0 -X github.com/jhonma82/engineering-platform/internal/version.Commit=<sha> -X github.com/jhonma82/engineering-platform/internal/version.BuildDate=<date>" ./cmd/eng
//
// Commit and BuildDate report as "unknown" when unstamped. Every other
// version reference in the tree (app provenance, cli version output)
// delegates to this package; no second CoreVersion may exist.
//
// Only the standard library is used: semver comparison is a small local
// implementation, not a third-party dependency.
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// CoreVersion is the eng binary release line. Stamp at release time with
// -ldflags -X (see above); working-tree builds report "dev".
var CoreVersion = "dev"

// Commit is the stamped VCS revision, or "unknown" when unstamped.
var Commit = "unknown"

// BuildDate is the stamped build date, or "unknown" when unstamped.
var BuildDate = "unknown"

// ParseVersion parses semantic versions of the forms "1.0.0", "v1.0.0" and
// "1.0" (short form: missing parts default to zero). A leading "v"/"V" and
// any pre-release/build suffix ("-rc.1", "+meta") are stripped before
// comparison, so release lines compare by their numeric core. Anything else
// (notably the "dev" working-tree marker) is an error.
func ParseVersion(s string) ([3]int, error) {
	var zero [3]int
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "v")
	t = strings.TrimPrefix(t, "V")
	if i := strings.IndexAny(t, "-+"); i >= 0 {
		t = t[:i]
	}
	parts := strings.Split(t, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return zero, fmt.Errorf("invalid version %q (want X.Y.Z, vX.Y.Z or X.Y)", s)
	}
	var out [3]int
	for i, p := range parts {
		if p == "" {
			return zero, fmt.Errorf("invalid version %q (want X.Y.Z, vX.Y.Z or X.Y)", s)
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return zero, fmt.Errorf("invalid version %q (want X.Y.Z, vX.Y.Z or X.Y)", s)
		}
		out[i] = n
	}
	return out, nil
}

// CompareVersions compares two semantic versions: -1 when a < b, 0 when
// equal, +1 when a > b. Either side may use the v-prefix or short form.
func CompareVersions(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if va[i] != vb[i] {
			if va[i] < vb[i] {
				return -1, nil
			}
			return 1, nil
		}
	}
	return 0, nil
}

// CheckCompatibility enforces the core/catalog contract: a catalog declaring
// minCore (and optionally maxCore) loads only when core satisfies both
// bounds. Failures use the exact release-hardening error vocabulary:
//
//	catalog requires core >= 1.1.0, running core is 1.0.0
//	catalog requires core <= 2.0.0, running core is 2.1.0
//
// The "dev" working-tree marker bypasses both gates: unreleased builds must
// keep loading the in-tree catalog under development. Release binaries always
// carry a stamped numeric line, so the gate is live exactly where it
// matters. An empty minCore/maxCore bound is unset and skipped.
func CheckCompatibility(core, minCore, maxCore string) error {
	if strings.TrimSpace(core) == "dev" {
		return nil
	}
	if strings.TrimSpace(minCore) != "" {
		cmp, err := CompareVersions(core, minCore)
		if err != nil {
			return fmt.Errorf("catalog declares invalid min_core_version %q: %v", minCore, err)
		}
		if cmp < 0 {
			return fmt.Errorf("catalog requires core >= %s, running core is %s", minCore, core)
		}
	}
	if strings.TrimSpace(maxCore) != "" {
		cmp, err := CompareVersions(core, maxCore)
		if err != nil {
			return fmt.Errorf("catalog declares invalid max_core_version %q: %v", maxCore, err)
		}
		if cmp > 0 {
			return fmt.Errorf("catalog requires core <= %s, running core is %s", maxCore, core)
		}
	}
	return nil
}
