# Cuándo no agregar una API separada

No crear Hono/FastAPI solo por estandarización cuando:
- Stardrive puede ser completamente estático;
- SpeedPy/Django+HTMX cubre frontend y backend con menos complejidad;
- Tauri puede operar localmente;
- una automatización de n8n no necesita dominio propio;
- un script o proceso batch es suficiente.

Agregar una API cuando exista al menos una razón real: múltiples clientes, integración externa, boundary de seguridad, escalado separado, contrato público o ciclo de despliegue independiente.
