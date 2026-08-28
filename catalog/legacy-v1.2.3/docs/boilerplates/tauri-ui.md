# Tauri UI

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `desktop` |
| Uso predeterminado | Aplicaciones de escritorio y herramientas locales |
| Repositorio | [https://github.com/agmmnn/tauri-ui](https://github.com/agmmnn/tauri-ui) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Tauri UI cubre el espacio de aplicaciones de escritorio y herramientas locales. Es especialmente útil para usuarios no técnicos que necesitan instalar una aplicación, trabajar con archivos o ejecutar procesos localmente. El scaffolder permite elegir el frontend, por lo que no conviene fijar Next.js por defecto.

## Qué ofrece el repositorio

- Scaffolder de aplicaciones Tauri con shadcn/ui.
- Defaults orientados a escritorio y baterías opcionales.
- Elección de frontend compatible con el flujo del generador.
- Base Tauri para integrar capacidades nativas mediante Rust y plugins.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Procesamiento local de archivos Excel, CSV, audio, video o documentos.
- Herramientas offline para oficinas con conectividad limitada.
- Clientes de escritorio de sistemas internos.
- Aplicaciones para usuarios no técnicos que requieren instalador.
- Utilidades que necesitan filesystem, ventanas, bandeja, updater o integración nativa.
- Empaquetado de soluciones Python mediante servicio local o sidecar, cuando esté justificado.

## Ejemplos por tipo de cliente

- **Gobierno:** herramienta local de captura o procesamiento en una dependencia.
- **Escuela:** utilidad de control o generación de documentos en equipos administrativos.
- **Sindicato:** gestor local de expedientes o credenciales.
- **Pyme:** conversor, gestor documental, app de operación o cliente de un servicio.

## Cuándo no usarlo

- Sitios públicos.
- Aplicaciones web que no requieren capacidades locales.
- Proyectos donde una PWA satisface instalación y offline con menor costo.
- Equipos que no puedan mantener Rust, firma, empaquetado y actualizaciones.

## Ventajas estratégicas

- Binarios más ligeros que stacks basados en Chromium embebido en muchos casos.
- Permite reutilizar habilidades web sin renunciar a capacidades nativas.
- El scaffolder evita encerrar todos los proyectos en un solo frontend.
- Buena opción para distribución a usuarios no técnicos.

## Riesgos, madurez y límites

- La superficie de seguridad depende de permisos, comandos y plugins Tauri.
- Firma, autoactualización y artefactos por sistema operativo requieren disciplina.
- Frontend y Rust forman dos capas que los agentes deben comprender.
- Empaquetar Python o modelos locales aumenta tamaño y soporte.

## Relación con otras opciones del catálogo

- **Frente a Ignite:** Tauri UI es la base de escritorio; Ignite continúa como base móvil.
- **Frente a PWA:** elegir Tauri cuando archivos, sistema operativo o distribución nativa aporten valor.
- **Frontend recomendado:** TanStack Start para apps operativas; Vite/React simple para utilidades pequeñas; Astro solo para interfaces predominantemente estáticas.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Crear presets de permisos mínimos.
- [ ] Agregar updater, logging y diagnóstico como módulos opcionales.
- [ ] Preparar CI de artefactos para Windows, Linux y macOS según alcance real.
- [ ] Documentar patrón de comandos Tauri y validación de inputs.
- [ ] Definir almacenamiento local, migraciones y respaldo.
- [ ] Crear un ejemplo de integración segura con servicio Python local.

## Evaluación AI-friendly

**Media-alta.** Los agentes necesitan un mapa explícito de límites: frontend no accede libremente al sistema; toda capacidad nativa pasa por comandos tipados y permisos mínimos. Debe prohibirse generar comandos genéricos de shell o filesystem.

## Despliegue y operación

- Seleccionar únicamente sistemas operativos comprometidos contractualmente.
- Firmar artefactos cuando el canal lo exija.
- Probar actualización, rollback y migraciones de datos.
- Entregar logs diagnósticos sin exponer datos sensibles.

## Decisión final

**Adoptado como scaffolder de escritorio**, no como plantilla única e inmutable. El frontend se decide por proyecto.

## Fuentes oficiales

- [https://github.com/agmmnn/tauri-ui](https://github.com/agmmnn/tauri-ui)
- [https://github.com/tauri-apps/tauri](https://github.com/tauri-apps/tauri)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
