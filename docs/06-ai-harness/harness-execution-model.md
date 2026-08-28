# Cómo trabaja el harness — ejemplo completo

Tarea: “El director puede aprobar solicitudes”.

1. **Classify:** feature + authorization.
2. **Read manifest:** GP-02, auth/rbac/audit.
3. **Load skills:** Hono, TanStack, authorization, audit.
4. **Load canonical examples:** permission policy + audited action.
5. **Plan:** contract → use case → policy → audit → UI → tests.
6. **Strategy:** Gentle elige ejecución directa o SDD según riesgo y ambigüedad.
7. **Implement:** Gentle realiza cambios; Pi puede seguir como harness operativo.
8. **Guards:** policy test, API contract, audit test, UI types.
9. **Review:** humano revisa diff y comportamiento.
10. **Knowledge:** no se registra como patrón global salvo que aparezca aprendizaje reusable.

No cargar Ignite, Tauri, SpeedPy o multitenancy porque no intervienen.
