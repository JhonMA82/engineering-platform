# React Starter Kit

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `saas-edge` |
| Uso predeterminado | Productos SaaS full-stack TypeScript con despliegue en el edge de Cloudflare |
| Repositorio | [https://github.com/kriasoft/react-starter-kit](https://github.com/kriasoft/react-starter-kit) |
| Revisión de fuentes | 2026-08-05 |

## Tesis de adopción

React Starter Kit es una apuesta full-stack TypeScript distinta de las demás del catálogo: un monorepo SaaS con React 19, TanStack Router, tRPC y Cloudflare Workers que declara estar listo para producción y que incluye autenticación, organizaciones, pagos y UI. Entra como alternativa al producto comercial, sin desplazar a los defaults institucionales ni a Open SaaS, que sigue siendo la vía basada en Wasp.

## Qué ofrece el repositorio

- Monorepo full-stack con apps `web` (Astro), `app` (React 19 SPA), `api` (Hono + tRPC) y `email` (React Email).
- Paquetes `ui` (shadcn/ui), `core` e infraestructura Terraform para Cloudflare.
- Type-safety del contrato desde base de datos hasta la UI: TypeScript, tRPC y Drizzle ORM.
- Autenticación con Better Auth: OTP por correo, passkeys, Google OAuth y organizaciones.
- Billing con Stripe (suscripciones y price IDs configurables).
- Drizzle ORM con Neon PostgreSQL, migraciones y seed.
- TanStack Router (file-based), TanStack Query, Jotai, Tailwind CSS v4.
- Runtime Bun, Vite, Vitest, ESLint y Prettier.
- Despliegue de cada app como Worker independiente de Cloudflare con service bindings.
- AGENTS.md, CLAUDE.md, comandos para Claude y configuración para Gemini en el repositorio.
- Asistentes de IA entrenados sobre el propio código y documentación en reactstarter.com.
- Licencia MIT.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Producto SaaS en TypeScript cuando el cliente acepta Cloudflare como plataforma.
- Aplicación full-stack React con auth, organizaciones y suscripciones desde el inicio.
- Producto con sitio público (Astro), área privada y API en un solo monorepo.
- Proyectos donde la consultoría prefiera un stack React/TanStack coherente en toda la pila.

## Ejemplos por tipo de cliente

- **Gobierno:** solo si el cliente acepta infraestructura edge y el modelo organizacional cubre instituciones; caso poco frecuente.
- **Escuela:** producto SaaS vendido a múltiples planteles, con tenant isolation validado.
- **Sindicato:** plataforma ofrecida a múltiples organizaciones, con aislamiento diseñado.
- **Pyme:** software por suscripción o portal B2B full-stack TypeScript.

## Cuándo no usarlo

- Sistema interno single-tenant sin billing.
- Proyectos donde el cliente imponga su propio hosting o Node.js clásico.
- Datos institucionales sensibles sin validar región, residencia de datos y cumplimiento de Cloudflare/Neon.
- API-first con múltiples clientes externos: tRPC acopla cliente y servidor; FastAPI es más abierto.
- Equipo sin disposición a adoptar Bun y el ciclo de vida de Workers.

## Ventajas estratégicas

- Pila coherente y type-safe de punta a punta.
- Capacidades comerciales reales incluidas: auth, organizaciones, Stripe y correo.
- Proyecto maduro y activo desde 2014, con gran comunidad y documentación propia.
- Alta orientación AI-friendly: AGENTS.md, CLAUDE.md, docs y asistentes entrenados.
- Despliegue simple y reproducible a Cloudflare, con Terraform incluido.

## Riesgos, madurez y límites

- Cloudflare es un proveedor con peso: Workers, Neon y Resend deben aceptarse desde el inicio.
- Bun reemplaza Node.js/npm; hay que verificar compatibilidad con el resto de la cadena.
- El edge impone consideraciones de límites, cold starts, precio y portabilidad.
- tRPC acopla el frontend al backend; no es un contrato de API pública abierto.
- Organizaciones de Better Auth no equivalen a tenant isolation ni RBAC empresarial.
- Stripe, webhooks y OAuth aumentan superficie de seguridad y compliance.
- No afirmar multi-tenancy ni certificación de producción sin verificación por proyecto.

## Relación con otras opciones del catálogo

- **Frente a Open SaaS:** ambas cubren SaaS comercial; [comparación directa](../comparisons/open-saas-vs-react-starter-kit.md).
- **Frente a TanStack/Next Admin:** los dashboards son capa de UI; React Starter Kit es producto full-stack con backend y billing.
- **Frente a FastAPI:** Python API-first y multi-client vs. TypeScript full-stack acoplado.
- **Frente a Institutional Operations Starter:** producto comercial externo vs. activo interno con trazabilidad institucional.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Fijar commit/release y revisar licencia, seguridad y compatibilidad antes de cada proyecto.
- [ ] Diseñar tenant isolation, memberships y auditoría; no asumir que organizaciones lo resuelven.
- [ ] Probar webhooks de Stripe, OAuth y recuperación de cuenta en staging.
- [ ] Definir región de Workers y residencia de datos para clientes institucionales.
- [ ] Preparar i18n, accesibilidad, errores y validaciones en español.
- [ ] Incluir CI, pruebas Vitest y observabilidad para los tres Workers.
- [ ] Documentar costos del edge y una posible salida/migración fuera de Cloudflare.

## Evaluación AI-friendly

**Alta.** El repositorio incluye AGENTS.md, CLAUDE.md, comandos de agente, docs propios y asistentes entrenados sobre el código. Los agentes deben respetar el flujo de migraciones Drizzle, la configuración de Wrangler y los service bindings, y no editar output generado sin mapa del proyecto.

## Despliegue y operación

- Secretos vía Wrangler; separar auth, Stripe, OAuth, correo e IA.
- Deployar en orden: email, web, api, app.
- Staging obligatorio para billing, webhooks y correo.
- Respaldos independientes de PostgreSQL y control de versiones de migraciones.
- Documentar límites de Workers y precios antes de comprometer al cliente.

## Decisión final

**Adoptado como opción full-stack TypeScript para SaaS comercial**, propuesta por el usuario. Coexiste con Open SaaS como alternativa con distinta base de framework y hosting; ninguno es default institucional sin piloto.

## Fuentes oficiales

- [https://github.com/kriasoft/react-starter-kit](https://github.com/kriasoft/react-starter-kit)
- [https://reactstarter.com](https://reactstarter.com)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
