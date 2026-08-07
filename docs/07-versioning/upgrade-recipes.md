# Upgrade Recipe — ejemplo

`hono-api 0.4 → 0.5`

1. `detect`: verifica versión y patrón viejo.
2. `plan`: lista archivos y migraciones.
3. `apply`: codemod/cambio controlado.
4. `verify`: types + tests + OpenAPI.
5. actualizar manifest.

Si falla verify, no marcar la nueva versión como instalada.
