# Ignite

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Seleccionado** |
| Procedencia | Propuesto por el usuario |
| Categoría | `mobile` |
| Uso predeterminado | Aplicaciones móviles React Native |
| Repositorio | [https://github.com/infinitered/ignite](https://github.com/infinitered/ignite) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Ignite se mantiene como base móvil porque evita reconstruir desde cero la arquitectura recurrente de React Native. Debe elegirse cuando una aplicación instalada aporta valor real —notificaciones, cámara, distribución, capacidades nativas o experiencia móvil dedicada— y no simplemente porque el cliente pidió una app.

## Qué ofrece el repositorio

- Boilerplate React Native mantenido por Infinite Red.
- CLI para crear proyectos y generadores de componentes, modelos y otros elementos.
- TypeScript y una arquitectura móvil opinionada.
- Soporte y documentación para flujos modernos de React Native.
- Base de pruebas y tooling para proyectos reales.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Aplicaciones para ciudadanos, alumnos, socios o empleados.
- Captura de campo con cámara, geolocalización o archivos.
- Notificaciones push.
- Aplicaciones que consumen APIs institucionales.
- Experiencias con operación parcial offline y sincronización diseñada.
- Productos móviles comerciales para pymes.

## Ejemplos por tipo de cliente

- **Gobierno:** reporte de incidencias, inspecciones o consulta de trámites.
- **Escuela:** app para alumnos, tutores, horarios, avisos y tareas.
- **Sindicato:** credencial digital, beneficios, avisos y seguimiento de gestiones.
- **Pyme:** fuerza de ventas, servicio en campo, inventario móvil o app de cliente.

## Cuándo no usarlo

- Herramientas internas que una web responsive o PWA resuelve con menos costo.
- Proyectos sin presupuesto para mantener builds, tiendas y versiones de sistema operativo.
- Duplicar exactamente una aplicación web sin una estrategia de canal.
- Asumir que React Native resuelve por sí solo sincronización offline.

## Ventajas estratégicas

- Arquitectura probada y generadores que aceleran el inicio.
- Reduce decisiones básicas repetitivas.
- Facilita mantener consistencia entre proyectos móviles.
- Permite integrar módulos nativos cuando el caso lo exige.

## Riesgos, madurez y límites

- El ecosistema React Native exige actualizaciones periódicas y pruebas por plataforma.
- Permisos, notificaciones, deep links y publicación requieren trabajo operativo.
- Una app móvil agrega soporte de versiones, distribución y observabilidad.
- La seguridad del dispositivo no reemplaza autorización en el backend.

## Relación con otras opciones del catálogo

- **Frente a PWA:** preferir PWA cuando no se necesiten capacidades nativas fuertes ni distribución en tiendas.
- **Frente a Tauri:** Ignite es móvil; Tauri es escritorio y, aunque puede abarcar móvil, no es la opción predeterminada aquí.
- **Frente a TanStack/Next:** el frontend móvil debe consumir contratos de API compartidos, no duplicar reglas de negocio.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Definir preset Expo/EAS y preset nativo según requerimientos.
- [ ] Agregar autenticación, configuración por ambientes y observabilidad.
- [ ] Crear módulo opcional offline con cola, reintentos y resolución de conflictos.
- [ ] Preparar diseño, errores y validaciones en español.
- [ ] Incluir CI de build y pruebas mínimas Android/iOS.
- [ ] Documentar almacenamiento seguro y manejo de secretos.

## Evaluación AI-friendly

**Media-alta.** La estructura y generadores ayudan, pero los agentes deben recibir reglas estrictas sobre navegación, estado, llamadas API, código nativo, permisos, offline y pruebas en dispositivos. Un agente no debe introducir módulos nativos sin justificar su mantenimiento.

## Despliegue y operación

- Definir desde el inicio canales de desarrollo, staging y producción.
- Automatizar versionado y release notes.
- No incluir secretos de backend en el bundle.
- Probar dispositivos de gama media y escenarios de red deficiente.

## Decisión final

**Adoptado como boilerplate móvil principal**, condicionado a demostrar que una app instalada supera a una PWA o web responsive.

## Fuentes oficiales

- [https://github.com/infinitered/ignite](https://github.com/infinitered/ignite)
- [https://docs.infinite.red/ignite-cli](https://docs.infinite.red/ignite-cli)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
