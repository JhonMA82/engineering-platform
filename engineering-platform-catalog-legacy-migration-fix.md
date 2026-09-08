# Engineering Platform 1.0 — Corrección de migración del catálogo legacy
## Instrucciones operativas para Pi Agent

**Estado:** Ready for implementation  
**Repositorio objetivo:** `JhonMA82/engineering-platform`  
**Fuente histórica:** `JhonMA82/engineering-platform-legacy`  
**Alcance:** corregir la migración de boilerplates sin rediseñar el core Go  
**Objetivo:** restaurar las referencias canónicas del catálogo legacy, migrar entradas omitidas y añadir pruebas que impidan volver a inventar repositorios/pins en futuras migraciones.

---

# 0. Instrucción maestra

Corregir exclusivamente la migración del catálogo desde:

```text
JhonMA82/engineering-platform-legacy
```

hacia:

```text
JhonMA82/engineering-platform
```

La arquitectura Go actual está aprobada.

No:

- reescribir el Resolver;
- cambiar Composer/Planner/Materializer salvo que un test de catálogo revele un bug real;
- crear nuevos boilerplates para sustituir los existentes;
- renombrar repositorios basándose en el ID lógico;
- inventar tags como `v1.0.0`;
- asumir que `id == nombre del repositorio`;
- investigar alternativas nuevas antes de restaurar el catálogo histórico;
- promover automáticamente opciones experimentales o `catalog-only`;
- usar la red como requisito para unit tests.

La fuente histórica canónica para esta tarea es:

```text
engineering-platform-legacy/platform/boilerplates.json
```

El repositorio legacy sirve como **fuente de procedencia**, no como runtime dependency.

---

# 1. Problema confirmado

Durante la reescritura algunos IDs lógicos del catálogo fueron interpretados como si fueran nombres de repositorio.

Ejemplos incorrectos:

```text
id: hono-api
repo: JhonMA82/hono-api
```

cuando históricamente:

```text
id: hono-api
repo: JhonMA82/api-starter
```

También se introdujeron pins `v1.0.0` que no provenían del catálogo legacy.

Además, el catálogo legacy contenía:

```text
11 boilerplates
```

mientras que el catálogo Go actual contiene:

```text
8 boilerplates
```

Por lo tanto la corrección debe incluir:

```text
A. restaurar mappings incorrectos
B. restaurar pins históricos
C. recuperar entradas omitidas
D. actualizar referencias internas
E. añadir pruebas de integridad de migración
F. ejecutar validación/pilots
```

---

# 2. Regla de identidad

Separar siempre estos conceptos:

```text
catalog id
display name
repository URL
upstream pin
adapter id/name
Surface
```

Nunca derivar uno a partir de otro por semejanza textual.

Ejemplo correcto:

```text
catalog id:
hono-api

repository:
https://github.com/JhonMA82/api-starter

Surface:
api

framework:
hono
```

No existe ninguna obligación de que:

```text
hono-api == api-starter
```

como string.

---

# 3. Mappings canónicos a restaurar

## 3.1 Stardrive

### Legacy canónico

```text
id:
stardrive

repository:
https://github.com/peltmonger/stardrive

pin:
5c449810b763140ac72133ff4ae63d8497cce77a

decision_status:
default

delivery_status:
curated

maintenance_tier:
A

category:
public-web
```

### Estado incorrecto a retirar

No conservar como identidad canónica:

```text
id:
stardrive-public-web

repository:
https://github.com/JhonMA82/stardrive-public-web

pin:
v1.0.0
```

### Acción

Restaurar el ID histórico:

```text
stardrive
```

y adaptar la entrada al schema v1 actual.

La Surface:

```text
public-web
```

ya expresa el propósito del boilerplate.

No es necesario duplicar esa semántica dentro del ID.

---

## 3.2 TanStack Admin

### Legacy canónico

