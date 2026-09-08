# Curation evidence — tanstack-transactional-pwa

- Upstream: <https://github.com/JhonMA82/tanstack-transactional-pwa>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Pin: `f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists, 2026-09-04) and
  equal to the advertised upstream HEAD, so the reviewed snapshot is
  current, not stale.
- Legacy review: `platform/boilerplates.json` entry
  `tanstack-transactional-pwa` (specialized/curated, tier B, category
  typescript-public-web). The v1 migration restores the historical commit
  pin (replacing the invented `v1.0.0` tag, which the upstream never
  advertised).
- Adapter: fetch+copy placeholder (`tanstack`); no setup/check commands are
  declared. Real setup (install/build) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license re-verify on pilot; no setup/checks; no pilot;
  offline-sync/PWA behavior beyond the fixture path unverified.
- Decision/delivery: specialized / pilot-ready (downgraded from curated
  under H3: curated requires a Pilot success record, which does not exist
  — see docs/decisions/migration-notes.md).
