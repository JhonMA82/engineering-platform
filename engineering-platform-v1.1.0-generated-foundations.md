# Engineering Platform v1.1.0 — Generated Foundations
## Instrucciones de implementación para Pi Agent

**Base:** Engineering Platform `v1.0.1`  
**Tipo de release recomendado:** `v1.1.0` (minor, porque amplía el contrato arquitectónico de materialización)  
**Objetivo único:** permitir que Engineering Platform consuma boilerplates/foundations que generan un proyecto mediante comandos, perfiles y argumentos, sin introducir lógica específica de cada proveedor en el core.

---

# 0. Decisión principal

Engineering Platform 1.0/1.0.1 ya sabe resolver arquitectura, seleccionar providers, componer Surfaces, crear un `MaterializationPlan`, trabajar en staging y hacer commit atómico del proyecto.

El nuevo problema es distinto:

```text
No todas las foundations son repositorios que deban copiarse.
Algunas son factories/generators que deben ejecutarse para producir
la variante mínima adecuada del proyecto.
```

Ejemplos reales:

```text
api-starter
→ bun run create:project -- --profile=authenticated --out=<temp>

shadcn-next-boilerplate
→ npm run generate:project -- inventory-admin --profile minimal

futuro tanstack admin fork
→ npm run generate:project -- inventory-admin --profile minimal

Ignite
→ ignite new ... con argumentos curados
```

Engineering Platform debe soportar ambos tipos de foundation:

```text
Foundation
├── Static
│   └── fetch → copy/prune
│
└── Generated
    └── fetch/acquire generator
        → prepare generator
        → execute exact curated command + arguments
        → validate generated output
        → copy generated output into project staging
```

La factory/generator **NO forma parte del proyecto final**.

---

# 1. No convertir esto en v1.0.2

Esta capacidad cambia el contrato de materialización.

No es un bugfix pequeño.

Publicar como:

```text
v1.1.0 — Generated Foundations
```

No modificar el significado de `v1.0.1`.

---

# 2. Objetivos

La v1.1.0 debe:

1. conservar compatibilidad con foundations estáticas existentes;
2. añadir foundations generadas;
3. soportar comandos como arrays de argv, nunca shell strings;
4. soportar perfiles/argumentos curados;
5. resolver la configuración antes de ejecutar;
6. registrar la configuración en `MaterializationPlan`;
7. incluir esa configuración en el fingerprint;
8. ejecutar generators en un workspace temporal aislado;
9. llevar únicamente el output generado al staging del proyecto;
10. ejecutar validaciones sobre el output;
11. registrar provenance suficiente para reproducir la generación;
12. conservar `Resolver`, `Composer` y `Materializer` con responsabilidades claras;
13. evitar cualquier `if provider == "api-starter"` o equivalente;
14. permitir que futuros generators se añadan principalmente mediante catálogo.

---

# 3. No objetivos

No implementar en v1.1:

- Better-T-Stack como provider;
- deployment strategy;
- deployment automation;
- selección de Clerk/Resend/Trigger.dev;
- skills de Gentle;
- caché avanzada de generators;
- ejecución remota;
- plugin runtime;
- un DSL genérico complejo;
- un stack builder universal;
- actualización automática de boilerplates;
- generación de features durante desarrollo (`generate:crud`, `generate:feature`, etc.);
- configuración interactiva del generator.

Esos temas se evaluarán después.

---

# 4. Principio rector

La regla arquitectónica debe ser:

> Engineering Platform decide qué foundation usar y con qué configuración curada.  
> El Materializer solamente ejecuta la decisión.

Nunca:

```text
Materializer:
"veo que es api-starter, voy a elegir authenticated"
```

Sí:

```text
MaterializationPlan:
foundation = hono-api
strategy = generate
profile = authenticated
arguments = [...]
```

y después:

```text
Materializer:
execute exactly this validated generation plan
```

---

# 5. Mantener intactas las responsabilidades actuales

## Resolver

Continúa respondiendo:

```text
¿Qué arquitectura/family/recipe corresponde?
```

No debe saber:

```text
--profile minimal
--removeDemo
npm run generate:project
```

## Composer

Continúa respondiendo:

```text
¿Qué provider satisface cada Surface?
¿Dónde irá cada Surface?
¿Qué database profile corresponde?
```

No debe ejecutar generators.

Idealmente tampoco debe conocer sintaxis de argumentos.

## Foundation Configuration / Planner

La configuración materializable debe resolverse en una fase **pura y determinista**, después de seleccionar el provider y antes de ejecutar side effects.

Se puede implementar de dos formas aceptables.

### Opción recomendada

Introducir una pequeña capa pura:

```text
internal/foundationconfig/
```

Flujo:

```text
ProjectIntent
    ↓
Resolver
    ↓
ArchitectureDecision
    ↓
Composer
    ↓
Foundation Configurator
    ↓
Planner
    ↓
MaterializationPlan
```

`Foundation Configurator` recibe:

- `ProjectIntent`;
- `ArchitectureDecision`;
- `Composition`;
- catálogo.

Y produce para cada componente:

```text
MaterializationConfig
```

### Opción mínima aceptable

Mantenerlo dentro de `planner` mediante funciones puras de configuración.

No meter esta selección dentro de `materializer`.

La elección entre estas dos opciones debe favorecer menor complejidad y mejor separación según el código v1.0.1 real.

---

# 6. Dos estrategias de materialización

Normalizar a dos estrategias públicas:

```text
copy
generate
```

## `copy`

Comportamiento equivalente a v1.0.1:

```text
fetch pinned source
→ copy
→ prune
→ setup
→ checks
```

## `generate`

Nuevo comportamiento:

```text
acquire pinned generator/foundation
→ prepare generator if required
→ execute curated generator command
→ locate declared output
→ validate output
→ copy output into project staging destination
→ optional setup on generated output
→ checks
```

