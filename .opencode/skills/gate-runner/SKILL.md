---
name: gate-runner
description: Selecciona y ejecuta verificaciones proporcionales al cambio y a la Recipe.
---

# Gate Runner

Usa `./eng check --project . --changed-files <archivos>` para seleccionar gates. Después mapea cada gate a un comando existente del starter materializado. No inventes scripts; si no existe comando, reporta la brecha. Registra comando, código de salida y resultado. Los fallos de seguridad, migración, contrato o build bloquean la entrega.
