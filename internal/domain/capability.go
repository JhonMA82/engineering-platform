package domain

import "strings"

// Capability is a catalog-registered architectural property.
type Capability struct {
	ID          CapabilityID `json:"id"`
	Description string       `json:"description,omitempty"`
}

// Validate checks structural presence of a capability record.
func (c Capability) Validate() error {
	if strings.TrimSpace(string(c.ID)) == "" {
		return Validation("capability id is required")
	}
	return nil
}
