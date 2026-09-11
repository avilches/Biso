# Biso

`biso` es una herramienta de línea de comandos para llevar las tareas de un proyecto, pensada para
que la use un agente automático que trabaja dentro de ese proyecto. La salida es predecible, los
errores se distinguen por su código sin leer el mensaje, y ningún comportamiento depende de dónde se
ejecute el programa.

## Estado: especificación cerrada, sin una línea de código

Lo que hay es [`docs/spec/`](docs/spec/index.md), y es la especificación de la que se implementa todo. Define
todos los comandos (su recuento vive en ["Los comandos"](docs/spec/cmd/index.md)) con su firma, su tabla de parámetros, su
comportamiento en los casos límite, la salida literal que imprimen, su esquema JSON, sus códigos de
salida y el texto exacto de su ayuda.
Está escrito para que alguien lo implemente entero sin preguntar nada.

**No hay que rediseñar nada por libre.** Si al implementar aparece un caso que la especificación no
cubre, lo correcto es añadirlo a la especificación y luego implementarlo, no resolverlo solo en el
código. Y si una regla parece arbitraria, su razón está en
[`docs/DECISIONES.md`](docs/DECISIONES.md) antes de cambiarla.

**Y antes de dar por hueca una laguna, hay que mirar
[`docs/PENDIENTES.md`](docs/PENDIENTES.md)**, que es la lista de lo que queda por cerrar: los rincones
donde la especificación todavía no decide, las frases que dicen algo falso sin cambiar el
comportamiento, y las decisiones aplazadas a propósito. Nada de eso impide empezar a implementar, y
saber que ya está anotado evita volver a descubrirlo. Cuando se cierre una entrada, se quita de ahí.

["Por dónde empezar a implementar"](docs/spec/por-donde-empezar.md) dice el orden en que cada pieza paga lo que cuesta: el modelo de
datos, el algoritmo de coincidencia (una función pura de la que dependen todos los comandos), los
cuatro comandos del trabajo diario, los verbos del ciclo, el mensaje de arranque, el lote y la
exportación, y el resto.

## El modelo de estados está cerrado

