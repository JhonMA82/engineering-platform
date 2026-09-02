# Flujo: cambio en proyecto existente

Inicia con `/evolve-project` en Pi o lee el manifest y usa el CLI. Nunca vuelvas a ejecutar `bootstrap` sobre el proyecto.

## Ejemplo A — nueva librería
Necesitas exportar XLSX.
1. Leer manifest.
2. Ejecutar `eng plan --change-type <tipo>`.
3. Evaluar si ya existe capacidad.
4. Añadir librería únicamente al módulo correspondiente.
5. Registrar versión y rationale si es significativa.
6. Ejecutar los gates seleccionados y sus comandos reales.

`eng add <feature> --project .` muestra el impacto. Con `--apply` actualiza decisiones y handoff, pero la capacidad permanece `pending-implementation` hasta que Gentle la implemente y los gates la verifiquen.

## Ejemplo B — agregar una app móvil

El proyecto ya tiene TanStack Admin + API Starter y ahora necesita app:

1. Ejecutar `eng extend ignite --project .`.
2. Revisar destino, dependencias, archivos preservados y runtimes.
3. Confirmar con `eng extend ignite --project . --apply`.
4. Ejecutar `eng doctor --project .` y entregar el `GENTLE.md` actualizado.

Engineering agrega `apps/mobile`, conserva `services/api` y `apps/admin`, y actualiza el CI raíz. La integración de contratos, autenticación y navegación es el siguiente incremento de Gentle; no se declara resuelta automáticamente.

## Ejemplo C — aparece multi-tenancy
La solución pasa de 1 escuela a 20.
1. NO regenerar proyecto.
2. Actualizar intake, volver a resolver la Recipe y registrar ADR.
3. Crear plan de migración de datos.
4. Añadir organizations/memberships/tenant scope.
5. Ejecutar isolation suite.
6. Actualizar manifest y ejecutar `eng doctor`.
