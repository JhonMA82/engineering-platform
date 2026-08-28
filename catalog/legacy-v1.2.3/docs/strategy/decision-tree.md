# Árbol de decisión

```mermaid
flowchart TD
    A[1. Naturaleza del entregable] --> B{¿Contenido público, SEO o documentación?}
    B -->|Sí| Stardrive[Stardrive]
    B -->|No| C{¿Aplicación instalada?}
    C -->|Móvil| Ignite[Ignite]
    C -->|Escritorio o local| Tauri[Tauri UI]
    C -->|No: web operativa| D[2. Datos y backend]
    D --> E{¿Dominado por Excel, documentos, formularios y reglas Python?}
    E -->|Pocas pantallas| Lite[SpeedPy Lite]
    E -->|Sistema multiusuario que crecerá| Full[SpeedPy Full]
    D --> F{¿API consumida por web, móvil, desktop o terceros?}
    F -->|Sí| FastAPI[Full Stack FastAPI Template]
    D --> G{¿Capacidad aislada como OCR, inferencia o webhook?}
    G -->|Sí| PyService[Python Service Starter]
    D --> H[3. Frontend administrativo]
    H --> I{¿Privada, operativa y consume un backend independiente?}
    I -->|Sí| TanStack[TanStack Shadcn Admin Dashboard]
    I -->|Sitio público + privado, Next aporta| Next[Next Shadcn Admin Dashboard]
    H --> J[4. Producto comercial: auth, pagos, jobs, correo y archivos]
    J -->|Hosting convencional, aceptar Wasp| OpenSaaS[Open SaaS<br/>evaluar con piloto]
    J -->|TypeScript puro, edge de Cloudflare| RSK[React Starter Kit]
    J --> K[5. IA]
    K -->|Interfaz principal: asistente personalizado| AIAssistant[AI Assistant Starter]
    K -->|Laboratorio privado local| SelfHosted[Self-hosted AI Starter Kit<br/>solo POC]
    K --> L[6. Especialización: ventaja medible de Go en SSE o concurrencia]
    L -->|Sí| GoShip[GoShip<br/>pilotar]
```

> [!IMPORTANT]
> Pregunta final obligatoria: ¿esta elección reduce complejidad total, o solo mueve la complejidad a otro framework? Si no puede responderse, crear un ADR y hacer un spike antes de adoptar.

## 1. Naturaleza del entregable

### ¿El valor principal es contenido público, SEO o documentación?

Usar **Stardrive**.

### ¿Es una aplicación instalada?

- Móvil: **Ignite**.
- Escritorio/local: **Tauri UI**.

### ¿Es una aplicación web operativa?

Continuar.

## 2. Datos y backend

### ¿El problema está dominado por Excel, documentos, formularios y reglas Python?

- Pocas pantallas y poca infraestructura: **SpeedPy Lite**.
- Sistema multiusuario que crecerá: **SpeedPy Full**.

### ¿La API será consumida por web, móvil, desktop o terceros?

Usar **Full Stack FastAPI Template** o su perfil backend-only.

### ¿Es una capacidad aislada como OCR, inferencia o webhook?

Usar **Python Service Starter**.

## 3. Frontend administrativo

### ¿La aplicación es privada, operativa y consume un backend independiente?

**TanStack Shadcn Admin Dashboard**.

### ¿Sitio público y app privada deben convivir y Next aporta capacidades específicas?

**Next Shadcn Admin Dashboard**.

## 4. Producto comercial

### ¿Se requiere auth, pagos, jobs, correo y archivos desde el inicio?

- Hosting convencional y aceptar Wasp: evaluar **Open SaaS** mediante piloto.
- Stack TypeScript puro y aceptar edge de Cloudflare: **React Starter Kit**.

## 5. IA

### ¿La interfaz principal será un asistente personalizado?

Usar **AI Assistant Starter**, derivado de Vercel Chatbot.

### ¿Se requiere un laboratorio privado local?

Usar **Self-hosted AI Starter Kit** solo para POC.

## 6. Especialización

### ¿Existe una ventaja medible de Go en SSE, concurrencia o footprint?

Pilotar **GoShip**.
