# Flujo: actualizar un starter

1. Detectar nueva release upstream.
2. Leer changelog/release notes.
3. Actualizar mirror/vendor.
4. Crear integration branch.
5. Ejecutar baseline tests y evals.
6. Revisar cambios en arquitectura, dependencies y seguridad.
7. Adaptar customization mínima.
8. Crear release candidate interno.
9. Probar con sandbox o proyecto piloto.
10. Publicar curated release.
11. Crear upgrade recipe para proyectos existentes si aporta valor.

No actualizar todos los proyectos automáticamente solo porque upstream publicó una versión.
