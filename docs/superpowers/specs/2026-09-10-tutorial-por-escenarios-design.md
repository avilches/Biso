# Tutorial de `biso` por escenarios, generado de fixtures

Diseñado el 2026-09-10. Este documento es el encargo del que se implementa el tutorial.

> **La prosa de este documento cita la especificación por el número de sus secciones**, como se
> escribió en su día. Ese mismo día `docs/SPEC.md` se repartió en los documentos de `docs/spec/`, que
> se citan por el título de sus secciones. Estas referencias narrativas se conservan sin tocar porque
> este documento es el encargo de una tarea ya cerrada; para traducir una, mira
> `tools/mapa-de-secciones.txt`. La única excepción es el campo `origen` de los fixtures, que sí se
> tradujo porque es un contrato activo que valida el generador, y que ahora cita una ruta de
> `docs/spec/` en vez de un número.

## El problema

`docs/SPEC.md` es completa y es exacta, pero es una referencia: dice qué hace cada comando, no en qué
orden se conocen los conceptos ni por qué uno querría usarlos. Alguien que no ha visto nunca `biso`
(ni Backlog.md, ni ningún gestor de tareas para agentes) no puede empezar por ella.

Hace falta un documento que enseñe desde cero. Y como no hay implementación todavía, hace falta además
que sus ejemplos no sean prosa inventada, porque un ejemplo inventado que contradice la spec es peor
que no tener ejemplo: parece autoridad y no lo es.

## La decisión de fondo: no ir comando por comando

El tutorial se organiza por **situaciones**, no por comandos. Un capítulo empieza con un apuro que el
lector reconoce ("a mitad de la tarea descubres que falta un criterio") y los comandos aparecen porque
la situación los pide.

La razón no es estética. Un tutorial ordenado por comandos es una segunda copia de la sección 10 de la
spec, y en cuanto existen dos copias empiezan a divergir. Ordenado por situaciones, el tutorial aporta
lo único que la spec no tiene, que es el porqué y el orden, y no compite con ella como fuente de verdad.

**Queda fuera a propósito** la referencia exhaustiva de parámetros. Está pedida como documento aparte,
y reutilizará la misma maquinaria de fixtures que se define aquí.

## Las tres piezas

1. **Los fixtures**, en `tutorial/escenarios/NN-nombre.yaml`: un fichero por escenario, donde cada paso
   declara el comando, su salida esperada, su código de salida y la prosa que lo acompaña.
2. **El generador**, `tutorial/generate.py`: convierte los fixtures en `docs/TUTORIAL.md`.
3. **La página**, `docs/TUTORIAL.md`: producto generado, entra en el sitio de MkDocs que ya existe y
   **no se edita a mano**. Lleva una cabecera que lo dice.

El tutorial vive dentro del sitio de documentación en vez de ser una página suelta para heredar gratis
la navegación, el buscador y el modo oscuro, y para que no haya un segundo sitio que mantener.

## El tablero de ejemplo

**El tutorial usa el mismo tablero de ejemplo que la spec**, el de la sección 9.7. Eso no es comodidad:
es lo que permite copiar salidas literales del documento en vez de inventarlas, y hace que cualquier
discrepancia entre el tutorial y la spec sea detectable.

Tablero `Kex`, con `task_prefix` fijado a `TASK`, estados `To Do` (54), `In Progress` (4) y `Done` (190),
tipos `idea, memory, task, bug, docs`, prioridades `high, medium, low`, e identidad `@claude`.

Inventario de partida, tal y como lo imprime `biso prime` en 9.7:

| Id | Estado | Tipo | Prior. | Título | ac | Asignada | Notas |
|---|---|---|---|---|---|---|---|
| `TASK-11` | In Progress | bug | high | Normalize CRLF in the diff | 1/2 | `@claude` | |
| `TASK-52` | In Progress | task | low | Document the release checklist | 0/1 | | arrendamiento vencido el 2026-09-05, era de `@bob` |
| `TASK-40` | In Progress | task | medium | Split the config loader | 0/2 | `@claude` | |
| `TASK-60` | In Progress | task | high | Confirm the retry budget for the upload endpoint | 0/2 | `@claude` | tiene pregunta abierta |
| `TASK-61` | To Do | docs | medium | Rewrite the install section | 0/1 | `@claude` | |
| `TASK-33` | To Do | task | medium | Add a retry counter to the upload log | 1/3 | `@claude` | |
| `TASK-7` | To Do | bug | high | Crash on an empty repository | 0/4 | | vence el 2026-09-08 |
| `TASK-19` | To Do | task | high | Retry the upload on 5xx | 0/2 | | |
| `TASK-44` | To Do | bug | low | Wrong column width on narrow ttys | 0/1 | | |

