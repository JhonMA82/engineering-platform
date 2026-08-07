# Cómo trabaja el harness — ejemplo completo

Tarea: “El director puede aprobar solicitudes”.

1. **Classify:** feature + authorization.
2. **Read manifest:** GP-02, auth/rbac/audit.
3. **Load skills:** Hono, TanStack, authorization, audit.
4. **Load canonical examples:** permission policy + audited action.
5. **Plan:** contract → use case → policy → audit → UI → tests.
6. **Implement:** OpenCode realiza cambios.
7. **Guards:** policy test, API contract, audit test, UI types.
8. **Review:** humano revisa diff y comportamiento.
9. **Knowledge:** no se registra como patrón global salvo que aparezca aprendizaje reusable.

No cargar Ignite, Tauri, SpeedPy o multitenancy porque no intervienen.
