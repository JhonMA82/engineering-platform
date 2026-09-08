package app

import (
	"encoding/json"
	"fmt"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/planner"
	"github.com/jhonma82/engineering-platform/internal/project"
	"github.com/jhonma82/engineering-platform/internal/version"
)

// CoreVersion delegates to the canonical release line in internal/version
// so provenance stamps whatever the binary was built with (see package
// version; release builds stamp it via ldflags).
var CoreVersion = version.CoreVersion

// MaterializeProject loads a plan document and materializes it into
// outputDir without re-resolving architecture. Intent and decision
// documents are optional: when provided they are persisted as the
// .engineering copies and enrich the agent-context files; when absent the
// remaining artifacts are still generated from plan plus catalog.
func MaterializeProject(planJSON, intentJSON, decisionJSON []byte, catalogDir, outputDir string) (*project.Manifest, error) {
	var plan planner.MaterializationPlan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return nil, domain.Validation(fmt.Sprintf("parse plan: %v", err))
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	res, err := materializer.Materialize(materializer.Request{
		Plan:         plan,
		Catalog:      cat,
		OutputDir:    outputDir,
		IntentJSON:   intentJSON,
		DecisionJSON: decisionJSON,
		CoreVersion:  CoreVersion,
	})
	if err != nil {
		return nil, err
	}
	return &res.Manifest, nil
}
