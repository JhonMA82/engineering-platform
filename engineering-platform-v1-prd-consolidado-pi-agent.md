# Engineering Platform 1.0 — PRD consolidado de reconstrucción desde cero
## Especificación normativa e instrucciones de ejecución para Pi Agent

**Estado:** Baseline aprobado para iniciar implementación  
**Fecha de referencia:** 2026-09-06  
**Objetivo de release:** `v1.0.0`  
**Implementación principal:** Go 1.27.x  
**Producto anterior de referencia:** `JhonMA82/engineering-platform` v0.x  
**Estrategia:** reconstrucción limpia; no refactor incremental de `scripts/eng.py`  
**Principio operativo:** congelar el diseño base y comenzar por el núcleo de routing antes de recuperar amplitud funcional

---

# 0. Instrucción maestra para Pi Agent

Pi Agent debe tratar este documento como la especificación principal para crear **Engineering Platform 1.0 desde cero**.

El repositorio anterior:

`https://github.com/JhonMA82/engineering-platform`

es una **fuente de requisitos, catálogo, decisiones históricas, adapters, evidencias y comportamiento esperado**, pero **NO es la arquitectura a portar**.

## Regla principal

> No traducir `eng.py` a Go.  
> No conservar la distribución de responsabilidades de `eng.py`.  
> No crear un nuevo archivo monolítico equivalente.  
> No permitir que la CLI, Pi, los prompts, el catálogo o los adapters decidan arquitectura por su cuenta.

La reconstrucción debe conservar la idea del producto, no su implementación histórica.

Pi debe trabajar por fases, cerrando cada una con código, tests, documentación y criterios de aceptación antes de avanzar.

Si una decisión técnica menor no está definida aquí, Pi puede elegir la opción más simple y mantenible, documentarla en un ADR breve y continuar. No debe detener el proyecto por decisiones reversibles.

---

# 1. Visión del producto

Engineering Platform es una plataforma de ingeniería local, minimalista y opinionada para convertir una idea de software en un proyecto reproducible construido a partir de **boilerplates curados como foundations arquitectónicas**.

La experiencia objetivo es:

```text
Idea del usuario
      │
      ▼
Pi Discovery
      │
      │ preguntas controladas cuando aplique
      ▼
ProjectIntent
      │
      ▼
Motor de decisión
      │
      ├── normalización
      ├── restricciones
      ├── candidatos
      ├── scoring
      ├── confianza
      └── explicación
      │
      ▼
ArchitectureDecision
      │
      ▼
Composer
      │
      ▼
MaterializationPlan
      │
      ▼
Materializer
      │
      ▼
Proyecto verificado
      │
      ▼
Agent Context Generation
      │
      ├── AGENTS.md raíz
      ├── project-map.json
      ├── implementation-brief.md
      └── GENTLE.md
      │
      ▼
Development Handoff
      │
      ▼
Gentle AI toma control
      │
      ├── implementación directa
      └── SDD si todavía faltan decisiones de producto
```

El producto debe evitar que una persona o un agente tenga que volver a discutir desde cero:

- framework;
- backend;
- frontend;
- mobile;
- desktop;
- TUI;
- base de datos;
- composición de aplicaciones;
- estructura de proyecto;
- bootstrap;
- quality gates;
- documentación base para agentes.

La plataforma debe distinguir claramente entre:

```text
Arquitectura / foundation
        ≠
Features concretas del producto
```

Un boilerplate no necesita implementar todas las features solicitadas. Debe ser una base óptima para implementarlas.

La filosofía se conserva:

> “Omarchy para proyectos”: pocos defaults buenos, automatización visible, decisiones reproducibles, contexto explícito para agentes y posibilidad de inspeccionar por qué se tomó cada decisión.

# 2. Problema a resolver

La implementación 0.x consiguió validar el concepto, pero acumuló responsabilidades en un único CLI Python de aproximadamente 3,600+ líneas.

Además, el resolver actual depende excesivamente de `project_type`: la definición de entrada obliga al agente a clasificar previamente el proyecto y el resolver asigna un peso dominante a esa clasificación.

Esto crea un problema conceptual:

```text
Agente clasifica arquitectura
        ↓
Resolver confirma clasificación
```

cuando debería funcionar así:

```text
Agente describe necesidad
        ↓
Resolver infiere arquitectura
```

Engineering Platform 1.0 debe corregir esta inversión.

---

# 3. Objetivos de producto

## 3.1 Objetivos principales

1. Convertir necesidades de negocio/producto en una representación estructurada llamada `ProjectIntent`.
2. Resolver una arquitectura de manera determinista y explicable.
3. Seleccionar uno o más boilerplates curados compatibles.
4. Componer diferentes superficies de producto sin multiplicar artificialmente los Golden Paths.
5. Generar un plan reproducible antes de escribir archivos.
6. Materializar únicamente fuentes versionadas/pinneadas.
7. Ejecutar setup y quality gates declarados.
8. Generar contexto mínimo y correcto para el agente que desarrollará el producto.
9. Permitir evolución posterior sin perder procedencia ni decisiones.
10. Mantener el core desacoplado de Pi, Gentle AI y cualquier otro agente.

## 3.2 Objetivos arquitectónicos

1. Binario único `eng`.
2. Core implementado en Go.
3. Dominio independiente de CLI, filesystem y procesos.
4. Catálogo declarativo.
5. Resolver puro y altamente testeable.
6. Explicaciones de selección y rechazo como datos de primera clase.
7. Tests de routing basados en escenarios reales.
8. Fronteras de paquetes explícitas.
9. Sin dependencias cíclicas.
10. Sin archivos “god object”.

---

# 4. No objetivos de v1.0

Engineering Platform 1.0 **NO** debe convertirse en:

- portal web;
- Backstage;
- servicio SaaS;
- daemon;
- servidor siempre encendido;
- marketplace abierto de boilerplates;
- framework universal;
- generador de código mediante LLM;
- sistema multiagente complejo;
- motor de workflows genérico;
- reemplazo de Gentle AI;
- reemplazo de Pi;
- sistema de deployment;
- sistema de gestión de secretos;
- sistema de infraestructura cloud;
- base de datos central de proyectos;
- motor probabilístico basado en un LLM.

La versión 1.0 debe ser una herramienta local determinista.

---

# 5. Principios no negociables

## P1. Pi describe; Engineering Platform decide

Pi puede:

- conversar;
- preguntar;
- resumir;
- normalizar lenguaje natural a un contrato estructurado;
- preservar restricciones técnicas explícitas del usuario;
- pedir confirmación al usuario.

Pi no puede decidir por su cuenta:

- Golden Path / Recipe;
- boilerplate final;
- arquitectura multi-app;
- base de datos definitiva;
- surface provider;
- score;
- compatibilidad.

Estas decisiones pertenecen al core Go y al catálogo.

---

## P2. `project_type` deja de ser input autoritativo

`project_type` no debe existir como campo obligatorio de `ProjectIntent`.

Si se conserva algún concepto parecido, debe aparecer únicamente como:

- dato derivado;
- etiqueta explicativa;
- resultado del resolver.

Nunca debe ser la llave principal de selección.

---

## P3. El router debe poder decir “no sé todavía”

El resolver no está obligado a seleccionar un candidato cuando la evidencia es insuficiente.

Estados mínimos:

```text
resolved
ambiguous
catalog-gap
unsupported
invalid
```

`catalog-gap` significa que la necesidad es arquitectónicamente comprensible, pero el catálogo activo no contiene una foundation, Surface, provider o profile capaz de satisfacer un requisito arquitectónico obligatorio.

No usar `catalog-gap` por una feature de negocio faltante.

---

## P4. Toda decisión debe ser explicable

No se acepta un resultado como:

```json
{"recipe": "GP-06"}
```

Debe existir una explicación estructurada:

```json
{
  "selected": "GP-06",
  "score": 91,
  "confidence": "high",
  "reasons": [
    "requires admin surface",
    "requires public unauthenticated intake",
    "requires shared API"
  ],
  "rejected": [
    {
      "id": "GP-02",
      "score": 67,
      "reasons": [
        "cannot satisfy required public-intake composition"
      ]
    }
  ]
}
```

---

## P5. El conocimiento del catálogo debe vivir en datos

No hardcodear en Go:

- lista de Surfaces;
- aliases;
- capabilities arquitectónicas;
- qué boilerplate provee una Surface;
- Recipes;
- database profiles;
- feature metadata de boilerplates;
- maintenance tiers;
- estado de una opción;
- combinaciones validadas.

El código conoce **cómo evaluar contratos**.

El catálogo conoce **qué opciones existen**.

---

## P6. Separar decisión de efectos secundarios

El resolver nunca:

- clona;
- crea carpetas;
- ejecuta Git;
- instala dependencias;
- escribe manifests;
- ejecuta shell.

Debe poder probarse usando únicamente estructuras en memoria.

---

## P7. Dry-run antes de mutación

Toda operación compleja debe poder producir primero un plan.

Ejemplo:

```text
eng resolve intent.json
eng plan intent.json
eng materialize plan.json
```

Los comandos de conveniencia pueden encadenar fases, pero internamente estas responsabilidades deben permanecer separadas.

---

## P8. El catálogo no es código de aplicación

Agregar correctamente:

- un nuevo boilerplate;
- una nueva Surface, como `tui`;
- un nuevo database profile;
- una capability arquitectónica;
- una Recipe;
- un adapter;
- una combinación validada;

no debe requerir modificar ni recompilar el resolver si los contratos existentes ya pueden representarlo.

---

## P9. Compatibilidad demostrada, no asumida

Que dos tecnologías puedan convivir teóricamente no significa que Engineering Platform pueda componerlas.

La compatibilidad materializable debe estar declarada y probada.

---

## P10. Mantener el producto pequeño

Antes de agregar una abstracción preguntar:

> ¿Esta abstracción elimina una responsabilidad mezclada real o solo añade arquitectura ceremonial?

Evitar enterprise patterns sin necesidad.

---

## P11. Boilerplates = foundations, no feature bundles

Los boilerplates son bases arquitectónicas mínimas, óptimas y escalables.

No deben seleccionarse exigiendo que ya implementen todas las features del usuario.

Ejemplo:

```text
Usuario:
TUI para Excel + comparación + PDF

Engineering Platform:
selecciona una foundation TUI/Python correcta

Gentle:
implementa comparación, PDF y demás features faltantes
```

