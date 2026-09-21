---
name: project-discovery
description: Turn a product idea into an Engineering Platform ProjectIntent and orchestrate eng resolve, plan, materialize and doctor. Use when starting a new project with /newproject.
---

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
- When `.engineering/bootstrap.json` exists (the current directory is an
  `eng init` workspace), NEVER derive `--output` from the intent/project
  name. Materialize into the workspace itself (`--output .`) and run
  `eng doctor --project .`. Creating a project-named subdirectory inside
  the workspace (for example `reloj_checador/reloj_checador_escolar/`) is
  forbidden; the core refuses it.
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
  → eng materialize --plan plan.json --output <dir> (. when eng init prepared it: bootstrap.json exists, never a project-named subdirectory)
  → eng doctor --project <dir> (. in an init workspace)
  → hand the directory to Gentle AI (GENTLE.md has the takeover steps)
  → clean up eng-owned bootstrap only (.pi/prompts/newproject.md,
    .pi/skills/project-discovery/, .engineering/bootstrap.json)
```

Recipe selection, scoring rationale and compatibility stay inside
`eng explain` output. Quote it; do not paraphrase it into new claims.

## Dev-only audit (opt-in, ENG_AUDIT=1)

When the user runs with `ENG_AUDIT=1`, stitch the conversational side of
the audit trail so the Go commands can complete it. Without the env gate
write nothing: auditing is strictly dev-only and off by default.

- Session: reuse `--audit-session <id>` when supplied; otherwise use the
  timestamp form `YYYYMMDD-HHMMSS` (UTC) and pass the same id to every
  `eng ... --audit --audit-session <id>` call.
- Base dir: `<workspace>/.engineering/audit/<session>/events.jsonl`.
  Create it with `mkdir -p`; never write inside the generated project.
- Log the idea first (`phase=idea/kind=idea`), every questionnaire turn
  (`phase=discovery/kind=question|answer`, including
  `unresolved_dimensions`), and each intent version
  (`phase=intent/kind=intent`). Truncate answers to 500 chars and never
  log secrets, tokens, passwords or API keys.
- Thread `--audit --audit-session <id>` through resolve/plan/materialize/
  doctor (or once via `eng start --audit --audit-session <id>`). The CLI
  finalizes `AUDIT.md` + `audit.json` next to `events.jsonl`.
