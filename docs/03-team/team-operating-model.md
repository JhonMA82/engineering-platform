# Modelo operativo del equipo — con ejemplo

## Flujo diario de un junior
Tarea: “agregar prioridad a solicitudes”.

1. Lee `.engineering/project.json` → GP-02, auth/rbac/audit.
2. Pide al harness contexto para `feature + db-change`.
3. Revisa plan antes de aplicar.
4. IA modifica schema, migration, API y UI.
5. Junior revisa el diff.
6. Ejecuta gates: migration + contracts + unit/integration + UI.
7. Abre PR explicando riesgo y verificación.

## Qué no debe hacer
- pedir “hazlo completo” sin leer manifest;
- aceptar librerías nuevas sin razón;
- saltarse migration porque “Drizzle ya sabe”; 
- copiar la solución de otro repo sin validar licencia/arquitectura.
