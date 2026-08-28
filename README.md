# Engineering Platform

Una base de ingeniería minimalista, práctica y opinionada para que una consultoría de **1 a 20 personas** inicie y mantenga proyectos sin volver a discutir el stack desde cero.

> Versión: **0.3.0** · Revisión: **2026-08-28**

La idea es “Omarchy para proyectos”: pocos defaults buenos, comandos directos y automatización visible. No es un portal, un framework universal ni una colección de repositorios de moda.

## Qué resuelve

Un intake pequeño se convierte en una **Project Recipe** reproducible:

```mermaid
flowchart TD
    A["Intake del proyecto"] --> B["Resolver"]
    B --> C["Project Recipe"]
    C --> D["Blueprint o starter liberado"]
    D --> E["Skills y Quality Gates"]
    E --> F["Manifest y conocimiento"]
```

La Recipe fija:

- Golden Path y versión;
- boilerplates y estrategia de actualización;
- perfil de base de datos;
- feature packs incluidos y excluidos;
- skills que el asistente debe cargar;
- quality gates que deben aportar evidencia;
- ownership de archivos para evitar sobrescrituras.

La misma automatización sirve a una persona y a veinte. Lo que cambia con el tamaño es la revisión humana, no la arquitectura de la plataforma.

## Inicio rápido

No hay dependencias externas: requiere Python 3.11 o superior.

```bash
make check
./eng recommend --input examples/intakes/school-requests.json
./eng new --from examples/intakes/school-requests.json --output /tmp/school-requests
./eng doctor --project /tmp/school-requests
./eng plan --project /tmp/school-requests --change-type permission
./eng add api-keys --project /tmp/school-requests
```

Mientras un starter no tenga artefacto interno `released`, `eng new` crea un **blueprint**: decisiones y contexto correctos, sin fingir que existe código productivo curado.

## Ejemplo de decisión

Necesidad: “Una escuela necesita solicitudes internas, adjuntos y aprobación”.

| Dimensión | Resultado |
|---|---|
| Recipe | `GP-02@1.0.0` |
| Starters | `tanstack-admin` + `hono-api` |
| Datos | `postgresql-managed` |
| Features | `auth`, `rbac`, `audit`, `observability`, `files` |
| Skills | selector, bootstrap, autorización, datos y gates |
| Exclusiones | multitenancy, jobs y mobile hasta que exista una necesidad |
| Entrega actual | `blueprint`, porque Hono API aún no está liberado |

## Catálogo sin duplicados

El catálogo anterior fue integrado, no descartado. Las 17 entradas y sus fichas se conservan en [`catalog/legacy-v1.2.3`](catalog/legacy-v1.2.3/README.md); el registro normalizado vigente es [`platform/boilerplates.json`](platform/boilerplates.json).

Cada entrada tiene dos estados independientes:

| Eje | Valores | Pregunta que responde |
|---|---|---|
| Decisión | default, alternative, specialized, experimental, reference, deprecated, rejected | ¿Conviene para este problema? |
| Entrega | catalog-only, pilot-ready, curated, released | ¿Qué tan reproducible y probado está? |

Para proponer una URL:

```bash
./eng boilerplate evaluate https://github.com/kriasoft/react-starter-kit
```

Resultado actual: `ALREADY_REGISTERED`. React Starter Kit conserva una sola entrada, pin al upstream revisado y actualización `merge-seed`. Si se observa un commit distinto:

```bash
./eng boilerplate evaluate \
  https://github.com/kriasoft/react-starter-kit \
  --observed-commit <commit>
```

El resultado pasa a `ALREADY_REGISTERED_REFRESH`, no a una entrada duplicada. El skill [`boilerplate-curator`](.opencode/skills/boilerplate-curator/SKILL.md) guía la inspección, comparación y registro completo.

## Golden Paths vigentes

| Recipe | Default | Canal |
|---|---|---|
| GP-01 Public Web | Stardrive | stable |
| GP-02 Admin Application | TanStack Admin + Hono API | stable |
| GP-03 Python Data | SpeedPy | stable |
| GP-04 Mobile | Ignite + Hono API | stable |
| GP-05 Desktop | Tauri UI | stable |
| GP-06 Multi-App | TanStack Admin + Hono API | stable |
| GP-07 Commercial SaaS | React Starter Kit | trial |
| GP-08 Governed AI Assistant | AI Assistant Starter | trial |

El resolver incluye alternativas, datos permitidos, features, skills, gates y exclusiones; esta tabla es solo el mapa rápido.

## Base de datos: perfiles, no marcas sueltas

- `postgresql-managed`: default estable para backends multiusuario.
- `sqlite-local`: default estable para escritorio/local.
- `turso-libsql`: trial para casos edge/libSQL explícitos.
- `turso-sync`: experimental para replicas y sync diseñado.
- `turso-database`: experimental; nunca default sin piloto.

Así Turso se evalúa por capacidad y riesgo, no como reemplazo global de PostgreSQL.

## Comandos del ciclo de vida

| Comando | Utilidad |
|---|---|
| `eng catalog` | Consulta estados, categoría y cobertura |
| `eng boilerplate evaluate` | Detecta duplicado, refresh o candidata |
| `eng recommend` | Resuelve intake a Recipe sin escribir |
| `eng new` | Crea manifest y documentos del proyecto |
| `eng doctor` | Detecta divergencia e ids inválidos |
| `eng plan` | Selecciona skills y gates por tipo de cambio |
| `eng check` | Selecciona gates según manifest y archivos |
| `eng add` | Planea un feature; `--apply` actualiza los archivos gestionados |
| `eng update` | Aplica la estrategia upstream de cada starter |

## Modelo de equipo

| Tamaño | Operación |
|---|---|
| 1 persona | El manifest sustituye decisiones recordadas de memoria; el autor revisa excepciones |
| 2–5 | Un responsable revisa cambios de Recipe o starter; cambios normales siguen el manifest |
| 6–12 | Owners por Tier A y revisión cruzada de plataforma; proyectos conservan autonomía |
| 13–20 | Rotación de mantenimiento y ventana de upgrades; sin comité para tareas normales |

No se necesita Backstage, base de datos de inventario ni servicio siempre encendido. Esa complejidad solo se justificaría cuando exista fleet real que el repositorio y la automatización ya no puedan observar.

## Fuentes de verdad

- [`platform/boilerplates.json`](platform/boilerplates.json): candidatos, aliases, estados e integración.
- [`platform/golden-paths.json`](platform/golden-paths.json): Project Recipes.
- [`platform/database-profiles.json`](platform/database-profiles.json): decisiones de datos.
- [`platform/feature-packs.json`](platform/feature-packs.json): capacidades componibles.
- [`skills/registry.json`](skills/registry.json): skills reales y rutas.
- `.engineering/project.json`: estado instalado de cada proyecto.

## Estructura

```text
platform/       registros y Recipes legibles por máquina
schemas/        contratos JSON
catalog/        fichas y memoria histórica
curation/       pins y adapters reproducibles
.opencode/      skills del asistente
feature-packs/  contratos de capacidades opcionales
upgrades/       actualización por estrategia upstream
examples/       intakes y proyectos canónicos
evals/          tareas de evaluación
scripts/        CLI y validación
tests/          pruebas del resolver y curador
docs/           guías humanas y decisiones
```

## Validación

```bash
make check
```

El validator comprueba JSON, aliases, URLs duplicadas, referencias entre Recipes/boilerplates/datos/features/skills, pins, adapters, schemas y enlaces Markdown. Las pruebas ejercitan resolución, Turso, detección de duplicados, refresh de React Starter Kit, generación segura y doctor.
