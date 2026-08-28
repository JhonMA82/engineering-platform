# Python Service Starter

## Rol

Starter pequeño para capacidades que merecen un proceso, despliegue o escalado independiente. No es un sustituto de SpeedPy ni una copia reducida del Full Stack FastAPI Template.

## Perfiles

### Minimal

- FastAPI.
- Pydantic.
- pytest.
- logging estructurado.
- health/readiness.
- Docker.

### Database

- PostgreSQL.
- SQLAlchemy/Alembic o decisión documentada.
- transacciones.
- pruebas de integración.

### Worker

- Cola elegida por necesidad.
- retries idempotentes.
- dead-letter.
- métricas.

## Casos

- OCR.
- Parsing documental.
- Inferencia.
- Embeddings.
- Webhooks.
- Integración con sistemas.
- Generación de PDF/Excel.
- Procesamiento pesado.

## Contratos

- OpenAPI.
- Errores estandarizados.
- Timeouts.
- Límites de tamaño.
- Idempotency keys.
- Autenticación service-to-service.
- Cliente generado cuando exista consumidor TypeScript.

## Regla anti-microservicios

No extraer un servicio si no tiene al menos una de estas razones:

- Escalado independiente.
- Dependencias incompatibles.
- Aislamiento de seguridad.
- Ciclo de despliegue independiente.
- Propietario claro.
- Carga asíncrona o hardware específico.
