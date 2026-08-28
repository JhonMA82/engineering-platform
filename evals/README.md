# Evals

Los evals separan dos niveles:

- `recipe-resolution.json` prueba que un intake produzca la arquitectura correcta y mínima.
- `boilerplate-curation.json` prueba duplicado, refresh y candidata.
- Los directorios de tareas describen cambios que se ejecutarán contra fixtures de starters cuando estos se liberen.

Los casos de decisión ya forman parte de `python -m unittest discover -s tests -v`. Los evals de modificación permanecen documentados hasta que exista un fixture de código liberado; no se reportan como ejecutados.