**`TASK-19` es la espina dorsal del tutorial.** Se asigna en el escenario 4 y se sigue hasta que se
cierra en el 10, pasando por el arrendamiento, los criterios, la pregunta y los comentarios. Que el
lector vea una tarea entera de principio a fin es lo que convierte trece capítulos en una historia.

## Los escenarios

Precedidos de una sección de conceptos que no da nada por sabido: qué es un tablero, qué son los
estados, en qué se diferencian los criterios de aceptación de la definición de hecho, qué es la
urgencia y qué significa tener una identidad.

| # | Situación | Comandos | Tareas que toca |
|---|---|---|---|
| 1 | Llegas a un proyecto que no conoces | `prime`, `where`, mención de `init` | ninguna, solo lee |
| 2 | Se te ocurre algo y no quieres que se pierda | `new` | crea `TASK-62` |
| 3 | ¿Y ahora qué hago? | `ls`, `get` | lee `TASK-33` |
| 4 | Repartes el trabajo | `config set me`, `set --assignee`, `ls --mine`, `prime` | asigna `TASK-19` |
| 5 | Te pones con ella | `start` | `TASK-19` pasa a In Progress |
| 6 | Dos a la vez sobre el mismo tablero | escrituras con arrendamiento ajeno | `TASK-11`, `TASK-52` |
| 7 | A mitad, falta un criterio | `set --dod`, `--rm-dod`, `--set-dod` | `TASK-19` |
| 8 | Te atascas y alguien tiene que decidir | `ask`, `answer` | `TASK-19`, mira `TASK-60` |
| 9 | La persona y el agente hablan | `comment`, `note` | `TASK-19` |
| 10 | Cierras | `set --check`, `finish` | `TASK-19` pasa a Done |
| 11 | Te equivocas | filtro imposible, referencia ambigua, inexistente | ninguna, todo falla |
| 12 | Aparcas algo a medias | `biso delete` (que no existe), `archive`, `ls --archived` | `TASK-52` |
| 13 | Tocas veinte de golpe | `set` con varias referencias, `--dry-run` | varias |

Lo que cada escenario tiene que dejar enseñado, y que no es adivinable:

- **1**: que el arranque se lee entero una vez y ya no hace falta nada más, y que `where` existe para
  cuando dudas de qué tablero estás tocando.
- **2**: que el nombre desnudo de una bandera añade, y que `new` imprime el identificador y nada más.
- **3**: que la lista corta a 30 y esconde las terminadas, y que lo dice por stderr. Que la urgencia es
  derivada y no se escribe.
- **4**: que `--mine` no funciona sin identidad declarada, y que asignar es una decisión de una persona
  sobre lo que otro debe hacer, no un reparto automático.
- **5**: que empezar es un comando propio, y que al empezar nace un arrendamiento sin pedirlo.
- **6**: que el arrendamiento **no es un estado**, que cualquier escritura de quien lo tiene lo renueva
  sola sin comando de latido, que una escritura ajena avisa pero no se bloquea, y que vencido deja de
  privilegiar a nadie. Es el escenario con más regla por línea de todo el tutorial.
- **7**: las cuatro formas de un campo de lista (añadir, sustituir la lista, quitar uno, vaciar) y que
  las claves `#N` son estables y no se desplazan al quitar un elemento.
- **8**: que preguntar aparca la tarea y que eso se ve desde fuera, y que preguntar es preferible a
  adivinar.
- **9**: la diferencia entre una nota de implementación y un comentario con autor.
- **10**: los formatos que aceptan `--check` (uno, un rango, una lista, `all`, o el texto del criterio).
- **11**: el principio 1 de la spec, que es el más importante de todos: un valor que el tablero no
  conoce es un error y nunca una lista vacía, de modo que una lista vacía es un hecho.
- **12**: que archivar no es terminar, y qué le pasa al arrendamiento de una tarea archivada.
- **13**: que todo lo que se hace una vez se puede hacer cien, y que se valida antes de escribir nada.

## El formato de los fixtures

