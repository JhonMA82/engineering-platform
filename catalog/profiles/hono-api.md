# Consulting Hono API Starter

| Campo | Decisión |
|---|---|
| Estado de decisión | Default para API TypeScript separada |
| Estado de entrega | `catalog-only` |
| Mantenimiento | Tier A |
| Repositorio | Pendiente de extracción interna |

## Tesis

El equipo necesita una API TypeScript pequeña y aburrida cuando varios clientes comparten dominio o hay integraciones. Hono es el mecanismo; el valor interno debe estar en contratos, autorización, errores, migraciones, observabilidad y generadores comprobados.

## Límite honesto

Todavía no existe un release interno. La Recipe puede generar el blueprint y los contratos de decisión, pero no debe presentarlo como código productivo hasta que haya un piloto real, pin, tests, CI y estrategia de actualización.

## Estructura objetivo

- `contracts`: esquemas, OpenAPI y errores estables.
- `application`: casos de uso sin acoplarse al transporte.
- `domain`: reglas del negocio.
- `adapters`: persistencia e integraciones.
- `http`: routing, autenticación, autorización y serialización.
- `tests`: unidad, contrato, integración y seguridad.

## Criterios para liberar

- [ ] Extraído o validado en un proyecto real.
- [ ] Pin reproducible de runtime y dependencias.
- [ ] Auth, RBAC y migraciones como packs opcionales, no acoplados al núcleo.
- [ ] OpenAPI y cliente de ejemplo.
- [ ] Gates de lint, types, unidad, contrato, integración, build y seguridad.
- [ ] Upgrade recipe y responsable Tier A.
