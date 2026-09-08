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
  features can never become hard constraints by construction. Technical
  constraints split into channels: legacy untyped values, explicit-target
  framework/language/runtime/provider constraints, database must-use /
  prefer / must-not-use values, and prefer/avoid constraints (any
  non-database target), which join `preferences` and stay ranking-only.
  Deployment targets have no catalog metadata and are recorded, not
  evaluated.
- R4 `Derive` (`derive.go`): visible rules turning intent signals into
  canonical requirements — anonymous access → `anonymous-public-access`,
  admin plus mobile sharing data (or several clients on one api) →
  `shared-backend`, offline/realtime signals → their capabilities, and (Fase
  9) `ops.background_jobs → background-processing`, because a worker
  topology conditions the foundation. Covered by `derive_test.go`.
- R5 `Candidates`: every active recipe, never a project_type assumption.
- R6 `ApplyConstraints` (`constraints.go`): required surfaces, composition
  subset check, required refs, derived requirements, must-use/must-not-use
  tech matching (case-insensitive substring either way). Untyped values
  match recipe `tech_tags` (pre-H2 rule, pinned by fixtures); typed
  framework/language/runtime/provider constraints match the catalog
  technology metadata for that target (`compatible with must-use
  <target>=<value>`). must-use database values eliminate only recipes
  whose policy offers no matching profile — a value with no curated
  profile anywhere is left for R7. prefer/avoid never eliminate.
- R7 `DiagnoseGap` (`gaps.go`): required work covered by NO active recipe
  becomes `catalog-gap` with research criteria; covered-but-eliminated
  becomes `unsupported`. A must-use database value with no curated profile
  is a `database-profile` gap even when recipes are eligible
  (`DiagnoseDatabaseGap`, checked before winner selection) — see
  `docs/concepts/technical-constraints.md`.
- R8 `ScoreAll` (`scoring.go`): weights surface 35, arch capabilities 20,
  data 15, ops 10, curation 8, simplicity 7, preference 5, product-feature
  bonus ≤3. `prefer`/`avoid` from `preferences[]` and from prefer/avoid
  technical constraints score identically against recipe tech tags.
- R9/R10 confidence plus `Unresolved` (`confidence.go`, `questions.go`):
  possible-strength refs always surface neutral dimensions; a close call
  (margin < 12) with open dimensions resolves `ambiguous` instead of
  guessing.

Behavior is specified by `testdata/routing/` (46 scenarios asserted by
`resolver_test.go`, floor-guarded by `minRoutingFixtures = 40`): resolved
selections with reason assertions, ambiguous dimensions, catalog-gap
(including `database-profile`), unsupported and invalid cases. Typed-constraint
unit coverage lives in `typed_constraints_test.go` (in-memory catalog: TanStack vs Next,
prefer-ranking, database gap); legacy untyped fixtures are untouched and
pin the backward-compat rule.

`open_product_questions[]` on the intent is routing-invisible by
construction: no R-phase reads it (`TestOpenProductQuestionsDoNotAffectRouting`
asserts identical decisions with and without questions). It exists so product
unknowns reach Gentle through the handoff without polluting architecture —
see `docs/architecture/handoff.md`.
