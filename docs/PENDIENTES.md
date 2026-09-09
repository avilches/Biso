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

---

## 1. Huecos que un implementador tendría que rellenar solo

Son los que más valen, porque cada uno es un sitio donde dos personas escribirían programas distintos.

**`--board` sigue descrito como una cadena opaca.** La sección 3 dice que acepta "el nombre o el
localizador de un tablero, en la forma que el almacenamiento imponga". Eso era razonable mientras la
persistencia estuviera sin decidir, y ahora deja sin decir si admite el nombre, el identificador de ocho
hexadecimales o una ruta, sobre qué raíces busca, y qué pasa cuando el nombre encaja con dos tableros de
la máquina, que es un caso que la sección 3.2 declara legal. Tampoco hay ningún `code` para "el tablero
que has nombrado no existe": `no_board` dice otra cosa. *(De la persistencia.)*

**La instantánea que "cruza a otra máquina" no tiene camino para cruzar.** La sección 14 dice que lo que
viaja es la instantánea que `biso snapshot` deja en git, pero ese repositorio vive dentro del directorio
del tablero, que está fuera del proyecto y no tiene remoto, y ningún comando hace push ni añade uno. El
historial es estrictamente local. Hay que decir que publicar ese directorio es trabajo de quien lo use, o
dar la vía. *(De la persistencia.)*

**Dos `biso snapshot` simultáneos no están cubiertos.** Es el único comando que escribe ficheros de texto
con nombre fijo en una ubicación compartida, y no aparece en ninguna de las seis garantías de la sección
4.10, que hablan de escrituras del tablero. Dos llamadas a la vez pueden entrelazar la escritura de
`tasks.ndjson` y sus dos `git commit` chocan en `index.lock`. Hace falta exigir escritura por fichero
temporal y renombrado atómico, y decir qué código sale cuando el commit choca. *(De la persistencia.)*

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
sugerir. *(De la persistencia.)*

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
puntos de código, grafemas o celdas de terminal, que es lo único que alinea de verdad. *(Anterior.)*

**`--print` y `--dry-run` no tienen comportamiento definido en `biso init` ni en `biso config set`.** No
están en la lista de comandos de lectura, así que se aceptan, pero los dos comandos escriben sin afectar a
ninguna tarea y `--print` está definido como "la ficha completa de cada tarea afectada". *(Anterior.)*

**`biso board` y `biso help` aceptan `--json` y no tienen sobre.** La sección 3 dice que las banderas
globales valen para todos los comandos y que las restricciones adicionales "son exactamente cuatro en todo
el documento", ninguna de las cuales los cubre, pero la tabla de `kind` de la sección 12.1 no tiene fila
para ellos y no hay esquema en ninguna parte. *(Anterior, heredado al declarar el "exactamente cuatro".)*

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

## 3. Dos cosas que el documento presenta mejor de lo que son

**Los 8,7 milisegundos se midieron leyendo un JSON, no una base de datos SQLite.** La sección 12 de
`docs/ESTADO-DEL-ARTE.md` lo dice literalmente; la sección 13 de `docs/DECISIONES.md` recoge la cifra sin
la mitad que dice de dónde lee, y sobre ella construye las "tres veces de margen" que la sección 14 usa
dos veces para apoyar la elección de Go. La cifra no está inventada, pero mide otra carga de trabajo. Una
frase que lo diga refuerza el párrafo final de la sección 14, que ya admite que el controlador de SQLite
está sin medir. *(De la persistencia.)*

**El token de vallado que se cita no es el que se implementa.** La sección 9.2 de `DECISIONES.md` presenta
la comprobación del tenedor como el token de vallado del artículo citado, que "rechaza las escrituras del
propietario antiguo". Lo que `biso` hace es aceptar la escritura del tenedor viejo y solo negarse a tocar
los dos campos del arrendamiento, con un aviso. Es coherente con el "avisa, no impide" que gobierna todo
el resto, pero entonces hay que decirlo así en vez de decir que cierra el agujero, y la lista de riesgos
aceptados de la sección 11 debería llevar esta fila. *(De la persistencia.)*

---

## 4. Incoherencias que no cambian el comportamiento

Ninguna hace fallar una implementación, pero cada una es una frase que dice algo falso.

- **`leaseExpired` sale en dos de los cuatro bloques del JSON de `biso prime`** y no en los otros dos, con
  su razón escrita (ahí valdría siempre `false`). La sección 12.4 promete que todas las claves
  documentadas aparecen siempre "para que nadie tenga que distinguir entre no está y no tiene valor", que
  es exactamente la distinción que esto obliga a hacer. *(De la persistencia.)*
- **La ficha de `biso get` no tiene fila para `project` ni para `ordinal`**, que son campos escalares de la
  misma clase que `milestone` y `parent`, que sí la tienen. *(Anterior.)*
