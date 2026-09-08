package composer

import (
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/domain"
)

// DefaultDestinations maps each known surface to its conventional
// repository destination. Unknown surfaces fall back to apps/<surface>.
var DefaultDestinations = map[domain.SurfaceID]string{
	"api":           "services/api",
	"web-admin":     "apps/admin",
	"mobile-native": "apps/mobile",
	"public-web":    "apps/web",
	"public-intake": "apps/intake",
}

// DefaultDestination returns the conventional destination for a surface.
func DefaultDestination(surface domain.SurfaceID) string {
	if d, ok := DefaultDestinations[surface]; ok {
		return d
	}
	return "apps/" + string(surface)
}

// AssignDestinations resolves one destination per surface: explicit
// overrides win, otherwise the conventional default applies. Every result
// is validated: absolute paths, path traversal, empty paths, duplicate
// destinations and nested overlaps are rejected with typed errors.
func AssignDestinations(surfaces []domain.SurfaceID, overrides map[domain.SurfaceID]string) (map[domain.SurfaceID]string, error) {
	out := make(map[domain.SurfaceID]string, len(surfaces))
	for _, s := range surfaces {
		dest := DefaultDestination(s)
		if o, ok := overrides[s]; ok {
			dest = o
		}
		cleaned, err := ValidateDestination(s, dest)
		if err != nil {
			return nil, err
		}
		out[s] = cleaned
	}
	ordered := append([]domain.SurfaceID{}, surfaces...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	seen := map[string]domain.SurfaceID{}
	for _, s := range ordered {
		d := out[s]
		if prev, dup := seen[d]; dup {
			return nil, domain.Composition(
				"destination collision: surfaces " + string(prev) + " and " + string(s) +
					" both target " + d)
		}
		seen[d] = s
	}
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			a, b := out[ordered[i]], out[ordered[j]]
			if isNested(a, b) {
				return nil, domain.Composition(
					"nested destinations: " + string(ordered[i]) + " targets " + a +
						" inside " + string(ordered[j]) + " target " + b)
			}
		}
	}
	return out, nil
}

// ValidateDestination rejects unsafe destinations and returns the cleaned
// relative path.
func ValidateDestination(surface domain.SurfaceID, dest string) (string, error) {
	where := "surface " + string(surface)
	if strings.TrimSpace(dest) == "" {
		return "", domain.Composition("invalid destination for " + where + ": path is empty")
	}
	if filepath.IsAbs(dest) || path.IsAbs(dest) || hasDriveLetter(dest) {
		return "", domain.Composition(
			"invalid destination " + dest + " for " + where + ": absolute paths are rejected")
	}
	slashed := strings.ReplaceAll(dest, "\\", "/")
	for _, seg := range strings.Split(slashed, "/") {
		switch seg {
		case "..":
			return "", domain.Composition(
				"invalid destination " + dest + " for " + where + ": path traversal is rejected")
		case "", ".":
			return "", domain.Composition(
				"invalid destination " + dest + " for " + where + ": path is not a clean relative directory")
		}
	}
	return path.Clean(slashed), nil
}

func hasDriveLetter(dest string) bool {
	if len(dest) < 2 || dest[1] != ':' {
		return false
	}
	c := dest[0]
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func isNested(a, b string) bool {
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
