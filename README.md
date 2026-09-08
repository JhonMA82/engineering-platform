# Engineering Platform 1.0 — M1 Deterministic Routing Core

M1 implements the minimal vertical slice:

```text
ProjectIntent → Catalog → Resolver → ArchitectureDecision
```

It covers routing for **Public Web (GP-01), Admin (GP-02), Python/Data (GP-03),
Mobile (GP-04), Desktop (GP-05), Multi-App (GP-06) and Commercial SaaS (GP-07)**
with hard constraints, scoring, confidence, rejected reasons, ambiguous
cases and architectural catalog-gap cases (46 routing scenarios,
floor-guarded at 40).

Out of scope for M1: materializer, composer, planner, handoff, Pi integration.

## M2 — Reproducible Composition

M2 adds the reproducible slice:

```text
ArchitectureDecision + Catalog → Composer → Composition → Planner → MaterializationPlan
```

`eng plan --input intent.json` resolves, composes and plans without
touching the filesystem: one pinned provider per required surface
(`tanstack-admin`, `hono-api`, `stardrive-public-web`,
`tanstack-transactional-pwa`, plus Fase 9: `ignite`, `tauri-ui`, `speedpy`,
`react-starter-kit`), conventional destinations
(`services/api`, `apps/admin`, `apps/mobile`, `apps/web`, `apps/intake`,
`apps/desktop`),
database profile from recipe policy, and a sha256 fingerprint binding the
intent fingerprint to the catalog version. `--json` prints the full
`MaterializationPlan` (see `schemas/materialization-plan.schema.json`).

```bash
go build -o /tmp/eng ./cmd/eng
python3 -c "import json; d=json.load(open('testdata/routing/admin-only.json')); open('/tmp/admin-intent.json','w').write(json.dumps(d['intent']))"
/tmp/eng plan --input /tmp/admin-intent.json
/tmp/eng plan --input /tmp/admin-intent.json --json
```

Details: `docs/architecture/composition.md`. Composition fixtures live in
`testdata/composition/` (11 scenarios incl. mobile-only, desktop-only,
python-job and saas-marketing plus destination collision and path
traversal); plan goldens in `testdata/plans/`.

## M3 — End-to-End Project Bootstrap

M3 closes the spine:

```text
ProjectIntent → Resolve → Plan → Materialize → Doctor → Gentle handoff
```

`eng materialize --plan plan.json --output <dir>` executes the plan in a
staging dir (fetch pinned source → verify pin → copy+prune with
path-safety checks → collision check), writes `.engineering/` state
(manifest, provenance, project map) plus agent-context artifacts (root
`AGENTS.md` router, `ARCHITECTURE.md`, `GENTLE.md`, per-surface stubs,
`implementation-brief.md`, `handoff.json`), verifies, and atomically
renames staging into place. `eng doctor --project <dir>` re-validates
consistency (exit non-zero on drift). Side effects live only in
`internal/materializer`; `internal/project` owns state and doctor;
`internal/handoff` is pure generation (no `eng handoff` command).
Adapters are catalog data: `adapter` accepts the legacy string or an
object `{operations, prune_paths, setup, checks, managed_files}`,
sources accept `git` (clone `--depth 1 --branch <pin>` with tag-match
verification) and `local` (offline fixtures in
`testdata/fixtures/boilerplates/`, pin-checked via a `PIN` marker).
Pi discovery lives in `integrations/pi/` (core never imports it).

```bash
/tmp/eng plan --input /tmp/admin-intent.json --catalog-dir <overlay> --json > /tmp/plan.json
/tmp/eng materialize --plan /tmp/plan.json --output /tmp/demo-proj --catalog-dir <overlay>
/tmp/eng doctor --project /tmp/demo-proj
```

`eng start` chains the same phases as one convenience command (plan →
materialize → doctor; `--dry-run` prints the plan with zero writes):

```bash
/tmp/eng start --intent /tmp/admin-intent.json --output /tmp/demo-proj --catalog-dir <overlay>
/tmp/eng start --intent /tmp/admin-intent.json --dry-run
```

