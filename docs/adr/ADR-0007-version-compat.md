# ADR-0007 — Core/Catalog version compatibility

Date: 2026-09-08 · Status: accepted

The catalog declares `min_core_version` (and optional `max_core_version`
plus `schema_version`); the loader enforces them as a real semver gate at
load time (`catalog requires core >= X, running core is Y`). The binary
line lives in exactly one place (`internal/version.CoreVersion`, ldflags-
stampable) and `eng version` reports core, catalog, schema, commit and
build date separately.

Consequence: an incompatible catalog cannot load, while catalog curation
and core releases version independently.
