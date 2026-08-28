# Arquitecturas compuestas recomendadas

## Sitio público + operación interna

```text
Stardrive
    │ enlaces / auth
TanStack Admin
    │ API
SpeedPy o FastAPI
    │
PostgreSQL + S3/MinIO
```

## Aplicación de datos en un solo stack

```text
SpeedPy
├── Django/HTMX
├── pandas/Polars
├── PostgreSQL
├── Django Admin
└── Celery opcional
```

Elegir esta composición antes de separar React y Python cuando la UI sea sencilla.

## Plataforma multicanal

```text
TanStack Admin ─┐
Ignite ─────────┼── FastAPI ── PostgreSQL
Tauri UI ───────┤       ├── workers
Integraciones ──┘       └── S3/MinIO
```

## SaaS comercial con capacidad Python especializada

```text
Open SaaS
├── auth / billing / app
└── Python Service
    ├── IA
    ├── OCR
    └── procesamiento
```

## IA institucional

```text
AI Assistant Starter
├── auth / roles / audit
├── retrieval + citations
├── tool gateway
└── provider adapters
    ├── cloud
    └── private
```

## Escritorio de procesamiento local

```text
Tauri UI
├── frontend TanStack/Vite
├── comandos Tauri
└── sidecar Python opcional
    ├── documentos
    └── datos
```

## Realtime especializado

```text
GoShip pilot
├── SSE
├── notifications
├── worker
└── PostgreSQL/Redis
```

Solo después de comparar contra una implementación SpeedPy/FastAPI suficiente.
