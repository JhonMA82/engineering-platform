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
- v1 adapter (H4): the generic `generate` operation runs the pinned
  generator (`npx ignite-cli@11.5.0 new {name} --yes`; npm metadata
  confirms 11.5.0 exists) and validates the scaffolded directory through
  the same copy+prune path as fetched trees. `--yes` accepts CLI defaults
  (bundle id, git init, dep install); no extra setup/check commands are
  declared until the network pilot confirms post-generation behavior —
  see `docs/decisions/ignite-materialization.md`.
- Pilot: not run — no materialization pilot was executed for this
  foundation in legacy or v1 (generator-based materialization is H4 work;
  see docs/decisions/ignite-materialization.md).
- Decision/delivery: default / pilot-ready (downgraded from curated under
  H3: curated requires a Pilot success record, which does not exist — see
  docs/decisions/migration-notes.md).
