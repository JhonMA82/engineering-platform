package resolver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Normalized is the R2 output: aliases resolved, stable order, pure
// semantic defaults. No technology is selected here.
type Normalized struct {
	Intent           domain.ProjectIntent
	RequiredSurfaces []domain.SurfaceID
	RequiredRefs     []string
	PossibleRefs     []string
}

// canonicalTerm lowercases a term and resolves catalog aliases.
func canonicalTerm(term string, idx catalog.Index) string {
	t := strings.ToLower(strings.TrimSpace(term))
	if c, ok := idx.CanonicalAlias(t); ok {
		return strings.ToLower(c)
	}
	return t
}

// Normalize runs R2 alias resolution plus stable sorting.
func Normalize(intent domain.ProjectIntent, idx catalog.Index) Normalized {
	n := Normalized{Intent: intent}
	for i := range n.Intent.Surfaces {
		s := &n.Intent.Surfaces[i]
		s.Kind = domain.SurfaceID(canonicalTerm(string(s.Kind), idx))
		if s.Scope == "" {
			s.Scope = domain.ScopeRequiredNow
		}
	}
	sort.Slice(n.Intent.Surfaces, func(i, j int) bool {
		a, b := n.Intent.Surfaces[i], n.Intent.Surfaces[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Scope < b.Scope
	})
	for i := range n.Intent.ArchitectureRequirements {
		r := &n.Intent.ArchitectureRequirements[i]
		r.Ref = canonicalTerm(r.Ref, idx)
		if r.Strength == "" {
			r.Strength = "required"
		}
	}
	sort.Slice(n.Intent.ArchitectureRequirements, func(i, j int) bool {
		return n.Intent.ArchitectureRequirements[i].ID < n.Intent.ArchitectureRequirements[j].ID
	})
	sort.Slice(n.Intent.ProductRequirements, func(i, j int) bool {
		return n.Intent.ProductRequirements[i].ID < n.Intent.ProductRequirements[j].ID
	})
	n.RequiredSurfaces = n.Intent.RequiredSurfaces()
	seen := map[string]bool{}
	for _, r := range n.Intent.ArchitectureRequirements {
		if r.Ref == "" || seen[r.Ref+"|"+r.EffectiveStrength()] {
			continue
		}
		seen[r.Ref+"|"+r.EffectiveStrength()] = true
		if r.EffectiveStrength() == "possible" {
			n.PossibleRefs = append(n.PossibleRefs, r.Ref)
		} else {
			n.RequiredRefs = append(n.RequiredRefs, r.Ref)
		}
	}
	sort.Strings(n.RequiredRefs)
	sort.Strings(n.PossibleRefs)
	return n
}

// FingerprintNormalized returns the sha256 of the normalized intent.
func FingerprintNormalized(n Normalized) string {
	raw, _ := json.Marshal(n.Intent)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// FingerprintRaw fingerprints a pre-validation intent.
func FingerprintRaw(intent domain.ProjectIntent) string {
	raw, _ := json.Marshal(intent)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
