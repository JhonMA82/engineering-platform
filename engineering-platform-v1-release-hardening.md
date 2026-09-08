# Engineering Platform 1.0 — Release Hardening
## Instrucciones para corregir observaciones antes de etiquetar `v1.0.0`

**Estado:** Ready for implementation  
**Objetivo:** cerrar los pendientes detectados en la auditoría de la reescritura Go sin modificar la arquitectura general  
**Repositorio objetivo:** `JhonMA82/engineering-platform`  
**Legacy:** `JhonMA82/engineering-platform-legacy`  
**Principio:** hardening de release, no nueva reescritura

---

# 0. Instrucción maestra para Pi Agent

Aplicar este documento sobre la implementación actual de Engineering Platform.

La arquitectura general está aprobada y **no debe reabrirse**.

No:

- reestructurar nuevamente todo el repositorio;
- volver a discutir Go vs Python;
- reintroducir `project_type`;
- mezclar Resolver con Materializer;
- mover decisiones del core a Pi;
- crear un plugin runtime;
- ampliar scope con features nuevas que no estén relacionadas con el cierre de v1.0.

El objetivo es corregir únicamente los contratos incompletos y endurecer la release.

Orden recomendado:

```text
1. Core ↔ Catalog compatibility
2. Typed technical constraints
3. Catalog / curation enforcement
4. Ignite materialization completeness
5. Gentle handoff / SDD contract
6. CI + release hardening
7. Final v1.0 readiness audit
```

Cada bloque debe cerrarse con:

```text
code
+ tests
+ docs
+ acceptance criteria
```

antes de avanzar al siguiente.

---

# 1. Core ↔ Catalog Compatibility

## Problema

El catálogo declara una versión mínima de core, pero la validación actual no garantiza que el binario en ejecución satisfaga realmente ese requisito.

Esto puede permitir cargar un catálogo que dependa de una capacidad que el engine todavía no entiende.

## Objetivo

Implementar una comprobación real de compatibilidad semver entre:

```text
eng CoreVersion
catalog.min_core_version
```

y, si aplica:

```text
catalog.max_core_version
catalog.schema_version
```

---

## 1.1 Contrato mínimo

`catalog/metadata.json` debe declarar como mínimo:

```json
{
  "catalog_version": "2026.09.07",
  "schema_version": 1,
  "min_core_version": "1.0.0"
}
```

Opcionalmente:

```json
"max_core_version": "1.x"
```

solo si aparece una necesidad real.

No añadir complejidad innecesaria.

---

## 1.2 Validación requerida

Al cargar catálogo:

```text
if core_version < min_core_version:
    reject catalog
```

Error esperado:

```text
catalog requires core >= 1.1.0, running core is 1.0.0
```

Si el schema del catálogo no es soportado:

```text
unsupported catalog schema version: 2
```

---

## 1.3 Fuente única de versión

La versión del binario debe existir en un lugar canónico.

No mantener versiones manuales divergentes en múltiples archivos.

Preferencia:

```go
package version

var CoreVersion = "dev"
```

y durante release:

```text
-ldflags
```

o mecanismo equivalente.

`eng version` debe mostrar:

```text
Core: 1.0.0
Catalog: 2026.09.07
Catalog schema: 1
Commit: ...
Build date: ...
```

cuando los valores estén disponibles.

---

## 1.4 Tests obligatorios

Agregar:

```text
TestCatalogRejectsNewerRequiredCore
TestCatalogAcceptsCompatibleCore
TestCatalogRejectsUnsupportedSchema
TestVersionCommandDisplaysCoreAndCatalog
```

No depender de red.

---

## Acceptance Criteria

- [ ] Un catálogo incompatible no puede cargarse.
- [ ] Un catálogo compatible sí carga.
- [ ] `eng version` diferencia claramente core y catálogo.
- [ ] Existe un único origen canónico de CoreVersion.
- [ ] Los tests cubren versiones mayores/menores/iguales.

---

# 2. Technical Constraints tipadas

## Problema

Las restricciones técnicas explícitas del usuario todavía no están suficientemente tipadas por dominio.

Ejemplo:

