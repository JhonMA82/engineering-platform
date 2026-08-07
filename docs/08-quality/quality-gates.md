# Quality Gates — con ejemplos

Pipeline base:
```text
format → lint → typecheck → architecture → unit → integration → contracts → security → build
```

## Gates condicionales
- DB: migration test.
- API: OpenAPI diff.
- tenant: isolation suite.
- permission: policy matrix.
- UI: E2E/accessibility cuando importa.
- bug: regression test.

**Ejemplo:** cambiar color de un badge no necesita migration test. Cambiar `role` sí requiere authorization tests.
