# CLI reference

The complete `eng` command surface, derived from `internal/cli`
(`eng help` prints the same summary). Flags use stdlib syntax:
`--flag value` or `--flag=value`.

## Conventions

- Exit `0`: success. Exit `1`: runtime failure (unreadable input,
  resolution/composition error, `doctor` error-severity findings).
  Exit `2`: usage error (unknown command, missing required flag).
- `--catalog-dir DIR` (resolve, plan, materialize, start, surface add,
  extend, add, update, catalog): reason over an organization overlay
  instead of the embedded default catalog.
- `--json` (resolve, plan, start, doctor, add-excluded, update, catalog
  list/show): machine-readable output on stdout.
- Run reports: `start`, `materialize` and `doctor` persist one JSON trace
  per failed run (or doctor run with error-severity findings) under
  `.engineering/runs/<timestamp>-<cmd>.json` in the project/output dir
  when it exists, otherwise in the workspace. Each report records the
  intention (name, problem, surfaces, recipe/plan fingerprint, output)
  vs the result (status, error class/message, hint, findings, manifest
  fingerprint). stderr always names the file plus the hint; a report
  failure warns only and never masks the run error.

## Commands

```text
eng init [--agent pi|opencode] [--force] [--no-agent]
```

Prepares the current directory as a workspace (agent integration plus
`.engineering/` state). OpenCode is the default agent (targets v1 and v2:
the installed `.opencode/commands/` + `.opencode/skills/` layout is the
v2 preferred form and remains v1-compatible, so no version flag is
needed); `--agent pi` installs the Pi integration instead. Prints what
it created, updated, or kept, then
the next step (`opencode`, `pi`, or `eng start --intent intent.json
--output .` for `--no-agent`). Re-running reports "already initialized"
and writes nothing; `--force` repairs eng-owned files even when
customized.

```text
eng resolve --input intent.json [--json] [--verbose] [--catalog-dir DIR]
```

Resolves a `ProjectIntent` to an architecture decision and prints the
human-readable explanation (`--json` prints the full decision;
`--verbose` adds fingerprint and candidate detail).

```text
eng plan --input intent.json [--json] [--catalog-dir DIR]
```

Prints the deterministic materialization plan (recipe, database
profile, per-component foundation@pin, strategy/profile, destinations,
fingerprint). Writes nothing.

```text
eng materialize --plan plan.json --output <dir> [--intent intent.json] [--decision decision.json] [--catalog-dir DIR]
```

Executes a plan into a new or empty directory, or into the eng-init
workspace itself (`--output .` when `.engineering/bootstrap.json`
exists). Re-running into a non-empty directory is refused, as is a new
or empty subdirectory nested inside an init workspace (never derive the
output from the intent/project name there). `--intent` / `--decision`
embed copies in the project state.

```text
eng start --intent intent.json --output <dir> [--catalog-dir DIR] [--dry-run] [--json]
```

Chains resolve → plan → materialize → doctor. `--dry-run` prints the
plan and performs zero filesystem writes (`--output` may be omitted
with `--dry-run`). Inside an `eng init` workspace use `--output .`;
a nested project-named subdirectory is refused. After successful
validation, workspaces prepared by `eng init` shed their disposable
bootstrap resources (warns, never fails, on cleanup errors).

```text
eng doctor [--project <dir>] [--json]
```

Re-validates a materialized project (manifest↔filesystem,
project-map↔dirs, provenance↔manifest fingerprint, plan-copy drift).
Defaults to the current directory. Prints `doctor: <dir> is consistent`
when clean; error-severity findings exit `1`.

```text
eng surface add --project <dir> --surface <id> [--provider <boilerplate>] [--catalog-dir DIR]
eng extend --project <dir> --surface <id> [--catalog-dir DIR]
eng add --project <dir> --requirement "<text>" [--scope required_now|planned_later] [--catalog-dir DIR]
eng update --project <dir> [--json] [--catalog-dir DIR]
```

Project evolution (full workflow: `docs/guides/evolve-project.md`):
`surface add` materializes only the delta for one new required surface
(aborts when the evolved intent would change recipe); `extend`
promotes a `planned_later` surface; `add` records a product
requirement without touching architecture; `update` prints a
`{component, current, catalog, strategy, action}` pin-comparison
report and mutates nothing (exit is always `0` on success).

```text
eng explain --input decision.json
```

Prints why a saved decision resolved the way it did (same explanation
as `resolve`).

```text
eng catalog [list] [--catalog-dir DIR] [--json]
eng catalog show <id> [--catalog-dir DIR] [--json]
eng catalog validate [--catalog-dir DIR]
```

Bare `eng catalog` lists recipes, boilerplates, and surfaces; `show`
renders one entry by id across every catalog kind (flags accepted on
either side of the id); `validate` enforces structural validity plus
the curation evidence bar and prints
`catalog OK: version <v>, <n> recipes, <m> boilerplates`.

```text
eng version
```

Prints core release line, catalog version, schema, commit, and build
date. Release builds stamp these via ldflags; worktree builds report
`dev` / `unknown`.

```text
eng self-update [--version X.Y.Z|latest] [--repo OWNER/NAME] [--check] [--yes]
```

Updates `eng` itself from the GitHub release line: resolves the target
tag (`latest` via the release redirect, no token needed), downloads the
platform binary plus `checksums.txt`, verifies sha256, and atomically
replaces the running executable. `--check` (and a bare run without
`--yes`) resolves and reports without writing; `--yes` applies.
`ENG_GITHUB_REPO`, `ENG_VERSION`, `GITHUB_TOKEN` override the defaults.
Remote install/uninstall scripts (`scripts/install.sh|ps1`,
`scripts/uninstall.sh|ps1`, bundled with every release) cover fresh
machines and removal.
