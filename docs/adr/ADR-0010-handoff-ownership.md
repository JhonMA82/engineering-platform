# ADR-0010 — Gentle ownership handoff

Date: 2026-09-08 · Status: accepted

Every materialization transfers ownership Pi → Engineering Platform →
Gentle through generated artifacts (`GENTLE.md`, `handoff.json`,
implementation brief, root `AGENTS.md` router, `project-map.json`). Gentle
decides direct implementation vs a product-scoped SDD session; the platform
always emits `implementation_mode: "gentle-decides"` and never prescribes
SDD. Architecture, topology, foundations and database profile are locked;
product questions travel in `open_product_questions`, separate from routing
ambiguity, under a no-repeat rule.

Consequence: Gentle takes over without re-eliciting recorded facts and
without reopening settled architecture absent a concrete contradiction.
