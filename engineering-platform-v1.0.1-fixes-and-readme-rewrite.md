# Engineering Platform v1.0.1 — Small Fixes + README Rewrite
## Instrucciones operativas para Pi Agent

**Estado:** Ready for implementation  
**Repositorio objetivo:** `JhonMA82/engineering-platform`  
**Base:** `v1.0.0`  
**Objetivo:** publicar una `v1.0.1` pequeña, sin rediseñar la arquitectura, corrigiendo dos edge cases funcionales y rehaciendo el README como puerta de entrada real al proyecto.

---

# 0. Instrucción maestra

Aplicar este documento como un **patch release**.

No:

- rediseñar Engineering Platform;
- reabrir ProjectIntent/Resolver/Composer/Materializer;
- agregar nuevas Recipes;
- agregar nuevos boilerplates salvo que sea estrictamente necesario para un test;
- ampliar el scope del catálogo;
- introducir nuevas features no relacionadas con estos fixes;
- convertir el README en documentación interna del core;
- documentar comandos o installation paths que no hayan sido validados.

La release `v1.0.1` debe contener únicamente:

```text
1. fix: must-not-use database
2. fix/contract: deployment technical constraint
3. docs: README completo orientado a usuario
4. docs cleanup necesario por la migración del catálogo
5. tests de regresión
```

---

# 1. Fix: `must-not-use database`

## Problema

Una restricción:

```text
must-not-use database=postgresql
```

es una hard constraint.

Si la Recipe/composición solo dispone de PostgreSQL, Engineering Platform no puede:

```text
mantener PostgreSQL
+
mostrar una nota
```

porque estaría violando una condición obligatoria del usuario.

## Comportamiento correcto

Caso:

```text
must-not-use database=postgresql
allowed profiles:
- postgresql
```

Resultado:

```text
NO valid database profile
```

La resolución/composición debe terminar en un estado explícito:

```text
CatalogGap
```

o:

```text
composition failure
```

según la frontera actual del código.

Preferencia:

- si la arquitectura es válida pero el catálogo no tiene ningún profile compatible, usar `CatalogGap`;
- si el plan quedó internamente contradictorio, usar un error de composición tipado.

No devolver un plan que contenga el database prohibido.

---

## 1.1 No cambiar la semántica de `prefer`

Esto solo aplica a:

```text
must-not-use
```

Una preferencia:

```text
avoid database=postgresql
```

sí puede terminar usando PostgreSQL cuando no exista alternativa, siempre que la explicación deje claro por qué.

Separar:

```text
must-not-use → obligatorio
avoid        → preferencia
```

---

## 1.2 Tests obligatorios

Agregar:

```text
TestDatabaseMustNotUseWithAlternative
TestDatabaseMustNotUseWithoutAlternative
TestDatabaseAvoidWithoutAlternativeMayFallback
```

Caso crítico:

```text
must-not-use postgresql
recipe only allows postgresql
```

Debe comprobar:

```text
no plan with postgresql
```

---

## Acceptance Criteria

- [ ] `must-not-use` nunca produce el profile prohibido.
- [ ] Existe un resultado explicable cuando no hay alternativa.
- [ ] `avoid` continúa siendo soft preference.
- [ ] El test de regresión reproduce el bug anterior.
- [ ] Resolver/Composer no tienen excepciones por nombre de database.

---

# 2. Fix/Contract: `deployment` Technical Constraint

## Problema

`deployment` aparece como target técnico soportado, pero actualmente no existe metadata suficiente en catálogo para evaluar de manera determinista:

```text
must-use deployment=edge
must-not-use deployment=serverless
```

Un hard constraint que el engine acepta pero no puede comprobar es un contrato engañoso.

## Decisión recomendada para v1.0.1

**No implementar todavía un sistema de deployment.**

Retirar `deployment` de los targets hard soportados hasta que exista un modelo real en catálogo.

Esto mantiene v1.0.1 pequeña.

---

## 2.1 Comportamiento esperado

Si el usuario/Pi produce:

```json
{
  "target": "deployment",
  "kind": "must-use",
  "value": "edge"
}
```

el validator debe devolver algo explícito:

```text
unsupported technical constraint target: deployment
```

o equivalente.

No:

```text
accept + ignore
```

---