No introducir nombres como:

```text
api-starter
ignite-mode
next-admin-mode
tanstack-mode
```

El core solo entiende:

```text
copy
generate
```

---

# 7. Compatibilidad hacia atrás

Las entradas actuales que no declaren estrategia deben seguir significando:

```text
strategy = copy
```

No obligar a migrar todo el catálogo de inmediato.

Un adapter antiguo:

```json
{
  "operations": ["fetch", "copy"]
}
```

debe continuar funcionando.

---

# 8. Contrato declarativo de Generated Foundation

Agregar un contrato genérico equivalente a:

```json
{
  "strategy": "generate",
  "generator": {
    "prepare": [],
    "run": [],
    "output": "",
    "default_profile": "",
    "profiles": []
  }
}
```

Los nombres exactos pueden adaptarse a las convenciones Go/JSON actuales.

La semántica sí debe conservarse.

---

# 9. `prepare`

Algunas factories necesitan preparar dependencias antes de generar.

Ejemplo conceptual para un repo Bun:

```json
"prepare": [
  {
    "run": ["bun", "install", "--frozen-lockfile"]
  }
]
```

Ejemplo npm:

```json
"prepare": [
  {
    "run": ["npm", "ci"]
  }
]
```

`prepare` corre en el workspace temporal de la **factory**, no en el destino final.

No confundir con `setup` del proyecto generado.

---

# 10. `run`

El generator se describe como argv.

Ejemplo conceptual:

```json
"run": [
  "npm",
  "run",
  "generate:project",
  "--",
  "{name}",
  "--profile",
  "{profile}",
  "--destination",
  "{output}"
]
```

o:

```json
"run": [
  "bun",
  "run",
  "create:project",
  "--",
  "--profile={profile}",
  "--out={output}"
]
```

Reglas:

- no shell string;
- no `bash -c`;
- no `sh -c`;
- no interpolación arbitraria;
- placeholders conocidos únicamente;
- cada token se pasa como argumento separado;
- validar metacaracteres con la política existente;
- timeout obligatorio.

---

# 11. Placeholders permitidos

Empezar con un vocabulario pequeño.

Recomendados:

```text
{name}
{project}
{surface}
{profile}
{output}
```

Semántica:

```text
{name}
→ nombre lógico de la Surface generada
  ejemplo: inventory-admin

{project}
→ nombre normalizado del proyecto
  ejemplo: inventory

{surface}
→ id de Surface
  ejemplo: admin

{profile}
→ perfil de generación ya resuelto

{output}
→ ubicación temporal controlada por Engineering Platform
```

No permitir placeholders desconocidos.

---

# 12. Nombre lógico vs destino

No hacer que el generator decida topología del monorepo.

Ejemplo:

```text
Project:
inventory
```

Engineering puede generar:

```text
{name} = inventory-admin
```

El generator produce:

```text
inventory-admin/
```

pero Composer/Planner decide que el resultado termina en:

```text
apps/admin/
```

Por tanto:

```text
generator name
≠
final destination
```

Esto mantiene la factory reutilizable fuera de Engineering Platform.

---

# 13. Estandarización de nombres

Implementar una función determinista de naming.

Sugerencia:

```text
single-surface:
{project}

multi-surface:
{project}-{surface}
```

Ejemplos:

```text
inventory
inventory-admin
inventory-api
inventory-mobile
inventory-web
```

Si una foundation tiene restricciones de naming, declararlas en su adapter/catalog entry.

No codificar reglas específicas en el core por provider.

---

# 14. Perfiles

Un Generated Foundation puede declarar perfiles curados.

Ejemplo:

```json
"profiles": [
  {
    "id": "minimal",
    "arguments": ["--profile", "minimal"]
  },
  {
    "id": "authenticated",
    "arguments": ["--profile", "authenticated"]
  }
]
```

El core no debe saber qué significa:

```text
minimal
authenticated
multi-tenant-core
```

Son conceptos de la foundation.

---

# 15. Selección de perfiles

La selección debe ser determinista y declarativa.

Evitar construir un lenguaje de reglas complejo en v1.1.

El contrato debe poder mapear **señales ya estructuradas de ProjectIntent/ArchitectureDecision** a un perfil.

Señales válidas iniciales:

- required Surfaces;
- ArchitectureRequirement refs;
- architectural Capability IDs;
- `DataIntent` estructurado;
- `OperationalIntent` estructurado;
- technical constraints explícitos;
- ProductRequirement IDs solo cuando sean IDs estables/conocidos y el mapping esté declarado en catálogo.

No interpretar texto libre.

No usar LLM.

No consultar Internet.

---

# 16. Regla “smallest valid profile”

Inspirarse en la filosofía:

> elegir la variante más pequeña que cubra las necesidades conocidas.

Si una factory define:

```text
minimal
authenticated
multi-tenant-core
integration-platform
platform
```

Engineering debe seleccionar la variante mínima compatible con las señales conocidas.

No elegir automáticamente:

```text
platform
```

solo porque incluye todo.

---

# 17. No convertir product features en routing

Muy importante:

```text
PDF
Excel import
reports
QR
emails
```

siguen sin ser razones para descartar una foundation arquitectónicamente correcta.

La configuración del generator puede aprovechar una capacidad ya soportada por la foundation, pero:

```text
missing product feature
≠
CatalogGap
```

Ejemplo:

```text
SaaS multi-tenant + PDF
```

Puede producir:

```text
api-starter profile = multi-tenant-core
```

y dejar:

```text
PDF implementation → Gentle
```

---

# 18. Product requirements sí pueden ayudar a configurar

Esto NO contradice la regla anterior.

Después de elegir correctamente la foundation:

```text
selected foundation
        ↓
configuration optimization
```

puede aprovecharse un requisito explícito si existe una opción curada.