Los cuatro requisitos que lo tocaban ya están decididos: dos resueltos sin ningún papel de estado
nuevo, uno resuelto con la decisión de persistencia ya tomada (el arrendamiento con caducidad) y uno
retirado. El detalle, con la evidencia detrás de cada decisión, está en
["El modelo de estados: cuatro requisitos, cerrados"](docs/DECISIONES.md#el-modelo-de-estados-cuatro-requisitos-cerrados).

## La persistencia está decidida

Un tablero es una base de datos SQLite en un directorio propio fuera del proyecto, localizado por un
fichero puntero versionado en git, con una exportación de texto que sí se commitea para el historial
(`biso snapshot`). Sin daemon, y sin fusionar nunca dos almacenes escritos por separado. El porqué de
cada pieza, con su aritmética, está en ["La decisión de persistencia"](docs/DECISIONES.md#la-decisión-de-persistencia),
que enlaza a su vez a [`docs/ESTADO-DEL-ARTE.md`](docs/ESTADO-DEL-ARTE.md), la investigación sobre las
demás herramientas del espacio y por qué fallan.

## El lenguaje es Go

El requisito que sale de la especificación es que el programa arranque rápido, con el presupuesto de
["El presupuesto de arranque"](docs/spec/presupuestos.md#el-presupuesto-de-arranque): 25 milisegundos de reloj para `biso ls` y `biso prime` sobre un tablero
de 300 tareas, cuyo origen está en ["El origen de la cifra de 25 milisegundos"](docs/DECISIONES.md#el-origen-de-la-cifra-de-25-milisegundos). Eso descarta los lenguajes
interpretados, y entre los compilados la elección es **Go**, no por rendimiento sino por lo que cuesta
escribir el programa: la ventaja de Rust son seis décimas de milisegundo sobre un presupuesto que ya
sobra tres veces, y su modelo de propiedad de memoria obliga a rondas de corrección que alargan el ciclo
sin dejar nada mejor en el producto. El razonamiento completo, con las cifras, está en
["El lenguaje de implementación es Go"](docs/DECISIONES.md#el-lenguaje-de-implementación-es-go).

**Y el controlador de SQLite es `modernc.org/sqlite`, sobre la interfaz estándar `database/sql` y sin
`cgo`.** Medido el 2026-09-10: los cuatro candidatos cumplen el presupuesto con mucho margen, así que no
decide el reloj sino la distribución del binario, y el enlace con la biblioteca en C no puede compilar de
forma cruzada para Linux ni para Windows y su modo sin `cgo` produce un binario que falla al ejecutarse
en vez de al construirse. El razonamiento está en
["El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`"](docs/DECISIONES.md#el-controlador-de-sqlite-es-moderncorgsqlite-sin-cgo), y el banco de
pruebas del que salen las cifras en `bench/sqlite-driver/`, con su propio `README.md`.

Con eso **ya no queda nada que bloquee empezar a escribir código**: el paso 1 del orden de
["Por dónde empezar a implementar"](docs/spec/por-donde-empezar.md) es el almacén, y su decisión de controlador está tomada.

## Cosas que conviene tener presentes al implementar

- **El mensaje de arranque tiene un tope duro de 5.120 bytes.** No es un objetivo, es una prueba de la
  suite, y es el único de los números de tamaño que congela
  ["El contrato de estabilidad"](docs/spec/estabilidad.md). El reparto entre la parte fija y el resumen, y lo que mide hoy el texto, están
  en ["El presupuesto de tamaño"](docs/spec/presupuestos.md#el-presupuesto-de-tamaño).
- **La simetría entre `biso export` y `biso new --from` es una prueba, no una intención.** Exportar
  un tablero e importarlo en otro vacío tiene que dar dos tableros idénticos campo a campo, con
  identificadores, fechas y claves de criterios incluidas.
- **Los ejemplos de salida del documento se generan, no se escriben a mano.** El algoritmo de
  columnas de `biso ls` y de `biso prime` está especificado como algoritmo justamente para eso.
  Cualquier cambio en él obliga a regenerar los ejemplos y a comprobar que coinciden carácter a
  carácter.
- **Un valor que no existe es siempre un error, se esté escribiendo o leyendo.** Un filtro mal
  escrito nunca puede devolver una lista vacía, porque quien la lee la interpreta como un hecho sobre
  el tablero.

## Reglas de este repositorio

- **El trabajo va en un worktree**, en `.claude/worktrees/<rama>`, nunca editando `main`
  directamente.
- **La documentación y los comentarios van en español.** Los identificadores del código y todo lo que
  es interfaz del programa (comandos, banderas, textos de ayuda, mensajes de error, claves JSON) van
  en inglés.
- **Nunca em-dash**, en ningún texto: ni en documentación, ni en código, ni en mensajes de commit.
- **Los mensajes de commit y las descripciones de PR no llevan coautoría** ni mención de haber sido
  generados por un agente.

## El sitio de documentación

Los cuatro documentos de `docs/` se sirven como un sitio navegable con MkDocs y el tema Material,
definido en `mkdocs.yml`. Es utillaje de documentación, no parte del programa: no toca el código
Go ni cambia el contenido de los `.md`, que siguen siendo la fuente de verdad. Las dependencias de
Python están fijadas con versión exacta en `docs-requirements.txt`.

Se ejecuta con `uv` (ya instalado en esta máquina) sin crear un entorno virtual dentro del
repositorio ni instalar nada en el Python del sistema: `uv` resuelve las dependencias fijadas a un
caché propio y las descarta al terminar.

- Servir en local con recarga automática al editar los `.md`:
  `uv run --with-requirements docs-requirements.txt --no-project mkdocs serve`
- Construir el sitio estático en `site/` (no se versiona, ver `.gitignore`):
  `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`