Si el boilerplate ya trae Excel, PDF u otra feature útil, se aprovecha.

La ausencia de una feature concreta **no lo descalifica** cuando la arquitectura es correcta.

---

## P12. Las features existentes son bonus, no requisitos de elegibilidad

Entre dos foundations con ajuste arquitectónico equivalente, una que ya incluya features solicitadas puede recibir un bonus razonable.

Pero el algoritmo nunca debe perseguir un “boilerplate perfecto” para cada combinación de features.

Regla:

> Prefer existing capability, never require feature completeness.

---

## P13. Catalog gaps son gaps arquitectónicos

Sí son gaps:

```text
se requiere TUI y no existe ninguna foundation TUI
se requiere native mobile y el catálogo solo tiene web
se requiere una topología shared-backend que ninguna composición soporta
```

No son gaps:

```text
falta PDF
falta Excel
falta facturación
falta un reporte
falta un formulario concreto
```

Estas últimas son tareas de implementación para Gentle.

---

## P14. Usuario técnico: restricciones explícitas se preservan

Engineering Platform debe servir tanto a usuarios no técnicos como técnicos.

Ejemplo no técnico:

> “Necesito que el dashboard siga funcionando sin Internet.”

El resolver decide tecnología adecuada.

Ejemplo técnico:

> “Quiero TanStack con datos offline y usar Turso.”

El Intent conserva TanStack/Turso como restricción o preferencia explícita, según el lenguaje del usuario.

El resolver debe respetarla si es compatible y curada; si existe una contradicción real debe explicarla, no ignorarla.

---

## P15. Evolución del catálogo != release del core

Regla normativa:

> Nuevos datos y conocimiento → actualización de catálogo.  
> Nueva capacidad del motor → release del core.

Agregar una nueva TUI foundation, un nuevo framework o una nueva base de datos no debe requerir `eng v1.0.1` si el motor ya sabe representarlos.

Sí puede requerir core nuevo una operación fundamental inexistente, por ejemplo un nuevo tipo de transformación materializable que el engine aún no entiende.

---

## P16. El proyecto generado debe ser navegable por agentes sin discovery del repositorio

Engineering Platform ya conoce qué materializó.

Por ello debe generar un mapa de contexto que permita a Gentle ir directamente al lugar correcto.

Un agente no debería explorar cientos de archivos para descubrir:

- qué apps existen;
- cuál es el backend;
- dónde vive mobile;
- qué boilerplate originó cada componente;
- qué `AGENTS.md` local debe leer.

---

## P17. Gentle recibe estado consolidado, no historial conversacional

El handoff contiene:

- intención final;
- arquitectura final;
- mapa del proyecto;
- requirements;
- decisiones bloqueadas;
- preguntas abiertas.

No debe pasar toda la conversación de Pi como fuente principal.

El historial puede contener ideas descartadas y ruido que no debe contaminar la implementación.

---

## P18. Gentle decide profundidad de implementación

Después del handoff:

```text
contexto suficiente
→ implementación directa

faltan decisiones de producto
→ sesión SDD
```

Engineering Platform no debe imponer SDD siempre.

Si Gentle abre SDD, sus preguntas deben profundizar reglas de producto, no redescubrir arquitectura ya resuelta.

# 6. Arquitectura objetivo

La arquitectura deberá seguir un modelo de **core por dominios + adapters externos**, sin mezclar transporte, decisión y efectos.

```text
┌─────────────────────────────────────────────────────────────┐
│                         CLI / Pi                            │
│                     Adaptadores externos                    │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Services                     │
│   ResolveProject / PlanProject / Materialize / Doctor       │
└───────────────┬─────────────────────┬───────────────────────┘
                │                     │
                ▼                     ▼
┌────────────────────────┐   ┌───────────────────────────────┐
│       Domain/Core      │   │          Catalog             │
│                        │   │                               │
│ Intent                 │   │ Recipes                       │
│ Resolver               │◄──│ Boilerplates                  │
│ Scoring                │   │ Surfaces                      │
│ Constraints            │   │ Capabilities                  │
│ Composer               │   │ Databases                     │
│ Decisions              │   │ Feature packs                 │
└───────────────┬────────┘   └───────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure                          │
│ Git / Filesystem / Process / Runtime / Materialization      │
└─────────────────────────────────────────────────────────────┘
```

---

# 7. Estructura inicial del repositorio

Pi debe iniciar con una estructura cercana a:

```text
engineering-platform/
├── cmd/
│   └── eng/
│       └── main.go
│
├── internal/
│   ├── domain/
│   │   ├── intent.go
│   │   ├── requirement.go
│   │   ├── constraint.go
│   │   ├── surface.go
│   │   ├── recipe.go
│   │   ├── boilerplate.go
│   │   ├── database.go
│   │   ├── decision.go
│   │   └── errors.go
│   │
│   ├── catalog/
│   │   ├── catalog.go
│   │   ├── source.go
│   │   ├── loader.go
│   │   ├── overlay.go
│   │   ├── validator.go
│   │   └── index.go
│   │
│   ├── resolver/
│   │   ├── normalizer.go
│   │   ├── requirements.go
│   │   ├── constraints.go
│   │   ├── candidates.go
│   │   ├── scoring.go
│   │   ├── coverage.go
│   │   ├── confidence.go
│   │   ├── gaps.go
│   │   ├── questions.go
│   │   └── resolver.go
│   │
│   ├── composer/
│   │   ├── composer.go
│   │   ├── surfaces.go
│   │   ├── providers.go
│   │   ├── compatibility.go
│   │   └── destinations.go
│   │
│   ├── planner/
│   │   ├── planner.go
│   │   └── plan.go
│   │
│   ├── materializer/
│   │   ├── materializer.go
│   │   ├── git.go
│   │   ├── filesystem.go
│   │   ├── process.go
│   │   └── checks.go
│   │
│   ├── project/
│   │   ├── manifest.go
│   │   ├── definition.go
│   │   ├── provenance.go
│   │   ├── doctor.go
│   │   └── projectmap.go
│   │
│   ├── handoff/
│   │   ├── generator.go
│   │   ├── brief.go
│   │   ├── agents.go
│   │   ├── gentle.go
│   │   └── status.go
│   │
│   ├── app/
│   │   ├── resolve.go
│   │   ├── plan.go
│   │   ├── materialize.go
│   │   ├── handoff.go
│   │   └── evolve.go
│   │
│   └── cli/
│       ├── root.go
│       ├── resolve.go
│       ├── explain.go
│       ├── plan.go
│       ├── materialize.go
│       ├── start.go
│       ├── catalog.go
│       ├── doctor.go
│       └── version.go
│
├── catalog/
│   ├── metadata.json
│   ├── recipes/
│   ├── boilerplates/
│   ├── surfaces/
│   ├── capabilities/
│   ├── vocabulary/
│   ├── database-profiles/
│   ├── feature-packs/
│   ├── compatibilities/
│   ├── adapters/
│   └── curation/
│
├── schemas/
│   ├── project-intent.schema.json
│   ├── architecture-decision.schema.json
│   ├── catalog-gap.schema.json
│   ├── materialization-plan.schema.json
│   ├── project-map.schema.json
│   ├── development-handoff.schema.json
│   └── project-manifest.schema.json
│
├── integrations/
│   └── pi/
│       ├── package.json
│       ├── extensions/
│       ├── skills/
│       └── prompts/
│
├── templates/
│   ├── ARCHITECTURE.md.tmpl
│   ├── ROOT_AGENTS.md.tmpl
│   ├── IMPLEMENTATION_BRIEF.md.tmpl
│   └── GENTLE.md.tmpl
│
├── testdata/
│   ├── routing/
│   ├── composition/
│   ├── handoff/
│   └── catalogs/
│
├── tests/
│   └── e2e/
│
├── docs/
│   ├── architecture/
│   ├── concepts/
│   ├── guides/
│   └── adr/
│
├── .github/
│   └── workflows/
│
├── go.mod
├── go.sum
├── Makefile
├── AGENTS.md
├── README.md
├── CONTRIBUTING.md
└── LICENSE
```

La estructura puede simplificarse si Pi encuentra una razón concreta, pero debe respetar las fronteras.

## Catálogo físicamente separado del core

El layout debe dejar clara la diferencia entre:

```text
motor Go
≠
conocimiento del catálogo
```

El catálogo puede vivir inicialmente en el mismo repo por conveniencia de desarrollo, pero debe cargarse mediante una frontera que posteriormente permita:

```text
built-in/default catalog
organization catalog overlay
user catalog overlay
project-local overrides
```

sin recompilar el binario.

No implementar un sistema complejo de plugins dinámicos para lograrlo.

JSON/YAML + schemas + adapters declarativos son suficientes como baseline.

# 8. Dependencias entre paquetes

Dirección permitida:

```text
cli ──────────────► app
                    │
                    ├────► resolver ───► domain
                    ├────► composer ───► domain
                    ├────► planner ─────► domain
                    ├────► materializer
                    ├────► project
                    └────► catalog ─────► domain
```

Reglas:

```text
domain       → no importa infraestructura
resolver     → no importa cli
resolver     → no importa materializer
composer     → no importa cli
catalog      → no importa cli
materializer → puede usar domain/plan, no resolver decisiones
cli          → no contiene lógica de negocio
pi           → solo invoca contratos públicos/CLI
```

No introducir un paquete genérico `utils` para ocultar responsabilidades.

---

# 9. Modelo de dominio

## 9.1 ProjectIntent

`ProjectIntent` representa **lo que el producto necesita**, no cómo se implementará.

Debe separar al menos cuatro clases de información:

```text
ProjectIntent
├── Product Requirements
├── Architecture Requirements
├── Technical Constraints / Preferences
└── Lifecycle Scope
```

Propuesta conceptual:

```go
type ProjectIntent struct {
    SchemaVersion int
    Name          string

    Problem string
    Users   []Actor
    Flows   []ProductFlow

    ProductRequirements      []ProductRequirement
    ArchitectureRequirements []ArchitectureRequirement

    Surfaces []SurfaceIntent
    Data     DataIntent
    Ops      OperationalIntent

    TechnicalConstraints []TechnicalConstraint
    Preferences          []Preference

    Scope ScopeIntent
    Notes []string
}
```

