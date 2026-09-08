# Evolve a materialized project (Fase 10)

Four commands grow a project after bootstrap without losing provenance.
Architecture evolution (`surface add`, `extend`) reuses the whole spine —
resolve → compose → plan → delta materialize — against the stored intent;
`add` records product requirements without touching architecture; `update`
is a report, never an auto-apply.

Starting point: a materialized project (see `docs/guides/new-project.md`),
for example a GP-06 project with `web-admin` + `api` required now:

```bash
go build -o /tmp/eng ./cmd/eng
```

## 1. `eng surface add` — architecture evolution

Adds one new required surface to the stored intent, re-resolves, composes
the full evolved project and materializes **only the delta** (existing
directories are never rewritten):

```bash
/tmp/eng surface add --project /tmp/demo-proj --surface public-intake --catalog-dir /tmp/fixture-catalog
/tmp/eng doctor --project /tmp/demo-proj
```

Rules:

- The surface must be known to the catalog (aliases resolve first, so
  `mobile app` means `mobile-native`) and must not already be present.
- The evolved intent must resolve (`resolved`, else a typed abort with no
  partial mutation).
- The evolved intent must select the **same recipe** as the project.
  Otherwise the command aborts: `recipe migration is out of v1 scope`.
  Materialize a new project for the new architecture instead.
- `--provider <boilerplate>` pins the foundation for the new surface
  (must exist, serve the surface and be eligible); empty selects the
  composed default.

On success the command refreshes `project-map.json` (new surface plus
relationships), the manifest (new fingerprint and pins),
`implementation-brief.md`, `handoff.json`, the intent/decision/plan copies
and appends a `surface-add` event to `provenance.json` (history is never
rewritten).

## 2. `eng extend` — scope promotion (`planned_later` → now)

Ships a surface the original intent deliberately deferred, without
reinstalling anything. Same pipeline as `surface add`, except the surface
must already exist in the stored intent with scope `planned_later` (else
the error names the fitting command):

```bash
/tmp/eng extend --project /tmp/demo-proj --surface public-intake --catalog-dir /tmp/fixture-catalog
/tmp/eng doctor --project /tmp/demo-proj
```

On success the stored intent flips that surface to `required_now` and a
`scope-extend` event is appended to provenance.

## 3. `eng add` — product requirements (never architecture)

Appends to the intent `product_requirements`, refreshes the handoff
requirements (pending implementation) and the brief. It never adds
components: resolution before and after must select the same recipe.

```bash
/tmp/eng add --project /tmp/demo-proj --requirement "monthly PDF export for auditors"
/tmp/eng add --project /tmp/demo-proj --requirement "kiosk self-check-in" --scope planned_later
```

- `--scope required_now` (default) records the requirement as-is;
  `planned_later` additionally tracks its id under the intent
  `planned_later` lifecycle list.
- New ids are deterministic (`REQ-001`, `REQ-002`, …).
- If the text names catalog architecture vocabulary (a surface,
  capability or alias such as `mobile app`), the requirement is still
  recorded, but a warning points at `eng surface add` / `eng extend` in
  case a scope change was meant.

## 4. `eng update` — report only (no auto-apply in v1)

Compares each materialized component pin with the active catalog pin and
prints `{component, current, catalog, strategy, action}`. Strategies come
from the optional boilerplate `update_strategy` field
(`replace|merge-seed|fork-track|manual`, default `manual`); entries
without a declared strategy report as `manual`. Nothing is mutated.

```bash
/tmp/eng update --project /tmp/demo-proj --catalog-dir /tmp/fixture-catalog
/tmp/eng update --project /tmp/demo-proj --catalog-dir /tmp/fixture-catalog --json
```

Exit is always 0 with a human-readable summary (`all current` when every
pin matches); `--json` emits the machine-readable report.

## Recipe-change policy

`surface add` and `extend` never migrate a project across recipes. When
the evolved intent resolves elsewhere (for example admin-only GP-02 plus
`mobile-native` resolves GP-06), the command aborts with an explanatory
error and leaves the project untouched — `doctor` stays green and the
manifest fingerprint is unchanged. Recipe migration is out of v1 scope by
design: foundations, destinations and database profiles are chosen per
recipe, so a silent cross-recipe move would invalidate the project's
provenance.

## Provenance model

`provenance.json` keeps the bootstrap record (`materialized_at`, original
pins) forever. Evolution only appends events and rolls the
intent/plan-fingerprint and pin pointers forward:

```json
{ "type": "surface-add", "at": "2026-09-08T12:00:00Z",
  "surface": "public-intake", "provider": "tanstack-transactional-pwa", "recipe": "GP-06" }
```

`doctor` cross-checks events against the manifest and map; an event naming
an unknown surface is a warning (forward-compatible with newer cores),
never an error.
