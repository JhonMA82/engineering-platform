---
name: change-planner
description: Convierte un cambio de producto en el conjunto mínimo de skills, archivos y quality gates.
---

# Change Planner

1. Lee `.engineering/project.json` y el diff o alcance solicitado.
2. Clasifica el cambio: `api`, `schema`, `permission`, `ui`, `upgrade` o `incident`.
3. Ejecuta `eng plan --project . --change-type <tipo>`.
4. Añade gates solo cuando el riesgo observable lo justifique; nunca elimines un gate exigido por la Recipe.
5. Identifica archivos `managed`, `managed_sections`, `seeded` y `user_owned` antes de editar.
6. Termina con evidencia concreta de cada gate, no con “debería funcionar”.

Para un feature pack existente usa primero `eng add <feature> --project .`; revisa el plan y solo entonces repite con `--apply`.
