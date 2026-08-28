---
name: project-bootstrap
description: Genera el blueprint reproducible de un proyecto después de resolver su Recipe.
---

# Project Bootstrap

1. Valida el intake contra `schemas/intake.schema.json` y resuélvelo con `./eng recommend`.
2. Muestra la Recipe y las advertencias antes de materializar dependencias externas.
3. Ejecuta `./eng new --from <intake> --output <directorio>` para crear `.engineering/project.json`, `ARCHITECTURE.md`, `AGENTS.md` y README.
4. Si cualquier starter no está `released`, conserva `scaffold_status: blueprint`; no clones `main` ni inventes un pin.
5. Cuando exista un adapter liberado, materializa desde su pin y registra commit, estrategia de actualización y patches.
6. Ejecuta `./eng doctor --project <directorio>` antes de entregar.

No sobrescribas un directorio no vacío. El código de dominio siempre es `user_owned`.
