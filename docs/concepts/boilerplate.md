# Boilerplate

A boilerplate (`internal/domain/boilerplate.go`, data in
`catalog/boilerplates/`) is a **foundation, not a feature bundle**: a minimal
architectural base that is optimal to extend. `included_product_features` is
informative ranking metadata (≤3 tie-break points); a missing feature never
disqualifies an architecturally valid foundation — Gentle implements the rest.

Boundary: a boilerplate is selectable for materialization only with
repository, upstream pin, adapter and an allowed
decision/delivery status — checked by `composer.Eligible`, not by the
resolver. Two status axes stay separate: decision (`default`, `alternative`,
`specialized`, …) steers selection scope, delivery (`curated`,
`pilot-ready`, …) records review maturity. Curation evidence lives alongside
the catalog in `catalog/curation/<id>.md`.

Adding a representable foundation is catalog work — entry plus adapter plus
evidence plus tests — and needs no core release. Runbook:
`docs/guides/curate-boilerplate.md`.
