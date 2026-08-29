# Handoff a Gentle AI

## Idea confirmada

Sistema interno para registrar, adjuntar y aprobar solicitudes escolares.

## Stack, patrones y estructura

- Recipe: `GP-02@1.0.0`.
- Boilerplates: `tanstack-admin` (`curated`) y `hono-api` (`released`).
- Datos: `postgresql-managed`.
- Features: `auth`, `rbac`, `audit`, `observability`, `files` y `notifications`.
- Skills: `architecture-selector`, `project-bootstrap`, `contracts`, `authorization`, `database`, `security-review` y `gate-runner`.
- Patrones: monolito modular, contratos explícitos, mínimo privilegio y single-tenant-first.
- Código del proyecto: `apps/**`, `packages/**`, `src/**` y `tests/**`.
- Estado: `blueprint` porque este ejemplo es un fixture ligero; un bootstrap real queda `materialized`.

## Alcance mínimo

- Crear una solicitud y adjuntar archivos.
- Aprobar o rechazar con actor, fecha e historial.
- Consultar el estado y la trazabilidad.

## Ejecución

Lee la definición, el manifest, `ARCHITECTURE.md` y `AGENTS.md`. Decide entre ejecución directa y SDD según riesgo y ambigüedad, registra el motivo y conserva la Recipe. Ejecuta los quality gates del manifest antes de terminar.