```text
id:
tanstack-admin

repository:
https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard

pin:
e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0

legacy_id:
tanstack-shadcn-admin-dashboard

decision_status:
default

delivery_status:
curated

maintenance_tier:
A

category:
admin-web
```

### Estado incorrecto

```text
repository:
https://github.com/JhonMA82/tanstack-admin

pin:
v1.0.0
```

### Acción

Mantener:

```text
id = tanstack-admin
```

pero restaurar:

```text
repo
pin
legacy identity metadata
```

según el catálogo histórico.

---

## 3.3 Hono API / API Starter

### Legacy canónico

```text
id:
hono-api

name:
Consulting API Starter

repository:
https://github.com/JhonMA82/api-starter

pin:
360eb274cc5936fee5aab88eb8bd94977e95dfc9

decision_status:
default

delivery_status:
released

maintenance_tier:
A

category:
typescript-api
```

### Estado incorrecto

```text
repository:
https://github.com/JhonMA82/hono-api

pin:
v1.0.0
```

### Acción

Mantener el ID lógico:

```text
hono-api
```

pero restaurar el repositorio real:

```text
JhonMA82/api-starter
```

No renombrar el ID a `api-starter` solo por coincidir con GitHub.

El ID del catálogo y el repo son conceptos diferentes.

---

## 3.4 TanStack Transactional PWA

### Legacy canónico

```text
id:
tanstack-transactional-pwa

repository:
https://github.com/JhonMA82/tanstack-transactional-pwa

pin:
f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c

decision_status:
specialized

delivery_status:
curated

maintenance_tier:
B
```

### Estado incorrecto

La URL actual es correcta, pero el pin:

```text
v1.0.0
```

no proviene del legacy.

### Acción

Restaurar el commit histórico:

```text
f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c
```

y revalidar el adapter actual contra ese commit.

---

# 4. Boilerplates que ya conservan la identidad correcta

No modificar innecesariamente estas entradas si sus datos actuales ya son compatibles con el legacy.

## Ignite

```text
id:
ignite

repository:
https://github.com/infinitered/ignite

legacy pin:
e829d2f922c5568a59a77bfb6232aeb500be3f13
```

La operación `generate` de v1 puede conservarse.

No reemplazarla por la implementación Python anterior.

---

## Tauri UI

```text
id:
tauri-ui

repository:
https://github.com/agmmnn/tauri-ui

pin:
8eb86d894c19b6df04ff883ab28b412b1e5f23ea
```

---

## SpeedPy

```text
id:
speedpy

repository:
https://github.com/speedpy/speedpy

pin:
3fbf725d8e9cf6b8aadb3aeaf1db2822522282b9
```

---

## React Starter Kit

```text
id:
react-starter-kit

repository:
https://github.com/kriasoft/react-starter-kit

pin:
0aa7603435f16159ad0b8fef68fb7f6280be7ca1
```

---

# 5. Entradas legacy omitidas que deben recuperarse

El catálogo legacy tenía tres entradas adicionales.

Deben migrarse conservando su intención histórica.

---

## 5.1 Next Admin

```text
id:
next-admin

legacy_id:
next-shadcn-admin-dashboard

repository:
https://github.com/arhamkhnz/next-shadcn-admin-dashboard

pin:
15e0a081bc1acad2b47adc638471b6e67fa36f10

decision_status:
alternative

delivery_status:
curated

maintenance_tier:
B

category:
admin-web
```

### Regla

No convertirlo en default.

Debe seguir siendo una alternativa para escenarios donde una restricción concreta favorezca Next.

Agregarlo al catálogo v1 con:

```text
adapter
technology metadata
Surface
curation reference
```

según el contrato actual.

Si la evidence actual necesaria para `curated` no ha sido migrada o no cumple los gates v1:

```text
no inventar evidence
```

En ese caso:

```text
preservar provenance histórica
+
dejar temporalmente delivery_status compatible con pilot-ready
+
documentar por qué todavía no fue promovido
```

No perder la entrada.

---

