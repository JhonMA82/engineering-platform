# Handoff — Gentle ownership transfer (§5)

Every materialization ends with an ownership transfer:

```text
Pi → Engineering Platform → Gentle
```

The platform resolves architecture; Gentle implements product. The handoff
artifacts are generated purely by `internal/handoff` (rendered files, no
disk access — the materializer writes them) on every materialization:

| File | Role |
| --- | --- |
| `GENTLE.md` | Ownership instructions: read order + Direct vs SDD decision |
| `.engineering/handoff.json` | Machine-readable transfer: status, owner, locked, questions, mode |
| `.engineering/implementation-brief.md` | What is pending, provided, open and locked |
| `AGENTS.md` (root) | Router: surface → path → responsibility → instructions |
| `<surface>/AGENTS.md` | Per-surface stub (never overwrites foundation files) |
| `ARCHITECTURE.md` | Parts, relations, locked decisions |
| `.engineering/project-map.json` | Machine-readable surface routing (see `internal/project`) |

## Direct implementation vs SDD session

`GENTLE.md` gives Gentle exactly two paths after reading the brief, router,
architecture, project map and the relevant surface instructions:

- **A. Direct implementation** — when product requirements are sufficiently
  defined.
- **B. SDD session** — only when important product/domain rules remain
  undefined, and then strictly about product behavior, workflows, rules,
  permissions, edge cases and domain decisions.

The platform never sets `implementation_mode: sdd`; `handoff.json` always
carries `"implementation_mode": "gentle-decides"`. Gentle decides.

## Locked decisions

`handoff.json → locked` lists what Gentle must not rediscover:

```text
architecture, surface-topology, selected-foundations, database-profile
```

Reopen them only on a concrete contradiction. The no-repeat rule applies
both ways: Gentle must not ask the user to repeat anything already present
in the handoff.

## Open product questions vs routing ambiguity

Two question channels, never mixed:

- `open_product_questions[]` — product/domain unknowns from the intent
  (approval roles, report fields, cancellation behavior). They never affect
  routing (`unresolved_dimensions` stays 0) and flow intent → `handoff.json`
  → brief untouched.
- `open_questions` — the historic routing-ambiguity content from
  `unresolved_dimensions`, kept for compatibility.

The brief mirrors the split: **Pending product implementation**,
**Already provided by foundation**, **Open product questions** (product),
**Unresolved routing questions** (architecture), **Locked architecture
decisions**.
