# Ejemplo completo — Sistema de solicitudes internas para una escuela

Este ejemplo sirve como guía de referencia para juniors.

## 1. Requerimiento inicial
“Una escuela recibe solicitudes internas por WhatsApp y hojas de cálculo. Quiere que empleados registren solicitudes y que el director pueda aprobarlas o rechazarlas.”

## 2. Intake
- una institución;
- 30 usuarios;
- uso en navegador;
- no app móvil;
- no pagos;
- no múltiples escuelas;
- necesita historial de aprobación.

## 3. Architecture Brief
**Golden Path:** GP-02.  
**Frontend:** TanStack Admin.  
**Backend:** Hono API.  
**DB:** PostgreSQL.  
**Features:** auth, rbac, audit.  
**No usar:** multitenancy, jobs, Redis, webhooks, mobile, Tauri.

## 4. Por qué no multi-tenancy
Solo hay una escuela. Agregar organizations/memberships/tenant scope aumentaría schema, tests y permisos sin resolver un requisito actual.

## 5. Manifest
Ver `examples/school-requests/.engineering/project.json`.

## 6. Primera feature: Request
Campos:
- id;
- title;
- description;
- status: draft/submitted/approved/rejected;
- createdBy;
- createdAt;
- approvedBy/approvedAt opcionales.

Roles:
- CAPTURIST: crea y edita sus borradores;
- DIRECTOR: lee y aprueba/rechaza;
- ADMIN: administra usuarios, no obtiene “manage all” implícito sobre todo negocio.

## 7. Flujo con harness
Tarea: “Implementa creación de solicitudes”.

El harness carga:
- project manifest;
- GP-02;
- Hono skill;
- TanStack skill;
- auth + authorization;
- canonical endpoint/form examples.

Plan esperado:
1. schema/migration;
2. domain/use case;
3. HTTP contract;
4. policy;
5. repository;
6. tests;
7. TanStack form;
8. E2E básico.

## 8. Quality Gates
- typecheck;
- migration desde DB vacía;
- create request integration test;
- policy test: capturist puede crear;
- OpenAPI route documentada;
- UI build.

## 9. Segunda feature: aprobación
El backend valida `request.approve`. La UI oculta el botón a quien no tiene permiso, pero eso solo es UX.

Audit event:
```text
actor: director-123
action: request.approved
resource: request-456
before: submitted
after: approved
requestId: req-...
```

## 10. Cambio seis meses después: 20 escuelas
Ahora sí aparece multi-tenancy.

No regenerar.

Proceso:
1. crear ADR;
2. ejecutar detector del feature pack;
3. planear organizations/memberships;
4. migrar datos actuales a organization `school-001`;
5. hacer organizationId obligatorio en recursos;
6. tenant-scoped repositories;
7. suite de aislamiento A/B;
8. selector de organización si aplica;
9. actualizar manifest.

## 11. Cambio posterior: integración con ERP
Agregar feature pack `webhooks`.
No rehacer auth ni tenancy.

## 12. Incidente de ejemplo
Bug: CAPTURIST pudo llamar directamente `POST /requests/:id/approve`.

Respuesta correcta:
1. prueba de regresión que reproduce 403 faltante;
2. arreglar policy backend;
3. ejecutar permission matrix;
4. registrar knowledge entry;
5. revisar si canonical example/guard debe mejorar.

## 13. Qué aprende un junior
- seleccionar antes de programar;
- no sobrediseñar;
- manifest evita redescubrir arquitectura;
- harness reduce contexto;
- seguridad vive en backend;
- tests/guards deciden Done;
- el proyecto puede evolucionar sin regenerarse.