## 5.2 FastAPI

### Legacy

```text
id:
fastapi

legacy_id:
full-stack-fastapi-template

repository:
https://github.com/fastapi/full-stack-fastapi-template

decision_status:
alternative

delivery_status:
pilot-ready

maintenance_tier:
B

category:
python-api
```

### Importante

El catálogo legacy mostrado no tenía un upstream commit curado para esta entrada.

Por lo tanto:

```text
NO INVENTAR PIN
NO USAR v1.0.0
NO PROMOVER A STABLE
```

Procedimiento:

1. migrar la entrada como conocimiento del catálogo;
2. mantenerla no estable;
3. verificar upstream actual;
4. elegir un commit inmutable únicamente durante un proceso real de curación/pilot;
5. guardar evidence;
6. solo entonces hacerla materializable.

Hasta ese momento debe poder existir como:

```text
alternative
pilot-ready/catalog-only
```

según lo permitido por el schema actual.

---

## 5.3 GoShip

### Legacy

```text
id:
goship

repository:
https://github.com/leomorpho/goship

decision_status:
experimental

delivery_status:
catalog-only

maintenance_tier:
C

category:
go-realtime
```

### Regla

Mantenerlo exactamente como conocimiento experimental.

No:

```text
crear adapter falso
inventar pin
hacerlo selectable
promoverlo
```

Su presencia demuestra que:

```text
catalog entry
!=
materializable foundation
```

Debe permanecer fuera del pool estable hasta un proceso posterior de curación.

---

# 6. Migración de status: preservar semántica, no falsificar gates

El schema v1 puede usar nombres diferentes al legacy.

Por ello no copiar strings mecánicamente si el contrato cambió.

Pero sí preservar el significado.

Ejemplo:

```text
legacy:
decision_status = default
delivery_status = curated
```

debe seguir significando:

```text
opción preferida
+
históricamente curada
```

Sin embargo, si los nuevos gates de v1 requieren:

```text
curation evidence
pilot evidence
adapter v1
```

y todavía falta alguno:

```text
no mentir al validator
```

Usar una transición explícita:

```text
historical_delivery_status: curated
current_delivery_status: pilot-ready
migration_reason: evidence pending v1 validation
```

o metadata/documentación equivalente.

La plataforma debe distinguir:

```text
historial
vs
estado actual verificado
```

---

# 7. Adapter migration

No copiar adapters Python literalmente.

La tarea correcta es:

```text
legacy adapter semantics
        ↓
understand required behavior
        ↓
express with v1 declarative operations
```

Para cada foundation corregida revisar:

```text
fetch
copy
generate
prune
setup
checks
managed_files
```

## Regla

Repo y pin provienen del legacy.

La forma del adapter puede evolucionar a v1.

No mezclar:

```text
historical source identity
```

con:

```text
new materialization implementation
```

---

# 8. Actualizar referencias internas por ID

Al restaurar:

```text
stardrive-public-web
→ stardrive
```

buscar referencias en todo el repositorio.

Pi debe ejecutar equivalente a:

```bash
rg -n "stardrive-public-web|JhonMA82/stardrive-public-web|JhonMA82/tanstack-admin|JhonMA82/hono-api|v1\.0\.0" .
```

Revisar especialmente:

```text
catalog/recipes/
catalog/compatibilities/
catalog/curation/
testdata/
internal/*_test.go
docs/
templates/
golden files
pilot fixtures
```

No hacer reemplazo global ciego de cada `v1.0.0`.

Solo modificar refs relacionadas con estas foundations.

---

# 9. Recipe integrity

Después de restaurar IDs/repos/pins, validar todos los Recipes.

Comprobar:

```text
every referenced boilerplate exists
every referenced Surface exists
every provider is eligible for the Recipe
no Recipe references stardrive-public-web
no dangling aliases
```

Especial atención a:

```text
GP-01 Public Web
GP-02 Admin
GP-04 Mobile
GP-06 Multi-App
```

