package domain

// ResolutionStatus is the resolver output state.
type ResolutionStatus string

const (
	StatusResolved    ResolutionStatus = "resolved"
	StatusAmbiguous   ResolutionStatus = "ambiguous"
	StatusCatalogGap  ResolutionStatus = "catalog-gap"
	StatusUnsupported ResolutionStatus = "unsupported"
	StatusInvalid     ResolutionStatus = "invalid"
)

// SelectedRecipe is the winning candidate, when one exists.
type SelectedRecipe struct {
	Recipe        string `json:"recipe"`
	RecipeVersion string `json:"recipe_version,omitempty"`
	Score         int    `json:"score"`
}

// Candidate records eligibility, score and reasons per recipe.
type Candidate struct {
	Recipe          string   `json:"recipe"`
	Eligible        bool     `json:"eligible"`
	Score           int      `json:"score"`
	PositiveReasons []string `json:"positive_reasons,omitempty"`
	NegativeReasons []string `json:"negative_reasons,omitempty"`
}

// DerivedRequirement is an architectural need inferred by R4 rules.
type DerivedRequirement struct {
	ID     string `json:"id"`
	Source string `json:"source,omitempty"`
}

// UnresolvedDimension is a neutral discriminant for Pi to ask about.
type UnresolvedDimension struct {
	Dimension string   `json:"dimension"`
	Reason    string   `json:"reason,omitempty"`
	Options   []string `json:"options,omitempty"`
}

// MissingPiece describes an architectural gap with curation criteria.
type MissingPiece struct {
	Kind             string   `json:"kind"`
	Ref              string   `json:"ref"`
	ResearchCriteria []string `json:"research_criteria,omitempty"`
}

// Confidence is derived from score, margin and unresolved dimensions.
type Confidence struct {
	Level  string `json:"level"`
	Margin int    `json:"margin"`
}

// ArchitectureDecision is the persistible resolver output.
type ArchitectureDecision struct {
	SchemaVersion        int                   `json:"schema_version"`
	Status               ResolutionStatus      `json:"status"`
	IntentFingerprint    string                `json:"intent_fingerprint,omitempty"`
	Selected             *SelectedRecipe       `json:"selected,omitempty"`
	Confidence           Confidence            `json:"confidence,omitempty"`
	Reasons              []string              `json:"reasons,omitempty"`
	Candidates           []Candidate           `json:"candidates,omitempty"`
	DerivedRequirements  []DerivedRequirement  `json:"derived_requirements,omitempty"`
	UnresolvedDimensions []UnresolvedDimension `json:"unresolved_dimensions,omitempty"`
	MissingArchitecture  []MissingPiece        `json:"missing_architecture,omitempty"`
}
