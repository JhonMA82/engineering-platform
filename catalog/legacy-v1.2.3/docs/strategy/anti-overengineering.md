# Regla contra el sobrediseño

## Síntoma

Elegir la solución “más completa” antes de entender el proceso.

## Ejemplos a evitar

- Next.js + FastAPI para subir un CSV y descargar otro.
- Celery + Redis para una tarea de dos segundos.
- Multi-tenancy para una instalación de un solo cliente.
- App móvil cuando una PWA cubre el flujo.
- Go por rendimiento sin benchmark.
- Microservicios sin despliegue o escalado independiente.
- RAG para datos que deberían consultarse con SQL.
- Chat para una tarea mejor resuelta con formulario.

## Estrategia

1. Modelar el proceso.
2. Identificar datos, usuarios y canales.
3. Elegir la base mínima.
4. Definir umbrales de crecimiento.
5. Agregar capacidades cuando se alcance el umbral.

## Preguntas

- ¿Qué complejidad elimina?
- ¿Qué complejidad introduce?
- ¿Quién la operará?
- ¿Cómo se actualiza?
- ¿Cómo se abandona?
- ¿Qué métrica justificaría escalar?
