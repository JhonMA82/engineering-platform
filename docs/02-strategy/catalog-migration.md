# Migración desde boilerplates-catalog

La evolución no elimina el catálogo: separa conocimiento, decisión y entrega.

| Antes | Ahora | Motivo |
|---|---|---|
| `catalog.json` mezclaba tipo y madurez | `decision_status` + `delivery_status` | Una opción útil puede no estar lista para entregar |
| ID largo ligado al repositorio | ID estable + `legacy_ids` | Permite renames sin duplicar |
| Cada starter era una elección aislada | Project Recipe | Stack, datos, features, skills, gates y exclusiones se deciden juntos |
| “Seleccionado” podía parecer “productivo” | Canales `catalog-only` a `released` | Evita promesas sin artefacto probado |
| Actualización genérica por fork | Estrategia por entrada | Respeta `merge-seed`, Copier, overlays o releases internos |

## Entradas recuperadas

Se preservan las 17 entradas anteriores, incluidas Vercel Chatbot, GoShip, Self-hosted AI Starter Kit y los cinco starters/perfiles internos. Los renames conservan alias:

- `tanstack-shadcn-admin-dashboard` → `tanstack-admin`;
- `next-shadcn-admin-dashboard` → `next-admin`;
- `full-stack-fastapi-template` → `fastapi`.

El snapshot completo se encuentra en [`catalog/legacy-v1.2.3`](../../catalog/legacy-v1.2.3/README.md). Sirve como evidencia histórica; las decisiones vigentes se consultan en `platform/boilerplates.json`.
