# Por dónde empezar a implementar

En el orden en que cada pieza paga lo que cuesta:

1. **El almacén**: abrir o crear la base de datos, el modo WAL y la transacción dentro de la que
   ocurre cualquier escritura, que es de lo que dependen las seis garantías de la sección 4.10. El
   controlador ya está elegido, `modernc.org/sqlite` sobre `database/sql` y sin `cgo`, con las cifras
   que lo deciden en el apartado 14.1 de `DECISIONES.md`. **Lo que se escribe aquí es la prueba del
   presupuesto de arranque de la sección 4.13**, que no se puede escribir antes de que el almacén
   exista y que hasta ahora solo se ha medido con un programa de prueba y no con el comando de verdad.
2. **El modelo de datos lógico** de la sección 5, con las claves estables de los criterios y el
   rechazo explícito de lo desconocido.
3. **El algoritmo de coincidencia** de la sección 6.1, que es una función pura de veinte líneas y de
   la que dependen todos los comandos.
4. **`init` y `where`**, que son crear un tablero y saber cuál es. Van aquí y no al final porque no hay
   forma de ejecutar ni de probar el paso siguiente sin un tablero, y crear un tablero es `init`.
5. **`new`, `ls`, `get` y `set`**, que son el trabajo diario.
6. **Los seis verbos de ciclo** de 10.7, `start`, `note`, `comment`, `finish`, `ask` y `answer`, que
   son azúcar sobre `set` y se escriben encima.
7. **`prime`**, que es lo que hace que todo lo anterior se use bien sin leer nada más.
8. **El lote de `new --from`, `export`, `snapshot` e `init --from`**, los cuatro juntos porque la prueba
   de simetría los necesita a la vez: exportar un tablero e importarlo tiene que dar dos tableros
   idénticos campo a campo, y lo mismo una instantánea restaurada.
9. **El resto**: `archive`, `config`, `doctor`, `board` y `help`.

Las garantías de la sección 4.10 no son un paso de esta lista: hay que respetarlas desde el primer
comando que escriba. El paso 1 es el que da las herramientas para respetarlas, no una excepción a esa
regla.
