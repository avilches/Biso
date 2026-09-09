# Lo que queda por cerrar en la especificación

Este documento existe porque `docs/SPEC.md` está escrito para implementarse sin preguntar nada, y estas
son las preguntas que todavía se le pueden hacer. Ninguna impide empezar a implementar: la especificación
es coherente y completa en todo lo que el trabajo diario toca. Lo que sigue son huecos que un
implementador encontraría al llegar a un rincón concreto, incoherencias que no cambian el comportamiento,
y decisiones aplazadas a propósito.

Cada entrada dice de dónde viene, porque cambia quién la arregla y con qué urgencia. **De la persistencia**
es de la rama que decidió el almacén. **Anterior** significa que ya estaba antes y sobrevivió a varias
revisiones, así que probablemente no duela tanto como parece.

La mayoría salió de dos revisiones completas de la rama de persistencia, hechas el 2026-09-08: una
mecánica, que verificó recuentos, referencias cruzadas, ejemplos y códigos de salida, y una de fondo, que
buscó decisiones que no se sostuvieran contra un escenario concreto.

**Cuando una entrada se cierra, se quita de aquí.** El 2026-09-09 se cerraron las dos de la sección que
señalaba lo que el documento vendía mejor de lo que era, diez de las incoherencias de la sección 3, y las
dos decisiones que esperaban al usuario, y ese mismo día seis más de esa sección, que necesitaban elegir
entre dos arreglos válidos. Y una se cerró por una vía que conviene recordar, porque no es
la habitual: el hueco de `--board` no se rellenó, se disolvió al retirar la bandera, cuando se vio que
`-C` ya llegaba a todo lo que ella prometía. La sección 12 de `DECISIONES.md` guarda el porqué. Merece la
pena preguntarse lo mismo ante cada entrada que queda: si el hueco existe porque falta decidir algo, o
porque sobra la cosa que lo abre. Una advertencia que se pagó caro entonces y conviene no repetir:
al revisar este documento se descubrió que **sus propios recuentos estaban desincronizados tres veces**,
que es exactamente el fallo que él denuncia como dominante en la especificación. No escribas aquí una
frase que cuente elementos sin contarlos.

---

## 1. Huecos que un implementador tendría que rellenar solo

Son los que más valen, porque cada uno es un sitio donde dos personas escribirían programas distintos.

**La instantánea que "cruza a otra máquina" no tiene camino para cruzar.** La sección 14 dice que lo que
viaja es la instantánea que `biso snapshot` deja en git, pero ese repositorio vive dentro del directorio
del tablero, que está fuera del proyecto y no tiene remoto, y ningún comando hace push ni añade uno. El
historial es estrictamente local. Hay que decir que publicar ese directorio es trabajo de quien lo use, o
dar la vía. **Tiene una consecuencia que se ve desde `INTEGRATION.md`**: las decisiones que un agente toma
en nombre de su humano, que son el argumento más fuerte de ese documento a favor de guardarlas como
comentarios de `biso`, pasan de morir con la sesión a morir con la máquina, y la frase de Superpowers de
que "la historia de git es el registro ahora" solo vuelve a ser cierta cuando alguien ejecuta `snapshot` y
commitea. *(De la persistencia.)*

**Dos `biso snapshot` simultáneos no están cubiertos.** Es el único comando que escribe ficheros de texto
con nombre fijo en una ubicación compartida, y no aparece en ninguna de las seis garantías de la sección
4.10, que hablan de escrituras del tablero. Dos llamadas a la vez pueden entrelazar la escritura de
`tasks.ndjson` y sus dos `git commit` chocan en `index.lock`. Hace falta exigir escritura por fichero
temporal y renombrado atómico, y decir qué código sale cuando el commit choca. **Y hay una mitad más
pequeña del mismo hueco**: `snapshot` escribe dos ficheros y la especificación no dice qué queda en disco
si el primero se escribe y el segundo no, porque ninguna de las seis garantías habla de ficheros.
*(De la persistencia.)*

**El código de salida 8 significa tres cosas con tres remedios distintos**: no hay tablero (ejecuta
`init`), el puntero nombra uno ausente (ejecuta `init`, adopta el identificador) y la base de datos está
dañada (restaura una copia, y `init` no arregla nada). El principio de la sección 1 dice que los errores
se distinguen por su código sin leer el mensaje, y aquí solo se distinguen por la clave `code`, que además
solo aparece con `--json`. El tercer caso lo añadió la persistencia y es el que no comparte remedio.
*(De la persistencia.)*

