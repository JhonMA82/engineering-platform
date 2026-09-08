# ADR-0011 — Catalog evolution independent from core releases

Date: 2026-09-08 · Status: accepted

The catalog evolves additively (new entries, surfaces, capabilities,
evidence links) without core changes or core releases: surfaces and
capabilities are open vocabulary validated by catalog data, overlays extend
the base without touching it, and `catalog_version` moves on curation
changes while `min_core_version` guards the floor. Renaming or removing an
id is the only breaking change and must be recorded in the migration notes.

Consequence: curation velocity never waits on the binary, and the binary
never needs a release to learn a new foundation.
