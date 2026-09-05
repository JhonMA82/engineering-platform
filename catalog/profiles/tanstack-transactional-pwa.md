# TanStack Transactional PWA

| Campo | Decisión |
|---|---|
| Estado de decisión | Especializado: solo cuando la captura pública offline transaccional es necesaria |
| Estado de entrega | `curated` mediante pin `f2571ea8` (tanstack-transactional-pwa 0.1.0) |
| Mantenimiento | Tier B |
| Repositorio | `JhonMA82/tanstack-transactional-pwa` |

## Tesis

Los portales públicos de captura —quejas, solicitudes, reportes de campo, módulos de autoservicio— deben funcionar con conexiones intermitentes y en kioscos, y entregar un comprobante de seguimiento incluso antes de sincronizar. Esa combinación (outbox idempotente en IndexedDB, adjuntos Blob, token opaco de 256 bits, kiosco con limpieza por inactividad) no la cubre ningún starter del catálogo: Stardrive resuelve contenido público y tanstack-admin resuelve administración privada. Para contenido web o flujos administrativos privados, este starter no es la opción.

## Entrega

El adapter clona el repositorio en el commit fijado y materializa el workspace completo en la raíz del proyecto: es un frontend autónomo que convive con un backend independiente mediante `docs/API-CONTRACT.md` (POST/GET `/v1/public/intake/submissions`). Instala con Bun y ejecuta lint (Biome), tipos, pruebas (Vitest con persistencia offline simulada) y build de producción con service worker. La actualización se realiza por diff curado sobre un nuevo pin, conservando el id.

## Estructura objetivo

- `apps/intake`: aplicación pública PWA.
  - `features/intake`: esquema Zod, formulario y tipos del dominio.
  - `hooks`: conectividad y sesión de kiosco.
  - `lib`: cliente HTTP, Dexie (borradores, outbox, adjuntos), token de seguimiento.
  - `routes`: formulario, seguimiento público por token y raíz.
- `packages/config`, `packages/env`, `packages/ui`: TypeScript compartido, variables tipadas y primitivas shadcn/ui.

## Límites que el proyecto consumidor debe resolver

- La outbox es tolerancia a fallos, no almacenamiento permanente: el navegador puede borrar IndexedDB y no hay Background Sync.
- Antiabuso, CSP, monitoreo, rate limiting y accesibilidad se definen por despliegue.
- La API debe revalidar todo; el cliente nunca es autoridad de seguridad.
- Para datos sensibles, evaluar expresamente si la captura offline es aceptable.

## Criterios de curación

- [x] Piloto real en el pin: install, check, check-types, test:run y build en verde con Bun 1.4.0.
- [x] Licencia MIT con atribución upstream declarada (Better-T-Stack 3.42.2).
- [x] Capacidades verificadas contra el código fuente, no solo el README.
- [x] Contrato de backend explícito con obligaciones de seguridad del servidor.
- [x] Estrategia de actualización declarada (diff curado) y checks del adapter registrados.
- [ ] Promoción a `released` exige materialización en un proyecto real con su backend.
