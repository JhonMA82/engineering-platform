---
name: gate-runner
description: Selecciona y ejecuta verificaciones proporcionales al cambio y a la Recipe.
---

# Gate Runner

Usa `eng check --project . --changed-files <archivos>` para seleccionar gates y añade `--run` para ejecutar los comandos registrados por los adapters. No inventes scripts; si no existe comando, reporta la brecha. Los fallos de seguridad, migración, contrato o build bloquean la entrega.
