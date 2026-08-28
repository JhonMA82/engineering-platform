# Versionado

## Catálogo

SemVer:

- MAJOR: cambio incompatible de estructura o política.
- MINOR: nueva entrada o documento.
- PATCH: corrección sin cambio de decisión.

## Starters internos

Cada starter tendrá sus propios tags:

- `pilot`.
- `production-approved`.
- `deprecated`.
- `reviewed-YYYY-MM-DD`.

## Upstream pinning

Registrar siempre:

- URL.
- commit.
- release/tag.
- fecha.
- licencia observada.
- patches internos.
- resultado de tests.

No iniciar un proyecto desde `main` sin snapshot reproducible.
