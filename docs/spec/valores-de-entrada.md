# Cómo se pasa un valor

## Tres formas de pasar un valor largo

Todo parámetro de tipo texto largo (`--desc`, `--plan`, `--note`, `--summary`, `--comment` y el texto
de un criterio o de un elemento de la definición de hecho) acepta las tres:

| Forma | Significado |
|---|---|
| `--desc "texto"` | el texto literal |
| `--desc @ruta/fichero.md` | el contenido del fichero, interpretado como UTF-8 |
| `--desc -` | todo lo que llegue por la entrada estándar hasta el fin de fichero |

Reglas:

- **Un texto que empieza de verdad por `@` se escribe `@@`.** El primer `@` se descarta y el resto es
  literal. Es la única secuencia de escape del programa.
- **Los campos de persona nunca interpretan el `@`.** `--assignee`, `--reporter` y `--comment-author`
  toman su valor tal cual, así que `--comment-author @trello:juan` guarda ese texto y no intenta leer
  ningún fichero.
- **`-` solo puede aparecer una vez por invocación.** Dos parámetros que pidan la entrada estándar son
  un error de uso con código 2, porque el segundo leería un flujo agotado y guardaría el vacío sin
  que se note.
- **Un fichero que no existe es código 4**, con el mensaje `error: --desc: file not found: docs/x.md`.
  Un fichero que existe pero no se puede leer es código 7.
- **Un valor vacío, venga de donde venga, no borra nada.** Ver 4.6.

## El valor vacío

Un valor vacío es una cadena sin ningún carácter, o solo con espacios, tanto si llega literalmente
como si llega de un fichero vacío o de una entrada estándar vacía. La regla es única:

| Dónde | Qué pasa |
|---|---|
| En una bandera que añade (`--note`, `--label`, `--ac`, `--desc`) | No se añade nada, se emite `warning: --note: empty value, nothing was added` y el código sigue siendo 0 |
| En una bandera que sustituye (`--set-notes`, `--set-label`) | Deja el campo vacío, igual que `--clear-notes`. Sustituir por nada es vaciar, y eso sí es explícito |
| En un campo escalar (`--type ""`, `--priority ""`) | Error 3. **La cadena vacía nunca es la forma de borrar un escalar**; para eso está `--clear-type` |
| En el título, al crear | Error 2: `error: title cannot be empty` |

**El `code` de un escalar vacío depende de si el campo tiene vocabulario cerrado.** Para `status`,
`type`, `priority` y `project`, una cadena vacía es un valor que no coincide con nada configurado, así
que sigue la regla de 6.1 y el `code` es el de un valor desconocido (`unknown_status` y análogos, con
el mensaje de 6.2). Para los demás escalares (`--reporter ""`, `--ordinal ""`, `--due ""`), que no
tienen vocabulario, el `code` es `empty_scalar_value`.

## Valores que empiezan por guion

Tres mecanismos, en orden de preferencia:

1. **`--flag=valor`** funciona siempre y es la forma recomendada: `--desc=-5 grados`.
2. **`--`** termina el análisis de opciones: `biso new -- "-n no es una bandera"`.
3. **Un valor que empieza por guion detrás de una bandera que exige valor se acepta tal cual**, sin
   heurísticas. `biso set TASK-1 --note -x` guarda `-x` como nota.

Como consecuencia de la regla 3, olvidar el valor de una bandera se detecta por lo que sobra después,
no por lo que parece: `biso set TASK-1 --note --priority high` guarda la nota `--priority` y luego
falla con código 2 y `error: unexpected argument: high`.

## Repetición y listas separadas por comas

Para toda bandera marcada como repetible:

- Repetirla acumula: `--label a --label b` deja dos etiquetas.
- Si además acepta lista, separar por comas acumula igual: `--label a,b` deja las mismas dos.
- Las dos formas se pueden mezclar.
- **Una coma dentro de un valor se escapa con `\,`.** Es la única forma de meter una coma en una
  etiqueta o en una referencia.
- Los campos de texto largo y los criterios **nunca** se parten por comas.
- Un valor repetido dentro de la misma bandera se guarda una vez y produce
  `warning: --label: "urgent" given twice, kept once`.

Para toda bandera **no** repetible, es decir, los campos escalares, pasarla dos veces con valores
distintos es un error de uso con código 2:

```
error: --status given twice with different values: "In Progress" and "Done"
```

