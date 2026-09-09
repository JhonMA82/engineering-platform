# Materialization (M3)

## Staging → verify → move

`internal/materializer` is the first and only package with
filesystem/process side effects. Resolver, composer and planner stay pure
and are only read for types. The materializer never re-resolves
architecture: the `MaterializationPlan` is the decision.

```text
validate output dir (empty-or-new)
  → staging temp dir (sibling of the output dir)
  → per component: fetch source (or run the curated generator) → verify pin → copy + prune
  → path-safety checks + collision check
  → write manifest, provenance, project map, agent-context files (staging)
  → post-materialize checks (staging)
  → atomic rename staging → output
```

Every check that needs no side effects runs before staging exists, so a
malformed plan, unknown provider, drifted pin, unsafe destination,
collision or malicious adapter command fails with the output directory
untouched. Any failure before the final rename removes staging. The
output directory is only ever touched by the rename commit.

## Safety rules

- **Destinations** are project-relative, clean slash paths. Absolute
  paths, `.`, empty segments and `..` are rejected, never sanitized.
- **Collisions**: duplicate and nested destinations (e.g. `apps` vs
  `apps/web`) are rejected in pre-validation and re-checked post-copy.
- **Symlinks** are never followed during copy. A symlink whose resolved
  target escapes the source tree aborts materialization; internal links
  are reproduced as links. Permission bits (including exec) are
  preserved; sockets, devices and other special files are rejected.
- **Prune paths** come only from the adapter declaration, must be
  relative and clean, and missing entries are skipped.
- **Idempotency**: re-running into a non-empty directory fails cleanly
  demanding an empty or new directory. There is no merge mode in v1.
- **Commands** run only from curated adapter declarations, argv-only
  with no shell: shell interpreters, path-like binaries and shell
  metacharacters (`; & | $ \` \` () < >` and control bytes) are rejected
  before execution. Environment is an allowlist (PATH, locale, temp,
  CI signals); it is never logged. Each command has a timeout via
  context and bounded failure output.

## Git-pin policy

- Real boilerplates are fetched with
  `git clone --depth 1 --branch <pin>` for named refs, or with a
  single-commit fetch (`init` + `fetch --depth 1 origin <sha>` +
  detached checkout) for full SHA pins.
- After cloning, HEAD must resolve to the pin
  (`rev-list -n 1 <pin>` == HEAD); otherwise the fetch is discarded.
- A missing pin fails the clone itself, so default-branch drift can
  never satisfy a pinned fetch. Pins are immutable commit SHAs by
  convention (boilerplate pin policy H6); a full SHA is fetched as one
  commit (`init` + `fetch --depth 1 origin <sha>` + detached checkout)
  because `clone --branch` cannot resolve a SHA. The verification, not
  the naming, is the guarantee.
- `local` sources exist for fixtures and tests. They verify the pin
  against a `PIN` marker file at the source root (trimmed content must
  equal the plan pin) — the offline analog of the tag match. The marker
  ships with the copied tree like any other source file.

## Adapter contract

`catalog/boilerplates/*.json` accept `adapter` as a legacy plain string
(name only, fetch+copy semantics) or as an object:

```json
"adapter": {
  "name": "hono",
  "operations": ["fetch", "copy"],
  "prune_paths": [],
  "setup": [],
  "checks": [],
  "managed_files": ["AGENTS.md"]
}
```

`operations` uses the §71 vocabulary (`fetch|copy|prune|template|
compose|generate`); adding a boilerplate that combines them needs no core
release. `generate` covers generator CLIs that scaffold their own output
directory instead of shipping a copyable tree (e.g.
`npx ignite-cli@11.5.0 new {name} --yes`):

```json
"operations": ["generate"],
"generate": {"run": ["npx", "ignite-cli@11.5.0", "new", "{name}", "--yes"], "output": "{name}"}
```

`run` is argv-only (never a shell string; the same argv gate as
setup/checks) and `output` is the path the command must produce.
Placeholders are closed (`{name}`, `{project}`, `{surface}`,
`{profile}`, `{output}`); every other byte is literal. Generated
foundations run in a per-component sandbox: an optional `prepare`
step runs inside the acquired factory, the command executes with the
factory (or a fresh directory for external tools) as cwd, and only
the declared non-empty output — strictly inside the sandbox — is
copied into project staging. Profiles, deterministic selection
(smallest valid wins) and the full contract live in
`docs/architecture/generated-foundations.md`.
See `docs/decisions/ignite-materialization.md`. `setup`/`checks`
are argv arrays (`["npm","run","build"]`) or
`{"run": [...]}` / `{"command": ..., "args": [...]}` objects — never
shell strings. v1 catalog entries declare `fetch+copy` with no commands,
so materialization is fully offline; recipe `quality_gates` stay
plan-level informational checks until curated executable checks exist.

`sources` accept `{"type": "git"}` (repo from the entry or its own
`repo` override) and `{"type": "local", "path": "<absolute dir>"}`.
An object adapter without `name` falls back to the boilerplate id so
composer eligibility and tech-signal matching keep working.

## What gets written

- Components at their planned destinations (source files verbatim,
  minus pruned paths).
- `.engineering/project.json` — manifest: fingerprint, catalog
  version, pins, file list. No timestamps.
- `.engineering/provenance.json` — core/catalog versions, intent and
  plan fingerprints, pins, per-component generation records
  (strategy, logical name, profile, arguments, adapter
  fingerprint), `materialized_at`. The only timestamped document in
  the system.
- `.engineering/project-map.json` — surfaces `{path, provider,
  instructions}` plus `consumes` edges toward a shared `api` backend.
- Agent context via `internal/handoff` (pure generation, no separate
  `eng handoff` command): intent/decision/plan copies,
  `implementation-brief.md`, `handoff.json`
  (`ready_for_implementation` → `gentle-ai`), root `AGENTS.md` router,
  `ARCHITECTURE.md`, `GENTLE.md`, and per-surface `AGENTS.md` stubs
  wherever the foundation shipped none (shipped ones are preserved).

`eng doctor` re-validates all of it: manifest↔filesystem,
project-map↔dirs, provenance↔manifest fingerprint, plan-copy drift,
plus warnings for untracked top-level entries.
