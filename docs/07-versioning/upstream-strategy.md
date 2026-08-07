# Upstream Strategy — ejemplo

TanStack Admin publica cambios.

```text
upstream/main
↓
vendor-sync
↓
integration/tanstack-1.x
↓
CI + visual checks + harness evals
↓
tag consulting-tanstack-admin@1.4.0
```

Nunca mezclar modificaciones de clientes en la rama espejo de upstream.
