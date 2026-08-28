# Open SaaS vs. dashboards Next/TanStack

## Diferencia esencial

- **Open SaaS:** aplicación full-stack comercial con auth, billing, jobs, correo y archivos.
- **Next/TanStack Admin:** capa de interfaz y patrones administrativos; el backend se diseña aparte.

## Open SaaS

Adecuado cuando el objetivo es lanzar un producto por suscripción y Wasp reduce trabajo real.

No asumir:

- organizaciones;
- tenant isolation;
- RBAC empresarial;
- auditoría institucional;
- cumplimiento sectorial.

## TanStack Admin

Adecuado para operación interna, backoffice y frontend de APIs.

## Next Admin

Adecuado si el producto integra sitio público y área privada sobre Next.js.

## Decisión

No introducir Open SaaS en proyectos institucionales solo porque trae autenticación. Auth no equivale a arquitectura institucional.
