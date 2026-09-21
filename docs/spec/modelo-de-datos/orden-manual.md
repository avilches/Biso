# El orden manual y su clave

`ordinal` es el orden manual de una tarea: el que decide quien mira el tablero cuando el orden por
urgencia de [`biso ls`](../cmd/ls.md) no es el que quiere. Es opcional, y lo normal es que la mayoría
de las tareas de un tablero no lo tengan.

**No es un número, es una clave de texto.** El porqué, con las alternativas descartadas, está en
["El orden manual es una clave de texto"](../../decisiones/detalles.md#el-orden-manual-es-una-clave-de-texto).
Esta página define qué es una clave, cómo se compara y cómo se calcula la que le toca a una tarea al
colocarla. Los flags que la escriben, con sus errores literales, están en
["El orden manual"](../familias-de-flags.md#el-orden-manual).

## Qué es una clave de orden

Una clave es una cadena que cumple estas dos condiciones, y cualquier otra cadena no es una clave:

- **Sus símbolos salen de `0-9a-z`**, dígitos y letras latinas minúsculas sin acento, y la cadena no
  está vacía.
- **No termina en `0`.**

**Dos claves se comparan por sus puntos de código Unicode**, de menor a mayor, exactamente igual que
`--sort title` (["La regla de orden, completa"](../cmd/ls.md#la-regla-de-orden-completa)). Como el
alfabeto son los dígitos y las minúsculas sin acento, esa comparación es la del diccionario y no
depende de la configuración regional de ninguna máquina: `9` va antes que `a`, `a` antes que `ai`, y
`ai` antes que `b`.

**Una clave es una fracción escrita en base 36**, y de ahí salen las dos condiciones. La clave `m` es
la fracción `0.m`, la clave `mi` es `0.mi`, y comparar las dos cadenas por sus puntos de código es
comparar las dos fracciones. El cero final se prohíbe porque `m0` vale lo mismo que `m` como fracción
y, aun así, se ordena detrás: entre las dos no cabría ninguna clave, y ese hueco sería un sitio del
orden donde no se podría insertar nada. Sin ceros finales, cada fracción tiene una sola grafía y
entre dos claves distintas siempre cabe otra.

**La clave no se teclea nunca.** Se escribe colocando la tarea (`--ordinal first`, `--ordinal last`,
`--above <ref>`, `--below <ref>`) y se quita con `--clear-ordinal`.

**La clave cruda solo viaja en el JSON**: en el sobre de [`biso ls`](../cmd/ls.md#el-esquema-json) y
en el de [`biso get`](../cmd/get.md#el-esquema-json), en la exportación de
[`biso export`](../cmd/export.md) y en el lote de [`biso new --from`](../cmd/new.md#el-modo-lote).
Ahí viaja como cadena, tal cual está guardada, o como `null` si la tarea no tiene ninguna, y eso es
lo que hace exacta la ida y vuelta entre `export` y `new --from`. La ficha de texto de `biso get` no
la imprime, porque no se puede teclear y no dice nada que quien la lee pueda usar: imprime `manual`
cuando la tarea tiene clave y `-` cuando no (["`biso get`"](../cmd/get.md#salida)).

**Una tarea sin clave no está la última de nada**: las tareas con clave van todas antes que las que
no la tienen, y entre estas últimas manda la urgencia (["La regla de orden,
completa"](../cmd/ls.md#la-regla-de-orden-completa)).

## El algoritmo del punto medio

Colocar una tarea es escribirle una clave que caiga dentro de un hueco, es decir, estrictamente entre
la clave de la tarea que tiene que quedar por encima y la de la que tiene que quedar por debajo.
Cualquiera de las dos puede faltar, y entonces el hueco llega hasta el extremo del orden.

**`clave_entre(anterior, siguiente)`** es la función que la calcula. `anterior` es la clave que tiene
que quedar por encima, es decir la menor de las dos, o nada; `siguiente` es la que tiene que quedar
por debajo, o nada. El valor de un símbolo es su posición dentro de `0123456789abcdefghijklmnopqrstuvwxyz`,
de 0 para el `0` a 35 para la `z`:

1. Si no hay `anterior`, vale la cadena vacía.
2. Si hay `siguiente`, se toma el prefijo común de las dos claves, comparando símbolo a símbolo y
   tomando `0` allí donde `anterior` ya se haya acabado. Si ese prefijo común no está vacío, el
   resultado es el prefijo seguido de `clave_entre` de las dos colas que quedan detrás de él.
3. Llegados aquí, las dos claves difieren ya en su primer símbolo. Sea `a` el valor del primer
   símbolo de `anterior`, o 0 si `anterior` está vacía, y sea `b` el valor del primer símbolo de
   `siguiente`, o 36 si no hay `siguiente`.
4. Si entre `a` y `b` cabe algún valor, es decir si `b` menos `a` es mayor que uno, el resultado es
   el símbolo cuyo valor es la parte entera de la media de los dos.
5. Si no cabe ninguno y `siguiente` tiene más de un símbolo, el resultado es el primer símbolo de
   `siguiente`.
6. Si no cabe ninguno y `siguiente` no existe o tiene un solo símbolo, el resultado es el símbolo de
   valor `a` seguido de `clave_entre` de la cola de `anterior` y nada.

**El resultado es siempre una clave válida**, estrictamente mayor que `anterior` y estrictamente
menor que `siguiente`, y la recursión siempre termina. Esto es lo que hace cierta la promesa de que
entre dos claves cualesquiera cabe otra, de que siempre hay una menor que la menor y de que siempre
hay una mayor que la mayor.

Estos son los casos que la implementación tiene que reproducir carácter a carácter:

| `anterior` | `siguiente` | Clave nueva | Qué caso ejercita |
|---|---|---|---|
| nada | nada | `i` | el tablero no tiene ninguna clave todavía |
| nada | `i` | `9` | colocar delante de todas |
| `i` | nada | `r` | colocar detrás de todas |
| `i` | `r` | `m` | el punto medio corriente |
| `a` | `c` | `b` | entre los dos primeros símbolos cabe uno |
| `a` | `b` | `ai` | símbolos consecutivos: la clave crece un símbolo |
| `9` | `a` | `9i` | el salto del último dígito a la primera letra |
| `mm` | `mn` | `mmi` | prefijo común, y detrás el mismo caso de arriba |
| `i` | `j` | `ii` | `siguiente` es de un solo símbolo |
| `zz` | nada | `zzi` | el extremo de arriba, con la clave mayor posible de su longitud |
| nada | `01` | `00i` | el extremo de abajo, con una clave que empieza por cero |

**El caso peor añade un símbolo cada cinco inserciones**, y es meter siempre la tarea en el mismo
hueco. Insertando una y otra vez delante de todas, las claves van `i`, `9`, `4`, `2`, `1`, `0i`,
`09`, `04`, `02`, `01`, `00i`: cinco inserciones por cada símbolo nuevo, que es lo que cuesta partir
en dos un alfabeto de treinta y seis símbolos.

## El hueco de cada colocación

Cada forma de colocar una tarea se traduce a un hueco, y el hueco a la clave con la función de
arriba:

| Petición | `anterior` | `siguiente` |
|---|---|---|
| `--ordinal first` | nada | la menor clave del tablero |
| `--ordinal last` | la mayor clave del tablero | nada |
| `--above <ref>` | la mayor clave estrictamente menor que la de la vecina | la clave de la vecina |
| `--below <ref>` | la clave de la vecina | la menor clave estrictamente mayor que la de la vecina |

Precisiones que hacen determinista el cálculo:

- **El hueco se calcula sobre el tablero entero**, con las tareas archivadas y las terminadas
  incluidas. La clave es global (["Los grupos de una vista no tienen clave
  propia"](#los-grupos-de-una-vista-no-tienen-clave-propia)), así que dónde cae una tarea no puede
  depender de los filtros de nadie.
- **Las claves de las tareas que la propia llamada mueve no cuentan** al buscar la vecina anterior o
  la siguiente. Así, mover una tarea justo por encima de la que ya tenía debajo no gasta símbolos de
  más, y una llamada que mueve un bloque entero no se estorba a sí misma.
- **Los extremos no fallan nunca.** `--above` sobre la tarea de clave menor y `--below` sobre la de
  clave mayor dejan una de las dos vecinas vacía, y la función siempre encuentra una clave menor que
  la menor y otra mayor que la mayor.
- **Si el tablero no tiene ninguna clave**, `--ordinal first` y `--ordinal last` dejan las dos
  vecinas vacías y escriben la misma clave, `i`.
- **Dos tareas pueden compartir clave**, porque nada obliga a que sean únicas y un lote de
  `biso new --from` puede traer dos iguales. La vecina anterior es la de la mayor clave
  **estrictamente** menor, y la siguiente la de la menor clave **estrictamente** mayor, así que un
  empate no deja ningún hueco vacío: la tarea nueva queda por encima o por debajo de las dos
  empatadas, nunca entre ellas. En el listado, dos claves iguales las desempata el identificador,
  como cualquier otro empate (["La regla de orden, completa"](../cmd/ls.md#la-regla-de-orden-completa)).

## Varias tareas en la misma llamada

Una llamada puede mover varias tareas de una vez, y entonces todas caen dentro del mismo hueco **en
el orden en que se escribieron las referencias**, que es lo que hace que `biso set A B --below C`
deje `C`, `A`, `B`, y que `biso set A B --above C` deje `A`, `B`, `C`.

La regla es una sola: se calcula el hueco una vez, y las tareas lo van ocupando por orden. La primera
toma `clave_entre(anterior, siguiente)`; a partir de ahí, `anterior` pasa a ser la clave que se acaba
de escribir y la siguiente tarea toma el punto medio del hueco que queda. Una tarea nombrada dos
veces en la misma llamada se escribe una sola vez, como en cualquier otra escritura
(["`biso set`"](../cmd/set.md#comportamiento-caso-a-caso)), así que ocupa un solo sitio.

Con un tablero donde la tarea `MYP-40` tiene la clave `m` y la siguiente que tiene clave es `t`,
`biso set MYP-7 MYP-19 MYP-23 --below MYP-40` escribe `p` a `MYP-7`, `r` a `MYP-19` y `s` a `MYP-23`.

## No hay renumerado

**La longitud de una clave no tiene tope y ningún comando reparte los huecos de nuevo.** No hace
falta: el caso peor de arriba añade un símbolo cada cinco inserciones en el mismo sitio, así que una
clave larga es la huella de un tablero que se ha reordenado mucho en el mismo punto, no una avería.
Si algún día hiciera falta, sería un comando nuevo, que es un añadido compatible con
["El contrato de estabilidad"](../estabilidad.md) y no un cambio de este campo.

## Una clave guardada que no cumple la regla

**Esto rige al escribir.** Una clave ya guardada que no cumpla el alfabeto o que acabe en `0`, porque
entró por una vía que no pasa por esta validación, no es un error nuevo distinto: es un dato que el
programa no puede interpretar, y se trata con la regla general de
["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
Es la misma regla que ya gobierna una etiqueta guardada fuera de su alfabeto
(["El juego de caracteres de un token"](../valores-de-entrada.md#el-juego-de-caracteres-de-un-token)).

## Los grupos de una vista no tienen clave propia

**Cada tarea tiene una sola clave, y es global.** Las agrupaciones de [`biso board`](../cmd/board.md)
son presentación: reordenar una tarea dentro de un grupo escribe la clave que le toca entre sus
vecinas **de ese grupo**, y como la clave es una sola, la posición de la tarea en la lista sin
agrupar puede moverse como consecuencia. Es el precio de tener una clave sola, y está aceptado en la
decisión. Cruzar de un grupo a otro no reordena nada: edita el campo por el que se agrupa
(["Qué escribe un arrastre"](../cmd/board.md#qué-escribe-un-arrastre)).

---
