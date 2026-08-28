# Open SaaS

| Campo | Decisión |
|---|---|
| Estado en el catálogo | **Recomendado con piloto** |
| Procedencia | Recomendado por el asistente |
| Categoría | `commercial-saas` |
| Uso predeterminado | Productos SaaS comerciales con auth, billing, jobs, correo y archivos |
| Repositorio | [https://github.com/wasp-lang/open-saas](https://github.com/wasp-lang/open-saas) |
| Revisión de fuentes | 2026-08-02 |

## Tesis de adopción

Open SaaS se incorpora porque cubre un escenario distinto: lanzar productos SaaS comerciales con autenticación, pagos, correo, jobs y archivos ya integrados. No debe usarse como base institucional genérica y tampoco debe venderse como multi-tenant empresarial sin diseñar organizaciones, aislamiento y RBAC.

## Qué ofrece el repositorio

- Wasp como framework full-stack.
- React, Node.js, Prisma y PostgreSQL.
- Autenticación por correo y proveedores sociales.
- Pagos mediante proveedores soportados por el proyecto.
- Correo, background jobs, landing page y carga S3.
- shadcn/ui.
- AGENTS.md, skills y plugin orientado a Claude Code.
- Licencia MIT.

> [!NOTE]
> Esta sección resume capacidades declaradas u observables en las fuentes oficiales. No implica que cada capacidad esté lista para las políticas de producción de la consultoría.

## Casos de uso donde encaja

- Micro-SaaS.
- Producto de suscripción.
- Herramienta de IA comercial con planes o créditos.
- Portal B2B que necesita auth, billing, jobs y archivos desde el inicio.
- Producto propio de la consultoría.
- MVP comercial donde la infraestructura SaaS repetitiva sea el mayor costo inicial.

## Ejemplos por tipo de cliente

- **Gobierno:** rara vez como base principal; solo si el proyecto realmente es un servicio multiinstitución con modelo comercial.
- **Escuela:** producto SaaS vendido a múltiples planteles, después de diseñar tenancy.
- **Sindicato:** plataforma ofrecida a múltiples organizaciones, con aislamiento validado.
- **Pyme:** software por suscripción, portal B2B o producto vertical.

## Cuándo no usarlo

- Sistema interno single-tenant sin billing.
- Proyecto centrado en datos Python o documentos.
- Equipo que no quiera adoptar Wasp y comprender su código generado.
- Asumir que auth equivale a organizaciones, tenancy y permisos empresariales.
- Infraestructura pública donde el cliente exija stack estándar sin capa adicional.

## Ventajas estratégicas

- Acelera capacidades comerciales repetitivas.
- Full-stack TypeScript con ORM y jobs.
- Buenas bases para agentes de programación.
- Permite concentrarse antes en propuesta de valor del producto.
- Licencia permisiva.

## Riesgos, madurez y límites

- Wasp agrega lenguaje declarativo, CLI y ciclo de actualización propio.
- La depuración requiere entender qué es fuente y qué es generado.
- El propio proyecto reconoce que pueden faltar capacidades.
- Multi-tenancy y RBAC avanzados deben diseñarse explícitamente.
- Billing y providers aumentan superficie de seguridad y compliance.

## Relación con otras opciones del catálogo

- **Frente a Next/TanStack Admin:** Open SaaS incluye backend y capacidades comerciales; los dashboards son principalmente UI.
- **Frente a Institutional Operations Starter:** Open SaaS para producto comercial; starter institucional para operación y trazabilidad.
- **Frente a SpeedPy:** TypeScript/SaaS comercial vs. Python/datos/proceso.
- **Frente a FastAPI:** Wasp acelera full-stack; FastAPI da control y ecosistema Python.

## Curación necesaria antes de usarlo en proyectos reales

- [ ] Ejecutar un piloto antes de declararlo adoptado.
- [ ] Documentar a fondo Wasp, generación y actualización.
- [ ] Diseñar organizations, memberships, tenant isolation y auditoría.
- [ ] Crear variante sin billing para casos internos, solo si sigue teniendo sentido usar Wasp.
- [ ] Agregar pruebas de webhooks y pagos.
- [ ] Verificar deploy fuera del camino feliz.
- [ ] Preparar compatibilidad con OpenCode, no solo Claude Code.

## Evaluación AI-friendly

**Alta en intención.** AGENTS.md y skills son una ventaja. Sin embargo, los agentes deben comprender el DSL de Wasp y evitar editar output generado. Se requiere un mapa de generación, comandos y límites del framework.

## Despliegue y operación

- Probar el proveedor objetivo antes de comprometerse.
- Staging obligatorio para billing, correo y webhooks.
- Respaldos independientes de PostgreSQL y archivos.
- Separar secretos de auth, pagos, correo y storage.
- Documentar una salida o migración si Wasp deja de ser conveniente.

## Decisión final

**Recomendado con piloto previo para SaaS comercial.** No es el default para gobiernos, escuelas o sindicatos internos.

## Fuentes oficiales

- [https://github.com/wasp-lang/open-saas](https://github.com/wasp-lang/open-saas)
- [https://github.com/wasp-lang/wasp](https://github.com/wasp-lang/wasp)
- [https://docs.opensaas.sh](https://docs.opensaas.sh)

---

[Volver al catálogo](../../README.md) · [Ver árbol de decisión](../strategy/decision-tree.md)
