# Flujo: cambio en proyecto existente

## Ejemplo A — nueva librería
Necesitas exportar XLSX.
1. Leer manifest.
2. Evaluar si ya existe capacidad.
3. Añadir librería únicamente al módulo correspondiente.
4. Registrar versión y rationale si es significativa.
5. Tests.

## Ejemplo B — aparece multi-tenancy
La solución pasa de 1 escuela a 20.
1. NO regenerar proyecto.
2. Ejecutar detector del feature pack `multitenancy`.
3. Crear plan de migración de datos.
4. Añadir organizations/memberships/tenant scope.
5. Ejecutar isolation suite.
6. Actualizar manifest + ADR.