No es obligatorio usar exactamente esta estructura, pero sí conservar la separación semántica.

---

## 9.2 Product Requirements

Describen **qué debe hacer el producto**.

Ejemplos:

```text
importar archivos Excel
generar reportes PDF
comparar libros
gestionar clientes
emitir facturas
crear códigos QR
aprobar solicitudes
buscar registros
```

Estas requirements:

- deben conservarse;
- deben llegar a Gentle;
- pueden dar bonus a un boilerplate que ya las incluya;
- normalmente NO son requisitos de elegibilidad del boilerplate.

No crear una capability arquitectónica por cada feature de producto.

---

## 9.3 Architecture Requirements

Describen propiedades que sí condicionan la foundation.

Ejemplos:

```text
terminal user interface
native mobile
desktop-native
offline-first
shared backend
anonymous public access
local filesystem access
background processing
realtime
edge execution
multiple clients consuming same API
```

Estas requirements sí pueden:

- eliminar candidatos;
- cambiar Recipe;
- exigir otra Surface;
- exigir otro database profile;
- generar un CatalogGap.

---

## 9.4 Technical Constraints y Preferences

El usuario puede expresar decisiones técnicas.

Ejemplos:

```text
must-use: tanstack
must-use: turso
prefer: python
prefer: bun
avoid: nextjs
```

Distinguir:

```text
constraint
→ obligatoria salvo contradicción imposible

preference
→ influye en ranking
```

Engineering Platform no debe tratar a un usuario técnico como si no hubiera expresado una decisión.

Tampoco debe convertir una preferencia casual en obligación.

---

## 9.5 Actor

Ejemplos:

```text
anonymous-public
customer
employee
operator
administrator
field-worker
partner
developer
```

Los actores son señales de necesidades, no roles de autorización definitivos.

---

## 9.6 SurfaceIntent

Debe expresar interfaces/clientes requeridos.

Ejemplo:

```json
{
  "kind": "web-admin",
  "access": "authenticated"
}
```

Surfaces iniciales pueden incluir:

```text
public-web
web-admin
public-intake
mobile-native
desktop
api
worker
tui
```

Pero esta lista es **catálogo**, no enum rígido compilado en Go.

Mañana debe ser posible registrar:

```text
tui
voice-client
kiosk
extension
```

si el modelo existente puede representarlos, sin release del núcleo.

Importante:

- `landing` normalmente es product/site scope sobre `public-web`;
- `blog` normalmente es feature/content capability;
- `PDF report generation` es feature;
- `Excel import` es feature;
- `offline` es architecture requirement/capability;
- `TUI` sí es Surface.

---

## 9.7 DataIntent

Debe capturar necesidad antes que marca.

Ejemplos de dimensiones:

```text
persistence: none | local | shared
multi_user: true/false
relational: true/false/unknown
offline_sync: true/false
edge_replication: true/false
attachments: true/false
sensitive_data: true/false
expected_scale: small | medium | large | unknown
```

Si el usuario no fijó tecnología, el resolver decide después el `database-profile`.

Si el usuario dijo explícitamente “Turso”, debe preservarse como constraint/preference técnica y validarse contra el resto del Intent.

---

## 9.8 OperationalIntent

Ejemplos:

```text
background-jobs
webhooks
realtime
scheduled-tasks
offline-operation
push-notifications
file-processing
external-integrations
```

Solo promover una operación a requisito arquitectónico cuando realmente cambie la foundation o composición.

---

## 9.9 Constraint

Las restricciones deben distinguir:

### Hard constraints

Un candidato incompatible debe eliminarse.

Ejemplos:

```text
must-run-offline
must-have-native-mobile
must-support-anonymous-public-input
must-share-backend-across-clients
must-use-tanstack
must-use-existing-corporate-database
```

### Soft preferences

Afectan score, pero no eliminan.

Ejemplos:

```text
prefer-python
prefer-single-repository
prefer-edge
prefer-low-ops
prefer-existing-team-skill
```

Nunca tratar una preferencia como hard constraint accidentalmente.

---

## 9.10 ScopeIntent: ahora vs futuro

Cada requirement/surface relevante debe poder distinguir:

```text
required_now
planned_later
explicitly_excluded
```

Una app mobile futura no debe materializar Ignite hoy.

Sí puede influir en decisiones que convenga hacer compatibles con una fase posterior, por ejemplo una API reusable.

---

## 9.11 FoundationFit vs FeatureCoverage

No confundir:

```text
Architectural/Foundation Fit
```

con:

```text
Product Feature Coverage
```

Para seleccionar un boilerplate importa principalmente el primero.

Ejemplo:

```text
TUI Python foundation
Architectural fit: excellent

Features already present:
- Excel parsing

Features to implement:
- PDF reports
- comparison workflow
- email export
```

Resultado válido:

```text
resolved
```

No `catalog-gap`.

# 10. Vocabulary y normalización

El vocabulario no debe estar embebido en código.

Crear dentro del catálogo:

```text
catalog/vocabulary/
catalog/capabilities/
catalog/surfaces/
```

El normalizador debe transformar aliases a términos canónicos.

Ejemplos:

```text
panel          → web-admin
app de campo   → mobile-native + field-operations
terminal app   → tui
sin internet   → offline-operation
```

La normalización debe ser determinista.

## Regla de extensibilidad

El core no debe declarar un enum cerrado para cada Surface/capability posible.

Los IDs pueden modelarse como tipos nominales/string validados por catálogo.

Ejemplo conceptual:

```go
type SurfaceID string
type CapabilityID string
```

El catálogo determina qué IDs existen en la revisión activa.

Esto permite agregar una Surface `tui` sin modificar `resolver.go`.

## Features de producto no deben contaminar el vocabulario arquitectónico

No llenar `capabilities` con toda feature imaginable:

```text
pdf-generation
excel-read
invoice-export
send-whatsapp
...
```

salvo que una de ellas demuestre tener consecuencias arquitectónicas reales y generalizables.

Las features de negocio permanecen en `ProductRequirements`.

## Términos desconocidos

Si aparece un término desconocido:

```text
unknown vocabulary
```

o:

```text
unclassified product requirement
```

según corresponda.

No asumir silenciosamente significado.

Pi puede conservar texto libre cuando no sea necesario canonizar una feature para el routing.

# 11. Architecture Profiles / Recipes

Los Golden Paths pueden mantenerse conceptualmente, pero en v1 deben ser **resultados candidatos**, no categorías proporcionadas por el usuario.

Nombres iniciales derivados del catálogo actual:

```text
GP-01 Public Web
GP-02 Admin Application
GP-03 Python/Data Application
GP-04 Mobile Application
GP-05 Desktop Application
GP-06 Multi-App
GP-07 Commercial SaaS
```

Antes de copiarlos, revisar sus contratos y simplificarlos.

Una Recipe debe declarar como mínimo:

```json
{
  "id": "GP-06",
  "version": "1.0.0",
  "status": "stable",

  "provides": {
    "surfaces": [],
    "capabilities": []
  },

  "requires": {},
  "supports": {},
  "forbids": {},

  "primary_boilerplates": [],
  "allowed_surface_composition": [],

  "database_policy": {},
  "feature_policy": {},
  "quality_gates": []
}
```

No incluir un campo del tipo:

```json
"project_types": ["multi-app"]
```

como mecanismo principal de matching.

---

# 12. Pipeline del resolver

El resolver debe ejecutarse en fases explícitas.

## Fase R1 — Validar Intent

Verificar:

- schema;
- vocabulario arquitectónico;
- duplicados;
- contradicciones;
- required fields mínimos;
- constraints técnicas explícitas.

Resultado:

```text
valid | invalid
```

---

## Fase R2 — Normalizar

Transformar:

- aliases;
- términos equivalentes;
- orden estable;
- defaults puramente semánticos.

No seleccionar tecnología.

---

## Fase R3 — Clasificar requirements

Separar:

```text
Product Requirements
Architecture Requirements
Technical Constraints
Preferences
```

No permitir que “PDF”, “Excel” o una feature concreta se convierta automáticamente en hard constraint.

---

## Fase R4 — Derivar requirements arquitectónicos

Convertir Intent a requisitos canónicos.

Ejemplo:

```text
anonymous user
+ form flow
→ requires public-access surface
```

Otro:

```text
admin web
+ native mobile
+ same business data
→ shared-backend requirement
```

Estas reglas deben estar visibles y probadas.

---

## Fase R5 — Generar candidatos

Evaluar todas las Recipes/providers activas y compatibles con la revisión del catálogo.

No empezar desde un supuesto `project_type`.

Las restricciones técnicas explícitas pueden reducir candidatos.

Ejemplo:

```text
must-use tanstack
```

debe excluir foundations incompatibles, pero no debe impedir componer otros componentes necesarios.

---

## Fase R6 — Aplicar hard constraints arquitectónicas

Cada candidato debe producir:

```text
eligible
rejected
```

con razones.

Ejemplo:

```text
GP-02 rejected:
- required surface mobile-native cannot be composed
```

No rechazar:

```text
python-tui
```

porque no traiga PDF si PDF es una feature que Gentle puede implementar.

---

## Fase R7 — Verificar cobertura arquitectónica del catálogo

Antes de declarar `resolved`, verificar que exista foundation/composición para todos los requisitos arquitectónicos obligatorios.

Si falta una pieza:

```text
status: catalog-gap
```

Ejemplo:

```text
required:
- tui

active catalog:
- no TUI provider

catalog gap:
- missing architectural foundation for surface=tui
```

El gap debe producir criterios de investigación/curación, no buscar automáticamente cualquier repo y confiar en él.

---

## Fase R8 — Scoring

Solo los candidatos elegibles reciben score.

Propuesta de dimensiones:

```text
required surface coverage      0..35
architectural capability fit   0..20
data fit                       0..15
operational fit                0..10
curation/stability             0..8
composition simplicity         0..7
technical preference fit       0..5
-----------------------------------
total                          0..100
```

### Bonus limitado por features ya presentes

Opcionalmente, cuando el catálogo tenga metadata confiable:

```text
requested product feature already provided
```

puede funcionar como tie-breaker o bonus pequeño.

Nunca debe pesar más que el ajuste arquitectónico.

Los pesos deben poder configurarse centralmente, preferiblemente en policy data.

---

## Fase R9 — Confidence

Determinar confianza por:

