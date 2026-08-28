---
name: github-readme
description: Escribe y mejora README de GitHub en español, con foco en boilerplates y starters públicos. Úsalo cuando el usuario pida crear, redactar, generar, mejorar, auditar, corregir o enriquecer el README de un repositorio, aunque no lo llame por su nombre (ej. "escribe la documentación del repo", "el readme está incompleto", "hazme el readme del proyecto", "documenta este boilerplate", "el readme está muy plano, agrega badges, iconos o gráficos"). También cuando el README existente tenga secciones faltantes, comandos desactualizados, enlaces rotos o formato que no renderiza en GitHub, y cuando haya que aplicar mecanismos como badges de shields.io, alerts, diagramas Mermaid, emojis o task lists.
---

# GitHub README en español

## Por qué existe este skill

Un README mal hecho daña la adopción de un boilerplate: promete capacidades que no existen, lista comandos que ya no funcionan o no renderiza en GitHub. Este skill produce README que dicen la verdad sobre el repo, muestran cómo empezar en 60 segundos y se mantienen al día con el código. El objetivo no es la longitud ni los adornos: es que un visitante entienda qué es, si le sirve y cómo lo usa sin adivinar nada.

## Principios que gobiernan todo el trabajo

- **Veracidad primero.** Solo afirmar lo que se puede verificar en el repositorio (manifiestos, código, docs, CI). Nunca inventar características, screenshots, badges ni URLs.
- **Sin inflar.** Sin lluvia de badges, sin emojis (salvo que el usuario los pida), sin secciones de relleno ni frases genéricas de marketing. Cada elemento visual debe aportar información real.
- **Español claro**, conservando los nombres técnicos oficiales (nombres de paquetes, comandos, términos del ecosistema).
- **Respetar el contexto del repo.** Leer AGENTS.md, convenciones existentes y el idioma de los docs antes de decidir el tono.
- **Cambios mínimos al mejorar.** No reescribir por reescribir: conservar lo válido y corregir lo roto.
- **No eliminar matices.** Un documento puede hacerse más legible, pero no hay que recortar información o matices para que quede más corto. Si una sección sobra, decirlo y esperar confirmación antes de quitarla.

## Mecanismos oficiales de GitHub (GFM)

Estos son los mecanismos nativos que GitHub renderiza en archivos `.md` y que se usan para enriquecer los documentos. Conocerlos evita inventar sintaxis que no renderiza o introducir HTML innecesario.

### Badges (shields.io)

- **Estático:** `https://img.shields.io/badge/<label>-<mensaje>-<color>` con `_` = espacio, `--` = guion. Colores: nombrados o hex.
- **Dinámico desde la API de GitHub:** CI (`github/actions/workflow/status/<owner>/<repo>/<workflow>.yml`), estrellas (`github/stars/...`), último commit (`github/last-commit/...`), releases, issues.
- Estilos: `flat`, `flat-square`, `plastic`, `for-the-badge`, `social`. Logo de simple-icons con `?logo=<slug>`.
- Reglas: solo badges que muestren datos reales y útiles (CI que existe, repositorio real). Sin badge de licencia si la licencia no está decidida en el repo. Los parámetros con caracteres especiales (acentos, espacios) van URL-encoded.

### Alerts (admoniciones)

- Sintaxis: `> [!NOTE]`, `> [!TIP]`, `> [!IMPORTANT]`, `> [!WARNING]`, `> [!CAUTION]`, seguido del contenido en blockquote.
- Uso: resaltar información crítica o avisos de contexto. Máximo uno o dos por documento; no anidarlos ni encadenarlos.
- Reemplazan blockquotes planos cuando el aviso merece jerarquía visual (p. ej. "esto es una capacidad declarada, no verificada").

### Diagramas Mermaid

- Sintaxis: bloque de código con lenguaje `mermaid`.
- Soporta flowcharts (`flowchart TD`), secuencia, clase, estado, pie, gantt.
- Útil para reemplazar árboles ASCII por diagramas reales (mapas de decisión, arquitectura, flujos).
- Verificar la sintaxis contra la versión de Mermaid que usa GitHub; las etiquetas con paréntesis y llaves no deben romper el diagrama.

### Emojis

- Sintaxis `:codigo:` (p. ej. `:tada:`). Solo si el usuario los pide o el repo ya los usa. Uso sobrio: encabezados de sección, sin saturar el texto.

### Task lists

- Sintaxis: `- [ ] tarea` / `- [x] hecha`. GitHub las renderiza con checkboxes interactivos.
- Ideales para listas de pendientes o pasos de adopción ("curación necesaria", "antes de publicar").

### Otros nativos

- **Tablas GFM:** fila de encabezado + separador `|---|---|` + filas consistentes.
- **Footnotes:** `texto[^1]` y al pie `[^1]: referencia`.
- **TOC automático:** GitHub genera el índice desde los encabezados; no hace falta un TOC manual.
- **Enlaces relativos:** preferir rutas relativas (`./docs/...`) a absolutas; GitHub las resuelve por rama.
- **Colores:** `#RRGGBB` dentro de backticks muestra una muestra de color solo en issues/PRs/discussions, no en `.md`.
- **geoJSON/topoJSON/STL:** mapas y modelos 3D nativos, solo si aportan al documento.

## Workflow: redactar desde cero