y cualquier Recipe que use:

```text
stardrive
tanstack-admin
hono-api
tanstack-transactional-pwa
```

No cambiar la lógica de routing solo para hacer pasar la migración.

---

# 10. Legacy migration baseline

Crear un fixture de prueba local que preserve la identidad histórica mínima.

Recomendación:

```text
testdata/catalog/legacy-boilerplate-baseline.json
```

Contenido conceptual:

```json
[
  {
    "id": "stardrive",
    "repo": "https://github.com/peltmonger/stardrive",
    "pin": "5c449810b763140ac72133ff4ae63d8497cce77a"
  },
  {
    "id": "tanstack-admin",
    "repo": "https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard",
    "pin": "e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0"
  },
  {
    "id": "hono-api",
    "repo": "https://github.com/JhonMA82/api-starter",
    "pin": "360eb274cc5936fee5aab88eb8bd94977e95dfc9"
  },
  {
    "id": "tanstack-transactional-pwa",
    "repo": "https://github.com/JhonMA82/tanstack-transactional-pwa",
    "pin": "f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c"
  }
]
```

Este fixture:

```text
NO
```

debe ser cargado por el runtime normal.

Solo sirve como prueba de integridad/migración.

---

# 11. Tests obligatorios

Crear una suite específica.

Ejemplos de nombres:

```text
TestLegacyCatalogCanonicalRepositories
TestLegacyCatalogCanonicalPins
TestLegacyCatalogIDsPreserved
TestNoInventedRepositoryFromCatalogID
TestNoInventedLegacyPins
TestAllLegacyEntriesRepresented
TestLegacyExperimentalEntriesRemainNonSelectable
TestLegacyAlternativeEntriesDoNotBecomeDefault
```

---

## 11.1 Canonical repository test

Debe fallar si:

```text
hono-api.repo != JhonMA82/api-starter
```

o:

```text
tanstack-admin.repo != arhamkhnz/tanstack-shadcn-admin-dashboard
```

etc.

---

## 11.2 Pin test

Debe asegurar los pins conocidos:

```text
stardrive
tanstack-admin
hono-api
tanstack-transactional-pwa
ignite
tauri-ui
speedpy
react-starter-kit
next-admin
```

cuando existe commit histórico.

No exigir pin histórico para entradas donde el legacy no lo tenía.

---

## 11.3 Cardinality/inventory test

No basta con:

```text
current count >= 8
```

Crear una baseline de IDs históricos:

```text
stardrive
tanstack-admin
next-admin
ignite
tauri-ui
speedpy
fastapi
react-starter-kit
goship
hono-api
tanstack-transactional-pwa
```

y comprobar que cada uno:

```text
exists
OR
has an explicit migration tombstone with reason
```

Para esta corrección se espera que los 11 existan.

---

## 11.4 Non-selectable test

Comprobar:

```text
goship
```

no entra en candidate/materialization pool estable.

Comprobar que:

```text
fastapi
```

no se promueve automáticamente solo por haber sido migrado.

---

# 12. Provenance metadata

Agregar una forma ligera de registrar procedencia.

Ejemplo conceptual en cada entrada migrada:

```json
{
  "provenance": {
    "source": "engineering-platform-legacy",
    "legacy_catalog_version": "2.1.0",
    "legacy_id": "hono-api"
  }
}
```

No es obligatorio usar exactamente este schema si el catálogo actual ya tiene un mecanismo equivalente.

El objetivo es poder responder:

```text
¿de dónde salió esta foundation?
```

sin revisar commits históricos.

---

# 13. Network verification

Después de restaurar el catálogo local:

```text
NO
```

hacer unit tests que dependan de GitHub.

Crear/usar un pilot separado que verifique:

```text
repository reachable
commit exists
adapter works
setup works
checks work
```

para foundations materializables.

Prioridad:

```text
stardrive
tanstack-admin
hono-api
tanstack-transactional-pwa
next-admin
```

