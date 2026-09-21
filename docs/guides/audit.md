# Dev-only audit trail

Trace one `idea → discovery → resolve → plan → materialize → doctor`
journey into a human report plus machine JSON. Strictly for development:
off by default, gated by `ENG_AUDIT=1`, never baked into generated
projects, never affecting fingerprints.

## Enable

```bash
export ENG_AUDIT=1
eng start --intent intent.json --output . --audit --audit-session demo-1
# → .engineering/audit/demo-1/AUDIT.md
# → .engineering/audit/demo-1/audit.json
# → .engineering/audit/demo-1/events.jsonl
```

- `--audit` opts one invocation into auditing; without `ENG_AUDIT=1` it
  warns and writes nothing.
- `--audit-session <id>` stitches skill + CLI events into one session.
  Default is a UTC timestamp; pass an explicit id to correlate retries.
- `--audit-dir <dir>` overrides the base (default: workspace or output
  dir that already exists, otherwise `.`).
- Per-command auditing works too:
  `eng resolve|plan|materialize|doctor --audit --audit-session <id>`.
  `eng start --audit` is the stitched full journey.

## What it records

| Phase | Who writes | Content |
|---|---|---|
| `idea` | skill | initial `$ARGUMENTS` text (truncated, no secrets) |
| `discovery` | skill | each questionnaire question + answer, including `unresolved_dimensions` loops |
| `intent` | CLI + skill | intent snapshot per version (name, problem, surface count) |
| `resolve` | CLI | argv, decision status (`resolved`/`ambiguous`/`catalog-gap`), recipe, intent fingerprint, counts |
| `plan` | CLI | argv, recipe, plan fingerprint, components, database profile |
| `materialize` | CLI | argv, manifest project/recipe/fingerprint, file counts |
| `doctor` | CLI | argv, findings count, status |
| `done` | CLI (`start`) | final status |

Every `audit.json` carries `schema_version`, `session_id`,
`started_at`, `core_version`, `catalog_version` and the ordered `events`
(`seq`, `at`, `phase`, `kind`, `summary`, `command`, `argv`, `status`,
`data`). `AUDIT.md` renders the same session as Timeline + per-phase
detail + reproduction hints. `events.jsonl` is the append-only log both
sides share.

## Skill side (questionnaire)

The `/newproject` skill owns discovery logging: idea first, then one
JSON line per question/answer/intent version into
`.engineering/audit/<session>/events.jsonl`. Answers truncate to 500
chars; anything matching token/secret/password/api_key is dropped. The
CLI never invents Q&A: without skill events the report still covers
resolve → doctor, but discovery stays empty.

## Guarantees

- **Off by default:** no env gate → zero audit writes, zero overhead.
- **Determinism preserved:** timestamps live only in audit files
  (like `provenance.json` and `runs/`); decisions and plans stay
  timestamp-free and fingerprints unchanged.
- **Layering:** types + redaction in `internal/project` (pure),
  markdown in `internal/handoff` (pure render), writes only in
  `internal/materializer`; `internal/domain` and `internal/resolver`
  untouched.
- **Secrets:** `RedactValue` replaces secret-looking keys/values with
  `[redacted]` at both builder and writer boundaries; fingerprints and
  pins survive.
- **Bootstrap safety:** `eng start` cleanup never deletes
  `.engineering/audit/`; a failed run still finalizes its tail so the
  report names the failing phase.

## Example

```bash
export ENG_AUDIT=1
export AUDIT=demo-1
mkdir -p .engineering/audit/$AUDIT
echo '{"phase":"idea","kind":"idea","summary":"tienda para panadería"}' >> .engineering/audit/$AUDIT/events.jsonl
eng start --intent intent.json --output . --audit --audit-session $AUDIT
cat .engineering/audit/$AUDIT/AUDIT.md
```

## Troubleshooting

- `audit: disabled — set ENG_AUDIT=1…`: you passed `--audit` without
  the env gate. Export it and re-run.
- Empty Discovery section: the skill did not log (prod run or manual
  CLI without agent). Re-run via `/newproject` with the gate open.
- `audit: could not …`: best-effort warning only; the run result stands.
  Check permissions on `.engineering/audit/`.
