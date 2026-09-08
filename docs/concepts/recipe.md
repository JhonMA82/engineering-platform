# Recipe

A recipe (`internal/domain/recipe.go`, data in `catalog/recipes/`) is a
candidate architecture profile: GP-01 public web, GP-02 admin, GP-03
python/data, GP-04 mobile, GP-05 desktop, GP-06 multi-app, GP-07 commercial
SaaS.

Boundary: recipes are **results, not inputs**. Pi never writes
`selected_recipe`; the resolver picks from `Active()` recipes
(`stable`/`active`) using provides plus allowed composition. A recipe carries
no `project_types` matching — if two recipes cover the same surfaces, the
finer foundation fit wins on scoring, not on labels.

A recipe declares what families it serves (`provides`), which foundations
serve them (`primary_boilerplates`), which surface sets compose
(`allowed_surface_composition`, subset semantics) and the data policy
(`database_policy`, including whether the family shares one backend).
Runbook: `docs/guides/add-recipe.md`.
