# Open SaaS vs. React Starter Kit

## Diferencia esencial

- **Open SaaS:** producto comercial sobre Wasp (lenguaje declarativo + código generado), Node.js, Prisma y PostgreSQL; se despliega en hosting convencional.
- **React Starter Kit:** producto comercial full-stack TypeScript con React 19, TanStack Router, tRPC, Drizzle ORM y Cloudflare Workers; se despliega en el edge.

## Open SaaS

Adecuado cuando se acepta Wasp como framework y se quiere backend Node + PostgreSQL desplegable en hosting clásico.

No asumir: organizaciones, tenant isolation, RBAC empresarial, auditoría institucional ni cumplimiento sectorial sin diseño explícito.

## React Starter Kit

Adecuado cuando el stack completo en TypeScript importa más que la portabilidad de hosting y se acepta Cloudflare como plataforma.

No asumir: tenant isolation solo por tener organizaciones de Better Auth, portabilidad fuera de Cloudflare ni contrato de API pública abierto (tRPC acopla cliente y servidor).

## Criterios de elección

- **Hosting del cliente:** si exige servidores propios o nube general, Open SaaS; si acepta edge de Cloudflare, React Starter Kit.
- **Idioma/framework:** preferencia por TypeScript puro en toda la pila apunta a React Starter Kit; aceptar Wasp y su ciclo propio apunta a Open SaaS.
- **Múltiples clientes externos:** ninguno es API-first abierto; para eso está Full Stack FastAPI.
- **Madurez y comunidad:** ambos son proyectos grandes y activos; React Starter Kit existe desde 2014 y Open SaaS nació sobre Wasp.
- **AI-friendly:** ambos son fuertes; React Starter Kit trae AGENTS.md, CLAUDE.md, docs propias y asistentes entrenados sobre el código.

## Decisión

Son dos apuestas distintas para la misma categoría comercial. React Starter Kit está seleccionado por propuesta del usuario; Open SaaS sigue siendo la recomendación con piloto. Un proyecto comercial debe elegir una de las dos y documentar hosting, tenancy y salida/migración antes de comprometerse.