## 2.2 Posible tratamiento como nota/preferencia

Si se desea preservar la intención:

```text
prefer deployment=edge
```

puede permanecer como:

```text
unresolved/non-routing preference
```

solo si el modelo actual ya tiene una forma clara de representarlo.

No complicar esta release para conservarlo.

La solución más limpia es:

```text
deployment unsupported in v1.0.1
```

y documentarlo.

---

## 2.3 Roadmap futuro

Crear issue/nota para una versión futura:

```text
Deployment Profiles
```

que podría incluir:

```text
edge
serverless
container
vm
static
managed-platform
```

pero **no implementarlo ahora**.

---

## Acceptance Criteria

- [ ] `deployment` no se anuncia como hard target soportado si no se evalúa.
- [ ] Un `must-use deployment=*` no puede ser aceptado silenciosamente.
- [ ] Existe test de rechazo.
- [ ] Documentación refleja los targets realmente soportados.
- [ ] No se agrega arquitectura nueva en v1.0.1.

---

# 3. README — objetivo de reescritura

## Problema actual

El README actual funciona más como:

```text
registro de milestones
+
notas de implementación
+
manual parcial del core
```

que como presentación del producto.

Empieza hablando de:

```text
M1 Deterministic Routing Core
M2 Reproducible Composition
M3 End-to-End Project Bootstrap
```

y entra rápidamente en:

```text
ProjectIntent
ArchitectureDecision
Composer
MaterializationPlan
```

antes de explicar de forma simple:

```text
qué es
para quién sirve
qué problema resuelve
cómo se instala
cómo se usa
```

También mezcla información histórica de implementación y referencias que ya fueron corregidas en catálogo.

Para un visitante nuevo, esto crea fricción.

---

# 4. Nuevo propósito del README

El README debe ser la **landing page del proyecto en GitHub**.

Debe permitir que alguien responda en pocos minutos:

```text
¿Qué es Engineering Platform?
¿Qué problema resuelve?
¿Para quién es?
¿Cómo funciona a grandes rasgos?
¿Cómo lo instalo?
¿Cómo empiezo?
¿Qué pasa después de generar el proyecto?
¿Cómo agrego un nuevo boilerplate?
¿Dónde está la documentación técnica?
```

Regla:

> README explains the product first, internals second.

---

# 5. Audiencias del README

Escribir para tres perfiles, en este orden.

## 5.1 Usuario que apenas descubre el proyecto

Necesita entender:

```text
problema
beneficio
flujo
ejemplo
instalación
```

No necesita conocer `R1–R10`.

## 5.2 Usuario técnico que quiere usarlo

Necesita:

```text
binario
Pi integration
commands
ProjectIntent
catalog
project generation
Gentle handoff
```

## 5.3 Contribuidor

Necesita:

```text
cómo modificar catálogo
cómo agregar boilerplate
cómo correr tests
arquitectura interna
docs
```

Esta información va después del quickstart.

---

# 6. Mensaje principal del proyecto

Pi debe redactar una introducción clara y poco técnica.

Idea semántica a conservar:

> Engineering Platform ayuda a convertir una idea de software en una base de proyecto lista para desarrollar.
>
> A partir de lo que el usuario necesita, selecciona entre boilerplates previamente evaluados, decide una arquitectura adecuada, combina las piezas necesarias —por ejemplo API, dashboard, web o mobile— y genera el proyecto con contexto para que un agente de desarrollo pueda continuar sin empezar de cero.

No copiar necesariamente estas frases literalmente.

Evitar como primera descripción:

```text
deterministic local engine for selecting, composing and materializing curated stacks
```

Esa descripción puede conservarse en una sección técnica secundaria.

---

# 7. Explicar qué NO es

Agregar una sección breve:

```text
Engineering Platform no intenta desarrollar toda la aplicación.
```

Explicar:

```text
Engineering Platform:
- selecciona una foundation
- compone la arquitectura
- prepara el proyecto
- deja instrucciones/contexto

Gentle AI:
- implementa las features del producto
- profundiza requisitos si hace falta
```

Esto evita que alguien espere un generador de aplicaciones completas.

---

# 8. Mostrar el flujo en palabras simples

Usar un diagrama pequeño.

Ejemplo conceptual:

