# TanStack Shadcn Admin Dashboard

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario y confirmado como incorporación útil |
| Categoría | `admin-web-tanstack` |
| Uso predeterminado | Aplicaciones operativas privadas y frontends conectados a APIs |
| Repositorio | [https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard](https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

La variante TanStack Start pasa a ser la opción preferente para aplicaciones administrativas privadas, dashboards operativos y frontends desacoplados. La razón no es moda: routing y search params tipados, loaders explícitos, integración natural con TanStack Query, Vite y menor fricción para UIs mayormente interactivas.

## Qué ofrece el repositorio

- Versión oficial del mismo dashboard construida con TanStack Start.
- TanStack Router, TypeScript, React, Vite, shadcn/ui, Tailwind CSS y TanStack Table.
- Estructura orientada a rutas y componentes colocados.
- SSR y funciones de servidor disponibles sin exigir replicar la arquitectura RSC de Next.js.
- AGENTS.md con convenciones del proyecto.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- CRM, ERP ligero y backoffice.
- Trámites, expedientes y seguimiento.
- Inventarios y mesas de ayuda.
- Portales escolares o sindicales.
- Aplicaciones con filtros, tablas y estado de URL complejo.
- Frontend para Django, FastAPI, Go, .NET, Directus o APIs institucionales.
- Despliegues autoalojados donde Vite y runtime explícito sean convenientes.

## Ejemplos por tipo de cliente

- **Gobierno:** tablero operativo, gestión de solicitudes, control de áreas o expedientes.
- **Escuela:** administración académica, incidencias y procesos internos.
- **Sindicato:** afiliados, beneficios, gestiones y seguimiento.
- **Pyme:** CRM, operaciones, inventario y portal B2B.

## Cuándo no usarlo

- Sitios centrados en contenido donde Stardrive sea suficiente.
- Productos que dependan profundamente de capacidades específicas de Next.js.
- Procesos de datos sencillos que SpeedPy resuelve sin separar frontend y backend.
- Usarlo sin fijar versiones en una fase de evolución rápida del framework.

## Ventajas estratégicas

- Search params tipados para dashboards con filtros complejos.
- Carga de datos y caché explícitas.
- Frontend natural para backends independientes.
- Menor ruido de `use client` en interfaces altamente interactivas.
- Vite y composición más neutral respecto al proveedor.

## Riesgos, madurez y límites

- TanStack Start tiene un ecosistema menor que Next.js.
- La plantilla sigue siendo una capa de UI, no un sistema completo.
- Autenticación, organizaciones, RBAC, auditoría y persistencia deben implementarse.
- Dependencias centrales deben fijarse y probarse.

## Relación con otras opciones del catálogo

- **Frente a Next.js:** TanStack es el default para operación privada; Next se reserva para producto híbrido o capacidades específicas.
- **Frente a SpeedPy:** TanStack conviene si habrá API separada y UI React rica; SpeedPy para un monolito Python productivo.
- **Frente al template FastAPI:** pueden combinarse, pero hay que decidir si se conserva o sustituye el frontend incluido en el template.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Fijar versiones concretas de TanStack Start, Router y runtime.
- [ ] Agregar pruebas unitarias, integración y E2E.
- [ ] Definir adapters de API, manejo de errores y autenticación.
- [ ] Crear generadores de features y contratos de búsqueda.
- [ ] Compartir tokens, layouts y componentes con la variante Next.
- [ ] Documentar route tree generado y no editarlo manualmente.

## Evaluación AI-friendly

**Alta potencialmente.** Routing, loaders, search schemas y contratos tipados son explícitos y comprensibles para agentes. El AGENTS.md interno debe explicar generated files, invalidación de Query, funciones de servidor y límites de dominio.

## Despliegue y operación

- Probar el runtime objetivo antes de cerrar arquitectura.
- Preferir Docker/VPS o plataforma compatible según cliente.
- Mantener observabilidad coordinada con el backend.
- No asumir compatibilidad de cualquier plugin de Vite con SSR sin prueba.

## Decisión final

**Adoptado como dashboard operativo predeterminado** dentro de la familia Admin.

## Fuentes oficiales

- [https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard](https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard)
- [https://tanstack.com/start/latest](https://tanstack.com/start/latest)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
