# Engineering Platform

Sistema interno de ingeniería de la consultoría para convertir el stack aprobado, OpenCode y el harness en un proceso **repetible, versionado y verificable**.

> Versión bootstrap: **0.2.0**  
> Revisión: **2026-08-07**  
> Repositorio anterior: `JhonMA82/boilerplates-catalog` → **deprecated / reference only**.

## Para juniors: empieza aquí

No intentes leer todo el repositorio. Sigue esta ruta:

1. [Guía de inicio](docs/00-start-here/START_HERE.md)
2. [Glosario visual con ejemplo de cada concepto](docs/01-concepts/CONCEPTS_WITH_EXAMPLES.md)
3. [Ejemplo completo: sistema de solicitudes de una escuela](docs/13-examples/END_TO_END_SCHOOL_REQUESTS.md)
4. [Flujo de proyecto nuevo](docs/04-project-lifecycle/new-project.md)
5. [Cómo trabajar con OpenCode + harness](docs/06-ai-harness/harness-execution-model.md)
6. [Quality Gates](docs/08-quality/quality-gates.md)

## Idea central

```text
Problema del cliente
  ↓
Architecture Brief
  ↓
Golden Path
  ↓
Starter(s) + Feature Packs mínimos
  ↓
Project Manifest
  ↓
OpenCode + Harness + Skills
  ↓
Implementación
  ↓
Quality Gates
  ↓
Deploy / Operación
  ↓
Knowledge Loop
```

## Ejemplo en 30 segundos

Cliente: “Una escuela necesita registrar solicitudes internas y que el director las apruebe”.

La plataforma decide:

```text
Golden Path → GP-02 Admin Application
Frontend    → TanStack Admin
Backend     → Hono API
DB          → PostgreSQL
Features    → auth + rbac + audit
NO agregar  → multitenancy, mobile, Tauri, Redis, jobs
```

¿Por qué? Porque es una sola escuela. Si más tarde 20 escuelas comparten la misma instalación, se agrega el feature pack `multitenancy` mediante un upgrade plan; no se regenera el proyecto.

## Fuente de verdad

- `platform/catalog.json` — tecnologías aprobadas y referencias.
- `platform/golden-paths.json` — caminos de arquitectura.
- `platform/feature-packs.json` — capacidades opcionales.
- `.engineering/project.json` — cómo está compuesto cada proyecto real.

## Stack congelado

| Necesidad | Opción principal |
|---|---|
| Sitio público | Stardrive |
| Admin web | TanStack Admin |
| Next cuando existe razón concreta | Next Admin |
| Mobile | Ignite |
| Desktop | Tauri UI |
| Python datos/documentos | SpeedPy |
| API TypeScript | Hono Starter |
| API/data/AI Python | FastAPI cuando corresponda |
| SaaS opinionado | React Starter Kit / Open SaaS |
| Automatización | n8n |
| Multi-app | patrón `apps/* + packages/*` inspirado en T3 Turbo |

## Regla de oro

> Elige la arquitectura más simple que cubra el problema correctamente y pueda mantenerse.

Una tecnología nueva no se agrega porque sea popular. Debe cubrir una categoría que el stack no resuelve o demostrar una mejora material.

## Estructura

```text
platform/              fuentes de verdad
golden-paths/          caminos aprobados
feature-packs/         capacidades opcionales
harness/               contrato de integración con el harness
skills/                routing de skills
canonical-examples/    patrones que los agentes deben imitar
evals/                 pruebas del harness/modelos
knowledge/             problemas y soluciones comprobados
upgrades/              recipes para proyectos existentes
docs/                  guías humanas y técnicas
templates/             intake, brief, ADR, manifest
examples/              ejemplos completos para el equipo
```

## Validar

```bash
python scripts/validate_platform.py
```
