# Catálogo de Boilerplates para Consultoría

![CI](https://img.shields.io/github/actions/workflow/status/JhonMA82/boilerplates-catalog/validate.yml?label=validaci%C3%B3n)
![Estrellas](https://img.shields.io/github/stars/JhonMA82/boilerplates-catalog?label=estrellas)
![Último commit](https://img.shields.io/github/last-commit/JhonMA82/boilerplates-catalog?label=%C3%BAltimo%20commit)
![Versión del catálogo](https://img.shields.io/badge/cat%C3%A1logo-1.2.3-blue)

Catálogo versionado de bases técnicas para una consultoría enfocada en desarrollo a medida, automatización, IA aplicada y optimización de procesos para gobiernos, escuelas, sindicatos y pymes.

> **Versión:** 1.2.3 · **Revisión:** 2026-08-05

Este repositorio conserva no solo una lista de enlaces, sino el razonamiento de selección: casos de uso, límites, comparaciones, riesgos, madurez, curación, AI-friendly, despliegue y relación entre opciones.

## ✅ Decisiones centrales

1. No usar Next.js por defecto.
2. TanStack Start es el dashboard operativo preferido.
3. SpeedPy es la base principal para datos, documentos y procesos Python.
4. FastAPI cubre API-first y múltiples clientes; no sustituye a SpeedPy.
5. GoShip es especializado y requiere piloto.
6. Open SaaS se reserva para producto comercial y necesita validar Wasp/tenancy.
7. Vercel Chatbot es upstream de un starter propio, no solución institucional terminada.
8. El activo más importante por construir es Institutional Operations Starter.
9. Cada upstream debe convertirse en un pack curado AI-friendly.
10. La arquitectura mínima adecuada tiene prioridad sobre acumular frameworks.

> [!IMPORTANT]
> Ningún estado del catálogo evita la revisión por proyecto: antes de producción se exige commit fijado, licencia revisada, threat model, pruebas, despliegue reproducible y plan de actualización.

## 📦 Catálogo

<!-- CATALOG_TABLE_START -->
| Estado | Opción | Uso principal | Upstream | Ficha |
|---|---|---|---|---|
| Seleccionado | **Stardrive** | Landing pages, blogs, documentación y sitios públicos | [Repositorio](https://github.com/peltmonger/stardrive) | [Detalle](docs/boilerplates/stardrive.md) |
| Seleccionado | **Ignite** | Aplicaciones móviles React Native | [Repositorio](https://github.com/infinitered/ignite) | [Detalle](docs/boilerplates/ignite.md) |
| Seleccionado | **Tauri UI** | Aplicaciones de escritorio y herramientas locales | [Repositorio](https://github.com/agmmnn/tauri-ui) | [Detalle](docs/boilerplates/tauri-ui.md) |
| Seleccionado | **Next Shadcn Admin Dashboard** | Productos híbridos con sitio público y aplicación privada sobre Next.js | [Repositorio](https://github.com/arhamkhnz/next-shadcn-admin-dashboard) | [Detalle](docs/boilerplates/next-shadcn-admin-dashboard.md) |
| Seleccionado | **TanStack Shadcn Admin Dashboard** | Aplicaciones operativas privadas y frontends conectados a APIs | [Repositorio](https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard) | [Detalle](docs/boilerplates/tanstack-shadcn-admin-dashboard.md) |
| Seleccionado | **SpeedPy** | Datos, documentos, formularios y procesos administrativos con Python | [Repositorio](https://github.com/speedpy/speedpy) | [Detalle](docs/boilerplates/speedpy.md) |
| Recomendado | **Full Stack FastAPI Template** | Aplicaciones API-first con backend Python, frontend React y varios clientes | [Repositorio](https://github.com/fastapi/full-stack-fastapi-template) | [Detalle](docs/boilerplates/full-stack-fastapi-template.md) |
| Piloto recomendado | **Open SaaS** | Productos SaaS comerciales con auth, billing, jobs, correo y archivos | [Repositorio](https://github.com/wasp-lang/open-saas) | [Detalle](docs/boilerplates/open-saas.md) |
| Seleccionado | **React Starter Kit** | Productos SaaS full-stack TypeScript con despliegue en el edge de Cloudflare | [Repositorio](https://github.com/kriasoft/react-starter-kit) | [Detalle](docs/boilerplates/react-starter-kit.md) |
| Recomendado | **Vercel Chatbot** | Base técnica de un asistente o copiloto de IA personalizado | [Repositorio](https://github.com/vercel/chatbot) | [Detalle](docs/boilerplates/vercel-chatbot.md) |
| Candidato especializado | **GoShip** | Pilotos de turnos, notificaciones, SSE, PWA y webhooks en Go | [Repositorio](https://github.com/leomorpho/goship) | [Detalle](docs/boilerplates/goship.md) |
| Infraestructura/POC | **Self-hosted AI Starter Kit** | Pruebas de concepto de automatización e IA privada autoalojada | [Repositorio](https://github.com/n8n-io/self-hosted-ai-starter-kit) | [Detalle](docs/boilerplates/self-hosted-ai-starter-kit.md) |
| Starter interno | **Consulting Admin Family** | Mantener una familia coherente con variantes TanStack Start y Next.js | Por crear | [Detalle](docs/internal-starters/consulting-admin-family.md) |
| Starter interno | **SpeedPy Lite** | Herramientas internas de datos con pocas pantallas y menor infraestructura | Por crear | [Detalle](docs/internal-starters/speedpy-lite.md) |
| Starter interno | **Institutional Operations Starter** | Trámites, expedientes, solicitudes, organizaciones, auditoría y operación institucional | Por crear | [Detalle](docs/internal-starters/institutional-operations-starter.md) |
| Starter interno | **AI Assistant Starter** | Asistentes institucionales con fuentes, herramientas, auditoría y control de costos | Por crear | [Detalle](docs/internal-starters/ai-assistant-starter.md) |
| Starter interno | **Python Service Starter** | OCR, inferencia, webhooks, documentos e integraciones como servicio aislado | Por crear | [Detalle](docs/internal-starters/python-service-starter.md) |
<!-- CATALOG_TABLE_END -->

## 🧭 Navegación esencial

- [Contexto de la consultoría](docs/strategy/consultancy-context.md)
- [Mapa del catálogo](docs/strategy/catalog-map.md)
- [Árbol de decisión](docs/strategy/decision-tree.md)
- [Arquitecturas compuestas](docs/strategy/architecture-compositions.md)
- [Casos por sector](docs/strategy/sector-use-cases.md)
- [Estándar AI-friendly](docs/strategy/ai-friendly-standard.md)
- [Mapa de cobertura de la conversación](docs/strategy/coverage-map.md)
- [Regla contra el sobrediseño](docs/strategy/anti-overengineering.md)
- [Servicios de consultoría](docs/strategy/service-offerings.md)
- [Política de selección](docs/governance/selection-policy.md)
- [Estrategia de forks y upstream](docs/governance/upstream-and-forks.md)

## ⚖️ Comparaciones clave

- [Next.js vs. TanStack Start](docs/comparisons/next-vs-tanstack.md)
- [SpeedPy vs. Full Stack FastAPI](docs/comparisons/speedpy-vs-fastapi.md)
- [SpeedPy vs. GoShip](docs/comparisons/speedpy-vs-goship.md)
- [Open SaaS vs. dashboards Admin](docs/comparisons/open-saas-vs-admin.md)
- [Open SaaS vs. React Starter Kit](docs/comparisons/open-saas-vs-react-starter-kit.md)
- [Tauri vs. PWA vs. Ignite](docs/comparisons/tauri-vs-pwa-vs-ignite.md)
- [Opciones de IA](docs/comparisons/ai-options.md)

## 🧩 Plataformas complementarias

n8n, Activepieces, Directus, Appsmith, Metabase, Dify, RAGFlow, Docling, MinIO y PostgreSQL se documentan aparte porque no todas son boilerplates:

[Ver índice de plataformas complementarias](docs/platforms/README.md)

## Skills de OpenCode

- [github-readme](.opencode/skills/github-readme/SKILL.md) — redacta y mejora README de GitHub en español, con foco en boilerplates y starters públicos.

## 🛠️ Mantenimiento

```bash
python scripts/generate_readme.py
python scripts/validate_catalog.py
```

GitHub Actions ejecuta la validación en push y pull request.

## 🗂️ Estructura

```text
.
├── README.md
├── AGENTS.md
├── catalog.json
├── docs/
│   ├── boilerplates/
│   ├── internal-starters/
│   ├── comparisons/
│   ├── strategy/
│   ├── governance/
│   └── platforms/
├── templates/
├── scripts/
├── .github/
└── .opencode/
    └── skills/
```

## 📄 Alcance

> [!NOTE]
> El catálogo no contiene copias de los upstream ni reemplaza sus documentos oficiales. Antes de cada proyecto se vuelve a verificar versión, licencia, seguridad y compatibilidad.
