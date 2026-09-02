# Monorepo multi-app

Se usa solo cuando hay múltiples apps reales.

```text
apps/
  admin/     TanStack
  mobile/    Ignite   # solo si existe
  desktop/   Tauri    # solo si existe
services/
  api/       API Starter
packages/
  contracts/
  sdk/
  config/
  testing/
```

**Ejemplo:** web y móvil consumen el mismo SDK.  
**No ejemplo:** crear `desktop/` vacío “para futuro”.

La composición inicial puede conservar lockfiles independientes de cada boilerplate. El workflow raíz ejecuta los checks en su directorio; unificar workspaces, contratos o SDK solo se hace cuando existe una necesidad compartida y mediante un incremento explícito.
