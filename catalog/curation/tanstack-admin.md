# Curation evidence — tanstack-admin

- Upstream: <https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Pin: `e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists, 2026-08-26). It is
  NOT the current upstream tip (HEAD `9b07657`), so the pin is an honest
  freeze of the reviewed legacy snapshot, not a floating reference. Re-pin
  only after re-review.
- Legacy review: `platform/boilerplates.json` entry `tanstack-admin`
  (legacy_id `tanstack-shadcn-admin-dashboard`, default/curated, tier A,
  category admin-web). The v1 migration restores this identity: the previous
  `JhonMA82/tanstack-admin` repo was a migration invention (id mistaken for
  repository name).
- Adapter: fetch+copy placeholder (`tanstack`); no setup/check commands are
  declared. Real setup (install/build) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license re-verify on pilot; pin frozen (not tip); no
  setup/checks; no pilot; CRUD/auth behavior beyond the fixture path
  unverified.
- Decision/delivery: default / pilot-ready (downgraded from curated under
  H3: curated requires a Pilot success record, which does not exist — see
  docs/decisions/migration-notes.md).
