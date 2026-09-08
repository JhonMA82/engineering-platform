# Composition (M2)

## Composer vs resolver boundary

The resolver decides the **recipe family and foundation fit** and persists
an `ArchitectureDecision`. The composer decides **how the required surfaces
are satisfied** with compatible providers. The two responsibilities never mix:

- `internal/resolver` — intent → decision (recipe, reasons, derived
  requirements). No providers, no destinations.
- `internal/composer` — decision + catalog → composition (one pinned
  provider boilerplate per required surface, destinations, database
  profile). No filesystem, network or process access.
- `internal/planner` — composition + decision + catalog → serializable
  `MaterializationPlan`. No timestamps, no randomness.
- `internal/app.PlanProject` — thin service: resolve, require a `resolved`
  decision, compose, plan. It fills `Composition.Project` from the intent
  name because a decision carries no name.

## Provider model

Required surfaces and architectural capabilities are recovered from the
decision itself (winner reasons: `covers required surface X`,
`covers capability X`, `covers derived requirement X`, plus
`compatible with must-use tech X`), never invented. Product-feature
mentions (`includes product feature: ...`) are ignored by construction, so
a missing product feature never adds a component.

A boilerplate is an eligible provider when it is curated
(`decision_status`: curated/default/alternative/specialized/reference),
available (`delivery_status`: stable/curated/pilot-ready/released) and
declares both pin and adapter.

Provider ranking per surface is deterministic: recipe primary boilerplates
first (catalog order), then must-use technology fit, then how many required
capabilities the provider covers, then smallest id. Example: for
`public-web` the offline-capable `tanstack-transactional-pwa` wins only
when offline operation is required; otherwise `stardrive` wins
by id order. A recipe with `database_policy.shared_backend` always gains
the shared `api` backend component (hono-api), even when the intent did not
name the api surface.

Compatibility is checked, not assumed: recipe surface coverage, allowed
surface composition over the full component set, capability coverage by the
union of chosen providers, database policy versus catalog profiles, tech
compatibility between components, and explicit must-use/must-not-use
constraints (legacy untyped values plus typed `target=value` pairs
recovered from decision reasons; database targets are enforced by profile
selection while deployment targets support no hard constraints in v1.0.1 —
never against providers). Every
failure is a typed `composition` domain error.

## Database profile selection

`SelectDatabaseProfileFor` resolves the recipe policy while honoring
recovered database constraints (§2.4, full contract in
`docs/concepts/technical-constraints.md`): must-use selects the allowed
profile identifying the value (id, engine or provider) even when it is not
the default; must-not-use avoids that profile when an alternative exists
and fails with a typed composition error when every allowed profile is
forbidden; prefer selects a matching allowed profile or falls back to the default
with the deviation explained in `Composition.DatabaseNote`; avoid picks
any other allowed profile. Preferences never fail selection. Profiles
declare `engine`, `provider` and `supports` (`postgresql-managed`,
`sqlite-local`); Turso is deliberately absent — contract first, profiles
later.

## Destination rules

Conventional defaults: `services/api`, `apps/admin`, `apps/mobile`,
`apps/web`, `apps/intake`; unknown surfaces fall back to `apps/<surface>`.
Explicit overrides ride on top of defaults (`ComposeWithDestinations`,
used by `eng plan` only through defaults). Rejected with typed errors:
absolute paths, `..` traversal (rejected, never sanitized), empty or
non-clean paths, duplicate destinations and nested overlaps.
