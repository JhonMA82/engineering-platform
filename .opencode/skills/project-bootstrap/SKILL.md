---
name: project-bootstrap
description: Materializa un proyecto reproducible después de resolver su Recipe.
---

# Project Bootstrap

Antes de iniciar, el destino puede estar vacío o contener únicamente `.atl/`, `.gitignore` y `.git`. Si contiene cualquier otro contenido del usuario, detente y no sobrescribas trabajo.

1. Valida `.engineering/project-definition.json` y exige confirmación explícita del usuario.
2. Muestra la Recipe y las advertencias antes de materializar dependencias externas.
3. Ejecuta `eng bootstrap --from .engineering/project-definition.json --output .` para materializar cada adapter desde su pin exacto.
4. Comprueba `.engineering/materialization.json`: fuentes, destinos, hashes, setup y checks deben corresponder al código real.
5. Ejecuta `eng doctor --project .`; si readiness es `code-ready`, ejecuta además `eng check --project . --run`.
6. Entrega el proyecto a Gentle únicamente con `scaffold_status: materialized` y un `GENTLE.md` actualizado.

Para automatización sin conversación se conserva la ruta legacy `eng new --from <intake> --output <directorio>`.

No sobrescribas contenido de usuario. El código de dominio siempre es `user_owned`.
