package domain

import "strings"

// DatabaseProfile is a curated persistence option selected by policy.
type DatabaseProfile struct {
	ID          string `json:"id"`
	Engine      string `json:"engine,omitempty"`
	Managed     bool   `json:"managed,omitempty"`
	Description string `json:"description,omitempty"`
}

// Validate checks structural presence of a database profile record.
func (d DatabaseProfile) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return Validation("database profile id is required")
	}
	return nil
}
