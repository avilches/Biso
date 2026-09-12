# El tutorial de `biso`

Esta carpeta es la fuente de `docs/TUTORIAL.md`, que es **producto generado y no se edita a mano**.
Si has llegado aquí para cambiar algo del tutorial, lo que se toca es un fichero de esta carpeta y
luego se regenera la página.

Su diseño, con el porqué de cada decisión, está en
[`docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md`](../docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md).

## La idea

El tutorial va **por situaciones y no por comandos**. Un capítulo abre con un apuro que el lector
reconoce ("a mitad de la tarea descubres que falta un criterio") y los comandos aparecen porque la
situación los pide.

No es una decisión estética. Un tutorial ordenado por comandos sería una segunda copia de
["Los comandos"](../docs/spec/cmd/index.md), y en cuanto existen dos copias empiezan a divergir. Ordenado por situaciones
aporta lo único que la especificación no tiene, que es el porqué y el orden en que se conocen los
conceptos, y no compite con ella como fuente de verdad.

Los trece escenarios siguen la vida de `TASK-19` desde que alguien se la asigna hasta que se cierra.
Que el lector vea una tarea entera de principio a fin es lo que convierte trece capítulos en una
historia.

## Qué hay en cada sitio

| Ruta | Qué es |
|---|---|
| `escenarios/NN-nombre.yaml` | Un capítulo. El orden lo fija el prefijo numérico |
| `conceptos.md` | La sección de conceptos desde cero que abre la página |
| `tablero.yaml` | El tablero de ejemplo: su configuración y sus nueve tareas |
| `lagunas/*.md` | Lo que la especificación no decide y hubo que suponer |
| `generate.py` | Escribe `docs/TUTORIAL.md` |
| `urgency.py` | Calcula la urgencia de cada tarea |
| `continuity.py` | Comprueba que los escenarios encadenan |

## El contrato de un fixture

Cada paso declara el comando, su salida, su código de salida y **de dónde sale esa salida**:

```yaml
  - narracion: |
      Prosa que prepara el comando.
    comando: biso note TASK-11 "Reescrito el parser de fechas"
    salida: |
      warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z
      TASK-11  In Progress  ac 1/2  dod 0/2  urgency 41.0
    codigo_salida: 0
    origen: literal spec/salida-y-terminal.md#notas-y-avisos
    comentario: |
      Opcional. Lo que hay que mirar en esa salida y por qué.
```

Las reglas que no se negocian:

- **`origen` solo admite dos formas**, `literal spec/<ruta>` si la salida está copiada carácter a
  carácter de ese fichero de `docs/spec/`, o `derivada spec/<ruta>[, spec/<ruta>...]` si se construyó
  aplicando las reglas de esos ficheros. `<ruta>` lleva ancla cuando la regla está dentro de un
  apartado y no cuando es el fichero entero. El generador **marca las derivadas en la página
  enlazando cada ruta**, y esa marca es el aparato de validación: lo literal ya está validado por
  estar en la especificación, lo derivado es lo que hay que revisar. De 59 pasos, 7 son literales.
- **`codigo_salida` es obligatorio en todos los pasos**, también en los que valen cero. Un tutorial
  que solo declara el código cuando falla enseña que el código solo importa al fallar, y en `biso` es
  al revés.
- **`salida` es stdout y stderr juntos**, en el orden en que los ve una persona en su terminal,
  porque es lo que el lector va a comparar. Cuando la distinción importa se dice en `comentario`.
- **Nada se inventa.** Si la especificación no decide algo que un escenario necesita, no se rellena
  con lo que parezca razonable: se anota en `lagunas/` con la pregunta concreta y lo que se supuso.
  Esa carpeta alimenta `docs/PENDIENTES.md`, y es uno de los productos valiosos de escribir esto.
- **Los escenarios comparten un tablero y ocurren en orden.** Cada fichero declara en
  `tablero_entra` y `tablero_sale` qué estado recibe y qué estado entrega, y eso es lo que hace
  visible una rotura de continuidad.

Los ficheros de esta carpeta son documentación con forma de datos, así que **sus claves y su prosa
van en español**. Los tres scripts son código, así que van enteros en inglés, con la única excepción
de las cadenas que el generador emite dentro de la página, agrupadas en un bloque marcado al
principio de `generate.py`.

## Los tres comandos

Los tres llevan el mismo prefijo de `uv` que MkDocs, que no instala nada en el sistema:

```
uv run --with-requirements docs-requirements.txt --no-project python tutorial/generate.py
uv run --with-requirements docs-requirements.txt --no-project python tutorial/urgency.py
uv run --with-requirements docs-requirements.txt --no-project python tutorial/continuity.py
```

`generate.py` escribe la página, y **falla sin escribir nada** si a un paso le falta el código de
salida, si `origen` no tiene una de las dos formas admitidas o si un fichero no es YAML válido. Un
generador que se traga un fixture incompleto destruye la única garantía de este diseño.

Los otros dos existen porque había números en los fixtures que nadie podía verificar leyendo un solo
fichero:

- `urgency.py` calcula la urgencia de cada tarea según ["La urgencia"](../docs/spec/modelo-de-datos.md#la-urgencia), con el
  desglose de cada término, y las ordena por la regla de `biso ls`. Los escenarios ordenan listas por
  urgencia, y ese orden hay que poder reproducirlo en vez de creerse un número escrito por alguien.
  Reproduce el `urgency 19.0` que la especificación imprime para `TASK-11`, que es la comprobación de
  que el tablero de ejemplo está bien modelado.
- `continuity.py` comprueba que los contadores del tablero encadenan entre escenarios consecutivos.
  Sale 1 si no cuadran. Ya cazó un desfase de uno que cuatro escenarios arrastraban por no contar la
  `TASK-62` que crea el escenario 2.

**Después de tocar cualquier fixture, hay que pasar los tres**: regenerar, y luego los dos
comprobadores. Y construir el sitio, que es la comprobación de que la página entra bien:

```
uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict
```

## Lo que viene

El día que exista el binario, un script recorrerá estos mismos ficheros, sembrará un tablero real con
`tablero.yaml`, lanzará cada `comando` y comparará su salida y su código con lo declarado. Los
ejemplos dejarán de ser simulados y pasarán a ser una batería de pruebas de salida literal.

Por eso `comando` es una línea ejecutable y no una ilustración, y por eso `codigo_salida` es
obligatorio. Cualquier atajo en el formato que hoy parezca inofensivo se paga entonces.
