---
name: engineering-project-discovery
description: Turn a product idea into an Engineering Platform ProjectIntent and orchestrate resolve, plan, materialize and doctor.
---

# Engineering Project Discovery — OpenCode skill

OpenCode is a conversational adapter, not the decision brain. This skill
turns a human idea into a `ProjectIntent` JSON document and walks it through
`eng resolve → eng plan → eng materialize`. The Go core decides
architecture; OpenCode never does.

## Hard rules

- OpenCode produces `ProjectIntent` JSON. It NEVER writes a `recipe_id`,
  edits scores, or invents compatibilities.
- OpenCode preserves explicit user constraints (`must-use`, `must-not-use`,
  `prefer`, `avoid`) verbatim into the intent.
- OpenCode never asks for information already present in the intent and never
  repeats resolved questions.
- The `/newproject` command argument is the starting discovery context.

## Progressive questionnaire

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

For non-technical users never ask framework, database-brand or monorepo
questions when the need can be inferred. For technical users, record
volunteered stack choices as constraints/preferences instead of re-asking
them.

## Discriminating dimensions

When `eng resolve` returns `ambiguous`, each `unresolved_dimensions` entry
becomes one question with the given options plus a free-text escape hatch.
When it returns `catalog-gap`, explain the architectural gap and recommend
foundation research/curation instead of forcing a fit.

## Flow

```text
/newproject <idea>
  → progressive questions (never repeat what the idea already supplied) → project-intent.json
  → eng resolve
  → resolved: present the proposal, ask for confirmation
  → ambiguous: ask discriminating dimensions, update intent, resolve again
  → catalog-gap: explain the gap, stop
  → eng plan → show components and pins
  → eng materialize --plan plan.json --output <dir> (. when eng init prepared it)
  → eng doctor --project <dir>
  → hand the directory to the development agent (GENTLE.md has the takeover steps)
  → clean up eng-owned bootstrap only (.opencode/commands/newproject.md,
    .opencode/skills/engineering-project-discovery/, .engineering/bootstrap.json)
```

Recipe selection, scoring rationale and compatibility stay inside
`eng explain` output. Quote it; do not paraphrase it into new claims.