```yaml
id: 06-el-arrendamiento
titulo: Dos agentes a la vez sobre el mismo tablero
situacion: |
  Prosa que plantea el apuro, antes de que aparezca ningún comando.
ensena:
  - Un arrendamiento no es un estado, es hasta cuándo vale la reserva.
  - Cualquier escritura de quien lo tiene lo renueva. No hay comando de latido.
pasos:
  - narracion: |
      Prosa que prepara el comando.
    comando: biso note TASK-11 "Reescrito el parser de fechas"
    salida: |
      warning: TASK-11's lease is held by @sara until 2026-09-08T14:00:00Z
      TASK-11  In Progress  ac 1/2  dod 0/2  urgencia 41
    codigo_salida: 0
    origen: literal spec/salida-y-terminal.md#notas-y-avisos
    comentario: |
      Opcional. Lo que hay que mirar en esa salida y por qué.
```

Reglas del formato:

- **`salida` es lo que sale por stdout y por stderr juntos**, en el orden en que los ve una persona en
  su terminal, porque es lo que el lector va a comparar. Cuando la distinción importe (que importa en el
  escenario 3), se dice en `comentario`.
- **`codigo_salida` es obligatorio en todos los pasos**, también en los que valen cero. Un tutorial que solo
  declara el código cuando falla enseña que el código solo importa al fallar, y en `biso` es al revés.
- **`origen` es obligatorio** y solo admite dos formas:
  - `literal spec/<ruta>`: la salida está copiada carácter a carácter de ese fichero de
    `docs/spec/`.
  - `derivada spec/<ruta>[, spec/<ruta>...]`: la salida se ha construido aplicando las reglas de
    esos ficheros, porque la especificación no trae ese caso exacto escrito. Con varios, separados
    por comas.

  `<ruta>` es la ruta relativa a `docs/spec/`, con ancla cuando la regla está dentro de un apartado
  (`cmd/set.md#selectores-de-criterios`) y sin ella cuando es el fichero entero. El generador
  **marca visualmente las derivadas** en la página, enlazando cada ruta al fichero real: esa marca
  es el aparato de validación: al revisar, lo literal ya está validado por estar en la especificación
  y lo derivado es lo que hay que mirar. Un enlace que apunte a un fichero o ancla que no existe lo
  dice el build de MkDocs con `--strict`.
- **Los pasos de un escenario ocurren en orden y el estado del tablero se arrastra** entre escenarios.
  El identificador `NN` del fichero fija ese orden.

## Cómo se mantiene la continuidad con varios agentes escribiendo a la vez

Este es el riesgo real del reparto: tres agentes escribiendo escenarios en paralelo producen tres
tableros distintos. Se ataja por dos vías.

**Primera, el inventario de arriba es normativo.** Ningún escenario inventa una tarea que no esté en esa
tabla, salvo la `TASK-62` que crea el escenario 2. Los contadores del tablero (`To Do 54`, etc.) solo
cambian cuando un escenario mueve una tarea de estado, y el escenario que la mueve es el responsable de
reflejarlo.

**Segunda, cada fichero declara qué recibe y qué entrega.** Dos campos al final:

```yaml
tablero_entra: TASK-19 en To Do, sin asignar, ac 0/2. Contadores 54/4/190.
tablero_sale: TASK-19 en In Progress, asignada a @claude, con arrendamiento de @claude. Contadores 53/5/190.
```

No los usa el generador: los usa el agente revisor, y los usa una persona al leer. Es el mecanismo más
barato que hace visible una rotura de continuidad sin ejecutar nada.

## El generador

`tutorial/generate.py`, sin dependencias fuera de la biblioteca estándar más el `PyYAML` que ya arrastra
MkDocs. Lee los ficheros de `tutorial/escenarios/` en orden, y escribe `docs/TUTORIAL.md`.

Por cada escenario emite: el título como sección, la situación como prosa, lo que enseña como un aviso
destacado (`admonition`, que ya está activada en `mkdocs.yml`), y luego cada paso con su narración, su
comando, su salida en un bloque de código y su comentario.

Dos cosas que el generador tiene que hacer bien:

- **Marcar las salidas derivadas.** Una nota discreta y consistente junto al bloque, no un aviso
  aparatoso: son la mayoría de los pasos y saturaría la página.
- **Fallar en vez de generar algo malo.** Si a un paso le falta `codigo_salida`, si `origen` no tiene una de las
  dos formas admitidas, o si un fichero no es YAML válido, el script termina con código distinto de cero
  y dice qué fichero y qué paso. Un generador que se traga un fixture incompleto y produce una página
  incompleta destruye la única garantía que este diseño ofrece.

Se ejecuta con `uv`, igual que MkDocs, y queda anotado en `CLAUDE.md` junto a los otros dos comandos.

## Lo que se ejecutará el día que exista el binario

No se construye ahora, pero el formato está pensado para ello: un script recorrerá los mismos ficheros,
lanzará cada `comando` contra un tablero real sembrado con el inventario de arriba, y comparará la salida y
el código con lo declarado. Los ejemplos dejarán de ser simulados y pasarán a ser una batería de pruebas
de salida literal.

