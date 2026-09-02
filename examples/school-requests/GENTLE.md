# Handoff a Gentle AI

## Contexto

- Proyecto: `school-requests`
- Recipe: `GP-02@1.0.0`
- Boilerplates: `hono-api` → `services/api`, `tanstack-admin` → `apps/admin`
- Base de datos: `postgresql-managed`
- Patrones: `modular-monolith`, `explicit-contracts`, `least-privilege`, `incremental-delivery`, `single-tenant-first`
- Estado: `blueprint` · readiness `code-ready`
- Capacidades pendientes de implementación: `files`, `notifications`

## Fuentes de verdad

- Idea y alcance: `.engineering/project-definition.json`
- Stack, skills, gates y ownership: `.engineering/project.json`
- Arquitectura: `ARCHITECTURE.md`
- Reglas: `AGENTS.md`
- Materialización, pins y verificación: `no materializado`

## Instrucciones de ejecución

1. Lee, en orden: `GENTLE.md`, `.engineering/project-definition.json`, `.engineering/project.json`, `ARCHITECTURE.md`, `AGENTS.md`. La idea no se duplica aquí.
2. Decide entre ejecución directa y SDD según riesgo, ambigüedad, contratos, datos y permisos; registra brevemente el motivo.
3. Conserva la Recipe, el stack, los patrones y las exclusiones; consulta las fuentes antes de desviarte.
4. Implementa únicamente las capacidades con estado `pending-implementation` que pertenezcan al incremento solicitado; no asumas que una feature declarada ya existe.
5. Implementa el incremento vertical mínimo y ejecuta los quality gates indicados en `.engineering/project.json`.
6. Si readiness es `code-ready`, ejecuta `eng check --run` antes de tratar el proyecto como verificado.
