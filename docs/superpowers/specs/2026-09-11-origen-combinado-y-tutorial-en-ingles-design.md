# El campo `origen` con varias citas, y el tutorial en inglés

**Sustituye a la tarea 10 de [`docs/superpowers/plans/2026-09-10-reparto-de-la-spec.md`](../plans/2026-09-10-reparto-de-la-spec.md).**
Esa tarea solo contemplaba traducir `origen` de citar por número a citar por fichero y título, con una
sola sección por cita. Este documento decide dos cosas más que aparecieron al mirar el estado real de
`worktree-tutorial`: cómo se cita más de una sección a la vez, y que el tutorial entero (no solo el campo
`origen`) nace en inglés.

## El problema de partida

El campo `origen` de un paso de un escenario dice de dónde sale su salida: `literal SPEC <número>` cuando
el texto está copiado carácter a carácter de la especificación, `derivada SPEC <número>` cuando se ha
construido aplicando las reglas de esa sección porque la especificación no trae ese caso exacto escrito.
Con la especificación repartida en `docs/spec/`, `origen` tiene que citar por fichero y título en vez de
por número, como ya hace el resto del repositorio desde la tarea 5.

El paso a paso original de la tarea 10 solo cubría una sección por cita
(`literal spec/salida-y-terminal.md "Notas y avisos"`). Al escribirse ese plan solo existía el escenario
`01-llegas-a-un-proyecto.yaml`, con dos `origen` de una sola sección cada uno. `worktree-tutorial` tiene
hoy doce escenarios más, y muchos de sus pasos citan varias secciones a la vez separadas por coma
(`derivada SPEC 5.2, 5.3, 10.5, 10.7.6`). Ni esa forma nueva de una sola cita ni el generador de fixtures
dicen cómo escribir una cita combinada.

## Alcance de este documento

Decide la forma del campo de citas del tutorial y confirma que el tutorial se escribe en inglés. **No**
decide, y queda fuera a propósito:

- **El cambio de idioma de todo el repositorio.** Es un proyecto aparte, con su propia decisión razonada
  en `DECISIONES.md` cuando se plantee. Este documento no toca la convención de `CLAUDE.md` de que la
  documentación del repositorio va en español; el tutorial es una excepción declarada aquí, no un cambio
  de regla general.
- **La traducción literal de cada cadena de texto del generador** (los títulos de las secciones de la
  página, el aviso de "documento generado", etc.). Este documento fija qué cadenas cambian de idioma y
  con qué criterio; el texto exacto de cada una es trabajo del plan de implementación, no una decisión de
  diseño.

## El esquema de los fixtures, en inglés

El tutorial nace en inglés porque es contenido nuevo, y con él sus doce escenarios ya escritos, que hoy
están en español de arriba a abajo: título, situación, qué enseña, la narración y el comentario de cada
paso, y las dos frases de balance del tablero. Los nombres de los campos también cambian, no solo el
contenido:

| Campo hoy (español) | Campo nuevo (inglés) |
|---|---|
| `id` | `id` |
| `titulo` | `title` |
| `situacion` | `situation` |
| `ensena` | `teaches` |
| `pasos` | `steps` |
| `tablero_entra` | `board_in` |
| `tablero_sale` | `board_out` |
| `narracion` | `narration` |
| `comando` | `command` |
| `salida` | `output` |
| `codigo_salida` | `exit_code` |
| `origen` | `source_kind` + `source` (ver más abajo) |
| `comentario` | `remark` |

Los nombres nuevos coinciden con los identificadores que `tutorial/generate.py` ya usa internamente
(`KEY_TITLE`, `KEY_NARRATION`, etc.), así que el cambio de esquema no inventa vocabulario: adopta el que
el propio generador ya eligió para razonar sobre el formato.

## El campo `source_kind` y `source`

`origen` se separa en dos campos:

- **`source_kind`**: `literal` o `derived`, una vez por paso. Sustituye a las palabras `literal`/`derivada`
  que hoy encabezan la cadena de `origen`.
