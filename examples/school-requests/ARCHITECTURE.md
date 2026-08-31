# Arquitectura: school-requests

La definición vive en `.engineering/project-definition.json`; stack, features, exclusiones y gates viven únicamente en `.engineering/project.json`.

## Patrones

- Monolito modular por servicio, contratos explícitos y mínimo privilegio.
- Clientes sin autoridad de seguridad ni reglas de dominio duplicadas.
- Cambios de datos con migración, recuperación y auditoría proporcional al riesgo.
- Single-tenant por defecto.

## Límites materiales

- `services/api` pertenece a `hono-api`; respeta sus instrucciones y comandos.
- `apps/admin` pertenece a `tanstack-admin`; respeta sus instrucciones y comandos.

Las desviaciones permanentes requieren actualizar la Recipe o registrar una decisión.
