# El tutorial de `biso`

Esta carpeta es la fuente de las páginas del tutorial (`docs/tutorial/*.md`), que son **producto
generado y no se editan a mano**. La página de Conceptos (`docs/concepts.md`, publicada bajo el tab
Inicio del sitio) ya no se genera desde aquí: es un Markdown normal y se edita en su sitio. Si has llegado aquí para cambiar algo del tutorial, lo que se toca es un
fichero de esta carpeta y luego se regeneran las páginas.

Su diseño, con el porqué de cada decisión, está en
[`docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md`](../../superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md).
La forma de las citas a varias secciones a la vez y que el contenido del tutorial nace en inglés
(a diferencia del resto de la documentación del repositorio) se decidió después, en
[`docs/superpowers/specs/2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md`](../../superpowers/specs/2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md).
Que el tutorial se generara como una página por escenario, y que Conceptos se sacara a su propia
página bajo Inicio manteniendo el inglés, se decidió en
[`docs/superpowers/specs/2026-09-12-navegacion-del-sitio-por-paginas-design.md`](../../superpowers/specs/2026-09-12-navegacion-del-sitio-por-paginas-design.md).

## La idea

El tutorial va **por situaciones y no por comandos**. Un capítulo abre con un apuro que el lector
reconoce ("a mitad de la tarea descubres que falta un criterio") y los comandos aparecen porque la
situación los pide.

No es una decisión estética. Un tutorial ordenado por comandos sería una segunda copia de
["Los comandos"](../../spec/cmd/index.md), y en cuanto existen dos copias empiezan a divergir. Ordenado por situaciones
aporta lo único que la especificación no tiene, que es el porqué y el orden en que se conocen los
conceptos, y no compite con ella como fuente de verdad.

Los trece escenarios siguen la vida de `TASK-19` desde que alguien se la asigna hasta que se cierra.
Que el lector vea una tarea entera de principio a fin es lo que convierte trece capítulos en una
historia.

## Qué hay en cada sitio

| Ruta | Qué es |
|---|---|
| `escenarios/NN-nombre.yaml` | Un capítulo. El orden lo fija el prefijo numérico |
| `tablero.yaml` | El tablero de ejemplo: su configuración y sus nueve tareas |
| `lagunas/*.md` | Lo que la especificación no decide y hubo que suponer |
| `generate.py` | Escribe `docs/tutorial/*.md` |
| `urgency.py` | Calcula la urgencia de cada tarea |
| `continuity.py` | Comprueba que los escenarios encadenan |

## El contrato de un fixture

Cada paso declara el comando, su salida, su código de salida y **de dónde sale esa salida**:

```yaml
  - narration: |
      Prose that sets up the command.
    command: biso note TASK-11 "Rewrote the date parser"
    output: |
      warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z
      TASK-11  In Progress  ac 1/2  dod 0/2  urgency 41.0
    exit_code: 0
    source_kind: literal
    source:
      - spec/salida-y-terminal.md#notas-y-avisos "Notas y avisos"
    remark: |
      Optional. What to look at in that output, and why.
```

Las reglas que no se negocian:

- **`source_kind` es `literal` o `derived`, y `source` es siempre una lista**, de una sola cita si
  `source_kind` es `literal` (un texto copiado carácter a carácter solo puede salir de un sitio), de
  una o varias si es `derived`. Cada cita de `source` tiene la forma `ruta/spec/fichero.md "Título"`,
  con la ruta relativa a `docs/` y **el título tal cual aparece en `docs/spec/`, sin traducir**: el
  generador lo usa literal como texto del enlace, y `docs/spec/` se queda en español aunque el
  tutorial nazca en inglés. El generador **enlaza cada cita de un paso `derived`** en la página, y esa
  marca es el aparato de validación: lo literal ya está validado por estar en la especificación, lo
  derivado es lo que hay que revisar. De 59 pasos, 7 son literales.
- **`exit_code` es obligatorio en todos los pasos**, también en los que valen cero. Un tutorial
  que solo declara el código cuando falla enseña que el código solo importa al fallar, y en `biso` es
  al revés.