```text
must-use: turso
```

no indica si `turso` restringe:

```text
database
runtime
framework
provider
deployment
```

Esto impide resolver correctamente escenarios técnicos como:

> "Quiero un dashboard TanStack offline usando Turso."

## Objetivo

Modelar restricciones técnicas explícitas con un target estable.

---

## 2.1 Modelo propuesto

Conceptualmente:

```go
type TechnicalConstraint struct {
    Target string
    Kind   string
    Value  string
}
```

Ejemplo JSON:

```json
{
  "target": "framework",
  "kind": "must-use",
  "value": "tanstack"
}
```

```json
{
  "target": "database",
  "kind": "must-use",
  "value": "turso"
}
```

Targets iniciales recomendados:

```text
framework
language
runtime
database
deployment
provider
```

No crear una taxonomía enorme inicialmente.

---

## 2.2 Constraint vs Preference

Mantener diferencia explícita:

```text
must-use
must-not-use
prefer
avoid
```

Ejemplo:

```json
{
  "target": "language",
  "kind": "prefer",
  "value": "python"
}
```

`prefer` afecta ranking.

`must-use` puede eliminar candidatos incompatibles.

---

## 2.3 Integración en Resolver

El Resolver debe considerar las restricciones antes del scoring final.

Ejemplo:

```text
must-use framework=tanstack

candidate Next:
rejected

candidate TanStack:
eligible
```

Pero la evaluación debe ocurrir únicamente si el catálogo conoce la metadata técnica correspondiente.

No hardcodear:

```text
tanstack
turso
python
go
```

dentro del resolver.

---

## 2.4 Integración en Composer

El Composer debe resolver database profile usando constraints cuando corresponda.

Ejemplo:

```text
must-use database=turso
```

Si existe profile curado compatible:

```text
select Turso profile
```

Si no existe:

```text
CatalogGap
kind: database-profile
value: turso
```

si la restricción era obligatoria.

Si solo era:

```text
prefer database=turso
```

puede seleccionar una alternativa compatible y explicar la desviación.

---

## 2.5 Catalog metadata

Boilerplates y profiles deben poder declarar metadata técnica.

Ejemplo:

```json
{
  "technology": {
    "framework": ["tanstack"],
    "language": ["typescript"],
    "runtime": ["bun"]
  }
}
```

Database profile:

```json
{
  "id": "turso",
  "engine": "libsql",
  "provider": "turso",
  "supports": [
    "offline",
    "sync"
  ]
}
```

No agregar Turso si todavía no ha sido validado/curado.

Primero implementar el contrato.

Después agregar profiles reales.

---

## 2.6 Tests mínimos

Crear escenarios:

```text
TanStack must-use selects TanStack provider
Next candidate rejected by must-use TanStack
Prefer Python influences TUI selection
Must-use database with missing profile → CatalogGap
Prefer unavailable database → explanation + fallback allowed
```

---

## Acceptance Criteria

- [ ] Technical constraints tienen `target`.
- [ ] Resolver no depende de nombres hardcodeados.
- [ ] `must-use` puede rechazar candidatos incompatibles.
- [ ] `prefer` no se trata como hard constraint.
- [ ] Composer puede resolver database profile mediante constraints.
- [ ] Missing mandatory database profile produce CatalogGap arquitectónico.
- [ ] Tests cubren TanStack/database/language scenarios.

---

# 3. Catalog y Curation Enforcement

## Problema

Una foundation puede aparecer como seleccionable aunque la evidencia de curación no esté vinculada formalmente por contrato.

La plataforma se basa en boilerplates curados, por lo que esto debe ser garantizado por validación.

## Objetivo

Una foundation no puede declararse lista para materialización sin evidencia mínima de curación.

---

## 3.1 Contrato requerido

Cada boilerplate seleccionable debe declarar algo equivalente a:

```json
{
  "curation": {
    "status": "curated",
    "evidence": "curation/tanstack-admin.md"
  }
}
```

o:

```text
curation_evidence
```

según el esquema actual.

No importa el nombre exacto.

Importa que exista la referencia verificable.

---

## 3.2 Reglas por status

