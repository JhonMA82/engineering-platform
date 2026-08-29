# Consulting Hono API Starter

| Campo | Decisión |
|---|---|
| Estado de decisión | Default para API TypeScript separada |
| Estado de entrega | `released` en `platform-0.5.0` |
| Mantenimiento | Tier A |
| Repositorio | Starter interno `starters/hono-api` |

## Tesis

El equipo necesita una API TypeScript pequeña y aburrida cuando varios clientes comparten dominio o hay integraciones. Hono es el mecanismo; el valor interno debe estar en contratos, autorización, errores, migraciones, observabilidad y generadores comprobados.

## Entrega

El adapter copia el release interno a `apps/api`, instala el lock npm y ejecuta typecheck y pruebas. Auth, persistencia y OpenAPI continúan como capacidades de proyecto o Feature Packs: no se anuncian como incluidas en el núcleo.

## Estructura objetivo

- `contracts`: esquemas, OpenAPI y errores estables.
- `application`: casos de uso sin acoplarse al transporte.
- `domain`: reglas del negocio.
- `adapters`: persistencia e integraciones.
- `http`: routing, autenticación, autorización y serialización.
- `tests`: unidad, contrato, integración y seguridad.

## Criterios para liberar

- [x] Starter interno versionado y validado.
- [x] Pin reproducible de runtime y dependencias.
- [ ] Auth, RBAC y migraciones como packs opcionales, no acoplados al núcleo.
- [ ] OpenAPI y cliente de ejemplo cuando lo requiera el contrato.
- [x] Gates base de types, build y unidad; los demás se añaden por Recipe.
- [x] Actualización por release interno y responsable Tier A.
