# Consulting Admin Family

## Propósito

Evitar que las variantes Next.js y TanStack Start se conviertan en dos forks que divergen y duplican cada corrección visual.

## Estructura propuesta

```text
consulting-ui-system/
├── tokens/
├── themes/
├── icons/
├── patterns/
├── dashboard-specs/
├── shared-domain-components/
└── shadcn-registry/

starters/
├── next-admin/
└── tanstack-admin/
```

## Regla de selección

- **TanStack Start:** opción predeterminada para aplicaciones operativas privadas, tablas, filtros y backend independiente.
- **Next.js:** opción específica para productos híbridos con sitio público, contenido y aplicación privada, o cuando las capacidades de Next sean parte del diseño.

## Qué sí compartir

- Tokens de diseño.
- Temas.
- Componentes sin dependencia del router.
- Layouts y patrones de tabla/formulario.
- Accesibilidad.
- Especificaciones de dashboard.
- Schemas de dominio y contratos de UI.
- Registry privado de shadcn.
- Generadores de features.

## Qué no forzar a compartir

- Router.
- Loaders.
- Server Actions / server functions.
- Manejo de caché.
- Metadatos de páginas.
- Integraciones específicas del runtime.

## Resultado esperado

Una sola identidad técnica y visual, con dos adapters de framework. Cada corrección común debe aplicarse en el paquete compartido; cada decisión de framework queda en su starter.

[Comparación Next vs. TanStack](../comparisons/next-vs-tanstack.md)
