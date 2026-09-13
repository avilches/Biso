# Identificador de tarea

- Un identificador es `<PREFIX>-<n>`, con `n` entero positivo. `PREFIX` viene de la configuración
  (`task_prefix`).
- **`task_prefix` no tiene un valor fijo por defecto: se deriva del nombre del tablero
  (`project_name`) en mayúsculas.** Dos tableros con `MYP` como valor fijo colisionarían los dos en
  `MYP-1`, y eso haría inservible cualquier vista que junte tareas de varios proyectos.
- **Los ejemplos de esta especificación pertenecen todos al mismo tablero ficticio, llamado `My
  project`, con `task_prefix` fijado a `MYP`.** Es el prefijo que aparece en los identificadores de
  ejemplo del resto del documento (`MYP-11`, `MYP-60`...), fijado a mano en vez de derivarse del
  nombre (que daría `MYPROJECT`), para que los ejemplos se lean como parte de un mismo tablero
  coherente.
- **La derivación quita del nombre los caracteres que no son letras y pasa el resto a mayúsculas**,
  así que un tablero llamado `mi-proyecto-2` da el prefijo `MIPROYECTO`. Si al quitarlos no queda
  ninguna letra, como en un tablero llamado `2026`, `biso init` no se inventa un valor: falla y pide
  el prefijo explícitamente con `--prefix` (error 2, `code` [`invalid_prefix`](../contrato-json.md#los-identificadores-de-error)), la misma
  clave que ya cubre un `--prefix` con algo que no sean letras (la sección [`biso init`](../cmd/init.md)).
- **"Letra" no incluye los diacríticos**, para que un nombre de tablero con cualquier carácter
  Unicode derive un prefijo predecible. La derivación pasa primero el nombre por el paso de
  `normalizar(x)` (sección ["El algoritmo de coincidencia"](../vocabularios.md#el-algoritmo-de-coincidencia)) que quita los acentos, las diéresis y las cedillas, y solo entonces
  se queda con lo que sean letras ASCII. Así un tablero llamado `Peña` deriva `PENA`, y uno llamado
  `Café` deriva `CAFE`.
- **Un identificador no se reutiliza jamás**, ni después de archivar una tarea ni después de
  eliminarla por cualquier vía.
- Los identificadores se asignan de forma creciente, pero **la especificación no promete que la
  secuencia no tenga huecos**. Un hueco es normal y nunca es un error.
- El tablero sabe en todo momento cuál es el identificador más alto que ha llegado a asignar, y ese
  dato se usa en los mensajes de la sección ["Los tres mensajes de \"no la encuentro\""](../referencias.md#los-tres-mensajes-de-no-la-encuentro) y en `biso doctor`.
