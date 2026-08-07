# Flujo: crear o actualizar una skill

Crear una skill cuando existe conocimiento específico que el agente necesita repetidamente y que no debe vivir en AGENTS.md global.

1. Definir trigger.
2. Definir stacks/features compatibles.
3. Escribir instrucciones mínimas y ejemplos canónicos.
4. Agregar guard o eval si la skill puede afectar arquitectura.
5. Registrar en `skills/registry.json`.
6. Probar contra fixture.
7. Medir si mejora corrección/contexto antes de promoverla.

Una skill no debe cargar documentación completa de una herramienta si solo necesita un patrón concreto.
