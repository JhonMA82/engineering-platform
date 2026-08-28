# Stardrive

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `public-web` |
| Uso predeterminado | Landing pages, blogs, documentación y sitios públicos |
| Repositorio | [https://github.com/peltmonger/stardrive](https://github.com/peltmonger/stardrive) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Stardrive se conserva como la base para presencia web pública. Su valor no es únicamente Astro: incorpora de forma explícita fundamentos que los agentes suelen omitir —seguridad, SEO, metadatos, accesibilidad, contenido estructurado y documentación para IA—, por lo que encaja con la estrategia de desarrollar sitios rápidos sin convertir cada proyecto en una aplicación React innecesaria.

## Qué ofrece el repositorio

- Astro con TypeScript y Tailwind CSS.
- Configuración centralizada mediante `theme.config.ts`.
- Soporte de internacionalización, blog, eventos, FAQ y documentación basada en Markdown.
- Preparación para `llms.txt`, Schema.org y herramientas WebMCP experimentales.
- AGENTS.md y carpeta `.ai/` con guías para agentes.
- Optimización para Cloudflare Workers, sin impedir otros hosts.
- Bases de accesibilidad, SEO, social previews, RSS, sitemaps, headers y redirects.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Landing pages corporativas o institucionales.
- Blogs, centros de noticias y documentación.
- Portales de programas, convocatorias y eventos.
- Sitios multilingües.
- Presencia pública de un producto que tendrá su aplicación privada en otro frontend.
- Micrositios de campañas con vida limitada y bajo costo operativo.

## Ejemplos por tipo de cliente

- **Gobierno:** portal público de un programa, convocatoria, transparencia temática o agenda de eventos.
- **Escuela:** sitio institucional, oferta educativa, noticias y documentación para aspirantes.
- **Sindicato:** portal informativo, convenios, comunicados y calendario.
- **Pyme:** sitio corporativo, blog, casos de éxito, landing de servicio o captación.

## Cuándo no usarlo

- Backoffice con tablas, filtros, workflows y estado complejo.
- Aplicaciones cuya experiencia principal ocurre autenticada.
- Sistemas que requieren edición colaborativa o tiempo real bidireccional.
- Usarlo como excusa para incluir React en todo el sitio cuando Astro ya cubre el caso.

## Ventajas estratégicas

- Reduce el riesgo de que la IA entregue un sitio visualmente atractivo pero técnicamente incompleto.
- Astro permite una salida estática muy económica y estable.
- La arquitectura de contenido favorece mantenimiento por personal no especializado.
- Puede convivir con React, Vue, Svelte o Solid solo en islas donde sea necesario.
- Es la opción más coherente del catálogo para SEO, contenido y accesibilidad pública.

## Riesgos, madurez y límites

- Es opinionado y contiene contenido de demostración que debe eliminarse.
- La optimización hacia Cloudflare no debe convertirse en dependencia contractual.
- WebMCP y `llms.txt` deben tratarse como capacidades emergentes, no como requisito de negocio.
- Una mala personalización puede dejar páginas, datos o configuraciones del demo.

## Relación con otras opciones del catálogo

- **Frente a Next.js:** elegir Stardrive cuando el problema principal sea contenido público; elegir Next cuando sitio y aplicación privada deban compartir una arquitectura integrada.
- **Frente a TanStack Start:** Stardrive cubre presencia pública; TanStack cubre operación interna.
- **Frente a SpeedPy:** Stardrive no sustituye formularios, datos ni procesos de backend.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Crear presets de marca: institucional, corporativo y campaña.
- [ ] Eliminar automáticamente contenido y archivos exclusivos del demo.
- [ ] Agregar plantillas legales y de privacidad adaptables a México, sin presentarlas como asesoría legal.
- [ ] Definir accesibilidad mínima y prueba Lighthouse en CI.
- [ ] Fijar versión del generador y dependencias.
- [ ] Agregar un mapa de contenido y convenciones en español para OpenCode.

## Evaluación AI-friendly

**Alta.** El repositorio está diseñado expresamente para agentes: incluye AGENTS.md, guías específicas y una estructura centralizada. El pack interno debe agregar reglas sobre tono institucional, contenido verificable, accesibilidad, i18n y prohibición de inventar datos legales o públicos.

## Despliegue y operación

- Preferir generación estática cuando no haya lógica dinámica.
- Usar Cloudflare Workers solo cuando sea conveniente para el cliente.
- Mantener una ruta alternativa documentada para VPS, CDN o hosting estático.
- Validar redirects, cache headers, robots y metadatos por ambiente.

## Decisión final

**Adoptado como boilerplate principal para sitios públicos.** No compite con el dashboard administrativo; forma la capa pública del portafolio.

## Fuentes oficiales

- [https://github.com/peltmonger/stardrive](https://github.com/peltmonger/stardrive)
- [https://astro.build](https://astro.build)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
