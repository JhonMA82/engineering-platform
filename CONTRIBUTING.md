# Contributing

## Build and verify

Requirements: Go 1.27+ and Git.

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
./eng catalog validate
```

`make build`, `make vet`, and `make test` wrap the first three. Keep the
tree green locally first (the cross-compilation matrix runs in CI).

## Repository layout

- `catalog/` — catalog data (recipes, boilerplates, surfaces,
  capabilities, database profiles, curation evidence).
- `cmd/` — `eng` entrypoint.
- `internal/` — Go core (domain, resolver, composer, planner,
  materializer, project state, CLI).
- `integrations/pi/` — Pi conversational adapter (`/newproject`,
  discovery skill).
- `schemas/` — JSON schemas for intents, decisions, and plans.
- `docs/` — architecture, concepts, guides, and decision records.
- `testdata/` — routing, composition, and plan fixtures.

## Development rules

Route by task (full router: `AGENTS.md`):

- Domain contracts → `internal/domain/`, decisions in `docs/adr/`.
- Filesystem/process effects live **only** in `internal/materializer`;
  `internal/domain` and `internal/resolver` stay free of infra imports
  (`os`, `os/exec`, `net/http`, `cobra`).
- `internal/project` (reads + pure builders + doctor) and
  `internal/handoff` (pure render) never write.
- No `project_type` as a routing key; product features never create
  catalog gaps.

Releases tag the core version (`eng version` reports binary, catalog,
commit, and date); the catalog carries its own revision in
`catalog/metadata.json` — core version ≠ catalog revision. Runbook:
`docs/maintainers/release.md`.
