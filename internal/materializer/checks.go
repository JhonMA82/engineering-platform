package materializer

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
)

// Verify runs the post-materialize checks over a staged (pre-commit) or
// final project directory: every planned destination exists, the manifest
// matches the plan fingerprint component by component, and no destinations
// collide. It returns the first inconsistency as a typed error.
func Verify(projectDir string, plan planner.MaterializationPlan, manifest project.Manifest) error {
	if manifest.PlanFingerprint != plan.Fingerprint {
		return domain.Materialization(fmt.Sprintf(
			"manifest fingerprint %q does not match plan fingerprint %q",
			shortOut(manifest.PlanFingerprint), shortOut(plan.Fingerprint)))
	}
	if len(manifest.Components) != len(plan.Components) {
		return domain.Materialization(fmt.Sprintf(
			"manifest tracks %d components but the plan declares %d",
			len(manifest.Components), len(plan.Components)))
	}
	planned := map[string]planner.PlanComponent{}
	dests := make([]string, 0, len(plan.Components))
	for _, c := range plan.Components {
		key := c.Destination + "\x00" + c.Surface
		planned[key] = c
		dests = append(dests, c.Destination)
	}
	for _, m := range manifest.Components {
		key := m.Destination + "\x00" + m.Surface
		want, ok := planned[key]
		if !ok {
			return domain.Materialization(fmt.Sprintf(
				"manifest component %s@%s is not in the plan", m.Destination, m.Surface))
		}
		if m.Boilerplate != want.Boilerplate || m.Pin != want.Pin {
			return domain.Materialization(fmt.Sprintf(
				"manifest component %s pins %s@%s but the plan pins %s@%s",
				m.Destination, m.Boilerplate, m.Pin, want.Boilerplate, want.Pin))
		}
		if want.Materialization.Strategy != "" && m.Strategy != want.Materialization.Strategy {
			return domain.Materialization(fmt.Sprintf(
				"manifest component %s records strategy %q but the plan declares %q",
				m.Destination, m.Strategy, want.Materialization.Strategy))
		}
		if want.Materialization.Profile != "" && m.Profile != want.Materialization.Profile {
			return domain.Materialization(fmt.Sprintf(
				"manifest component %s records profile %q but the plan declares %q",
				m.Destination, m.Profile, want.Materialization.Profile))
		}
		if want.Materialization.AdapterFingerprint != "" && m.AdapterFingerprint != want.Materialization.AdapterFingerprint {
			return domain.Materialization(fmt.Sprintf(
				"manifest component %s adapter fingerprint does not match the plan",
				m.Destination))
		}
	}
	if err := CheckCollisions(dests); err != nil {
		return err
	}
	for _, dest := range dests {
		if err := ValidateDestination(dest); err != nil {
			return err
		}
		st, err := os.Stat(projectSubdir(projectDir, dest))
		if err != nil || !st.IsDir() {
			return domain.Materialization(fmt.Sprintf("destination %q is missing after materialization", dest))
		}
	}
	return nil
}
