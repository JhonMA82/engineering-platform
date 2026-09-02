# Changelog

## [Unreleased]

## [0.8.0] - 2026-09-02

- Completada la evidencia de AI Assistant y React Starter Kit; `eng boilerplate verify` ahora refleja en su código de salida cuando la verificación falla y conserva el JSON.
- Separada la verdad de capacidades: cada feature queda `materialized`, `verified` o `pending-implementation`; Engineering ya no presenta como implementado lo que corresponde desarrollar a Gentle.
- Añadido preflight declarativo de runtimes y versiones antes de clonar o instalar, con Bun 1.4 para API Starter 0.12.2.
- Añadido `eng extend` transaccional para incorporar un starter compatible a un proyecto existente sin regenerarlo, además del flujo Pi `/evolve-project`.
- Generado un workflow CI raíz por proyecto compuesto, manteniendo los checks propios de cada boilerplate y sus directorios de trabajo.
- Añadidos el workflow `release-pilots` y `scripts/release_pilot.py` para materializar y verificar los seis Recipes estables con evidencia descargable.
- Añadido un entorno nativo determinista y declarado para comandos de materialización de starters curados, incluido Stardrive.
- Fijado el generador anidado de la fuente Tauri UI (`create-tauri-app@4.6.2`) para reproducir el scaffold de escritorio.
- Conservada compatibilidad de lectura con proyectos 0.7; el siguiente `add` o `extend` completa los metadatos nuevos sin regenerar el proyecto.
- Reducido el ruido AI de fuentes upstream mediante poda declarativa y overlays compactos, preservando una fuente de verdad por decisión.
- Corregida la salida curada de Ignite 11.5.0 con dependencias Expo fijadas y un guard mínimo de navegación, manteniendo typecheck y tests activos.
- Endurecidos schemas, doctor, validadores y ownership para extensiones y manifiestos evolutivos.

## [0.7.0] - 2026-08-31

- Integrado `api-starter v0.12.0` como implementación real de `hono-api`, conservando el id estable y materializando perfiles exactos en `services/api` mediante el nuevo adapter genérico `git-generator`.
- Añadido mapping declarativo de PostgreSQL y features; las capacidades no equivalentes quedan explícitas para Gentle en vez de activar módulos inesperados.
- Auditados los seis boilerplates principales y añadida evidencia AI-friendly, overlays para Stardrive, Ignite y Tauri UI, y preservación de instrucciones upstream sin duplicarlas en el handoff.
- Corregido Tauri UI para ejecutar el generador fijado en lugar del shell interno provisional.
- Generalizados validator y adapters: overlays, evidencia y estrategias se validan sin condiciones especiales por boilerplate.
- Añadidos `eng boilerplate verify`, `add` y `remove`, con dry-run predeterminado y bloqueo de bajas todavía referenciadas por Recipes.
- Reducidos `ARCHITECTURE.md` y `GENTLE.md` para que la idea, decisiones y materialización permanezcan en fuentes únicas.
- Actualizados Recipes y ownership para composiciones `services/api` + `apps/*`, además de pruebas de mapping, overlays y gestión segura del catálogo.

## [0.6.0] - 2026-08-30

- Reducido el handoff para que `GENTLE.md` y `.engineering/gentle-handoff.json` apunten a fuentes únicas en lugar de copiar materialización, árbol y verificaciones.
- Conservados los checks no seleccionados en ejecuciones parciales y reutilizado el setup ya aprobado; añadido `--force-setup` para repetirlo.
- Conservados `.env.example` y `.env.sample` en la instalación global; los proyectos materializados inicializan Git en `main` y usan schemas versionados.

## [0.5.4] - 2026-08-29

- Corregido el instalador global para retirar fuentes Git anteriores de Engineering Platform y evitar colisiones con la copia local actual en Pi.
- Añadidas regresiones para preservar paquetes externos y paquetes instalados en el ámbito del proyecto.

## [0.5.3] - 2026-08-29

- Corregido el instalador global para tolerar copias locales gestionadas que ya no están registradas en Pi, sin ocultar fallos reales de retiro.
- Añadida una regresión para el escenario de directorio gestionado sin registro.

## [0.5.2] - 2026-08-29

- Corregido `eng install --global --target pi` para retirar del registro las instalaciones locales versionadas anteriores antes de eliminar sus directorios, evitando que Pi cargue skills obsoletos.
- Añadidas regresiones para preservar las instalaciones cuando `pi remove` falla y conservar paquetes externos.

## [0.5.1] - 2026-08-29

- Ajustado `project-discovery` para iniciar con la idea libre del usuario, preguntar progresivamente sin formularios ni opciones predeterminadas y registrar la procedencia del patrón de entrevista.

