# Tauri vs. PWA vs. Ignite

| Necesidad | Opción |
|---|---|
| Web instalable básica | PWA |
| Móvil con capacidades nativas y tiendas | Ignite |
| Escritorio con filesystem y procesos locales | Tauri UI |

## PWA primero cuando

- La app es principalmente formularios y consulta.
- No se requieren APIs nativas complejas.
- Se desea una sola distribución.
- El offline es limitado y controlable.

## Ignite cuando

- Push, cámara, sensores o experiencia móvil dedicada son centrales.
- Se requiere publicación en tiendas.
- La operación en campo es importante.

## Tauri cuando

- Se procesan archivos locales.
- Se requiere instalador de escritorio.
- Hay integración con sistema operativo.
- El usuario no debe operar terminal o entorno Python.

## Advertencia

Offline no aparece automáticamente en ninguna opción. Debe diseñarse almacenamiento, sincronización, conflicto y recuperación.
