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
- Adapter: fetch+copy placeholder (`hono`); no setup/check commands are
  declared. Real setup (install/typecheck) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license re-verify on pilot; no setup/checks; no pilot;
  shared-backend behavior beyond the fixture path unverified.
- Decision/delivery: default / pilot-ready (downgraded from released under
  H3: release-grade evidence bars require a Pilot success record, which does
  not exist — see docs/decisions/migration-notes.md).
