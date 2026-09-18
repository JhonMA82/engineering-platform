# Changelog

All versions and dates below are derived from Git tags (`git tag`,
`git log <tag>..`) — no release history is invented here. Commit subjects
are reproduced verbatim from the tag ranges. Catalog revisions per release
were read with `git show <tag>:catalog/metadata.json`.

## Unreleased

Nothing yet.

## v1.2.4 — 2026-09-18

Documentation hardening, no behavior changes. Retired `docs/decisions/`
(migration notes, readiness audit, ignite research, bootstrap proposal)
after migrating durable rules into current guides; new docs landing page
(`docs/index.md`), first-project tutorial, CLI reference, and
`CONTRIBUTING.md`; curation runbook rewritten upstream-first; README
rewritten as the project landing page.

## v1.2.3 — 2026-09-09

Tag subject: `fix(generate): execute Materialization.Arguments and embed base catalog`.

- `af5578d` fix(generate): execute Materialization.Arguments and embed base catalog

## v1.2.2 — 2026-09-09

Tag subject: `fix(bootstrap): pass newproject idea via ARGUMENTS`.

- `49b95c3` fix(bootstrap): pass newproject idea via ARGUMENTS

## v1.2.1 — 2026-09-09

Tag subject: `fix(pi): require skill frontmatter and bump bootstrap to 1.0.1`.

- `6438bf9` fix(pi): require skill frontmatter and bump bootstrap to 1.0.1

## v1.2.0 — 2026-09-09

Catalog at this tag: `1.1.0` (released alongside the core).

- `5510ba4` feat(bootstrap): eng init with local PI/OpenCode integration
- `3fd24d9` chore: remove root handoff docs that should not live in repo
- `5958dfd` docs: add v1.0.1 fixes and v1.1.0 generated foundations handoffs
- `350af9c` chore(catalog): release catalog 1.1.0 alongside core v1.1.0

## v1.1.0 — 2026-09-09

Catalog at this tag: `1.0.0`.

Generated-foundations release (copy vs generate strategies, profiles,
sandboxed execution, generation metadata in manifest/provenance).

- `115e5f9` docs(catalog): record hono-api generation pilots
- `adea6f9` test(planner): regenerate v2 golden plans
- `8cfbf5c` docs(catalog): record hono-api generation contract evidence
- `57a35d6` docs: document generated foundations and curation workflow
- `664c679` feat(cli): show generation strategy and profiles
- `ca01d51` refactor(catalog): materialize hono-api through its generator
- `acc9a48` feat(evolve): refresh generation records on architecture evolution
- `540555b` feat(handoff): expose generation strategy and profile context
- `94dc771` feat(project): record generation metadata in manifest and provenance
- `4f55af3` feat(materializer): execute generated foundations in isolated sandboxes
- `b9b0200` feat(planner): plan v2 with resolved generation configuration
- `addd266` feat(foundationconfig): resolve deterministic generation configuration
- `d4a09f5` feat(domain): model generated foundation contracts
- `1a0ca55` docs(readme): rewrite landing page and document v1.0.1 contracts
- `40be740` fix(domain): reject unevaluated deployment hard constraints
- `4347042` fix(composer): enforce must-not-use database constraints

## v1.0.0 — 2026-09-07

Catalog at this tag: `1.0.0`.

Initial Go deterministic-core release: intent/resolver/composer/planner/
materializer pipeline, declarative catalog, CLI, Pi adapter, handoff.

- `fix(catalog)`: remove superseded stardrive-public-web entry
- `docs(catalog)`: document legacy provenance and identity rule
- `test(catalog)`: legacy migration integrity baseline and upstream pin pilot
- `feat(domain)`: provenance metadata with catalog-only validation relief
- `fix(materializer)`: fetch immutable SHA pins and read legacy manifest alias
- `fix(recipes)`: use canonical stardrive provider id
- `feat(catalog)`: migrate omitted legacy entries next-admin, fastapi and goship
- `fix(catalog)`: restore canonical legacy boilerplate sources and historical pins
- `docs`: normalize markdown table formatting
- `test(routing)`: minimum dataset guard and architecture regression tests
- `docs`: hardening ADRs, catalog concept, release runbook and readiness audit
- `ci(release)`: format gate, build matrix, release workflow and pinned dependencies
- `feat(handoff)`: direct-build vs SDD contract with open product questions
- `feat(cli)`: wire curation validation into catalog validate and show
- `feat(materializer)`: generic generate operation for generator-based foundations
- `feat(curation)`: formal evidence links with per-status validator enforcement
- `feat(constraints)`: typed technical constraints with catalog-driven matching
- `feat(version)`: canonical CoreVersion with semver catalog compatibility gate
- `docs`: concepts, guides, ADRs, migration notes and DoD audit
- `feat(cli)`: resolve, plan, materialize, doctor, start, catalog and evolution commands
- `feat(pi)`: conversational adapter with project discovery skill
- `feat(handoff)`: project manifest, doctor and Gentle handoff context
- `feat(materializer)`: pinned-source materialization with safety checks
- `feat(composer,planner)`: reproducible composition and materialization plan
- `feat(resolver)`: deterministic R1-R10 routing with scenario dataset
- `feat(catalog)`: declarative catalog with loader, validator and index
- `feat(domain)`: project intent and architecture decision contracts
- `chore`: bootstrap Go module, Makefile and CI
