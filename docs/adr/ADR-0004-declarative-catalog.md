# ADR-0004 — Declarative catalog

Date: 2026-09-07 · Status: accepted

Surfaces, capabilities, recipes, boilerplates and vocabulary aliases live in
`catalog/` as JSON data validated by `internal/catalog`. `SurfaceID` and
`CapabilityID` are nominal string types checked against the catalog, not
closed Go enums.

Consequence: registering a future surface (e.g. `tui`) needs no core
release; the loader merges overlays over the base catalog and the validator
rejects dangling refs, missing pins/adapters and bad `min_core_version`.