- **`source`**: una lista de citas, siempre lista aunque solo traiga una. Cada cita tiene la forma
  `ruta/spec/fichero.md "Título de la sección"`, con la ruta relativa a `docs/` y el título tal cual
  aparece en el fichero citado, entre comillas. La cita no lleva ancla escrita a mano: el generador la
  calcula del título con la misma función de slugify que ya usa `tools/generar_mapa.py` para
  `docs/spec/`, así que una cita nunca puede traer un ancla que no cuadre con su propio título.

Ejemplo de un paso con una sola cita:

```yaml
steps:
  - command: biso where
    output: |
      ...
    exit_code: 0
    source_kind: literal
    source:
      - spec/cmd/where.md "`biso where`"
```

Ejemplo de un paso con varias, el caso real de `08-te-atascas.yaml` en `worktree-tutorial`
(`derivada SPEC 5.2, 5.3, 10.5, 10.7.6`):

```yaml
steps:
  - command: biso answer TASK-19 "..."
    output: |
      ...
    exit_code: 0
    source_kind: derived
    source:
      - spec/modelo-de-datos.md "Los comentarios"
      - spec/modelo-de-datos.md "Las fechas"
      - spec/cmd/get.md "`biso get`"
      - spec/cmd/verbos-del-ciclo.md "`biso answer`"
```

**El título de cada cita va en el idioma del fichero citado, no traducido.** `docs/spec/` se queda en
español, así que el título de la cita tiene que ser el título real de esa sección, carácter a carácter,
para que el ancla calculada por slugify exista de verdad. El efecto es que una página en inglés va a
enlazar con texto en español (`[Los comentarios](...)`, no `[Comments](...)`): es la consecuencia directa
de que el tutorial pase a inglés mientras la especificación se queda en español, y no algo que este
documento pueda evitar sin traducir también `docs/spec/`, que quedó fuera de alcance.

**Regla de la lista:** cuando `source_kind` es `literal`, `source` tiene que traer exactamente una cita,
porque un texto copiado carácter a carácter solo puede salir de un sitio. En los datos de hoy no existe
ningún `literal` con más de una sección, así que la regla no rompe ningún caso real; el generador la
comprueba y falla si aparece uno. Cuando `source_kind` es `derived`, `source` admite una o varias.

## Renderizado en la página generada

`literal` sigue sin marcarse en la página, igual que hoy: ya está validado por estar en la especificación
tal cual.

Cada cita de un paso `derived` se convierte en un enlace real de Markdown, con el título de la cita como
texto del enlace, dentro de la nota que ya existe para marcar las salidas derivadas:

```
*(derived output, see [Los comentarios](spec/modelo-de-datos.md#los-comentarios),
[Las fechas](spec/modelo-de-datos.md#las-fechas), [`biso get`](spec/cmd/get.md),
[`biso answer`](spec/cmd/verbos-del-ciclo.md#biso-answer); not literal spec text)*
```

Como `docs/TUTORIAL.md` es un único fichero dentro de `docs/`, la ruta de cada enlace es la misma que
usa la cita en `source`, sin ningún prefijo `../` que calcular.

## Validación del generador

El generador comprueba, por cada paso:

- que `source_kind` sea `literal` o `derived`, sin ninguna otra palabra;
- que `source` no esté vacío;
- que un `source_kind: literal` traiga exactamente un elemento en `source`;
- que cada elemento de `source` tenga la forma `ruta "Título"`.

**No comprueba que el fichero o el título citados existan de verdad.** Esa comprobación ya la hace gratis
`uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict` en cuanto la cita se
convierte en un enlace real dentro de `docs/`, igual que ya hace con las 374 referencias de la
especificación de la tarea 5. Escribir una segunda comprobación que valide lo mismo sería duplicar la red
de seguridad que ya existe.

## Dónde se ejecuta esto

El contenido que cambia (los trece escenarios, `tutorial/generate.py` y el documento de diseño del
tutorial) vive en `worktree-tutorial`, no en este worktree. Una sesión aislada en
`.claude/worktrees/reparto-de-la-spec` no puede ejecutar comandos de git contra `worktree-tutorial`: hace
falta una sesión que trabaje directamente ahí, con este documento como referencia, para escribir el plan
de implementación y ejecutarlo.
