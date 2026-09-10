# DoD audit — Engineering Platform 1.0 (§46, 28 criteria)

Date: 2026-09-08. Scope: PRD `engineering-platform-v1-prd-consolidado-pi-agent.md` §46
plus the §23 initial commands and the §55 doc list. Method: every row cites a
file, test name, or observed command output. Verdicts: **pass** (proven in
repo), **contract-ready** (repo artifacts complete; live Pi/Gentle session
behavior is human-side and cannot be observed from here), **partial**
(genuine gap with follow-up). No inflated passes: criteria 23–26 are
contract-ready, not observed.

How to re-verify: `go vet ./...`, `go test -count=1 ./...`,
`go build -o /tmp/eng ./cmd/eng`, `go test ./internal/app/ -run 'Pilot' -count=1 -v`.

## Result: 24 pass · 4 contract-ready · 0 partial · 0 fail

| # | §46 criterion | Status | Evidence |
| --- | --- | --- | --- |
| 1 | no depende de Python para `eng` | pass | `go.mod` (Go 1.27, zero third-party deps); `find . -name '*.py'` returns nothing; `go build ./...` green |
| 2 | `eng` compila como binario Go | pass | `go build -o /tmp/eng ./cmd/eng` → `BUILD_OK`; `cmd/eng/main.go` over `internal/cli` |
| 3 | el resolver no requiere `project_type` | pass | `internal/domain/intent.go` documents "intentionally no project_type field"; only mentions in repo are comments rejecting it; 42 routing fixtures resolve without it |
| 4 | Pi no selecciona Recipes | pass | `integrations/pi/skills/project-discovery/SKILL.md` → intent → `eng resolve` selects; resolver R1–R10 is pure core (`internal/resolver/`) |
| 5 | Pi usa `@juicesharp/rpiv-ask-user-question` en la integración oficial | pass | `integrations/pi/package.json` declares the dependency; `SKILL.md:16-17` and `prompts/newproject.md:15` route controlled questions through `ask_user_question`. Live Pi usage is human-side (see note) |
| 6 | catálogo, Surfaces y vocabulario no hardcodeados en el resolver | pass | `catalog/*.json` + `internal/catalog/loader.go` + `index.go`; resolver consumes them via `catalog.NewIndex`, no closed lists |
| 7 | nuevo boilerplate representable sin recompilar el core | pass | `MergeOverlay` replaces entries by id; pilots repoint `ignite` at a fixture source with zero core changes (`internal/app/pilots_test.go`); `docs/guides/add-recipe.md`, `docs/guides/curate-boilerplate.md`, `catalog/curation/` evidence |
| 8 | routing dataset ≥ 40 casos útiles | pass | `testdata/routing/` holds **42** fixtures, all asserted by `internal/resolver` tests |
| 9 | casos ambiguos producen dimensiones discriminantes | pass | `ambiguous-admin-multiapp.json`, `ambiguous-mobile-scope.json`, `ambiguous-offline-scope.json` → `unresolved_dimensions`; `TestPlanProjectRejectsUnresolved` |
| 10 | cada decisión explica selección y rechazos | pass | `resolver.Explain`, `eng explain --input decision.json`; `must_include_reasons` / `must_not_select` asserted per fixture |
| 11 | `catalog-gap` solo para gaps arquitectónicos | pass | `tui-gap.json` → catalog-gap; `pdf-excel-no-gap.json` resolves despite missing product features |
| 12 | features faltantes no descalifican foundation válida | pass | `pdf-excel-no-gap.json`; P12 bonus-only scoring (`internal/resolver`, R8) |
| 13 | composer valida combinaciones | pass | `CheckCompatibility` + `testdata/composition/` (**11** scenarios incl. collision, traversal); `eng plan` fails closed on mismatch |
| 14 | materialization plan reproducible | pass | `TestPlanDeterministic`, `TestPlannerGolden` (goldens regenerated via sanctioned `-update` after a whitespace-only drift; fingerprints unchanged), sha256 fingerprint binds intent + catalog |
| 15 | materializer trabaja con pins | pass | plan pins every provider; fetch verifies (`PIN` marker for `local`, tag-match for `git`); `docs/architecture/materialization.md` |
| 16 | `eng doctor` comprueba consistencia | pass | `internal/project/doctor.go`, `eng doctor [--project] [--json]`, exit non-zero on errors (`project.HasErrors`) |
| 17 | Golden Paths principales materializados en CI | pass | **new** `.github/workflows/pilots.yml` runs `go test ./internal/app/ -run 'Pilot'` offline: GP-02 single, GP-06 multi, GP-04 mobile-equivalent, surface-add evolution — each ends doctor-clean (`internal/app/pilots_test.go`) |
| 18 | multi-surface genera `AGENTS.md` raíz como router | pass | asserted in `TestEndToEndAdminOnlyOffline` + `assertPilotArtifacts`; observed at `/tmp/demo-proj/AGENTS.md` via `eng start` |
| 19 | cada Surface conserva su `AGENTS.md` especializado | pass | adapter `managed_files: ["AGENTS.md"]`; observed `/tmp/demo-proj/apps/admin/AGENTS.md` from fixture source |
| 20 | se genera `.engineering/project-map.json` | pass | asserted in e2e + pilots; observed in `/tmp/demo-proj/.engineering/` |
| 21 | se genera `.engineering/implementation-brief.md` | pass | same as 20 |
| 22 | se genera `.engineering/handoff.json` | pass | same as 20 |
| 23 | Pi ejecuta el flujo end-to-end | contract-ready | skill `project-discovery` + prompt `newproject.md` + all `eng` commands exist; a live Pi session was not observed here |
| 24 | Gentle recibe el proyecto sin re-describir la idea | contract-ready | handoff trio + `GENTLE.md` + stored intent produced deterministically; Gentle-side consumption is human-side |
| 25 | Gentle puede decidir direct build vs SDD | contract-ready | `GENTLE.md` + brief expose scope/foundation state for the call; the call itself is Gentle behavior |
| 26 | una sesión SDD no re-descubre decisiones cerradas | contract-ready | locked artifacts (decision json, provenance, project-map) persist; SDD session discipline is human-side |
| 27 | no existe un nuevo `eng.py` | pass | no Python anywhere in repo; single Go binary; CLI thin (`internal/cli`), effects only in `internal/materializer` |
| 28 | documentación refleja la arquitectura real | pass | full §55 list present (`README.md`, 4 concepts, 3+1 architecture incl. `composition.md`, 4 guides, 6 ADRs); README fixed this pass (stale `testdata/routing/admin-only/intent.json` path, new `start`/`catalog show`/`pilots` sections) |

