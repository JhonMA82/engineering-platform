# Handoff a Gentle AI

## Contexto breve

- Proyecto: `school-requests`
- Idea: Sistema interno para registrar, adjuntar y aprobar solicitudes escolares.
- Problema: Las solicitudes se pierden entre mensajes y no existe trazabilidad de responsables o decisiones.
- Recipe: `GP-02@1.0.0`
- Boilerplates: `tanstack-admin` → `apps/admin`, `hono-api` → `apps/api`
- Base de datos: `postgresql-managed`
- Patrones: `modular-monolith`, `explicit-contracts`, `least-privilege`, `incremental-delivery`, `single-tenant-first`
- Estado: `blueprint` · readiness `code-ready`

## Fuentes de verdad

- Idea y alcance: `.engineering/project-definition.json`
- Stack, skills, gates y ownership: `.engineering/project.json`
- Arquitectura: `ARCHITECTURE.md`
- Reglas: `AGENTS.md`
- Materialización, pins y verificación: `no materializado`

## Instrucciones de ejecución

1. Lee, en orden: `GENTLE.md`, `.engineering/project-definition.json`, `.engineering/project.json`, `ARCHITECTURE.md`, `AGENTS.md`.
2. Decide entre ejecución directa y SDD según riesgo, ambigüedad, contratos, datos y permisos; registra brevemente el motivo.
3. Conserva la Recipe, el stack, los patrones y las exclusiones; consulta las fuentes antes de desviarte.
4. Implementa el incremento vertical mínimo y ejecuta los quality gates indicados en `.engineering/project.json`.
5. Si readiness es `code-ready`, ejecuta `eng check --run` antes de tratar el proyecto como verificado.
