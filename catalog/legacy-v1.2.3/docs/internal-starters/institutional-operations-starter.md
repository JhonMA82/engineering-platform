# Institutional Operations Starter

## Por qué es el activo estratégico principal

Gobiernos, escuelas, sindicatos y pymes comparten capacidades operativas: organizaciones, usuarios, áreas, solicitudes, estados, adjuntos, comentarios, fechas límite, auditoría y reportes. Reutilizar estas capacidades puede ahorrar más que mantener diez dashboards.

El objetivo **no** es crear un mega-framework que modele todos los sectores. Se construirá por extracción progresiva desde proyectos reales.

## Núcleo inicial

```text
organizations
memberships
departments
roles-and-permissions
requests
workflows
assignments
comments
files
deadlines
audit
notifications
exports
```

## Capacidades

### Organización

- Instituciones.
- Dependencias o departamentos.
- Sedes y planteles.
- Usuarios y membresías.
- Roles y permisos.
- Configuración por organización.

### Operación

- Solicitudes y trámites.
- Expedientes.
- Formularios.
- Estados y transiciones.
- Asignaciones.
- Adjuntos.
- Comentarios internos.
- Fechas límite.

### Gobernanza

- Bitácora de auditoría.
- Historial de cambios.
- Exportación.
- Retención.
- Privacidad.
- Permisos por dato o acción cuando sea necesario.

### Comunicación

- Notificaciones internas.
- Correo.
- WhatsApp/SMS como adapters opcionales.
- Plantillas.
- Recordatorios.

### Analítica

- Tiempos de respuesta.
- Carga por área.
- Indicadores.
- Reportes.
- Excel/PDF.

## Interfaces posibles

- **SpeedPy/HTMX:** para sistemas donde backend, datos y UI sencilla deben vivir juntos.
- **TanStack Start:** para experiencia operativa React rica.
- **Next.js:** solo cuando también exista producto público integrado.
- **Ignite/Tauri:** clientes adicionales sobre el mismo contrato.

## Multi-tenancy

La instalación **single-tenant** debe ser una opción de primera clase. No imponer multi-tenancy a todos los clientes. Cuando se active, el aislamiento debe probarse en consultas, archivos, jobs, caché y auditoría.

## Primer hito

Extraer el núcleo desde un proyecto real de solicitudes o expedientes. No construir módulos sectoriales hipotéticos antes de validar el flujo.
