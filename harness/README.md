# Harness Integration

El harness convierte contexto en ejecución controlada.

## Ejemplo
Tarea: “agregar aprobación”.
- manifest dice GP-02 + auth/rbac/audit;
- harness carga skills Hono/TanStack/authorization/audit;
- exige policy + audit tests;
- no carga mobile/desktop.

El objetivo es menos contexto, menos tokens y menos decisiones inventadas.
