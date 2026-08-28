---
name: project-bootstrap
description: Genera el blueprint reproducible de un proyecto después de resolver su Recipe.
---

# Project Bootstrap

1. Valida `.engineering/project-definition.json` y exige confirmación explícita del usuario.
2. Muestra la Recipe y las advertencias antes de materializar dependencias externas.
3. Ejecuta `eng bootstrap --from .engineering/project-definition.json --output .` para crear manifest, arquitectura, reglas y handoff para Gentle.
4. Si cualquier starter no está `released`, conserva `scaffold_status: blueprint`; no clones `main` ni inventes un pin.
5. Cuando exista un adapter liberado, materializa desde su pin y registra commit, estrategia de actualización y patches.
6. Ejecuta `eng doctor --project .` antes de entregar.

Para automatización sin conversación se conserva la ruta legacy `eng new --from <intake> --output <directorio>`.

No sobrescribas un directorio no vacío. El código de dominio siempre es `user_owned`.
