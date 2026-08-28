# Automatización: n8n y Activepieces

## Decisión inicial

**n8n es el estándar inicial de automatización de la consultoría.** Activepieces queda como alternativa cuando su arquitectura TypeScript, extensibilidad o modelo operativo encajen mejor.

## Casos

- Sincronizar sistemas.
- Procesar correos y archivos.
- Enviar notificaciones.
- Programar tareas.
- Integrar CRMs, formularios, storage y mensajería.
- Orquestar llamadas a servicios de IA.
- Prototipar una automatización antes de convertir la parte crítica en código.

## n8n

Ventajas:

- Amplio catálogo de integraciones.
- Nodos de IA.
- Autoalojamiento.
- Combinación de flujo visual y código.
- Ecosistema y documentación.

Riesgos:

- Reglas críticas ocultas en workflows.
- Credenciales y datos distribuidos.
- Cambios manuales sin revisión.
- Flujos difíciles de probar.
- Modelo de licencia que debe revisarse para el servicio ofrecido.

## Activepieces

Evaluar cuando:

- Se prefiera TypeScript para conectores.
- Se necesiten piezas propias.
- Su modelo de despliegue y licencia resulte más conveniente.
- El equipo valore una experiencia más orientada a desarrolladores.

## Política de versionado

- Exportar workflows.
- Guardarlos en Git.
- Separar ambientes.
- Evitar credenciales embebidas.
- Añadir descripción, propietario, inputs, outputs y manejo de errores.
- Probar con datos sintéticos.
- Migrar a código cuando el workflow sea demasiado crítico o complejo.

## Regla de arquitectura

n8n coordina; no debe convertirse en la única implementación de reglas de negocio sensibles, permisos o contabilidad.
