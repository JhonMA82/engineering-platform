# SpeedPy Lite

## Motivo

SpeedPy completo incluye capacidades útiles para SaaS multiusuario, pero resulta excesivo para una herramienta sencilla que solo debe cargar un archivo, procesarlo, revisar resultados y descargar una salida.

## Perfil mínimo

```text
Django
HTMX
Tailwind CSS
pandas o Polars
SQLite o PostgreSQL
pruebas
Docker opcional
```

## Eliminar por defecto

- Billing.
- OAuth2 Provider.
- MCP público.
- Webhooks.
- Celery y Redis cuando las tareas sean breves.
- Multi-tenancy cuando sea una instalación para un solo equipo.
- Páginas de precios y contenido SaaS.

## Conservar

- Autenticación cuando existan datos sensibles.
- Formularios y validación Django.
- Django Admin para soporte.
- Auditoría básica.
- Importación/exportación.
- Separación por dominios.
- Preparación para elevar el proyecto al perfil Full.

## Casos

- Validador de Excel.
- Consolidador de CSV.
- Generador de reportes.
- Herramienta interna de una dependencia.
- Limpieza y conciliación de datos.
- Generación de documentos por lotes.

## Regla de evolución

No agregar Redis, Celery o multi-tenancy por anticipación. Incorporarlos cuando una métrica o requisito real lo exija.
