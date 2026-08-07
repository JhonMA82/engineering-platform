# Guards

| Cambio | Guard |
|---|---|
| ruta API | OpenAPI + contract test |
| schema DB | migration test |
| recurso tenant | isolation suite |
| permiso | policy/matrix test |
| auth | security/session suite |
| UI | typecheck + a11y + E2E relevante |
| dependencia | lockfile + compatibility + build |
| deploy | smoke + health/readiness |
| bug | regression test |
| feature pack | profile/install/upgrade test |

Un agente no puede marcar Done con un guard obligatorio fallando.
