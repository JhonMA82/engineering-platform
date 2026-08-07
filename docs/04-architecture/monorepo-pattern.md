# Patrón multi-app

Inspirado conceptualmente en create-t3-turbo, no adoptado como boilerplate.

```text
apps/{web,api,mobile,desktop,worker}
packages/{contracts,sdk,auth,authorization,database,config,ui,testing}
```

Generar solo apps requeridas. Compartir contratos/SDK, no código específico de framework. Seguridad en backend.
