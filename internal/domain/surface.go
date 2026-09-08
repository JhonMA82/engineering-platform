package domain

import "strings"

// Surface is a catalog-registered interface/client kind.
type Surface struct {
	ID          SurfaceID `json:"id"`
	Description string    `json:"description,omitempty"`
}

// Validate checks structural presence of a surface record.
func (s Surface) Validate() error {
	if strings.TrimSpace(string(s.ID)) == "" {
		return Validation("surface id is required")
	}
	return nil
}
