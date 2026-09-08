# AGENTS.md — root router

This repo is a Go deterministic core. Route by task:

- Domain contracts → `internal/domain/`, decisions in `docs/adr/`.
- Catalog data or validation → `catalog/`, loader in `internal/catalog/`.
- Routing behavior → `internal/resolver/`, fixtures in `testdata/routing/`.
- Composition/planning → `internal/composer/`, `internal/planner/` (pure, no FS).
- Materialization → `internal/materializer/` (ONLY package with FS/process effects).
- Project state + doctor → `internal/project/`; agent context → `internal/handoff/`.
- CLI surface → `internal/cli/`, entrypoint `cmd/eng/main.go`.
- Schemas → `schemas/`.

Rules: keep `internal/domain` and `internal/resolver` free of infra imports
(`os`, `os/exec`, `net/http`, `cobra`). Mutating filesystem/process effects
(fetch, copy, command execution, project writes) live ONLY in
`internal/materializer`; `internal/project` (reads + pure builders +
doctor) and `internal/handoff` (pure render) never write. No `project_type`
as a routing key.
Product features never create catalog gaps. See `docs/adr/ADR-0001-rewrite-in-go.md`.
