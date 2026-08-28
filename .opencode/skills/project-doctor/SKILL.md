---
name: project-doctor
description: Comprueba que un proyecto sigue una Recipe válida y no contiene decisiones indefinidas.
---

# Project Doctor

Ejecuta `eng doctor --project <directorio>`. Trata como error ids inexistentes, `PINNED`, dependencias de feature incumplidas, handoff divergente, documentos gestionados ausentes y cambios de stack no registrados. Un starter no liberado es advertencia. No corrijas automáticamente una decisión arquitectónica: actualiza la definición, vuelve a resolver y registra ADR si cambia el resultado.
