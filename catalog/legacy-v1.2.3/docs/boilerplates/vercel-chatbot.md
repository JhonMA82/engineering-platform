# Vercel Chatbot

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Recomendado** |
| Procedencia | Recomendado por el asistente |
| Categoría | `ai-application` |
| Uso predeterminado | Base técnica de un asistente o copiloto de IA personalizado |
| Repositorio | [https://github.com/vercel/chatbot](https://github.com/vercel/chatbot) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Vercel Chatbot no se adopta como producto terminado, sino como base técnica para construir un AI Assistant Starter propio. El valor está en AI SDK, UI conversacional, tool calling y persistencia; el trabajo importante de la consultoría será gobernanza, fuentes, permisos, evaluación y neutralidad de proveedor.

## Qué ofrece el repositorio

- Next.js App Router.
- AI SDK para texto, objetos estructurados y tool calls.
- Hooks para chat y UI generativa.
- shadcn/ui y Tailwind CSS.
- Persistencia de chat y almacenamiento de archivos mediante servicios incluidos en la plantilla.
- Auth.js.
- Soporte para múltiples proveedores a través de AI Gateway y posibilidad de usar proveedores directos.
- Pruebas E2E en el repositorio.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Asistente de reglamentos, convenios o manuales.
- Copiloto para empleados.
- Mesa de ayuda conectada con herramientas.
- Generación guiada de documentos.
- Chat con datos internos y acciones controladas.
- Interfaces generativas para procesos complejos.

## Ejemplos por tipo de cliente

- **Gobierno:** asistente de procedimientos o apoyo interno, con citas y permisos.
- **Escuela:** consulta de reglamentos, orientación y soporte.
- **Sindicato:** convenios, prestaciones, estatutos y seguimiento de gestiones.
- **Pyme:** soporte, ventas asistidas o copiloto de operación.

## Cuándo no usarlo

- RAG serio sin pipeline de ingestión, calidad y evaluación.
- Agentes autónomos que ejecutan acciones sensibles sin confirmación.
- Tareas donde un formulario o búsqueda estructurada sea mejor que chat.
- Clientes que requieren infraestructura local sin adaptar storage, auth y modelos.

## Ventajas estratégicas

- UI completa y extensible.
- Abstracción de proveedores mediante AI SDK.
- Tool calling y datos estructurados.
- Punto de partida fuerte para asistentes personalizados.
- Ecosistema activo.

## Riesgos, madurez y límites

- La configuración por defecto puede acercar el proyecto a servicios Vercel.
- No trae organizaciones, RBAC institucional, presupuestos, citas o evaluación completos.
- Prompt injection y tool abuse deben tratarse como amenazas.
- Costos, retención y privacidad dependen de proveedor y configuración.
- Una interfaz convincente puede ocultar respuestas no confiables.

## Relación con otras opciones del catálogo

- **Frente a Dify:** código altamente personalizado vs. plataforma configurable.
- **Frente a RAGFlow:** aplicación conversacional general vs. enfoque documental especializado.
- **Frente a Self-hosted AI Starter Kit:** UI/producto vs. infraestructura local.
- **Frente al AI Assistant Starter interno:** el repo externo es base; el starter interno será el producto gobernado.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Abstraer modelos, storage, auth y base de datos.
- [ ] Agregar organizaciones, roles y políticas por herramienta.
- [ ] Incluir fuentes, citas y trazabilidad.
- [ ] Registrar tool calls, inputs, resultados y aprobaciones.
- [ ] Implementar límites, presupuestos y observabilidad de tokens.
- [ ] Construir datasets de evaluación y feedback.
- [ ] Protección frente a prompt injection y exfiltración.
- [ ] Preparar español y accesibilidad.

## Evaluación AI-friendly

**Media-alta como base de producto.** El reto no es que un agente programe el chat; es impedir que introduzca herramientas sin política, mezcle prompts con reglas de negocio o acople el sistema a un proveedor. El starter interno debe tener catálogo de herramientas y pruebas deterministas.

## Despliegue y operación

- Ofrecer modalidad nube y modalidad privada cuando sea viable.
- No depender de AI Gateway si el cliente exige proveedor directo.
- Cifrar y limitar retención de conversaciones.
- Separar ambientes y datos de evaluación.
- Agregar rate limiting y protección de abuso.

## Decisión final

**Recomendado únicamente como upstream del AI Assistant Starter interno.**

## Fuentes oficiales

- [https://github.com/vercel/chatbot](https://github.com/vercel/chatbot)
- [https://ai-sdk.dev](https://ai-sdk.dev)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
