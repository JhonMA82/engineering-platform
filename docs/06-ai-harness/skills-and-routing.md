# Skills y routing — ejemplos

| Cambio | Skills |
|---|---|
| nuevo endpoint | contracts + authorization si es sensible |
| cambio schema | database |
| botón + tabla | gate-runner + reglas del starter materializado |
| permiso | authorization + security-review |
| actualización React Starter Kit | react-starter-kit-updater + gate-runner |
| Excel o transformación | data + database si persiste |

El registro ejecutable está en `skills/registry.json`, que incluye `project-discovery` (platform) y `project-evolution` (project): la creación y la evolución post-bootstrap se resuelven por registry, no por paths sueltos. `./eng plan` selecciona el subconjunto para el manifest y tipo de cambio.

Error típico: cargar todas las skills o mencionar una que no tiene implementación.