Ejemplo:

```text
required: notifications
foundation already has optional notifications module
```

Entonces la configuración podría activar esa feature.

Pero esto debe ser:

```text
optimization of selected foundation
```

Nunca:

```text
eligibility requirement for foundation
```

---

# 19. No exigir soporte de todas las features

Si la foundation seleccionada no sabe generar una feature:

```text
no failure
```

La feature queda en:

```text
implementation-brief.md
handoff.json
```

para Gentle.

---

# 20. MaterializationPlan debe guardar la decisión

`MaterializationPlan` v1 identifica:

- boilerplate;
- pin;
- destination;
- surface.

Para v1.1 debe poder incluir materialización por componente.

Ejemplo conceptual:

```json
{
  "boilerplate": "hono-api",
  "pin": "abc123",
  "surface": "api",
  "destination": "services/api",
  "materialization": {
    "strategy": "generate",
    "name": "inventory-api",
    "profile": "authenticated",
    "arguments": ["--profile=authenticated"],
    "adapter_fingerprint": "sha256:..."
  }
}
```

Para static:

```json
"materialization": {
  "strategy": "copy",
  "adapter_fingerprint": "sha256:..."
}
```

---

# 21. No guardar rutas temporales en el plan

No serializar:

```text
/tmp/eng-1234/output
C:\Temp\...
```

El plan debe ser reproducible independientemente de la máquina.

`{output}` se resuelve únicamente durante materialización.

---

# 22. Schema version

Recomendación:

```text
MaterializationPlan schema_version = 2
```

porque la semántica por componente se amplía.

Si es razonablemente sencillo:

- seguir leyendo planes v1 como `strategy=copy`;
- generar nuevos planes v2.

No romper proyectos ya materializados ni manifests históricos.

---

# 23. Fingerprint

Actualizar `FingerprintPlan`.

Debe incluir, por cada componente:

```text
surface
boilerplate
pin
destination
strategy
logical name
profile
resolved arguments
adapter fingerprint
```

Objetivo:

Cambiar:

```text
profile=minimal
```

por:

```text
profile=authenticated
```

debe cambiar el fingerprint.

---

# 24. Adapter fingerprint

Agregar un fingerprint determinista del contrato de adapter/generator.

Debe cubrir al menos:

```text
strategy
prepare commands
run command template
output declaration
profile definitions
setup
checks
managed files
```

Esto evita que un plan antiguo ejecute silenciosamente un adapter modificado aunque el provider/pin no haya cambiado.

Si el catálogo ya tiene un mecanismo equivalente, reutilizarlo.

---

# 25. Pipeline de materialización generado

Para cada componente generated:

```text
1. validar plan + adapter
2. obtener pinned source/generator
3. crear sandbox temporal por componente
4. preparar factory
5. resolver placeholders runtime seguros
6. ejecutar exact argv
7. localizar output declarado
8. verificar que output esté dentro del sandbox
9. verificar que output exista
10. verificar que output no esté vacío
11. copiar output a project staging destination
12. aplicar prune post-generation si está declarado
13. ejecutar setup sobre el proyecto generado
14. ejecutar checks
15. continuar con manifest/provenance/handoff
16. commit atómico final
```

---

# 26. Reutilizar staging v1.0.1

No crear un segundo sistema de commit.

Engineering Platform ya usa:

```text
project staging
→ verification
→ atomic rename
```

Mantenerlo.

Añadir solamente un sandbox temporal para generation.

Ejemplo conceptual:

```text
/tmp/eng-fetch-*/
    source/

<project-parent>/.staging-*/
    apps/admin/
    services/api/

temporary generation sandbox:
    factory/
    generated/
```

El resultado generado se copia dentro de `.staging-*`.

El output final sigue tocándose solo en el commit.

---

# 27. La factory nunca debe filtrarse al proyecto final

Agregar una prueba explícita.

Nunca debe terminar algo como:

```text
apps/admin/
├── templates/
├── generator/
├── scripts/create-project...
└── source showcase...
```

salvo que esos archivos formen legítimamente parte del output generado.

El proyecto final recibe:

```text
generated output only
```

---

# 28. Fetch temporal vs caché

Para v1.1:

```text
usar fetch temporal actual
```

No implementar caché persistente todavía.

Futuro posible:

```text
~/.cache/engineering-platform/foundations/<id>/<pin>/
```

Pero es optimización, no requisito funcional.

Primero lograr generación reproducible y segura.

---

# 29. Source pin y generator version

Mantener pins inmutables.

Si el generator vive dentro del repo:

```text
repo commit SHA
```

fija su comportamiento.

Si se invoca una herramienta externa:

```text
npx ignite-cli@11.5.0
```

debe tener versión exacta.

Prohibir en catálogo curado:

```text
@latest
main
master
unversioned global executable
```

salvo excepción explícita documentada y no selectable como stable.

---

# 30. Instalación de dependencias de la factory

Ejemplo `api-starter`:

```text
bun install --frozen-lockfile
```

ocurre en la factory temporal.

Ejemplo `shadcn-next-boilerplate`:

```text
npm ci
```

ocurre en la factory temporal si es necesario para ejecutar scripts TS.

Estas dependencias NO justifican copiar `node_modules` al proyecto generado.

---

# 31. Setup del output

Separar:

```text
generator.prepare
```

de:

```text
generated_output.setup
```

Ejemplo:

```text
prepare:
npm ci

generate:
npm run generate:project ...

setup:
npm install / npm ci
```

No asumir que un proyecto recién generado necesita instalación.

Cada provider debe declararlo.

---

# 32. Seguridad de comandos

Reutilizar y ampliar la política actual.

Obligatorio:

