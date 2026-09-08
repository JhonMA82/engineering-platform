# Curation evidence — tanstack-admin

- Upstream: <https://github.com/JhonMA82/tanstack-admin>
- License: unverified (gap). No license record was carried over from legacy
  and the upstream repository is not publicly reachable (see Pin), so the
  license MUST be confirmed on pilot before materializing for a real project.
- Pin: `v1.0.0` (declared, NOT verified). `git ls-remote` against the
  upstream URL fails with "Repository not found" (checked 2026-09-08): the
  repository is private, renamed, or removed. Offline pilots repoint this id
  at a local fixture source; a network pilot must re-verify the pin first.
- Adapter: fetch+copy placeholder (`tanstack`); no setup/check commands are
  declared. Real setup (install/build) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license unverified; pin unreachable; no setup/checks;
  no pilot; CRUD/auth behavior beyond the fixture path unverified.
- Decision/delivery: curated / pilot-ready (downgraded from stable under H3:
  release-grade evidence bars require a Pilot success record, which does
  not exist — see docs/decisions/migration-notes.md).
