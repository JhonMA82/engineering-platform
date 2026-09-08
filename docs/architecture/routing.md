# Routing

`internal/resolver` runs the R1–R10 pipeline as pure functions over the
intent and the catalog index; `Resolve` orchestrates, the other files each
own one phase.

- R1 `ValidateIntent` (`intent.go:Validate`): structural rules only. Unknown
  architectural vocabulary is not invalid — uncovered architecture becomes a
  `catalog-gap` later.
- R2 `Normalize` (`normalizer.go`): alias resolution via
  `catalog/vocabulary/aliases.json`, stable sorting, scope defaulting. No
  technology is selected here.
- R3 `SplitRequirements` (`requirements.go`): product vs architecture
  (required/possible) vs must-use/must-not-use vs preferences. Product
  features can never become hard constraints by construction.
- R4 `Derive` (`derive.go`): visible rules turning intent signals into
  canonical requirements — anonymous access → `anonymous-public-access`,
  admin plus mobile sharing data (or several clients on one api) →
  `shared-backend`, offline/realtime signals → their capabilities, and (Fase
  9) `ops.background_jobs → background-processing`, because a worker
  topology conditions the foundation. Covered by `derive_test.go`.
- R5 `Candidates`: every active recipe, never a project_type assumption.
- R6 `ApplyConstraints` (`constraints.go`): required surfaces, composition
  subset check, required refs, derived requirements, must-use/must-not-use
  tech matching (case-insensitive substring either way).
- R7 `DiagnoseGap` (`gaps.go`): required work covered by NO active recipe
  becomes `catalog-gap` with research criteria; covered-but-eliminated
  becomes `unsupported`.
- R8 `ScoreAll` (`scoring.go`): weights surface 35, arch capabilities 20,
  data 15, ops 10, curation 8, simplicity 7, preference 5, product-feature
  bonus ≤3.
- R9/R10 confidence plus `Unresolved` (`confidence.go`, `questions.go`):
  possible-strength refs always surface neutral dimensions; a close call
  (margin < 12) with open dimensions resolves `ambiguous` instead of
  guessing.

Behavior is specified by `testdata/routing/` (41 scenarios asserted by
`resolver_test.go`): resolved selections with reason assertions, ambiguous
dimensions, catalog-gap, unsupported and invalid cases.
