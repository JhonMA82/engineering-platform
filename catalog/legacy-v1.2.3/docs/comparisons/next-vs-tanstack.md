# Next.js vs. TanStack Start

## Conclusión

**TanStack Start es el default para aplicaciones operativas privadas. Next.js permanece como variante para productos híbridos o capacidades específicas.**

## Elegir TanStack Start

- Backend independiente.
- Tablas y filtros complejos.
- Search params como parte del estado de aplicación.
- CRM, ERP ligero, trámites o expedientes.
- Vite y despliegue neutral.
- UI principalmente interactiva.
- Carga/caché explícitas.

## Elegir Next.js

- Sitio público y app privada en un solo producto.
- Contenido indexable y metadatos.
- Uso deliberado de RSC/Server Actions.
- Ecosistema o infraestructura existente.
- Integraciones específicas de Next.

## Trabajo de migración

La interfaz shadcn, tablas, formularios, Zustand, Zod y Tailwind son altamente reutilizables. Lo que cambia es:

| Next.js | TanStack Start |
|---|---|
| `app/`, layouts y pages | `src/routes/` |
| `next/link` | TanStack Router Link |
| `next/navigation` | hooks del router |
| `searchParams` | schemas de search tipados |
| Server Actions | server functions |
| Route Handlers | server routes |
| loading/error/not-found | pending/error/notFound components |
| RSC boundaries | modelo React tradicional + SSR |

El autor ya mantiene una versión oficial TanStack, por lo que no se debe repetir una migración manual.

## Riesgo de dos forks

Mantener cambios visuales duplicados provoca divergencia. La solución es la [Consulting Admin Family](../internal-starters/consulting-admin-family.md).
