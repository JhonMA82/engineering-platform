---
name: database
description: Diseña datos y migraciones reversibles para el perfil de base seleccionado.
---

# Database

Lee el perfil en `platform/database-profiles.json`. Toda modificación de esquema incluye migración, compatibilidad durante despliegue, respaldo/recuperación y prueba con datos representativos. No cambies de PostgreSQL a Turso por semejanza de API: valida características usadas, concurrencia, latencia, migraciones, restore y lock-in. En sincronización define autoridad, conflictos y recuperación antes del código.
