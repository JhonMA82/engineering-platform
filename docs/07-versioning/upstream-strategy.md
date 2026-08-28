# Upstream Strategy

No existe una actualización universal. Cada entrada declara `integration.mode` y `update_strategy`.

| Modo | Uso | Actualización |
|---|---|---|
| direct | generador externo sin capa interna | regenerar y revisar diff |
| overlay | upstream más configuración de consultoría | integrar upstream y reaplicar overlay probado |
| seed-fork | upstream diseñado para divergencia | merge desde remoto seed |
| internal | activo propio | release SemVer y Upgrade Recipe |
| reference-only | fuente de patrones | revisión manual, nunca sync automático |

El adapter registra URL, commit/release, licencia, checks, ownership y patches. Nunca se inicia desde `main` sin snapshot reproducible.

React Starter Kit es `seed-fork`: debe leerse su skill nativo `merge-seed`, preservar identidad y alcance del proyecto, no reescribir migraciones aplicadas y cambiar el pin solo después de todos los checks.
