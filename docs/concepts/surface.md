# Surface

A surface (`catalog/surfaces/`, e.g. `public-web`, `web-admin`,
`public-intake`, `api`, `mobile-native`, `desktop`, `tui`) is one
required interface or client of the product. Surfaces are catalog data, never
a closed enum in Go (`domain.SurfaceID` is a nominal string validated by the
catalog).

Boundary: a surface is **architectural scope, not product scope**. A landing
page is product scope over the `public-web` surface; a lobby kiosk is an
anonymous intake flow over `public-intake`. User synonyms resolve through
`catalog/vocabulary/aliases.json` during normalization (`desktop app →
desktop`, `kiosk → public-intake`), so new wording rarely needs a new
surface — add one only when a genuinely new client kind needs its own
foundation (as `desktop` did for Tauri).

A surface with no eligible provider is an honest `catalog-gap`, never a
silent substitution (see the `tui` fixtures).