- score;
- margen con #2;
- dimensiones críticas desconocidas;
- constraints no resueltas;
- cobertura arquitectónica.

---

## Fase R10 — Output

Estados:

```text
resolved
ambiguous
catalog-gap
unsupported
invalid
```

`resolved` significa:

> existe una foundation/composición curada que satisface los requisitos arquitectónicos obligatorios.

No significa:

> todas las features del producto ya vienen implementadas.

# 13. Confidence model

La confianza debe ser derivada del resultado, no inventada por el agente.

Propuesta:

```text
high:
- score ganador >= threshold
- diferencia con #2 >= margin
- no unresolved critical dimensions

medium:
- ganador claro pero existe una dimensión relevante desconocida

low:
- diferencia pequeña
- dos arquitecturas plausibles
```

Estados:

```go
type ResolutionStatus string

const (
    ResolutionResolved    ResolutionStatus = "resolved"
    ResolutionAmbiguous   ResolutionStatus = "ambiguous"
    ResolutionUnsupported ResolutionStatus = "unsupported"
    ResolutionInvalid     ResolutionStatus = "invalid"
)
```

---

# 14. Preguntas discriminantes

Cuando el resultado sea `ambiguous`, Engineering Platform no debe redactar una pregunta natural compleja.

Debe devolver una estructura neutral:

```json
{
  "dimension": "public_access",
  "reason": "distinguishes GP-02 from GP-06",
  "options": [
    "authenticated-only",
    "anonymous-public-access"
  ]
}
```

Pi transforma esa dimensión en lenguaje natural.

Ejemplo:

> ¿El formulario debe poder ser utilizado por personas que no tengan una cuenta?

Así se conserva la frontera:

```text
core descubre qué falta
Pi decide cómo preguntarlo
```

---

# 15. ArchitectureDecision

El output del resolver debe ser persistible.

Contrato conceptual:

```json
{
  "schema_version": 1,
  "status": "resolved",
  "intent_fingerprint": "...",

  "selected": {
    "recipe": "GP-06",
    "recipe_version": "1.0.0",
    "score": 92
  },

  "confidence": {
    "level": "high",
    "margin": 21
  },

  "reasons": [],

  "candidates": [
    {
      "recipe": "GP-06",
      "eligible": true,
      "score": 92,
      "positive_reasons": [],
      "negative_reasons": []
    }
  ],

  "derived_requirements": [],
  "unresolved_dimensions": []
}
```

Debe ser estable para:

- tests;
- debug;
- auditoría;
- handoff;
- `eng explain`.

---

# 16. Composer

El resolver decide la **familia/Recipe y foundation fit**.

El Composer decide **cómo satisfacer las Surfaces requeridas** utilizando providers compatibles.

No mezclar ambas responsabilidades.

Ejemplo:

```text
Intent:
API + Dashboard + Mobile

Composer:
services/api
apps/dashboard
apps/mobile
```

Cada componente puede provenir de un boilerplate diferente.

El Composer debe verificar:

- provider disponible;
- provider curado;
- pin válido;
- Surface cubierta;
- architectural capabilities cubiertas;
- dependencias;
- colisiones de destino;
- compatibilidad entre componentes;
- database compatibility;
- runtime compatibility.

## Features del producto

El Composer no busca componentes adicionales solo porque una feature concreta no venga incluida.

Ejemplo:

```text
dashboard foundation seleccionada
usuario requiere PDF reports
boilerplate no incluye PDF
```

Resultado:

```text
composición válida
feature queda como implementation requirement
```

No:

```text
buscar otro dashboard hasta encontrar uno con PDF
```

## Technical constraints

Si el usuario dijo:

```text
dashboard con TanStack
offline
Turso
```

el Composer/resolver deben usar estas señales para:

- seleccionar provider TanStack;
- elegir perfil offline compatible;
- validar Turso/libSQL si está curado y es compatible;
- explicar cualquier incompatibilidad.

No deben sustituir tecnologías explícitamente requeridas sin informar.

# 17. Boilerplate Registry

Mantener el concepto de catálogo curado del proyecto actual.

Cada boilerplate es una **foundation**, no una promesa de feature completeness.

Cada entrada debe separar:

## Decisión

```text
default
alternative
specialized
experimental
reference
deprecated
rejected
```

## Entrega

```text
catalog-only
pilot-ready
curated
released
```

Estos ejes deben conservarse.

## Metadata útil

Un boilerplate puede declarar features ya incluidas:

```text
included_product_features
```

pero esta metadata es informativa/de ranking secundario.

No convertirla en requisito de selección salvo que una feature esté clasificada explícitamente como requirement arquitectónico.

## Nuevas foundations

Agregar un boilerplate nuevo debe consistir principalmente en:

```text
catalog entry
+ adapter
+ curation evidence
+ tests
```

Ejemplo futuro:

```text
python-tui
```

Después de validarlo, el resolver debe poder seleccionarlo para proyectos TUI sin que se modifique el core Go.

# 18. Contrato de Boilerplate

Cada boilerplate debe declarar al menos:

```text
id
display_name
repository
license
upstream pin
observed_at
decision_status
delivery_status
maintenance_tier
provides_surfaces
architectural_capabilities
technology_metadata
included_product_features (optional)
requires
integration adapter
curation evidence
update strategy
agent instructions location
```

Un boilerplate no es seleccionable para materialización si falta:

- repository/local source;
- pin;
- adapter;
- evidencia mínima;
- estado de entrega permitido.

## Agent instructions

Toda foundation curada debe declarar dónde viven sus instrucciones especializadas para agentes.

Preferencia:

```text
AGENTS.md
```

en la raíz del boilerplate/surface materializada.

Estas instrucciones describen:

- arquitectura local;
- patrones;
- convenciones;
- commands;
- testing;
- boundaries;
- cosas que el agente no debe hacer.

Engineering Platform no debe duplicar este contenido en el `AGENTS.md` raíz del proyecto.

# 19. Adapters

Los adapters contienen detalles de materialización.

Un adapter puede declarar:

```text
source
destination
prune paths
environment
setup commands
check commands
runtime requirements
managed files
extension destination
dependencies
```

El adapter NO declara:

> “úsame cuando el proyecto sea un CRM”.

Eso pertenece al catálogo/Recipe.

---

# 20. MaterializationPlan

Antes de escribir código debe existir un plan serializable.

Ejemplo conceptual:

```json
{
  "schema_version": 1,
  "project": "school-requests",
  "recipe": "GP-06",

  "components": [
    {
      "boilerplate": "hono-api",
      "pin": "...",
      "destination": "services/api"
    },
    {
      "boilerplate": "tanstack-admin",
      "pin": "...",
      "destination": "apps/admin"
    }
  ],

  "database_profile": "postgresql-managed",

  "operations": [],
  "setup": [],
  "checks": []
}
```

`eng plan` debe poder generar este documento sin alterar filesystem.

---

# 21. Materializer

Solo consume un `MaterializationPlan` válido.

Responsabilidades:

1. verificar destino;
2. crear staging;
3. obtener fuentes;
4. comprobar pin;
5. aplicar adapter;
6. validar paths seguros;
7. detectar colisiones;
8. mover resultado;
9. ejecutar setup permitido;
10. ejecutar checks;
11. escribir procedencia;
12. generar manifest.

No debe volver a resolver arquitectura.

---

# 22. Seguridad de filesystem y procesos

Reglas mínimas:

- nunca escribir fuera del proyecto por path traversal;
- rechazar destinos absolutos cuando no estén explícitamente permitidos;
- no seguir symlinks peligrosos durante composición;
- no ejecutar comandos provenientes del input del usuario;
- comandos ejecutables solo desde adapters curados;
- environment allowlist;
- logs sin secretos;
- staging antes de commit final;
- operación idempotente cuando sea razonable;
- cleanup en fallo.

Agregar tests específicos para:

```text
../
absolute paths
nested destinations
destination collisions
symlink escape
malformed adapter
command injection
```

---

# 23. CLI v1

La CLI debe ser una capa delgada.

Comandos mínimos iniciales:

```text
eng version

eng catalog
eng catalog show <id>
eng catalog validate

eng resolve --input intent.json
eng explain --input decision.json

eng plan --input intent.json
eng materialize --plan plan.json --output <dir>

eng start <name>
eng doctor [--project <dir>]
```

Después de cerrar el núcleo pueden incorporarse:

```text
eng add
eng surface add
eng extend
eng update
eng handoff
eng install
eng uninstall
```

No implementar todos los comandos antiguos durante la primera fase.

Primero demostrar:

```text
Intent → Resolve → Plan → Materialize
```

---

# 24. Librería CLI

Se permite usar Cobra para la capa CLI si reduce código repetitivo.

Condición:

> Cobra solo puede vivir en `internal/cli` / `cmd/eng`.

El dominio y application services no deben recibir:

- `cobra.Command`;
- flags;
- argumentos CLI;
- stdout.

Si Pi considera suficiente `flag`/stdlib, también es válido.

No añadir un framework adicional sin necesidad.

---

# 25. Runtime Go

Usar como baseline:

```text
Go 1.27.x
```

Durante la creación inicial fijar la versión estable 1.27 disponible en ese momento.

Priorizar:

- standard library;
- pocas dependencias;
- dependencias justificadas;
- `context.Context` en operaciones externas;
- errores tipados donde aporten valor;
- interfaces únicamente en fronteras reales.

Evitar interfaces “por si acaso”.

---

# 26. Integración con Pi

Pi debe ser un **adapter conversacional**, no el cerebro del core.

## Proveedor oficial de preguntas controladas

La integración oficial debe utilizar:

```text
@juicesharp/rpiv-ask-user-question
```

y su herramienta:

```text
ask_user_question
```

cuando una pregunta pueda presentarse como opciones controladas.

Esta dependencia pertenece a:

```text
integrations/pi/
```

No al core Go.

El core debe devolver dimensiones faltantes/ambiguas como datos.

Pi decide cómo presentarlas.

---

## Flujo deseado

