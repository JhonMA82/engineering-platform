# AGENTS.md — Engineering Platform

## Fuente de verdad
1. `platform/catalog.json`
2. `platform/golden-paths.json`
3. `platform/feature-packs.json`
4. `platform/compatibility.json`
5. `docs/decisions/*.md`

## Invariantes
- Selecciona la arquitectura más simple suficiente.
- No agregues una nueva tecnología si una adoptada resuelve el problema.
- Multi-tenancy es opcional.
- El frontend nunca es autoridad de autorización.
- Un proyecto generado es propietario de su código; no se mantiene con merges permanentes del starter.
- Los upgrades usan recipes y validación.
- Un bug corregido debe producir una prueba de regresión.
- Una solución específica de cliente no se promueve automáticamente a la plataforma.
- Los cambios de plataforma deben incluir documentación, tests/evals y compatibilidad.
- Los agentes no deciden que una tarea está terminada: los quality gates lo determinan.

## Flujo de cambio
Clasificar → ADR si aplica → cambio mínimo → validar → changelog → upgrade recipe si rompe → eval si afecta harness.

## Contexto para proyectos derivados
Leer primero `.engineering/project.json`, `AGENTS.md`, Golden Path, feature packs y ADRs. No cargar skills irrelevantes.
