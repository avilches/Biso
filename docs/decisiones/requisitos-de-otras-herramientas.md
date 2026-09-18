# Cuatro requisitos aprendidos de otras herramientas

Estos no salen del estudio de uso, sino de comparar Backlog.md con Taskwarrior y mirar qué falla en
cada uno.

**Ningún campo desconocido se acepta ni se descarta en silencio.** Las dos herramientas lo hacen mal:
Backlog.md borra el campo sin avisar en la siguiente reescritura del fichero, y Taskwarrior, ante un
atributo personalizado no declarado, en el peor caso se come la descripción entera de la tarea. La
regla correcta es rechazar con un error explícito, y de ahí salen las claves declaradas de `ext` y la
regla de que una tarea con una clave no declarada falla en una lectura dirigida en vez de perderse.

**Cualquier fecha se puede fijar al importar.** Ninguna de las dos lo permite desde su interfaz
pública. Backlog.md no tiene un solo parámetro de fecha, ni para la de creación ni para el instante de
un comentario, y por eso un importador de histórico no puede preservar las fechas reales usando su
línea de comandos, que es su única vía legítima de escritura. Un detalle que conviene recordar: dentro
de Backlog.md existe una función interna que sí acepta una fecha explícita para un comentario, pero no
está expuesta ni por su CLI, ni por su servidor, ni como librería.

**La exportación es literalmente el formato de importación.** Ninguna de las dos lo cumple: la
exportación de Backlog.md es un informe de solo lectura, y la de Taskwarrior mezcla datos reales con
derivados como el identificador de sesión y la urgencia sin separarlos. En `biso`, `export` escribe
todos los campos no derivados y ninguno derivado, y la ida y vuelta es una prueba de la suite.

Hay una trampa concreta en esto, y cuesta verla: **si el formato de lote admite los criterios de
aceptación como simples cadenas de texto, la simetría es falsa por construcción**, porque un criterio
tiene clave, texto y marca de cumplido, y al reimportar se pierden las claves y las marcas. Por eso el
formato admite objetos.

**Una funcionalidad no se retira sin anunciarla, y un cambio de formato lleva su camino de migración.**
Taskwarrior 3.0 retiró el historial de una de sus vistas sin aviso, y su cambio de motor de
almacenamiento provocó un caso real documentado de pérdida total de la base de tareas en Arch Linux,
porque el paquete se actualizó sin incluir el script de migración. De ahí sale el ciclo de aviso
obligatorio del contrato de estabilidad.