## [0.5.0] - 2026-08-29

- `eng bootstrap` ahora materializa código real de forma transaccional desde adapters con pin exacto.
- Curados Stardrive, TanStack Admin, Ignite, Tauri UI, SpeedPy y React Starter Kit; liberados Hono API y Governed AI Assistant internos.
- Añadidos `.engineering/materialization.json`, readiness `code-ready|verified`, remotes `seed-*` y detección del árbol, paquetes, scripts y archivos de entorno reales.
- `GENTLE.md` recibe idea, fuentes, stack, patrones, estructura real, criterios y evidencia para elegir ejecución directa o SDD.
- `eng check --run` ejecuta checks registrados; `eng update --check` consulta ramas upstream reales.
- Corregido `/new-project` para admitir `.atl/`, `.gitignore` y `.git` sin aceptar contenido de usuario.
- Ampliada la suite a 38 pruebas, incluyendo materialización y rollback transaccional.

## [0.4.4] - 2026-08-28

- Migración automática del launcher gestionado entre instalaciones versionadas.
- Desinstalación limpia de todas las copias gestionadas y sus registros en Pi, preservando archivos ajenos.
- Añadidas regresiones para la migración y la limpieza segura de instalaciones globales.

## [0.4.3] - 2026-08-28

- Derivada la versión de la extensión Pi y del CLI `eng` desde `package.json`.
- Corregido el frontmatter del skill `project-discovery` para aceptar correctamente valores con dos puntos.
- Añadida una regresión para el launcher ejecutado mediante un enlace simbólico.

## [0.4.2] - 2026-08-28

- Formalizada la política de entrega versionada antes del primer push y la preservación de los tags anteriores.

## [0.4.1] - 2026-08-28

- Corregido el launcher global cuando `eng` se invoca mediante un enlace simbólico.
- Permitido reutilizar destinos que contienen únicamente `.atl/` y `.gitignore`, preservando el rechazo de contenido de usuario.
- Añadidas regresiones para la resolución del launcher y la seguridad de destinos con metadatos gestionados.

## [0.4.0] - 2026-08-28

- Convertida la plataforma en paquete Pi nativo con extensión TypeScript, descubrimiento, 13 skills operativos y prompt declarados en `package.json`.
- Añadidos `/new-project` y `/engineering-status` para descubrir proyectos y comprobar manifests desde Pi.
- Añadidos `eng install --global --target pi`, diagnóstico y desinstalación segura de la copia versionada.
- Añadido `eng start <nombre>` para crear `~/dev/<nombre>` y abrir Pi en el directorio correcto sin nesting.
- Añadido contrato `project-definition` y flujo progresivo idea → confirmación → Recipe → bootstrap.
- Añadidos `eng bootstrap` y `eng handoff`, que generan contexto estructurado y `GENTLE.md`.
- Gentle conserva la decisión entre desarrollo directo y SDD; Engineering Platform fija idea, stack, patrones, estructura, alcance y gates.
- Conservados los estados reales de boilerplates: los starters sin artefacto `released` continúan como blueprint.
- Ampliada la suite a 24 pruebas, incluyendo rutas seguras, paquete Pi, instalación aislada y handoff.

## [0.3.0] - 2026-08-28

- Recuperado el catálogo histórico `1.2.3`, con 17 entradas, perfiles internos, comparaciones y gobierno.
- Añadido registro normalizado con ids estables, aliases, decisión, entrega, Tier e integración.
- Convertidos Golden Paths en ocho Project Recipes con stack, datos, features, skills, gates y exclusiones.
- Añadidos perfiles PostgreSQL, SQLite y Turso con canales y gates explícitos.
- Implementado CLI `eng`: catalog, curator, recommend, new, doctor, plan, check y update.
- Añadidos 13 skills ejecutables para selección, bootstrap, curación, datos, seguridad y conocimiento.
- Revisado React Starter Kit al commit `0aa7603`, conservado en `pilot-ready`, con adapter y Upgrade Recipe `merge-seed`.
- Eliminado el marcador `PINNED`; los artefactos sin release se declaran como blueprint.
- Añadidos formularios de intake, evals, feature packs `billing` y `sync`, schemas y Makefile.
- Añadido validator cruzado, con 16 pruebas y CI.
- Recuperados formularios de propuesta, revisión y deprecación.

## [0.2.0] - 2026-08-07

- Ruta de aprendizaje para juniors.
- Glosario con ejemplo y error típico por concepto.
- Ejemplo end-to-end de sistema escolar single-tenant.

## [0.1.0]

- Bootstrap de Engineering Platform.
