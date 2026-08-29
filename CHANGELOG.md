# Changelog

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
