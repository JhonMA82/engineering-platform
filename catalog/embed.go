// Package catalogdata embeds the base catalog tree into the eng binary so
// a globally installed eng works from any directory: the embedded catalog
// is the normal base source, and --catalog-dir stays an optional overlay.
// Only data files are embedded (never curation/ evidence docs).
package catalogdata

import "embed"

// FS carries metadata.json plus every data subdirectory of the base catalog.
//
//go:embed metadata.json
//go:embed boilerplates/*.json
//go:embed capabilities/*.json
//go:embed database-profiles/*.json
//go:embed recipes/*.json
//go:embed surfaces/*.json
//go:embed vocabulary/*.json
var FS embed.FS
