# New project runbook (M3)

End-to-end: intent → resolve → plan → materialize → doctor → Gentle
handoff. Everything below runs offline using the fixture catalog
overlay; real boilerplates additionally need network for the pinned
`git clone`.

## 0. Build

```bash
go build -o /tmp/eng ./cmd/eng
```

## 1. Intent

Extract the admin-only intent from the routing fixture (or write your
own following `schemas/project-intent.schema.json`):

```bash
python3 -c "import json; d=json.load(open('testdata/routing/admin-only.json')); open('/tmp/intent.json','w').write(json.dumps(d['intent'], indent=2))"
```

## 2. Fixture catalog overlay (offline)

Point the curated providers at the local fixtures so no network is
needed. Read the expected pins from the fixture marker files:

```bash
export WEB_PIN=$(tr -d '\n' < testdata/fixtures/boilerplates/fixture-web/PIN)
export WEB_SRC=$PWD/testdata/fixtures/boilerplates/fixture-web
mkdir -p /tmp/fixture-catalog/boilerplates
cat > /tmp/fixture-catalog/metadata.json <<EOF
{"catalog_version":"0.0.0-runbook","min_core_version":"1.0.0","schema_version":1}
EOF
cat > /tmp/fixture-catalog/boilerplates/tanstack-admin.json <<EOF
{
  "id": "tanstack-admin",
  "repo": "https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard",
  "pin": "$WEB_PIN",
  "adapter": {"name": "tanstack", "operations": ["fetch", "copy"], "managed_files": ["AGENTS.md"]},
  "source": {"type": "local", "path": "$WEB_SRC"},
  "delivery_status": "stable",
  "decision_status": "curated",
  "provides": {"surfaces": ["web-admin"], "capabilities": []},
  "tech_tags": ["tanstack"],
  "included_features": ["crud", "auth"]
}
EOF
```

## 3. Resolve and plan (no filesystem writes)

`eng plan --json` shows per component which foundation, pin, strategy
(`copy` or `generate`), resolved profile and non-runtime arguments will
materialize, and where each will land — so `--dry-run` stays meaningful
for generated foundations (temporary sandbox paths are never shown).

```bash
/tmp/eng resolve --input /tmp/intent.json --catalog-dir /tmp/fixture-catalog
/tmp/eng plan --input /tmp/intent.json --catalog-dir /tmp/fixture-catalog --json > /tmp/plan.json
/tmp/eng resolve --input /tmp/intent.json --json --catalog-dir /tmp/fixture-catalog > /tmp/decision.json
```

## 4. Materialize

```bash
/tmp/eng materialize --plan /tmp/plan.json --intent /tmp/intent.json \
  --decision /tmp/decision.json --catalog-dir /tmp/fixture-catalog \
  --output /tmp/demo-proj
```

The output directory must be new or empty; re-running into it is
refused. Expected layout:

```text
/tmp/demo-proj/
├── AGENTS.md
├── ARCHITECTURE.md
├── GENTLE.md
├── apps/admin/            # fixture-web content + preserved AGENTS.md
└── .engineering/
    ├── project-intent.json
    ├── architecture-decision.json
    ├── materialization-plan.json
    ├── project-map.json
    ├── implementation-brief.md
    ├── handoff.json
    ├── project.json
    └── provenance.json
```

## 5. Doctor

```bash
/tmp/eng doctor --project /tmp/demo-proj
/tmp/eng doctor --project /tmp/demo-proj --json
```

A consistent project prints `doctor: /tmp/demo-proj is consistent` and
exits 0. Any error-severity finding exits non-zero.

## 6. Hand off to Gentle

Open `/tmp/demo-proj/GENTLE.md`: read the brief, the root `AGENTS.md`
router, `ARCHITECTURE.md` and `project-map.json`, then the surface
`AGENTS.md` before modifying a surface. Architecture, surface topology
and selected foundations are locked (`.engineering/handoff.json`).

## With real boilerplates

Drop `--catalog-dir` (or overlay only what you need) and ensure network
access: `eng materialize` clones each provider with
`git clone --depth 1 --branch <pin>` and verifies HEAD resolves to the
pin. Multi-surface intents (e.g. `admin-mobile`) additionally
materialize the shared `api` backend and emit `consumes` relationships.
