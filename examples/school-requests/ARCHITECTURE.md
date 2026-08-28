# Arquitectura: school-requests

Documento generado desde `GP-02@1.0.0`. El proyecto es single-tenant; si varias escuelas comparten instalación se vuelve a resolver el intake y se agrega multitenancy mediante una migración explícita.

## Idea

Sistema interno para registrar, adjuntar y aprobar solicitudes escolares con trazabilidad de responsables y decisiones.

## Stack

- `tanstack-admin` (`pilot-ready`, overlay).
- `hono-api` (`catalog-only`, internal).
- Datos: `postgresql-managed`.
- Features: auth, RBAC, audit, observability, files y notifications.
- Estado: `blueprint` hasta que los starters tengan releases internos reproducibles.

## Gates

Lint, types, tests, integración, migración, build, seguridad y backup/restore.

## Fuera de alcance

- multitenancy;
- jobs;
- webhooks;
- mobile y desktop.
