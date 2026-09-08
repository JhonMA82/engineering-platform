# ADR-0005 — Resolution/materialization separation

Date: 2026-09-07 · Status: accepted

The resolver decides the recipe and foundation fit; it performs no side
effects. Composition, planning and materialization are later stages that
consume the `ArchitectureDecision`, never the raw conversation.

Consequence: M1 exposes only `resolve`, `explain`, `catalog validate` and
`version`. `catalog-gap` fires solely for missing architectural
foundations; missing product features are Gentle's implementation scope and
at most a small scoring tie-breaker.
