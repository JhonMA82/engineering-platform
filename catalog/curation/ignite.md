# Curation evidence — ignite

- Upstream: <https://github.com/infinitered/ignite>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Maintenance signal: tier B; active upstream (pinned commit equals upstream
  `master` HEAD at review time).
- Pin: `e829d2f922c5568a59a77bfb6232aeb500be3f13` (`master`), verified
  2026-09-08 via `git ls-remote` (advertised HEAD) — the reviewed snapshot is
  the current upstream tip, not a stale freeze.
- Legacy review: `curation/ignite/evidence.json` (reviewed 2026-09-01,
  ai_friendly medium); gaps recorded there: generated project still needs the
  Engineering AGENTS.md overlay.
- v1 adapter gap (known, not blocking routing): legacy materializes Ignite via
  a generator command (`ignite-cli`), but the v1 adapter vocabulary only offers
  fetch/copy/prune/template/compose. The v1 adapter therefore records a
  fetch+copy placeholder; generator-based materialization is future
  materializer work, documented in `docs/decisions/migration-notes.md`.
- Decision/delivery: default / curated (unchanged from legacy).
