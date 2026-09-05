# Escenarios y combinaciones del flujo

Qué resultado produce cada petición al inicio del flujo y qué caminos tiene la iteración principal. Todos los estados y errores de este documento fueron verificados contra `eng` 0.9.0; las combinaciones rechazadas muestran el mensaje real.

## Puertas de entrada

La petición del usuario entra por una de tres puertas; la puerta decide el comando, no al revés.

| Situación | Puerta | Comando |
| --- | --- | --- |
| Proyecto nuevo, conversación | Pi | `eng start <nombre>` → `/new-project` |
| Proyecto nuevo, automatización o CI | CLI | `eng new --from intake.json` (blueprint) |
| Proyecto nuevo con código | CLI | `eng bootstrap --from definicion.json` |
| Cambio en proyecto existente | Pi o CLI | `/evolve-project`, `eng plan`, `eng add`, `eng extend`, `eng surface add` |

Nunca se vuelve a ejecutar `bootstrap` sobre un proyecto existente; la evolución usa los comandos de la sección [Evolución](#evolución-de-proyecto-existente).

## Inicio: petición → resolución

### Combinación Recipe por tipo de proyecto

`project_type` selecciona la Recipe; la Recipe fija starters, base de datos, features y gates.

| `project_type` | Recipe | Estado | Starters | Alternativas |
| --- | --- | --- | --- | --- |
| `public-web` | GP-01 | stable | stardrive | — |
| `admin`, `institutional-admin` | GP-02 | stable | hono-api, tanstack-admin | next-admin, speedpy |
| `python-data` | GP-03 | stable | speedpy | fastapi |
| `mobile` | GP-04 | stable | hono-api, ignite | — |
| `desktop` | GP-05 | stable | tauri-ui | — |
| `multi-app` | GP-06 | stable | hono-api, tanstack-admin | fastapi |
| `commercial-saas` | GP-07 | trial | react-starter-kit | — |

### Combinaciones dentro de la resolución

- **Features**: cada feature queda `materialized` (owner `boilerplate`) o `pending-implementation` (owner `gentle-ai`). Un starter que no materializa una feature no la promete; queda como capacidad pendiente con advertencia.
- **Base de datos**: el perfil debe estar entre los `allowed` de la Recipe; cada uno define canales y gates propios.
- **Surfaces** (opcional, desde 0.9.0): el intake o la definición confirmada puede declarar `surfaces` con `id` y `capabilities`. Ver matriz abajo.
- **Inferencia desde señales**: si el intake trae señales de dominio (`kiosk-mode`, `tracking-token`, `form-capture`...) sin `surfaces` declaradas, el motor implica la surface con esas capabilities y avisa en `warnings`; una señal ambigua entre surfaces (`offline`) no implica nada. La inferencia también filtra la Recipe: `admin` + señales de kiosco sugiere `multi-app` en vez de construir solo el dashboard.
- **Exclusiones**: lo excluido no se materializa ni se promete; activarlo después exige nuevo requisito y ADR cuando cambia arquitectura.

### Resultados posibles del inicio

| Resultado | Condición |
| --- | --- |
| Blueprint (sin código) | `eng new`; útil para CI y para decidir antes de generar |
| Proyecto `code-ready` | Materializado pero checks pendientes; exige `eng check --run` |
| Proyecto `verified` | Materializado y checks aprobados; el setup aprobado se reutiliza |
| `ERROR: Surface desconocida: X` | La Surface no está entre las reconocidas (`public-web`, `public-intake`, `mobile`, `desktop`) |
| `ERROR: <Recipe> no permite componer las Surfaces [...]` | El `project_type` no compone las Surfaces pedidas; reintentar con `project_type: multi-app` (`eng recommend --suggest` devuelve el intake corregido) |
| `ERROR: <provider> no cubre capabilities de <Surface>: ...` | Las capabilities pedidas exceden las declaradas por el provider |
| `ERROR: Capabilities no reconocidas ...; sugerencias: ...` | La capability no existe; el motor sugiere la traducción canónica (`qr-capture` → `form-capture, tracking-token`) |
| `ERROR: No existe provider materializable para la Surface X` | Ningún boilerplate curado declara esa Surface |

### Matriz de Surfaces — estado verificado en 0.9.0

Surfaces reconocidas por `eng`:

| Surface | Capabilities reconocidas |
| --- | --- |
| `public-web` | landing, pricing, blog, documentation, changelog, seo, public-content |
| `mobile` | authenticated-app, push-notifications, camera, offline, location, deep-links, media, background-tasks |
| `public-intake` | form-capture, offline-drafts, offline-outbox, attachments-offline, tracking-token, kiosk-mode, pwa-installable, connectivity-status |
| `desktop` | desktop-shell, installer, local-files, native-integration, offline, auto-update, system-tray, deep-links |

Estado real de composición por Recipe (ejecuciones verificado con `--dry-run`):

| Recipe | Surface | Resultado |
| --- | --- | --- |
| GP-06 multi-app | `public-intake` | Válida: tanstack-transactional-pwa entra como starter `role: surface` en `apps/intake` |
| GP-06 multi-app | `mobile` | Válida: Ignite entra como starter `role: surface` en `apps/mobile` |
| GP-06 multi-app | `desktop` | Válida: tauri-ui entra como starter `role: surface` en `apps/desktop` |
| GP-06 multi-app | `public-web` | Rechazada: GP-06 no la declara |
| GP-07 commercial-saas | `desktop` | Válida: tauri-ui como surface en `apps/desktop` |
| GP-07 commercial-saas | `public-intake` o `mobile` | Rechazada: GP-07 solo declara `desktop` |
| GP-02 admin | cualquier Surface | Rechazada con sugerencia: reintentar con `project_type: multi-app` |

Discrepancias conocidas registradas (no son escenarios válidos hoy):

- Stardrive no declara `provides_surfaces` para `public-web`; hoy no existe provider materializable para esa Surface.
- La matriz del workflow `release-pilots` anuncia `saas-public-web`, `saas-mobile`, `saas-complete`, `admin-complete` y `mobile-public-web`; `scripts/release_pilot.py` aún solo conoce GP-01..GP-06, por lo que esas ejecuciones fallan.

Para capabilities, la regla es una: lo pedido debe estar declarado por el provider (Ignite declara `authenticated-app`). Una capability aceptada es un requisito, nunca evidencia de implementación.

## Evolución de proyecto existente

| Petición | Comando | Efecto |
| --- | --- | --- |
| Nueva capacidad de producto | `eng add <feature>` | Registra decisión y trabajo pendiente; queda `pending-implementation` |
| Otro starter en ruta libre | `eng extend <id>` | Dry-run por defecto; `--apply` integra, valida dependencias y destinos |
| Nueva Surface | `eng surface add <id>` | Dry-run por defecto; `--apply` aplica transaccionalmente e idempotente |
| Cambio de código | `eng plan --change-type <tipo>` | Selecciona skills y gates; la implementación es de Gentle |

El proyecto debe pasar `eng doctor` antes de cualquier `--apply`; `surface add` exige además un proyecto materializado. `eng doctor` suma `evolution_hints` con las Surfaces componibles aún no instaladas.

## Iteración principal

Gentle recibe el handoff y ejecuta el ciclo; las combinaciones posibles son acotadas.

**Estrategia**: ejecución directa o SDD, según riesgo, ambigüedad, contratos, datos y permisos; el motivo se registra brevemente. No hay una tercera vía.

**Alcance del incremento**:

- Solo se implementan las capacidades `pending-implementation` que pertenecen al incremento pedido.
- Una capability de Surface solicitada no está terminada hasta que código y checks lo demuestren.
- Lo `materialized` no se reimplementa.

**Gates**: los definidos en `.engineering/project.json`; con `readiness: code-ready` se exige `eng check --run` antes de tratar el proyecto como verificado.

**Reglas de composición durante la iteración**:

1. No mover ni sustituir providers ni rutas de Surface.
2. No unificar runtimes entre starters.
3. No duplicar auth ni backend.
4. No compartir secretos con `public-web` sin cambio arquitectónico explícito.

**Salidas del ciclo**: incremento verificado (gates en verde y revisión humana) o incidente; los incidentes repetibles y la tercera repetición de una solución se evalúan para knowledge entry, canonical example, feature pack o skill.

## Resumen petición → resultado

| Petición | Resultado esperado |
| --- | --- |
| "Necesito un sitio público" | GP-01, blueprint o proyecto con Stardrive |
| "Sistema interno de solicitudes" | GP-02, hono-api + tanstack-admin, pending: files/notifications si aplican |
| "Admin en ecosistema Next.js" | GP-02 con `next-admin` (auto-selección por señales `next-ecosystem`) |
| "SaaS comercial" | GP-07 (trial), React Starter Kit; compone `desktop` |
| "App móvil para mi admin existente" | GP-06 + `eng surface add mobile` o `eng extend ignite`; capability `authenticated-app` |
| "Portal público de quejas con QR y kiosco" | GP-06 (`multi-app`) + surface `public-intake`; QR/folio/cola offline se traducen a `form-capture, tracking-token, offline-outbox, kiosk-mode` |
| "Dashboard ahora, app móvil en unos meses" | GP-06 desde el inicio (Recipe componible) sin surfaces iniciales; `eng surface add` cuando se confirme |
| "Exportar XLSX" | `eng plan` + `eng add`; librería solo en el módulo correspondiente |
| "Multiplico clientes (20 escuelas)" | No regenerar: intake → ADR → migración de datos → isolation suite |

## Referencias

- [Flujo de proyecto nuevo](new-project.md) — comandos paso a paso.
- [Cambio en proyecto existente](change-existing-project.md) — ejemplos A, B y C.
- [Modelo de ejecución del harness](../06-ai-harness/harness-execution-model.md) — el ciclo de Gentle en detalle.
- [Ejemplo end-to-end](../13-examples/END_TO_END_SCHOOL_REQUESTS.md) — GP-02 de la idea al handoff.