Propuesta:

### `catalog-only`

Puede existir sin pilot completo.

No seleccionable para materialización normal.

### `pilot-ready`

Debe tener:

```text
license checked
repository checked
pin defined
adapter present
basic evidence
```

### `curated`

Debe tener:

```text
all above
+ successful pilot
+ maintenance evidence
+ architectural fit evidence
```

### `released`

Debe cumplir todo lo anterior y estar habilitado para selección estable.

---

## 3.3 Validator

`eng catalog validate` debe detectar:

```text
missing curation evidence
missing adapter
missing pin
broken curation reference
released foundation without successful pilot evidence
```

No basta con que el archivo JSON sea sintácticamente válido.

---

## 3.4 Referencias seguras

Las rutas declaradas deben permanecer dentro del catálogo.

Rechazar:

```text
../
absolute paths
symlink escapes
```

si aplica.

---

## 3.5 Tests

Agregar:

```text
released boilerplate without evidence → invalid
catalog-only without full evidence → valid
broken evidence path → invalid
curated foundation with valid evidence → valid
```

---

## Acceptance Criteria

- [ ] Toda foundation `curated/released` tiene evidence vinculada.
- [ ] `catalog validate` detecta evidence faltante.
- [ ] Un boilerplate incompleto no entra al pool estable.
- [ ] La validación distingue catalog-only/pilot/curated/released.
- [ ] No se requiere modificar el core para añadir nueva evidence.

---

# 4. Ignite / Generated Boilerplate Materialization

## Problema

La foundation Ignite requiere una forma de materialización distinta de un simple `git clone + copy`.

La implementación actual debe verificar que el adapter realmente reproduzca el flujo necesario o, si no puede representarlo, extender el engine de manera mínima y genérica.

## Objetivo

Cerrar correctamente la materialización de Ignite sin introducir una excepción específica:

```text
if boilerplate == ignite
```

en el core.

---

## 4.1 Investigación previa

Antes de modificar engine:

1. determinar cómo se genera actualmente un proyecto Ignite;
2. verificar si el repositorio puede usarse directamente como starter;
3. comprobar si requiere generator/CLI;
4. identificar inputs mínimos;
5. identificar qué archivos genera;
6. identificar comandos de setup/check;
7. documentar el resultado.

Crear:

```text
docs/decisions/ignite-materialization.md
```

breve.

---

## 4.2 Preferencia

Si el proceso puede expresarse usando primitives existentes:

```text
fetch
copy
template
run curated command
check
```

usar adapter solamente.

No modificar core.

---

## 4.3 Si falta una primitive real

Solo si el flujo no puede representarse de manera segura, crear una operación genérica.

Ejemplo conceptual:

```text
generate
```

con contrato como:

```json
{
  "operation": "generate",
  "command": "...",
  "args": [],
  "output": "..."
}
```

La operación debe ser útil para más casos futuros, no solo Ignite.

No permitir shell arbitrario.

---

## 4.4 Seguridad

La nueva operación, si existe, debe:

- usar argv estructurado;
- no usar `sh -c`;
- controlar cwd;
- validar destination;
- validar output path;
- limpiar staging en error;
- capturar exit code;
- preservar logs útiles;
- no leer secretos arbitrariamente.

---

## 4.5 Pilot real

Agregar un pilot que pruebe la materialización real de Ignite cuando sea razonable.

Separar:

```text
offline deterministic test
```

de:

```text
upstream/network pilot
```

Los tests normales no deben depender de red.

---

## Acceptance Criteria

- [ ] Ignite puede materializarse correctamente.
- [ ] No existe `if ignite` especial en el core.
- [ ] Si se agregó una primitive, es genérica y documentada.
- [ ] Existe test offline del engine.
- [ ] Existe pilot real o evidencia equivalente del flujo.
- [ ] Adapter declara setup/check correctamente.

---

# 5. Gentle Handoff: Direct Build vs SDD

## Problema

El handoff actual ya indica que Gentle no debe redescubrir arquitectura, pero debe quedar explícito que Gentle decide si:

```text
implementa directo
```

o:

```text
inicia una sesión SDD
```

