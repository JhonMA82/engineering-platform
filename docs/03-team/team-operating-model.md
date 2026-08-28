# Modelo operativo para 1–20 personas

La plataforma no añade procesos distintos por tamaño. Intake, Recipe, manifest, skills y gates son iguales; solo cambia quién revisa una excepción.

| Equipo | Revisión mínima |
|---|---|
| 1 | Autorrevisión explícita de Recipe y warnings |
| 2–5 | Otra persona revisa cambios de stack, datos o seguridad |
| 6–12 | Owner por activo Tier A y revisión de cambios de plataforma |
| 13–20 | Rotación de mantenimiento y ventana planificada de upgrades |

## Flujo diario

Para “agregar prioridad a solicitudes”:

1. `eng doctor` confirma que el proyecto sigue GP-02.
2. `eng plan --change-type schema` selecciona skill de datos y gates.
3. El asistente modifica schema, migración, API y UI respetando ownership.
4. El desarrollador revisa el diff y ejecuta los comandos reales de migración, types y tests.
5. Solo una decisión reusable o un incidente repetido vuelve a la plataforma.

No se convoca comité para una feature normal. Una excepción de Recipe, un nuevo boilerplate o una promoción de delivery sí requiere evidencia y revisión.
