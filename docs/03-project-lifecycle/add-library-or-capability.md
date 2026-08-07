# Flujo: nueva librería o capacidad después del init

Un proyecto puede crecer sin volver a ejecutar todo el bootstrap.

1. Leer `.engineering/project.json`.
2. Clasificar: dependencia local, feature pack, cambio de Golden Path o nueva app.
3. Verificar si la plataforma ya tiene pack/patrón aprobado.
4. Evaluar licencia, mantenimiento, compatibilidad y superficie de seguridad.
5. Implementar en una rama corta.
6. Agregar únicamente la skill/contexto de la nueva capacidad.
7. Ejecutar guards específicos.
8. Actualizar manifest, ADR y lockfile.
9. Si la capacidad es reusable, proponerla a la plataforma después de validarla en el proyecto.

No regenerar módulos que no cambiaron ni volver a ejecutar todo el init.