según el estado de los requisitos de producto.

## Objetivo

Formalizar el ownership transition:

```text
Pi
→ Engineering Platform
→ Gentle
```

sin que Gentle vuelva a preguntar lo ya resuelto.

---

## 5.1 `GENTLE.md`

Agregar instrucciones explícitas equivalentes a:

```text
You are taking ownership of a project generated by Engineering Platform.

Do not rediscover or replace the selected architecture unless a concrete
contradiction is found.

Read:
1. .engineering/implementation-brief.md
2. AGENTS.md
3. ARCHITECTURE.md
4. .engineering/project-map.json
5. the AGENTS.md of the Surface you will modify

Then decide:

A. Direct implementation
   Use this path when the product requirements are sufficiently defined.

B. SDD session
   Use this path only when important product/domain rules remain undefined.

An SDD session must focus on product behavior, workflows, rules, permissions,
edge cases and domain decisions.

Do not ask the user to repeat information already present in the handoff.
Do not reopen framework, database, boilerplate or topology choices unless a
real contradiction is discovered.
```

---

## 5.2 Open Product Questions

No depender únicamente de `ArchitectureDecision.unresolved_dimensions`.

Después de resolver arquitectura normalmente:

```text
unresolved_dimensions = 0
```

pero todavía pueden existir preguntas de producto.

Agregar soporte para:

```text
open_product_questions
```

en `ProjectIntent` o en un objeto de handoff derivado.

Ejemplos:

```text
define approval roles
define report fields
define cancellation behavior
define audit requirements
```

Estas preguntas no afectan necesariamente el routing.

---

## 5.3 Handoff JSON

Extender si es necesario:

```json
{
  "status": "ready_for_implementation",
  "next_owner": "gentle-ai",

  "locked": [
    "architecture",
    "surface-topology",
    "selected-foundations",
    "database-profile"
  ],

  "product_requirements": [],
  "open_product_questions": [],

  "implementation_mode": "gentle-decides"
}
```

No establecer:

```text
implementation_mode = sdd
```

desde Engineering Platform.

Gentle decide.

---

## 5.4 Implementation Brief

Debe distinguir:

```text
Already provided by foundation
Pending product implementation
Open product questions
Locked architecture decisions
```

No duplicar todos los documentos.

---

## 5.5 Tests

Golden tests para:

```text
GENTLE.md contains direct-build/SDD rule
handoff.json exposes open product questions
root AGENTS routes correctly
project-map references all materialized Surfaces
```

---

## Acceptance Criteria

- [ ] Gentle puede tomar control sin reexplicar el proyecto.
- [ ] `GENTLE.md` define Direct Build vs SDD.
- [ ] SDD queda limitado a producto/dominio.
- [ ] Existen open product questions separadas de routing ambiguity.
- [ ] Architecture decisions aparecen como locked.
- [ ] `AGENTS.md` y `project-map.json` siguen siendo consistentes.

---

# 6. CI y Release Engineering

## Problema

La CI actual demuestra buena salud del código, pero antes de `v1.0.0` debe endurecerse el proceso de release.

## Objetivo

Validar consistentemente:

```text
format
vet
tests
catalog
routing
builds
pilots
release artifacts
```

---

## 6.1 CI principal

