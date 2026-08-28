# Flujo de proyecto nuevo

## 1. Intake

Parte de `templates/project-intake.json`. Describe tipo, señales, features, exclusiones, datos y restricciones. No elijas framework en el intake.

## 2. Resolución

```bash
./eng recommend --input intake.json
```

Revisa Recipe, estado, starters, base permitida, skills, gates, exclusiones y warnings. Una desviación se decide aquí, no después de generar.

## 3. Blueprint

```bash
./eng new --from intake.json --output my-project
./eng doctor --project my-project
```

Se crean manifest, intake, arquitectura y reglas para agentes. El comando no clona ramas sin pin ni convierte fichas en código. `scaffold_status: blueprint` significa que el equipo aún debe usar o construir el adapter liberado correspondiente.

## 4. Materialización

Solo un adapter `released` puede materializar automáticamente. Debe registrar upstream, pin, licencia, modo de integración, ownership, checks y Upgrade Recipe. Después se ejecutan gates vacíos y se crea la primera vertical de dominio.

## 5. Entrega y aprendizaje

Cada cambio usa `eng plan` y `eng check`. Los incidentes repetibles y la tercera repetición de una solución se evalúan para knowledge entry, canonical example, feature pack, skill o starter interno.
