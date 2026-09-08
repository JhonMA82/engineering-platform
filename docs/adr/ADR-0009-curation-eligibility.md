# ADR-0009 — Curated foundation eligibility

Date: 2026-09-08 · Status: accepted

No foundation is selectable for materialization without a formally linked
curation record (`curation: {status, evidence}` confined to the catalog).
`catalog-only` may exist without full evidence but is not selectable;
`pilot-ready` requires license/repo/pin/adapter/basic evidence; `curated`
and `released` additionally require a recorded successful pilot.
`eng catalog validate` enforces the bars; downgrade on doubt, never upgrade.

Consequence: an unpiloted foundation can never silently enter the stable
pool, and adding evidence stays a data-only operation.