```text
Idea
  ↓
Pi hace las preguntas necesarias
  ↓
Engineering Platform selecciona la base
  ↓
compone API / dashboard / mobile / etc.
  ↓
genera el proyecto
  ↓
Gentle AI continúa el desarrollo
```

Después, opcionalmente mostrar la versión técnica:

```text
ProjectIntent
→ Resolver
→ Composer
→ Materializer
→ Development Handoff
```

No invertir el orden.

---

# 9. Ejemplo práctico en el README

Agregar un ejemplo reconocible.

Ejemplo:

```text
"Necesito un sistema con dashboard administrativo,
API y app móvil para registrar operaciones de campo."
```

Explicar:

```text
Pi captura:
- usuarios
- interfaces
- offline/native
- restricciones

Engineering Platform puede resolver:
- API foundation
- admin foundation
- mobile foundation
- database profile
- estructura del repositorio

El resultado puede quedar:
services/api/
apps/admin/
apps/mobile/
```

Y cada Surface conserva sus instrucciones para agentes.

No prometer nombres de provider específicos si no son necesarios para explicar el producto.

---

# 10. Instalación — sección obligatoria

El README debe tener una sección:

```text
## Installation
```

cerca del inicio.

Prioridad de instalación:

```text
1. Prebuilt binary — recomendado
2. Build from source — contributors/developers
```

No obligar al usuario normal a instalar Go si ya existen releases binarias.

---

# 11. Instalación desde GitHub Release

La release workflow actualmente genera:

```text
eng-linux-amd64
eng-linux-arm64
eng-darwin-arm64
eng-windows-amd64.exe
checksums.txt
```

Antes de escribir comandos, Pi debe comprobar que los nombres siguen coincidiendo con `.github/workflows/release.yml`.

## Linux amd64

Incluir instrucciones equivalentes a:

```bash
curl -L -o eng   https://github.com/JhonMA82/engineering-platform/releases/latest/download/eng-linux-amd64

chmod +x eng
mkdir -p ~/.local/bin
mv eng ~/.local/bin/eng

eng version
```

Explicar que `~/.local/bin` debe estar en `PATH`.

Puede mostrar `/usr/local/bin` como alternativa, pero no exigir `sudo`.

## Linux ARM64

Mostrar que debe utilizarse:

```text
eng-linux-arm64
```

## macOS Apple Silicon

Mostrar:

```text
eng-darwin-arm64
```

No afirmar soporte Intel mientras no exista `darwin-amd64`.

## Windows

Explicar:

```text
download eng-windows-amd64.exe
rename optionally to eng.exe
place in a directory included in PATH
run eng version
```

No escribir un script PowerShell complejo si no fue probado.

## Checksums

Agregar verificación opcional/recomendada mediante:

```text
checksums.txt
```

---

# 12. Build from source

Agregar sección secundaria.

Requisitos reales:

```text
Go 1.27+
Git
```

Comandos:

```bash
git clone https://github.com/JhonMA82/engineering-platform.git
cd engineering-platform
go build -o eng ./cmd/eng
./eng version
```

Opcional:

```bash
make build
```

solo si el Makefile actual lo soporta.

No poner Python como requisito.

---

# 13. Pi Agent — instalación/integración

El README debe explicar que la experiencia recomendada utiliza Pi para discovery.

No asumir que el lector conoce Pi.

Explicar en una frase:

> Pi funciona como la interfaz conversacional que convierte la idea del usuario en un `ProjectIntent`.

Después indicar:

```text
integrations/pi/
```

y enlazar a una guía dedicada si ya existe.

El README no debe meter toda la implementación del plugin.

---

# 14. Dependencia de preguntas controladas

Mencionar que la integración Pi utiliza:

```text
@juicesharp/rpiv-ask-user-question
```

para preguntas estructuradas.

No convertir esto en protagonista del README.

Debe aparecer dentro de:

```text
Pi integration
```

o documentación avanzada.

---

# 15. Quick Start — experiencia recomendada

El README debe tener:

```text
## Quick Start
```

Flujo conceptual:

```text
1. Instala `eng`
2. Configura la integración de Pi
3. Describe tu idea
4. Pi genera/refina ProjectIntent
5. Engineering Platform resuelve y materializa
6. Gentle toma control
```

Si existe un comando real `/new-project`, documentarlo.

