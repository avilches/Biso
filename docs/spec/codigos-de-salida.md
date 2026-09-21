# Códigos de salida

Tabla global. Ningún comando usa un código fuera de esta tabla, y ningún código tiene dos
significados. Quien llama puede ramificar sobre el número sin leer el mensaje.

| Código | Nombre | Significado | Ejemplo |
|---:|---|---|---|
| 0 | `OK` | La operación terminó y se aplicó | `biso new "Algo"` |
| 1 | `INTERNAL` | Fallo no previsto del programa | una excepción no capturada |
| 2 | `USAGE` | La línea de comandos está mal formada | flag desconocido, falta un obligatorio, flags incompatibles, identificador mal formado, flag de escritura en un comando de lectura |
| 3 | `BAD_VALUE` | El valor que llega es sintácticamente correcto pero el tablero no lo reconoce, o un dato guardado no se puede interpretar | `--status "Pending"` en un tablero que no tiene `Pending` entre sus estados |
| 4 | `NOT_FOUND` | La entidad referida no existe | `biso get MYP-999` |
| 5 | `AMBIGUOUS` | La referencia encaja con más de una entidad | `biso get "parser"` con tres coincidencias |
| 6 | `PRECONDITION` | La operación es válida, pero el estado actual del tablero no la permite o no la satisface | `biso finish --strict` con criterios sin marcar, `biso doctor` con problemas pendientes, o `--below` sobre una tarea que no tiene clave de orden manual |
| 7 | `VALIDATION` | Una validación previa ha fallado y **no se ha escrito nada** | `biso new --from tareas.ndjson` con la línea 47 inválida |
| 8 | `ENVIRONMENT` | Falla el entorno, no la petición | el almacén no responde, no hay permisos, no se puede adquirir el acceso exclusivo, no hay terminal donde hace falta |
| 20 | `NO_BOARD` | No hay tablero accesible desde donde se ha llamado | cualquier comando fuera de un tablero, salvo `init`, `help`, `--help` y `--version`, que no necesitan uno; `biso where` también devuelve 20 cuando no encuentra ninguno |
| 21 | `DAMAGED` | El tablero está donde tiene que estar, y su almacén no se puede leer | `board.db` que no abre, o que falla su comprobación de integridad (["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) |
| 22 | `AMBIGUOUS_BOARD` | El mismo identificador de tablero aparece en dos sitios, y elegir uno sería escribir en el que nadie ha nombrado | dos raíces de la sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina) con una carpeta que lleva el mismo marcador (["Cómo se elige el tablero"](resolucion-del-tablero.md)) |

Los códigos del 20 en adelante son los desenlaces malos de resolver el tablero, y no los desenlaces
propios de ningún comando.

Las reglas que acompañan a la tabla, cada una en su propia sección para poder enlazarla:

## El código 7 garantiza que no se ha escrito nada, y el código específico siempre gana sobre él

Si un comando termina con 7, el tablero está exactamente como estaba antes, pero el 7 no es el
código general de "una validación falló dentro de un lote": un fallo que ya tiene su propio código
(3, 4, 5...) lo conserva sea cual sea el tamaño del lote. `biso set A B C` donde `C` no existe
devuelve 4, tanto si `C` es la única referencia como si hay cien. El 7 queda reservado en exclusiva
a los fallos que no se pueden atribuir a un elemento señalable, y en la versión 1.0 hay uno solo,
con su propio identificador en la sección ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error):
`batch_invalid`, la validación de conjunto de un lote que no apunta a un elemento concreto, como un
`new --from` con varias líneas inválidas por motivos distintos. El 7 llega entonces con el detalle
de **todos** los fallos encontrados, no solo del primero. Si un programa que llama a `biso` quisiera
tratar todo lote fallido igual, tendría que analizar `error.details` de todas formas, que es justo lo
que un código específico existe para evitar.

**`--dry-run` no tiene código propio, y por eso no hay ningún `dry_run_failed`.** Una vista previa
contesta lo que contestaría la llamada que simula, con el código del fallo que encontró
(["Los flags globales"](cmd/flags-globales.md)), así que el único `--dry-run` que sale con 7 es el
de un lote, que ya salía con 7 sin él. Los comandos que escriben sobre tareas y no tienen lote
(`set`, `archive` y los seis verbos del ciclo) no llevan, por tanto, ninguna fila del 7 en su tabla
de códigos: no hay forma de llegar a él desde ahí.

## Un listado vacío es siempre 0

Un tablero donde de verdad no hay nada que cumpla un filtro válido no es un error.

## El código 3 cubre dos direcciones

Un valor de entrada que el tablero no reconoce, y un dato ya guardado que el programa no sabe
interpretar. Las dos son "el vocabulario no cuadra", y el mensaje siempre dice cuál de las dos ha
ocurrido.

## El código 1 es un fallo del programa, no de quien llama

La reacción correcta es informar, no reintentar con otros parámetros.

## Los códigos 20, 21 y 22 son los tres desenlaces malos de resolver el tablero, y cada uno tiene un remedio distinto

- **20.** No hay ningún tablero accesible. El remedio es `biso init`, que crea uno.
- **21.** El tablero está donde tiene que estar, pero su almacén no se puede leer. El único remedio
  es reconstruirlo desde una instantánea con `biso init --from`.
- **22.** El mismo identificador de tablero aparece en dos sitios. El remedio es a mano: quitar o
  renombrar uno de los dos directorios.

Cada desenlace tiene su propio código, en vez de compartir uno, porque quien ramifica sobre el número tiene que poder elegir el remedio
sin leer el mensaje, que es el principio de la sección ["Los principios"](principios.md). **Ni el 21
ni el 22 aparecen en la tabla de códigos de salida de cada comando**, porque no son desenlaces propios
de ninguno sino del tablero entero, igual que el 1. Hay tres excepciones: `biso where`, que existe
justamente para explicar la resolución y los lleva los dos en su tabla
(["`biso where`"](cmd/where.md)); `biso doctor`, que lleva solo el 21 en la suya porque su base de
datos ilegible es justo uno de los daños que existe para diagnosticar (["`biso doctor`"](cmd/doctor.md));
y `biso init`, que lleva solo el 22, porque recorre las raíces de la máquina por su cuenta para
comprobar que el `id` que va a acuñar o a adoptar no exista ya, y puede darlo desde un proyecto que
todavía no tiene tablero (["`biso init`"](cmd/init.md#códigos-de-salida)).
`biso doctor` no lleva el 22, porque un tablero ambiguo aborta al resolverse, antes de que `doctor`
llegue a abrir ninguna base de datos.

---

