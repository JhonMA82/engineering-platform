# GoShip

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Candidato especializado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `go-realtime` |
| Uso predeterminado | Pilotos de turnos, notificaciones, SSE, PWA y webhooks en Go |
| Repositorio | [https://github.com/leomorpho/goship](https://github.com/leomorpho/goship) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

GoShip no entra como sexto boilerplate general. Se conserva como candidato especializado cuando Go ofrezca una ventaja medible en concurrencia, SSE, footprint o despliegue. Su documentación está en progreso, hereda elementos de Pagoda y de una aplicación real, y requiere saneamiento antes de usarse.

## Qué ofrece el repositorio

- Go + HTMX con servidor web y worker asíncrono.
- PostgreSQL, Redis y Ent ORM.
- Rutas y handlers sobre Echo.
- SSE para realtime.
- Notificaciones persistentes y programables.
- PWA y service worker.
- Playwright E2E, aunque el propio README señala contenido de prueba heredado.
- Kamal como ruta principal de despliegue documentada.
- El autor indica desarrollo activo y documentación WIP.
- Existe lógica heredada de Chérie, como amistad/perfiles, que puede eliminarse.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Sistemas de turnos y filas.
- Pantallas de llamado.
- Portales de notificaciones y alertas.
- Receptores concurrentes de webhooks.
- Servicios integradores con muchos eventos.
- PWA de campo, después de diseñar offline real.
- Consultas públicas simples de alta concurrencia.

## Ejemplos por tipo de cliente

- **Gobierno:** turnos, alertas, seguimiento en tiempo real o recepción de webhooks.
- **Escuela:** avisos y estado de procesos con actualizaciones en vivo.
- **Sindicato:** notificaciones y seguimiento, si la ventaja operativa de Go se justifica.
- **Pyme:** micro-SaaS concurrente, integración o monitor.

## Cuándo no usarlo

- Excel, OCR, reportes y documentos dominados por Python.
- CRUD institucional tradicional donde Django Admin reduzca trabajo.
- Dashboard React complejo.
- Introducir Go solo por rendimiento teórico sin benchmark.
- Proyectos donde el equipo no pueda mantener otro ecosistema.

## Ventajas estratégicas

- Go puede ofrecer operación compacta y concurrencia sencilla.
- HTMX evita SPA completa.
- SSE y notificaciones ya forman parte del diseño.
- Buena base conceptual para turnos y alertas.

## Riesgos, madurez y límites

- Documentación WIP; el README remite a Pagoda para gran parte de la arquitectura.
- Desarrollo activo y cambios después de probarse en una aplicación concreta.
- Código de dominio heredado que debe retirarse.
- Tests E2E heredados según el propio README.
- Introduce Go, Ent, HTMX y Kamal como ecosistema adicional.
- PWA no implica offline completo ni resolución de conflictos.

## Relación con otras opciones del catálogo

- **Frente a SpeedPy:** GoShip para concurrencia/realtime; SpeedPy para datos/documentos/admin.
- **Frente a FastAPI:** GoShip cuando Go demuestre mejor operación; FastAPI cuando IA/datos Python o API ecosystem sean decisivos.
- **Frente a TanStack:** GoShip puede renderizar UI con HTMX; TanStack para frontend React rico.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Piloto obligatorio con métricas.
- [ ] Eliminar lógica Chérie y pruebas heredadas.
- [ ] Crear documentación completa y AGENTS.md propio.
- [ ] Agregar organizaciones, auditoría, i18n y accesibilidad.
- [ ] Validar auth, permisos, storage y notificaciones.
- [ ] Definir despliegue alternativo a Kamal si se requiere.
- [ ] Benchmark frente a SpeedPy/FastAPI para el mismo caso.

## Evaluación AI-friendly

**Baja-media actualmente.** La documentación incompleta y el código heredado aumentan el riesgo de que un agente perpetúe supuestos del dominio original. No usar desarrollo intensivo por IA hasta tener mapa de arquitectura, glosario y tests actualizados.

## Despliegue y operación

- Contenedor y VPS como ruta inicial.
- Medir RAM, CPU, conexiones SSE y recuperación.
- Agregar observabilidad y runbook.
- Probar pérdida de conexión y reentrega de notificaciones.

## Decisión final

**Candidato especializado.** Solo promover después de un piloto exitoso en turnos, alertas o webhooks.

## Fuentes oficiales

- [https://github.com/leomorpho/goship](https://github.com/leomorpho/goship)
- [https://goship.run](https://goship.run)
- [https://github.com/mikestefanello/pagoda](https://github.com/mikestefanello/pagoda)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