FastAPI solo después de definir un pin real.

GoShip no requiere pilot estable mientras siga `catalog-only`.

---

# 14. Pin policy

Para foundations con commit histórico:

```text
usar commit SHA
```

No sustituir por:

```text
main
latest
v1.0.0
```

sin un proceso de actualización explícito.

Si posteriormente se desea actualizar una foundation:

```text
curation/update workflow
→ evaluate new upstream
→ pilot
→ update evidence
→ update pin
```

No mezclar:

```text
catalog migration
```

con:

```text
boilerplate upgrade
```

Esta tarea es migración fiel, no actualización tecnológica.

---

# 15. Curation files

Al cambiar:

```text
stardrive-public-web
→ stardrive
```

revisar la evidence existente.

Si el archivo actual se llama:

```text
catalog/curation/stardrive-public-web.md
```

y describe realmente el boilerplate incorrecto/ficticio:

```text
no conservarlo como evidence válida
```

Crear/migrar evidence basada en:

```text
peltmonger/stardrive
+
commit histórico
```

Si el contenido es genérico y correcto, renombrarlo/referenciarlo apropiadamente.

Aplicar lo mismo a:

```text
tanstack-admin
hono-api
tanstack-transactional-pwa
```

La evidence debe describir el repo real.

---

# 16. AGENTS.md de las foundations

Después de apuntar a repos reales, comprobar que cada materialización:

```text
preserva
o
genera
```

las instrucciones correspondientes.

No asumir que todos los upstream tienen `AGENTS.md`.

Si el boilerplate curado ya lo trae:

```text
preservarlo
```

Si Engineering Platform mantiene instrucciones overlay:

```text
aplicarlas mediante adapter
```

No falsificar `managed_files` declarando `AGENTS.md` si el proceso no lo produce.

Agregar test donde corresponda.

---

# 17. No cambiar Product/Architecture semantics

Esta corrección no debe alterar:

```text
ProjectIntent
ProductRequirements
ArchitectureRequirements
TechnicalConstraints
CatalogGap semantics
Gentle handoff
```

Salvo actualización de IDs de provider.

Ejemplo:

```text
public-web provider:
stardrive-public-web
```

debe convertirse en:

```text
public-web provider:
stardrive
```

pero el routing conceptual debe seguir siendo el mismo.

---

# 18. Compatibilidad de proyectos generados durante el periodo de reescritura

Buscar si algún fixture/project manifest usa:

```text
stardrive-public-web
```

Si solo son tests:

```text
migrarlos
```

Si existe posibilidad de proyectos reales ya generados:

crear alias de lectura/migración:

```text
stardrive-public-web → stardrive
```

solo para manifests antiguos.

No mantener dos boilerplates activos equivalentes en el catálogo.

El alias no debe aparecer como provider nuevo.

---

# 19. Validación final

