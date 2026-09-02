---
name: project-evolution
description: 'Analiza y aplica cambios sobre un proyecto existente de Engineering Platform, incluyendo nuevas aplicaciones o capacidades, sin regenerar ni sobrescribir el código actual.'
compatibility: 'Requiere un proyecto materializado con `.engineering/project.json` y Engineering Platform 0.8.0.'
---

# Project Evolution

Convierte una necesidad posterior al bootstrap en el cambio mínimo verificable.

## Flujo

1. Lee `GENTLE.md`, `.engineering/project.json` y `.engineering/materialization.json`.
2. Pregunta en texto libre qué necesita cambiar si el usuario todavía no lo explicó. No presentes un catálogo completo.
3. Clasifica el cambio:
   - nueva aplicación o servicio basado en un starter: usa `<eng> extend <starter> --project .`;
   - nueva capacidad: usa `<eng> add <feature> --project .`;
   - base de datos, tenancy, sustitución o retiro: genera un plan y exige SDD/ADR; no lo apliques como extensión simple.
4. Muestra el dry-run y explica en cinco líneas o menos lo agregado, actualizado, preservado y pendiente.
5. Solicita confirmación explícita antes de usar `--apply`.
6. Después de aplicar, ejecuta `<eng> doctor --project .` y entrega el control a Gentle mediante el `GENTLE.md` actualizado.

## Invariantes

- No vuelvas a ejecutar `bootstrap`.
- No agregues un starter en una ruta ocupada.
- No describas una capacidad como implementada si aparece `pending-implementation`.
- No cambies Recipe, base de datos, autenticación o tenancy silenciosamente.
- Si la operación falla, conserva el proyecto previo y reporta el bloqueo.

`<eng>` es la ruta absoluta proporcionada por la extensión Pi; si falta, usa `eng` desde `PATH`.
