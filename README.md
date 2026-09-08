# Engineering Platform

Turn a software idea into a ready-to-develop project foundation.

Engineering Platform takes what you need — an admin dashboard, a public site,
a mobile app, an API — selects from previously evaluated boilerplates, decides
a suitable architecture, composes the pieces, and generates the project with
context so a development agent can continue without starting from zero.

## Why Engineering Platform?

Starting a new product usually means answering the same questions twice: which
stack to use, how to lay out the repository, and how to hand the result to
whoever builds the features. Engineering Platform answers the first two
deterministically and prepares the third:

- **You describe the need.** Users, interfaces, data, offline or native
  requirements, constraints.
- **It selects a foundation.** From a catalog of curated boilerplates with a
  known origin and immutable pin — not from whatever is trending this week.
- **It composes the architecture.** API, dashboard, web, mobile: the pieces
  your product actually needs, with conventional destinations.
- **It generates the project.** Code scaffolding plus agent context
  (`AGENTS.md`, `GENTLE.md`, an implementation brief, a handoff document), so
  development continues with a map instead of a blank page.

## How it works

```text
Idea
  ↓
Pi asks the necessary questions
  ↓
Engineering Platform selects the base
  ↓
composes API / dashboard / mobile / etc.
  ↓
generates the project
  ↓
Gentle AI continues development
```

