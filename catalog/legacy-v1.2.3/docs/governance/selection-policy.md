# Política de selección

## Regla principal

Un boilerplate entra al catálogo solo si cubre una categoría clara o reduce de forma demostrable riesgo, tiempo o costo total.

## Criterios

- Encaje con el problema.
- Madurez.
- Licencia.
- Seguridad.
- Actualización.
- Despliegue.
- Pruebas.
- Observabilidad.
- Accesibilidad.
- i18n.
- AI-friendly real.
- Costo de mantener otro ecosistema.
- Salida/migración.

## Preguntas de descarte

- ¿Duplica una opción existente?
- ¿Solo cambia el diseño?
- ¿Introduce un lenguaje sin beneficio?
- ¿Las funciones son demos?
- ¿Obliga a un proveedor?
- ¿Podemos actualizarlo?
- ¿El cliente puede operarlo?
- ¿Una solución más sencilla cubre el caso?

## Estados

- `selected`: elegido/propuesto por el usuario.
- `recommended`: recomendación principal.
- `recommended_pilot`: útil, pero requiere piloto.
- `specialized_candidate`: solo con condición técnica.
- `internal_planned`: activo propio por construir.
- `infrastructure_pack`: composición de infraestructura.

## Gate de producción

Ningún estado evita la revisión por proyecto. Antes de producción se requiere commit fijado, licencia revisada, threat model, pruebas, despliegue reproducible y plan de actualización.
