# /newproject

Start a new project with Engineering Platform and hand it to a development agent.

The command accepts the initial project idea directly:

```text
/newproject Inventory system for three warehouses with mobile barcode scanning
```

Treat that argument as the starting discovery context. Never ask again
for information already supplied in the command.

## Steps

1. **Discover.** Follow the `engineering-project-discovery` skill: ask
   progressively (problem → users → flows → surfaces → access → data →
   offline/native → integrations → restrictions → volunteered stack
   preferences) and draft `project-intent.json`.
2. **Resolve.** Run `eng resolve --input project-intent.json`.
   - `resolved` → present the winning recipe and its reasons; ask for
     confirmation before writing anything.
   - `ambiguous` → ask each `unresolved_dimensions` entry with the given
     options plus a free-text escape hatch, update the intent, resolve again.
   - `catalog-gap` → explain which architectural need has no curated
     foundation and recommend research/curation. Stop here.
3. **Plan.** Run `eng plan --input project-intent.json --json > plan.json`
   and show the components (surface → destination ← provider@pin).
4. **Materialize.** Run
   `eng materialize --plan plan.json --intent project-intent.json
   --decision decision.json --output <dir>`
   (save the resolve output as `decision.json` first).
   When the workspace was prepared with `eng init`, materialize into the
   current workspace (`--output .`); otherwise use a new empty directory.
5. **Verify.** Run `eng doctor --project <dir>`; it must report a
   consistent project.
6. **Hand off.** Point the development agent at `<dir>/GENTLE.md`. Done:
   architecture is locked, product requirements are listed in
   `<dir>/.engineering/implementation-brief.md`.
7. **Clean up bootstrap.** After a successful `doctor`, remove only the
   eng-owned bootstrap resources (they created the project; they are not
   the product):
   - `.opencode/commands/newproject.md`
   - `.opencode/skills/engineering-project-discovery/`
   - `.engineering/bootstrap.json`
   `eng start` does this automatically. For the step-by-step commands,
   remove them here.
   Keep the generated application code, its agent tooling (`AGENTS.md`,
   `GENTLE.md`, surface files) and the minimal `.engineering/` provenance
   (`project.json`, `provenance.json`, plan, brief, handoff). If
   materialization or validation failed, clean up nothing: keep the state
   for diagnosis and retry.

Never write a recipe id by hand, never edit scores, never invent
compatibilities. If the user changes requirements after materialization,
update the intent and materialize into a new output directory — re-running
into a materialized project is refused by design.
