# Los comandos

Veinte comandos. Los once primeros son el ciclo de trabajo y aparecen en `biso --help`; los nueve
restantes son de administración y aparecen en `biso help all`. De esos once, diez son los que 9.6
cuenta como "las órdenes del ciclo de trabajo" del bloque `COMMANDS` de `biso prime`: `prime` es el
undécimo, y no se lista a sí mismo en su propio mensaje.

| Comando | Qué hace | En `biso --help` |
|---|---|---|
| `prime` | El mensaje de arranque de la sección 9 | sí |
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

**Las banderas globales de la sección 3 valen en todos ellos y no se repiten en las tablas de
parámetros de cada comando.** Un comando solo las menciona cuando le impone una restricción
adicional, y esas restricciones son exactamente cuatro en todo el documento: `prime --full` no se
combina con `--json`, `config` solo acepta `--json` en su subcomando `list`, `export` rechaza
`--json` con código 2, y `--print` y `--dry-run` no valen donde no tienen nada que hacer, cada una
en su propia lista, según la regla de la sección 3.

El caso de `export` merece una línea, porque es el único comando cuya salida ya es JSON sin pedirlo:
son objetos JSON, uno por línea, y `--json` pide el sobre único de la sección 12, que es otra forma
distinta. Pasarlo es error 2:

```
error: --json does not apply to export
       its output is already one JSON object per line
```

**Todos los comandos que escriben aceptan todas las banderas de campo de la sección 8**, con el mismo
nombre y el mismo significado. Eso vale para `new`, `set`, `start`, `note`, `comment`, `finish`, `ask`,
`answer` y `archive`. Lo que distingue a unos de otros no es qué campos aceptan, sino qué hacen por
defecto. Las tablas de parámetros de cada comando enumeran solo lo que es propio de ese comando.

