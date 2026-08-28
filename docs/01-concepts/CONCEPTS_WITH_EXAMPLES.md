# Conceptos de la Engineering Platform — con ejemplos

Este documento es el mapa de aprendizaje para juniors. Cada concepto incluye **qué es, cuándo se usa, ejemplo y error típico**.

## 1. Engineering Platform
**Qué es:** la fuente de verdad sobre cómo la consultoría construye software.  
**Cuándo:** siempre; no es una librería que se instale en producción.  
**Ejemplo:** indica que una app interna usa GP-02 y que multi-tenancy es opcional.  
**Error típico:** pensar que es un mega-framework que debe contener todo el código.

## 2. Starter / Boilerplate
**Qué es:** punto de partida técnico para una categoría.  
**Ejemplo:** TanStack Admin para una interfaz administrativa.  
**Error típico:** elegir un boilerplate porque “se ve moderno” aunque no corresponda al problema.

## 3. Golden Path / Project Recipe
**Qué es:** combinación versionada de starters, datos, features, skills, gates y exclusiones para un tipo de proyecto.

**Ejemplo:** GP-02 = TanStack Admin + Hono + PostgreSQL administrado para una app administrativa.
**Error típico:** convertir cada variante pequeña en un Golden Path nuevo.

## 4. Feature Pack
**Qué es:** capacidad opcional instalable sobre un proyecto.  
**Ejemplo:** una escuela single-tenant agrega `webhooks` seis meses después para notificar a un ERP.  
**Error típico:** instalar auth, tenancy, Redis y jobs “por si acaso”.

## 5. Project Manifest
**Qué es:** archivo `.engineering/project.json` que declara cómo nació y cómo está compuesto un proyecto.  
**Ejemplo:**
```json
{
  "platform_version": "0.4.0",
  "scaffold_status": "blueprint",
  "recipe": {"id": "GP-02", "version": "1.0.0"},
  "starters": [
    {"id": "tanstack-admin", "pin": null},
    {"id": "hono-api", "pin": null}
  ],
  "database": "postgresql-managed",
  "features": ["auth", "rbac", "audit", "observability"]
}
```
**Error típico:** volver a decidir el stack desde cero en cada tarea.

## 6. Architecture Brief
**Qué es:** documento corto que explica por qué se eligió una arquitectura.  
**Ejemplo:** “No usamos multitenancy porque la primera versión sirve a una sola escuela”.  
**Error típico:** documentar solo qué se eligió, no por qué se descartaron alternativas.

## 7. ADR
**Qué es:** Architecture Decision Record para decisiones relevantes y duraderas.  
**Ejemplo:** “ADR-003: usamos Hono API externa porque web y móvil compartirán backend”.  
**Error típico:** crear ADR para cambios triviales como renombrar un botón.

## 8. Harness
**Qué es:** la capa de orquestación que decide qué contexto, skills, guards y flujo debe usar el agente.  
**Ejemplo:** un cambio de schema carga `database` y gates de migración, no Tauri ni Ignite.
**Error típico:** cargar todo el conocimiento del stack en cada prompt.

## 9. Skill
**Qué es:** conocimiento especializado y operativo para una tarea o stack.  
**Ejemplo:** `contracts` enseña cómo cambiar un endpoint sin romper consumidores.
**Error típico:** una mega-skill de 5,000 líneas que cubre todos los frameworks.

## 10. Guard
**Qué es:** verificación determinista exigida por una clase de cambio.  
**Ejemplo:** cambio de schema → migration test obligatorio.  
**Error típico:** confiar en que el modelo “revisó mentalmente” la migración.

## 11. Quality Gate
**Qué es:** conjunto de validaciones que debe pasar una tarea/release.  
**Ejemplo:** lint → types → unit → integration → contracts → build.  
**Error típico:** considerar CI verde suficiente aunque no cubra el riesgo específico del cambio.

## 12. Canonical Example
**Qué es:** ejemplo pequeño, correcto y probado que el agente debe imitar.  
**Ejemplo:** un endpoint Hono autenticado con contrato + use case + test.  
**Error típico:** usar como canonical example código de una feature enorme o desactualizada.

## 13. Eval
**Qué es:** escenario reproducible para medir si un modelo/harness/skill trabaja correctamente.  
**Ejemplo:** pedir “agrega permiso approve” y puntuar si actualiza policy, tests y docs sin romper arquitectura.  
**Error típico:** comparar modelos por impresión subjetiva.

## 14. Knowledge Entry
**Qué es:** problema o patrón comprobado que vale reutilizar.  
**Ejemplo:** error de migration Drizzle + causa + fix + prueba de verificación.  
**Error típico:** convertir knowledge en un bloc de notas sin evidencia.

## 15. Upgrade Recipe
**Qué es:** procedimiento versionado para actualizar un proyecto generado.  
**Ejemplo:** React Starter Kit usa su Recipe `merge-seed`, ejecuta checks y actualiza el commit del manifest.
**Error típico:** aplicar una estrategia genérica e ignorar el mecanismo nativo del upstream.

## 16. Upstream / Curated Starter
**Qué es:** upstream es el proyecto original; curated starter es la versión probada y adaptada por la consultoría.  
**Ejemplo:** actualizar TanStack Admin upstream en branch de integración, probar y publicar tag interno.  
**Error típico:** desarrollar cambios propios directamente en la rama espejo del upstream.

## 17. Monorepo multi-app
**Qué es:** patrón `apps/* + packages/*` para proyectos que realmente tienen varios clientes.  
**Ejemplo:** web + mobile comparten SDK y contratos.  
**Error típico:** crear mobile/desktop/worker vacíos en todo proyecto.

## 18. Design System
**Qué es:** reglas y componentes visuales propios para consistencia.  
**Ejemplo:** tablas, formularios, estados vacíos y tipografía usados igual en dos productos.  
**Error típico:** dejar que cada agente improvise shadcn de forma distinta.

## 19. Project Change vs Platform Change
**Project change:** “la escuela necesita un campo prioridad”.  
**Platform change:** “todos nuestros proyectos necesitan un patrón reusable para auditoría”.  
**Error típico:** meter requisitos específicos de un cliente en el starter global.

## 20. Knowledge Loop
**Qué es:** convertir errores reales en prevención reusable.  
**Ejemplo:** bug de permisos → regression test → knowledge entry → guard mejorado.  
**Error típico:** documentar el bug pero no agregar una prueba que evite repetirlo.
