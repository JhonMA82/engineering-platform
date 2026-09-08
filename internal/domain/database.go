package domain

import "strings"

// techValueMatch is the shared technology-matching rule: a value matches
// when it equals or contains (either way) any of the catalog signals,
// case-insensitively. Matching stays catalog-driven: the signals come from
// catalog JSON, never from hardcoded brand names in Go.
func techValueMatch(signals []string, value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return false
	}
	for _, s := range signals {
		tag := strings.ToLower(strings.TrimSpace(s))
		if tag == "" {
			continue
		}
		if tag == v || strings.Contains(tag, v) || strings.Contains(v, tag) {
			return true
		}
	}
	return false
}

// DatabaseProfile is a curated persistence option selected by policy.
// Engine names the storage engine (postgres, sqlite, libsql); Provider
// names the backing service or distribution; Supports lists the catalog
// capabilities the profile offers (shared-backend, local, offline, sync).
// Matching for database-target constraints runs against id, engine and
// provider — all catalog data, never hardcoded brands in Go (see §2.3).
type DatabaseProfile struct {
	ID          string   `json:"id"`
	Engine      string   `json:"engine,omitempty"`
	Provider    string   `json:"provider,omitempty"`
	Supports    []string `json:"supports,omitempty"`
	Managed     bool     `json:"managed,omitempty"`
	Description string   `json:"description,omitempty"`
}

// MatchesTechnology reports whether a database-target constraint value
// identifies this profile via its id, engine or provider, using the same
// case-insensitive either-way substring rule as technology matching.
func (d DatabaseProfile) MatchesTechnology(value string) bool {
	return techValueMatch([]string{d.ID, d.Engine, d.Provider}, value)
}

// Validate checks structural presence of a database profile record.
func (d DatabaseProfile) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return Validation("database profile id is required")
	}
	return nil
}
