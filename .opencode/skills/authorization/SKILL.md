---
name: authorization
description: Agrega autenticación o permisos con denegación predeterminada y pruebas de aislamiento.
---

# Authorization

Modela `actor + acción + recurso + contexto`. Autoriza en el servidor y niega por defecto; ocultar botones no es control de acceso. Prueba permitido, denegado, otra organización, recurso inexistente y estado no válido. Audita acciones sensibles sin registrar secretos. Multitenancy exige aislamiento en consultas, archivos, jobs, caché y auditoría.
