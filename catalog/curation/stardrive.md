# Curation evidence — stardrive

- Upstream: <https://github.com/peltmonger/stardrive>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Pin: `5c449810b763140ac72133ff4ae63d8497cce77a` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists, 2026-08-28). It is
  NOT the current upstream tip (HEAD `41a42e0`), so the pin is an honest
  freeze of the reviewed legacy snapshot, not a floating reference. Re-pin
  only after re-review.
- Legacy review: `platform/boilerplates.json` entry `stardrive`
  (default/curated, tier A, category public-web). The v1 migration restores
  this identity: the previous `stardrive-public-web` id and the
  `JhonMA82/stardrive-public-web` repo were migration inventions and are
  retired (reader-side manifest alias only — see
  `docs/decisions/migration-notes.md`).
- Adapter: fetch+copy placeholder (`astro`); no setup/check commands are
  declared. Real setup (install/build) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license re-verify on pilot; pin frozen (not tip); no
  setup/checks; no pilot; Astro-specific build/adapter behavior unverified.
- Decision/delivery: default / pilot-ready (downgraded from curated under
  H3: curated requires a Pilot success record, which does not exist — see
  docs/decisions/migration-notes.md).
