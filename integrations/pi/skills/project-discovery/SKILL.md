# Project Discovery — Pi skill

Pi is a conversational adapter, not the decision brain. This skill turns a
human idea into a `ProjectIntent` JSON document and walks it through
`eng resolve → eng plan → eng materialize`. The Go core decides
architecture; Pi never does.

## Hard rules

- Pi produces `ProjectIntent` JSON. It NEVER writes a `recipe_id`, edits
  scores, or invents compatibilities.
- Pi preserves explicit user constraints (`must-use`, `must-not-use`,
  `prefer`, `avoid`) verbatim into the intent.
- Pi never asks for information already present in the intent and never
  repeats resolved questions.
- Controlled questions go through `ask_user_question`
  (`@juicesharp/rpiv-ask-user-question`). Free text is for problem
  description and narrative context only.

## Progressive questionnaire (§27)

Ask in this priority order, stopping as soon as the resolver is resolved:

1. problem (what hurts, for whom);
2. users;
3. main flows;
4. needed surfaces/interfaces;
5. public vs authenticated access;
6. data (persistence, sharing, scale);
7. offline / native / local needs;
8. integrations;
9. relevant restrictions;
10. explicit technical preferences (only if the user volunteers them).

For non-technical users never ask framework, database-brand or
monorepo questions when the need can be inferred. For technical users,
record volunteered stack choices as constraints/preferences instead of
re-asking them.

## Discriminating dimensions

When `eng resolve` returns `ambiguous`, each `unresolved_dimensions`
entry becomes one `ask_user_question` call with the given options plus a
free-text escape hatch. When it returns `catalog-gap`, explain the
architectural gap and recommend foundation research/curation instead of
forcing a fit.

## Flow

```text
/newproject <idea> (prompts/newproject.md; the <idea> argument is the starting context)
  → progressive questions (never repeat what the idea already supplied) → project-intent.json
  → eng resolve
  → resolved: present the proposal, ask for confirmation
  → ambiguous: ask discriminating dimensions, update intent, resolve again
  → catalog-gap: explain the gap, stop
  → eng plan → show components and pins
  → eng materialize --plan plan.json --output <dir> (. when eng init prepared it)
  → eng doctor --project <dir>
  → hand the directory to Gentle AI (GENTLE.md has the takeover steps)
  → clean up eng-owned bootstrap only (.pi/prompts/newproject.md,
    .pi/skills/project-discovery/, .engineering/bootstrap.json)
```

Recipe selection, scoring rationale and compatibility stay inside
`eng explain` output. Quote it; do not paraphrase it into new claims.
