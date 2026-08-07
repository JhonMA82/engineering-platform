# Security Baseline

Siempre: secrets fuera repo, dependencias fijadas, input validation, secure headers, logging sin secretos, backups según criticidad, least privilege. Auth→sesiones/revocación/CSRF. Tenant→backend authorization + scope + isolation tests. Files→MIME/size/signed URLs. Webhooks→firma/replay/idempotencia.
