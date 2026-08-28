# AI Assistant Starter

## Base

Partir de Vercel Chatbot solo como upstream técnico. El starter interno debe convertir un chat genérico en una aplicación institucional gobernada.

## Capacidades obligatorias

- Autenticación.
- Organizaciones.
- Roles y permisos.
- Conversaciones por ámbito.
- Fuentes y citas.
- Ingesta documental versionada.
- Catálogo de herramientas.
- Confirmación humana para acciones sensibles.
- Registro de tool calls.
- Límites de uso y presupuesto.
- Feedback.
- Evaluaciones.
- Proveedores intercambiables.
- Retención y eliminación de datos.

## Herramientas

Cada herramienta debe declarar:

```yaml
name:
purpose:
input_schema:
required_role:
data_scope:
side_effects:
requires_confirmation:
timeout:
audit_fields:
test_cases:
```

## Seguridad

- Tratar documentos y resultados como contenido no confiable.
- Separar system prompt de contenido recuperado.
- No permitir que un documento modifique políticas.
- Validar inputs y outputs de herramientas.
- Aplicar autorización antes de ejecutar, no solo antes de mostrar el botón.
- Minimizar datos enviados a proveedores.

## Tipos de producto

- Asistente de reglamentos.
- Copiloto de servidor público.
- Asistente escolar.
- Consulta sindical.
- Generador de documentos.
- Mesa de ayuda.
- Agente de productividad conectado a sistemas.

## Modalidades

- Nube administrada.
- Proveedores directos.
- Infraestructura privada.
- Modelo local, cuando su calidad y costo sean suficientes.

## Métrica de éxito

No medir únicamente “se ve inteligente”. Medir exactitud con fuentes, tasa de resolución, errores, costo, latencia y acciones abortadas por seguridad.
