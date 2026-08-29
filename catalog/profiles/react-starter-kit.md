# React Starter Kit

| Campo | Decisión |
|---|---|
| ID canónico | `react-starter-kit` |
| Decisión | `specialized` |
| Entrega | `curated` |
| Categoría | SaaS full-stack Cloudflare-first |
| Repositorio | <https://github.com/kriasoft/react-starter-kit> |
| Commit observado | `0aa7603435f16159ad0b8fef68fb7f6280be7ca1` |
| Revisión | 2026-08-28 |
| Licencia observada | MIT |

## Decisión de catálogo

La URL ya existía en el catálogo anterior y en Engineering Platform. Se conserva un solo id; proponerla otra vez produce `ALREADY_REGISTERED`, o `ALREADY_REGISTERED_REFRESH` si el commit observado es distinto.

Es el default de GP-07 únicamente cuando existen billing, organizaciones y aceptación explícita del enfoque Cloudflare. No reemplaza GP-02 para aplicaciones institucionales single-tenant.

## Stack observado en la fuente

- React 19, TanStack Router y Query, Jotai, Tailwind CSS 4 y shadcn/ui.
- tRPC, Hono y Drizzle sobre Neon PostgreSQL.
- Better Auth con capacidades de organización y billing del producto.
- Cloudflare Workers, Bun, TypeScript, Vite/Vitest y Terraform.
- Estructura multi-app con web Astro, aplicación React, API y paquetes compartidos.
- `AGENTS.md`, skills para `merge-seed` y shadcn, y registro de skills.

Estos puntos fueron inspeccionados en el repositorio al commit indicado. No equivalen a que instalación, tests, build, deploy y migración hayan pasado en infraestructura de la consultoría.

## Estrategia de actualización

El upstream incluye una estrategia nativa de forks con remoto `seed`. La plataforma registra `integration.mode: seed-fork` y `update_strategy: merge-seed`.

Antes de actualizar se lee `.agents/skills/merge-seed/SKILL.md` del upstream. Se preservan identidad y alcance del proyecto, no se reescriben migraciones aplicadas y se ejecutan instalación congelada, typecheck, lint, tests y build antes de cambiar el pin.

## Usar cuando

- El producto es SaaS comercial, no solo una app interna.
- Billing y organizaciones forman parte del MVP.
- Cloudflare y el stack completo son aceptables contractual y operacionalmente.
- El equipo puede mantener monorepo, infraestructura y estrategia seed-fork.

## Evitar cuando

- Una institución única necesita workflows y auditoría sin billing.
- Se exige neutralidad de proveedor o API pública como frontera principal.
- El producto cabe en GP-01 o GP-02 con menos capas.
- No existe presupuesto para piloto, upgrades y validación de migraciones.

## Para promover a curated/released

- [ ] Ejecutar un piloto desde el pin registrado.
- [ ] Guardar resultados de instalación, types, lint, tests y build.
- [ ] Probar auth, organizaciones, billing sandbox, migraciones y rollback.
- [ ] Validar deploy y observabilidad en la cuenta objetivo.
- [ ] Documentar patches internos y una actualización real por merge-seed.
- [ ] Publicar artefacto interno y versión antes de `released`.
