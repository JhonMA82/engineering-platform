# Estrategia de upstream, forks y packs

## No trabajar directamente sobre upstream

```text
upstream oficial
       ↓
upstream-sync
       ↓
consulting-base
       ↓
tag interno probado
       ↓
proyecto cliente
```

## Por qué

- Mantener historial limpio.
- Reaplicar cambios internos.
- Probar actualizaciones antes de proyectos.
- Evitar que cada cliente sea un fork irreconciliable.
- Conservar parches de seguridad.

## Qué pertenece al pack

- Saneamiento de demos.
- Arquitectura.
- Idioma.
- Accesibilidad.
- Observabilidad.
- CI.
- AGENTS.md.
- Generadores.
- Seguridad.
- Deployment.
- Módulos opcionales.

## Qué pertenece al proyecto cliente

- Dominio.
- Marca.
- Integraciones.
- Datos.
- Políticas.
- Requerimientos específicos.

## Actualización

1. Abrir rama de sync.
2. Integrar upstream.
3. Ejecutar tests y migraciones.
4. Revisar breaking changes.
5. Actualizar documentación.
6. Publicar tag interno.
7. Migrar clientes de forma planificada.