Si depende de cómo se instala la skill de Pi, enlazar a la guía real.

No inventar una UX que el repo no implemente.

---

# 16. Uso sin Pi

Engineering Platform también debe explicarse como CLI.

Agregar un pequeño bloque:

```bash
eng resolve --input project-intent.json
eng plan --input project-intent.json
eng materialize --plan materialization-plan.json --output ./my-project
eng doctor --project ./my-project
```

Y, si continúa existiendo:

```bash
eng start --intent project-intent.json --output ./my-project
```

No hacer del flujo JSON manual la experiencia principal.

---

# 17. Explicar `eng start`

Explicar:

```text
eng start
```

como el camino CLI corto:

```text
resolve
→ plan
→ materialize
→ doctor
```

Y:

```text
--dry-run
```

como forma de inspeccionar antes de escribir.

---

# 18. Qué genera

Agregar:

```text
## What gets generated?
```

Ejemplo:

```text
my-project/
├── AGENTS.md
├── ARCHITECTURE.md
├── GENTLE.md
├── apps/
├── services/
└── .engineering/
    ├── project-intent.json
    ├── architecture-decision.json
    ├── materialization-plan.json
    ├── project-map.json
    ├── implementation-brief.md
    └── handoff.json
```

Aclarar que la estructura concreta depende de las Surfaces seleccionadas.

---

# 19. AGENTS / Gentle Handoff

Explicar en lenguaje sencillo:

```text
Engineering Platform no solo copia boilerplates.
También deja un mapa para agentes.
```

Describir:

```text
root AGENTS.md
→ indica dónde trabajar

surface AGENTS.md
→ reglas específicas de la foundation

GENTLE.md
→ explica cómo continuar

implementation-brief.md
→ resume qué se quiere construir
```

Esta característica diferenciadora debe aparecer en el README.

---

# 20. Curated Foundations

Agregar:

```text
## Curated Foundations
```

Definición sencilla:

> Un boilerplate curado es una base evaluada y registrada con un origen y pin conocidos, adapter y evidencia suficiente para que Engineering Platform pueda usarla de forma reproducible.

No afirmar que todos los entries del catálogo son estables.

---

# 21. Tabla actual del catálogo

Si se incluye una tabla:

```text
NO escribirla manualmente sin comprobar catalog/
```

Pi debe derivarla del catálogo actual.

Nunca volver a dejar referencias antiguas como:

```text
stardrive-public-web
```

si el ID canónico es:

```text
stardrive
```

Preferencia: tabla pequeña de families/Recipes principales.

---

# 22. Extending Engineering Platform

Agregar:

```text
## Extending Engineering Platform
```

Explicar:

> Si mañana necesitas una foundation para una TUI y el catálogo no tiene una adecuada, puedes evaluar un boilerplate, crear su adapter, añadir evidencia de curación y registrarlo. Si usa operaciones que el engine ya entiende, no necesitas modificar ni publicar una nueva versión del core.

Mostrar:

```text
new boilerplate
+ catalog entry
+ adapter
+ curation evidence
+ tests
```

---

# 23. Cómo agregar/modificar un boilerplate

README: versión corta.

```text
1. crear catalog/boilerplates/<id>.json
2. definir repo + immutable pin
3. declarar Surfaces/technical metadata
4. crear adapter
5. agregar curation evidence
6. ejecutar eng catalog validate
7. agregar pilot/test
```

Luego enlazar a la guía real.

---

# 24. Cómo modificar Recipes

Explicar brevemente:

```text
Recipes viven en catalog/recipes/
```

Advertencia:

> No es necesario crear una Recipe nueva por cada feature.

Ejemplos como:

```text
PDF
Excel
reports
QR
```

son normalmente features del producto, no nuevas Recipes.

---

# 25. Catálogos y overlays

Mencionar:

```text
default catalog
+
organization overlay
```

Útil para:

```text
consultoría
empresa
boilerplates privados
```

Enlazar a documentación avanzada.

---

# 26. Project evolution

Agregar sección corta.

Comandos actuales a verificar:

```text
eng surface add
eng extend
eng add
eng update
```

Explicar:

```text
surface add → agrega una nueva parte arquitectónica
extend      → activa una Surface planeada
add         → agrega un requisito de producto
update      → compara pins/reporta actualizaciones
```

