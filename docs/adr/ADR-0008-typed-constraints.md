# ADR-0008 — Typed technical constraints

Date: 2026-09-08 · Status: accepted

Explicit user technology decisions carry a stable target
(`framework|language|runtime|database|deployment|provider`) alongside kind
and value. The resolver matches them against catalog technology metadata —
never brand literals in Go — with `must-use` eliminating incompatible
candidates and `prefer`/`avoid` scoring only. A mandatory database value
with no curated profile is an architectural `catalog-gap`
(`{kind: database-profile}`); a preferred one falls back with a recorded
deviation.

Consequence: `must-use framework=tanstack` rejects Next-based candidates
and `must-use database=turso` resolves a Turso profile or gaps honestly,
without hardcoding either name in the core.
