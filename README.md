# Engineering Platform

Sistema interno de ingeniería para estandarizar cómo la consultoría diseña, genera, desarrolla, verifica, despliega, documenta y evoluciona software asistido por IA.

> **Estado:** bootstrap / v0.1.0  
> **Última revisión:** 2026-08-07  
> **Fuente histórica deprecada:** https://github.com/JhonMA82/boilerplates-catalog

## Qué problema resuelve

La plataforma evita que cada proyecto dependa de prompts aislados, preferencias momentáneas o decisiones improvisadas del agente. Convierte el stack aprobado en **Golden Paths, feature packs, quality gates, conocimiento reusable y reglas de actualización**.

```text
Idea → Architecture Brief → Golden Path → Starter + features mínimas → AI Harness → Quality Gates → Deployment → Knowledge Capture → Platform Evolution
```

## Empieza aquí

- [Guía de inicio](docs/00-start-here/START_HERE.md)
- [Flujo completo de un proyecto](docs/03-project-lifecycle/project-lifecycle.md)
- [Cómo trabaja el equipo con IA](docs/02-team/team-operating-model.md)
- [Golden Paths](golden-paths/README.md)
- [Feature Packs](feature-packs/README.md)
- [Integración con el harness](harness/README.md)
- [Versionado y upgrades](docs/06-versioning/versioning-strategy.md)
- [Quality Gates](docs/07-quality/quality-gates.md)
- [Knowledge Loop](docs/10-knowledge/knowledge-loop.md)

## Stack congelado

| Necesidad | Camino principal |
|---|---|
| Sitio público | Stardrive |
| Admin / sistema operativo web | TanStack Admin |
| Next.js cuando exista razón concreta | Next Admin |
| Mobile | Ignite |
| Desktop | Tauri UI |
| Datos/documentos Python | SpeedPy |
| API TypeScript desacoplada | Consulting Hono API Starter |
| API/data/AI Python | FastAPI / Full Stack FastAPI cuando aplique |
| SaaS integrado | React Starter Kit / Open SaaS según caso |
| Automatización | n8n |
| Arquitectura multi-app | Patrón `apps/* + packages/*` inspirado en T3 Turbo |

Fuente estructurada: [`platform/catalog.json`](platform/catalog.json).

## Regla de oro

> Elegir la arquitectura más simple que cubra correctamente el problema y pueda mantenerse.

No agregar Hono, React, multi-tenancy, Redis, workers o microservicios por defecto. Cada capacidad debe justificar su existencia.

## Estructura

```text
engineering-platform/
├── platform/              # fuentes de verdad estructuradas
├── golden-paths/          # caminos aprobados
├── feature-packs/         # capacidades componibles
├── harness/               # contrato de integración con el harness
├── skills/                # catálogo y reglas de skills
├── canonical-examples/    # ejemplos que los agentes deben imitar
├── evals/                 # pruebas del harness/modelos
├── knowledge/             # problemas, soluciones y patrones validados
├── upgrades/              # recipes/codemods de actualización
├── docs/                  # documentación humana y técnica
├── templates/             # documentos y manifiestos
├── examples/              # ejemplos completos
└── scripts/               # validadores y utilidades
```

## Cómo usarlo en un proyecto

1. Crear un **Architecture Brief** usando `templates/architecture-brief.md`.
2. Seleccionar un Golden Path.
3. Crear `.engineering/project.json` con versiones y features.
4. Generar/clonar únicamente apps y features necesarias.
5. El harness lee `project.json`, `AGENTS.md` y el Golden Path.
6. Implementar mediante specs pequeñas y verificables.
7. Los quality gates determinan cuándo está listo.
8. Registrar decisiones en ADRs.
9. Desplegar usando el runbook del proyecto.
10. Al cerrar un bug o aprendizaje, decidir si pertenece solo al proyecto o si debe promoverse a `knowledge/` o a un feature pack.

## Qué NO es

- No es un mega-framework.
- No contiene código de clientes.
- No obliga a usar el mismo stack en todo.
- No sincroniza proyectos mediante merges permanentes con starters.
- No sustituye revisión humana, pruebas, seguridad ni operación.

## Validación

```bash
python scripts/validate_platform.py
```