```text
/new-project
    ↓
skill project-discovery
    ↓
preguntas progresivas
    │
    ├── ask_user_question cuando opciones conocidas ayudan
    └── pregunta libre cuando se necesita descripción
    ↓
project-intent.json
    ↓
eng resolve
    ↓
status ?
    ├── resolved
    │      ↓
    │   presentar propuesta
    │
    ├── ambiguous
    │      ↓
    │   unresolved_dimensions
    │      ↓
    │   ask_user_question
    │      ↓
    │   actualizar Intent
    │      ↓
    │   eng resolve
    │
    └── catalog-gap
           ↓
        explicar gap arquitectónico
           ↓
        recomendar investigar/curar foundation
    ↓
confirmación del usuario
    ↓
eng plan
    ↓
eng materialize
    ↓
eng handoff
    ↓
Gentle toma control
```

Pi no debe escribir un `recipe_id` manualmente.

Pi no debe modificar scores.

Pi no debe inventar compatibilidades.

Pi sí debe preservar restricciones explícitas del usuario.

---

## Reglas para `ask_user_question`

Preferir preguntas controladas cuando:

- existen opciones claras;
- se necesita distinguir candidatos;
- la respuesta puede mapearse a un campo canónico;
- varias preguntas relacionadas pueden agruparse sin cansar al usuario.

Usar respuesta libre cuando:

- se describe el problema;
- el flujo requiere contexto narrativo;
- las alternativas conocidas podrían sesgar la respuesta.

Reglas:

```text
- no preguntar información ya existente en ProjectIntent
- no repetir preguntas resueltas
- no convertir discovery en formulario rígido
- agrupar preguntas relacionadas cuando mejore fluidez
- conservar opción libre cuando las opciones puedan ser incompletas
- priorizar preguntas discriminantes producidas por el resolver
```

El core nunca debe llamar directamente a `ask_user_question`.

# 27. Diseño del cuestionario de descubrimiento

El cuestionario debe ser progresivo, no un formulario rígido.

Prioridad aproximada:

1. problema;
2. usuarios;
3. flujos principales;
4. interfaces/Surfaces necesarias;
5. acceso público/autenticado;
6. datos;
7. offline/native/local;
8. integraciones;
9. restricciones relevantes;
10. preferencias técnicas explícitas.

## Usuario no técnico

No preguntar:

```text
React o Vue
PostgreSQL o Turso
Go o Python
monorepo o multi-repo
```

si la necesidad se puede inferir de requisitos.

## Usuario técnico

Si el usuario espontáneamente define:

```text
TanStack
Turso
Python
Go
Ignite
```

Pi debe registrar esa decisión como constraint/preference, no descartarla.

Ejemplo:

> “Voy a desarrollar un dashboard usando TanStack y necesito datos offline.”

No es necesario volver a preguntar qué framework quiere.

La siguiente pregunta debe centrarse en información faltante que realmente cambie la arquitectura.

## Preguntas discriminantes

Cuando el resolver entregue:

```text
unresolved_dimensions
```

Pi debe utilizar preferentemente `ask_user_question`.

Ejemplo:

```text
dimension: public_access
options:
- authenticated-only
- anonymous-public-access
```

Pi puede renderizar:

> ¿Quién podrá utilizar esta parte?

con opciones controladas.

El core identifica **qué falta**.

Pi decide **cómo preguntarlo**.

# 28. Development Handoff a Gentle AI

El handoff es una fase formal del producto.

Engineering Platform no termina cuando los archivos fueron clonados.

Termina cuando:

```text
arquitectura resuelta
+ composición materializada
+ quality gates ejecutados
+ contexto de agentes generado
+ ownership transferido a Gentle
```

---

## 28.1 Objetivo

Gentle debe tomar control sin que el usuario tenga que volver a explicar:

- qué quiere construir;
- para quién;
- qué apps existen;
- qué foundation usa cada app;
- qué arquitectura se eligió;
- qué features faltan;
- qué decisiones están cerradas.

---

## 28.2 Outputs mínimos

```text
.engineering/
├── project-intent.json
├── architecture-decision.json
├── materialization-plan.json
├── project-map.json
├── implementation-brief.md
├── handoff.json
├── project.json
└── provenance.json

AGENTS.md
ARCHITECTURE.md
GENTLE.md
```

Y dentro de cada Surface:

```text
apps/dashboard/AGENTS.md
apps/mobile/AGENTS.md
services/api/AGENTS.md
```

cuando corresponda.

---

## 28.3 Agent Context Routing

El `AGENTS.md` raíz es un **router del repositorio**.

No debe repetir las reglas internas de cada boilerplate.

Ejemplo para API + Dashboard + Mobile:

```text
Project
├── services/api      → services/api/AGENTS.md
├── apps/dashboard    → apps/dashboard/AGENTS.md
└── apps/mobile       → apps/mobile/AGENTS.md
```

Debe contener:

- propósito del proyecto;
- tabla de Surfaces;
- paths;
- responsabilidades;
- dependencies entre Surfaces;
- ruta de `AGENTS.md` local;
- reglas cross-surface;
- fuentes de verdad.

Ejemplo conceptual:

```markdown
| Surface | Path | Responsibility | Instructions |
|---|---|---|---|
| API | services/api | Shared backend | services/api/AGENTS.md |
| Dashboard | apps/dashboard | Admin UI | apps/dashboard/AGENTS.md |
| Mobile | apps/mobile | Native client | apps/mobile/AGENTS.md |
```

---

## 28.4 `project-map.json`

Crear equivalente machine-readable:

```json
{
  "schema_version": 1,
  "surfaces": {
    "api": {
      "path": "services/api",
      "provider": "hono-api",
      "instructions": "services/api/AGENTS.md"
    },
    "dashboard": {
      "path": "apps/dashboard",
      "provider": "tanstack-admin",
      "instructions": "apps/dashboard/AGENTS.md"
    },
    "mobile": {
      "path": "apps/mobile",
      "provider": "ignite",
      "instructions": "apps/mobile/AGENTS.md"
    }
  },
  "relationships": [
    {
      "from": "dashboard",
      "to": "api",
      "type": "consumes"
    },
    {
      "from": "mobile",
      "to": "api",
      "type": "consumes"
    }
  ]
}
```

---

## 28.5 `implementation-brief.md`

Debe ser la síntesis humana del estado final.

Contenido:

```text
qué se construye
problema
usuarios
flujos
product requirements
architecture requirements
selected foundation(s)
features ya provistas por boilerplates
features pendientes
technical constraints
out-of-scope
open product questions
```

No duplicar todos los JSON.

Debe ser suficientemente corto para que un agente lo lea al iniciar.

---

## 28.6 `handoff.json`

Debe declarar formalmente el cambio de ownership.

Ejemplo:

```json
{
  "schema_version": 1,
  "status": "ready_for_implementation",
  "next_owner": "gentle-ai",
  "locked": [
    "architecture",
    "surface-topology",
    "selected-foundations"
  ],
  "requirements": [],
  "open_questions": []
}
```

Categorías:

### Locked

Decisiones que Gentle normalmente no vuelve a descubrir:

```text
architecture
selected boilerplates
surface topology
database profile
explicit technical constraints
folder structure
```

### Requirements

Lo que queda por implementar.

### Open questions

Detalles de producto que no eran necesarios para resolver arquitectura.

---

## 28.7 `GENTLE.md`

Debe ser corto y operativo.

Debe indicarle a Gentle:

1. leer `implementation-brief.md`;
2. leer `AGENTS.md`;
3. consultar `ARCHITECTURE.md`;
4. usar `project-map.json` para rutas;
5. leer el `AGENTS.md` de la Surface antes de modificarla;
6. no pedir al usuario que repita información ya registrada.

Regla clave:

> Do not rediscover the selected architecture.

---

## 28.8 Gentle decide Direct Build vs SDD

Después de leer el handoff:

```text
requirements suficientemente definidos
→ Gentle implementa directamente

faltan reglas/dominio/edge cases críticos
→ Gentle inicia SDD
```

Engineering Platform no impone SDD.

---

## 28.9 Alcance de una sesión SDD

La sesión SDD puede preguntar:

```text
estados de una orden
permisos
reglas de aprobación
edge cases
contenido exacto de reportes
workflows
dominio
```

No debe volver a preguntar:

```text
TanStack o Next
Postgres o Turso
qué boilerplate usar
dónde vive mobile
```

salvo que aparezca evidencia concreta de contradicción con una decisión ya cerrada.

---

## 28.10 Estado final

Al completar el flujo:

```text
Engineering Platform completed successfully

Project state:
READY_FOR_IMPLEMENTATION

Architecture:
✓ resolved
✓ materialized
✓ validated

Agent context:
✓ generated

Next owner:
Gentle AI
```

Este estado también debe ser machine-readable.

# 29. Fuentes de verdad

V1 debe definirlas explícitamente.

## Plataforma / catálogo

```text
catalog/metadata.json
catalog/recipes/
catalog/boilerplates/
catalog/surfaces/
catalog/capabilities/
catalog/vocabulary/
catalog/database-profiles/
catalog/feature-packs/
catalog/compatibilities/
catalog/adapters/
catalog/curation/
```

## Proyecto generado

```text
.engineering/project-intent.json
.engineering/architecture-decision.json
.engineering/materialization-plan.json
.engineering/project-map.json
.engineering/implementation-brief.md
.engineering/handoff.json
.engineering/project.json
.engineering/provenance.json
```

## Human agent routing

```text
AGENTS.md
ARCHITECTURE.md
GENTLE.md
<surface>/AGENTS.md
```

Evitar que README, package metadata y archivos individuales mantengan versiones independientes manualmente.

La versión del catálogo debe tener una fuente canónica.

El core y el catálogo deben versionarse de manera independiente.

# 30. Versionado

Separar:

```text
core binary version
catalog version/revision
schema versions
recipe versions
boilerplate catalog revision
adapter versions
```

No asumir que todos evolucionan juntos.

Ejemplo:

```text
eng core: 1.0.0
catalog: 2026.09.15
```

Después:

```text
eng core: 1.0.0
catalog: 2026.10.02
```

puede incluir nuevos boilerplates sin release del core.

El binario debe exponer:

```text
eng version
```

con:

```text
binary version
catalog version
build commit
build date
```

cuando estén disponibles.

## Compatibilidad

El catálogo debe declarar:

```text
minimum_core_version
schema_version
```

o mecanismo equivalente.

El core debe rechazar un catálogo que requiera operaciones/schema que no entiende.

## Regla de releases

No sacar release del núcleo por:

```text
nuevo boilerplate
nuevo pin
nueva Surface
nuevo database profile
nuevo alias
nueva Recipe
nuevo adapter declarativo
nueva combinación validada
```

