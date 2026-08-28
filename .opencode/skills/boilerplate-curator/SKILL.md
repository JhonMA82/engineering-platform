---
name: boilerplate-curator
description: Evalúa una URL de boilerplate, evita duplicados y registra solo opciones con una ventaja demostrable.
---

# Boilerplate Curator

## Objetivo

Ante una URL, responde una de estas decisiones y actualiza los artefactos correspondientes:

- `ALREADY_REGISTERED`
- `ALREADY_REGISTERED_REFRESH`
- `ADD_AS_CANDIDATE`
- `ADD_AS_REFERENCE`
- `ADD_AS_INFRASTRUCTURE_PACK`
- `ADD_AS_FEATURE_PACK`
- `REJECT_DUPLICATE`
- `REJECT_NO_MAINTENANCE`
- `REJECT_LICENSE`
- `REJECT_EXCESS_COMPLEXITY`
- `REPLACE_EXISTING_CANDIDATE`

## Flujo obligatorio

1. Normaliza la URL y ejecuta `eng boilerplate evaluate <url>`.
2. Busca coincidencia exacta, alias histórico, rename, fork y starter que cubra la misma categoría, incluyendo entradas deprecated o rechazadas.
3. Inspecciona fuente primaria: README, licencia, manifiestos, CI, actividad reciente, releases, estrategia de actualización y guías para agentes.
4. Separa capacidades declaradas de capacidades verificadas. Prueba instalación, tests y build en un piloto antes de `pilot-ready`.
5. Compara contra el default actual con criterios: tiempo hasta primer cambio, mantenimiento, seguridad, portabilidad, calidad, actualización y ajuste a proyectos reales.
6. Decide. “Popular” o “más completo” no son brechas.

## Registro de una entrada nueva

Actualiza en la misma operación:

- `platform/boilerplates.json`, usando un id corto estable y `legacy_ids` para renames;
- una ficha de evidencia en `catalog/profiles/`;
- Recipe o comparación afectada, solo si cambia una decisión;
- `CHANGELOG.md` y un ejemplo/eval de la decisión;
- pin, licencia observada, fecha, delivery status e integración.

Las altas empiezan `catalog-only` y nunca se convierten en default solo por registrarse.

## Actualización de una entrada existente

Conserva el id. Usa la estrategia declarada en `integration.update_strategy`. Para React Starter Kit respeta su mecanismo `seed-fork`/`merge-seed`: upstream posee mecanismos; el proyecto posee identidad y alcance. Ejecuta sus verificaciones documentadas antes de cambiar el pin.

## Caso canónico

`https://github.com/kriasoft/react-starter-kit` ya es `react-starter-kit`. No crear otra entrada. Si el commit observado difiere del pin, devuelve `ALREADY_REGISTERED_REFRESH`; si coincide, `ALREADY_REGISTERED`.