- argv arrays;
- no shell;
- timeout;
- no argumentos vacíos;
- no placeholders desconocidos;
- no path traversal;
- output dentro de sandbox;
- destination dentro de project staging;
- no environment secrets en catálogo;
- no interpolar ProductIntent libre dentro de shell;
- nombre sanitizado antes de insertarlo en argv;
- límite razonable de output/log de procesos.

---

# 33. Interactive generators

Generated Foundations usadas por Engineering Platform deben soportar modo no interactivo.

Si un generator exige preguntas:

```text
no selectable
```

hasta que exista:

- flag `--yes`;
- config file;
- argumentos completos;
- API programática;
- otra forma determinista.

Engineering Platform no debe automatizar TTY/prompt scraping.

---

# 34. Introspection de generators

Si una factory ofrece:

```text
--list-profiles --json
--list-features --json
```

puede usarse durante:

```text
curation
pilot
maintenance
```

No usarlo como dependencia obligatoria del Resolver en cada ejecución.

El catálogo curado debe contener la versión validada del contrato.

Esto mantiene:

```text
planning deterministic + offline-capable
```

cuando se usan fixtures.

---

# 35. Caso 1 — api-starter

Actualmente `hono-api` apunta a:

```text
JhonMA82/api-starter
```

Ese repo es una **factory**.

No debe copiarse como API final.

Migrarlo a:

```text
strategy = generate
```

Flujo esperado:

```text
fetch api-starter@pinned-commit
↓
bun install --frozen-lockfile
↓
bun run create:project -- --profile=<resolved> --out=<temporary-output>
↓
validate generated API
↓
services/api/
```

Perfiles conocidos del starter deben curarse desde la versión pinneada.

Ejemplos:

```text
minimal
data-api
authenticated
multi-tenant-core
integration-platform
platform
```

No asumir que estos nombres son eternos: son metadata del provider/pin.

---

# 36. api-starter — selección inicial sugerida

No intentar cubrir todos los mappings en el primer commit.

Empezar con señales claras.

Ejemplo conceptual:

```text
no persistence / public simple API
→ minimal

persistence + no user accounts
→ data-api

multi-user/auth required
→ authenticated

multi-tenant architecture
→ multi-tenant-core

multi-tenant + architectural jobs/webhooks/API integration requirements
→ integration-platform
```

`platform` no debe ser fallback automático.

Si una combinación no está suficientemente definida:

```text
smallest safe profile
+
remaining product requirements → Gentle
```

---

# 37. Caso 2 — shadcn-next-boilerplate

Cuando se actualice y se convierta en provider oficial `next-admin`, usar:

```text
strategy = generate
profile = minimal
```

Comando base conocido:

```text
npm run generate:project -- inventory-admin --profile minimal
```

Engineering debe producir un nombre de Surface determinista, por ejemplo:

```text
inventory-admin
```

y colocar el output final en:

```text
apps/admin/
```

La factory completa no llega a `apps/admin`.

---

# 38. Development tools de shadcn-next-boilerplate

El repo también puede ofrecer:

```text
generate:feature
generate:crud
generate:dashboard
```

NO ejecutarlos como parte del bootstrap v1.1 salvo que exista una necesidad arquitectónica explícita y curada.

Su uso principal futuro es:

```text
Gentle development tooling
```

Conservar esa información en:

```text
surface AGENTS.md
GENTLE.md / implementation brief
```

si ya viene en el output de la foundation.

No crear todavía un subsistema `development_tools`.

---

# 39. Caso 3 — TanStack Admin fork

El fork actual debe evolucionar siguiendo el mismo contrato humano que Next Admin.

Objetivo del fork:

```text
npm run generate:project -- inventory-admin --profile minimal
```

Recomendado:

```text
full
minimal
```

`minimal` debe conservar:

- admin shell;
- routing foundation;
- sidebar/navigation;
- theme/design system;
- forms/table patterns estructurales;
- error/not-found handling;
- AGENTS;
- validators;
- tooling útil.

Y eliminar:

- dashboards showcase no necesarios;
- datos mock;
- usuarios ficticios;
- demo charts;
- demo calendar/chat/kanban/etc.;
- contenido puramente demostrativo.

Cuando el fork esté validado:

```text
tanstack-admin
→ strategy=generate
→ profile=minimal
```

---

# 40. No exigir interfaces idénticas a todos los generators

Es positivo que tus forks propios converjan en algo como:

```text
generate:project
--profile
--destination / --out
--no-install
--json / list profiles
```

Pero Engineering Platform NO debe exigir que Ignite, Tauri u otros proyectos externos usen exactamente esa CLI.

El adapter traduce el contrato genérico a la CLI real.

---

# 41. Caso 4 — Ignite

Ignite ya aparece conceptualmente como generator.

Usarlo como prueba de que la abstracción no está diseñada solo para tus repos.

La configuración curada debe preferir una salida limpia cuando exista un argumento oficial equivalente a:

```text
remove demo
```

No hardcodear `Ignite` en Go.

El adapter declara sus argv.

---

# 42. Tauri, Stardrive y otros

No migrarlos obligatoriamente durante el primer PR.

Después de que el contrato funcione con:

```text
api-starter
Next Admin
Ignite
```

evaluar:

```text
Tauri UI
Stardrive
FastAPI/Copier
```

Una release v1.1 puede soportar el mecanismo aunque algunos providers sigan `copy`.

---

# 43. Better-T-Stack

No añadirlo en esta release.

Pero usarlo como prueba mental del contrato:

```text
¿podría un futuro adapter expresar Better-T-Stack
sin modificar el core?
```

Si la respuesta es sí, el contrato probablemente es suficientemente general.

No agregar campos específicos de Better-T-Stack.

---

# 44. Plan determinista

`eng plan --json` debe mostrar claramente:

```text
qué se va a generar
con qué foundation
qué pin
qué strategy
qué profile
qué argumentos no temporales
en qué destination terminará
```

