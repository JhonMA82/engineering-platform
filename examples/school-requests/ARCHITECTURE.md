# Arquitectura: school-requests

Documento generado desde `GP-02@1.0.0`. El proyecto es single-tenant; si varias escuelas comparten instalación se vuelve a resolver el intake y se agrega multitenancy mediante una migración explícita.

## Idea

Sistema interno para registrar, adjuntar y aprobar solicitudes escolares con trazabilidad de responsables y decisiones.

## Stack

- `tanstack-admin` (`curated`, overlay) fijado a `e6e5d3b`.
- `hono-api` (`released`, internal) fijado a `platform-0.5.0`.
- Datos: `postgresql-managed`.
- Features: auth, RBAC, audit, observability, files y notifications.
- Estado de este fixture documental: `blueprint`; `eng bootstrap` materializa ambos starters.

## Gates

Lint, types, tests, integración, migración, build, seguridad y backup/restore.

## Fuera de alcance

- multitenancy;
- jobs;
- webhooks;
- mobile y desktop.
