# Consulting API Starter

| Campo | Decisión |
|---|---|
| Estado de decisión | Default para API TypeScript separada |
| Estado de entrega | `released` mediante `api-starter v0.12.0` |
| Mantenimiento | Tier A |
| Repositorio | `JhonMA82/api-starter` |

## Tesis

El equipo necesita una API TypeScript pequeña y aburrida cuando varios clientes comparten dominio o hay integraciones. Hono es el mecanismo; el valor interno debe estar en contratos, autorización, errores, migraciones, observabilidad y generadores comprobados.

## Entrega

El adapter clona el commit fijado, ejecuta su generador con las capacidades resueltas y materializa el workspace en `services/api`. Instala con Bun y ejecuta lint, tipos y pruebas. El manifiesto `.api-starter/manifest.json` conserva procedencia y actualización.

## Estructura objetivo

- `contracts`: esquemas, OpenAPI y errores estables.
- `application`: casos de uso sin acoplarse al transporte.
- `domain`: reglas del negocio.
- `adapters`: persistencia e integraciones.
- `http`: routing, autenticación, autorización y serialización.
- `tests`: unidad, contrato, integración y seguridad.

## Criterios para liberar

- [x] Starter externo propio versionado y validado.
- [x] Pin reproducible de runtime y dependencias.
- [x] Auth, autorización, auditoría y migraciones componibles mediante generador.
- [x] OpenAPI validado por CI.
- [x] Gates base de types, build y unidad; los demás se añaden por Recipe.
- [x] Actualización mediante manifest/generator y responsable Tier A.
