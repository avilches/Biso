# Códigos de salida

Tabla global. Ningún comando usa un código fuera de esta tabla, y ningún código tiene dos
significados. Quien llama puede ramificar sobre el número sin leer el mensaje.

| Código | Nombre | Significado | Ejemplo |
|---:|---|---|---|
| 0 | `OK` | La operación terminó y se aplicó | `biso new "Algo"` |
| 1 | `INTERNAL` | Fallo no previsto del programa | una excepción no capturada |
| 2 | `USAGE` | La línea de comandos está mal formada | bandera desconocida, falta un obligatorio, banderas incompatibles, identificador mal formado, bandera de escritura en un comando de lectura |
| 3 | `BAD_VALUE` | El valor que llega es sintácticamente correcto pero el tablero no lo reconoce, o un dato guardado no se puede interpretar | `--status "Pending"` en un tablero cuyos estados son otros |
| 4 | `NOT_FOUND` | La entidad referida no existe | `biso get TASK-999` |
| 5 | `AMBIGUOUS` | La referencia encaja con más de una entidad | `biso get "parser"` con tres coincidencias |
| 6 | `PRECONDITION` | La operación es válida, pero el estado actual del tablero no la permite o no la satisface | `biso finish --strict` con criterios sin marcar, o `biso doctor` con problemas pendientes |
| 7 | `ENVIRONMENT` | Falla el entorno, no la petición | el almacén no responde, no hay permisos, no se puede adquirir el acceso exclusivo, no hay terminal donde hace falta |
| 8 | `NO_BOARD` | No hay tablero accesible desde donde se ha llamado | cualquier comando fuera de un tablero, salvo `init`, `help`, `--help` y `--version`, que no necesitan uno; `biso where` también devuelve 8 cuando no encuentra ninguno |
| 9 | `VALIDATION` | Una validación previa ha fallado y **no se ha escrito nada** | `biso new --from tareas.ndjson` con la línea 47 inválida |
| 10 | `DAMAGED` | El tablero está donde tiene que estar, y su almacén no se puede leer | `board.db` que no abre, o que falla su comprobación de integridad (["Qué pasa con un dato que no se puede interpretar"](garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar)) |
| 11 | `AMBIGUOUS_BOARD` | El mismo identificador de tablero aparece en dos sitios, y elegir uno sería escribir en el que nadie ha nombrado | dos raíces de la sección ["Configuración de máquina"](invocacion.md#configuración-de-máquina) con una carpeta que lleva el mismo marcador (["Cómo se elige el tablero"](resolucion-del-tablero.md)) |

Las reglas que acompañan a la tabla:

- **El código 9 garantiza que no se ha escrito nada.** Si un comando termina con 9, el tablero está
  exactamente como estaba antes. Por eso una validación fallida dentro de un lote se reporta como 9 y
  no como 3 ni como 4, y por eso el 9 llega siempre con el detalle de **todos** los fallos
  encontrados, no solo del primero.
- **Un listado vacío es siempre 0.** Un tablero donde de verdad no hay nada que cumpla un filtro
  válido no es un error.
- **El código 3 cubre dos direcciones.** Un valor de entrada que el tablero no reconoce, y un dato ya
  guardado que el programa no sabe interpretar. Las dos son "el vocabulario no cuadra", y el mensaje
  siempre dice cuál de las dos ha ocurrido.
- **El código 1 es un fallo del programa, no de quien llama.** La reacción correcta es informar, no
  reintentar con otros parámetros.
- **Los códigos 8, 10 y 11 son los tres desenlaces malos de resolver el tablero, y son tres porque el
  remedio de cada uno es otro.** Con el 8 no hay tablero y `biso init` lo crea; con el 10 el tablero
  está ahí y hay que reconstruirlo desde una instantánea con `biso init --from`, que es lo único que lo
  arregla; con el 11 hay dos y hay que quitar o renombrar uno de los dos directorios a mano. Por eso el
  daño del almacén no comparte número con la ausencia de tablero, aunque para quien llama las tres
  frases empiecen igual: quien ramifica sobre el número tiene que poder elegir el remedio sin leer el
  mensaje, que es el principio de la sección ["Los principios"](principios.md). **Ni el 10 ni el 11 aparecen en la tabla de códigos de
  salida de cada comando**, porque no son desenlaces propios de ninguno sino del tablero entero, igual
  que el 1. La excepción es `biso where`, que existe justamente para explicar la resolución y los lleva
  los dos en su tabla (["`biso where`"](cmd/where.md)).

---

