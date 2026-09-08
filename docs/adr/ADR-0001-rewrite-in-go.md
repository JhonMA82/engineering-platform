# ADR-0001 — Rewrite in Go

Date: 2026-09-07 · Status: accepted

The legacy Python assistant scripts are reference material, not architecture.
M1 rebuilds the deterministic core (intent → catalog → resolver → decision)
in Go for static typing, single-binary distribution and reproducible builds.

Consequence: no port of `scripts/eng.py`; behavior is re-specified by the
routing fixtures in `testdata/routing/`.
