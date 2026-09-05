---
name: project-discovery
description: 'Descubre, confirma y prepara una idea de proyecto nuevo con Engineering Platform. Úsala cuando el usuario quiera iniciar un producto desde cero, seleccionar una Recipe o generar el handoff para Gentle AI.'
compatibility: 'Requiere Python 3.11+, el ejecutable eng incluido y un destino vacío o que contenga únicamente `.atl/`, `.gitignore` y `.git`.'
---

# Project Discovery

Tu objetivo es reducir ambigüedad suficiente para iniciar, no redactar un documento enorme ni pedir al usuario que elija frameworks.

## Invariantes

- Trabaja únicamente en el `cwd` indicado por `/new-project`.
- El destino puede estar vacío o contener únicamente `.atl/`, `.gitignore` y `.git`. Si contiene cualquier otro contenido del usuario, detente: no crees un proyecto anidado ni sobrescribas trabajo.
- Haz de una a tres preguntas cortas por ronda. Empieza por problema, usuarios y resultado; pregunta detalles solo si cambian arquitectura, riesgo o alcance.
- No preguntes por stack, boilerplate o patrón salvo que exista una restricción técnica real. Engineering Platform toma esas decisiones.
- Separa `must_have` de `out_of_scope`. No conviertas deseos futuros en alcance inicial.
- Muestra un resumen y solicita confirmación explícita antes de escribir una definición con estado `confirmed`.
- No desarrolles el producto. Termina con el código base materializado, verificado y el handoff para Gentle.
- La idea inicial debe llegar como texto libre del usuario.
- Si todavía no existe una idea, hacé esa pregunta con la tool nativa («Describí la idea principal del proyecto con tus propias palabras; puede estar incompleta») acompañada de 2 a 4 ideas de ejemplo como opciones. Después espera.
- Preguntá SIEMPRE con la tool nativa de preguntas de Pi (`ask_user_question`): cada pregunta lleva de 2 a 4 opciones predefinidas con trade-offs, y la escritura libre sale sola por la fila de respuesta personalizada que la tool agrega (no inventes una opción «Otro»: la tool la rechaza).
- No sugieras resultados, features, stack, Recipes o boilerplates hasta poder resumir la idea sin inventar información.
- Si la idea sigue siendo demasiado vaga, realiza preguntas abiertas sobre qué quiere conseguir, quién lo usará o en qué contexto.
- Después de entender la idea, realiza de una a tres preguntas por ronda basadas en las respuestas anteriores.
- Las opciones describen necesidades, clientes y alcance (problema, quién lo usa, qué clientes necesita, qué queda fuera), nunca frameworks, starters ni Recipes: la plataforma decide el stack.
- Si todavía no hay idea, la primera pregunta trae 2 a 4 ideas de ejemplo como opciones y el usuario escribe la suya en la fila libre.
- Agrupá hasta 4 preguntas relacionadas en una sola llamada; no más de 3 rondas sin mostrar el resumen provisional con la Recipe tentativa.

## Flujo

1. Confirma que el destino está vacío o contiene únicamente `.atl/`, `.gitignore` y `.git`. Conoce el nombre kebab-case derivado de la carpeta.
2. Descubre progresivamente:
   - problema y usuarios;
   - resultado observable y alcance mínimo;
   - datos, permisos, integraciones y clientes necesarios;
   - restricciones, riesgos, incógnitas y criterios de aceptación;
   - elementos explícitamente fuera de alcance.
3. Deduce `project_type`, señales, features, surfaces y perfil de datos. Después de cada ronda de preguntas, ejecuta `<eng> recommend --input <archivo-temporal> --suggest` como lectura provisional y compartí una línea («con lo que sé hasta ahora iríamos a GP-X con estos clientes»); si el motor devuelve `corrected_intake`, adoptá el `project_type` corregido desde esa ronda, no al final. Si falla sin corrección, ajustá la clasificación; no inventes ids.
4. Explica la Recipe recomendada en cinco líneas o menos, incluyendo cualquier starter no liberado. Si hay una alternativa materialmente válida, menciónala; no presentes un menú completo.
5. Tras la confirmación, crea `.engineering/project-definition.json` conforme a `schemas/project-definition.schema.json`. Usa `discovery.status: confirmed` y `discovery.confirmed_by: user` solo después de la confirmación.
6. Ejecuta `<eng> bootstrap --from .engineering/project-definition.json --output .`.
7. Ejecuta `<eng> doctor --project .` y corrige únicamente problemas de definición o generación.
8. Entrega un resumen corto: Recipe, stack, readiness, checks y ruta `GENTLE.md`.

`<eng>` es la ruta absoluta recibida como argumento del skill. Si no fue proporcionada, usa `eng` desde `PATH`.

## Clientes → intake (regla de surfaces)

Deducí los clientes antes de fijar `project_type`:

| Clientes | `project_type` | `surfaces` iniciales |
|---|---|---|
| Solo operadores internos (dashboard) | `admin` o `institutional-admin` | `[]` |
| Operadores + ciudadanía anónima / portal público / kiosco | `multi-app` | `public-intake` con capabilities válidas |
| Operadores + campo móvil | `multi-app` | `mobile` con capabilities válidas |
| Combinación de las anteriores | `multi-app` | ambas surfaces |

Reglas:

- Señales de stack también deciden starter: `next-ecosystem` o ecosistema Next → el motor selecciona `next-admin` sobre `tanstack-admin` (misma categoría, señales más fuertes) y lo avisa; sin esas señales queda el default.
- Dos o más clientes reales, o uno más otro futuro explícito («en unos meses»), → `multi-app` desde el inicio, aunque las `surfaces` iniciales queden vacías. La Recipe componible se elige por el futuro; la surface se agrega cuando se confirma (`eng surface add`). Un deseo futuro nunca recorta la Recipe a una no componible.
- Todas las surfaces comparten la misma API y contratos; nunca propongas un backend por cliente.
- El vocabulario de capabilities es cerrado (`SURFACE_CAPABILITIES` en `scripts/eng.py`): el lenguaje del usuario se traduce, no se copia. QR / folio / token de seguimiento / kiosco / cola offline → `public-intake` con `form-capture`, `offline-outbox`, `tracking-token`, `kiosk-mode` (más `offline-drafts`, `attachments-offline`, `pwa-installable`, `connectivity-status` si aplican). Nunca inventes capabilities como `qr-capture`: `recommend` las rechaza.
- Si `recommend` responde «... no permite componer ... reintentá con `project_type='multi-app'`», corregí el `project_type` a `multi-app`; no quites la surface.
- Las señales de dominio bastan como respaldo: si el intake lleva `kiosk-mode` o `tracking-token` aunque falte `surfaces`, el motor implica `public-intake` y avisa. Igual declará siempre `surfaces` explícitas: la inferencia fija alcance mínimo, no el contrato completo.

## Contrato mínimo

La definición contiene:

- `idea`: summary, problem, users, outcomes, must_have y out_of_scope;
- `intake`: name, project_type, signals, features, excluded_features, database y constraints;
- `delivery`: acceptance_criteria, risks y unknowns;
- `discovery`: status, confirmed_by y confirmed_at.
## Procedencia

La entrevista progresiva adapta el patrón `grilling` de `mattpocock/skills`, licencia MIT, observado en el commit `6654f6b60cd9d5be8b54c6fafe44346dabeb3b76`.

No dejes texto de ejemplo, marcadores ni decisiones sin confirmar.
