# Curation evidence — next-admin

- Upstream: <https://github.com/arhamkhnz/next-shadcn-admin-dashboard>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Pin: `15e0a081bc1acad2b47adc638471b6e67fa36f10` (`main`), verified
  2026-09-08 via the GitHub commits API (commit exists, 2026-09-04). It is
  NOT the current upstream tip (HEAD `22af3d1`), so the pin is an honest
  freeze of the reviewed legacy snapshot, not a floating reference. Re-pin
  only after re-review.
- Legacy review: `platform/boilerplates.json` entry `next-admin`
  (legacy_id `next-shadcn-admin-dashboard`, alternative/curated, tier B,
  category admin-web). Recovered by the legacy-identity repair migration;
  it was omitted from the v1 catalog before, not rejected.
- Role: alternative for scenarios where a concrete constraint favors Next;
  never the default (the admin track default stays `tanstack-admin`). It is
  eligible for `web-admin` but wins no recipe primary.
- Adapter: fetch+copy placeholder (`next`); no setup/check commands are
  declared. Real setup (install/build) and check commands are a pilot gap.
- Pilot: not run — no materialization pilot was executed for this foundation
  in legacy or v1.
- Gaps (explicit): license re-verify on pilot; pin frozen (not tip); no
  setup/checks; no pilot; CRUD/auth behavior unverified.
- Decision/delivery: alternative / pilot-ready (downgraded from curated
  under H3: curated requires a Pilot success record, which does not exist
  — see docs/decisions/migration-notes.md).