Por eso `comando` es una línea de comando ejecutable y no una ilustración, y por eso
`codigo_salida` es obligatorio.
Cualquier atajo en el formato que hoy parezca inofensivo se paga entonces.

## El reparto del trabajo

Cinco encargos en paralelo, más una revisión al final:

- **El generador**: `tutorial/generate.py`, el fichero de inventario inicial y la sección de conceptos.
- **Escenarios 1 a 3**: llegar, crear, orientarse.
- **Escenarios 4 a 6**: asignar, empezar, el arrendamiento.
- **Escenarios 7 a 10**: la vida de `TASK-19` hasta que se cierra.
- **Escenarios 11 a 13**: los errores, archivar y el lote.
- **La revisión**: continuidad del tablero de punta a punta, que ningún escritor individual puede ver, y
  que cada `origen` declarado como literal lo sea de verdad.


## Lo que cambio al implementarlo

Tres desviaciones de este diseno, con su motivo. Se anotan aqui porque el documento tiene que seguir
describiendo lo que hay.

**El escenario 12 archiva `TASK-52` y no `TASK-40`.** La leccion de ese escenario es que archivar
vacia el arrendamiento, y `TASK-40` no tiene ninguno, asi que con ella la regla no se podria ensenar.
`TASK-52` si: el agente se la queda en el escenario 6 reclamando el arrendamiento vencido de `@bob`,
asi que archivarla en el 12 cierra ese hilo y muestra la reserva vaciandose. De paso el escenario
ensena que `biso delete` no existe, cuyo mensaje la especificacion da literal.

**Hay tres herramientas y no una.** Ademas del generador hicieron falta dos comprobadores, y los dos
por el mismo motivo: habia numeros en los fixtures que nadie podia verificar leyendo un solo fichero.
`tutorial/urgency.py` calcula la urgencia de cada tarea desde `tutorial/tablero.yaml` con el
desglose de cada termino, y `tutorial/continuity.py` comprueba que los contadores del tablero
encadenan entre escenarios. El segundo cazo un desfase de uno que cuatro escenarios arrastraban por
no contar la `TASK-62` que crea el escenario 2.

**`tutorial/tablero.yaml` declara cuatro cosas que la especificacion no da.** El dia que el tutorial
toma como hoy, las fechas de creacion de las nueve tareas, la dependencia de `TASK-40` sobre
`TASK-11`, y que la clave `me` del tablero se queda sin configurar. Las cuatro hacen falta para que
la urgencia y el orden de las listas sean reproducibles, y las cuatro estan razonadas en
`tutorial/lagunas/04-06.md`. La de `me` es la mas interesante, porque la seccion 10.10 de la
especificacion exige dejarla sin configurar en un tablero compartido y su propio ejemplo la
configura.

**Las claves de los fixtures van todas en español y el código va todo en inglés.** Los ficheros de
`tutorial/escenarios/` son documentación con forma de datos, así que sus claves son españolas; dos se
habían quedado en inglés (`cmd` y `exit`) y se renombraron a `comando` y `codigo_salida`. Los tres
scripts son código, así que van enteros en inglés, incluidos comentarios y mensajes, con la única
excepción de las cadenas que el generador emite dentro de la página, agrupadas en un bloque marcado.
La regla está escrita en el `CLAUDE.md` de la raíz.

**El detalle operativo del tutorial vive en `tutorial/CLAUDE.md`.** El de la raíz solo lo referencia
con un enlace, sin `@import`, para que un agente lo cargue cuando trabaje en esa carpeta y no antes.
Referenciar en vez de importar tiene además una ventaja medida en esta máquina: opencode no expande
los imports, así que un `@import` en la raíz obligaría a mantener un `opencode.json` en paralelo, y
`link-agent-instructions.sh --check` confirma que sin él los tres agentes siguen leyendo lo mismo.

**`origen` cita una ruta de `docs/spec/` y no un número de sección.** Así se escribió mientras
`docs/SPEC.md` era un solo fichero con secciones numeradas, pero `main` lo repartió en los treinta y
cuatro ficheros de `docs/spec/` y pasó a citar por título en todo el repositorio. Las cincuenta y
nueve citas de los fixtures se tradujeron mecánicamente con `tools/mapa-de-secciones.txt`, que ya
existía para ese reparto, y el generador pasó a enlazar cada ruta derivada al fichero real en vez de
limitarse a mostrar un número que ya no significaba nada.
