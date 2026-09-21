# Public contracts

This page is the integration contract between Engineering Platform and
repository-intelligence tooling (AiContext). Everything listed here is a
public artifact: it may be consumed by external tools. Anything not listed
here is internal and must not be coupled to.

## Generated project layout (`.engineering/`)

Every materialized project carries:

| Artifact | Owner | Schema | Notes |
|---|---|---|---|
| `project.json` | Engineering | `schemas/project-manifest.schema.json` (`schema_version` 1–2) | Reproducible record: no timestamps. v2 adds optional generation fields (`strategy`, `name`, `profile`, `arguments`, `adapter_fingerprint`). Readers must ignore unknown optional fields. |
| `provenance.json` | Engineering | `schemas/provenance.schema.json` (`schema_version` `"2"`) | Dated authorship record. The ONLY document allowed to carry timestamps (`materialized_at`, `events[].at`). v1 records carry no `components[]`; v2 adds them. |
| `project-map.json` | Engineering | `schemas/project-map.schema.json` (`schema_version` 1) | Machine-readable surface routing: the counterpart of the root `AGENTS.md` router table. `surfaces.<name> = {path, provider, instructions}`; `relationships[]` currently only `consumes` edges toward `api`. |
| `project-intent.json` | Engineering | `schemas/project-intent.schema.json` | Canonical copy of the resolved intent. |
| `architecture-decision.json` | Engineering | `schemas/architecture-decision.schema.json` | Canonical copy of the decision. |
| `materialization-plan.json` | Engineering | `schemas/materialization-plan.schema.json` | Canonical copy of the executed plan. |
| `implementation-brief.md` | Engineering | — (Markdown, human) | What to build next. Do NOT parse it for machine facts; the JSON copies above are authoritative. |
| `handoff.json` | Engineering | `schemas/development-handoff.schema.json` | Formal ownership transfer (`status: ready_for_implementation`, `next_owner: gentle-ai`, `locked: [architecture, surface-topology, selected-foundations, database-profile]`). |
| `runs/*.json` | Engineering | internal (`RunReport`, `schema_version: "1"`) | One report per command run. Diagnostic, not routing input. |
| `bootstrap.json` | Engineering (workspace only) | internal | `eng init` workspace state. Never present as project truth. |

AiContext-owned files that may coexist in the same directory (never written,
required, or deleted by Engineering operations): `aicontext.toml`,
`PROJECT_STATE.md`, `PATTERNS.md`, `consistency.yml`, `subprojects.yml`,
`rules/**`. The manifest file list (`project.json: files`) never includes
them; `eng doctor` never requires them; `eng evolve` / `materialize` /
`cleanup` never touch them.

## Version compatibility

- Consumers must read `schema_version` explicitly and accept only known
  versions (`project.json` 1–2, `project-map.json` 1, `provenance.json`
  `"2"` with forward tolerance for missing `components[]`).
- Unknown future versions must fail or degrade visibly — never be silently
  interpreted as a known version.
- Changes are additive: new optional fields, never renames of existing ones.
- No recipe coupling: consumers must work against `project-map.json` /
  `project.json` destinations, never against recipe ids (`GP-01`…`GP-07`).
  New recipes and surfaces work without consumer changes while they respect
  these contracts.

## `AGENTS.md` sharing

- The root `AGENTS.md` router table and per-surface stubs are
  Engineering-owned, except the marked AiContext blocks:
  `<!-- aicontext:routing:start -->…<!-- aicontext:routing:end -->` and
  `<!-- aicontext:context:start -->…<!-- aicontext:context:end -->`.
- Engineering evolution refreshes the router but preserves those blocks
  byte-identically (see `internal/project/ownership.go`).
- Surface stubs are never overwritten (`Overwrite: false`), so
  foundation-shipped instructions survive both evolution and AiContext.

## Doctor boundaries

- `eng doctor` answers: "is the materialized foundation still coherent with
  Engineering?" (manifest↔filesystem, map↔manifest, provenance↔manifest,
  plan copy).
- `aicontext check` answers: "is the repository context still coherent with
  reality?" It observes the same contracts but never reimplements `eng
  doctor`.