- **`INTEGRATION.md` lleva las cifras del mensaje de arranque de antes del arrendamiento**: dice 1.441 de
  resumen y 4.696 de total, cuando las secciones que cita dicen 1.491 y 4.746. Y sigue describiendo el
  arrendamiento como "aplazado en la sección 9.2", que dejó de ser cierto. *(De la persistencia.)*
- **Un reparto de 3.456 y 1.664 no son "dos mitades exactas"**, como dicen cuatro sitios. Es un 67,5 y un
  32,5 por ciento. La palabra correcta es partes. *(Anterior.)*
- **La sección 3.2 llama a la tercera vía "la más específica de las cuatro"** aunque pierda contra
  `--board` y `BISO_BOARD`. El razonamiento que sigue solo necesita que sea más específica que el puntero.
  *(De la persistencia.)*
- **`biso ask --help` dice que la tarea aparece bajo `WAITING ON A PERSON`** y el bloque se llama
  `NEEDS ANSWER` en las diecinueve veces que sale en el documento. *(Anterior.)*
- **`docs/ESTADO-DEL-ARTE.md` cita `biso ls --ready`** como una de las tres cosas que nadie quiere perder,
  y esa bandera se retiró en favor de `--blocked` y `--not-blocked`. *(Anterior.)*
- **La fila `tasks` de `biso where` usa "active" con el sentido de "no archivada"**, cuando la tabla de
  vocabulario reserva esa palabra para el papel del estado. *(Anterior.)*
- **La sección 10.4 dice que "`--sort` sin valor aplica el orden por defecto"** y debería decir "sin
  `--sort`", porque `--sort` sin valor es `missing_value`, código 2. *(Anterior.)*
- **La firma de `biso export` escribe `-o <file|->` y su tabla titula la fila `--out <file>`.** El resto
  del documento pone las dos formas juntas en la firma. *(Anterior.)*
- **`urgencyBreakdown` aparece en el ejemplo JSON de `biso get` sin decir que necesita
  `--explain-urgency`**, mientras que para el texto sí se dice. *(Anterior.)*
- **La línea de estado de `biso finish` no es alcanzable desde el ciclo que el documento enseña**: muestra
  `dod 1/1`, y la invocación del ciclo no lleva `--check-dod`, así que diría `dod 0/1`. Los ejemplos de
  salida se generan, así que hay que arreglar uno de los dos. *(Anterior.)*
- **El informe de `biso doctor` no dice si un error ya reparado cuenta en el recuento de la primera
  línea.** El ejemplo dice "1 error found" y "1 problem fixed" habiendo encontrado dos. *(Anterior.)*
- **`--id` y `--match` están en la tabla de parámetros de `get` y de `set` y no en la de los otros siete
  comandos que las aceptan**, aunque su firma y su ayuda las lleven. *(Anterior.)*
- **La afirmación de que hay "una sola desviación de nombre en todo el programa" no se sostiene**: marcar
  un criterio es `--check` y marcar un elemento de la definición de hecho es `--check-dod`, así que una
  lleva el sufijo del campo y la otra no. Y la rejilla `FIELD FLAGS` del mensaje de arranque, de la que se
  dice que lleva "los nombres de todas las banderas de campo", no incluye `--check-dod` ni
  `--uncheck-dod`, que son justo los que no se pueden adivinar. *(Anterior.)*
- **`--milestone` es el único filtro de texto libre que no valida.** La sección 6 de `DECISIONES.md`
  explica por qué las etiquetas y las personas no tienen vocabulario cerrado al escribir pero sus filtros
  sí validan, y ese argumento no distingue en nada al hito. Hoy un hito mal escrito devuelve una lista
  vacía, que es el fallo que el principio 1 existe para evitar. *(Anterior, y es la única grieta que le
  queda al principio que el proyecto pone primero.)*
- **El argumento contra combinar `--overwrite-config` con `--from`** dice que fusionar dos almacenes "es
  justo lo que esta decisión de persistencia rechaza en todas partes", y no es en todas partes:
  `biso new --from` sobre un tablero con tareas está permitido y es, por cualquier lectura llana, fusionar.
  Lo que la decisión rechaza son dos copias vivas escritas por separado. *(De la persistencia.)*

---

## 5. Decisiones aplazadas a propósito

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

## 6. Dos cosas que necesitan una decisión de quien manda

**Las plantillas del ledger de `INTEGRATION.md`.** Sus cuatro em-dash se sustituyeron por guiones normales
para cumplir la regla del repositorio, pero ese fichero presenta el formato como fijado por el plugin de
Superpowers, y el ledger real que existe en la máquina sí los lleva. Nada consume ese formato, así que no
se rompe nada, pero el documento ya no cita el formato ajeno carácter a carácter.

**El identificador `lease_invariant`**, que nombra la comprobación nueva de `biso doctor`, se inventó al
escribirla. Los códigos de los hallazgos de `doctor` no están enumerados en la sección 12.3, así que no
hubo lista que actualizar, pero el nombre no viene de ninguna parte y se puede cambiar.
