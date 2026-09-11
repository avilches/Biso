# Por dónde empezar a implementar

En el orden en que cada pieza paga lo que cuesta:

1. **El almacén**: abrir o crear la base de datos, el modo WAL y la transacción dentro de la que
   ocurre cualquier escritura, que es de lo que dependen las garantías de la sección ["Concurrencia, atomicidad y garantías observables"](garantias.md#concurrencia-atomicidad-y-garantías-observables). El
   controlador ya está elegido, `modernc.org/sqlite` sobre `database/sql` y sin `cgo`, con las cifras
   que lo deciden en el apartado ["El controlador de SQLite es `modernc.org/sqlite`, sin `cgo`"](../DECISIONES.md#el-controlador-de-sqlite-es-moderncorgsqlite-sin-cgo) de `DECISIONES.md`. **Lo que se escribe aquí es la prueba del
   presupuesto de arranque de la sección ["El presupuesto de arranque"](presupuestos.md#el-presupuesto-de-arranque)**, que no se puede escribir antes de que el almacén
   exista y que hasta ahora solo se ha medido con un programa de prueba y no con el comando de verdad.
2. **El modelo de datos lógico** de la sección ["El modelo de datos de una tarea"](modelo-de-datos.md), con las claves estables de los criterios y el
   rechazo explícito de lo desconocido.
3. **El algoritmo de coincidencia** de la sección ["El algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia), que es una función pura de veinte líneas y de
   la que dependen todos los comandos.
4. **`init` y `where`**, que son crear un tablero y saber cuál es. Van aquí y no al final porque no hay
   forma de ejecutar ni de probar el paso siguiente sin un tablero, y crear un tablero es `init`.
5. **`new`, `ls`, `get` y `set`**, que son el trabajo diario.
6. **Los seis verbos de ciclo** de la sección ["Los verbos del ciclo: `start`, `note`, `comment`, `finish`, `ask`, `answer`"](cmd/verbos-del-ciclo.md), `start`, `note`, `comment`, `finish`, `ask` y `answer`, que
   son azúcar sobre `set` y se escriben encima.
7. **`prime`**, que es lo que hace que todo lo anterior se use bien sin leer nada más.
8. **El lote de `new --from`, `export`, `snapshot` e `init --from`**, los cuatro juntos porque la prueba
   de simetría los necesita a la vez: exportar un tablero e importarlo tiene que dar dos tableros
   idénticos campo a campo, y lo mismo una instantánea restaurada.
9. **El resto**: `archive`, `config`, `doctor`, `board` y `help`.

Las garantías de la sección ["Concurrencia, atomicidad y garantías observables"](garantias.md#concurrencia-atomicidad-y-garantías-observables) no son un paso de esta lista: hay que respetarlas desde el primer
comando que escriba. El paso 1 es el que da las herramientas para respetarlas, no una excepción a esa
regla.
