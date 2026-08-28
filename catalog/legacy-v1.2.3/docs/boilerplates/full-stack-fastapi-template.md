# Full Stack FastAPI Template

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Recomendado** |
| Procedencia | Recomendado por el asistente |
| Categoría | `python-api-first` |
| Uso predeterminado | Aplicaciones API-first con backend Python, frontend React y varios clientes |
| Repositorio | [https://github.com/fastapi/full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Este template cubre un hueco que SpeedPy no debe forzar: aplicaciones API-first donde un backend Python será consumido por varios clientes o donde la separación frontend/backend sea una decisión de arquitectura. No sustituye a SpeedPy; ambos se seleccionan según la frontera del sistema.

## Qué ofrece el repositorio

- Template oficial bajo la organización FastAPI.
- FastAPI, SQLModel, Pydantic y PostgreSQL.
- Frontend React + TypeScript + Vite.
- Tailwind CSS y shadcn/ui.
- Cliente frontend generado automáticamente desde OpenAPI.
- Playwright para E2E.
- Docker Compose y configuración de despliegue.
- Copier para generar proyectos y estructura de skills para agentes.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- API para web, móvil y escritorio.
- Backends Python con contrato OpenAPI formal.
- Sistemas que integran múltiples consumidores.
- Aplicación React donde el backend de datos/IA debe ser independiente.
- Plataforma que crecerá hacia servicios separados.
- Proyecto que necesita cliente TypeScript generado y pruebas E2E desde el inicio.

## Ejemplos por tipo de cliente

- **Gobierno:** API central para portal, app móvil e integraciones.
- **Escuela:** backend compartido por administración web y app de alumnos.
- **Sindicato:** servicios de afiliados consumidos por web y móvil.
- **Pyme:** plataforma B2B o producto con API para terceros.

## Cuándo no usarlo

- Aplicación de formularios y reportes que Django + HTMX resuelve con menos capas.
- Servicio aislado pequeño donde el Python Service Starter sea suficiente.
- Proyectos sin equipo o necesidad para mantener frontend y backend separados.
- Usarlo solo porque FastAPI es moderno.

## Ventajas estratégicas

- Fuente oficial y arquitectura API-first.
- Tipos y contrato OpenAPI entre backend y frontend.
- Pruebas backend/frontend/E2E contempladas.
- Docker Compose y automatización de proyecto.
- Base adecuada para combinar con Ignite o Tauri.

## Riesgos, madurez y límites

- Dos aplicaciones y dos toolchains aumentan coordinación.
- JWT y auth inicial no sustituyen un modelo completo de sesiones, revocación y RBAC.
- SQLModel debe validarse en dominios complejos.
- No incluye automáticamente multi-tenancy institucional, auditoría ni workers de negocio.
- Traefik y la topología incluida pueden ser excesivos para proyectos pequeños.

## Relación con otras opciones del catálogo

- **Frente a SpeedPy:** API-first y múltiples clientes vs. monolito productivo de datos y procesos.
- **Frente a Python Service Starter:** aplicación completa vs. servicio pequeño.
- **Frente a TanStack Admin:** el template trae frontend propio; si se usa la UI TanStack del catálogo, hay que evitar mantener dos frontends.
- **Frente a Open SaaS:** FastAPI favorece Python y control de API; Open SaaS acelera producto comercial full-stack TypeScript.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Crear perfil full-stack y perfil backend-only.
- [ ] Agregar módulos opcionales de organizaciones, permisos y auditoría.
- [ ] Definir estrategia de tokens, sesiones y revocación.
- [ ] Incluir OpenTelemetry, logging estructurado y métricas.
- [ ] Alinear el frontend con la familia Admin o retirar el frontend demo.
- [ ] Agregar jobs solo cuando sea necesario.
- [ ] Documentar generación de cliente y evitar modelos duplicados.

## Evaluación AI-friendly

**Alta potencialmente.** OpenAPI, Pydantic y cliente generado proporcionan contratos claros. Los agentes deben modificar primero el contrato backend y regenerar el cliente, no copiar tipos manualmente. Los generated files deben marcarse como no editables.

## Despliegue y operación

- Fijar commit o versión generada por Copier.
- Cambiar todos los secretos y dominios de ejemplo.
- Validar Docker en VPS y plataforma administrada.
- Separar escalado de frontend, API y workers solo si el tráfico lo exige.

## Decisión final

**Recomendado como base principal para proyectos API-first Python.** No reemplaza a SpeedPy ni debe añadirse a todo sistema.

## Fuentes oficiales

- [https://github.com/fastapi/full-stack-fastapi-template](https://github.com/fastapi/full-stack-fastapi-template)
- [https://fastapi.tiangolo.com](https://fastapi.tiangolo.com)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
