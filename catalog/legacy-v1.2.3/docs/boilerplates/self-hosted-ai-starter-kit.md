# Self-hosted AI Starter Kit

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Pack de infraestructura** |
| Procedencia | Recomendado por el asistente |
| Categoría | `private-ai-infrastructure` |
| Uso predeterminado | Pruebas de concepto de automatización e IA privada autoalojada |
| Repositorio | [https://github.com/n8n-io/self-hosted-ai-starter-kit](https://github.com/n8n-io/self-hosted-ai-starter-kit) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Este repositorio no se clasifica como boilerplate de aplicación, sino como pack de infraestructura para demostrar IA y automatización privadas. Es útil para prototipos y laboratorios, pero el propio proyecto advierte que no está completamente optimizado para producción.

## Qué ofrece el repositorio

- Docker Compose.
- n8n autoalojado.
- Ollama para modelos locales.
- Qdrant como vector store.
- PostgreSQL.
- Perfiles para CPU y GPUs compatibles.
- Workflow de demostración.
- La documentación oficial lo presenta como starter para comenzar, no como entorno productivo endurecido.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Prueba de concepto de IA privada.
- Automatización local.
- Resumen seguro de documentos en laboratorio.
- RAG básico con modelos locales.
- Demostración para clientes con requisitos de privacidad.
- Entorno interno para experimentar con agentes y workflows.

## Ejemplos por tipo de cliente

- **Gobierno:** laboratorio aislado antes de una arquitectura productiva.
- **Escuela:** análisis local de documentos no sensibles durante un piloto.
- **Sindicato:** pruebas de consulta privada de convenios y expedientes.
- **Pyme:** automatizaciones locales y evaluación de costos de modelos.

## Cuándo no usarlo

- Producción crítica sin hardening.
- Alta disponibilidad sin rediseño.
- Cliente sin capacidad de operar Docker, backups y modelos.
- Presentarlo como aplicación final.
- Asumir que autoalojado equivale automáticamente a seguro.

## Ventajas estratégicas

- Arranque rápido de componentes compatibles.
- Permite experimentar sin enviar todo a un proveedor externo.
- n8n aporta integraciones y orquestación visual.
- Separa modelo y vector store.

## Riesgos, madurez y límites

- No está optimizado para producción según su README.
- Los workflows visuales pueden quedar fuera de control de versiones.
- Modelos locales exigen recursos, parches y monitoreo.
- Secrets, TLS, SSO, backups y aislamiento deben añadirse.
- La licencia y modelo de uso de cada componente deben revisarse.

## Relación con otras opciones del catálogo

- **Frente a Vercel Chatbot:** infraestructura vs. aplicación.
- **Frente a Dify/RAGFlow:** composición base vs. plataformas especializadas.
- **Frente a un servicio Python:** laboratorio low-code vs. código versionado y probado.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Reverse proxy, TLS, SSO y gestión de secretos.
- [ ] Backups y prueba de restauración.
- [ ] Exportación/versionado de workflows.
- [ ] Ambientes separados.
- [ ] Monitoreo de CPU, GPU, disco y colas.
- [ ] Registro de modelos, embeddings y versiones.
- [ ] Runbooks de actualización y recuperación.

## Evaluación AI-friendly

**Media.** Los workflows deben exportarse y acompañarse de documentación, pruebas y restricciones. Un agente no debe poder modificar herramientas o credenciales sin revisión.

## Despliegue y operación

- Usar solo en red controlada durante piloto.
- No exponer n8n, Qdrant u Ollama directamente a internet.
- Dimensionar almacenamiento de modelos y vectores.
- Separar datos del cliente por ambiente.

## Decisión final

**Mantener como pack de infraestructura para POC y laboratorio.** La producción requiere una arquitectura derivada, no el Compose sin cambios.

## Fuentes oficiales

- [https://github.com/n8n-io/self-hosted-ai-starter-kit](https://github.com/n8n-io/self-hosted-ai-starter-kit)
- [https://github.com/n8n-io/n8n](https://github.com/n8n-io/n8n)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
