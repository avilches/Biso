# Lo que queda por cerrar en la especificación

Este documento existe porque `docs/SPEC.md` está escrito para implementarse sin preguntar nada, y estas
son las preguntas que todavía se le pueden hacer. Ninguna impide empezar a implementar: la especificación
es coherente y completa en todo lo que el trabajo diario toca. Lo que sigue son huecos que un
implementador encontraría al llegar a un rincón concreto, incoherencias que no cambian el comportamiento,
y decisiones aplazadas a propósito.

**Cuando una entrada se cierra, se quita de aquí.** El 2026-09-09 se cerró todo lo que quedaba de las dos
revisiones de la rama de persistencia: los nueve huecos que un implementador habría tenido que rellenar
solo, el orden de implementación de la sección 15, la última incoherencia de vocabulario y los cinco
huecos que habían aparecido al cerrar los demás. La sección 12.1 de `DECISIONES.md` guarda el porqué de
los que llevaban una decisión de fondo detrás, y la sección 2 de aquí conserva las decisiones que siguen
aplazadas a propósito.

Ese mismo día se cerró después **el contrato de ejecución del sistema de control de versiones**, que era
el primero de los tres huecos que la ronda había abierto: qué se hace con lo que escriben las órdenes, si
hay tiempo máximo de espera y qué rutas da `{files}`. Las reglas quedaron en un apartado de la sección
10.14 de `SPEC.md` que vale igual para `git` y para `custom`, y el porqué en la sección 12.2 de
`DECISIONES.md`. De paso salió un error real del documento, que `{files}` y la clave `files` del JSON de
`snapshot` nombraban conjuntos distintos de ficheros sin decirlo.

Y el 2026-09-10 se cerró **el controlador de SQLite**, que era la primera de las decisiones aplazadas y
la única que bloqueaba escribir código. El resultado y su porqué están en el apartado 14.1 de
`DECISIONES.md`, y las cifras completas en `bench/sqlite-driver/RESULTADOS.md`. Al medirlo aparecieron
dos correcciones al apartado 13, que daba por constante un suelo que era solo de macOS y por sin medir un
almacén que ya lo está.

Dos cosas que se aprendieron por el camino y conviene no volver a aprender. La primera: **ante cada hueco
merece la pena preguntarse si existe porque falta decidir algo o porque sobra la cosa que lo abre.** El
hueco de `--board` no se rellenó, se disolvió al retirar la bandera cuando se vio que `-C` ya llegaba a
todo lo que ella prometía, y el de la instantánea que no podía cruzar de máquina se cerró al revés, no
dándole una vía sino dejando de cablear git y admitiendo que el sistema de control de versiones es una
elección del usuario. La segunda: **no escribas aquí una frase que cuente elementos sin contarlos.** Al
revisar este documento se descubrió que sus propios recuentos estaban desincronizados tres veces, que es
exactamente el fallo que él denuncia como dominante en la especificación.

---

## 1. Huecos abiertos

Ninguno impide implementar la versión 1.0 con `git`, que es el valor por defecto y el único camino
medido. Los dos primeros los abrió la ronda del control de versiones del 2026-09-09; el tercero salió al
medir el controlador de SQLite el 2026-09-10.

**El catálogo tiene un solo miembro conocido, así que la palabra promete más de lo que hay.** `git` es el
único sistema cuya receta está escrita, y la de cualquier otro (Mercurial, Jujutsu, Subversion) no existe:
habría que decir cómo se le pregunta si un directorio está dentro de un repositorio, cuál es su raíz y si
ignora una carpeta, que son las dos preguntas de las que depende dónde acaba la revisión. Mientras eso no
se escriba, quien use otro sistema pasa por `custom` y pierde el identificador de la revisión.

**`biso doctor` no comprueba que el fichero de exclusión del tablero cuadre con el sistema configurado.**
La sección 10.1 declara que si alguien cambia la clave `vcs` después de crear el tablero, el fichero se
queda con el nombre del sistema anterior y hay que arreglarlo a mano. Es una comprobación que `doctor`
podría hacer y no hace, y añadirla obliga a tocar las tres frases que cuentan las dieciocho filas de su
tabla, así que no se hizo dentro de esa ronda.

**La sección 4.13 no dice el sistema operativo de la máquina de referencia, y ahora se sabe que eso
cambia mucho el margen.** Esa sección amarra el presupuesto de 25 milisegundos a "la máquina que ejecuta
la suite de integración continua", a propósito, para no llevar dentro del documento la ficha técnica de
un ordenador. Al medir el controlador (apartado 14.1 de `DECISIONES.md`) salió que el suelo para arrancar
un binario de Go que no hace nada es de 8,1 milisegundos en macOS y de 0,37 en Linux, veintidós veces
menos, y que la lectura real del tablero tarda 14,5 milisegundos en el primero y 3,2 en el segundo. Las
dos plataformas cumplen el presupuesto, así que nada está bloqueado, pero el mismo código pasa la prueba
con una vez y siete décimas de margen o con casi ocho según el sistema del ejecutor, y el documento no
dice cuál es. Cuando exista la integración continua habrá que fijarlo, o decir que da igual y por qué.

---

## 2. Decisiones aplazadas a propósito

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