- **`output` es stdout y stderr juntos**, en el orden en que los ve una persona en su terminal,
  porque es lo que el lector va a comparar. Cuando la distinción importa se dice en `remark`.
- **Nada se inventa.** Si la especificación no decide algo que un escenario necesita, no se rellena
  con lo que parezca razonable: se anota en `lagunas/` con la pregunta concreta y lo que se supuso.
  Esa carpeta alimenta las tareas de seguimiento del tablero, y es uno de los productos valiosos de
  escribir esto.
- **Los escenarios comparten un tablero y ocurren en orden.** Cada fichero declara en
  `board_in` y `board_out` qué estado recibe y qué estado entrega, y eso es lo que hace
  visible una rotura de continuidad.

**El contenido del tutorial nace en inglés, al revés que el resto de la documentación del
repositorio.** `title`, `situation`, `teaches`, `narration`, `remark`, `board_in` y `board_out` van
en inglés; `command` y `output` ya estaban en inglés porque son texto literal de la interfaz de
`biso`; y el título de cada cita de `source` se queda en español porque cita un fichero de
`docs/spec/` que sigue en español. Es una excepción declarada para esta carpeta, para
`docs/tutorial/*.md` y para `docs/concepts.md`, no un cambio de la regla general de `CLAUDE.md` de
que la documentación va en español: este fichero, `lagunas/*.md` y los documentos de diseño siguen
en español. Los nombres de los ficheros y de la carpeta (`escenarios/NN-nombre.yaml`) tampoco cambian de idioma, solo su contenido. Los tres scripts son código, así que
van enteros en inglés, con la única excepción de las cadenas que el generador emite dentro de las
páginas, que ahora también están en inglés y ya no son una excepción al idioma del código.

## Los tres comandos

Los tres llevan el mismo prefijo de `uv` que MkDocs, que no instala nada en el sistema:

```
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/generate.py
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/urgency.py
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project python docs/docs-tooling/tutorial/continuity.py
```

`generate.py` escribe las páginas, y **falla sin escribir nada** si a un paso le falta el código de
salida, si `origen` no tiene una de las dos formas admitidas o si un fichero no es YAML válido. Un
generador que se traga un fixture incompleto destruye la única garantía de este diseño.

Los otros dos existen porque había números en los fixtures que nadie podía verificar leyendo un solo
fichero:

- `urgency.py` calcula la urgencia de cada tarea según ["La urgencia"](../../spec/modelo-de-datos/urgencia.md#la-urgencia), con el
  desglose de cada término, y las ordena por la regla de `biso ls`. Los escenarios ordenan listas por
  urgencia, y ese orden hay que poder reproducirlo en vez de creerse un número escrito por alguien.
  Reproduce el `urgency 19.0` que la especificación imprime para `TASK-11`, que es la comprobación de
  que el tablero de ejemplo está bien modelado.
- `continuity.py` comprueba que los contadores del tablero encadenan entre escenarios consecutivos.
  Sale 1 si no cuadran. Ya cazó un desfase de uno que cuatro escenarios arrastraban por no contar la
  `TASK-62` que crea el escenario 2.

**Después de tocar cualquier fixture, hay que pasar los tres**: regenerar, y luego los dos
comprobadores. Y construir el sitio, que es la comprobación de que las páginas entran bien:

```
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project mkdocs build --strict -f docs/docs-tooling/mkdocs/mkdocs.yml
```

## Lo que viene

El día que exista el binario, un script recorrerá estos mismos ficheros, sembrará un tablero real con
`tablero.yaml`, lanzará cada `comando` y comparará su salida y su código con lo declarado. Los
ejemplos dejarán de ser simulados y pasarán a ser una batería de pruebas de salida literal.

Por eso `comando` es una línea ejecutable y no una ilustración, y por eso `codigo_salida` es
obligatorio. Cualquier atajo en el formato que hoy parezca inofensivo se paga entonces.
