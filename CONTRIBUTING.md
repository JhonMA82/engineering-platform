# Contributing

Un cambio entra a la plataforma cuando reduce trabajo repetido sin trasladar complejidad innecesaria a todos los proyectos.

## Comandos

```bash
make check
./eng boilerplate evaluate <url>
./eng recommend --input <intake.json>
```

## Cambio normal

- Mantén el alcance pequeño y sin datos de cliente.
- Añade o actualiza prueba si cambia resolución, curación o manifest.
- Registra ADR solo si cambia una decisión arquitectónica.
- Incluye Upgrade Recipe si proyectos existentes deben migrar.
- Actualiza `CHANGELOG.md`.

## Alta o actualización de boilerplate

Usa `.github/ISSUE_TEMPLATE/propose.yml` y el skill `boilerplate-curator`. La PR debe incluir evidencia de licencia, mantenimiento, instalación, tests/build, comparación y estrategia upstream. Actualiza `platform/boilerplates.json`, ficha, eval y changelog en una misma operación.

No promociones directamente a `released`. La secuencia normal es `catalog-only` → `pilot-ready` → `curated` → `released`, y cada paso exige evidencia adicional.
