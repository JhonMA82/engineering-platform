# Changelog

## [Unreleased]

## [0.10.0] - 2026-09-05

- README con tabla de boilerplates (cuándo se usa, dónde y URL, generada desde el registro) y retirada la fila GP-08.

- `next-admin` curado y seleccionable: piloto en verde (npm ci, biome check, next build) en el pin `15e0a081`, adapter `overlay` → `apps/admin`, evidencia AI-friendly y `delivery_status: curated`; `recommend` ahora auto-selecciona una alternativa curada de la misma categoría cuando sus señales pesan más que las del default (`next-ecosystem` → `next-admin` con aviso; empates y alternativas no materializables conservan el default).

- Se retira `self-hosted-ai-starter-kit` del catálogo (11 entradas; sin uso en Recipes).

- Borrado total de GP-08: se retiran `ai-assistant-starter` (entrada, `starters/ai-assistant/`, `curation/ai-assistant-starter/`), la Recipe GP-08 con su ficha, `ai-assistant` del enum de `project_type`, el eval `governed-assistant` y 6 tests de materialización offline (sin starter local, esa cobertura se pierde; el resto del flujo se verifica en blueprint).
- Restaurado `goship` en el catálogo (13 entradas) tras confirmar que sí tiene repo.

- Catálogo reducido a 12 entradas: se retiran `consulting-admin-family`, `speedpy-lite`, `institutional-operations-starter`, `python-service-starter`, `vercel-chatbot`, `open-saas` y `goship` (sin uso previsto); se limpian alternativas de GP-03/GP-07/GP-08, `solution_packs` de GP-02, fichas y matriz de escenarios. GP-08 se conserva con `ai-assistant-starter` pendiente de revisión.

- El handoff enlaza la documentación por starter: los paths `*.md` de cada `evidence.json` llegan a `composition.starter_docs` y a la sección «Documentación por starter» de `GENTLE.md` (p. ej. `apps/intake/docs/API-CONTRACT.md`), para que Gentle no tenga que descubrirlos.

- Handoff más AI-friendly: tildes consistentes en `GENTLE.md`, las capacidades pendientes indican destino por defecto (`services/api`), nueva regla de aceptación por capability (API → `contract`+`integration`, clientes → `build`+`security`, datos → `migration`), y `AGENTS.md` lista solo skills post-bootstrap.

- Test de scaffold punta a punta (`tests/test_scaffold_flow.py`, 3 escenarios): idea → blueprint generado; verifica starters/surfaces elegidos, archivos (`GENTLE.md`, `AGENTS.md`, `ARCHITECTURE.md`, `.engineering/*`), contenidos clave (`future_surfaces`, sección Superficies futuras) y `doctor` limpio.

- Test de flujo de intención (`tests/test_intent_flow.py`, 7 escenarios): idea completa, solo señales, admin que pide portal (error + autocorrección vía `--suggest`), admin con solo señales, vocabulario inventado (did-you-mean), móvil futuro (Recipe componible + `evolution_hints`) y admin solo mínimo.

- Discovery con opciones guiadas: `project-discovery` ahora exige la tool nativa de preguntas de Pi (`ask_user_question`) con 2-4 opciones predefinidas de necesidades/clientes/alcance (nunca stack ni Recipes) más la fila de escritura libre de la tool; se elimina la prohibición anterior de selectores.

- La idea manda aunque el intake venga imperfecto: `recommend`/`bootstrap`/`new` ahora infieren Surfaces desde señales de dominio (`kiosk-mode`/`tracking-token` → `public-intake` con esas capabilities, solo si la señal pertenece a una única surface) y avisan en `warnings`; la inferencia también filtra la Recipe (`admin` + señales de kiosco sugiere `multi-app` en vez de un dashboard mudo). Sumados 3 tests de inferencia, eval `portal-signals-only` y notas en `flow-scenarios.md` y `project-discovery`.

