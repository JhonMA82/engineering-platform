# Generated Foundations (v1.1)

Some foundations are not repositories to copy. They are factories that
must be executed to produce the right minimal project variant
(`api-starter` generates an API via `create:project`; an admin fork
generates via `generate:project --profile minimal`; Ignite scaffolds via
its CLI). Engineering Platform supports both kinds with two strategies
the core understands — `copy` and `generate` — and no provider-specific
branches.

## Static vs generated

```text
Foundation
├── Static (copy)
│   └── fetch pinned source → copy/prune → setup → checks
│
└── Generated (generate)
    └── acquire pinned generator → prepare → execute curated argv
        → validate declared output → copy output into project staging
        → setup/checks on the output
```

The factory never becomes part of the final project: only the declared
generated output is copied into staging. A generated project contains
the product, never the complete factory repository.

## Lifecycle

```text
ProjectIntent
    ↓
Resolver → ArchitectureDecision (what family/recipe fits; no CLI syntax)
    ↓
Composer → Composition (which provider per Surface, destinations,
           database profile; never executes generators)
    ↓
Foundation Configuration → MaterializationConfig per component
    (pure, deterministic: strategy, logical name, profile, arguments,
    adapter fingerprint)
    ↓
Planner → MaterializationPlan v2 (records the configuration)
    ↓
Materializer (the ONLY filesystem/process boundary; executes the plan
    verbatim, never selects profiles or architecture)
```

Planning is deterministic and offline-capable; `eng plan --json` shows
foundation, pin, strategy, profile, non-runtime arguments and the
destination each component will land in.

## Adapter contract

```json
"adapter": {
  "name": "hono",
  "operations": ["fetch", "generate"],
  "generate": {
    "prepare": [{"run": ["bun", "install", "--frozen-lockfile"]}],
    "run": ["bun", "run", "create:project", "--", "--out={output}"],
    "output": "{output}",
    "default_profile": "minimal",
    "profiles": [
      {"id": "minimal", "arguments": ["--profile=minimal"]},
      {"id": "authenticated", "arguments": ["--profile=authenticated"],
       "requires": {"capabilities": ["shared-backend"],
                    "unless_capabilities": ["anonymous-public-access"]}}
    ]
  },
  "managed_files": ["AGENTS.md"]
}
```

Rules:

- `run`/`prepare` are argv arrays only — never shell strings, no
  `bash -c`, validated by the same argv gate as setup/checks, with a
  mandatory timeout and an allowlisted environment.
- Placeholders are closed: `{name}` (logical surface name),
  `{project}`, `{surface}`, `{profile}`, `{output}` (sandbox location
  resolved at materialization time, never serialized in plans).
  Unknown placeholders fail before anything executes.
- `prepare` runs inside the factory workspace; `setup`/`checks` run
  against the staged generated output. The two are never confused.
- Pins are immutable: repo commit SHAs, exact tool versions
  (`ignite-cli@11.5.0`); no `@latest`, no floating branches.
- Generated foundations must run non-interactively (`--yes` or
  equivalent). Interactive-only generators are not selectable.
- Adapters declaring `fetch` acquire the pinned factory into the
  sandbox; adapters with only `generate` run an external pinned tool
  in a fresh sandbox directory.

## Profile selection

Profiles are foundation concepts (`minimal`, `authenticated`, ...) the
core never interprets. Each profile may declare `requires` over
structured decision signals — capabilities, surfaces, stable product
requirement ids — plus `unless_capabilities` exclusions. Profiles are
declared smallest-first; the richest fully-satisfied profile wins
(smallest valid profile). No free-text inspection, no LLM, no network.

Product requirements may optimize the configuration of an
already-selected foundation, but a missing product feature never
disqualifies it (`missing feature ≠ CatalogGap`): unmatched signals
fall back to a smaller profile and the feature stays pending for
Gentle. A hard `must-not-use` constraint that excludes the foundation
fails closed with an explicit configuration gap — never a silent
fallback.

## Staging and safety

Generation runs in a per-component sandbox under the fetch root.
The declared output must exist as a non-empty directory strictly
inside the sandbox (symlink escapes fail), and must not be the
factory root itself. Only that output is copied into the existing
project staging; the atomic rename commit is unchanged, so a failed
or timed-out generator leaves the destination untouched. Mixed
static/generated projects materialize component by component into
one staging tree with one handoff.

## Provenance, manifest, doctor, handoff

Every generated surface records foundation, repo, pin, strategy,
logical name, profile, resolved non-runtime arguments and adapter
fingerprint in `.engineering/provenance.json` (never temp paths,
never secrets). The manifest carries the same per-component
configuration; `eng doctor` reports `generation-config-drift` and
`provenance-component-mismatch` when plan, manifest and provenance
disagree. The handoff names strategy and profile per surface so
Gentle knows which foundation variant it owns.

Plan fingerprints bind strategy, name, profile, arguments and adapter
fingerprint: changing any of them changes the fingerprint. Schema v1
components without generation metadata read as `copy`.

## Curation

A selectable generated foundation needs, beyond the standard
evidence: validated generator command and version/pin, validated
profiles and default, validated output path, a non-interactive pilot,
a generated project that builds/tests, proof that factory files do
not leak, and preserved surface `AGENTS.md` behavior. A provider
converted from `copy` to `generate` stays `pilot-ready` until a real
generation pilot (minimal plus one richer profile) lands.

## Examples

- `hono-api` (`api-starter` factory): `fetch` + `bun install` +
  `create:project`, profiles `minimal → integration-platform`,
  default `minimal`.
- `ignite`: external pinned tool (`npx ignite-cli@11.5.0 new {name}
  --yes`), no profiles, output `{name}`.
- `next-admin` / `tanstack-admin`: still `copy` until their forks
  ship a validated non-interactive `generate:project --profile
  minimal` contract; the core release does not wait for them.
