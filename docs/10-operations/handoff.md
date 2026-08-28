# Handoff a Gentle y al equipo

Al crear el proyecto, `eng bootstrap` produce dos contratos:

- `GENTLE.md`: idea, stack, boilerplates, patrones, estructura, alcance, exclusiones y estrategia;
- `.engineering/gentle-handoff.json`: el mismo contexto legible por agentes.

Gentle debe leer primero la definición confirmada, el manifest, la arquitectura y `AGENTS.md`. Gentle decide si desarrolla directo o usa SDD según riesgo y ambigüedad; Engineering Platform no impone esa estrategia.

Al entregar el producto, el equipo agrega repo y tag, ADRs, ambientes, mapa de secretos sin valores, deploy/rollback, backup/restore, monitoreo, proveedores y limitaciones conocidas.

Objetivo: operar sin reconstruir mentalmente el proyecto.
