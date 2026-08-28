# Integración Pi-first

Engineering Platform `0.4.0` se distribuye como paquete Pi nativo. `package.json` declara una extensión, el skill de descubrimiento, los 13 skills operativos y prompts; Pi los carga globalmente mediante su gestor de paquetes.

## Instalación

Desde una copia estable del repositorio:

```bash
./eng install --global --target pi
eng doctor --global
```

El instalador conserva una copia versionada en `~/.local/share/engineering-platform/0.4.0`, crea `~/.local/bin/eng` y ejecuta `pi install` sobre esa copia. Si `~/.local/bin` no está en `PATH`, `eng doctor --global` muestra la ruta que falta.

Para retirarla, `eng uninstall --global --target pi` llama primero a `pi remove` y después elimina únicamente el launcher propio y la copia `0.4.0`.

Pi también permite instalar directamente un tag publicado:

```bash
pi install git:github.com/JhonMA82/engineering-platform@v0.4.0
```

La instalación directa habilita `/new-project`, `/engineering-status`, `/skill:project-discovery`, los skills registrados en cada Recipe y `/project-handoff`; el launcher `eng start` requiere además instalar el binario desde un checkout o ZIP.

## Dos entradas, un solo flujo

### Proyecto aún sin carpeta

```bash
eng start school-requests
```

Por defecto crea `~/dev/school-requests` y arranca Pi dentro del destino. Usa `ENG_WORKSPACE` o `--workspace` para otra raíz. El comando rechaza nombres fuera de kebab-case y destinos no vacíos.

### Pi ya está abierto

Dentro de una carpeta vacía:

```text
/new-project
```

La extensión verifica el directorio, nombra la sesión y carga el skill. Si la carpeta contiene otro proyecto, se detiene para evitar nesting y sobrescrituras.

## Responsabilidades

- Pi y `project-discovery`: preguntar, confirmar la idea y clasificar la necesidad.
- Engineering Platform: validar la definición, resolver la Recipe y generar contexto reproducible.
- Boilerplate: aportar una base solo conforme a su `delivery_status`; `pilot-ready` y `catalog-only` siguen siendo blueprint.
- Gentle AI: elegir `direct` o `SDD` y ejecutar el desarrollo a partir de `GENTLE.md`.

La extensión no cambia el `cwd` de una sesión, no auto-confía carpetas y no instala dependencias del proyecto. Estas decisiones evitan comportamiento implícito y respetan el modelo de seguridad de Pi.

## Referencias de compatibilidad

- [Pi Packages](https://pi.dev/docs/latest/packages)
- [Extensions](https://pi.dev/docs/latest/extensions)
- [Skills](https://pi.dev/docs/latest/skills)
- [CLI y sesiones](https://pi.dev/docs/latest/usage)
