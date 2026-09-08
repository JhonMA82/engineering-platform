# Add a recipe

A recipe is a candidate architecture profile: what a family of projects
provides, what it composes, and which foundations serve it. It is a routing
candidate, never a user input — Pi produces intents, the resolver selects
recipes.

## Contract

Write `catalog/recipes/<ID>-<slug>.json`:

- `id` (GP-NN), `version`, `status` (`stable` for proven families, `active`
  for selectable-but-young ones like GP-07; anything else is invisible to
  candidate generation).
- `provides.surfaces` / `provides.capabilities`: the architectural coverage
  used by hard constraints. Claim only what the primaries actually serve.
- `primary_boilerplates`: foundations the composer prefers per surface, in
  order. Every id must exist in `catalog/boilerplates/`.
- `allowed_surface_composition`: combos of surfaces this family may serve
  together. Matching is subset semantics: some combo must contain every
  required surface. Omit only when any combination is acceptable.
- `database_policy`: `default_profile` plus `allowed_profiles` (must name
  profiles in `catalog/database-profiles/`, default included) and
  `shared_backend` (true adds the shared `api` component automatically, even
  when the intent did not name the api surface).
- `tech_tags`: stack signals for must-use matching. A tag nobody asks for is
  harmless; a tag that over-matches (e.g. `react` also matching
  `react-native` via substring) must be checked against existing fixtures.
- `quality_gates`: gates the materialized project must pass.

Never add a `project_types` field: matching derives from provides plus
composition, not from user-supplied categories.

## When a new recipe is justified

A new recipe is needed when an intent family is covered by no existing
provides/composition combination — e.g. GP-07 exists although GP-06 already
serves public-web stacks, because a single seed-fork SaaS foundation is a
different foundation fit from a composed backoffice (see
`docs/decisions/migration-notes.md`). If an existing recipe already covers
the family, fold the change into its primaries or combos instead.

## Validate

```bash
/tmp/eng catalog validate
go test ./internal/resolver/ ./internal/composer/ ./internal/catalog/
```

Add routing fixtures proving the recipe wins its family and loses its
neighbors (especially the closest existing recipe), plus one composition
fixture per new surface the recipe introduces. All prior fixtures must stay
green.
