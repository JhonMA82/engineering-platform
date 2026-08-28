# AGENTS.md

Lee `.engineering/project.json` y `ARCHITECTURE.md` antes de cambiar el proyecto.

- No cambies stack, base de datos o features sin actualizar el intake.
- Lee `GENTLE.md` antes de decidir la estrategia de desarrollo.
- Autoriza en la API; ocultar botones no es seguridad.
- Todo cambio de schema incluye migración y backup/restore.
- No agregues multitenancy, jobs, mobile o desktop sin un nuevo requisito.
- Ejecuta `eng plan` antes del cambio y los gates aplicables al terminar.
