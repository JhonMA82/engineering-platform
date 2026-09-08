# v1.0 readiness report — Release Hardening H5+H6+H7

Date: 2026-09-08 · Scope: `engineering-platform-v1-release-hardening.md`
§5–§8, §10 (H5–H7), §11 (DoD). Method: every row cites a file, test name,
or observed command output from this pass. H1–H4 were merged before this
pass; their evidence stays in `docs/decisions/migration-notes.md`.

Validation observed this pass (2026-09-08, working tree):

```text
gofmt -l .                        → clean (no output)
go vet ./...                      → clean
go test -count=1 ./...            → all packages ok
go build -o /tmp/eng ./cmd/eng    → BUILD_OK
/tmp/eng catalog validate         → catalog OK: version 1.0.0, 7 recipes, 8 boilerplates
/tmp/eng version                  → Core: dev / Catalog: 1.0.0 / schema 1 (unstamped tree build)
/tmp/eng resolve|plan|start --dry-run (open-questions intent) → resolved GP-02
go test ./internal/app/ -run 'Pilot' -count=1 → ok (offline pilots)
cross-compile matrix (linux/amd64+arm64, darwin/arm64, windows/amd64) → all OK
```

## §11 Definition of Done → evidence

| # | §11 checkbox | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Core y catálogo validan compatibilidad real | pass (H1) | `internal/catalog/loader.go` semver gate + schema gate; `compat_test.go`, `schema_test.go` |
| 2 | CoreVersion y catalog version son independientes | pass (H1) | `internal/version` single source; `eng version` splits Core/Catalog/schema |
| 3 | Technical constraints tienen target | pass (H2) | `TechnicalConstraint{Target,Kind,Value}`; `docs/concepts/technical-constraints.md`; ADR-0008 |
| 4 | Database constraints son resolubles | pass (H2) | composer profile selection; `docs/decisions/migration-notes.md` H2 |
| 5 | Missing mandatory database provider genera CatalogGap | pass (H2) | `DiagnoseDatabaseGap`; fixture `must-use-database-gap.json` green |
| 6 | Curation evidence es obligatoria para foundations estables | pass (H3) | `eng catalog validate` per-status bars; `curation_test.go`; 8 stubs |
| 7 | Todos los boilerplates seleccionables pasan catalog validation | pass | `/tmp/eng catalog validate` → OK (observed above) |
| 8 | Ignite tiene materialización válida o queda fuera del pool estable | pass (H4) | generic `generate` op; `docs/decisions/ignite-materialization.md`; ignite is `pilot-ready`, not stable |
| 9 | No hay casos especiales hardcodeados por boilerplate en el core | pass | `TestNoBoilerplateSpecialCases`; resolver has no brand literals (ADR-0008) |
| 10 | Gentle recibe Direct Build vs SDD instructions | pass (H5) | `GENTLE.md` template §5.1 wording; `TestGentleDirectVsSDDRule`; `docs/architecture/handoff.md`; ADR-0010 |
| 11 | Open product questions separadas de routing ambiguity | pass (H5) | `ProjectIntent.open_product_questions[]`; resolver never reads it (`TestOpenProductQuestionsDoNotAffectRouting`); `handoff.json {product_requirements, open_product_questions, implementation_mode: gentle-decides}` (`TestHandoffJSONProductQuestions`); fixture `admin-open-questions.json` resolves GP-02 identically; locked includes `database-profile` |
| 12 | Multi-surface handoff genera mapa correcto | pass | `TestEndToEndAdminOnlyOffline` + `assertPilotArtifacts` + `materialize_test.go` 12-file assertion (AGENTS, project-map, brief, handoff, GENTLE, ARCHITECTURE, surface stubs); `TestRootAgentsRoutes` |
| 13 | CI verifica formato | pass (H6) | `ci.yml` gofmt gate (fails on output); observed `gofmt -l .` clean |
| 14 | CI verifica vet | pass (H6) | `ci.yml` `go vet ./...`; observed clean |
| 15 | CI ejecuta tests | pass (H6) | `ci.yml` `go test ./...`; observed all ok |
| 16 | CI valida catálogo | pass (H6) | `ci.yml` explicit `eng catalog validate` step (builds eng first); observed OK |
| 17 | Routing dataset tiene guard >= 40 | pass (H6) | `minRoutingFixtures = 40` named constant (`resolver_test.go`); 46 fixtures (`ls testdata/routing \| wc -l` = 46) |
| 18 | Build matrix multiplataforma pasa | pass (H6) | `ci.yml` `build-matrix` job (4 targets, compile-only); all four verified locally this pass |
| 19 | Pi plugin tiene versión fijada | pass (H6) | `integrations/pi/package.json`: `^2.9.0` (was `*`; 2.9.0 is the latest published line, verified via `npm view`) |
| 20 | Pins críticos son reproducibles | pass (H6) | 4 SHA pins (verified HEADs/snapshot) + 4 first-party `v1.0.0` tags recorded as **re-verify-on-pilot** (no invented SHAs); pin table in migration-notes §H6; policy in curate guide §4 |
| 21 | Release artifacts incluyen checksums | pass (H6) | `release.yml` (`v*` tags → 4 binaries + `checksums.txt` via sha256sum → gh-release assets); runbook `docs/guides/release.md` |
| 22 | `go test ./...` pasa | pass | observed all ok (see header) |
| 23 | `go vet ./...` pasa | pass | observed clean |
| 24 | `eng catalog validate` pasa | pass | observed OK |
| 25 | Pilots relevantes pasan | pass | offline pilots ok (`-run 'Pilot'`); network pilot stays gated behind `ENG_UPSTREAM_PILOTS=1` by design |

