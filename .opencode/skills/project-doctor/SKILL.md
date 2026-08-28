---
name: project-doctor
description: Comprueba que un proyecto sigue una Recipe válida y no contiene decisiones indefinidas.
---

# Project Doctor

Ejecuta `./eng doctor --project <directorio>`. Trata como error ids inexistentes, `PINNED`, dependencias de feature incumplidas y cambios de stack no registrados. Trata como advertencia un starter no liberado o documentos gestionados ausentes. No corrijas automáticamente una decisión arquitectónica: genera intake actualizado, vuelve a resolver y registra ADR si cambia el resultado.
