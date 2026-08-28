# Ciclo de vida — ejemplo resumido

```mermaid
flowchart TD
    A["Intake: solicitudes internas"] --> B["Recipe GP-02"]
    B --> C["Manifest y blueprint"]
    C --> D["Features y gates"]
    D --> E["Release y operación"]
    E --> F["Incidentes a knowledge"]
```

Cada fase tiene un artefacto verificable. `eng doctor`, `eng plan` y `eng check` acompañan el proyecto después del bootstrap; no depende de memoria oral del equipo.