cuando los contratos existentes basten.

Sí sacar release cuando:

```text
nuevo tipo de operación del engine
nuevo contrato incompatible
nueva capacidad de materialización no representable
fix del resolver/core
```

# 31. Dataset de routing: activo crítico

Antes de optimizar el scoring, crear un dataset de escenarios.

Directorio:

```text
testdata/routing/
```

Casos mínimos:

```text
01-simple-public-site.json
02-company-site-blog.json
03-internal-admin.json
04-internal-requests-approval.json
05-python-reporting-tool.json
06-mobile-field-app.json
07-desktop-local-tool.json
08-admin-plus-mobile.json
09-admin-plus-public-intake.json
10-commercial-saas.json
11-saas-plus-marketing.json
12-saas-plus-mobile.json
13-offline-public-intake.json
14-api-multiple-clients.json
15-public-docs-only.json
16-admin-with-files.json
17-kiosk-intake.json
18-desktop-offline.json
19-ambiguous-admin-or-multiapp.json
20-unsupported-specialized-case.json
```

Expandir hasta por lo menos 40–50 casos antes de declarar estable el router.

---

# 32. Formato de caso de routing

Ejemplo:

```json
{
  "name": "school-public-requests",

  "intent": {
    "...": "..."
  },

  "expect": {
    "status": "resolved",
    "selected_recipe": "GP-06",

    "must_include_reasons": [
      "public-intake",
      "shared-backend"
    ],

    "must_not_select": [
      "GP-02"
    ]
  }
}
```

Para casos ambiguos:

```json
{
  "expect": {
    "status": "ambiguous",
    "discriminating_dimensions": [
      "public_access"
    ]
  }
}
```

---

# 33. Tipos de pruebas

## Unit tests

Especialmente:

- normalizer;
- constraints;
- candidate generation;
- scoring;
- confidence;
- composition compatibility;
- destination safety.

## Contract tests

- catalogs válidos;
- ids únicos;
- referencias existentes;
- pins;
- adapters;
- capabilities;
- recipes.

## Routing evals

Dataset completo.

## Integration tests

```text
Intent → Decision → Plan
```

sin red cuando sea posible.

## E2E

Materializar fixtures locales pequeños.

Los E2E contra upstream reales deben quedar separados de los tests deterministas.

---

# 34. Golden tests

Para outputs estructurados importantes usar golden files con criterio.

Buenos candidatos:

```text
ArchitectureDecision
MaterializationPlan
project manifest
Gentle handoff
```

Evitar golden tests para mensajes humanos irrelevantes.

---

# 35. Reproducibilidad

Una misma combinación de:

```text
Engineering Platform version
catalog revision
ProjectIntent
```

debe producir la misma decisión salvo cambios explícitos en políticas.

Calcular fingerprint del Intent normalizado.

No usar:

- hora actual;
- orden aleatorio;
- respuesta de LLM;
- estado de red,

para resolver la arquitectura.

---

# 36. Observabilidad CLI

No se necesita telemetría remota en v1.

Sí se necesita visibilidad local.

`eng resolve` debe permitir:

```text
--json
--verbose
```

Salida humana:

```text
Selected: GP-06 Multi-App
Score: 92
Confidence: high

Why:
  + requires authenticated admin
  + requires anonymous public intake
  + both share domain data

Rejected:
  GP-02 Admin Application
    - public-intake cannot be satisfied safely
```

---

# 37. Errores

Definir clases/categorías conceptuales:

```text
validation
catalog
resolution
composition
materialization
runtime
filesystem
external-command
```

La CLI transforma estos errores a:

- mensaje;
- exit code;
- hints.

El dominio no imprime.

---

# 38. Catálogo inicial a migrar

Usar el catálogo 0.x como referencia y revalidar cada entrada antes de incorporarla.

Entradas relevantes actuales incluyen:

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

No copiar automáticamente todos sus campos.

Procedimiento:

1. leer entrada vieja;
2. leer adapter;
3. leer evidencia;
4. transformar al contrato v1;
5. validar;
6. agregar contract test;
7. agregar al menos un routing/composition scenario si participa en decisiones.

---

# 39. Qué conservar del proyecto anterior

Conservar como conocimiento:

- filosofía;
- Golden Paths útiles;
- catálogo curado;
- aliases;
- evidencias;
- pins;
- maintenance tiers;
- adapters;
- database profiles;
- feature packs;
- quality gates;
- concepto de surfaces;
- manifests;
- handoff;
- integración Pi;
- flujo Gentle;
- canonical examples;
- evals útiles.

---

# 40. Qué NO portar directamente

No copiar como arquitectura:

```text
scripts/eng.py
funciones gigantes
dict[str, Any] como dominio general
hardcoded SURFACE_CAPABILITIES
hardcoded SURFACE_SYNONYMS
project_type-first resolver
argparse handlers con lógica de negocio
lecturas directas de ROOT dentro del resolver
filesystem dentro del motor de decisión
```

Si una función antigua contiene una regla necesaria:

1. identificar la regla;
2. escribir un test;
3. reimplementarla en el componente correcto.

No copiar la función.

---

# 41. Política anti-monolito

Límites orientativos, no burocráticos:

- ninguna función de negocio > ~80 líneas sin razón clara;
- ningún archivo del core debe convertirse en contenedor de múltiples dominios;
- `cmd/eng/main.go` debe ser casi trivial;
- handlers CLI delegan en application services;
- parsers y serializers fuera del dominio cuando corresponda.

No dividir artificialmente funciones pequeñas solo para cumplir números.

El objetivo es cohesión.

---

# 42. Política anti-abstracción accidental

No crear inicialmente:

- event bus;
- DI container;
- plugin system dinámico;
- repository pattern para archivos JSON;
- CQRS;
- message broker;
- reflection framework;
- generic service registry.

Solo introducir una abstracción cuando exista una segunda implementación real o una frontera de testing clara.

---

# 43. Estrategia de implementación para Pi Agent

Pi debe implementar por slices verticales pequeños.

## Fase 0 — Bootstrap limpio

Entregables:

- repo Go;
- `go.mod`;
- Makefile;
- CI;
- lint/test/build;
- README mínimo;
- ADR-0001: Why Go / rewrite;
- estructura base.

Gate:

```text
go test ./...
go vet ./...
go build ./cmd/eng
```

---

## Fase 1 — Domain contracts

Implementar:

- ProjectIntent;
- Recipe;
- Boilerplate;
- Surface;
- Capability;
- Constraint;
- ArchitectureDecision.

Crear fixtures.

Todavía sin CLI sofisticada.

Gate:

- serialization roundtrip;
- validation tests;
- zero infrastructure imports in domain.

---

## Fase 2 — Catalog

Implementar:

- loader;
- schema/semantic validator;
- index;
- unique IDs;
- cross references.

Crear catálogo v1 mínimo con 2–3 Recipes y providers suficientes para probar.

No migrar todo aún.

Gate:

```text
eng catalog validate
```

---

## Fase 3 — Resolver v1

Implementar:

```text
validate
normalize
derive
candidates
hard constraints
score
confidence
explain
```

Cubrir primero:

```text
Public Web
Admin
Multi-App
```

Gate:

mínimo 15 routing scenarios pasando.

---

## Fase 4 — Composer

Implementar:

- surface providers;
- capabilities;
- compatibility;
- destinations;
- collisions;
- dependencies.

Gate:

casos:

```text
admin only
admin + public-intake
admin + mobile
SaaS + public-web
```

según el catálogo ya migrado.

---

## Fase 5 — MaterializationPlan

Convertir decisión + composición en plan puro.

Gate:

golden tests deterministas.

---

## Fase 6 — Materializer

Implementar primero con fixtures locales.

Después Git pinned sources.

Gate:

proyecto de prueba materializado en temp dir y validado.

---

## Fase 7 — CLI completa del flujo principal

Cerrar:

```text
eng resolve
eng explain
eng plan
eng materialize
eng doctor
```

Gate:

E2E local.

---

## Fase 8 — Pi integration

Crear el paquete Pi nuevo.

Pi produce Intent, nunca Recipe.

Gate:

flujo manual:

```text
/new-project
→ preguntas
→ intent
→ resolve
→ posible pregunta discriminante
→ confirmación
→ plan
→ materialize
```

---

## Fase 9 — Migración del catálogo útil

Migrar progresivamente:

- Stardrive;
- Hono API / api-starter;
- TanStack Admin;
- TanStack Transactional PWA;
- Ignite;
- React Starter Kit;
- Tauri;
- Python/Data choices.

Cada migración requiere tests.

---

## Fase 10 — Evolución del proyecto

Solo después de estabilizar creación:

```text
eng add
eng surface add
eng extend
eng update
```

No portar antes.

---

# 44. Regla de commits para Pi

Pi debe producir commits pequeños y semánticos.

Ejemplos:

```text
chore: bootstrap Go CLI
feat(domain): add project intent contracts
feat(catalog): validate recipe references
feat(resolver): apply hard surface constraints
test(routing): add admin plus public intake cases
feat(composer): resolve surface providers
feat(materializer): materialize pinned git source
feat(pi): add project discovery integration
```

No hacer un único commit masivo para toda la reescritura.

---

# 45. Definition of Done por fase

Una fase solo se considera cerrada cuando:

- compila;
- tests pasan;
- API pública relevante documentada;
- no existen TODO críticos escondidos;
- no se rompe una frontera arquitectónica;
- fixtures añadidos;
- decisiones nuevas documentadas si son importantes.

---

# 46. Definition of Done de Engineering Platform 1.0

V1 puede declararse candidata cuando:

