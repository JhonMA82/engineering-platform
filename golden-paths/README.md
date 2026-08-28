# Golden Paths y Project Recipes

Los documentos `GP-*.md` explican el propósito humano. La definición ejecutable está en `platform/golden-paths.json` y contiene stack, alternativas, perfiles de datos, features, skills, gates y exclusiones.

```bash
./eng recommend --input examples/intakes/school-requests.json
```

Una excepción modifica el intake y vuelve a resolver; no se cambia el stack directamente en el proyecto.
