# SpeedPy vs. Full Stack FastAPI Template

## Pregunta decisiva

¿La aplicación es un sistema Python completo con UI sencilla, o una API que será consumida por varios clientes?

## SpeedPy

Elegir cuando:

- Datos, Excel, documentos y formularios dominan.
- Django Admin aporta valor.
- HTMX cubre la interfaz.
- Un solo despliegue reduce costo.
- El equipo desea usar Python de extremo a extremo.

## Full Stack FastAPI

Elegir cuando:

- La API es un producto.
- Habrá web, móvil, desktop o terceros.
- Se necesita cliente TypeScript generado.
- La UI React requiere separación clara.
- El backend debe escalar o desplegarse independientemente.

## No hacer

- Agregar Next/TanStack delante de SpeedPy por reflejo.
- Convertir cada aplicación en microservicios.
- Duplicar validaciones entre Pydantic, Zod y formularios sin contrato.
- Usar FastAPI para CRUD institucional solo por modernidad.

## Combinación válida

SpeedPy puede conservar el sistema principal y extraer un Python Service para OCR o inferencia. No es obligatorio migrar todo a FastAPI.
