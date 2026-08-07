# Start Here — guía para nuevos integrantes

## Qué necesitas entender primero
No necesitas memorizar frameworks. Necesitas aprender el **flujo de decisión**.

### Paso 1 — entender el problema
Ejemplo: “La escuela necesita recibir solicitudes internas”.

Preguntas:
- ¿sitio público o aplicación interna? → interna.
- ¿una escuela o muchas? → una.
- ¿necesita móvil? → no.
- ¿necesita usuarios? → sí.
- ¿necesita aprobaciones? → sí.

### Paso 2 — seleccionar Golden Path
Resultado: `GP-02 Admin Application`.

### Paso 3 — agregar solo capacidades requeridas
- auth ✅
- rbac ✅
- audit ✅
- multitenancy ❌
- jobs ❌
- webhooks ❌

### Paso 4 — crear manifest
Ver [Project Manifest](../01-concepts/CONCEPTS_WITH_EXAMPLES.md#5-project-manifest).

### Paso 5 — trabajar con harness
El harness carga solo TanStack, Hono, DB, auth, permisos y audit.

### Paso 6 — terminar por gates
No es “Done” cuando el agente dice que funciona. Es Done cuando pasan tests, typecheck, migration y gates relevantes.

## Primer ejercicio recomendado
Lee y reproduce en sandbox el [ejemplo completo](../13-examples/END_TO_END_SCHOOL_REQUESTS.md).