Under the hood this maps to a deterministic pipeline
(`ProjectIntent → Resolver → Composer → Materializer → Development Handoff`),
described in [How it works internally](#how-it-works-internally).

## What Engineering Platform is not

Engineering Platform does **not** build your whole application:

- Engineering Platform: selects a foundation, composes the architecture,
  prepares the project, leaves instructions and context.
- Gentle AI (or your team): implements the product features, deepens
  requirements when needed.

If you expect a full-app generator, this is not it. It gives you the right
starting point — the features are still yours to build.

## Example

> "I need a system with an admin dashboard, an API, and a mobile app for
> recording field operations."

Pi captures the users, the interfaces, the offline/native needs, and the
restrictions. Engineering Platform can then resolve an API foundation, an
admin foundation, a mobile foundation, and a database profile, and lay out
the repository:

```text
services/api/
apps/admin/
apps/mobile/
```

Each surface keeps its own agent instructions, so whoever builds the field
 Flow continues from a map, not from guesses.

## Installation

### Prebuilt binary (recommended)

Every `v*` tag publishes binaries and a `checksums.txt` to the
[GitHub Releases page](https://github.com/JhonMA82/engineering-platform/releases).
Available assets:

```text
eng-linux-amd64
eng-linux-arm64
eng-darwin-arm64        # macOS Apple Silicon
eng-windows-amd64.exe
checksums.txt
```

**Linux (amd64):**

```bash
curl -L -o eng https://github.com/JhonMA82/engineering-platform/releases/latest/download/eng-linux-amd64
chmod +x eng
mkdir -p ~/.local/bin
mv eng ~/.local/bin/eng
eng version
```

Make sure `~/.local/bin` is on your `PATH`. As an alternative you may place
the binary in `/usr/local/bin` instead — `sudo` is not required when you use
`~/.local/bin`.

**Linux (ARM64):** same steps, downloading `eng-linux-arm64` instead.

**macOS (Apple Silicon):** same steps, downloading `eng-darwin-arm64`.
Intel macOS is not shipped (`darwin-amd64` is not built).

**Windows (amd64):** download `eng-windows-amd64.exe` from the releases page,
optionally rename it to `eng.exe`, place it in a directory on your `PATH`,
then run:

```text
eng version
```

**Checksums (recommended):** verify your download against `checksums.txt`
from the same release:

```bash
curl -L -O https://github.com/JhonMA82/engineering-platform/releases/latest/download/checksums.txt
sha256sum -c checksums.txt
```

### Build from source

For contributors and developers. Requirements: **Go 1.27+** and **Git**
(no Python needed).

```bash
git clone https://github.com/JhonMA82/engineering-platform.git
cd engineering-platform
go build -o eng ./cmd/eng
./eng version
```

## Quick Start

### Recommended: Pi workflow

The recommended experience uses [Pi](https://github.com/juicesharp/pi) as the
conversational interface: Pi turns your idea into a `ProjectIntent` document
that Engineering Platform can resolve. The integration lives in
[`integrations/pi/`](integrations/pi/) and works through structured questions
(`@juicesharp/rpiv-ask-user-question`), so you answer options instead of
writing JSON by hand.

1. Install `eng` (see [Installation](#installation)).
2. Set up the Pi integration from [`integrations/pi/`](integrations/pi/).
3. Describe your idea to Pi.
4. Run `/new-project` ([prompt](integrations/pi/prompts/new-project.md)):
   Pi refines the `ProjectIntent` with you.
5. Engineering Platform resolves, plans, and materializes the project.
6. Gentle AI takes over from `GENTLE.md`.

Details: [`integrations/pi/skills/project-discovery/SKILL.md`](integrations/pi/skills/project-discovery/SKILL.md)
and the [new-project runbook](docs/guides/new-project.md).

### CLI workflow

Without Pi, drive the same pipeline directly:

```bash
eng resolve --input project-intent.json
eng plan --input project-intent.json
eng materialize --plan materialization-plan.json --output ./my-project
eng doctor --project ./my-project
```

`eng start` is the short path chaining the same phases
(`resolve → plan → materialize → doctor`):

```bash
eng start --intent project-intent.json --output ./my-project
eng start --intent project-intent.json --dry-run   # inspect the plan, write nothing
```

`eng explain --input decision.json` prints why a decision resolved the way it
did. The manual JSON flow exists for scripting and debugging; for normal use,
prefer the Pi workflow above.

## What gets generated?

```text
my-project/
├── AGENTS.md
├── ARCHITECTURE.md
├── GENTLE.md
├── apps/
├── services/
└── .engineering/
    ├── project.json
    ├── provenance.json
    ├── project-map.json
    ├── materialization-plan.json
    ├── implementation-brief.md
    └── handoff.json
```

The concrete layout depends on the selected surfaces (`apps/` and `services/`
entries follow conventional destinations such as `services/api`,
`apps/admin`, `apps/mobile`, `apps/web`). `.engineering/` keeps the machine
state: manifest, provenance, project map, and the materialization plan the
project was built from.

## Handoff to Gentle AI

Engineering Platform does not just copy boilerplates — it leaves a map for
agents:

- Root `AGENTS.md` — where to work in the repository.
- Per-surface `AGENTS.md` — rules specific to each foundation.
- `GENTLE.md` — how to continue development.
- `.engineering/implementation-brief.md` — what is to be built.

Point your development agent at `GENTLE.md` and it starts with the
architecture locked and the product requirements listed.

## Curated foundations

A **curated boilerplate** is an evaluated, registered base with a known origin
and immutable pin, an adapter, and enough curation evidence for Engineering
Platform to use it reproducibly.

Main recipes in the default catalog:

| Recipe | What it covers | Primary foundation(s) |
| --- | --- | --- |
| GP-01 | Public website | stardrive |
| GP-02 | Authenticated backoffice | tanstack-admin |
| GP-03 | Python/data app | speedpy |
| GP-04 | Native mobile client + shared API | ignite, hono-api |
| GP-05 | Installed desktop tool, local database | tauri-ui |
| GP-06 | Multi-app (web + admin + API + PWA) | stardrive, tanstack-admin, hono-api, tanstack-transactional-pwa |
| GP-07 | Commercial SaaS | react-starter-kit |

Database profiles: `postgresql-managed` (shared multi-user backends) and
`sqlite-local` (single-node or local scope). The recipe policy picks the
default; explicit database constraints can steer it (see below).

## Technical constraints

Explicit user decisions travel as `{target, kind, value}` constraints, for
example `{target: "database", kind: "must-use", value: "sqlite-local"}`.
Hard constraints (`must-use`, `must-not-use`) can eliminate candidates; soft
ones (`prefer`, `avoid`) only affect ranking.

Supported hard targets in v1.0.1: `framework`, `language`, `runtime`,
`database`, `provider`. `deployment` supports **no** hard constraints yet —
the catalog curates no deployment metadata, so `must-use` / `must-not-use`
with `target: "deployment"` are rejected at validation
(`unsupported technical constraint target: deployment`) instead of being
silently ignored. `prefer` / `avoid` deployment wishes can still be recorded
as non-routing preferences. Deployment profiles (`edge`, `serverless`,
`container`, …) are future work.

Full contract: [docs/concepts/technical-constraints.md](docs/concepts/technical-constraints.md).

## When the catalog doesn't have a fit

Engineering Platform never forces a wrong solution. If it understands the
architectural need but no curated foundation matches, it returns a
`CatalogGap`:

```text
CatalogGap → research → curate/add a foundation → resolve again
```

A gap means architecture is missing from the catalog — not that your product
lacks a feature. Features like PDF export, Excel import, or QR codes are
product behavior: they belong to Gentle AI to implement, not to the catalog
to provide.

**Product feature vs architecture:** take *"a TUI that imports Excel and
generates PDFs"*. Engineering Platform decides the TUI foundation; Gentle
implements the Excel import, the PDF generation, and the business behavior.
If the chosen boilerplate already ships a feature, it is reused — but a
missing feature never triggers a different foundation.

## Extending the catalog

If tomorrow you need a foundation the catalog lacks (say, a TUI) and the
engine already understands the operations it uses, you can add it without
modifying or releasing a new core version: new boilerplate + catalog entry +
adapter + curation evidence + tests.

Short version (full guide: [docs/guides/add-boilerplate.md](docs/guides/add-boilerplate.md)):

1. Create `catalog/boilerplates/<id>.json`.
2. Define the repo and its immutable pin.
3. Declare surfaces and technical metadata.
4. Create the adapter.
5. Add curation evidence under `catalog/curation/`.
6. Run `eng catalog validate`.
7. Add a pilot/test.

Recipes live in `catalog/recipes/`. You do not need a new recipe per feature:
PDF, Excel, reports, or QR support are product features, not recipes
(see [docs/guides/add-recipe.md](docs/guides/add-recipe.md) for when a recipe
is actually warranted).

Catalogs compose: the default catalog plus an optional organization overlay
(`--catalog-dir`) for consultancies or companies with private boilerplates.

## Evolving an existing project

Generated projects keep evolving through the same deterministic pipeline:

```bash
eng surface add --project <dir> --surface <id> [--provider <boilerplate>]
eng extend --project <dir> --surface <id>
eng add --project <dir> --requirement "<text>" [--scope required_now|planned_later]
eng update --project <dir> [--json]
```

- `surface add` — adds a new architectural part (only the delta materializes;
  migrating to a different recipe aborts untouched, by design).
- `extend` — promotes a `planned_later` surface to `required_now`.
- `add` — records a product requirement; architecture is never touched.
- `update` — report-only pin comparison against the catalog; never mutates.

Runbook: [docs/guides/evolve-project.md](docs/guides/evolve-project.md).

## Commands

```bash
eng resolve --input intent.json [--json] [--verbose]
eng plan --input intent.json [--json]
eng materialize --plan plan.json --output <dir> [--intent intent.json] [--decision decision.json]
eng start --intent intent.json --output <dir> [--dry-run] [--json]
eng doctor --project <dir> [--json]
eng surface add --project <dir> --surface <id> [--provider <boilerplate>]
eng extend --project <dir> --surface <id>
eng add --project <dir> --requirement "<text>" [--scope required_now|planned_later]
eng update --project <dir> [--json]
eng explain --input decision.json
eng catalog [list]
eng catalog show <id>
eng catalog validate
eng version
```

All commands accept `--catalog-dir DIR` to use an organization overlay instead
of the default catalog.

## How it works internally

Engineering Platform is a deterministic local engine for selecting, composing,
and materializing curated stacks. Pure core, no network, no LLM calls: stable
sorts, pinned providers, and a fingerprint binding each plan to the intent and
the catalog version.

```text
ProjectIntent
    ↓
Resolver            → ArchitectureDecision
    ↓
Composer            → Composition
    ↓
Planner             → MaterializationPlan
    ↓
Materializer        → project + .engineering/ state
    ↓
Development Handoff → AGENTS.md / GENTLE.md / brief / handoff.json
```

Filesystem and process effects live only in the materializer; everything else
is pure and in-memory. Depth-first reading:

- [docs/architecture/core.md](docs/architecture/core.md) — pipeline and layering
- [docs/architecture/routing.md](docs/architecture/routing.md) — resolver rules
- [docs/architecture/composition.md](docs/architecture/composition.md) — composer and compatibility
- [docs/architecture/materialization.md](docs/architecture/materialization.md) — safety and pin policy
- [docs/architecture/handoff.md](docs/architecture/handoff.md) — Gentle ownership transfer

## Repository layout

- `catalog/` — catalog data (recipes, boilerplates, surfaces, capabilities, database profiles, curation evidence).
- `cmd/` — `eng` entrypoint.
- `internal/` — Go core (domain, resolver, composer, planner, materializer, project state, CLI).
- `integrations/pi/` — Pi conversational adapter (`/new-project`, discovery skill).
- `schemas/` — JSON schemas for intents, decisions, and plans.
- `docs/` — architecture, concepts, guides, and decision records.
- `testdata/` — routing, composition, and plan fixtures.

## Development

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
./eng catalog validate
```

`make build`, `make vet`, and `make test` wrap the first three.

Releases tag the core version (`eng version` reports binary, catalog,
commit, and date); the catalog carries its own revision in
`catalog/metadata.json` — core version ≠ catalog revision.

## Documentation

- Getting started: [docs/guides/new-project.md](docs/guides/new-project.md)
- Architecture: [docs/architecture/core.md](docs/architecture/core.md)
- Catalog concepts: [docs/concepts/catalog.md](docs/concepts/catalog.md),
  [recipes](docs/concepts/recipe.md),
  [boilerplates](docs/concepts/boilerplate.md),
  [technical constraints](docs/concepts/technical-constraints.md)
- Adding a boilerplate: [docs/guides/add-boilerplate.md](docs/guides/add-boilerplate.md)
  (curation: [docs/guides/curate-boilerplate.md](docs/guides/curate-boilerplate.md))
- Adding a recipe: [docs/guides/add-recipe.md](docs/guides/add-recipe.md)
- Project evolution: [docs/guides/evolve-project.md](docs/guides/evolve-project.md)
- Release process: [docs/guides/release.md](docs/guides/release.md)
- Pi integration: [integrations/pi/](integrations/pi/)
  ([discovery skill](integrations/pi/skills/project-discovery/SKILL.md))
- Materialization and safety: [docs/architecture/materialization.md](docs/architecture/materialization.md)

## Legacy

The original Python implementation is kept in
`JhonMA82/engineering-platform-legacy` as a historical reference.

## License

No license file is declared in this repository yet — check with the
maintainers before reuse beyond evaluation.
