# Curation evidence — tauri-ui

- Upstream: <https://github.com/agmmnn/tauri-ui>
- License: MIT declared in README and `packages/create-tauri-ui/package.json`;
  no LICENSE file at the pinned commit (deleted upstream in `0a9dce3`, 2023).
- Maintenance signal: tier B; active upstream (pinned commit equals upstream
  `master` HEAD at review time).
- Pin: `8eb86d894c19b6df04ff883ab28b412b1e5f23ea`, verified 2026-09-08 via
  `git ls-remote` (advertised HEAD). Note: legacy registry/adapter record the
  branch as `main`, but the upstream default branch is `master`.
- Legacy review: `curation/tauri-ui/evidence.json` with a full 2026-09-04
  pilot (bun install, template build, generated-app `tsc -b && vite build`
  green). Verified at source level: Tauri v2 shell, installer bundling config,
  greet IPC, logging, opener integration, offline-by-architecture frontend.
- Caveats carried over: local file I/O needs `tauri-plugin-fs` added by the
  project; auto-update, tray and deep-links are absent; Rust-layer build and
  real installer generation were not exercised; generation needs live network
  (registry fetches at generation time).
- The v1 `offline-operation` capability claim means offline-by-architecture
  (bundled assets plus local SQLite), not a data-sync layer.
- Pilot: 2026-09-04 full pilot passed (bun install, template build and
  generated-app `tsc -b && vite build` green; Rust layer and real
  installers not exercised — see caveats above).
- Decision/delivery: default / curated (unchanged from legacy).
