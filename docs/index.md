# Documentation

Three doors: **use** the platform, **understand** it, or **maintain** it.

## Use it

- [First project](getting-started/first-project.md) — install → `eng init`
  → Pi → `/newproject` → handoff. Start here.
- [CLI reference](reference/cli.md) — every command, flags, exit behavior.
- [Evolve a project](guides/evolve-project.md) — `surface add`,
  `extend`, `add`, `update`.
- [Dev-only audit trail](guides/audit.md) — `ENG_AUDIT=1` + `--audit`:
  idea → Q&A → commands → report (`AUDIT.md` + `audit.json`).
- [Changelog](../CHANGELOG.md) — release history from Git tags.

## Understand it

- [Architecture](architecture/core.md) — pipeline and layering; depth-first:
  [routing](architecture/routing.md),
  [composition](architecture/composition.md),
  [materialization](architecture/materialization.md),
  [generated foundations](architecture/generated-foundations.md),
  [handoff](architecture/handoff.md).
- [Catalog](concepts/catalog.md) — the declarative knowledge base.
- [Public contracts](concepts/public-contracts.md) — versioned
  `.engineering/` artifacts external tools (AiContext) may consume, plus
  the `.engineering/` ownership table.
- [Project intent](concepts/project-intent.md) — the only resolver input.
- [Surfaces](concepts/surface.md), [recipes](concepts/recipe.md),
  [boilerplates](concepts/boilerplate.md),
  [technical constraints](concepts/technical-constraints.md).
- [Decision records](adr/) — accepted architectural decisions (ADR-0001…
  ADR-0011).

## Maintain it

- [Curate a boilerplate](maintainers/curate-boilerplate.md) — upstream-first
  curation runbook (entry + adapter + evidence + tests).
- [Add a recipe](maintainers/add-recipe.md) — when a new architecture
  family is warranted.
- [Release](maintainers/release.md) — validation gates, tagging, binaries.
- [New-project pilot](maintainers/testing/newproject-pilot.md) — offline
  end-to-end verification with the fixture catalog.
- [Curation evidence](../catalog/curation/) — per-foundation stubs,
  linked from each boilerplate entry.

Agent integrations live in [integrations/](../integrations/) (Pi `/newproject` skill and
OpenCode equivalent); they are agent instructions, not reader docs.