**Con `board.db` dañado, el remedio que el usuario intentará no está definido.** La sección 4.12 aborta
con código 8 recomendando restaurar de una copia, y la sección 10.1 dice que si ya hay un tablero
accesible desde aquí, `init` da error 2. No está dicho cuál gana cuando alguien ejecuta
`biso init --from <instantánea>` en ese directorio, que es exactamente lo que el mensaje le acaba de
sugerir. **Y el mensaje empeora el enredo**: dice "restaura de una copia" sin nombrar `biso init --from`,
que es el único mecanismo de restauración que la especificación define. *(De la persistencia.)*

**Restaurar la instantánea de otra persona fija su identidad como la del tablero.** El `config.json` que
escribe `biso snapshot` lleva `me` y `default_limit`, así que un `init --from` sobre la instantánea de un
compañero deja su identidad en la configuración, que es lo que la sección 11 de `DECISIONES.md` avisa que
destruye `--mine`. La otra mitad de este hueco, que el identificador del tablero no viajaba en la
instantánea, ya está resuelta: el marcador `<id>.id` queda versionado. *(De la persistencia.)*

**El recorte del título a 100 caracteres admite dos implementaciones.** No está dicho si la cadena
resultante mide 100 caracteres, con 97 más los tres puntos, o 103. La misma ambigüedad la hereda el cuerpo
de la pregunta en la sección 9.7, que remite a "la misma cifra exacta". Ningún ejemplo del documento
ejerce el recorte, porque el título más largo mide 48 caracteres, así que es la única parte del algoritmo
de columnas que no queda comprobada carácter a carácter. *(Anterior.)*

**Nada dice cómo se alinean las columnas con texto que no es de anchura sencilla.** La sección 4.4 obliga a
UTF-8 y los títulos son texto libre, así que una tarea puede llevar ideogramas, emoji o marcas
combinantes. El algoritmo define el ancho como "la longitud del valor más largo" sin decir si se cuentan
puntos de código, grafemas o celdas de terminal, que es lo único que alinea de verdad. Ninguna de esas
palabras aparece en ningún documento del proyecto, así que no hay ni una pista de intención. Ojo al
decidir: la sección 4.1 prohíbe expresamente mirar el terminal para decidir qué se imprime, lo que cierra
la puerta a adaptarse a su ancho pero no resuelve en qué unidad se cuenta. *(Anterior.)*

**`--print` y `--dry-run` no tienen comportamiento definido en `biso init` ni en `biso config set`.** No
están en la lista de comandos de lectura, así que se aceptan, pero los dos comandos escriben sin afectar a
ninguna tarea y `--print` está definido como "la ficha completa de cada tarea afectada". *(Anterior.)*

**`biso board` y `biso help` aceptan `--json` y no tienen sobre.** La sección 3 dice que las banderas
globales valen para todos los comandos y que las restricciones adicionales "son exactamente cuatro en todo
el documento", ninguna de las cuales los cubre, pero la tabla de `kind` de la sección 12.1 no tiene fila
para ellos y no hay esquema en ninguna parte. *(Anterior.)*

---

## 2. El orden de implementación de la sección 15 está desfasado

Tres cosas a la vez, y las tres se arreglan juntas:

- **No hay ningún paso para el almacén**: abrir o crear la base de datos, el modo WAL y la transacción de
  la que dependen las seis garantías de la sección 4.10. La nota final dice que esas garantías no son un
  paso de la lista, lo cual es cierto como principio, pero el almacén sí es trabajo.
- **`init` va en el paso 7 y los comandos que necesitan un tablero en el paso 3.** No hay forma de
  ejecutar el paso 3 sin un tablero, y crear un tablero es `init`.
- **`biso snapshot` no aparece en ningún paso**, aunque la rama lo añadiera como vigésimo comando y
  actualizara el recuento, la tabla de `kind` y el bloque de `biso help all`. Y la prueba de simetría que
  el paso 6 exige usa `snapshot` y `init --from`, los dos del paso 7.

Al arreglarlo conviene que el paso nuevo del almacén diga que ahí se mide el presupuesto de la sección
4.13 con el mecanismo elegido, **sin nombrar ningún lenguaje**, para que `SPEC.md` siga siendo agnóstica.
Eso cierra el hueco que hoy tiene: la prueba del presupuesto no se puede escribir hasta que el almacén
exista y el controlador de SQLite esté elegido, y esa elección está anotada en `CLAUDE.md` como lo primero
que la implementación tiene que resolver. *(De la persistencia.)*

---

