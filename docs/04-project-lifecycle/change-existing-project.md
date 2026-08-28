# Flujo: cambio en proyecto existente

## Ejemplo A — nueva librería
Necesitas exportar XLSX.
1. Leer manifest.
2. Ejecutar `eng plan --change-type <tipo>`.
3. Evaluar si ya existe capacidad.
4. Añadir librería únicamente al módulo correspondiente.
5. Registrar versión y rationale si es significativa.
6. Ejecutar los gates seleccionados y sus comandos reales.

## Ejemplo B — aparece multi-tenancy
La solución pasa de 1 escuela a 20.
1. NO regenerar proyecto.
2. Actualizar intake, volver a resolver la Recipe y registrar ADR.
3. Crear plan de migración de datos.
4. Añadir organizations/memberships/tenant scope.
5. Ejecutar isolation suite.
6. Actualizar manifest y ejecutar `eng doctor`.