Así `--dry-run` continúa siendo útil.

---

# 45. Ejemplo de plan esperado

Ejemplo conceptual:

```json
{
  "schema_version": 2,
  "project": "inventory",
  "components": [
    {
      "surface": "admin",
      "boilerplate": "tanstack-admin",
      "pin": "<sha>",
      "destination": "apps/admin",
      "materialization": {
        "strategy": "generate",
        "name": "inventory-admin",
        "profile": "minimal",
        "arguments": ["--profile", "minimal"]
      }
    },
    {
      "surface": "api",
      "boilerplate": "hono-api",
      "pin": "<sha>",
      "destination": "services/api",
      "materialization": {
        "strategy": "generate",
        "name": "inventory-api",
        "profile": "authenticated",
        "arguments": ["--profile=authenticated"]
      }
    }
  ]
}
```

No guardar `{output}` ya resuelto.

---

# 46. Provenance

Ampliar `.engineering/provenance.json`.

Por componente registrar:

```text
surface
foundation
repo
pin
strategy
logical name
profile
resolved non-secret arguments
adapter fingerprint
```

Opcionalmente:

```text
generator tool/version
```

si puede conocerse determinísticamente.

Nunca registrar:

- temp absolute paths;
- secrets;
- tokens;
- credentials.

---

# 47. Manifest / Doctor

`eng doctor` debe ser capaz de comprobar que el proyecto materializado sigue asociado a:

```text
foundation + pin + generation configuration
```

No necesita volver a ejecutar el generator.

Debe detectar inconsistencias de metadata/provenance.

---

# 48. Handoff a Gentle

Agregar al handoff información útil, no excesiva.

Ejemplo:

```text
admin:
  generated from tanstack-admin
  profile: minimal

api:
  generated from api-starter
  profile: authenticated
```

Esto ayuda a Gentle a entender qué foundation recibió.

No pedirle a Gentle que redescubra por qué se eligió el profile.

---

# 49. Preservar AGENTS del output

Si la foundation generada produce:

```text
AGENTS.md
```

conservarlo.

El root `AGENTS.md` generado por Engineering sigue actuando como router.

No sobrescribir el `AGENTS.md` específico de una Surface salvo que el contrato de managed files lo indique explícitamente.

---

# 50. Curation contract

Toda Generated Foundation selectable debe tener evidencia adicional:

```text
generator command validated
generator version/pin validated
profiles validated
default profile validated
output path validated
non-interactive run validated
generated output builds/tests
factory files do not leak
AGENTS behavior validated
```

Añadir esto a la política de curation.

---

# 51. Estados de delivery

No promover automáticamente un provider a `curated/released` solo porque el schema valida.

Para un provider convertido de `copy` a `generate`:

```text
downgrade to pilot-ready
```

si hace falta hasta completar materialization pilot real.

Luego promover nuevamente.

---

# 52. Tests de dominio/catalog

Agregar tests equivalentes a:

```text
TestAdapterAcceptsCopyStrategy
TestAdapterAcceptsGenerateStrategy
TestGenerateRequiresRunCommand
TestGenerateRequiresSafeOutputDeclaration
TestGenerateRejectsUnknownPlaceholder
TestGenerateRejectsInteractiveOnlyContract
TestStaticAdapterRemainsBackwardCompatible
TestGeneratorProfilesHaveUniqueIDs
TestDefaultGeneratorProfileExists
TestGeneratorCommandsRejectShellExecution
```

---

# 53. Tests de configuración

Agregar:

```text
TestGeneratedFoundationUsesDefaultMinimalProfile
TestProfileSelectionIsDeterministic
TestSmallestCompatibleProfileWins
TestProductFeatureDoesNotDisqualifyFoundation
TestUnknownProductFeatureRemainsForGentle
TestTechnicalConstraintIsRespectedDuringGenerationConfig
```

Si no se implementa selección automática completa en el primer PR, ajustar los nombres al alcance real sin fingir cobertura.

---

# 54. Tests de planner

Agregar:

```text
TestPlanV2ContainsGenerationStrategy
TestPlanContainsResolvedProfile
TestPlanFingerprintChangesWhenProfileChanges
TestPlanFingerprintChangesWhenArgumentsChange
TestPlanDoesNotContainTemporaryPaths
TestCopyComponentStillPlansCorrectly
```

---

# 55. Tests de materializer

Obligatorios:

```text
TestGeneratedFoundationRunsInTemporaryWorkspace
TestGeneratedFoundationCopiesOnlyOutput
TestGeneratorSourceDoesNotLeakIntoProject
TestGeneratedOutputCannotEscapeSandbox
TestMissingGeneratedOutputFails
TestEmptyGeneratedOutputFails
TestGeneratorFailureLeavesDestinationUntouched
TestGeneratorTimeoutLeavesDestinationUntouched
TestGeneratedSetupRunsAgainstOutputNotFactory
TestGeneratedChecksRunAgainstOutput
TestMixedCopyAndGenerateProjectMaterializes
```

---

# 56. Offline fixtures

Crear generators mínimos de fixture.

No depender de npm/Bun/Internet en unit tests.

Ejemplo fixture generator:

```text
testdata/fixtures/generators/fake-admin/
```

que acepte:

```text
name
profile
output
```

y genere archivos simples.

Usarlo para probar:

- placeholders;
- profiles;
- sandbox;
- errors;
- output;
- mixed projects.

---

# 57. Pilots reales

Unit tests deben ser offline.

Pilots pueden usar providers reales.

Pilots mínimos recomendados:

## Pilot A — api-starter minimal

```text
API simple
→ hono-api
→ generate
→ minimal
```

Verificar:

- output real;
- no factory;
- build/check;
- provenance.

## Pilot B — api-starter richer profile

Ejemplo:

```text
multi-user / auth
→ authenticated
```

