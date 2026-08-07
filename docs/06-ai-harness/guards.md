# Guards — ejemplos

- API nueva → OpenAPI diff.
- Tabla/campo nuevo → migration test.
- Recurso tenant → isolation A/B test.
- Permiso → matrix/policy test.
- Bug → regression test.
- UI crítica → E2E/accessibility.

Ejemplo: agregar `priority` toca DB + API; por tanto requiere migration test y contract validation aunque la UI “se vea bien”.
