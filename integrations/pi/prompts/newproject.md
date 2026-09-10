---
description: Start a new Engineering Platform project and hand it to a development agent
argument-hint: "<idea>"
---

# /newproject

Start a new project with Engineering Platform and hand it to Gentle AI.

The initial project idea arrives as the command argument:

```text
${ARGUMENTS:-<no idea supplied — ask the user for it>}
```

- If an idea is supplied above, treat it as the starting discovery
  context. Never ask again for information it already contains.
- If no idea was supplied, start by asking the user what they want to
  build (problem, users, main flows).

Prerequisites: the `ask_user_question` extension
(`pi install npm:@juicesharp/rpiv-ask-user-question`) for structured
questions. If it is unavailable, ask the same questions as plain chat
lists instead of guessing.

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
   When the workspace was prepared with `eng init`, materialize into the
   current workspace (`--output .`); otherwise use a new empty directory.
5. **Verify.** Run `eng doctor --project <dir>`; it must report a
   consistent project.
6. **Hand off.** Point Gentle AI at `<dir>/GENTLE.md`. Done: architecture
   is locked, product requirements are listed in
   `<dir>/.engineering/implementation-brief.md`.
7. **Clean up bootstrap.** After a successful `doctor`, remove only the
   eng-owned bootstrap resources (they created the project; they are not
   the product):
   - `.pi/prompts/newproject.md`
   - `.pi/skills/project-discovery/`
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
update the intent and start a new output directory — re-running into a
non-empty directory is refused by design.