o un perfil arquitectónico más fuerte según los mappings implementados.

## Pilot C — Next Admin minimal

Cuando el repo esté actualizado/curado:

```text
next-admin
→ generate
→ minimal
```

Verificar que demos retirados no aparezcan.

## Pilot D — mixed project

```text
admin generated
+
api generated
```

Resultado:

```text
apps/admin/
services/api/
```

y un solo root handoff.

## Pilot E — Ignite

Validar que un generator externo también funciona con el contrato.

---

# 58. TanStack Admin no debe bloquear v1.1

Si el fork todavía está en desarrollo:

```text
no esperar a terminarlo para diseñar el core
```

Implementar v1.1 contra:

```text
api-starter
Next Admin
Ignite
```

y registrar TanStack Admin cuando su `generate:project --profile minimal` esté validado.

Esto evita acoplar release del core al calendario del fork.

---

# 59. Migración del catálogo

Migrar providers gradualmente.

Orden recomendado:

```text
1. hono-api → generated
2. ignite → generated contract definitivo
3. next-admin → generated después de actualizar fork
4. tanstack-admin → generated cuando el fork esté listo
5. revisar Tauri/Stardrive después
```

No modificar providers que no necesitan cambio.

---

# 60. `hono-api` actual es el caso que corrige un bug conceptual

El entry actual representa `api-starter` como foundation API.

Después de v1.1, no debe materializar la factory completa.

Definition:

```text
hono-api = generated API foundation
```

La materialización correcta es:

```text
api-starter factory
→ generator
→ independent generated API
```

---

# 61. `next-admin` recomendado

Una vez actualizado:

```text
provider ID: next-admin
repo: JhonMA82/shadcn-next-boilerplate
pin: immutable SHA
strategy: generate
curated profile: minimal
```

Conservar legacy IDs/provenance si la migración lo requiere.

No crear otro provider duplicado solo por cambiar repo/fork.

---

# 62. `tanstack-admin` recomendado

Cuando el fork actual tenga tooling equivalente:

```text
provider ID: tanstack-admin
repo: JhonMA82/tanstack-shadcn-admin-dashboard
pin: immutable SHA
strategy: generate
curated profile: minimal
```

No cambiar el ID lógico si sigue representando la misma foundation.

---

# 63. No imponer estructura T3 a las factories

Engineering controla la topología final.

Ejemplo:

```text
apps/admin
apps/mobile
apps/web
services/api
packages/*
```

Una factory genera una Surface independiente.

No exigir que `api-starter` genere directamente en `services/api`.

No exigir que Next Admin conozca `apps/admin`.

---

# 64. Topología futura

La discusión sobre unificar:

```text
apps/api
```

vs:

```text
services/api
```

queda fuera de v1.1.

No introducirla como efecto secundario de Generated Foundations.

---

# 65. Commands after bootstrap

Distinguir dos categorías.

## Materialization generators

Usados por Engineering:

```text
generate:project
create:project
ignite new
```

## Development generators

Usados después por Gentle/desarrollador:

```text
generate:feature
generate:crud
generate:dashboard
add:feature
```

v1.1 implementa formalmente la primera categoría.

La segunda se documenta/preserva pero no necesita un nuevo subsystem todavía.

---

# 66. CLI de Engineering

No es necesario crear muchos comandos nuevos.

`eng plan`, `eng start`, `eng materialize` deben funcionar con generated foundations de manera transparente.

`eng plan --json` muestra la configuración.

`eng start --dry-run` muestra lo que se generaría.

Opcional:

```text
eng catalog show <provider>
```

puede mostrar:

```text
materialization: generate
profiles:
...
```

si encaja naturalmente.

---

# 67. Errores tipados

Agregar errores claros.

Ejemplos:

```text
generator profile unavailable
generator configuration unresolved
generator command invalid
generator preparation failed
generator execution failed
generator output missing
generator output escaped sandbox
generator output validation failed
adapter drift
```

No devolver únicamente:

```text
exit status 1
```

Incluir contexto:

```text
surface
provider
stage
```

sin secretos.

---

# 68. CatalogGap vs generator failure

Distinguir:

```text
No existe foundation compatible
→ CatalogGap
```

de:

```text
Foundation seleccionada pero generator falló
→ Materialization failure
```

y:

```text
Foundation válida, pero ninguna configuración curada puede satisfacer
una hard architectural requirement
→ composition/configuration gap
```

No etiquetar todo como CatalogGap.

---

# 69. Hard constraints

Si una configuración del generator violaría una hard constraint:

```text
NO fallback
```

Ejemplo:

```text
must-not-use database=postgresql
```

y un profile generado fuerza PostgreSQL:

```text
ese profile es inválido
```

Si no existe alternativa:

```text
fail/gap explícito
```

No generar de todos modos.

---

# 70. Reproducibilidad

Para afirmar que una generated foundation es reproducible, registrar:

```text
provider repo
provider pin
adapter contract
generator command
generator profile
generator arguments
catalog version
plan fingerprint
```

No es necesario registrar el contenido completo de la factory.

---

# 71. README v1.1

Actualizar README sin volverlo técnico en exceso.

Añadir dentro de “How it works” una explicación sencilla:

> Some foundations are copied from a curated template; others are generated from a pinned factory using a validated profile. Engineering Platform runs the factory in a temporary workspace and keeps only the generated project.

Agregar ejemplo:

```text
Next Admin → minimal
API Starter → authenticated
```

No añadir todo el schema al README.

La documentación profunda va a:

```text
docs/architecture/generated-foundations.md
```

---

# 72. Documentación nueva

Crear:

```text
docs/architecture/generated-foundations.md
```

Contenido:

- static vs generated;
- lifecycle;
- adapter contract;
- profile selection;
- staging;
- security;
- provenance;
- curation;
- examples.

Actualizar:

