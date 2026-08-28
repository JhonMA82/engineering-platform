# Start Here

No memorices frameworks. Aprende este ciclo: **idea → definición → Recipe → handoff → desarrollo → gates → conocimiento**.

## Primeros diez minutos

```bash
make check
./eng install --global --target pi --dry-run
./eng bootstrap --from examples/project-definitions/school-requests.json --output /tmp/school-requests --dry-run
./eng boilerplate evaluate https://github.com/kriasoft/react-starter-kit
```

En la primera salida identifica Recipe, starters, datos, features, skills, gates, exclusiones y advertencias. En la segunda comprueba que una URL ya registrada no genera otra entrada.

## Primer proyecto

1. Instala la integración Pi con `./eng install --global --target pi`.
2. Ejecuta `eng start <nombre>` o abre Pi en una carpeta vacía y usa `/new-project`.
3. Responde sobre problema, usuarios, resultado, alcance y restricciones; no elijas framework.
4. Confirma la definición y lee `.engineering/project.json`, `GENTLE.md`, `ARCHITECTURE.md` y `AGENTS.md` generados.
5. Si el resultado dice `blueprint`, falta materialización curada: no lo presentes como starter productivo.
6. Para un cambio usa `eng plan`; al terminar usa `eng check` y los comandos reales del starter.

## Ejemplo mental

Una escuela única necesita solicitudes y aprobación. GP-02 cubre el problema; auth, RBAC y audit son necesarios. Multitenancy, jobs y mobile quedan fuera hasta que un requisito los active.

Después revisa el [ejemplo completo](../13-examples/END_TO_END_SCHOOL_REQUESTS.md) y el [flujo de proyecto nuevo](../04-project-lifecycle/new-project.md).
