# Los comandos

Los comandos que aparecen en `biso --help` son el ciclo de trabajo; los que no aparecen ahí son de
administración y solo se listan con `biso help all`. De los que aparecen en `biso --help`, todos salvo
`prime` son los que la sección ["Qué entra en el mensaje y qué se relega a `--help`"](prime.md#qué-entra-en-el-mensaje-y-qué-se-relega-a---help)
cuenta como "las órdenes del ciclo de trabajo" del bloque `COMMANDS` de `biso prime`: `prime` no se
lista a sí mismo en su propio mensaje.

| Comando | Qué hace | En `biso --help` |
|---|---|---|
| `prime` | El mensaje de arranque de la sección ["`biso prime`, el arranque de una sesión"](prime.md) | sí |
| `ls` | Lista tareas | sí |
| `get` | Muestra una tarea | sí |
| `new` | Crea una o muchas tareas | sí |
| `set` | Cambia campos de una o varias tareas | sí |
| `start` | Toma una tarea y la pone en curso | sí |
| `note` | Añade una nota de implementación | sí |
| `comment` | Añade un comentario con autor | sí |
| `finish` | Cierra una tarea | sí |
| `ask` | Aparca una tarea en una pregunta abierta | sí |
| `answer` | Responde la pregunta abierta y desaparca la tarea | sí |
| `archive` | Saca una tarea del tablero activo | no |
| `export` | Vuelca el tablero en el formato de entrada de `new --from` | no |
| `init` | Crea un tablero | no |
| `where` | Explica qué tablero se está usando y por qué | no |
| `config` | Lee y cambia la configuración | no |
| `doctor` | Comprueba y repara la integridad | no |
| `board` | Abre la interfaz interactiva | no |
| `help` | La ayuda de primer nivel y la de cada comando | no |
| `snapshot` | Escribe la instantánea del tablero en su propio directorio y la guarda en el control de versiones | no |

**Las banderas globales de la sección ["Banderas globales"](flags-globales.md#banderas-globales) valen en todos ellos y no se repiten en las tablas de
parámetros de cada comando.** Un comando solo las menciona cuando le impone una restricción
adicional, y estas son todas las restricciones que hay en todo el documento: `prime --full` no se
combina con `--json`, `config` solo acepta `--json` en su subcomando `list`, `export` rechaza
`--json` con código 2, y `--print` y `--dry-run` no valen donde no tienen nada que hacer, cada una
en su propia lista, según la regla de la sección ["Banderas globales"](flags-globales.md#banderas-globales).

El caso de `export` merece una línea, porque es el único comando cuya salida ya es JSON sin pedirlo:
son objetos JSON, uno por línea, y `--json` pide el sobre único de la sección ["El contrato JSON"](../contrato-json.md), que es otra forma
distinta. Pasarlo es error 2:

```
error: --json does not apply to export
       its output is already one JSON object per line
```

**Todos los comandos que escriben aceptan todas las banderas de campo de la sección ["Las familias de banderas"](../familias-de-flags.md)**, con el mismo
nombre y el mismo significado. Eso vale para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`,
`answer` y `archive`. Lo que distingue a unos de otros no es qué campos aceptan, sino qué hacen por
defecto. Las tablas de parámetros de cada comando enumeran solo lo que es propio de ese comando.

