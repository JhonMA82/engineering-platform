# Changelog

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
