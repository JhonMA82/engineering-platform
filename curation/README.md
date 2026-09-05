# Curation adapters

Un adapter registra cómo convertir una entrada evaluada en un artefacto reproducible: pin upstream, materializador genérico, overlay opcional, checks y actualización. `evidence.json` documenta su auditoría AI-friendly.

Los tipos soportados son `git-copy`, `local-copy`, `command-generator` y `git-generator`. El motor no contiene condiciones por nombre de boilerplate. Para agregar o retirar entradas usa `eng boilerplate add|remove`; ambos operan en dry-run salvo `--apply`.

Los paths `*.md` declarados en cada `evidence.json` se exponen en el handoff (`composition.starter_docs` y sección «Documentación por starter» de `GENTLE.md`) para que Gentle lea los límites y contratos de cada starter en su destino.
