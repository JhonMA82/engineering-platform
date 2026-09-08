# Technical constraints

`TechnicalConstraint` (`internal/domain/intent.go`) is an explicit user
technology decision: `{target, kind, value}`.

```json
{ "target": "framework", "kind": "must-use", "value": "tanstack" }
{ "target": "database", "kind": "must-use", "value": "turso" }
{ "target": "language", "kind": "prefer", "value": "python" }
```

## Targets

The taxonomy stays small: `framework | language | runtime | database |
deployment | provider`. Unknown targets are rejected by intent validation
with the full list. `deployment` is a known target but supports no hard
constraints in v1.0.1: the catalog curates no deployment metadata, so
`must-* deployment=*` is rejected at validation with
`unsupported technical constraint target: deployment` instead of being
accepted and silently ignored. `prefer | avoid deployment=*` remain valid
as recorded, non-routing preferences. `target` is what lets the resolver tell a database
from a framework: `must-use: turso` alone cannot.

## Kinds

`must-use | must-not-use` are hard constraints: they can eliminate
candidates. `prefer | avoid` affect the score only and never eligibility —
including when written inside `technical_constraints` (they behave exactly
like `preferences[]`). The legacy `must-run | must-support | must-share`
kinds remain accepted but are not evaluated.

## How matching works (no hardcoded brands)

Matching runs against catalog data only; the Go core contains no
`tanstack` / `turso` / `python` literals:

- `framework | language | runtime`: the recipe primary boilerplates'
  `technology` metadata for that target, with recipe `tech_tags` as a
  documented fallback.
- `provider`: boilerplate id, adapter name and tech tags.
- `database`: curated database profiles by id, engine or provider
  (`DatabaseProfile.MatchesTechnology`); resolved at composition time
  against the recipe policy.
- `deployment`: no catalog metadata is curated yet, so hard deployment
  constraints are rejected at validation (`unsupported technical constraint
  target: deployment`, v1.0.1); `prefer`/`avoid` deployment values are only
  recorded in decision reasons, never evaluated (§2.3). Deployment profiles
  (`edge`, `serverless`, `container`, ...) are future work.

## Database constraints (§2.4)

The composer selects the database profile honoring recovered constraints:

- `must-use database=X` selects the allowed profile identifying `X`, even
  when it is not the recipe default. No allowed profile identifies `X`:
  typed composition error.
- `must-use database=X` with **no curated profile anywhere** never reaches
  composition: the resolver reports `catalog-gap` with
  `MissingArchitecture {kind: database-profile, value: X}`.
- `prefer database=X` with no matching profile falls back to the default
  and records the deviation in `Composition.DatabaseNote` — a reason, never
  a gap. `prefer`/`avoid` never affect eligibility.

Deliberately not present: Turso. The contract above ships first; real
profiles are curated later.

## Backward compatibility

Intents written before targets carry `{kind, value}` only. A missing
target preserves the pre-H2 rule exactly: the value matches recipe
`tech_tags` (fixtures `must-use-tanstack`, `python-reporting-tool`,
`commercial-saas` and the `unsupported-*` cases pin this behavior). New
intents should always set an explicit target.

## Behavior map

- Resolver R3 (`requirements.go:SplitRequirements`): separates untyped,
  typed, database and preference channels.
- Resolver R6 (`constraints.go:ApplyConstraints`): typed filtering with
  `compatible with must-use <target>=<value>` reasons; legacy strings
  unchanged.
- Resolver R7 (`gaps.go:DiagnoseDatabaseGap`): missing mandatory profile
  → `database-profile` gap.
- Composer (`compatibility.go:SelectDatabaseProfileFor`,
  `surfaces.go:MustUseConstraints`): constraint-driven profile selection
  and provider matching recovered from decision reasons.
