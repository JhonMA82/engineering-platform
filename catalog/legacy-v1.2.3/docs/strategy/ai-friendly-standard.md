# Estándar AI-friendly de la consultoría

Un repositorio no es AI-friendly solo por incluir `AGENTS.md`.

## Requisitos

### Contexto

- `AGENTS.md` corto como índice.
- Mapa de arquitectura.
- Glosario.
- Decisiones ADR.
- Invariantes de dominio.
- Ejemplos canónicos.
- Mocks identificados.
- Generated files identificados.

### Ejecución

- Comandos de setup, test, lint y build.
- Datos de ejemplo reproducibles.
- Validaciones automáticas.
- Scripts para generar features.
- Límites de cambios permitidos.

### Calidad

- Tests de dominio.
- Tests de contrato.
- E2E de flujos críticos.
- Type checking.
- Lint.
- Migraciones probadas.
- Observabilidad.

### Seguridad

- Matriz de roles.
- Catálogo de herramientas de IA.
- No ejecutar side effects sin autorización.
- Secrets fuera del repositorio.
- Datos sensibles y retención documentados.
- Prompt injection tratada como amenaza.

## Pack por boilerplate

Cada upstream seleccionado debe convertirse en un pack curado con:

```text
AGENTS.md
docs/architecture.md
docs/glossary.md
docs/decisions/
docs/patterns/
docs/ai/project-map.yaml
docs/ai/canonical-examples.yaml
scripts/validate-architecture.*
scripts/generate-ai-context.*
templates/
```

El pack debe adaptarse al repositorio; no copiar una estructura universal sin revisar el stack.