Agregar pasos explícitos:

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./cmd/eng
```

El check de formato debe fallar si:

```text
gofmt -l
```

produce salida.

---

## 6.2 Catalog validation

Ejecutar:

```text
eng catalog validate
```

o el application service equivalente en CI.

No confiar únicamente en tests indirectos.

---

## 6.3 Routing guard

El dataset actual ya supera 40 casos.

Cambiar cualquier guard similar a:

```go
if len(cases) < 15
```

por:

```go
if len(cases) < 40
```

o preferiblemente:

```text
minimum defined as constant/policy
```

para evitar regresiones silenciosas.

---

## 6.4 Build matrix

Construir:

```text
linux/amd64
linux/arm64
darwin/arm64
windows/amd64
```

Opcionalmente:

```text
darwin/amd64
```

si se considera útil.

No es necesario ejecutar todos los binarios en sus OS dentro del primer workflow si la cross-compilation es suficiente.

---

## 6.5 Release artifacts

Para tags:

```text
v1.0.0
```

generar:

```text
eng-linux-amd64
eng-linux-arm64
eng-darwin-arm64
eng-windows-amd64.exe
checksums.txt
```

Si existe packaging adicional, hacerlo después.

No bloquear v1 con Homebrew/AUR/etc.

---

## 6.6 Immutable pins

Revisar adapters.

Preferencia de máxima reproducibilidad:

```text
commit SHA
```

Si se usa tag:

```text
tag
+ expected commit SHA
```

cuando sea razonable.

Evitar depender exclusivamente de refs mutables.

---

## 6.7 Pi dependency pin

No usar:

```json
"@juicesharp/rpiv-ask-user-question": "*"
```

Fijar una versión o rango compatible.

Ejemplo:

```json
"@juicesharp/rpiv-ask-user-question": "^2.9.0"
```

si esa versión ha sido validada con la integración actual.

No actualizar automáticamente a cualquier major futura.

---

## 6.8 Pilots workflow

Mantener separado:

```text
fast deterministic CI
```

de:

```text
upstream pilots
```

Los pilots pueden validar:

- materialization;
- real boilerplate pins;
- setup commands;
- checks;
- adapters.

No hacer que una caída temporal de GitHub convierta cada unit test en fallo.

---

## Acceptance Criteria

- [ ] CI falla por formato incorrecto.
- [ ] CI ejecuta `go vet`.
- [ ] CI ejecuta todos los tests.
- [ ] CI valida catálogo explícitamente.
- [ ] Routing guard exige >= 40 casos.
- [ ] Build matrix genera las plataformas objetivo.
- [ ] Pi question plugin está pinneado.
- [ ] Pins críticos son reproducibles.
- [ ] Release genera checksums.

---

# 7. Pruebas de no regresión arquitectónica

Antes de cerrar v1, agregar o conservar tests que garanticen las decisiones fundamentales del PRD.

## 7.1 ProjectIntent

Debe comprobarse:

```text
no project_type required
```

---

## 7.2 Product features

Caso:

```text
TUI foundation supports architecture
PDF missing
```

debe producir:

```text
resolved
```

no:

```text
catalog-gap
```

---

## 7.3 CatalogGap

Caso:

```text
required Surface = tui
no TUI provider in catalog
```

debe producir:

```text
catalog-gap
```

---

## 7.4 New Surface without core changes

Fixture de catálogo:

```text
surface = test-surface
provider = test-provider
```

Debe cargarse y resolverse si los contratos existentes son suficientes.

El test demuestra que no existe un enum cerrado escondido.

---

## 7.5 Technical user

Caso:

```text
must-use framework=tanstack
```

debe excluir provider incompatible.

---

## 7.6 Multi-surface handoff

Fixture:

```text
API
Dashboard
Mobile
```

Debe generar:

```text
root AGENTS.md
project-map.json
surface AGENTS references
ARCHITECTURE.md
implementation-brief.md
handoff.json
GENTLE.md
```

---

## 7.7 Resolver purity

Resolver tests no deben necesitar:

```text
filesystem
git
network
Pi
process execution
```

---

# 8. Documentación a actualizar

Actualizar únicamente la documentación afectada.

Mínimos:

```text
README.md
docs/architecture/routing.md
docs/architecture/materialization.md
docs/architecture/handoff.md
docs/concepts/catalog.md
docs/concepts/technical-constraints.md
docs/guides/add-boilerplate.md
docs/guides/release.md
```

Si esos archivos no existen, crear solo los necesarios.

No generar documentación redundante.

---

# 9. ADRs recomendados

Crear o actualizar brevemente:

```text
ADR: Core/Catalog version compatibility
ADR: Typed technical constraints
ADR: Curated foundation eligibility
ADR: Gentle ownership handoff
ADR: Catalog evolution independent from core releases
```

Solo si todavía no están cubiertos por ADR existentes.

No crear ADRs duplicados.

---

# 10. Orden de implementación

## Fase H1 — Compatibility

Implementar:

```text
core/catalog semver
schema compatibility
eng version
```

Cerrar tests.

---

## Fase H2 — Technical constraints

Implementar:

```text
targeted constraints
resolver filtering
database profile selection
catalog-gap for mandatory missing profile
```

Cerrar tests.

---

## Fase H3 — Curation contract

Implementar:

```text
evidence references
eligibility validator
catalog validation
```

Cerrar todos los boilerplates existentes hasta que el catálogo completo sea válido.

---

## Fase H4 — Ignite

Investigar y cerrar materialización real.

No avanzar con workaround específico.

---

## Fase H5 — Gentle contract

Actualizar:

```text
GENTLE.md
handoff.json
implementation-brief
open product questions
```

Agregar golden tests.

---

## Fase H6 — CI/release

Agregar:

```text
format
catalog validate
routing >= 40
build matrix
release artifacts
checksum
dependency pins
```

---

## Fase H7 — Final Audit

Ejecutar:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
```

