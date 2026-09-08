# Materialization (M3)

## Staging → verify → move

`internal/materializer` is the first and only package with
filesystem/process side effects. Resolver, composer and planner stay pure
and are only read for types. The materializer never re-resolves
architecture: the `MaterializationPlan` is the decision.

```text
validate output dir (empty-or-new)
  → staging temp dir (sibling of the output dir)
  → per component: fetch source → verify pin → copy + prune
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
  `git clone --depth 1 --branch <pin>`.
- After cloning, HEAD must resolve to the pin
  (`rev-list -n 1 <pin>` == HEAD); otherwise the fetch is discarded.
- A missing pin fails the clone itself, so default-branch drift can
  never satisfy a pinned fetch. Pins are tags by convention
  (`v1.0.0`); the verification, not the naming, is the guarantee.
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
compose`); adding a boilerplate that combines them needs no core
release. `setup`/`checks` are argv arrays (`["npm","run","build"]`) or
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
  plan fingerprints, pins, `materialized_at`. The only timestamped
  document in the system.
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
