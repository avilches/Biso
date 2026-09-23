# Las relaciones entre tareas

Una tarea puede apuntar a otras de tres maneras, y cada una contesta una pregunta distinta. Esta
página dice qué significa cada una y cómo se elige entre ellas. Lo que el programa hace con ellas
vive en la página del comando o del cálculo donde se nota, y se enlaza desde aquí en vez de
repetirse.

| Campo | La pregunta que contesta | Cuántos |
|---|---|---|
| `parent` | ¿de qué trabajo mayor es parte esto? | una tarea como mucho |
| `dependencies` | ¿qué tiene que estar hecho antes que esto? | las que hagan falta |
| `references` | ¿qué otra cosa hay que mirar? | las que hagan falta |

## `parent`, la contención

Dice que una tarea es parte de otra. Una tarea tiene **como mucho un padre**, porque una cosa es
parte de una sola cosa mayor, y por eso el campo es un escalar y no una lista.

La prueba para saber si dos tareas están en esta relación: **¿al terminar la grande tiene sentido
preguntar si la pequeña está hecha?** Si la respuesta es sí, hay parentesco. Si es no, no lo hay,
por muy relacionadas que estén las dos.

Esa pregunta no es retórica: es exactamente lo único que el programa hace con el campo, el aviso de
subtareas sin terminar de ["`biso finish`"](../cmd/verbos-del-ciclo.md#biso-finish). Lo demás que
toca al parentesco es lectura, el filtro `--parent` de ["`biso ls`"](../cmd/ls.md), que devuelve las
hijas directas y no las nietas. Cualquier tarea con hijas actúa como agrupación sin que haya que
marcarla de ninguna forma, y por eso no hay ningún campo dedicado a agrupar
(["Lo que se deja fuera a propósito"](../fuera-de-alcance.md)).

## `dependencies`, la precedencia

Dice que una tarea no se puede hacer antes que otra.

**La dirección es siempre la misma, y es lo que más se confunde: `dependencies` son las tareas que
bloquean a esta, no las que esta bloquea.** `biso set MYP-10 --add-deps MYP-4` significa que `MYP-4`
va primero y que `MYP-10` espera por ella. La relación contraria no se escribe nunca: `blocks` es su
inversa y se calcula al leer, como el resto de
["Los campos derivados"](index.md#los-campos-derivados). La ayuda de
["`biso set`"](../cmd/set.md#biso-set---help) y la de
["`biso new`"](../cmd/new.md#biso-new---help) lo dicen con ese mismo ejemplo, porque escribirla al
revés es válido y el programa no lo detecta, y el mensaje de arranque lo repite en una regla
(["La ayuda enseña la dirección de una dependencia"](../../decisiones/detalles.md#la-ayuda-enseña-la-dirección-de-una-dependencia)).

La prueba: **¿si hago esta primero, el trabajo se tira o se rehace?** Si la respuesta es sí, es una
dependencia.

El caso débil, el de "va mejor después" sin que nada llegue a romperse, también cabe aquí, y es
deliberado: una dependencia no cierra ninguna puerta, porque ["`biso start`"](../cmd/verbos-del-ciclo.md#biso-start)
sobre una tarea bloqueada avisa y la empieza igual. Una cadena de tareas que solo expresa un orden
preferido es información para quien elige, no una verja.

Las dependencias son la única de estas relaciones que entra en un cálculo, con los dos términos de
["La urgencia"](urgencia.md#la-urgencia) y los filtros `--blocked` y `--not-blocked` de
["`biso ls`"](../cmd/ls.md), y la única que se valida al escribirla, con la existencia de la tarea y
la detección de ciclos de ["Las familias de flags"](../familias-de-flags.md#campos-de-lista-que-admiten-coma).

## Los punteros: `references`

Todo lo demás que haya que mirar y que no cambie ni el alcance del trabajo ni su orden: la página
que lo gobierna, el informe del que salió, una dirección web, incluso el identificador de otra tarea
con la que hay que ser coherente.

**Es un solo campo, y un documento es una referencia más.** No hay un segundo campo `documentation`
para las páginas y las rutas de documentos: la distinción no la sostenía ningún uso ni ninguna
regla, y por qué se retiró está en
["Se retira `documentation` y `references` queda como único campo de punteros"](../../decisiones/detalles.md#se-retira-documentation-y-references-queda-como-único-campo-de-punteros).
Tampoco hay un campo para los ficheros que tocó el trabajo: una ruta que valga la pena señalar es
una referencia más, y por qué se retiró está en
["Se retira `modifiedFiles`"](../../decisiones/detalles.md#se-retira-modifiedfiles).

Son texto libre, sin el alfabeto cerrado que sí tienen las etiquetas y las personas
(["El juego de caracteres de un token"](../valores-de-entrada.md#el-juego-de-caracteres-de-un-token)).
No se resuelven a ninguna tarea, y **no los alcanza la búsqueda de texto**
(["La búsqueda por texto"](../referencias.md#la-búsqueda-por-texto)). **Sí se valida una cosa:** que
ningún elemento lleve un `\r` o un `\n` literal, la misma regla que rige cualquier `string` de la tarea
(["El salto de línea en un campo `string`"](../valores-de-entrada.md#el-salto-de-línea-en-un-campo-string));
fuera de eso, ninguna otra forma se rechaza. Se guardan, se imprimen y viajan en la exportación.

La prueba, que es por descarte: si no cambia qué hay que hacer para dar por terminada una tarea
mayor, ni en qué orden hay que hacer las cosas, es un puntero.

## Cómo se elige entre los tres

No son alternativas, y lo normal es que se usen a la vez sobre las mismas tareas: una tarea grande
con sus hijas, las hijas encadenadas entre sí por dependencias, y cada una apuntando al documento
que la gobierna.

**Cuidado con la palabra "depende", que en la lengua corriente sirve igual para las dos primeras.**
"Esto depende de aquello" puede querer decir que no puede empezar hasta que la otra termine, que es
una dependencia, o que es parte de ella, que es un parentesco. Cuando la frase que llega es ambigua,
quien la recibe la desambigua con las dos preguntas de arriba, no mirando lo que la tarea ya tenga
escrito.

Y hay dos cosas que no hace ninguna de las tres relaciones. **Ninguna cambia el identificador de una
tarea**, que no codifica jamás su lugar en ninguna de ellas
(["Identificador de tarea"](identificadores.md#identificador-de-tarea)). Y **ninguna cambia el orden
de un listado**, que no agrupa nunca, ni por estado ni por parentesco
(["La regla de orden, completa"](../cmd/ls.md#la-regla-de-orden-completa)).
