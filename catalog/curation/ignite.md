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
- v1 adapter: the generic `generate` operation runs the pinned
  generator (`npx ignite-cli@11.5.0 new {name} --yes`; npm metadata
  confirms 11.5.0 exists) and validates the scaffolded directory through
  the same copy+prune path as fetched trees. `--yes` accepts CLI defaults
  (bundle id, git init, dep install); the CLI inputs are app name
  (positional, becomes the output directory) plus `parameter.option`
  flags (`yes`, `bundle`, `git`, `installDeps`, `packager`), so no
  interactive prompt is needed. No extra setup/check commands are
  declared until the network pilot confirms post-generation behavior.
- Pilot: not run — no materialization pilot was executed for this
  foundation in legacy or v1. Confirming or extending setup/checks from
  the generated tree is the pilot's explicit follow-up, not guessed here.
- Decision/delivery: default / pilot-ready (downgraded from curated:
  curated requires a Pilot success record, which does not exist).