1. no depende de Python para `eng`;
2. `eng` compila como binario Go;
3. el resolver no requiere `project_type`;
4. Pi no selecciona Recipes;
5. Pi usa `@juicesharp/rpiv-ask-user-question` para preguntas controladas en la integración oficial;
6. catálogo, Surfaces y vocabulario no están hardcodeados como listas cerradas en el resolver;
7. un nuevo boilerplate representable puede agregarse mediante catálogo + adapter + evidencia + tests sin recompilar el core;
8. routing dataset contiene >= 40 casos útiles;
9. casos ambiguos producen dimensiones discriminantes;
10. cada decisión explica selección y rechazos;
11. `catalog-gap` se usa únicamente para gaps arquitectónicos;
12. features faltantes no descalifican una foundation arquitectónicamente válida;
13. composer valida combinaciones;
14. materialization plan es reproducible;
15. materializer trabaja con pins;
16. `eng doctor` comprueba consistencia;
17. al menos los Golden Paths principales han sido materializados en CI;
18. cada proyecto multi-surface genera `AGENTS.md` raíz como router;
19. cada Surface materializada conserva/expone su `AGENTS.md` especializado cuando existe;
20. se genera `.engineering/project-map.json`;
21. se genera `.engineering/implementation-brief.md`;
22. se genera `.engineering/handoff.json`;
23. Pi ejecuta el flujo end-to-end;
24. Gentle recibe el proyecto sin requerir que el usuario vuelva a describir la idea;
25. Gentle puede decidir direct build vs SDD;
26. una sesión SDD no vuelve a descubrir decisiones arquitectónicas cerradas;
27. no existe un nuevo archivo equivalente a `eng.py`;
28. documentación refleja la arquitectura real.

# 47. Métricas de calidad

No medir éxito por cantidad de features.

Medir:

## Routing

```text
exact routing accuracy
acceptable alternative rate
false confident routing
ambiguity detection quality
```

Objetivo especialmente importante:

> Un routing dudoso detectado como ambiguo es mejor que un routing incorrecto con confianza alta.

## Ingeniería

```text
unit test coverage del resolver
routing scenario coverage
catalog validation
materialization success rate
reproducibility
```

No imponer porcentaje global de coverage como objetivo aislado.

---

# 48. Casos que deben influir en el diseño desde el inicio

Engineering Platform debe manejar correctamente, entre otros:

## SaaS con landing

No crear un nuevo Project Type.

```text
Commercial SaaS primary
+ public-web Surface
```

## Admin + app móvil

```text
shared backend
+ web-admin
+ mobile-native
→ Multi-App
```

## Admin + formulario público offline

```text
web-admin
+ public-intake
+ offline architectural requirement
→ Multi-App
```

## Blog simple

```text
public-web
+ blog product requirement
→ Public Web foundation
```

## API para múltiples clientes

El hecho de tener varios clientes debe influir en shared backend/composition, no depender de una etiqueta manual.

## App móvil futura

Si `mobile` es solo posibilidad futura y no requirement actual, no instalar Ignite todavía.

## TUI para Excel y PDF

Usuario:

> “Necesito crear una TUI para manejar datos de Excel, compararlos y generar PDF.”

Clasificación:

```text
Surface:
tui

Architecture requirements:
local execution
local filesystem

Product requirements:
Excel
comparison
PDF reports
```

Si el catálogo tiene una foundation Python TUI adecuada:

```text
resolved
```

aunque no incluya PDF.

Gentle implementa las features restantes.

Si no existe ningún provider TUI:

```text
catalog-gap
```

y Pi recomienda investigar/curar una foundation.

## TUI: Python vs Go

No fijar:

```text
tui = go
```

ni:

```text
tui = python
```

El catálogo debe permitir distintas foundations.

El fit depende del contexto.

Ejemplo data-heavy/Excel puede favorecer Python.

Ejemplo system tooling/concurrency puede favorecer Go.

El usuario también puede imponer/preferir un lenguaje.

## Usuario técnico: TanStack + offline + Turso

Usuario:

> “Voy a desarrollar un dashboard usando TanStack y quiero datos offline con Turso.”

Pi no pregunta qué framework usar.

Registra:

```text
Surface:
web-admin

Architecture:
offline requirement

Technical:
TanStack
Turso
```

Engineering Platform valida catálogo/compatibilidad y enruta a la foundation adecuada.

Si TanStack + Turso son compatibles y curados:

```text
resolved
```

Si Turso no existe en el catálogo pero la foundation TanStack sí:

- determinar si Turso es una restricción obligatoria;
- si lo es y no hay profile/provider representable, puede existir `catalog-gap`;
- si solo es una preferencia, explicar alternativa compatible.

## Boilerplate con features adicionales

Si una foundation ya incluye:

```text
Excel support
PDF generation
auth
```

y son features solicitadas:

```text
aprovecharlas
```

pero no convertirlas en requisito de elegibilidad universal.

# 49. Reglas para “ahora vs futuro”

ProjectIntent debe distinguir:

```text
required_now
planned_later
explicitly_excluded
```

Evitar materializar tecnología solo por una posible fase futura.

Sí conservar compatibilidad cuando sea razonable.

Ejemplo:

```text
mobile planned later
```

puede favorecer una API reusable, pero no agrega una app móvil hoy.

---

# 50. Feature status

Mantener la disciplina útil del proyecto anterior, pero separar dos planos.

## Foundation-provided

```text
provided-by-foundation
```

## Product implementation

```text
requested
pending-implementation
in-progress
verified
```

Una feature solicitada no debe marcarse como implementada si el boilerplate no la aporta realmente.

Pero su ausencia tampoco implica que el boilerplate sea inválido.

Ejemplo:

```text
PDF report
requested
pending-implementation
```

es un estado normal después de materializar una foundation TUI.

El handoff a Gentle debe incluir claramente:

```text
already provided
vs
still to implement
```

# 51. Curation workflow

El sistema de curación es importante, pero no forma parte del resolver runtime.

Pipeline conceptual:

```text
candidate repository
      ↓
dedupe
      ↓
license
      ↓
maintenance
      ↓
architecture fit
      ↓
AI friendliness
      ↓
security
      ↓
adapter
      ↓
pilot
      ↓
curated/released
```

Mantener evidencia junto al catálogo.

---

# 52. Update strategies

Cada boilerplate puede tener:

```text
replace
merge-seed
fork-track
manual
```

o un vocabulario final equivalente.

No implementar actualización automática agresiva en v1 inicial.

`eng update` deberá generar primero un plan/reporte.

---

# 53. CI

Pipeline mínimo del repositorio:

```text
format check
go vet
unit tests
routing tests
catalog validation
build linux amd64
build linux arm64
build darwin arm64
build windows amd64
```

Posteriormente:

```text
materialization pilots
```

No hacer que tests básicos dependan de GitHub/upstreams.

---

# 54. Releases

Construir binarios reproducibles mediante tags.

Artefactos deseados:

```text
eng-linux-amd64
eng-linux-arm64
eng-darwin-arm64
eng-windows-amd64.exe
checksums.txt
```

## Core y catálogo tienen ciclos distintos

El core se libera cuando cambia el motor.

El catálogo puede actualizarse con mucha mayor frecuencia.

Ejemplo:

```text
Core 1.0.0
Catalog 2026.09.01

Core 1.0.0
Catalog 2026.09.08
  + python-tui

Core 1.0.0
Catalog 2026.09.14
  + new TanStack pin
  + Turso profile
```

No crear releases artificiales del core solo para distribuir conocimiento nuevo.

## Catálogos adicionales

El diseño debe permitir evolucionar hacia:

```text
default catalog
organization overlay
user overlay
```

sin plugin runtime ejecutable.

Esto permite, por ejemplo, un catálogo privado de una consultoría manteniendo el core común.

# 55. Documentación mínima

Antes de v1:

```text
README.md
docs/concepts/project-intent.md
docs/concepts/recipe.md
docs/concepts/surface.md
docs/concepts/boilerplate.md
docs/architecture/core.md
docs/architecture/routing.md
docs/architecture/materialization.md
docs/guides/new-project.md
docs/guides/curate-boilerplate.md
docs/guides/add-recipe.md
docs/adr/
```

Documentar fronteras, no repetir implementación línea por línea.

---

# 56. ADRs iniciales

Crear al menos:

```text
ADR-0001-rewrite-in-go.md
ADR-0002-deterministic-core.md
ADR-0003-project-intent-not-project-type.md
ADR-0004-declarative-catalog.md
ADR-0005-resolution-and-materialization-separation.md
ADR-0006-pi-as-adapter.md
```

Breves y prácticos.

---

# 57. Preguntas que Pi NO debe hacer al propietario durante la construcción inicial

Si este PRD ya define la decisión, Pi no debe volver a preguntar:

- ¿Python o Go?
- ¿Portar `eng.py`?
- ¿Debe Pi seleccionar stack?
- ¿Mantener `project_type`?
- ¿Usar LLM dentro del resolver?
- ¿Separar resolver/materializer?
- ¿Catálogo declarativo?
- ¿Binario único?
- ¿Conservar boilerplates curados?

Ya están resueltas.

---

# 58. Decisiones reversibles que Pi puede tomar sin bloquear

Pi puede decidir razonablemente:

- Cobra vs stdlib CLI;
- librería JSON Schema concreta;
- logger ligero;
- layout menor de paquetes;
- nombre de tipos internos;
- library de diff;
- formato concreto de mensajes CLI.

Condiciones:

- simplicidad;
- mantenimiento;
- tests;
- documentar si afecta contrato.

---

# 59. Investigación permitida durante implementación

Pi puede consultar documentación actual para:

- Go 1.27;
- Pi packages/extensions/skills;
- Cobra si se adopta;
- JSON Schema;
- GitHub Actions;
- seguridad de ejecución de procesos;
- release tooling.

No debe utilizar investigación externa para cambiar la visión de producto sin necesidad.

---

# 60. Señales de arquitectura incorrecta

Pi debe detener y refactorizar si aparece cualquiera de estas señales:

```text
main.go contiene routing
CLI handler conoce scoring
resolver ejecuta Git
materializer elige Recipe
Pi escribe selected_recipe
catalog loader conoce cwd del proyecto
domain importa Cobra
domain importa os/exec
nuevo archivo supera continuamente responsabilidades
capabilities vuelven a hardcodearse
tests de routing necesitan mocks de filesystem
```

---

# 61. Primer milestone funcional

El primer milestone NO será “portar todos los comandos”.

Será:

## M1 — Deterministic Routing Core

Entrada:

```text
ProjectIntent
```

Salida:

```text
ArchitectureDecision
```

Debe cubrir correctamente:

```text
Public Web
Admin
Multi-App
```

con:

- hard constraints;
- score;
- confidence;
- rejected reasons;
- ambiguous case;
- 15+ scenarios.

Solo después avanzar al materializer.

---

# 62. Segundo milestone

## M2 — Reproducible Composition

Entrada:

```text
ArchitectureDecision
+ Catalog
```