- Flujo por intención de cliente (8 refuerzos): `skills/registry.json` registra `project-discovery` y `project-evolution` (antes huérfanos); `eng recommend --suggest` devuelve el intake corregido a `multi-app` con su manifest en vez de solo fallar; `project-discovery` valida con `recommend --suggest` provisional en cada ronda; nuevo `SURFACE_SYNONYMS` con sugerencias did-you-mean ante capabilities inventadas (`qr-capture` → `form-capture, tracking-token`); nuevo `evals/discovery-cases.json` con 4 casos usuario→intake→Recipe (incluye cliente móvil futuro sin surfaces iniciales); `eng doctor` suma `evolution_hints` con las surfaces componibles aún no instaladas; el handoff (`gentle-handoff.json` + `GENTLE.md`) declara `future_surfaces` con guardarraíles de composición; y el help de `eng` agrupa comandos por fase (`descubrir/componer/evolucionar/verificar/entorno`). Documentación sincronizada: `flow-scenarios.md` refleja los mensajes y matrices reales (verificados contra el motor), `new-project.md` documenta la validación provisional y la regla `multi-app`, y `skills-and-routing.md` registra el routing por registry de discovery/evolution.

- Ruteo por intención de cliente en `recommend`: la selección de Recipe ahora filtra por las Surfaces solicitadas y, si el `project_type` no permite componerlas, falla con la corrección exacta (`reintentá con project_type='multi-app'`, GP-06); GP-06 amplía sus señales de dominio (`public-intake`, `kiosk-mode`, `tracking-token`, `offline-sync`, `form-capture`, `offline-outbox`, `pwa-installable`, `field-app`, `camera`, `push`, `offline`); `project-discovery` y `architecture-selector` documentan la tabla clientes→intake (dos o más clientes, o uno más otro futuro explícito, → `multi-app` compartiendo la misma API; el vocabulario de capabilities es cerrado y el lenguaje QR/kiosco/folio se traduce, no se copia); sumados evals `dashboard-plus-portal` y `dashboard-plus-mobile` más la regresión `admin+surface sugiere multi-app`.

- Habilitada la Surface `desktop` con proveedor curado: `tauri-ui` declara `provides_surfaces` respaldado por la evidencia del piloto en el pin `8eb86d89` (desktop-shell, installer verificado a nivel de configuración, native-integration en alcance estrecho y offline arquitectural), GP-07 y GP-06 la admiten en `composable_surfaces`, y se retira `local-files` de su `use_when` porque la plantilla no incluye plugin fs.

- Habilitada la composición real de Surfaces: `public-intake` entra al vocabulario de capabilities de Surface, GP-06 declara `composable_surfaces` (`public-intake` y `mobile`), `tanstack-transactional-pwa` e Ignite registran `provides_surfaces` limitado a capabilities verificadas en su evidencia de curación, el adapter declara `apps/intake` como destino de Surface adicional y la verificación de cobertura de capabilities del provider se aplica a toda Surface solicitada, no solo `public-web`.

- Registrado `tanstack-transactional-pwa` como `specialized`/`curated` tras un piloto completo en el pin `f2571ea8` (install, Biome, tipos, pruebas y build en verde con Bun 1.4.0), con adapter `git-copy` de modo directo, evidencia AI-friendly y ficha en `catalog/profiles/`; cubre captura pública offline con outbox idempotente, adjuntos Blob, kiosco y token de seguimiento, brechas que ningún starter del catálogo cubría.

## [0.9.0] - 2026-09-04

- Añadidas Composable Project Surfaces sin multiplicar Golden Paths.
- Añadidas `public-web` y `mobile` como primeras Surfaces; Stardrive e Ignite son sus providers curados.
- Permitida la composición adicional en SaaS, Admin y Mobile según una matriz explícita de cada Recipe.
- Añadidos capabilities por Surface, roles de starter y composición semántica en el manifest.
- Añadido `eng surface add` con dry-run, aplicación transaccional e idempotencia.
- Actualizados Architecture, Gentle, CI, discovery y evolución para respetar providers, destinos y requisitos pendientes.
- Conservada la inspección de proyectos 0.8 mediante inferencia de provider sin inventar capabilities.
- Añadidos evals, regresiones y escenarios multi-Surface para SaaS, Admin y Mobile.

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
