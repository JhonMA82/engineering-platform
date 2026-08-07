# Evals — ejemplo

Eval `add-permission`:
- fixture: app GP-02.
- tarea: “crear permiso `request.approve` para DIRECTOR”.
- esperamos: policy backend, test permitido/denegado, UI capability, audit si aplica.
- penalizar: permiso solo en frontend, nueva librería innecesaria, tests ausentes.

Esto permite comparar harness/modelos de forma objetiva.
