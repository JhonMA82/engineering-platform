# Release Flow — ejemplo

Release de `priority`:
1. CI verde.
2. revisar migration.
3. staging.
4. backup si cambio es sensible.
5. deploy API.
6. migration controlada.
7. deploy web.
8. smoke: crear/editar solicitud.
9. observar errores/latencia.
10. cerrar release.
