# Herramientas internas, backends rápidos y BI

## Directus

Útil para:

- Exponer REST/GraphQL sobre datos.
- Administrar catálogos.
- Crear un backoffice técnico rápido.
- Prototipos de APIs y contenido estructurado.

No sustituye:

- Dominio complejo.
- Workflows institucionales con invariantes.
- UX altamente personalizada.

## Appsmith

Útil para:

- Herramientas internas.
- Paneles de soporte.
- Prototipos operativos.
- Interfaces para bases y APIs.

No usar como default para:

- Producto público.
- UX compleja.
- Sistemas cuyo código debe evolucionar durante años sin dependencia de plataforma.

## Metabase

Útil para:

- Explorar datos.
- Dashboards analíticos.
- Indicadores.
- Consultas de negocio.
- Embedding cuando el modelo de licencia y seguridad lo permita.

No reconstruir en React cada reporte exploratorio. Construir dentro de la aplicación solo los indicadores que forman parte del workflow y requieren interacción específica.

## Criterio

Estas plataformas aceleran administración y análisis. El boilerplate sigue siendo responsable de autenticación, permisos, dominio, auditoría e integración segura.
