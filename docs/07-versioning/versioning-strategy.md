# Versionado

```text
engineering-platform 0.8.0 (candidata a v1.0)
Project Recipe GP-02 1.0.0
starter release       versión o commit exacto
feature pack          versión propia al liberarse
```

- MAJOR: contrato incompatible o Recipe que exige migración.
- MINOR: capacidad compatible, nueva Recipe o nueva entrada.
- PATCH: corrección sin cambio de selección.

El manifest registra versión de plataforma y Recipe. Un starter externo registra commit/release solo cuando fue observado; `null` significa que todavía es blueprint, nunca “usar latest”. Cada promoción o cambio de pin actualiza changelog, adapter, tests y Upgrade Recipe aplicable.

La candidata 0.8 congela los contratos de definición, manifest, materialización, capacidades y extensión. Para promover el mismo código a v1.0 se exige una ventana de uso real, pilotos verdes de los caminos principales y ausencia de correcciones incompatibles; después, las mejoras ordinarias deben entrar como minor o patch.
