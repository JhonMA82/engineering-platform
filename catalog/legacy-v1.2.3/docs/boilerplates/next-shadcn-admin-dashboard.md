# Next Shadcn Admin Dashboard

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `admin-web-next` |
| Uso predeterminado | Productos híbridos con sitio público y aplicación privada sobre Next.js |
| Repositorio | [https://github.com/arhamkhnz/next-shadcn-admin-dashboard](https://github.com/arhamkhnz/next-shadcn-admin-dashboard) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

La variante Next.js sigue siendo útil, pero deja de ser la elección automática para todo dashboard. Se reserva para productos donde el sitio público, contenido indexable y aplicación privada conviven, o donde React Server Components, Server Actions y el ecosistema Next aporten una ventaja concreta.

## Qué ofrece el repositorio

- Dashboard moderno con Next.js, React, TypeScript, Tailwind CSS y shadcn/ui.
- Amplio conjunto de pantallas y patrones administrativos.
- Componentes para tablas, formularios, gráficas y múltiples layouts.
- Base visual y estructural; no debe confundirse con un backend empresarial terminado.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Producto híbrido con marketing, contenido y área privada.
- SaaS que usará activamente capacidades de Next.js.
- Catálogo o portal público con dashboard autenticado.
- Cliente con infraestructura, equipo o integraciones ya estandarizadas en Next.js.
- Aplicación donde la renderización y composición del servidor reduzcan JavaScript de forma medible.

## Ejemplos por tipo de cliente

- **Gobierno:** portal público + área de gestión en un producto unificado, solo si existe una razón técnica.
- **Escuela:** portal de aspirantes + cuenta privada.
- **Sindicato:** sitio público + área de afiliados.
- **Pyme:** SaaS o plataforma comercial con marketing y producto integrados.

## Cuándo no usarlo

- Herramientas de datos que SpeedPy resuelve con un único stack Python.
- Aplicaciones operativas privadas conectadas a API externa donde TanStack Start sea más explícito.
- Introducir Next.js únicamente por popularidad.
- Asumir que las pantallas demo incluyen auth, RBAC, auditoría o multi-tenancy productivos.

## Ventajas estratégicas

- Ecosistema amplio y abundancia de integraciones.
- Capacidades integradas para contenido, metadatos, streaming y servidor.
- Permite unificar sitio y aplicación cuando esa unificación es valiosa.
- La base visual reduce trabajo inicial de UI.

## Riesgos, madurez y límites

- Los límites servidor/cliente pueden añadir complejidad en dashboards muy interactivos.
- Server Components y caché requieren convenciones claras.
- Puede duplicar backend si se conecta sin criterio con Django o FastAPI.
- La plantilla es principalmente de interfaz.

## Relación con otras opciones del catálogo

- **Frente a TanStack Start:** Next para productos híbridos y ecosistema Next; TanStack como default de operación interna.
- **Frente a Stardrive:** Stardrive gana en sitios públicos principalmente editoriales.
- **Frente a Open SaaS:** Open SaaS incluye capacidades comerciales y backend; este dashboard no.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Separar todas las pantallas mock de capacidades reales.
- [ ] Agregar autenticación, autorización, auditoría y contratos de datos.
- [ ] Crear paquete compartido de UI con la variante TanStack.
- [ ] Añadir pruebas unitarias y E2E.
- [ ] Documentar caché, revalidación, Server Actions y fronteras RSC.
- [ ] Evitar dependencia accidental de servicios Vercel.

## Evaluación AI-friendly

**Media.** La IA puede interpretar erróneamente pantallas demo como módulos implementados. El pack debe incluir un mapa de mocks, rutas, capacidades reales, reglas RSC/client y contratos. Cada feature nueva debe declarar dónde vive la lógica de negocio.

## Despliegue y operación

- Vercel es una opción, no un requisito.
- Validar Docker/VPS cuando el cliente requiera autoalojamiento.
- Documentar estrategia de imágenes, caché y assets.
- Fijar versiones y probar build independiente del entorno del desarrollador.

## Decisión final

**Adoptado como variante especializada de la familia Admin**, no como dashboard predeterminado universal.

## Fuentes oficiales

- [https://github.com/arhamkhnz/next-shadcn-admin-dashboard](https://github.com/arhamkhnz/next-shadcn-admin-dashboard)
- [https://nextjs.org/docs](https://nextjs.org/docs)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