---

# 27. Qué pasa si no existe una foundation adecuada

Explicar:

> Engineering Platform no fuerza una solución incorrecta. Si entiende la necesidad arquitectónica pero el catálogo no tiene una foundation compatible, devuelve un `CatalogGap`.

```text
CatalogGap
→ investigar
→ curar/agregar foundation
→ volver a resolver
```

No significa que falte una feature como PDF.

---

# 28. Product feature vs architecture

Agregar un ejemplo corto:

```text
"TUI que importa Excel y genera PDFs"
```

Engineering Platform decide:

```text
foundation TUI
```

Gentle implementa:

```text
Excel
PDF
business behavior
```

Si el boilerplate ya incluye una feature, se aprovecha.

No se busca otro boilerplate solo porque falte una feature.

---

# 29. Architecture section

La arquitectura interna puede aparecer al final.

Título:

```text
## How it works internally
```

Diagrama:

```text
ProjectIntent
    ↓
Resolver
    ↓
ArchitectureDecision
    ↓
Composer
    ↓
MaterializationPlan
    ↓
Materializer
    ↓
Development Handoff
```

Después enlazar `docs/architecture/`.

---

# 30. Repository layout

Mantener una sección breve:

```text
catalog/
cmd/
internal/
integrations/pi/
schemas/
docs/
testdata/
```

Una línea por carpeta.

---

# 31. Desarrollo/contribución

Sección:

```text
## Development
```

Comandos:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
./eng catalog validate
```

---

# 32. Release model

Explicar brevemente:

```text
core version
≠
catalog revision
```

No entrar en todos los detalles de semver en la introducción.

---

# 33. Legacy

Agregar nota pequeña:

> La implementación original en Python se conserva en `JhonMA82/engineering-platform-legacy` como referencia histórica.

No dedicarle una sección extensa.

---

# 34. Documentación

Crear un índice hacia documentación existente:

```text
Getting started
Architecture
Catalog
Adding a boilerplate
Project evolution
Release process
Pi integration
Materialization/security
```

Verificar todos los paths antes de enlazar.

---

# 35. Estructura final sugerida del README

```text
# Engineering Platform

one-line value proposition
short paragraph

## Why Engineering Platform?
## How it works
## Example
## Installation
### Linux
### macOS
### Windows
### Build from source

## Quick Start
### Recommended: Pi workflow
### CLI workflow

## What gets generated?
## Handoff to Gentle AI
## Curated foundations
## When the catalog doesn't have a fit
## Extending the catalog
## Evolving an existing project
## Commands
## How it works internally
## Repository layout
## Development
## Documentation
## Legacy
## License
```

Regla de jerarquía:

```text
product
→ install
→ use
→ extend
→ internals
```

---

# 36. Qué eliminar del README actual

Mover o eliminar del README principal:

```text
M1 — Deterministic Routing Core
M2 — Reproducible Composition
M3 — End-to-End Project Bootstrap
Fase 9
Fase 10
Out of scope for M1
```

Pueden permanecer en PRD, ADRs, history o release notes.

Ya no describen la experiencia de un producto publicado.

---

# 37. Limpiar referencias obsoletas

Buscar:

```bash
rg -n   "stardrive-public-web|M1|M2|M3|Fase 9|Fase 10|Out of scope for M1"   README.md docs/
```

No modificar ADRs históricos donde la mención sea deliberada.

Sí corregir README y guías actuales.

---

# 38. No inventar instrucciones

Antes de terminar README:

1. comprobar `.github/workflows/release.yml`;
2. comprobar assets reales de la release;
3. comprobar `eng version`;
4. probar comandos documentados;
5. verificar enlaces;
6. verificar IDs del catálogo.

Toda instrucción debe corresponder al producto real.

---

# 39. README smoke tests

Si es práctico, añadir un check ligero para:

```text
documented eng commands exist
README relative links resolve
release binary names match workflow
no obsolete provider id stardrive-public-web
```

No crear un framework complejo de documentación.

---

# 40. Release notes v1.0.1

Contenido recomendado:

```text
v1.0.1

Fixes:
- enforce must-not-use database constraints
- stop accepting unsupported deployment hard constraints

