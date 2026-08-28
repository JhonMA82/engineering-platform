# Datos y almacenamiento

## PostgreSQL

Default para sistemas multiusuario por:

- madurez;
- transacciones;
- ecosistema;
- extensiones;
- compatibilidad con los stacks seleccionados.

No es obligatorio para una herramienta local de un usuario: SQLite puede ser suficiente.

## MinIO / S3

Usar object storage para:

- adjuntos;
- evidencias;
- documentos;
- exports;
- archivos de IA.

Diseñar:

- claves y namespaces;
- tenancy;
- URLs temporales;
- antivirus cuando aplique;
- retención;
- backups;
- cifrado.

## Qdrant / pgvector

- Qdrant puede ser útil en infraestructura especializada.
- pgvector reduce componentes cuando PostgreSQL es suficiente.
- Elegir mediante pruebas de escala, filtros, operación y recuperación.

## DuckDB

Excelente como componente de análisis local o por lotes en SpeedPy/Python, no como sustituto automático de la base transaccional.

## Principio

Separar base transaccional, object storage y analítica según sus responsabilidades. No introducir un servicio nuevo si una capacidad ya existente cubre el volumen real.
