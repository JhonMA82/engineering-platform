# AGENTS.md

## Propósito
Gobierna cómo un agente trabaja con la Engineering Platform.

## Antes de modificar un proyecto
Leer en orden:
1. `.engineering/project.json`;
2. `AGENTS.md` del proyecto;
3. Golden Path;
4. feature packs instalados;
5. ADRs relevantes;
6. canonical examples del cambio.

## Ejemplo
Tarea: “agrega campo `priority` a solicitudes”.

Correcto:
- detectar GP-02;
- cargar TanStack + Hono + DB;
- crear migration;
- actualizar contrato API;
- actualizar UI;
- ejecutar migration test + OpenAPI diff + tests.

Incorrecto:
- agregar Redis;
- crear microservicio;
- cambiar ORM;
- introducir multitenancy.

## Invariantes
- arquitectura mínima suficiente;
- frontend no es autoridad de seguridad;
- todo cambio DB tiene migration;
- toda ruta pública tiene contrato OpenAPI;
- todo bug produce regression test;
- recursos tenant requieren tenant isolation test;
- no usar `latest` en dependencias productivas;
- no copiar código de referencias con licencia incompatible.
