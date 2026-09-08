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
when offline operation is required; otherwise `stardrive-public-web` wins
by id order. A recipe with `database_policy.shared_backend` always gains
the shared `api` backend component (hono-api), even when the intent did not
name the api surface.

Compatibility is checked, not assumed: recipe surface coverage, allowed
surface composition over the full component set, capability coverage by the
union of chosen providers, database policy versus catalog profiles, tech
compatibility between components, and explicit must-use/must-not-use
constraints. Every failure is a typed `composition` domain error.

## Destination rules

Conventional defaults: `services/api`, `apps/admin`, `apps/mobile`,
`apps/web`, `apps/intake`; unknown surfaces fall back to `apps/<surface>`.
Explicit overrides ride on top of defaults (`ComposeWithDestinations`,
used by `eng plan` only through defaults). Rejected with typed errors:
absolute paths, `..` traversal (rejected, never sanitized), empty or
non-clean paths, duplicate destinations and nested overlaps.