## §23 initial commands (closed this pass)

| Command | Status |
| --- | --- |
| `eng version` (now: binary + catalog + min-core + schema + commit/date) | pass |
| `eng catalog` (list default) / `catalog show <id>` / `catalog validate` | pass — **new** (bare `catalog` previously errored) |
| `eng resolve --input` / `eng explain --input` | pass (pre-existing) |
| `eng plan --input` / `eng materialize --plan --output` | pass (pre-existing) |
| `eng start --intent --output [--catalog-dir] [--dry-run] [--json]` | pass — **new** P7 chain; `--dry-run` provably writes nothing (`TestStartDryRunWritesNothing`); unresolved exits non-zero with the typed composition error |
| `eng doctor` | pass (pre-existing; also chained by `start`) |

Note: §23 sketches `eng start <name>`; the implemented
`--intent/--output` form is the refined P7 contract from the delegation
(intent file in, project dir out) and is what the runbook documents.

## Adjacent §30 closures (this pass)

- `eng version` prints commit/date (ldflags-stampable `BuildCommit`/`BuildDate`,
  `unknown` when unstamped) alongside binary + catalog versions.
- Core now rejects unknown catalog schemas: `SupportedSchemaVersion = 1`
  enforced in `LoadDir` with a typed catalog error (`schema_test.go`).
  `min_core_version` semver-vs-binary enforcement is deliberately deferred:
  core reports `0.1.0` (pre-1.0) while the shipped catalog declares floor
  `1.0.0`; enforcing now would reject the in-repo catalog. Revisit at 1.0.

## Pre-existing issue found and fixed

- `TestPlannerGolden` failed before this pass: goldens carried single-line
  arrays while `json.MarshalIndent` emits multi-line. Content (incl. both
  fingerprints) was identical — pure formatting drift. Regenerated via the
  repo's own `-update` path and re-ran without it: green.
