# Flujo de proyecto nuevo

## Ruta principal: Pi

Si la carpeta aún no existe:

```bash
eng start my-project
```

El launcher crea `~/dev/my-project`, cambia a ese directorio y abre Pi. Si Pi ya está abierto dentro de una carpeta vacía usa `/new-project`. Nunca se crea una carpeta de proyecto dentro de otra sesión o repositorio no vacío.

## 1. Descubrimiento

El skill `project-discovery` pregunta progresivamente por problema, usuarios, resultados, alcance, datos, permisos y restricciones. No pide un framework. Antes de escribir solicita confirmación explícita y guarda `.engineering/project-definition.json`.

## 2. Resolución y bootstrap

```bash
eng bootstrap --from .engineering/project-definition.json --output .
```

Se valida la definición, se resuelve la Recipe y se crean manifest, arquitectura, reglas para agentes y handoff a Gentle. Una desviación se decide aquí, no después de generar.

## 3. Verificación

```bash
eng doctor --project .
```

`scaffold_status: blueprint` significa que el equipo aún debe usar o construir el adapter liberado correspondiente. El comando no clona ramas sin pin ni convierte fichas en código.

## 4. Materialización

Solo un adapter `released` puede materializar automáticamente. Debe registrar upstream, pin, licencia, modo de integración, ownership, checks y Upgrade Recipe. Después se ejecutan gates vacíos y se crea la primera vertical de dominio.

## 5. Gentle y aprendizaje

Gentle lee `GENTLE.md` y decide entre desarrollo directo o SDD. Cada cambio usa `eng plan` y `eng check`. Los incidentes repetibles y la tercera repetición de una solución se evalúan para knowledge entry, canonical example, feature pack, skill o starter interno.

La ruta manual con `templates/project-intake.json`, `eng recommend` y `eng new` se conserva para automatización y CI sin conversación.
