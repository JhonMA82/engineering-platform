---
name: architecture-selector
description: Selecciona una Project Recipe a partir del problema, restricciones y clientes del sistema.
---

# Architecture Selector

## Resultado obligatorio

Entrega una selección explícita de Recipe, starters, base de datos, features, skills, gates y exclusiones. Explica cada excepción al default; no produzcas una lista abierta de tecnologías.

## Flujo

1. Lee `platform/golden-paths.json`, `platform/boilerplates.json`, `platform/database-profiles.json` y `platform/feature-packs.json`.
2. Obtén `name`, `project_type`, usuarios, clientes, integraciones, restricciones operativas y capacidades realmente necesarias.
3. Ejecuta `eng recommend --input <intake.json>`. Si responde «... no permite componer ... reintentá con `project_type='multi-app'`», corregí el intake a `project_type: multi-app` y reintentá una vez; solo si sigue sin Recipe exacta, avanzá al paso 4.
4. Si no existe Recipe exacta, detente y propone una decisión de plataforma; no mezcles Golden Paths de forma improvisada.
5. Prefiere el camino estable más simple. Un starter especializado exige que el intake contenga la señal que lo justifica. Si las señales favorecen a una alternativa curada de la misma categoría (`next-ecosystem` → `next-admin`), `recommend` la selecciona sola y lo avisa; no hace falta forzarla.
6. Enumera lo que deliberadamente no se incluirá.

## Reglas

- Single-tenant es primera clase; no activar multitenancy por anticipación.
- PWA antes que app nativa si cámara, GPS, push, offline o tienda no aportan valor comprobable.
- Monolito modular antes que microservicios.
- Dos o más clientes (dashboard + portal público/kiosco/móvil), o uno más otro futuro explícito, → `project_type: multi-app` (GP-06) con todas las surfaces compartiendo la misma API; un solo cliente sin futuro → su Recipe específica sin surfaces.
- Turso no es default general: usa el perfil concreto y respeta su canal.
- Un `catalog-only` genera blueprint, nunca una promesa de código listo para producción.
