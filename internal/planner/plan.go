package planner

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/foundationconfig"
)

// PlanComponent is one materialization unit inside the plan.
// Materialization records the resolved per-component configuration:
// strategy (copy or generate), logical generated name, resolved profile,
// non-runtime arguments and the adapter fingerprint. Components written
// by schema v1 plans carry an empty Materialization and read as copy.
type PlanComponent struct {
	Boilerplate     string                                 `json:"boilerplate"`
	Pin             string                                 `json:"pin"`
	Destination     string                                 `json:"destination"`
	Surface         string                                 `json:"surface"`
	Materialization foundationconfig.MaterializationConfig `json:"materialization,omitempty"`
}

// EffectiveStrategy reports the component strategy, defaulting legacy v1
// components without materialization metadata to copy.
func (c PlanComponent) EffectiveStrategy() string {
	if c.Materialization.Strategy == "" {
		return domain.StrategyCopy
	}
	return c.Materialization.Strategy
}

// MaterializationPlan is the serializable pre-write document consumed by
// the future materializer. JSON tags are stable.
type MaterializationPlan struct {
	SchemaVersion   int             `json:"schema_version"`
	Project         string          `json:"project"`
	Recipe          string          `json:"recipe"`
	RecipeVersion   string          `json:"recipe_version"`
	Components      []PlanComponent `json:"components"`
	DatabaseProfile string          `json:"database_profile"`
	Operations      []string        `json:"operations"`
	Setup           []string        `json:"setup"`
	Checks          []string        `json:"checks"`
	Fingerprint     string          `json:"fingerprint"`
}

// FingerprintPlan binds the plan to the intent fingerprint and the catalog
// version: sha256 hex over intent_fingerprint, catalog_version, recipe,
// recipe version, database profile and every component (surface,
// boilerplate, pin, destination, strategy, logical name, resolved profile,
// resolved arguments, adapter fingerprint), NUL-separated with components
// ordered by surface. Changing a profile, an argument or the adapter
// contract therefore changes the fingerprint.
func FingerprintPlan(intentFingerprint, catalogVersion string, comp composer.Composition, mats map[domain.SurfaceID]foundationconfig.MaterializationConfig) string {
	ordered := append([]composer.Component{}, comp.Components...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Surface < ordered[j].Surface })
	h := sha256.New()
	write := func(s string) {
		h.Write([]byte(s))
		h.Write([]byte{0})
	}
	write(intentFingerprint)
	write(catalogVersion)
	write(comp.Recipe)
	write(comp.RecipeVersion)
	write(comp.DatabaseProfile)
	for _, c := range ordered {
		write(string(c.Surface))
		write(c.Boilerplate)
		write(c.Pin)
		write(c.Destination)
		m := mats[c.Surface]
		strategy := m.Strategy
		if strategy == "" {
			strategy = domain.StrategyCopy
		}
		write(strategy)
		write(m.Name)
		write(m.Profile)
		for _, arg := range m.Arguments {
			write(arg)
		}
		write("\x00args")
		write(m.AdapterFingerprint)
	}
	return hex.EncodeToString(h.Sum(nil))
}
