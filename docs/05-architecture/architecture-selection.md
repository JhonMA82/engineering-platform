# Selección arquitectónica

| Caso | Recipe | Señal decisiva |
|---|---|---|
| Blog municipal | GP-01 / Stardrive | contenido y SEO |
| Solicitudes de una escuela | GP-02 / TanStack + Hono | operación privada y API |
| Limpiar 50 Excel | GP-03 / SpeedPy | Python y datos dominan |
| Inspectores en campo | GP-04 / Ignite | capacidades nativas |
| Procesador local de PDF | GP-05 / Tauri | filesystem y offline |
| Web + móvil | GP-06 | varios clientes, dominio compartido |
| SaaS con billing | GP-07 / React Starter Kit | cobro y organizaciones reales |
| Asistente con citas y tools | GP-08 | IA es el producto y necesita gobierno |

Si dos caminos sirven, elige el de menor complejidad operacional. Una señal no basta para activar todos los packs: `billing` justifica GP-07, pero no jobs o webhooks ajenos al flujo de cobro.

Los datos se seleccionan dentro de los perfiles permitidos por la Recipe. Turso/libSQL necesita ventaja edge o sync y un piloto explícito; no sustituye PostgreSQL por default.
