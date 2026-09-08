# Project intent

`ProjectIntent` (`internal/domain/intent.go`, schema
`schemas/project-intent.schema.json`) is what the product needs, stated
without deciding how it is built. It is the only input the resolver accepts.

Boundary: there is intentionally **no `project_type` field**. Routing derives
the recipe from surfaces, requirements and constraints — a label supplied by
the user must never pre-select the architecture.

- `surfaces[]` with per-surface `scope` (`required_now` vs `planned_later`
  vs `explicitly_excluded`): only `required_now` surfaces constrain routing;
  a future mobile app favors a reusable API but installs no mobile foundation
  (fixture `mobile-future`).
- `architecture_requirements[]` with `strength` (`required` vs `possible`):
  required refs condition the foundation; possible refs only surface
  discriminating dimensions for questions.
- `product_requirements[]`: what the product must do. Never eligibility, only
  a small scoring tie-break — a missing PDF feature never disqualifies a
  foundation.
- `technical_constraints[]` (`must-use` / `must-not-use` …): explicit user
  decisions, preserved verbatim and enforced as hard constraints.
- `preferences[]` (`prefer` / `avoid`): ranking influence only.
