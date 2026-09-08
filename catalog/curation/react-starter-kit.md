# Curation evidence — react-starter-kit

- Upstream: <https://github.com/kriasoft/react-starter-kit>
- License: MIT (legacy registry record; verify on pilot before materializing).
- Maintenance signal: tier C as recorded in legacy (kept as `specialized`
  accordingly: eligible for explicit commercial-SaaS signals, never the
  default for generic intents).
- Pin: `0aa7603435f16159ad0b8fef68fb7f6280be7ca1` (`main`), verified
  2026-09-08 via `git ls-remote` (advertised HEAD).
- Legacy review: `curation/react-starter-kit/evidence.json` (reviewed
  2026-09-01, ai_friendly high); adapter declares seed-fork ownership,
  frozen install, and typecheck/lint/test/build gates.
- Gaps carried over (no pilot was executed in legacy either): install,
  typecheck, lint, test and build were not run from the pinned snapshot;
  tenant isolation, billing sandbox, migrations, rollback, deployment and
  observability remain unverified; Cloudflare Workers, Neon, Bun and
  provider-specific integrations require project-level acceptance.
- Decision/delivery: specialized / curated (unchanged from legacy).
  Specialized keeps the foundation eligible while preventing it from winning
  generic intents on simplicity alone — GP-07 is selected by explicit
  commercial-SaaS signals (React/Cloudflare stack, SaaS composition).