```text
docs/architecture/materialization.md
docs/guides/add-boilerplate.md
docs/guides/new-project.md
```

según nombres reales existentes.

---

# 73. Guía de curation para forks propios

Añadir checklist:

```text
Does generator run non-interactively?
Does it accept output destination?
Does it accept stable profile IDs?
Does minimal remove showcase/demo code?
Does generated project retain AGENTS?
Can generated project validate/build independently?
Does it avoid runtime dependency on factory?
Can profile/options be listed machine-readably?
```

Las dos últimas son recomendadas; no hacerlas hard requirements si no son necesarias.

---

# 74. Recomendación para tus forks

Intentar converger en una UX similar:

```text
npm run generate:project -- <name> --profile minimal
```

Opcionales deseables:

```text
--destination <path>
--no-install
--json
--list-profiles
--list-features
```

Pero Engineering Platform no debe depender de que todos usen exactamente estos nombres.

---

# 75. Orden de implementación interno

## Fase 1 — Contracts

- strategy static/generated;
- generator spec;
- profiles;
- placeholders;
- validation;
- schema v2;
- fingerprints.

No ejecutar nada todavía.

## Fase 2 — Planner/configurator

- resolve name;
- resolve strategy;
- resolve profile;
- resolve non-runtime arguments;
- serialize plan v2.

## Fase 3 — Materializer

- generator sandbox;
- prepare;
- execute;
- output validation;
- copy output to existing staging;
- setup/checks.

## Fase 4 — Provenance/Doctor/Handoff

- record config;
- doctor validation;
- handoff summary.

## Fase 5 — Fixtures/tests

- complete offline coverage.

## Fase 6 — Real providers

- api-starter;
- Ignite;
- Next Admin when ready.

## Fase 7 — Pilots/CI/docs

- pilots;
- README/docs;
- release.

---

# 76. Evitar un error importante

No diseñar el contrato directamente como:

```json
{
  "npm_script": "generate:project",
  "profile": "minimal"
}
```

Eso acopla Engineering a npm.

Debe poder representar:

```text
npm
bun
npx
binary
cargo generate
copier
future Better-T-Stack
```

con el mismo modelo de argv.

---

# 77. Evitar otro error importante

No permitir:

```text
catalog adapter contains arbitrary code
```

El catálogo sigue siendo declarativo.

Si un provider necesita lógica compleja para generar:

- esa lógica pertenece a su factory;
- Engineering solo le pasa argumentos curados.

---

# 78. No crear un plugin runtime

No necesitamos:

```text
Go plugins
JS plugins
WASM adapters
dynamic provider code
```

en v1.1.

El contrato declarativo + CLI factory es suficiente para los casos actuales.

---

# 79. Acceptance Criteria funcionales

- [ ] Foundations existentes `copy` siguen funcionando.
- [ ] Existe estrategia `generate`.
- [ ] Generator commands son declarativos/argv.
- [ ] `MaterializationPlan` fija profile/configuration.
- [ ] Plan fingerprint incluye configuración generada.
- [ ] Materializer no selecciona perfiles.
- [ ] Generator corre en temp/sandbox.
- [ ] Factory no aparece en output final.
- [ ] Solo output declarado llega al project staging.
- [ ] Fallo del generator deja destino final intacto.
- [ ] Paths generados no pueden escapar sandbox.
- [ ] Profiles desconocidos fallan antes de ejecución.
- [ ] Placeholder desconocido falla.
- [ ] No se permiten shell commands.
- [ ] Provenance registra strategy/profile/args.
- [ ] `doctor` entiende generated foundations.
- [ ] Handoff informa qué profile produjo cada Surface.

---

# 80. Acceptance Criteria de arquitectura

- [ ] No existe `if hono-api`.
- [ ] No existe `if ignite`.
- [ ] No existe `if next-admin`.
- [ ] No existe `if tanstack-admin`.
- [ ] Resolver continúa sin side effects.
- [ ] Composer continúa sin side effects.
- [ ] Configuración es pura/determinista.
- [ ] Materializer sigue siendo único boundary de procesos/filesystem.
- [ ] Product features no se convierten en routing hard constraints.
- [ ] Nuevo provider generated puede añadirse principalmente por catálogo.
- [ ] Better-T-Stack podría expresarse en el futuro sin cambiar primitives del core.

---

# 81. Acceptance Criteria de providers

## api-starter

- [ ] ya no se copia la factory como API final;
- [ ] usa generator;
- [ ] pin inmutable;
- [ ] al menos `minimal` pilot;
- [ ] al menos un segundo profile pilot;
- [ ] output independiente;
- [ ] provenance correcto.

## Ignite

- [ ] usa contrato generated;
- [ ] generator version pinneada;
- [ ] configuración no interactiva;
- [ ] output validado.

## Next Admin

Cuando esté actualizado:

- [ ] repo AI-friendly correcto;
- [ ] profile `minimal`;
- [ ] demo/showcase removido;
- [ ] `apps/admin` contiene solo output derivado;
- [ ] AGENTS preservado;
- [ ] validate/build pasa.

## TanStack Admin

No bloqueante para primer merge:

- [ ] registrar después de que fork tenga generator estable;
- [ ] profile minimal;
- [ ] mismo contrato generic.

---

# 82. CI

Mantener gates existentes y añadir:

```text
generated foundation unit tests
mixed static/generated tests
offline fixture materialization
plan v1 compatibility test if supported
plan v2 golden tests
```

Pilots reales pueden estar en workflow separado si requieren Internet/npm/Bun.

---

# 83. Release gate

Antes de tag `v1.1.0`:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
./eng catalog validate
```

Además:

```text
all existing pilots green
generated foundation pilots green
README/docs current
release build matrix green
```

---

# 84. Definition of Done

`v1.1.0` está terminada cuando este escenario funciona:

```text
User:
"Necesito un sistema de inventario
con dashboard administrativo y API con usuarios."