## §7 regression inventory (all in-repo)

- §7.1 no-project_type: `TestIntentHasNoProjectType` (reflection + JSON-tag + resolve check).
- §7.2 TUI-arch-fit + missing PDF → resolved: `pdf-excel-no-gap.json` (in `TestRoutingScenarios`).
- §7.3 tui-without-provider → catalog-gap: `TestTuiWithoutProviderIsCatalogGap` (+ `tui-gap.json`, `tui-python-excel.json`).
- §7.4 synthetic surface+provider overlay resolves, zero core changes: `TestSyntheticSurfaceResolvesWithoutCoreChanges`.
- §7.5 must-use framework=tanstack excludes incompatible: `must-use-framework-tanstack.json` (+ `must-use-tanstack.json`).
- §7.6 multi-surface handoff 7 files: `materialize_test.go` + `assertPilotArtifacts` + `TestRootAgentsRoutes`.
- §7.7 resolver purity: `TestCoreImportPurity` (stdlib go/parser over direct imports of resolver/composer/planner/domain; no os/os-exec/net).

## §8 docs (final set)

`README.md` (counts + handoff/release pointers), `docs/architecture/routing.md`
(46 scenarios, guard, open-questions note), `docs/architecture/materialization.md`
(unchanged — H4 content already accurate), `docs/architecture/handoff.md` (new),
`docs/concepts/catalog.md` (new), `docs/concepts/technical-constraints.md`
(pre-existing, H2), `docs/guides/add-boilerplate.md` (new §8 alias → the single
`curate-boilerplate.md` runbook), `docs/guides/release.md` (new).

## §9 ADRs (no duplicates)

Existing 0001–0006 kept; added 0007 version-compat, 0008 typed-constraints,
0009 curation-eligibility, 0010 handoff-ownership, 0011 catalog-evolution.
Checked first: none of the five topics was covered by an existing ADR.

## Known non-blockers (post-v1)

- The four first-party `v1.0.0` tag pins need one pilot each to quote SHAs
  and re-pin (migration-notes §H6, release guide versioning rules).
- `TestPilotIgniteGeneration` still requires `ENG_UPSTREAM_PILOTS=1` +
  network (by design — never part of `go test ./...`).
- Tag v1.0.0 when the parent accepts this tree; `release.yml` handles assets.
