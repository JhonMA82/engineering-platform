# Upgrade Recipe — ejemplo real

React Starter Kit declara `seed-fork` y `merge-seed`.

1. `detect`: compara el pin del manifest con upstream.
2. `plan`: lee el skill nativo `.agents/skills/merge-seed/SKILL.md` y revisa el rango.
3. `apply`: crea rama, fetch del remoto `seed` y merge por intención.
4. `preserve`: mantiene identidad, alcance y migraciones aplicadas.
5. `verify`: instalación congelada, types, lint, tests, build y gates de la Recipe.
6. `record`: evidencia, patches persistentes y pin nuevo.

La Recipe completa está en [`upgrades/react-starter-kit/merge-seed`](../../upgrades/react-starter-kit/merge-seed/README.md). Si falla verify, el pin no cambia.