Runbook: `docs/guides/new-project.md`. Safety and pin policy:
`docs/architecture/materialization.md`. Gentle ownership transfer:
`docs/architecture/handoff.md` (`GENTLE.md` Direct vs SDD decision,
`handoff.json` with `open_product_questions` and `gentle-decides` mode,
locked `database-profile`). Release process: `docs/guides/release.md`
(CI gates, build matrix, `v*` tags → binaries + `checksums.txt`).

## Commands

```bash
go build ./...        # or: make build
go vet ./...          # or: make vet
go test ./...         # or: make test
make cover

./eng catalog validate
./eng catalog                  # list: recipes + boilerplates + surfaces
./eng catalog show GP-06       # recipe|boilerplate|surface|capability|database-profile
./eng resolve --input <intent>.json
./eng resolve --input <intent>.json --json
./eng start --intent <intent>.json --output <dir> [--dry-run]
./eng explain --input <decision>.json
./eng version                  # binary + catalog + commit/date (ldflags-stamped releases)
```

Routing fixtures wrap the intent (`{"intent": {...}, "expect": {...}}`);
extract it first (see the M2 snippet above). Offline materialization
pilots (GP-02 single, GP-06 multi, GP-04 mobile, surface-add evolution)
run with fixture sources only:

```bash
go test ./internal/app/ -run 'Pilot' -count=1 -v   # also in .github/workflows/pilots.yml
```

The `eng` binary is built from `./cmd/eng`:

```bash
go build -o eng ./cmd/eng
```

## Layout

- `internal/domain/` — pure contracts, zero infra imports.
- `internal/catalog/` — declarative JSON catalog loader, validator, index.
- `catalog/` — catalog data (recipes, boilerplates, surfaces, capabilities).
- `internal/resolver/` — pure in-memory R1–R10 pipeline.
- `internal/app/` — thin services (`Resolve`/`Plan`/`Materialize`/`Doctor`/evolution) over the core.
- `internal/cli/` — stdlib-flag CLI; `cmd/eng` entrypoint.
- `schemas/` — JSON schemas for intent and decision.
- `testdata/routing/` — routing fixtures asserted by resolver tests.
- `docs/adr/` — architecture decision records.

## Fase 10 — Project evolution

```text
Stored intent → Resolve → Compose → Plan → Delta materialize → Doctor
```

- `eng surface add --project <dir> --surface <id> [--provider <boilerplate>]` —
  architecture evolution: the evolved intent must resolve and select the
  same recipe, then only the delta components materialize (existing
  directories untouched). Recipe migration aborts with an explanatory
  error and zero mutation — it is out of v1 scope by design.
- `eng extend --project <dir> --surface <id>` — promotes a `planned_later`
  surface to `required_now` through the same pipeline.
- `eng add --project <dir> --requirement "<text>" [--scope required_now|planned_later]` —
  records a product requirement (brief + handoff refresh, deterministic
  `REQ-NNN` ids); architecture is never touched, and architecture
  vocabulary in the text raises a warning pointing at the scope commands.
- `eng update --project <dir> [--json]` — report-only pin comparison
  `{component, current, catalog, strategy, action}` with the
  `replace|merge-seed|fork-track|manual` strategy vocabulary (boilerplate
  `update_strategy`, default `manual`); exit 0, no mutation.

Provenance is append-only (`surface-add` / `scope-extend` /
`requirement-add` events; `doctor` warns — never errors — on unknown
surfaces). Runbook: `docs/guides/evolve-project.md`.

## Catalog (Fase 9)

| Recipe | Family | Primary foundation(s) |
| --- | --- | --- |
| GP-01 | Public web | stardrive-public-web |
| GP-02 | Admin | tanstack-admin |
| GP-03 | Python/Data | speedpy |
| GP-04 | Mobile | ignite, hono-api |
| GP-05 | Desktop | tauri-ui |
| GP-06 | Multi-app | stardrive-public-web, tanstack-admin, hono-api, tanstack-transactional-pwa |
| GP-07 | Commercial SaaS | react-starter-kit |

Surfaces: `public-web`, `web-admin`, `public-intake`, `api`,
`mobile-native`, `desktop`, `tui`. Capabilities include `shared-backend`,
`anonymous-public-access`, `offline-operation`, `realtime`,
`local-filesystem` and `background-processing`. Curation evidence per
foundation lives in `catalog/curation/`; migration decisions and skips in
`docs/decisions/migration-notes.md`.
