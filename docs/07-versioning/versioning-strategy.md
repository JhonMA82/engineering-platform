# Versionado — ejemplo

```text
engineering-platform 0.2.0
hono-api            0.4.0
tanstack-admin      1.3.2
auth-pack           0.2.1
audit-pack          0.1.0
```

Proyecto A registra versiones exactas en manifest.

Si `hono-api 0.5` rompe una API interna, es MAJOR cuando sea estable y debe incluir upgrade recipe. El proyecto no se actualiza por `git merge` con el starter.
