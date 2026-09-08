# /new-project

Start a new project with Engineering Platform and hand it to Gentle AI.

## Steps

1. **Discover.** Follow `skills/project-discovery/SKILL.md`: ask
   progressively (problem → users → flows → surfaces → access → data →
   offline/native → integrations → restrictions → volunteered stack
   preferences) and draft `project-intent.json`.
2. **Resolve.** Run `eng resolve --input project-intent.json`.
   - `resolved` → present the winning recipe and its reasons; ask for
     confirmation before writing anything.
   - `ambiguous` → ask each `unresolved_dimensions` entry via
     `ask_user_question`, update the intent, resolve again.
   - `catalog-gap` → explain which architectural need has no curated
     foundation and recommend research/curation. Stop here.
3. **Plan.** Run `eng plan --input project-intent.json --json > plan.json`
   and show the components (surface → destination ← provider@pin).
4. **Materialize.** Run
   `eng materialize --plan plan.json --intent project-intent.json
   --decision decision.json --output <dir>`
   (save the resolve output as `decision.json` first).
5. **Verify.** Run `eng doctor --project <dir>`; it must report a
   consistent project.
6. **Hand off.** Point Gentle AI at `<dir>/GENTLE.md`. Done: architecture
   is locked, product requirements are listed in
   `<dir>/.engineering/implementation-brief.md`.

Never write a recipe id by hand, never edit scores, never invent
compatibilities. If the user changes requirements after materialization,
update the intent and start a new output directory — re-running into a
non-empty directory is refused by design.
