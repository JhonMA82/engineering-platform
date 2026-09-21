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
- When `.engineering/bootstrap.json` exists (the current directory is an
  `eng init` workspace), NEVER derive `--output` from the intent/project
  name. Materialize into the workspace itself (`--output .`) and run
  `eng doctor --project .`. Creating a project-named subdirectory inside
  the workspace (for example `reloj_checador/reloj_checador_escolar/`) is
  forbidden; the core refuses it.

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
  → eng materialize --plan plan.json --output <dir> (. when eng init prepared it: bootstrap.json exists, never a project-named subdirectory)
  → eng doctor --project <dir> (. in an init workspace)
  → hand the directory to the development agent (GENTLE.md has the takeover steps)
  → clean up eng-owned bootstrap only (.opencode/commands/newproject.md,
    .opencode/skills/engineering-project-discovery/, .engineering/bootstrap.json)
```

Recipe selection, scoring rationale and compatibility stay inside
`eng explain` output. Quote it; do not paraphrase it into new claims.

## Dev-only audit (opt-in, ENG_AUDIT=1)

When the user runs with `ENG_AUDIT=1`, stitch the conversational side of
the audit trail so the Go commands can complete it. Without the env gate
write nothing: auditing is strictly dev-only and off by default.

- Session: reuse `--audit-session <id>` when the user supplied one;
  otherwise use the timestamp form `YYYYMMDD-HHMMSS` (UTC) and pass the
  same id to every `eng ... --audit --audit-session <id>` call.
- Base dir: `<workspace>/.engineering/audit/<session>/events.jsonl`.
  Create it with `mkdir -p`; never write inside the generated project
  output (audit lives in the workspace that ran the commands).
- Log the idea first: one JSON line with
  `{"phase":"idea","kind":"idea","summary":"<first 160 chars of $ARGUMENTS>"}`.
- Log every questionnaire turn as it happens: one line per question
  (`{"phase":"discovery","kind":"question","summary":"<question>"}`) and
  one per answer (`{"phase":"discovery","kind":"answer","summary":"<answer>"}`).
  Discriminating dimensions from `unresolved_dimensions` log the same way.
- Log each intent version after writing `project-intent.json`:
  `{"phase":"intent","kind":"intent","summary":"intent v<n>: <name>"}`.
- Never log secrets, tokens, passwords or API keys: truncate answers to
  500 chars and drop anything matching token/secret/password/api_key.
- Thread the flags through the whole flow:
  `eng resolve --audit --audit-session <id>`,
  `eng plan --audit --audit-session <id>`,
  `eng materialize --audit --audit-session <id>`,
  `eng doctor --audit --audit-session <id>`
  (or once via `eng start --audit --audit-session <id>`).
  The CLI appends command/result events and finalizes
  `AUDIT.md` + `audit.json` next to `events.jsonl`.