### 1. Inspeccionar el repositorio antes de escribir nada

Nunca escribir un README sin mirar primero el código. Revisar, en orden:

- Manifiesto de dependencias (`package.json`, `Cargo.toml`, `pyproject.toml`, `go.mod`, etc.): stack, scripts, binarios, engines, campos de metadata.
- Estructura de directorios: qué apps, paquetes o componentes incluye el proyecto.
- `docs/`, `AGENTS.md`, `README` anterior, config de CI: contexto y afirmaciones ya existentes; el workflow de CI permite un badge veraz.
- Licencia (`LICENSE` o campo de licencia en el manifiesto): solo mencionarla si está verificada; si es ambigua o está pendiente de decisión, omitir o preguntar, nunca inventar.
- `git log`/tags (opcional): nociones de madurez y versión.

### 2. Identificar la audiencia

¿Quién usará este boilerplate? Un desarrollador con una necesidad concreta (una app móvil, un dashboard, un SaaS). El README debe responder: qué es, para qué sirve, para quién, qué stack usa, cómo empezar.

### 3. Escribir con la plantilla

Usar siempre esta estructura, omitiendo las secciones que no apliquen:

```
# <Nombre>

<badges opcionales: CI, estrellas, último commit>

<Una línea: qué es (starter/boilerplate de X)>

<Descripción de 2-4 líneas: problema que resuelve y para quién>

## Stack / Tecnología
<tabla o lista breve, con nombres oficiales y versiones si son verificables>

## Requisitos previos
<herramientas e instalaciones necesarias>

## Instalación y uso rápido
<comandos reales, verificados contra los scripts del repo, en el orden real>

## Scripts principales
<qué hace cada comando útil del manifiesto>

## Estructura del proyecto
<árbol resumido con un comentario corto por directorio>

## Configuración
<variables de entorno o ajustes, solo si existen; sin inventar valores>

## Documentación
<enlaces a docs oficiales o a docs/ del repositorio>

## Licencia
<solo si está verificada; con enlace al archivo LICENSE>
```

No agregar secciones cosméticas (Contribución, FAQ, Screenshots, Changelog) salvo que tengan contenido real y verificable o que el usuario las pida explícitamente.

### 4. Verificar antes de terminar

- Cada comando listado existe en el manifiesto/scripts del repo.
- Cada archivo referenciado existe (rutas exactas, respetando mayúsculas).
- No hay afirmaciones sin soporte en el código.
- Los badges apuntan al owner/repo real y a workflows que existen.
- El Markdown renderiza en GitHub.

## Workflow: modificar documentos existentes

Este flujo aplica a cualquier modificación (mejorar, enriquecer, corregir o reestructurar un README u otro documento Markdown del repo):

1. **Leer antes de tocar.** Leer el documento completo, los documentos vinculados que pueda afectar (CHANGELOG, mapas de decisión, tablas de cobertura, índices) y las reglas del repo en AGENTS.md.
2. **Diagnosticar.** Problemas concretos: secciones faltantes, comandos desactualizados o inexistentes, afirmaciones no verificadas, enlaces rotos, formato que no renderiza (p. ej. sangrado de 4 espacios que rompe listas y tablas, mezcla de estilos de encabezado), ausencia de mecanismos visuales que el contexto amerita.
3. **Proponer y consultar antes de eliminar o reestructurar.** Si el cambio elimina contenido, cambia el orden de columnas/secciones o modifica decisiones documentadas, presentar el análisis y esperar la confirmación explícita del usuario antes de aplicar. Nunca aplicar decisiones de este tipo en el mismo turno en que se propusieron.
4. **Conservar lo válido.** Mantener contenido correcto, tono y nombres técnicos. Cambios mínimos que acerquen al estándar.
5. **Actualizar los documentos vinculados.** Si el repo usa changelog, versionado, mapas de cobertura o índices, actualizarlos en la misma operación: el README no es un documento aislado.
6. **Ejecutar las validaciones del repo.** Si existen scripts de validación (p. ej. `validate_catalog.py`) o generadores, correrlos y dejar todo en verde antes de terminar.
7. **Explicar qué se corrigió y por qué** al entregar el resultado.

## Reglas de formato para que renderice bien en GitHub

- Jerarquía de encabezados correcta (`#` título, `##` secciones, sin saltar niveles).
- Tablas bien formadas: fila de encabezado, línea separadora `|---|---|`, filas consistentes en número de columnas.
- Bloques de código con lenguaje declarado (```bash, ```ts, ```mermaid) y precedidos de línea en blanco.
- Línea en blanco entre bloques de Markdown que GitHub junta (listas, tablas, bloques de código).
- Enlaces relativos con rutas reales del repo (`./docs/...`).
- Preferir Markdown puro sobre HTML cuando Markdown alcanza.
- Código de ejemplo verificable: debe ser el de los scripts reales, no inventado.
- Alertas `[!NOTE]`-`[!CAUTION]` limitadas y nunca anidadas.
- Diagramas Mermaid en bloques propios, sin texto suelto que los rompa.

## Recordatorio final

El README es la primera impresión del boilerplate: que un visitante sepa en menos de un minuto qué es, si le sirve y cómo arrancar, sin que el repositorio tenga que defenderse luego de promesas que no cumplió. Y cada modificación posterior debe dejar el documento mejor que como estaba, sin perder información ni decisiones.