Ejecutar:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
```

Luego:

```text
eng catalog validate
eng version
```

Después ejecutar pilots relevantes.

La suite debe confirmar:

```text
11 legacy entries represented
canonical repos restored
canonical historical pins restored where known
no fake v1.0.0 pins remain for migrated entries
no dangling stardrive-public-web references
recipes valid
routing still passes
handoff still passes
```

---

# 20. Checklist exacta

## Correcciones de identidad

- [ ] `stardrive` usa `peltmonger/stardrive`.
- [ ] `stardrive` usa commit `5c449810b763140ac72133ff4ae63d8497cce77a`.
- [ ] `stardrive-public-web` deja de ser el ID canónico.
- [ ] `tanstack-admin` usa `arhamkhnz/tanstack-shadcn-admin-dashboard`.
- [ ] `tanstack-admin` usa commit `e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0`.
- [ ] `hono-api` usa `JhonMA82/api-starter`.
- [ ] `hono-api` usa commit `360eb274cc5936fee5aab88eb8bd94977e95dfc9`.
- [ ] `tanstack-transactional-pwa` usa commit `f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c`.

## Entradas omitidas

- [ ] `next-admin` restaurado.
- [ ] `fastapi` restaurado sin inventar pin.
- [ ] `goship` restaurado como experimental/catalog-only.
- [ ] El inventario histórico de 11 IDs está representado.

## Integridad

- [ ] Recipes actualizados.
- [ ] Curation references apuntan al upstream real.
- [ ] Tests/golden fixtures actualizados.
- [ ] No quedan repos ficticios creados desde IDs.
- [ ] No quedan pins ficticios `v1.0.0` para estas entradas.
- [ ] Aliases legacy solo se usan para compatibilidad, no como providers duplicados.

## Calidad

- [ ] `go test ./...` pasa.
- [ ] `go vet ./...` pasa.
- [ ] `eng catalog validate` pasa.
- [ ] Routing dataset sigue verde.
- [ ] Pilots de foundations materializables pasan o quedan explícitamente pendientes con status correcto.

---

# 21. Definition of Done

Esta corrección se considera terminada cuando:

1. el catálogo Go representa fielmente las 11 entradas históricas;
2. los repositorios canónicos coinciden con el legacy;
3. los pins históricos conocidos coinciden con el legacy;
4. las entradas sin pin histórico no reciben uno inventado;
5. no existe un provider activo `stardrive-public-web` duplicando `stardrive`;
6. Recipes apuntan a IDs canónicos;
7. curation/evidence corresponde a repos reales;
8. el schema v1 sigue validando;
9. routing tests siguen pasando;
10. materialization tests siguen pasando;
11. existe una prueba permanente de integridad de migración;
12. ninguna corrección requirió rediseñar el core.

---

# 22. Commit strategy recomendada

Usar commits pequeños:

```text
fix(catalog): restore canonical legacy boilerplate sources
fix(catalog): restore historical immutable pins
feat(catalog): migrate omitted legacy entries
fix(recipes): use canonical stardrive provider id
fix(curation): align evidence with canonical upstreams
test(catalog): add legacy migration integrity baseline
test(pilots): validate migrated foundations
docs(catalog): document legacy provenance
```

No hacer un único commit gigante si puede evitarse.

---

# 23. Prompt operativo para Pi Agent

```text
Repair the Engineering Platform v1 catalog migration using
JhonMA82/engineering-platform-legacy/platform/boilerplates.json as the
historical source of truth.

Do not redesign the Go core.

Restore canonical repository identity and historical immutable pins:

stardrive
  repo: https://github.com/peltmonger/stardrive
  pin: 5c449810b763140ac72133ff4ae63d8497cce77a

tanstack-admin
  repo: https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard
  pin: e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0

hono-api
  repo: https://github.com/JhonMA82/api-starter
  pin: 360eb274cc5936fee5aab88eb8bd94977e95dfc9

tanstack-transactional-pwa
  repo: https://github.com/JhonMA82/tanstack-transactional-pwa
  pin: f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c

Restore the canonical catalog id `stardrive`; do not keep
`stardrive-public-web` as a duplicate active provider.

Also migrate the three omitted legacy entries:

next-admin
fastapi
goship

Preserve their historical decision/delivery intent. Do not promote FastAPI or
GoShip and do not invent pins where the legacy catalog did not contain one.

Update all Recipes, curation references, fixtures and tests that depend on the
incorrect IDs/repos.

Add a local legacy migration baseline test so future rewrites cannot infer a
repository URL from the catalog id or invent a release tag.

Do not make unit tests depend on GitHub. Use separate pilots for network/upstream
verification.

When complete run:
- gofmt
- go vet ./...
- go test ./...
- go build ./cmd/eng
- eng catalog validate
- relevant pilots

Produce a concise migration report showing:
- all 11 legacy IDs
- canonical repo
- canonical pin or "unvalidated"
- current v1 status
- pilot status
- any intentional status downgrade required by v1 curation gates.
```
