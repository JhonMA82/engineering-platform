# Git Workflow

## Engineering Platform
- `main`: estado integrado y revisado.
- `feature/*`: cambios cortos.
- `release/*`: solo cuando una release necesita estabilización.
- tags SemVer para releases.

## Curated starters
Mantener upstream separado del trabajo interno. Nunca usar `main` como espejo y customization al mismo tiempo.

## Proyectos de clientes
- `main`: siempre desplegable según política del proyecto.
- ramas cortas por feature/fix.
- PR + CI.
- evitar ramas permanentes `multitenant`, `next`, `tanstack`, etc.

Las variantes de producto se modelan como starters/apps/features, no como ramas eternas.
