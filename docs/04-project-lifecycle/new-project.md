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

Se valida la definición, se resuelve la Recipe, se materializan los starters en un staging temporal, se instalan dependencias, se ejecutan checks y solo entonces se publica el destino. Una desviación se decide aquí, no después de generar.

## 3. Verificación

```bash
eng doctor --project .
eng check --project . --run
```

`scaffold_status: materialized` confirma que el código existe. `readiness: verified` confirma además que pasaron los checks; `code-ready` exige ejecutar `eng check --run`.
El setup ya aprobado se reutiliza en comprobaciones posteriores; usa `eng check --project . --run --force-setup` solo cuando necesites reinstalar dependencias.

## 4. Materialización

Cada starter default tiene adapter, pin, licencia, destino, setup y checks. Los externos se copian desde el commit exacto; los internos desde `starters/`; Ignite usa su generador versionado. Si algo falla, el destino original queda intacto.

## 5. Gentle y aprendizaje

Gentle lee `GENTLE.md` y decide entre desarrollo directo o SDD. Cada cambio usa `eng plan` y `eng check`. Los incidentes repetibles y la tercera repetición de una solución se evalúan para knowledge entry, canonical example, feature pack, skill o starter interno.

La ruta manual con `templates/project-intake.json`, `eng recommend` y `eng new` se conserva para automatización y CI sin conversación.
