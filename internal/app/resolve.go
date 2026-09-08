// Package app is the thin service boundary over the resolver.
package app

import (
	"encoding/json"
	"fmt"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/resolver"
)

// ResolveProject parses intentJSON and resolves it against the catalog.
// An empty catalogDir selects the in-repo base catalog. Invalid intents
// yield a decision with status invalid, not an error; only unreadable
// input or catalog failures return errors.
func ResolveProject(intentJSON []byte, catalogDir string) (*domain.ArchitectureDecision, error) {
	var intent domain.ProjectIntent
	if err := json.Unmarshal(intentJSON, &intent); err != nil {
		return nil, domain.Validation(fmt.Sprintf("parse intent: %v", err))
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return nil, err
	}
	decision := resolver.Resolve(intent, cat)
	return &decision, nil
}
