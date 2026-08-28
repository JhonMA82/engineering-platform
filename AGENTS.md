# AGENTS.md

## Propósito

Gobierna cómo un agente modifica esta Engineering Platform. Para un proyecto generado prevalece su propio `AGENTS.md` y `.engineering/project.json`.

## Fuente de verdad

1. `platform/boilerplates.json` para entradas y madurez;
2. `platform/golden-paths.json` para Recipes;
3. `platform/database-profiles.json` y `platform/feature-packs.json` para composición;
4. `skills/registry.json` para routing;
5. fichas y ADRs para evidencia y motivo.

## Antes de cambiar la plataforma

1. Lee el registro afectado, su schema, ficha, Recipe y tests.
2. Distingue cambio de proyecto de cambio reusable de plataforma.
3. Conserva ids; usa `legacy_ids` para renames.
4. No promociones `delivery_status` sin artefacto, pin, checks y piloto correspondientes.
5. Actualiza registro, ficha, ejemplo/eval, validator y changelog juntos.
6. Ejecuta `make check`.

## Boilerplates

Ante una URL usa el skill `boilerplate-curator` y `./eng boilerplate evaluate`. Revisa coincidencia exacta, alias, fork y cobertura antes de añadir. Una candidata inicia `catalog-only`; popularidad no justifica un default.

## Invariantes

- arquitectura mínima suficiente;
- single-tenant y monolito modular como opciones de primera clase;
- frontend nunca es autoridad de seguridad;
- todo cambio de datos tiene migración y recuperación;
- todo bug produce regression test;
- recursos tenant requieren prueba de aislamiento;
- no usar `latest` ni marcadores como `PINNED`;
- no copiar referencias sin licencia compatible;
- no afirmar que un starter está listo si solo existe su ficha.