## 3. Incoherencias que no cambian el comportamiento

Ninguna hace fallar una implementación, pero cada una es una frase que dice algo falso. Queda una, y
sigue abierta porque la clave que discute está congelada por el contrato de estabilidad, así que el
texto y el JSON no se pueden arreglar de la misma forma.

- **La fila `tasks` de `biso where` usa "active" con el sentido de "no archivada"**, cuando la tabla de
  vocabulario reserva esa palabra para el papel del estado. Cuidado: la clave JSON `active` está congelada
  por el contrato de estabilidad de la sección 13, así que el texto y el JSON no se pueden arreglar igual.
  *(Anterior.)*

---

## 4. Decisiones aplazadas a propósito

**El controlador de SQLite.** Es lo primero que la implementación tiene que resolver, y está en el
`CLAUDE.md`: el enlace con la biblioteca en C obliga a compilar con `cgo` y complica generar binarios para
otras plataformas, la traducción a Go puro compila en cualquier sitio, y la diferencia afecta al arranque.
Sin medir contra el presupuesto de la sección 4.13.

**La exportación al formato de Backlog.md.** No es para uso propio: es para que quien ya lo usa pueda
probar `biso` sin salto al vacío. La advertencia es que ese formato tiene bugs abiertos que se heredarían,
entre ellos que al editar una tarea pierde las claves de frontmatter que no conoce. La introducción de
`SPEC.md` ya declara que la compatibilidad con Backlog.md es de modelo de datos y no de formato de fichero.

**La interfaz multiproyecto**, fuera de alcance a propósito y con su propio brainstorming pendiente. De
ella solo se recogió el requisito que afecta al almacenamiento: los tableros se enumeran por convención,
con una raíz por defecto más una lista explícita de raíces adicionales, sin ningún registro que se
actualice solo.

**El mensaje del commit `bd0c874`** describe cuatro cambios que en realidad entraron en `aaa5b96`. El
árbol es correcto, solo el mensaje se adelantó. Sin enmendar por no reescribir historia sin petición.

---

## 5. Huecos que aparecieron al cerrar los demás

Salieron el 2026-09-09, mientras se arreglaban las incoherencias de la sección 3, y ninguno estaba
apuntado antes. Se listan aparte porque todavía no se ha decidido si entran en esta ronda.

**El error de identidad duplicada de tablero no existe.** Dos sitios distintos de la especificación
remiten a "el error de que el mismo `id` aparezca en dos raíces", uno en la sección 3.3 y otro en la 10.1
al hablar de `init --from`, y ese error no está escrito en ninguna parte: no tiene mensaje literal, ni
código de salida, ni identificador de la clave `code`. `biso doctor` tampoco lo comprueba, y su tabla de
dieciocho filas está cerrada y contada, así que añadir la comprobación obliga a tocar tres frases que la
cuentan.

**No hay mensaje literal para "ya hay un tablero accesible desde aquí"**, que es con diferencia el error
más probable de `biso init`. Tiene código 2 y el `code` `board_exists`, pero ningún texto, mientras los
otros dos errores de resolución de tablero sí lo tienen.

**El ejemplo de texto de `biso config list` omite cuatro claves que su propio esquema JSON incluye**
(`projects`, `labels`, `assignees` y los siete coeficientes de `urgency`). Como el `config.json` que
escribe `biso snapshot` se define por referencia a ese esquema, la ambigüedad se propaga al fichero que
`init --from` lee de vuelta, y la prueba de simetría promete "toda su configuración" ejemplificando solo
cuatro cosas.

**El sobre de error tiene cuatro claves que varían según los datos.** Al reescribir la promesa de la
sección 12.4 quedó dicho con todas las letras que ninguna clave va ni viene según los datos, y que la
única excepción son las que gobierna una bandera, porque quien llama sabe qué ha escrito. El sobre de
error de la sección 12.2 incumple eso: `field`, `given`, `valid` y `details` aparecen o no según qué
error sea. Es la única tensión que le queda a la promesa recién escrita, y no se tocó porque quedaba
fuera del encargo. Puede que la salida sea documentarlas por `code`, o admitir que un sobre de error no
es una salida de datos y gobernarlo aparte.

**`ESTADO-DEL-ARTE.md` afirma que `biso` distingue por diseño quién escribió qué**, y la lista de riesgos
aceptados de la sección 11 de `DECISIONES.md` reconoce que un tablero con la clave `me` configurada anula
esa distinción, porque entonces todo el mundo comparte identidad. Las dos frases no pueden ser ciertas a
la vez.