Documentation:
- completely rewritten README
- added installation and quick-start guidance
- clarified Pi → Engineering Platform → Gentle workflow
- documented catalog extension and project evolution
- removed stale milestone-era references
```

---

# 41. Validación final

Ejecutar:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./cmd/eng
./eng catalog validate
./eng version
```

Además:

```text
CI
pilots
release validation
```

---

# 42. Definition of Done — v1.0.1

## Functional

- [ ] `must-not-use database` nunca devuelve profile prohibido.
- [ ] Existe test sin alternativa disponible.
- [ ] `avoid database` sigue siendo soft.
- [ ] `deployment` no se acepta como hard constraint no evaluable.
- [ ] Existe test para `deployment`.

## README

- [ ] Explica el producto antes de la arquitectura interna.
- [ ] Una persona no técnica puede entender qué hace.
- [ ] Incluye instalación por binario.
- [ ] Incluye Linux amd64.
- [ ] Incluye Linux arm64.
- [ ] Incluye macOS arm64.
- [ ] Incluye Windows amd64.
- [ ] Incluye build from source.
- [ ] Incluye `eng version`.
- [ ] Incluye Quick Start.
- [ ] Explica Pi.
- [ ] Explica Gentle handoff.
- [ ] Explica qué genera.
- [ ] Explica curated foundations.
- [ ] Explica CatalogGap.
- [ ] Explica cómo agregar un boilerplate.
- [ ] Explica overlays brevemente.
- [ ] Explica project evolution.
- [ ] Enlaza documentación técnica.
- [ ] No contiene `stardrive-public-web` como ID vigente.
- [ ] No abre con M1/M2/M3.
- [ ] No presenta hitos internos como introducción principal.
- [ ] Todos los comandos documentados existen.
- [ ] Todos los links internos funcionan.

## Release

- [ ] `go test ./...` pasa.
- [ ] `go vet ./...` pasa.
- [ ] `eng catalog validate` pasa.
- [ ] CI pasa.
- [ ] pilots pasan.
- [ ] release notes preparadas.
- [ ] tag `v1.0.1` solo después de todos los gates.

---

# 43. Commit strategy sugerida

```text
fix(composer): enforce must-not-use database constraints
test(composer): cover forbidden database without fallback

fix(domain): reject unevaluated deployment hard constraints
test(resolver): reject unsupported deployment target

docs(readme): rewrite project introduction and installation
docs(readme): add Pi quickstart and Gentle handoff
docs(readme): document catalog extension and project evolution
docs: remove stale milestone-era references

test(docs): guard README commands and canonical provider ids
```

---

# 44. Prompt operativo para Pi Agent

```text
Prepare Engineering Platform v1.0.1 as a small patch release.

Do not redesign the architecture.

Fix exactly these functional issues:

1. A `must-not-use database=<profile>` hard constraint must never result in
   selecting the forbidden database. If no allowed profile exists, return the
   appropriate catalog/composition gap instead of falling back.

2. `deployment` is currently advertised as a supported technical-constraint
   target but is not evaluated. For v1.0.1, remove/reject it as a hard
   constraint target rather than implementing a new deployment subsystem.
   Add regression tests.

Then completely rewrite README.md as the public landing page for the project.

The README must explain the product in simple language before internal
architecture. A new visitor should quickly understand:

- what Engineering Platform does
- why it exists
- the Pi → Engineering Platform → Gentle AI workflow
- how to install the released binary
- how to use it
- what it generates
- how to add or curate boilerplates
- how catalog gaps work
- how to evolve an existing project
- where to find technical documentation

Do not lead with M1/M2/M3/Fase 9/Fase 10 implementation history.

Verify installation instructions against the actual release workflow and
release assets. The current release workflow builds:
- eng-linux-amd64
- eng-linux-arm64
- eng-darwin-arm64
- eng-windows-amd64.exe
- checksums.txt

Remove stale current-use references such as `stardrive-public-web`; use the
canonical current catalog ids.

Keep advanced resolver/composer/materializer details in a later
"How it works internally" section and link to docs for depth.

Do not make the README a duplicate of the PRD.

Before tagging v1.0.1:
- gofmt
- go vet ./...
- go test ./...
- go build ./cmd/eng
- eng catalog validate
- CI
- pilots

Produce a concise readiness report and release notes when complete.
```
