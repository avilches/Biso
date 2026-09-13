# Parte 3. Lo que la gente quiere conservar, y lo que rechaza

De todo el ruido salen tres cosas que nadie quiere perder: **el grafo de dependencias**, **poder dejar
fuera del listado lo que no se puede coger ahora**, y **poder fichar una tarea de forma atómica**. Las
tres están ya en la especificación de `biso` como `dependencies`, como los filtros que se combinan en
`biso ls --not-blocked --not-waiting`, y como `biso start`. La segunda no es una sola bandera a
propósito: ninguna puede decir por sí misma que una tarea esté lista, porque cuántos filtros hace falta
descartar depende de qué se busque, y ["El porqué de reglas concretas"](../decisiones/comandos-y-flags.md#el-porqué-de-reglas-concretas) cuenta por qué se retiró el
nombre `--ready`, que lo prometía sin poder cumplirlo.

Y dos que rechaza de forma consistente: **el proceso en segundo plano** y **el almacén opaco**. El
primero está descartado. El segundo es la tensión real de esta decisión, y la respuesta es que el
almacén no sea opaco por fuera aunque sea una base de datos por dentro: hay una exportación fiel y
probada, versionada en el propio directorio del tablero, y `biso where` explica siempre qué tablero se
usó y por qué.

El ciclo del espacio se ha completado una vez en menos de un año. Los ficheros markdown sueltos
generaron Beads para darles dependencias y un grafo. La complejidad de Beads generó un reemplazo
deliberadamente simple, de vuelta al texto plano pero conservando el grafo. Y el motor de Beads tuvo que
publicar una marcha atrás para recuperar a quien trabaja solo. Merece la pena tenerlo presente antes de
añadir cualquier cosa.
