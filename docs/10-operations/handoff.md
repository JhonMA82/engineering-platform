# Handoff a Gentle y al equipo

Al crear el proyecto, `eng bootstrap` produce un handoff y un índice:

- `GENTLE.md`: contexto breve, stack base, patrones, estado e instrucciones;
- `.engineering/gentle-handoff.json`: índice pequeño de fuentes de verdad y estrategia para agentes;
- `.engineering/project-definition.json`: idea, alcance, riesgos y criterios;
- `.engineering/project.json`: Recipe, stack, skills, gates y ownership;
- `.engineering/materialization.json`: fuentes exactas, pins, estructura y verificaciones.

Gentle debe leer primero `GENTLE.md` y después las fuentes que este índice señala. Decide si desarrolla directo o usa SDD según riesgo y ambigüedad; Engineering Platform no impone esa estrategia.

Al entregar el producto, el equipo agrega repo y tag, ADRs, ambientes, mapa de secretos sin valores, deploy/rollback, backup/restore, monitoreo, proveedores y limitaciones conocidas.

Objetivo: operar sin reconstruir mentalmente el proyecto.
