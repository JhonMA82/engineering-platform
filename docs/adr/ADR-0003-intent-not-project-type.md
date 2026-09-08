# ADR-0003 — Intent, not project_type

Date: 2026-09-07 · Status: accepted

`ProjectIntent` carries no `project_type` field. Recipes are derived
candidates matched on surfaces, capabilities, data/ops fit and explicit
technical constraints — never on a user-supplied category label.

Consequence: `landing` normalizes to the `public-web` surface, `blog`/`PDF`
stay product requirements, and scope (`required_now` vs `planned_later`)
decides what gets built now (a future mobile app must not materialize
mobile foundations today).
