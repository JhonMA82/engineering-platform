---
name: project-discovery
description: Descubre, confirma y prepara una idea de proyecto nuevo con Engineering Platform. Úsala cuando el usuario quiera iniciar un producto desde cero, seleccionar una Recipe o generar el handoff para Gentle AI.
compatibility: Requiere Python 3.11+, el ejecutable eng incluido y un destino vacío o que contenga únicamente los metadatos gestionados de Gentle AI: `.atl/` y `.gitignore` en la raíz, con el estado anidado permitido dentro de `.atl/`.
---

# Project Discovery

Tu objetivo es reducir ambigüedad suficiente para iniciar, no redactar un documento enorme ni pedir al usuario que elija frameworks.

## Invariantes

- Trabaja únicamente en el `cwd` indicado por `/new-project`.
- El destino puede estar vacío o contener únicamente `.atl/` y `.gitignore` en la raíz; `.atl/` contiene metadatos de Gentle AI y puede contener estado anidado. Si contiene cualquier otro contenido del usuario, detente: no crees un proyecto anidado ni sobrescribas trabajo.
- Haz de una a tres preguntas cortas por ronda. Empieza por problema, usuarios y resultado; pregunta detalles solo si cambian arquitectura, riesgo o alcance.
- No preguntes por stack, boilerplate o patrón salvo que exista una restricción técnica real. Engineering Platform toma esas decisiones.
- Separa `must_have` de `out_of_scope`. No conviertas deseos futuros en alcance inicial.
- Muestra un resumen y solicita confirmación explícita antes de escribir una definición con estado `confirmed`.
- No desarrolles el producto. Termina con el blueprint y el handoff para Gentle.

## Flujo

1. Confirma que el destino está vacío o contiene únicamente `.atl/` y `.gitignore` en la raíz; `.atl/` contiene metadatos de Gentle AI y puede contener estado anidado. Conoce el nombre kebab-case derivado de la carpeta.
2. Descubre progresivamente:
   - problema y usuarios;
   - resultado observable y alcance mínimo;
   - datos, permisos, integraciones y clientes necesarios;
   - restricciones, riesgos, incógnitas y criterios de aceptación;
   - elementos explícitamente fuera de alcance.
3. Deduce `project_type`, señales, features y perfil de datos. Ejecuta `<eng> recommend --input <archivo-temporal>` para validar la Recipe. Si falla, corrige la clasificación; no inventes ids.
4. Explica la Recipe recomendada en cinco líneas o menos, incluyendo cualquier starter no liberado. Si hay una alternativa materialmente válida, menciónala; no presentes un menú completo.
5. Tras la confirmación, crea `.engineering/project-definition.json` conforme a `schemas/project-definition.schema.json`. Usa `discovery.status: confirmed` y `discovery.confirmed_by: user` solo después de la confirmación.
6. Ejecuta `<eng> bootstrap --from .engineering/project-definition.json --output .`.
7. Ejecuta `<eng> doctor --project .` y corrige únicamente problemas de definición o generación.
8. Entrega un resumen corto: Recipe, stack, estado real del scaffold y ruta `GENTLE.md`.

`<eng>` es la ruta absoluta recibida como argumento del skill. Si no fue proporcionada, usa `eng` desde `PATH`.

## Contrato mínimo

La definición contiene:

- `idea`: summary, problem, users, outcomes, must_have y out_of_scope;
- `intake`: name, project_type, signals, features, excluded_features, database y constraints;
- `delivery`: acceptance_criteria, risks y unknowns;
- `discovery`: status, confirmed_by y confirmed_at.

No dejes texto de ejemplo, marcadores ni decisiones sin confirmar.
