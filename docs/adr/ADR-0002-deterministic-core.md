# ADR-0002 — Deterministic core

Date: 2026-09-07 · Status: accepted

`internal/domain` and `internal/resolver` are pure: stable sorts, sha256
intent fingerprint, no time/rand/net/LLM/filesystem. The same intent over
the same catalog always yields the same decision, which makes routing
testable, auditable and safe to cache.

Consequence: infra imports (`os`, `os/exec`, `net/http`, `cobra`) are banned
in those packages; I/O lives in `internal/catalog`, `internal/app` and
`internal/cli`.
