# Harness Integration

El harness actual es deliberadamente pequeño:

- `.engineering/project.json` limita arquitectura y capacidades;
- `skills/registry.json` aporta instrucciones concretas;
- `eng plan` selecciona contexto por tipo de cambio;
- `eng check` selecciona quality gates por manifest y archivos;
- schemas, validator y CI detectan divergencia.

El starter materializado debe mapear cada gate a un comando real. Hasta entonces `eng check` declara `selection-only` y reporta la brecha; no simula haber ejecutado pruebas.
