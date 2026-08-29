# Versionado

```text
engineering-platform 0.4.3
Project Recipe GP-02 1.0.0
starter release       versión o commit exacto
feature pack          versión propia al liberarse
```

- MAJOR: contrato incompatible o Recipe que exige migración.
- MINOR: capacidad compatible, nueva Recipe o nueva entrada.
- PATCH: corrección sin cambio de selección.

El manifest registra versión de plataforma y Recipe. Un starter externo registra commit/release solo cuando fue observado; `null` significa que todavía es blueprint, nunca “usar latest”. Cada promoción o cambio de pin actualiza changelog, adapter, tests y Upgrade Recipe aplicable.
