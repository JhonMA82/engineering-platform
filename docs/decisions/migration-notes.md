# Migration notes — Fase 9 useful-catalog migration

Date: 2026-09-08 · Scope: PRD §38 procedure over the legacy 0.x catalog
(`platform/boilerplates.json`, `platform/golden-paths.json`,
`platform/database-profiles.json`, `curation/`).

Method per entry: read old entry → read adapter → read evidence → transform
to the v1 contract → validate (`eng catalog validate`) → contract test
(`internal/catalog/phase9_test.go`) → routing/composition fixture. Legacy is
read-only; no code was copied, only knowledge extracted. Pins were verified to
exist: `git ls-remote` for advertised HEADs, GitHub commits API for the one
frozen non-HEAD pin (speedpy).

## Migrated

### ignite → `catalog/boilerplates/ignite.json` · GP-04

- Source: legacy `ignite` (default/curated, tier B, category mobile).
- Pin `e829d2f922c5568a59a77bfb6232aeb500be3f13` (`master`): verified — it is
  the advertised upstream HEAD, so the reviewed snapshot is current, not stale.
- Status kept: default / curated. Provides `mobile-native`; tags
  `react-native, expo, typescript`.
- Evidence: `catalog/curation/ignite.md` (legacy `curation/ignite/`
  reviewed 2026-09-01; gap carried over: generated project needs the
  Engineering AGENTS.md overlay).
- Known adapter gap (documented, not blocking routing): legacy materializes
  via a generator command (`ignite-cli@11.5.0`), but the v1 adapter vocabulary
  is fetch/copy/prune/template/compose, so the v1 adapter records a fetch+copy
  placeholder. Generator-based materialization is future materializer work.

### tauri-ui → `catalog/boilerplates/tauri-ui.json` · GP-05

- Source: legacy `tauri-ui` (default/curated, tier B, category desktop).
- Pin `8eb86d894c19b6df04ff883ab28b412b1e5f23ea`: verified advertised HEAD.
  Correction carried over: legacy records the branch as `main`, but the
  upstream default branch is `master` (confirmed by the pilot notes).
- Status kept: default / curated. Provides `desktop` + `offline-operation`
  (offline-by-architecture: bundled assets plus local SQLite — not a
  data-sync layer), tags `tauri, rust, react, typescript`. Adapter setup and
  build check mirror the piloted legacy adapter (`bun install`,
  `bun run build`).
- Evidence: `catalog/curation/tauri-ui.md` (legacy full pilot 2026-09-04;
  caveats carried over: no `tauri-plugin-fs`, no auto-update/tray/deep-links,
  Rust layer and real installers not exercised, generation needs network).
- New surface `desktop` added first (`catalog/surfaces/desktop.json`); alias
  `desktop app → desktop`.

### speedpy → `catalog/boilerplates/speedpy.json` · GP-03

- Source: legacy `speedpy` (default/curated, tier A, category python-data).
- Pin `3fbf725d8e9cf6b8aadb3aeaf1db2822522282b9` (`main`): verified to exist
  via the commits API, but it is NOT the current upstream tip — an honest
  freeze of the reviewed snapshot. Re-pin only after re-review.
- Status kept: default / curated. Provides `api` + `background-processing`,
  tags `python, django`; adapter mirrors the legacy overlay adapter
  (`uv sync`, `manage.py check`, prune `.claude`).
- Why the new `background-processing` capability exists: the legacy profile
  documents Celery plus Redis async tasks and a Docker mode with
  PostgreSQL/Redis/Celery — a worker topology, i.e. an architectural
  consequence, not a product feature. Resolver rule added: R4 derives
  `background-processing` from `ops.background_jobs` (with unit test
  `TestDeriveBackgroundJobs`; documented in `docs/architecture/routing.md`).
- Evidence: `catalog/curation/speedpy.md` (legacy gaps: 85 KB AGENTS.md, no
  upstream CI at the pin, demo stripping required).

### react-starter-kit → `catalog/boilerplates/react-starter-kit.json` · GP-07

- Source: legacy `react-starter-kit` (specialized/curated, tier C,
  category saas-edge).
- Pin `0aa7603435f16159ad0b8fef68fb7f6280be7ca1` (`main`): verified
  advertised HEAD.
- Status kept: specialized / curated. Provides `public-web, api` +
  `shared-backend, anonymous-public-access`; tags `react, cloudflare, saas`.
- Why GP-07 exists although GP-06 already composes public-web stacks
  (checked, not assumed): GP-06 composes separate foundations
  (site + backoffice + api); GP-07 selects the single seed-fork commercial
  foundation (billing/organizations built in, Cloudflare-first). Billing stays
  a product feature (bonus-only — a bare billing landing still resolves
  GP-01, fixture `saas-billing-landing`); GP-07 wins on foundation fit for
  React/Cloudflare SaaS compositions. Legacy GP-07 status `trial` maps to v1
  `active` (selectable, not yet stable).
- Evidence: `catalog/curation/react-starter-kit.md` (legacy gaps: no pilot
  executed; tenant isolation, billing sandbox, migrations and observability
  unverified; provider integrations need project-level acceptance).

## Recipes added (simplified, no `project_types` field)

- GP-03 Python/Data (speedpy; `postgresql-managed` default, sqlite allowed).
- GP-04 Mobile (ignite + hono-api; shared backend, `postgresql-managed`).
- GP-05 Desktop (tauri-ui; `sqlite-local`, no shared backend).
- GP-07 Commercial SaaS (react-starter-kit; shared backend,
  `postgresql-managed`).
- GP-06 untouched: still the only recipe covering `web-admin` compositions,
  offline-capable multi-app and realtime. All 15 pre-existing routing and 7
  composition fixtures pass unchanged — the expansion is purely additive.

## Dataset: 15 → 41 routing cases

All 20 §31 canonical names are covered (reusing the overlapping 15 where the
mapping is exact), plus mobile/desktop/python/saas/kiosk/offline/api-client
pairs that discriminate the new recipes: `mobile-field-app` vs
`mobile-offline-field` (native vs offline-capable), `saas-plus-marketing` vs
`saas-plus-mobile` (single SaaS foundation vs composed multi-app),
`saas-billing-landing` (billing as bonus-only), `prefer-python-api`
(preferences never eliminate), `desktop-api-combo` (honest `unsupported`:
each surface is covered somewhere, no recipe composes both).

## Skipped (with reasons — no dishonest pins)

- `fastapi`: legacy status `alternative/pilot-ready` with NO upstream pin
  recorded and NO evidence file (`curation/fastapi/` does not exist). §38
  step 3 (read evidence) cannot be satisfied; re-curation with a fresh tag pin
  is future work, not a migration.
- `goship`: legacy `experimental/catalog-only`, reference-only integration,
  no pin recorded. Nothing honest to freeze.
- `next-admin`: `alternative/curated` but outside the Fase 9 scope list; the
  admin track is already covered by `tanstack-admin`. Defer to a later
  migration slice if a Next.js alternative is needed.
- `turso-libsql` / `turso-sync` database profiles: GP-05 allows only
  `sqlite-local` (GP-05 legacy lists `turso-sync` as allowed, but no Turso
  profile exists in the v1 catalog yet). Adding provider profiles is a
  separate catalog slice.
