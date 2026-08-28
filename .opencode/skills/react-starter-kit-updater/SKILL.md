---
name: react-starter-kit-updater
description: Actualiza un fork de React Starter Kit usando su estrategia nativa merge-seed.
---

# React Starter Kit Updater

1. Confirma que el manifest usa `react-starter-kit`, `seed-fork` y un commit.
2. Lee el skill upstream `.agents/skills/merge-seed/SKILL.md`; sus instrucciones actuales gobiernan el merge.
3. Configura `merge.conflictstyle=zdiff3`, actualiza el remoto `seed` y crea una rama de actualización.
4. Preserva identidad, alcance y variables del proyecto; upstream posee mecanismos. No reemplaces migraciones aplicadas.
5. Resuelve conflictos por intención y registra patches locales persistentes.
6. Ejecuta el set upstream: instalación congelada, typecheck, lint, tests y build; agrega gates de la Recipe.
7. Solo después actualiza el pin y la ficha de curación.
