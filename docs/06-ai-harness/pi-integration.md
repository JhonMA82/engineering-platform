# Integración Pi-first

Engineering Platform `0.8.0` se distribuye como paquete Pi nativo. `package.json` declara una extensión, los skills de descubrimiento y evolución, los skills operativos y prompts; Pi los carga globalmente mediante su gestor de paquetes.

## Instalación

Desde una copia estable del repositorio:

```bash
./eng install --global --target pi
eng doctor --global
```

El instalador conserva una copia versionada en `~/.local/share/engineering-platform/0.8.0`, crea `~/.local/bin/eng` y ejecuta `pi install` sobre esa copia. Si `~/.local/bin` no está en `PATH`, `eng doctor --global` muestra la ruta que falta.
Al actualizar, retira del registro de Pi las copias locales gestionadas anteriores, aunque alguna ya no esté registrada, y las fuentes Git anteriores de Engineering Platform antes de borrar sus directorios; no toca paquetes externos ni fuentes Git de otros paquetes.

Para retirarla, `eng uninstall --global --target pi` llama a `pi remove` para cada copia versionada gestionada, elimina el launcher propio y conserva cualquier launcher externo o archivo ajeno.

Pi también permite instalar directamente un tag publicado:

```bash
pi install git:github.com/JhonMA82/engineering-platform@v0.8.0
```

La instalación directa habilita `/new-project`, `/evolve-project`, `/engineering-status`, los skills registrados en cada Recipe y `/project-handoff`; el launcher `eng start` requiere además instalar el binario desde un checkout o ZIP.

## Dos entradas, un solo flujo

### Proyecto aún sin carpeta

```bash
eng start school-requests
```

Por defecto crea `~/dev/school-requests` y arranca Pi dentro del destino. Usa `ENG_WORKSPACE` o `--workspace` para otra raíz. El comando admite únicamente `.atl/`, `.gitignore` y `.git` preexistentes; cualquier contenido de usuario detiene el flujo.

### Pi ya está abierto

Dentro de un destino vacío o que contenga únicamente `.atl/`, `.gitignore` y `.git`:

```text
/new-project
```

La extensión verifica el directorio, nombra la sesión y carga el skill. Si la carpeta contiene otro proyecto, se detiene para evitar nesting y sobrescrituras.

### Proyecto ya materializado

Dentro de la raíz del proyecto usa `/evolve-project`. El skill clasifica la necesidad y muestra primero un dry-run: `eng extend` para otra app o servicio, `eng add` para una capacidad. Sustituir stack, base de datos o tenancy continúa siendo una migración con SDD/ADR, no una extensión automática.

## Responsabilidades

- Pi y `project-discovery`: preguntar, confirmar la idea y clasificar la necesidad.
- Engineering Platform: validar la definición, resolver la Recipe, materializar los pins y generar contexto reproducible.
- Boilerplate: aportar código mediante su adapter curado y registrar la procedencia exacta.
- Gentle AI: elegir `direct` o `SDD` y ejecutar el desarrollo a partir de `GENTLE.md`.

La extensión no cambia el `cwd` de una sesión ni auto-confía carpetas. El bootstrap instala las dependencias declaradas por cada adapter y conserva un registro verificable.

## Referencias de compatibilidad

- [Pi Packages](https://pi.dev/docs/latest/packages)
- [Extensions](https://pi.dev/docs/latest/extensions)
- [Skills](https://pi.dev/docs/latest/skills)
- [CLI y sesiones](https://pi.dev/docs/latest/usage)
