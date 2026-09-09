# Curation evidence — hono-api

- Upstream: <https://github.com/JhonMA82/api-starter>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Pin: `360eb274cc5936fee5aab88eb8bd94977e95dfc9` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists, 2026-09-01,
  `chore(release): v0.12.2`) and equal to the advertised upstream HEAD, so
  the reviewed snapshot is current, not stale.
- Legacy review: `platform/boilerplates.json` entry `hono-api`
  (name `Consulting API Starter`, default/released, tier A, category
  typescript-api). The v1 migration restores this identity: the previous
  `JhonMA82/hono-api` repo was a migration invention (catalog id mistaken
  for repository name). The id stays `hono-api`; only the repo is corrected.
- Adapter (v1.1.0): fetch+generate. The `api-starter` repository is a
  factory, not a copyable API: the pinned factory is acquired into a
  temporary sandbox, prepared with `bun install --frozen-lockfile`,
  executed as `bun run create:project -- --out={output} --profile=<id>`
  and only the declared output flows into the project staging.
- Profiles (curated for the v1.1.0 contract): `minimal` (default),
  `data-api` (shared-backend), `authenticated` (shared-backend without
  public access), `multi-tenant-core` (shared-backend plus the stable
  `organizations` requirement id), `integration-platform`
  (shared-backend plus background-processing). `platform` is deliberately
  not declared: it must never be an automatic fallback.
- Pilot: 2026-09-09, manual generation pilots against the pinned commit
  `360eb274` (equals upstream HEAD): Pilot A (`minimal`) generated
  successfully via `bun run create:project -- --out=<dir>
  --profile=minimal` after `bun install --frozen-lockfile`; the generated
  project passes `bun run typecheck`, ships `AGENTS.md`, and contains no
  `generator/` or `templates/` factory residue. Pilot B (`authenticated`)
  generated successfully with the same command shape. Both pilots used the
  exact argv declared in the catalog adapter, confirming the contract.
  Delivery stays `pilot-ready` until license re-verification and a
  maintainer promotion review (see Gaps).
- Gaps (explicit): license re-verify on pilot; no setup/checks; no pilot;
  shared-backend behavior beyond the fixture path unverified.
- Decision/delivery: default / pilot-ready (downgraded from released under
  H3: release-grade evidence bars require a Pilot success record, which does
  not exist — see docs/decisions/migration-notes.md).
