# SpeedPy

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `python-data-institutional` |
| Uso predeterminado | Datos, documentos, formularios y procesos administrativos con Python |
| Repositorio | [https://github.com/speedpy/speedpy](https://github.com/speedpy/speedpy) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

SpeedPy se considera una de las bases más relevantes del catálogo porque muchos proyectos de la consultoría estarán dominados por Excel, CSV, documentos, formularios, padrones, reportes y reglas de negocio. Django + HTMX evita desplegar un frontend Next.js y un backend Python separados cuando la interfaz puede renderizarse en servidor.

## Qué ofrece el repositorio

- Django con UI server-rendered mediante HTMX, Alpine.js y Tailwind CSS.
- Django REST Framework y documentación OpenAPI.
- Equipos, roles e invitaciones.
- OAuth2 Provider, personal access tokens y scopes.
- Celery para tareas asíncronas.
- MFA y campos cifrados.
- Webhooks.
- MCP, CLI, AGENTS.md y recetas para agentes.
- Modo Docker con PostgreSQL, Redis y Celery; modo local más ligero.
- Checklist explícito de preparación para producción y contenido demo por retirar.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Carga, validación y transformación de Excel/CSV.
- Padrones de beneficiarios, alumnos, afiliados, empleados o clientes.
- Generación de reportes, constancias y documentos.
- Formularios institucionales y workflows de revisión.
- Inventarios y catálogos.
- Aplicaciones que usan pandas, Polars, DuckDB, openpyxl, Docling, OCR o ML.
- Sistemas multiusuario que requieren admin técnico y API.

## Ejemplos por tipo de cliente

- **Gobierno:** padrones, programas, conciliación de datos, solicitudes y reportes.
- **Escuela:** inscripciones, expedientes, calificaciones, importación de listas y documentos.
- **Sindicato:** afiliados, cuotas, beneficios, expedientes y generación de constancias.
- **Pyme:** inventarios, conciliaciones, importadores, reportes y automatización administrativa.

## Cuándo no usarlo

- Herramientas de una sola acción sin usuarios, persistencia ni crecimiento.
- Dashboards React muy ricos donde una API separada sea necesaria.
- Microservicios de inferencia o webhooks que deban escalar de forma independiente.
- Aplicaciones donde el principal reto sea concurrencia extrema y no datos o proceso.

## Ventajas estratégicas

- Un solo lenguaje para backend, datos, IA y documentos.
- Django Admin reduce el costo de soporte y operación interna.
- HTMX permite interfaces dinámicas sin SPA completa.
- Amplio ecosistema Python para archivos, documentos y análisis.
- Incluye más capacidades AI-friendly que muchos starters Django.

## Riesgos, madurez y límites

- Es un boilerplate SaaS completo y puede sobredimensionar herramientas pequeñas.
- La arquitectura principal de una sola app (`mainapp`) debe modularizarse para dominios grandes.
- El contenido demo debe eliminarse antes de producción.
- Billing, OAuth2 Provider, Celery y Redis no deben conservarse por defecto si no aportan valor.
- Su comunidad es menor que Django o FastAPI como ecosistemas generales.

## Relación con otras opciones del catálogo

- **Frente a Full Stack FastAPI:** SpeedPy para aplicación monolítica de datos/procesos; FastAPI para API-first y varios clientes.
- **Frente a TanStack + API:** SpeedPy reduce capas si HTMX cubre la UX.
- **Frente a GoShip:** SpeedPy gana en datos, documentos y administración; GoShip solo cuando Go tenga una ventaja operativa.
- **Frente a SpeedPy Lite:** usar el perfil Lite cuando equipos, billing, OAuth2, Redis o Celery no sean necesarios.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Crear dos perfiles: Full y Lite.
- [ ] Eliminar de forma automatizada módulos demo, billing y providers no elegidos.
- [ ] Reorganizar proyectos grandes por dominios: organizations, datasets, imports, workflows, reports, documents, audit e integrations.
- [ ] Agregar packs opcionales para pandas, Polars, DuckDB, openpyxl y Docling.
- [ ] Preparar español, accesibilidad y plantillas institucionales.
- [ ] Agregar auditoría y permisos por organización/departamento.
- [ ] Neutralizar despliegue para no depender de Appliku.

## Evaluación AI-friendly

**Alta.** El repositorio incluye AGENTS.md, skills y MCP. El riesgo es que la IA conserve módulos demo o mezcle billing con lógica institucional. El pack interno debe marcar capacidades obligatorias, opcionales y eliminables, además de invariantes de dominio.

## Despliegue y operación

- SQLite solo para local, demo o herramienta de un usuario.
- PostgreSQL para sistemas multiusuario.
- Celery y Redis únicamente cuando existan tareas que lo justifiquen.
- Docker para el perfil completo; `uv` para desarrollo o perfil Lite.
- Backups, migraciones y restore deben probarse antes de entrega.

## Decisión final

**Adoptado como base Python principal para datos y operación institucional**, sujeto a construir un fork curado y un perfil Lite.

## Fuentes oficiales

- [https://github.com/speedpy/speedpy](https://github.com/speedpy/speedpy)
- [https://docs.speedpy.com](https://docs.speedpy.com)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
