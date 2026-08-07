# Monorepo multi-app

Se usa solo cuando hay múltiples apps reales.

```text
apps/
  web/       TanStack
  api/       Hono
  mobile/    Ignite   # solo si existe
packages/
  contracts/
  sdk/
  config/
  testing/
```

**Ejemplo:** web y móvil consumen el mismo SDK.  
**No ejemplo:** crear `desktop/` vacío “para futuro”.