Pi
↓
ProjectIntent

Engineering Resolver
↓
arquitectura

Composer
↓
admin provider
api provider

Foundation Configuration
↓
admin:
  generated
  minimal

api:
  generated
  authenticated

Planner
↓
MaterializationPlan v2

Materializer
↓
temporary admin factory
→ generated admin output
→ apps/admin

temporary api factory
→ generated API output
→ services/api

Engineering
↓
AGENTS
ARCHITECTURE
GENTLE
implementation brief
handoff
provenance

Gentle
↓
continúa desarrollo del producto
```

Y en el repo final NO existen las factories completas.

---

# 85. Commit strategy sugerida

```text
feat(domain): model generated foundation contracts
feat(planner): plan deterministic generation configuration
test(planner): fingerprint generated profiles and arguments

feat(materializer): execute generated foundations in isolated staging
test(materializer): cover generated output safety and atomic failures

feat(project): record generated foundation provenance
feat(handoff): expose generated foundation profile context

refactor(catalog): materialize hono-api through its generator
refactor(catalog): normalize ignite as generated foundation

test(pilots): add generated api and mixed-project pilots

docs: document generated foundations and curation workflow
```

No es obligatorio usar exactamente estos commits.

---

# 86. Prompt operativo para Pi Agent

```text
Upgrade Engineering Platform v1.0.1 to v1.1.0 with one architectural feature:
Generated Foundations.

Do not redesign routing, recipes, ProjectIntent, Pi discovery, Gentle handoff,
deployment, skills, or service selection.

Problem:
Engineering Platform currently works primarily as fetch/copy materialization,
but several curated foundations are factories/generators. In particular,
JhonMA82/api-starter must generate an independent API project instead of
copying the factory repository. JhonMA82/shadcn-next-boilerplate and the
forthcoming TanStack Admin fork also expose generate:project with a minimal
profile. Ignite is another independent generator case.

Required architecture:

Foundation
- copy
- generate

For generated foundations:
- acquire the pinned generator/factory into temporary space
- prepare it if needed
- execute a catalog-declared argv command with a deterministically resolved
  configuration/profile
- write output only inside a controlled generation sandbox
- validate the declared output
- copy only that output into the existing project staging destination
- run output setup/checks
- preserve the existing atomic project commit behavior
- never copy the factory itself into the final project

The Materializer must NOT choose profiles or architecture. It executes a
pre-resolved MaterializationPlan.

Add a pure deterministic configuration step before materialization, either
as a small foundationconfig package or as a clearly isolated planner concern.
Do not make Composer or Materializer provider-specific.

MaterializationPlan should evolve to schema v2 and record per-component:
- strategy
- logical generated name
- resolved profile
- resolved non-runtime arguments
- adapter fingerprint

Temporary paths must never be serialized into the plan.

Update plan fingerprints so profile/argument/adapter changes alter the
fingerprint.

Keep old static providers backward compatible. Prefer reading schema v1 plans
as copy semantics if practical.

Generator adapters remain declarative:
- argv arrays only
- no shell strings
- known placeholders only: name, project, surface, profile, output
- timeout
- safe output declaration
- immutable provider/tool pins
- non-interactive execution

Do not add a plugin runtime or provider-specific Go branches.

Profile selection must follow the smallest-valid-profile principle and use
only structured deterministic signals already present in ProjectIntent /
ArchitectureDecision / catalog metadata. Do not inspect narrative text and
do not use an LLM. Product requirements may optimize configuration after a
foundation is selected, but missing product features must never make an
otherwise valid foundation ineligible.

Migrate and pilot at least:
1. hono-api / JhonMA82/api-starter as a generated foundation
2. Ignite through the generic generated contract
3. Next Admin through JhonMA82/shadcn-next-boilerplate when its current
   AI-friendly fork is synced and curated

The TanStack Admin fork must not block the core release; register it after its
generate:project --profile minimal contract is stable.

Use offline fake generators for unit tests. Use real network/tooling only in
pilots.

Extend provenance, doctor, and handoff so a generated Surface records:
foundation, pin, strategy, logical name, profile, arguments, adapter
fingerprint.

Add docs/architecture/generated-foundations.md and update materialization,
new-project, add-boilerplate, and README documentation.

Critical invariant:
a generated project contains only the generated output, never the complete
factory repository.

Before release:
- gofmt
- go vet ./...
- go test ./...
- go build ./cmd/eng
- eng catalog validate
- all existing pilots
- generated-foundation pilots
- release matrix

Return a final readiness report showing:
- architectural changes
- catalog migrations
- tests added
- pilots
- backward compatibility
- remaining follow-ups
```

---

# 87. Recomendación posterior a v1.1.0

Después de publicar y probar v1.1.0:

```text
v1.1.x
→ corregir problemas descubiertos en proyectos reales

siguiente fase
→ actualizar/curar Next Admin + TanStack Admin minimal generators

después
→ proyectos reales Pi → Engineering → Gentle

solo después
→ Agent Preparation / skills / explicit integrations

mucho después
→ Deployment Strategy
```

No mezclar estas fases en v1.1.0.

---

# 88. Resumen final

Engineering Platform no debe convertirse en otro generador universal.

Su papel es:

```text
entender la necesidad
↓
resolver la arquitectura
↓
seleccionar una foundation curada
↓
decidir su configuración mínima válida
↓
ejecutar la factory de forma reproducible
↓
conservar solo el producto generado
↓
componer las Surfaces
↓
entregar contexto limpio a Gentle
```

La factory sabe **cómo generar**.

Engineering Platform sabe **cuándo usarla y con qué configuración**.

Gentle sabe **cómo continuar construyendo el producto**.

Esa frontera debe permanecer intacta.
