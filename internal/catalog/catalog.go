// Package catalog loads and validates the declarative JSON catalog.
// Only stdlib encoding/json is used for parsing.
package catalog

import "github.com/jhonma82/engineering-platform/internal/domain"

// Catalog is the loaded declarative knowledge base.
type Catalog struct {
	CatalogVersion   string
	MinCoreVersion   string
	MaxCoreVersion   string
	SchemaVersion    int
	Recipes          []domain.Recipe
	Boilerplates     []domain.Boilerplate
	Surfaces         []domain.Surface
	Capabilities     []domain.Capability
	DatabaseProfiles []domain.DatabaseProfile
	Aliases          map[string]string
	AliasNotes       map[string]string
}