Salida:

```text
MaterializationPlan
```

Debe demostrar al menos:

```text
Stardrive
Hono API
TanStack Admin
TanStack Transactional PWA
```

y resolver destinos/compatibilidad.

---

# 63. Tercer milestone

## M3 — End-to-End Project Bootstrap

```text
Idea
→ Pi discovery
→ ProjectIntent
→ resolve
→ decision
→ plan
→ materialize
→ tests
→ Gentle handoff
```

Una vez logrado, la nueva arquitectura habrá probado el concepto completo.

---

# 64. Compatibilidad con legacy

No se requiere que los manifests v0.x funcionen directamente en v1.0.

Si posteriormente se necesita migración:

```text
eng migrate legacy-project
```

debe ser una herramienta separada.

No contaminar el nuevo dominio con compatibilidad prematura.

El repositorio legacy se considera read-only durante la reconstrucción.

---

# 65. Estrategia recomendada de repositorio

Preferencia:

1. crear la nueva implementación en una rama/repo limpio;
2. conservar el histórico 0.x accesible;
3. no borrar legacy hasta demostrar M3;
4. cuando v1 sea funcional, decidir si:
   - reemplaza `main`, o
   - se publica como nueva major desde el mismo repo.

No mezclar Go nuevo y Python viejo mediante imports o fallback permanente.

Un wrapper temporal de transición es aceptable únicamente fuera del core y con fecha de eliminación.

---

# 66. Criterio final de arquitectura

Antes de aprobar cualquier módulo nuevo, responder:

1. ¿Qué responsabilidad tiene?
2. ¿Qué input recibe?
3. ¿Qué output produce?
4. ¿Tiene efectos secundarios?
5. ¿Quién puede importarlo?
6. ¿Está decidiendo producto o ejecutando una decisión?
7. ¿Su conocimiento debería ser código o catálogo?
8. ¿Se puede probar sin Pi, Git y filesystem?

Si estas respuestas son ambiguas, probablemente se están mezclando capas.

---

# 67. Resumen ejecutivo para Pi

Construir Engineering Platform 1.0 como un **motor local determinista de selección, composición y materialización de stacks curados**.

La reescritura debe corregir el problema central de 0.x:

> el agente no debe clasificar previamente la arquitectura para que el resolver solo la confirme.

Nuevo principio:

> **El usuario describe el producto. Pi captura intención. El resolver decide arquitectura. El Composer selecciona/completa surfaces. El Materializer ejecuta un plan. Gentle desarrolla el producto.**

Implementar el core en Go 1.27.x.

Priorizar primero:

```text
ProjectIntent
→ deterministic Resolver
→ explainable ArchitectureDecision
→ routing dataset
```

Luego:

```text
Composer
→ MaterializationPlan
→ Materializer
```

Finalmente:

```text
Pi integration
→ Gentle handoff
→ evolution commands
```

No migrar funciones por traducción.

Extraer requisitos y reimplementarlos en la capa correcta.

---

# 68. Checklist de inicio inmediato para Pi Agent

- [ ] Crear repositorio/worktree limpio para v1.
- [ ] Inicializar Go 1.27.x.
- [ ] Crear estructura de paquetes.
- [ ] Crear CI mínimo.
- [ ] Escribir ADR-0001 a ADR-0006.
- [ ] Definir `ProjectIntent`.
- [ ] Definir `ArchitectureDecision`.
- [ ] Crear catálogo v1 mínimo.
- [ ] Crear validator de catálogo.
- [ ] Crear 15 escenarios de routing.
- [ ] Implementar resolver puro.
- [ ] Implementar hard constraints.
- [ ] Implementar scoring.
- [ ] Implementar confidence.
- [ ] Implementar explicaciones.
- [ ] Implementar `ambiguous`.
- [ ] Exponer `eng resolve`.
- [ ] Exponer `eng explain`.
- [ ] Cerrar milestone M1 antes de portar otras funciones.

---

# 69. Fuentes de referencia de la reconstrucción

Repositorio legacy:

`https://github.com/JhonMA82/engineering-platform`

Archivos especialmente útiles como **referencia funcional, no arquitectónica**:

```text
scripts/eng.py
schemas/intake.schema.json
platform/golden-paths.json
platform/boilerplates.json
platform/database-profiles.json
platform/feature-packs.json
docs/06-ai-harness/pi-integration.md
curation/
canonical-examples/
evals/
tests/
```

Go:

`https://go.dev/doc/go1.27`

Historial de releases:

`https://go.dev/doc/devel/release`

---

# 70. Orden final de prioridad

```text
1. Correctitud del routing
2. Separación arquitectónica
3. Explicabilidad
4. Reproducibilidad
5. Seguridad de materialización
6. Simplicidad del catálogo
7. Integración Pi
8. Experiencia CLI
9. Evolución/update
10. Nuevas capabilities
```

No sacrificar los primeros cinco puntos para recuperar rápidamente todas las features de v0.x.

---


# 71. Arquitectura extensible sin releases del núcleo

Este principio es parte central de v1.

## 71.1 Operaciones dirigidas por catálogo

El core debe implementar un vocabulario pequeño de operaciones genéricas.

Ejemplos:

```text
fetch/clone pinned source
copy
prune
template
compose destination
set managed metadata
run curated setup command
run curated check
```

Un adapter combina estas operaciones.

Agregar otro boilerplate que use estas operaciones no requiere nuevo core.

## 71.2 Cuándo sí se modifica el core

Solo cuando aparece algo que no puede expresarse de manera segura con el contrato existente.

Ejemplo hipotético:

```text
AST-aware cross-project transformation
```

si el engine no dispone de esa primitive.

Entonces:

```text
Core 1.1
→ agrega operación declarativa nueva
```

Después múltiples adapters pueden usarla.

## 71.3 No crear un plugin runtime prematuramente

No usar inicialmente:

```text
Go shared libraries
dynamic executable plugins
RPC plugin host
arbitrary code downloaded from catalog
```

El catálogo debe ser declarativo y los comandos ejecutables deben venir de adapters curados y validados.

---

# 72. Estrategia de CatalogGap

Un `CatalogGap` debe ser accionable.

Ejemplo:

```json
{
  "status": "catalog-gap",
  "missing_architecture": [
    {
      "kind": "surface",
      "id": "tui"
    }
  ],
  "research_criteria": {
    "preferred_languages": ["python", "go"],
    "required_architecture": ["local-filesystem"]
  }
}
```

Pi puede explicar:

> El catálogo actual no contiene una foundation TUI curada. La necesidad arquitectónica está clara, pero no es materializable con las opciones disponibles.

Después se puede:

```text
1. buscar candidates
2. evaluar mantenimiento/licencia
3. pilotar
4. crear adapter
5. curar
6. registrar
7. volver a ejecutar resolve
```

Una vez registrada, futuros proyectos TUI pueden usarla sin cambiar `eng`.

No usar este mecanismo para features normales.

---

# 73. Agent Navigation Principle

Regla normativa:

> An agent receiving a materialized project must not need repository discovery to understand its topology.

La ruta de contexto es:

```text
GENTLE.md
   ↓
implementation-brief.md
   ↓
AGENTS.md raíz
   ↓
project-map.json / ARCHITECTURE.md
   ↓
Surface-specific AGENTS.md
```

Cada nivel tiene una responsabilidad.

## Root `AGENTS.md`

Qué existe y dónde trabajar.

## Surface `AGENTS.md`

Cómo trabajar correctamente dentro de esa foundation.

## `ARCHITECTURE.md`

Cómo se relacionan las partes.

## `implementation-brief.md`

Qué producto se está construyendo y qué falta.

## `GENTLE.md`

Cómo tomar control.

Evitar duplicación entre niveles.

---

# 74. Ownership Boundary: Pi → Engineering Platform → Gentle

Fronteras finales:

```text
Pi
└── discovery
    ├── entiende lenguaje del usuario
    ├── hace preguntas
    ├── usa ask_user_question
    └── produce/refina ProjectIntent

Engineering Platform
└── engineering decision
    ├── resolver
    ├── catalog
    ├── composer
    ├── plan
    ├── materializer
    ├── validation
    └── handoff context

Gentle
└── product implementation
    ├── lee handoff
    ├── decide direct-build vs SDD
    ├── implementa features
    ├── respeta AGENTS.md
    └── evoluciona el producto
```

No permitir responsabilidad difusa entre estas tres etapas.

---

# 75. Freeze de diseño para comenzar implementación

A partir de este PRD, Pi debe tratar las siguientes decisiones como cerradas para el milestone inicial:

```text
Go core
deterministic resolver
no project_type routing
catalog-driven Surfaces/providers
boilerplates as foundations
product features handled by Gentle
CatalogGap only for architectural gaps
catalog updates independent from core releases
Pi as conversational adapter
rpiv ask-user-question for controlled questions
Agent Context Routing
formal Development Handoff
Gentle chooses direct implementation vs SDD
```

No abrir nuevas discusiones arquitectónicas salvo que una implementación demuestre una contradicción concreta.

El objetivo inmediato es construir y validar M1.

---

# 76. Prompt operativo de arranque para Pi Agent

Pi puede tomar este bloque como instrucción inmediata:

```text
Create Engineering Platform 1.0 from a clean Go codebase following this PRD.

Do not port scripts/eng.py.

Begin only with Milestone M1:
ProjectIntent → Catalog → Resolver → ArchitectureDecision.

Implement Public Web, Admin and Multi-App first.
Create at least 15 routing scenarios including one ambiguous case and one
architectural catalog-gap case.

Keep Product Requirements separate from Architecture Requirements.
Missing product features must never create a CatalogGap.

Surfaces and providers must be catalog-driven and must not be closed enums
that require recompiling the core to add a TUI or another future Surface.

Do not implement materialization, legacy command parity, update workflows or
Gentle integration until M1 is proven by tests.

Commit in small semantic increments and keep the core independent from Pi,
filesystem side effects and external processes.
```

## Decisión de producto

**Engineering Platform 1.0 será una reconstrucción limpia en Go, orientada a dominio, con un core determinista, un catálogo extensible independiente del ciclo de releases y un handoff formal entre Pi y Gentle AI.**

La idea original se conserva.

Los boilerplates serán foundations arquitectónicas, no colecciones de features terminadas.

La arquitectura histórica no se porta.
