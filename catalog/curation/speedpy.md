# Curation evidence — speedpy

- Upstream: <https://github.com/speedpy/speedpy>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Maintenance signal: tier A; legacy evidence lists AGENTS.md, agent skills,
  design system, 40 test files and a production checklist.
- Pin: `3fbf725d8e9cf6b8aadb3aeaf1db2822522282b9` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists; it is NOT the current
  upstream tip, so the pin is an honest freeze of the reviewed snapshot, not a
  floating reference).
- Legacy review: `curation/speedpy/evidence.json` (reviewed 2026-08-31,
  ai_friendly high-heavy); gaps: 85 KB AGENTS.md, no upstream CI at the pinned
  commit, demo stripping required.
- Architectural consequence (why the `background-processing` capability
  exists): the legacy profile documents Celery plus Redis async tasks and a
  Docker mode with PostgreSQL/Redis/Celery — a worker topology, not just
  request handling. That is what GP-03 routes on.
- Pilot: not run — install, typecheck, lint, test and build were not run
  from the pinned snapshot in legacy or v1.
- Decision/delivery: default / pilot-ready (downgraded from curated under
  H3: curated requires a Pilot success record, which does not exist — see
  docs/decisions/migration-notes.md).
