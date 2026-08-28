# Start Here

No memorices frameworks. Aprende este ciclo: **intake → Recipe → manifest → cambio → gates → conocimiento**.

## Primeros diez minutos

```bash
make check
./eng recommend --input examples/intakes/school-requests.json
./eng boilerplate evaluate https://github.com/kriasoft/react-starter-kit
```

En la primera salida identifica Recipe, starters, datos, features, skills, gates, exclusiones y advertencias. En la segunda comprueba que una URL ya registrada no genera otra entrada.

## Primer proyecto

1. Copia `templates/project-intake.json` y responde con hechos, no preferencias de framework.
2. Ejecuta `./eng recommend --input <intake>`.
3. Si la selección es correcta, ejecuta `./eng new --from <intake> --output <directorio>`.
4. Lee `.engineering/project.json`, `ARCHITECTURE.md` y `AGENTS.md` generados.
5. Si el resultado dice `blueprint`, falta materialización curada: no lo presentes como starter productivo.
6. Para un cambio usa `eng plan`; al terminar usa `eng check` y los comandos reales del starter.

## Ejemplo mental

Una escuela única necesita solicitudes y aprobación. GP-02 cubre el problema; auth, RBAC y audit son necesarios. Multitenancy, jobs y mobile quedan fuera hasta que un requisito los active.

Después revisa el [ejemplo completo](../13-examples/END_TO_END_SCHOOL_REQUESTS.md) y el [flujo de proyecto nuevo](../04-project-lifecycle/new-project.md).
