# Biso

`biso` es una herramienta de línea de comandos para llevar las tareas de un proyecto, pensada para
que la use un agente automático que trabaja dentro de ese proyecto. La salida es predecible, los
errores se distinguen por su código sin leer el mensaje, y ningún comportamiento depende de dónde se
ejecute el programa.

## Estado: especificación cerrada, sin una línea de código

Lo que hay es [`docs/SPEC.md`](docs/SPEC.md), y es el documento del que se implementa todo. Define
todos los comandos (su recuento vive en la sección 10) con su firma, su tabla de parámetros, su
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

Su sección 15 dice por dónde empezar, en el orden en que cada pieza paga lo que cuesta: el modelo de
datos, el algoritmo de coincidencia (una función pura de la que dependen todos los comandos), los
cuatro comandos del trabajo diario, los verbos del ciclo, el mensaje de arranque, el lote y la
exportación, y el resto.

## El modelo de estados está cerrado

Los cuatro requisitos que lo tocaban ya están decididos: dos resueltos sin ningún papel de estado
nuevo, uno resuelto con la decisión de persistencia ya tomada (el arrendamiento con caducidad) y uno
retirado. El detalle, con la evidencia detrás de cada decisión, está en la sección 9 de
[`docs/DECISIONES.md`](docs/DECISIONES.md).

## La persistencia está decidida

Un tablero es una base de datos SQLite en un directorio propio fuera del proyecto, localizado por un
fichero puntero versionado en git, con una exportación de texto que sí se commitea para el historial
(`biso snapshot`). Sin daemon, y sin fusionar nunca dos almacenes escritos por separado. El porqué de
cada pieza, con su aritmética, está en la sección 12 de [`docs/DECISIONES.md`](docs/DECISIONES.md),
que enlaza a su vez a [`docs/ESTADO-DEL-ARTE.md`](docs/ESTADO-DEL-ARTE.md), la investigación sobre las
demás herramientas del espacio y por qué fallan.

## El lenguaje es Go

El requisito que sale de la especificación es que el programa arranque rápido, con el presupuesto de la
sección 4.13 de `docs/SPEC.md`: 25 milisegundos de reloj para `biso ls` y `biso prime` sobre un tablero
de 300 tareas, cuyo origen está en la sección 13 de `docs/DECISIONES.md`. Eso descarta los lenguajes
interpretados, y entre los compilados la elección es **Go**, no por rendimiento sino por lo que cuesta
escribir el programa: la ventaja de Rust son seis décimas de milisegundo sobre un presupuesto que ya
sobra tres veces, y su modelo de propiedad de memoria obliga a rondas de corrección que alargan el ciclo
sin dejar nada mejor en el producto. El razonamiento completo, con las cifras, está en la sección 14 de
`docs/DECISIONES.md`.

**Y el controlador de SQLite es `modernc.org/sqlite`, sobre la interfaz estándar `database/sql` y sin
`cgo`.** Medido el 2026-09-10: los cuatro candidatos cumplen el presupuesto con mucho margen, así que no
decide el reloj sino la distribución del binario, y el enlace con la biblioteca en C no puede compilar de
forma cruzada para Linux ni para Windows y su modo sin `cgo` produce un binario que falla al ejecutarse
en vez de al construirse. El razonamiento está en el apartado 14.1 de `docs/DECISIONES.md`, y el banco de
pruebas del que salen las cifras en `bench/sqlite-driver/`, con su propio `README.md`.

Con eso **ya no queda nada que bloquee empezar a escribir código**: el paso 1 del orden de la sección 15
de `docs/SPEC.md` es el almacén, y su decisión de controlador está tomada.

## Cosas que conviene tener presentes al implementar

- **El mensaje de arranque tiene un tope duro de 5.120 bytes.** No es un objetivo, es una prueba de la
  suite, y es el único de los números de tamaño que congela el contrato de estabilidad de la sección
  13 de `docs/SPEC.md`. El reparto entre la parte fija y el resumen, y lo que mide hoy el texto, están
  en la sección 9.5 de `docs/SPEC.md`.
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

Los documentos de `docs/` se sirven como un sitio navegable con MkDocs y el tema Material,
definido en `mkdocs.yml`. Es utillaje de documentación, no parte del programa: no toca el código
Go. Las dependencias de Python están fijadas con versión exacta en `docs-requirements.txt`.

Se ejecuta con `uv` (ya instalado en esta máquina) sin crear un entorno virtual dentro del
repositorio ni instalar nada en el Python del sistema: `uv` resuelve las dependencias fijadas a un
caché propio y las descarta al terminar.

- Servir en local con recarga automática al editar los `.md`:
  `uv run --with-requirements docs-requirements.txt --no-project mkdocs serve`
- Construir el sitio estático en `site/` (no se versiona, ver `.gitignore`):
  `uv run --with-requirements docs-requirements.txt --no-project mkdocs build --strict`

`docs/TUTORIAL.md` es la única excepción a "los `.md` son la fuente de verdad": es producto
generado a partir de los fixtures de `tutorial/escenarios/` y de `tutorial/conceptos.md`, y no se
edita a mano (lleva su propia cabecera que lo recuerda). Se regenera con:
`uv run --with-requirements docs-requirements.txt --no-project python tutorial/generar.py`

## El tutorial y sus tres herramientas

El tutorial se diseñó en
[`docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md`](docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md),
y va por situaciones y no por comandos, para no ser una segunda copia de la sección 10 de la
especificación. Cada paso de cada escenario declara su comando, su salida, su código de salida y de
dónde sale esa salida: `literal SPEC <sección>` si está copiada carácter a carácter, o
`derivada SPEC <sección>` si se construyó aplicando sus reglas. El generador marca las derivadas en
la página, y esa marca es el aparato de validación: lo literal ya está validado por estar en la
especificación, lo derivado es lo que hay que revisar.

Tres scripts, los tres con el mismo prefijo de `uv` que MkDocs:

- `python tutorial/generar.py` escribe `docs/TUTORIAL.md`. Valida los fixtures y **falla** si a un
  paso le falta el código de salida o si `origen` no tiene una de las dos formas admitidas.
- `python tutorial/urgencia.py` calcula la urgencia de cada tarea del tablero de ejemplo según la
  sección 5.4, con el desglose de cada término, y las ordena por la regla de `biso ls`. Existe para
  que ningún fixture tenga que hacer esa aritmética a mano: los escenarios ordenan listas por
  urgencia y ese orden hay que poder reproducirlo.
- `python tutorial/continuidad.py` comprueba que los contadores del tablero encadenan entre
  escenarios consecutivos, o sea que lo que uno entrega es lo que el siguiente recibe. Sale 1 si no
  cuadran. Ya cazó un desfase de uno que cuatro escenarios arrastraban.

`tutorial/lagunas/` es lo que salió de escribir el tutorial: lo que la especificación no decide y
hubo que suponer, con la pregunta concreta y qué se supuso. Alimenta `docs/PENDIENTES.md`.

El día que exista el binario, el mismo formato de fixture se ejecuta contra un tablero sembrado con
`tutorial/tablero.yaml` y se compara con lo declarado, así que los ejemplos dejarán de ser simulados
y pasarán a ser pruebas de salida literal. Por eso `cmd` es una línea ejecutable y no una
ilustración, y por eso el código de salida es obligatorio en todos los pasos.