Después:

```text
eng catalog validate
eng version
```

y pilots correspondientes.

---

# 11. Definition of Done — v1.0 Release Hardening

No etiquetar `v1.0.0` hasta cumplir:

- [ ] Core y catálogo validan compatibilidad real.
- [ ] CoreVersion y catalog version son independientes.
- [ ] Technical constraints tienen target.
- [ ] Database constraints son resolubles.
- [ ] Missing mandatory database provider genera CatalogGap.
- [ ] Curation evidence es obligatoria para foundations estables.
- [ ] Todos los boilerplates seleccionables pasan catalog validation.
- [ ] Ignite tiene materialización válida o queda explícitamente fuera del pool estable.
- [ ] No hay casos especiales hardcodeados por boilerplate en el core.
- [ ] Gentle recibe Direct Build vs SDD instructions.
- [ ] Open product questions están separadas de routing ambiguity.
- [ ] Multi-surface handoff genera mapa correcto.
- [ ] CI verifica formato.
- [ ] CI verifica vet.
- [ ] CI ejecuta tests.
- [ ] CI valida catálogo.
- [ ] Routing dataset tiene guard >= 40.
- [ ] Build matrix multiplataforma pasa.
- [ ] Pi plugin tiene versión fijada.
- [ ] Pins críticos son reproducibles.
- [ ] Release artifacts incluyen checksums.
- [ ] `go test ./...` pasa.
- [ ] `go vet ./...` pasa.
- [ ] `eng catalog validate` pasa.
- [ ] Pilots relevantes pasan.

---

# 12. Regla de alcance durante este hardening

Si durante estas correcciones aparece una idea nueva que no bloquea v1:

```text
documentar
crear issue
dejar para v1.1+
```

No ampliar la release.

Ejemplos para post-v1 si no son necesarios ahora:

```text
web UI
remote catalog registry
automatic boilerplate research
marketplace
plugin runtime
cloud sync
telemetry
additional agent integrations
advanced update workflows
```

---

# 13. Resultado esperado

Al terminar estas instrucciones, Engineering Platform debe poder considerarse:

```text
architecture stable
resolver stable
catalog contract stable
materialization stable
agent handoff stable
release pipeline stable
```

y entonces:

```text
tag v1.0.0
```

será razonable.

---

# 14. Prompt operativo para Pi Agent

Usar este bloque como instrucción de arranque:

```text
Apply the Engineering Platform 1.0 Release Hardening plan in this document.

Do not redesign or rewrite the architecture.

Work in this order:

1. Core/Catalog compatibility
2. Typed technical constraints
3. Curation enforcement
4. Ignite materialization completeness
5. Gentle direct-build vs SDD handoff
6. CI and release hardening

Close each phase with tests before moving to the next.

Preserve the existing architectural boundaries:
Pi → ProjectIntent
Resolver → ArchitectureDecision
Composer → MaterializationPlan
Materializer → project
Handoff → Gentle

Missing product features must never become CatalogGaps.

Do not add boilerplate-specific conditionals to the core.

Do not add new product scope unless required to close one of the audited gaps.

When complete, run the full validation suite and produce a concise v1.0 readiness report.
```
