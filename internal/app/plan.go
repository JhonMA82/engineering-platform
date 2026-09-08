package app

import (
	"encoding/json"
	"fmt"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/composer"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/planner"
)

// PlanProject resolves intentJSON against the catalog and converts the
// decision into a composition and a materialization plan. Only a resolved
// decision yields a plan: ambiguous, catalog-gap, unsupported and invalid
// outcomes return a typed composition error and no plan.
func PlanProject(intentJSON []byte, catalogDir string) (*domain.ArchitectureDecision, *composer.Composition, *planner.MaterializationPlan, error) {
	decision, err := ResolveProject(intentJSON, catalogDir)
	if err != nil {
		return nil, nil, nil, err
	}
	if decision.Status != domain.StatusResolved {
		return decision, nil, nil, domain.Composition(
			fmt.Sprintf("cannot plan: decision status is %q (need resolved)", decision.Status))
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return decision, nil, nil, err
	}
	comp, err := composer.Compose(*decision, cat)
	if err != nil {
		return decision, nil, nil, err
	}
	comp.Project = intentName(intentJSON)
	plan, err := planner.Plan(comp, *decision, cat)
	if err != nil {
		return decision, nil, nil, err
	}
	return decision, &comp, &plan, nil
}

// intentName recovers the project name for the plan. A decision carries no
// name, so it is read back from the intent; an unreadable name falls back
// to empty rather than failing planning.
func intentName(intentJSON []byte) string {
	var intent domain.ProjectIntent
	if err := json.Unmarshal(intentJSON, &intent); err != nil {
		return ""
	}
	return intent.Name
}
