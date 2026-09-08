# ADR-0006 — Pi as adapter

Date: 2026-09-07 · Status: accepted

Pi is the conversational adapter: it produces `ProjectIntent` and renders
the core's neutral `unresolved_dimensions` as natural questions. It never
selects recipes or invents providers.

Consequence: ambiguous results carry dimension/reason/options triples for
Pi to ask; Gentle receives the consolidated decision, never the chat
history.
