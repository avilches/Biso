# Plan de implementación: la persistencia de biso

> **Para agentes:** SUB-SKILL OBLIGATORIA: usa `superpowers:subagent-driven-development`
> (recomendada) o `superpowers:executing-plans` para ejecutar este plan tarea a tarea. Los pasos
> usan casillas (`- [ ]`) para el seguimiento.

**Objetivo:** llevar a `docs/SPEC.md` y a `docs/DECISIONES.md` la decisión de cómo se guardan los
datos, de forma que la especificación vuelva a estar completa y se pueda empezar a implementar
código.

**Arquitectura:** el trabajo es documental, no hay código. Un tablero pasa a ser una base de datos
SQLite en un directorio propio fuera del proyecto, localizada por un fichero puntero versionado en
git, con el historial en una exportación de texto. Las tareas de este plan son grupos de cambios en
la especificación que se pueden revisar y rechazar por separado.

**Herramientas:** Markdown, `grep`, `wc -c`. Ninguna dependencia nueva.

**Diseño:** [`../specs/2026-09-07-persistencia-design.md`](../specs/2026-09-07-persistencia-design.md).
La evidencia sobre otras herramientas está en [`../../ESTADO-DEL-ARTE.md`](../../ESTADO-DEL-ARTE.md).
Los tres documentos viajan juntos: quien ejecute este plan tiene que leer el diseño antes de empezar.

## Restricciones globales

Valen para todas las tareas y no se repiten en cada una.

- **La documentación y los comentarios van en español.** Los identificadores, los comandos, las
  banderas, los textos de ayuda, los mensajes de error y las claves JSON van en inglés.
- **Nunca em-dash**, en ningún texto. Se comprueba con `grep -c '—' <fichero>`, que tiene que dar 0.
- **Los mensajes de commit no llevan coautoría** ni mención de haber sido generados por un agente.
- **Las tildes de un mensaje de commit no son un requisito y no son un hallazgo.** La prosa de la
  documentación sí lleva su ortografía completa, pero el asunto o el cuerpo de un commit sin
  acentos no se reporta como defecto ni se arregla. Decidido por el usuario el 2026-09-07.
- **El mensaje de `biso prime` tiene un tope duro de 5.120 bytes**, repartido en 3.456 de parte fija y
  1.664 de resumen (sección 9.5). La medida de hoy es 4.689 bytes en total, 3.255 de parte fija y
  1.434 de resumen, con 431 de margen. **Cualquier cambio en el texto literal de la sección 9.7 obliga
  a recalcular con `wc -c`, nunca a estimar**, y a actualizar las cifras en `SPEC.md` y en
  `DECISIONES.md` a la vez.
- **Los ejemplos de salida se generan, no se escriben a mano.** Si cambia la salida de un comando, su
  ejemplo se regenera y se comprueba carácter a carácter.
- **Un valor que no existe es siempre un error**, se esté escribiendo o leyendo. Ningún filtro mal
  escrito puede devolver una lista vacía.
- **Se referencian secciones por su número y su título, nunca por número de línea**, porque las líneas
  se mueven a medida que las tareas editan el documento.
- **El trabajo va en el worktree** `.claude/worktrees/persistencia`, rama `worktree-persistencia`, que
  ya existe. No se toca `main`.

## Decisiones que este plan cierra

El diseño dejó seis cosas abiertas a propósito. Cinco se cierran aquí con un valor concreto, porque un
plan no puede llevar huecos. La sexta queda fuera de alcance.

| Cosa abierta | Valor que se fija | Por qué |
|---|---|---|
| Nombre del fichero puntero | `.biso.json` en la raíz del proyecto | Oculto, evidentemente de `biso`, y la extensión declara su formato |
| Forma del identificador del tablero | 8 caracteres hexadecimales en minúscula | Va dentro del nombre de la carpeta, así que se lee con los ojos y conviene que sea corto |
| Nombre de la carpeta del tablero | `<nombre>-<id>`, por ejemplo `kex-3f9a2b1c` | Permite que dos proyectos de la misma máquina se llamen igual, y hace que buscar el tablero sea una sola búsqueda de patrón sin leer la configuración de ninguno |
| Forma de la instantánea | Dos ficheros en el directorio del tablero, `tasks.ndjson` y `config.json` | La configuración cambia poco y las tareas mucho, así que en ficheros separados los diffs quedan limpios |
| Nombre del paso de exportar y commitear | `biso snapshot` | Es un sustantivo y dice lo que produce |
| Cifra del presupuesto de arranque | 25 ms de reloj para `biso ls` y `biso prime` sobre un tablero de 300 tareas | Deja tres veces de margen sobre los 8,7 ms medidos en Go y excluye los lenguajes interpretados, cuyo solo arranque cuesta 24,5 y 33 ms |
| **El lenguaje de implementación** | **Fuera de alcance** | La especificación es independiente del lenguaje. El presupuesto de arranque entra igual, y acota la lista sin cerrarla |

Y una decisión que reduce el trabajo: **liberar un arrendamiento vencido no necesita comando nuevo.**
La sección 9.2 de `DECISIONES.md` ya dice que reclamar una tarea sería la misma escritura que hace
`biso start`, así que quien la reclama libera el arrendamiento vencido en la misma transacción, con la
comprobación del tenedor dentro. Así el plan añade solo dos comandos, `biso snapshot` y `biso rename`.

---

## Tarea 1: El directorio del tablero, el puntero y la raíz por defecto

Es la base que todas las demás tareas referencian, así que va primera.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 3.2 "Cómo se elige el tablero"
- Modificar: `docs/SPEC.md`, sección 10.10 "biso config", tabla de claves
- Modificar: `docs/SPEC.md`, sección 10.1 "biso init"

**Vocabulario que esta tarea establece y que las demás usan:**
- **directorio del tablero**: la carpeta que contiene la base de datos y la instantánea
- **puntero**: el fichero `.biso.json` en la raíz del proyecto, versionado en git, con `id` obligatorio
  e inmutable y `path` opcional
- **raíz por defecto**: el directorio de la máquina donde caen los tableros nuevos
- **nombre del tablero**: la etiqueta humana, que vive en la configuración del tablero y da nombre a su
  carpeta

- [ ] **Paso 1: Leer el diseño**

Leer completas las secciones 1, 2, 3 y 11 de `docs/superpowers/specs/2026-09-07-persistencia-design.md`.
No empezar a editar sin haberlas leído: el resto de los pasos asume ese contexto.

- [ ] **Paso 2: Reescribir el paso 4 de la sección 3.2**

Hoy el paso 4 dice "Un tablero que el propio almacenamiento asocia al directorio de trabajo, buscado
con el mismo procedimiento hacia arriba y el mismo tope". Pasa a ser la raíz por defecto de la máquina.
Mantener intactos los pasos 1, 2 y 3, el tope de la búsqueda hacia arriba, y el mensaje de código 8 con
su texto literal.

- [ ] **Paso 3: Describir el puntero en la sección 3.2**

Añadir, después de la lista de cuatro vías, la descripción del puntero: que es el fichero `.biso.json`
en la raíz del proyecto, que va versionado en git, y que lleva estas claves:

| Clave | Tipo | Obligatoria | Notas |
|---|---|---|---|
| `version` | entero | sí | versión del formato del puntero |
| `id` | 8 caracteres hexadecimales | sí | identidad del tablero, inmutable |
| `path` | ruta | no | solo cuando el tablero no está en la raíz por defecto |

Y las tres reglas: que una clave desconocida es un error, que el `id` manda y el `path` es una pista que
puede no resolver, y cómo se busca.

**Cómo se busca**, que es el paso que hay que dejar sin ambigüedad. La carpeta de un tablero se llama
`<nombre>-<id>`, por ejemplo `kex-3f9a2b1c`, así que localizarlo desde el puntero es **una sola búsqueda
del patrón `*-<id>` en la raíz por defecto**, sin abrir ni leer la configuración de ningún tablero.
Documentar también las dos cosas que eso resuelve: que dos proyectos de la misma máquina se pueden llamar
igual sin chocar, y que el puntero sigue resolviendo aunque el tablero se haya renombrado, porque el
identificador viaja en el nombre de la carpeta.

Y que `biso init` genera el identificador de la fuente de números aleatorios del sistema comprobando que
no exista ya en la raíz por defecto, que es una lectura de directorio.

- [ ] **Paso 4: Enunciar los dos casos del diseño que faltaban**

En la misma sección 3.2, dejar dicho que **dos proyectos distintos pueden apuntar legalmente al mismo
tablero**, porque es el mismo mecanismo que hace que varias copias de trabajo lo compartan (cambio 16
del diseño). Y que un tablero cuyo proyecto ya no existe queda huérfano en la raíz por defecto y ningún
comando de hoy lo ve.

- [ ] **Paso 5: Situar el nombre del tablero en su configuración**

El nombre humano del tablero es una clave de la configuración del tablero, no del puntero, porque un
nombre es una etiqueta y las etiquetas colisionan entre personas, así que no puede ser la identidad.

**Antes de crear una clave nueva, comprobar si `project_name`, que ya existe en la tabla de la sección
10.10, es exactamente esta cosa.** Si lo es, no se crea ninguna clave y se documenta que ese es el nombre
del tablero y el que da nombre a su carpeta. Si es otra cosa, se crea la clave que haga falta y se dice
en una frase en qué se diferencian.

Esta decisión la necesita la tarea 3, que deriva el prefijo de los identificadores del nombre del
tablero, así que tiene que quedar resuelta aquí y no más adelante.

- [ ] **Paso 6: Añadir las claves de configuración de máquina**

En la sección 10.10, añadir a la tabla de claves la raíz por defecto y la lista de raíces adicionales,
con sus tipos y sus valores por defecto, y decir explícitamente que son de máquina y no de tablero,
porque el resto de esa tabla es configuración del tablero. Respetar el estilo de la tabla existente.

- [ ] **Paso 7: Ajustar `biso init` en la sección 10.1**

`biso init` sin `--at` crea el tablero en la raíz por defecto y escribe el puntero. Con `--at` lo pone
donde se le diga y escribe el puntero igual, con `path`. Hoy la sección 3.2 dice que el puntero se crea
"siempre que el tablero no quede dentro del propio proyecto", y esa condición desaparece: el puntero se
escribe siempre, porque el tablero nunca queda dentro del proyecto.

- [ ] **Paso 8: Comprobar la coherencia**

```bash
grep -c '—' docs/SPEC.md                 # tiene que dar 0
grep -n 'asocia al directorio de trabajo' docs/SPEC.md   # no debe quedar ninguna
grep -n '\.biso\.json' docs/SPEC.md      # aparece en 3.2 y en 10.1
```

Leer la sección 3.2 entera de una vez y confirmar que las cuatro vías siguen siendo cuatro, que el
mensaje de código 8 no ha cambiado, y que ninguna frase contradice a otra.

- [ ] **Paso 9: Commit**

```bash
git add docs/SPEC.md
git commit -m "Fija el directorio del tablero, el puntero y la raíz por defecto"
```

---

## Tarea 2: El error del puntero que nombra un tablero ausente

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 3.2
- Modificar: `docs/SPEC.md`, sección 10.1 "biso init"
- Modificar: `docs/SPEC.md`, sección 12, la lista de claves `code` del código 8

**No se toca la tabla de códigos de salida de la sección 2**, porque no se añade ningún código nuevo.

**Consume de la tarea 1:** el puntero `.biso.json` y su clave `id`.

- [ ] **Paso 1: Escribir el caso y su mensaje**

Añadir a la sección 3.2 el caso de un puntero que existe y nombra un tablero que en esta máquina no
está, que es lo que pasa al clonar el proyecto en otro ordenador. El mensaje de hoy,
`error: no board here, and none configured for this project`, sería falso, porque sí hay uno
configurado. Escribir un mensaje propio en inglés que diga las tres cosas: que hay un puntero, cuál es
el identificador que nombra, y que `biso init` crea el tablero aquí adoptando esa identidad. Seguir el
estilo de los mensajes de la sección 7.3, con `error:` y con una línea `hint:`.

- [ ] **Paso 2: Usar el código 8 con una clave `code` propia**

**No se añade ningún código de salida nuevo.** El código 8 `NO_BOARD` dice "no hay tablero accesible
desde donde se ha llamado", y eso es exactamente verdad en este caso: lo que estaba mal era solo el
texto del mensaje, que afirma que no hay ninguno configurado. Para quien llama la situación es la misma,
no puede trabajar, y el remedio también, ejecutar `biso init`.

Lo que sí cambia es la clave `code` del sobre JSON de la sección 12, que es el mecanismo que la
especificación tiene para distinguir dos situaciones que comparten código de salida sin leer el mensaje.
Añadir una clave nueva para este caso a la lista del código 8 en la sección 12, en el estilo de las que
ya existen (`malformed_id`, `never_allocated`, `not_found`).

- [ ] **Paso 3: Documentar que `biso init` adopta la identidad**

En la sección 10.1, decir que si existe un puntero cuyo tablero no está, `biso init` crea el tablero
usando el `id` del puntero en vez de acuñar uno nuevo, para que las dos máquinas sigan hablando del
mismo tablero. Y que no reescribe el puntero.

- [ ] **Paso 4: Comprobar**

```bash
grep -c '—' docs/SPEC.md                 # 0
grep -n 'no board here' docs/SPEC.md     # el mensaje viejo sigue existiendo para su caso
```

Confirmar que los dos mensajes coexisten y que cada uno dice cuándo aplica.

- [ ] **Paso 5: Commit**

```bash
git add docs/SPEC.md
git commit -m "Da su propio error al puntero que nombra un tablero que no está en esta máquina"
```

---

## Tarea 3: El prefijo de los identificadores

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 4.11 "Identificadores"
- Modificar: `docs/SPEC.md`, sección 10.10 "biso config", tabla de claves y tabla de casos
- Modificar: `docs/SPEC.md`, sección 10.1 "biso init"

**Consume de la tarea 1:** el nombre del tablero.

- [ ] **Paso 1: Cambiar el valor por defecto de `task_prefix`**

En la sección 4.11 y en la tabla de la sección 10.10, `task_prefix` deja de tener `TASK` como valor por
defecto y se deriva del nombre del tablero en mayúsculas. Explicar el motivo en una frase: que dos
tableros con `TASK` por defecto colisionan los dos en `TASK-1` y hacen inservible cualquier vista que
junte varios proyectos.

- [ ] **Paso 2: Documentar la regla para el nombre que no da un prefijo válido**

El tipo de `task_prefix` es "texto de solo letras", así que un tablero llamado `mi-proyecto-2` no da un
prefijo válido tal cual. La regla que se documenta es: **se quitan los caracteres que no son letras y el
resto se pasa a mayúsculas**, así que `mi-proyecto-2` da `MIPROYECTO`. Y si al quitarlos no queda ninguna
letra, por ejemplo en un tablero llamado `2026`, **`biso init` falla y pide el prefijo explícitamente**,
en vez de inventarse un valor en silencio.

Ese fallo no necesita nada nuevo: es el código 2 con la clave `code` **`invalid_prefix`**, que ya existe
en la lista de la sección 12 y que ya cubre un prefijo inválido. Reutilizarla en vez de crear otra.

- [ ] **Paso 3: Hacer `task_prefix` inmutable, que es el agujero de hoy**

En la tabla de casos de la sección 10.10, añadir que cambiar `task_prefix` cuando el tablero ya tiene
alguna tarea es error 6, con la misma forma que ya tiene quitar de `statuses` un estado que alguna tarea
usa. El mensaje remite a la ruta de exportar, reescribir los identificadores e importar en un tablero
nuevo.

Este paso cierra una contradicción que ya existe: la sección 10.10 termina diciendo que ningún cambio de
configuración toca ninguna tarea, nunca, y `task_prefix` es la única clave cuyo valor está incrustado en
datos que ya existen. Comprobar que esa frase final sigue siendo verdadera después del cambio, porque
ahora lo es de verdad.

- [ ] **Paso 4: Comprobar**

```bash
grep -c '—' docs/SPEC.md                            # 0
grep -n 'task_prefix' docs/SPEC.md                  # las cinco menciones siguen coherentes
grep -n 'por defecto `TASK`' docs/SPEC.md           # no debe quedar ninguna
grep -n 'Ningún cambio de configuración toca' docs/SPEC.md   # la frase sigue ahí
```

- [ ] **Paso 5: Commit**

```bash
git add docs/SPEC.md
git commit -m "Deriva el prefijo del nombre del tablero y lo hace inmutable con tareas ya creadas"
```

---

## Tarea 4: El nombre del tablero y `biso rename`

**Ficheros:**
- Crear: `docs/SPEC.md`, **subsección 10.12** para `biso rename`

**Consume de la tarea 1:** el nombre del tablero, ya situado en la configuración del tablero por el paso
5 de esa tarea, y su relación con el nombre de la carpeta.
**Produce para la tarea 5:** el comando `biso rename` y el hecho de que nombre e identidad son
distintos.
**Produce para la tarea 10:** un comando administrativo más.

El número de subsección, 10.12, está asignado de antemano a propósito, porque la tarea 8 también añade
una subsección de comando y las dos elegirían el mismo número por separado.

- [ ] **Paso 1: Especificar `biso rename`**

Escribir la sección del comando con la misma estructura que tienen los demás: firma, tabla de
parámetros, comportamiento, casos límite, salida literal, esquema JSON, códigos de salida y texto de
ayuda. El comando cambia el nombre y mueve la carpeta del tablero, **conservando el sufijo del identificador**:
`kex-3f9a2b1c` pasa a `nuevonombre-3f9a2b1c`. **No toca ningún puntero**, y eso es consecuencia directa
de que el identificador sea la identidad y viaje en el nombre de la carpeta: conviene decirlo
explícitamente en el comportamiento.

Como el sufijo se conserva, la carpeta de destino **no puede existir ya** salvo que algo esté corrupto,
así que ese caso deja de ser un caso de uso normal y pasa a ser un error de entorno, código 7. Un nombre
nuevo que coincida con el de otro tablero es perfectamente legal y no es un error, porque los sufijos
difieren.

- [ ] **Paso 2: Comprobar**

```bash
grep -c '—' docs/SPEC.md    # 0
grep -n '^### 10\.1[0-9]' docs/SPEC.md   # la 10.12 existe y no duplica ningun numero
```

Comparar la sección nueva con la de otro comando de administración, por ejemplo `biso where`, y
confirmar que tiene las mismas partes en el mismo orden.

- [ ] **Paso 3: Commit**

```bash
git add docs/SPEC.md
git commit -m "Especifica biso rename"
```

---

## Tarea 5: `biso where` dice las tres cosas

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 10.2 "biso where"

**Consume:** el puntero de la tarea 1, la identidad de la tarea 2, el nombre de la tarea 4.

- [ ] **Paso 1: Ampliar lo que informa**

Hoy `biso where` dice qué tablero se ha usado y por qué. Ahora hay tres cosas distintas que antes eran
una: el identificador del tablero, su nombre, y la ruta de su directorio. Documentar que informa de las
tres, y de cuál de las cuatro vías de la sección 3.2 ganó.

- [ ] **Paso 2: Regenerar el ejemplo de salida**

El ejemplo de salida de esta sección cambia, y los ejemplos de este documento se generan y se comprueban
carácter a carácter. Regenerar el ejemplo con el algoritmo de columnas que la especificación define, y
actualizar también el esquema JSON de la sección con las claves nuevas.

- [ ] **Paso 3: Comprobar**

```bash
grep -c '—' docs/SPEC.md    # 0
```

Confirmar que el ejemplo y el esquema JSON dicen lo mismo, con las mismas claves.

- [ ] **Paso 4: Commit**

```bash
git add docs/SPEC.md
git commit -m "Hace que biso where diga el identificador, el nombre y la ruta del tablero"
```

---

## Tarea 6: El arrendamiento con caducidad

Es la tarea que cierra la decisión que la sección 9.2 de `DECISIONES.md` dejó aplazada, y la única que
toca los dos documentos.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 5 (campos de una tarea) y sección 5 (lista de campos derivados)
- Modificar: `docs/SPEC.md`, sección 10 donde se especifica `biso start`
- Modificar: `docs/DECISIONES.md`, sección 9.2 "Saber si alguien está trabajando de verdad"

- [ ] **Paso 1: Leer la sección 6 del diseño y la 9.2 de DECISIONES**

Hay una contradicción entre las dos y hay que entenderla antes de escribir. La sección 9.2 dice que la
caducidad, al vencer, saca la tarea del estado activo. Eso no se sostiene, porque el estado es un campo
guardado y un campo derivado no puede cambiarlo.

- [ ] **Paso 2: Añadir los campos guardados a la sección 5**

Dos campos nuevos en la tabla de campos de una tarea: el instante de caducidad del arrendamiento y quién
lo tiene. Con su tipo, si son obligatorios, quién los fija y si son mutables, en el mismo formato que las
demás filas. Los dos son opcionales y solo tienen valor en una tarea activa y asignada.

- [ ] **Paso 3: Añadir el campo derivado**

Añadir a la lista canónica de campos derivados de la sección 5, que hoy tiene diez, el que dice si el
arrendamiento está vencido. **Esa lista es la única canónica**, así que hay que comprobar que no queden
otras enumeraciones de campos derivados en el documento diciendo diez.

- [ ] **Paso 4: Documentar que el estado guardado no cambia solo**

Escribir explícitamente que lo que vence es la reclamación y no el estado: una tarea con el arrendamiento
vencido sigue teniendo guardado el estado activo, y ningún comando la mueve por su cuenta. Decir también
por qué no se hace de otra forma, porque una escritura diferida haría que un comando toque tareas que no
nombró.

- [ ] **Paso 5: Documentar la reclamación en `biso start`**

`biso start` sobre una tarea cuyo arrendamiento venció la reclama, poniendo el arrendamiento a nombre de
quien llama, **comprobando el tenedor dentro de la misma transacción** para que el propietario viejo no
la robe de vuelta al despertar.

`biso start` sobre una tarea cuyo arrendamiento está **vivo y es de otro** avisa y la coge igual, con un
`warning:` por stderr que diga de quién era y cuándo vence. No se niega, y hay dos motivos escritos para
eso. El primero es que la especificación aplica "avisa, no impide" de forma consistente en todos los
demás sitios donde podría bloquear, incluidas las dependencias sin terminar y las preguntas abiertas,
tanto en `start` como en `finish`. El segundo es que un bloqueo de flujo no evita el trabajo duplicado,
solo empuja a rodear la herramienta modificando datos que no deberían tocarse, y un dato mal puesto para
esquivar un bloqueo es peor que un aviso ignorado.

Si al escribirlo se decide lo contrario, hay que dejar registrado el motivo en `DECISIONES.md`, porque
sería el único bloqueo de flujo de toda la especificación y una excepción así no puede quedar sin
explicación.

- [ ] **Paso 6: Enmendar la sección 9.2 de DECISIONES.md**

Sustituir la frase sobre sacar la tarea del estado activo, **dejando registrado que es una enmienda y por
qué**, en el estilo del documento, que registra la evidencia detrás de cada decisión. La enmienda no puede
aparecer como si la decisión siempre hubiera sido esta. Apuntar también lo que el estado del arte aporta:
que el patrón tiene nombre propio ahí fuera y que su segunda mitad, la comprobación del tenedor, era la
pieza que faltaba.

- [ ] **Paso 7: Decidir si `biso prime` lo menciona, y recalcular los bytes si lo hace**

Un agente que arranca una sesión querría saber que una tarea quedó reclamada por una sesión muerta, así
que hay un argumento fuerte para que `prime` lo diga. Si se decide que sí:

```bash
# extraer el bloque literal de la seccion 9.7 a un fichero y medirlo
wc -c /tmp/prime-nuevo.txt
```

Y actualizar **las cuatro cifras a la vez**: el total, la parte fija y el resumen en la sección 9.5 y en
la 9.7 de `SPEC.md`, y las mismas cifras en la sección 3 de `DECISIONES.md`. Las de hoy son 4.689 en
total, 3.255 de parte fija y 1.434 de resumen, contra un tope de 5.120 con 431 de margen. **Nunca
estimar**: este es el patrón de fallo que el propio proyecto identificó como el más frecuente.

- [ ] **Paso 8: Comprobar**

```bash
grep -c '—' docs/SPEC.md docs/DECISIONES.md      # 0 en los dos
grep -n 'saca la tarea del estado activo' docs/DECISIONES.md   # no debe quedar ninguna
grep -n '4\.689\|3\.255\|1\.434' docs/SPEC.md docs/DECISIONES.md  # si cambiaron, ninguna vieja queda
```

- [ ] **Paso 9: Commit**

```bash
git add docs/SPEC.md docs/DECISIONES.md
git commit -m "Cierra el arrendamiento con caducidad: lo que vence es la reclamación, no el estado"
```

---

## Tarea 7: Qué significa un dato que no se puede interpretar

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 4.12 "Qué pasa con un dato que no se puede interpretar"
- Modificar: `docs/SPEC.md`, sección 10.11 "biso doctor"

- [ ] **Paso 1: Separar los dos casos**

La sección 4.12 está escrita pensando en un fichero por tarea, y con una base de datos son dos casos
distintos que hoy se dicen como uno. El primero, una tarea cuyo contenido no se puede interpretar,
mantiene el comportamiento actual entero: error 3 en las lecturas dirigidas, saltar y avisar en las de
conjunto, código 6 en `export`, y reportar en `doctor`. El segundo, la base de datos que no abre o que
falla su comprobación de integridad, **no es una tarea ilegible**: es que no hay tablero legible, y
necesita su propio código y su propio mensaje. No puede presentarse como "una tarea se ha saltado".

- [ ] **Paso 2: Escribir el mensaje del segundo caso**

En inglés, con `error:` y una línea `hint:`, en el estilo de la sección 7.3. Documentar su código de
salida en la tabla correspondiente.

- [ ] **Paso 3: Ser honesto sobre lo que se pierde**

Añadir una frase diciendo que con un almacén por tarea el aislamiento del daño sale gratis y con una
base de datos hay que provocarlo, y que una corrupción puede llevarse más de una tarea. La
especificación de este proyecto dice las cosas incómodas en voz alta, no las esconde.

- [ ] **Paso 4: Añadir la comprobación de integridad a `biso doctor`**

En la sección 10.11, dos comprobaciones nuevas en la lista: la integridad de la base de datos, y un aviso
si el directorio del tablero está en un sistema de ficheros donde el modo WAL de SQLite no es seguro.
Decir para cada una si es reparable automáticamente, como hace el resto de la lista. La segunda no lo es:
es un aviso.

- [ ] **Paso 5: Comprobar**

```bash
grep -c '—' docs/SPEC.md    # 0
```

Leer la sección 4.12 entera y confirmar que los dos casos se distinguen sin ambigüedad y que la tabla de
tipos de lectura sigue cuadrando.

- [ ] **Paso 6: Commit**

```bash
git add docs/SPEC.md
git commit -m "Separa la tarea ilegible de la base de datos que no abre, y amplía biso doctor"
```

---

## Tarea 8: La instantánea, `biso snapshot` y la restauración

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 10.9 "biso export"
- Modificar: `docs/SPEC.md`, sección 10.3 "biso new --from"
- Crear: `docs/SPEC.md`, **subsección 10.13** para `biso snapshot`
- Modificar: `docs/SPEC.md`, sección 10.1 "biso init"

**Consume:** el directorio del tablero de la tarea 1.
**Produce para la tarea 10:** un comando administrativo más.

El número de subsección, 10.13, está asignado de antemano porque la tarea 4 ya ocupó la 10.12.

- [ ] **Paso 1: Documentar el contenido versionado del directorio del tablero**

El directorio del tablero es su propio repositorio de git, y lo que se versiona no es la base de datos:
se versionan `tasks.ndjson` y `config.json`, y se ignoran el fichero de la base de datos y sus auxiliares
de WAL. Explicar en una frase por qué no se versiona el binario: cada escritura reescribe páginas, así
que cada commit guardaría una copia completa, y git no puede diffearlo.

Decir además que **el repositorio de git es opcional y su ausencia no rompe nada**: la instantánea se
escribe igual y sigue sirviendo para restaurar, y lo único que se pierde es el historial. `biso` no puede
exigir que git esté instalado.

- [ ] **Paso 2: Especificar `biso snapshot`**

Con la estructura completa de un comando: firma, parámetros, comportamiento, casos límite, salida
literal, esquema JSON, códigos de salida y ayuda. Escribe los dos ficheros y, si el directorio del
tablero es un repositorio de git, commitea. Una bandera para no commitear. **Decir explícitamente que
ningún otro comando ejecuta git**, porque ejecutar `git` cuesta unos 12 milisegundos medidos y el
presupuesto de arranque de la tarea 9 no lo admite en el camino caliente.

- [ ] **Paso 3: Cerrar el agujero de la simetría**

Hoy la garantía de simetría de la sección 10.9 no basta para reconstruir un tablero, porque la
exportación lleva las tareas y no la configuración: el propio ejemplo de la suite vuelve a declarar los
estados, los tipos y las extensiones en el `init` del tablero de destino, y el documento avisa de que un
vocabulario distinto hace fallar el lote entero con código 9. Documentar que la instantánea incluye la
configuración, y **actualizar el ejemplo de la prueba de simetría** para que no tenga que declarar el
vocabulario a mano.

- [ ] **Paso 4: Especificar la restauración**

Restaurar tiene que ser una operación y no una receta. Especificar cómo se hace con `biso init`, leyendo
los dos ficheros de una instantánea, y qué pasa en los casos límite: una instantánea a medias, una con la
configuración pero sin tareas, y una cuyo `config.json` declara un vocabulario que las tareas no usan.

- [ ] **Paso 5: Comprobar**

```bash
grep -c '—' docs/SPEC.md    # 0
grep -n 'tasks.ndjson\|config.json' docs/SPEC.md   # coherentes en las cuatro secciones tocadas
```

Confirmar que la garantía de simetría, tal como queda escrita, es comprobable por la suite sin pasos
manuales.

- [ ] **Paso 6: Commit**

```bash
git add docs/SPEC.md
git commit -m "Añade la configuración a la instantánea y especifica biso snapshot y la restauración"
```

---

## Tarea 9: El presupuesto de arranque

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 4 (regla transversal nueva)
- Modificar: `docs/SPEC.md`, sección 13 "contrato de estabilidad"

- [ ] **Paso 1: Escribir el requisito con su cifra**

Una regla transversal nueva en la sección 4: `biso ls` y `biso prime` sobre un tablero de 300 tareas
terminan en menos de 25 milisegundos de reloj, y es una prueba de la suite. Decir que la cifra se mide en
una máquina de referencia y que la máquina de referencia hay que nombrarla, porque un número sin máquina
no se puede comprobar.

- [ ] **Paso 2: Escribir las tres reglas que lo protegen**

No hacer al arrancar trabajo que nadie pidió. No ejecutar git en el camino caliente. Y que la palanca
mayor no es hacer cada llamada más rápida sino hacer menos llamadas, que es para lo que existe
`biso prime`. Las tres salen de fallos ajenos documentados en el estado del arte, y conviene citarlo.

- [ ] **Paso 3: Dejar el presupuesto fuera del contrato de estabilidad, y decirlo**

La sección 13 congela hoy un solo número de tamaño, los 5.120 bytes del mensaje de arranque. **El
presupuesto de arranque no se suma a los congelados**, y la sección 13 tiene que decirlo explícitamente
en vez de dejarlo ambiguo.

El motivo es que los dos números no son de la misma naturaleza. Los 5.120 bytes son una propiedad del
texto, así que cualquiera puede medirlos y siempre dan lo mismo. Los 25 milisegundos son una propiedad de
una máquina, y congelar en un contrato de estabilidad un número que depende del hardware haría que la
herramienta incumpliera su propio contrato al ejecutarse en un ordenador más lento, sin que nadie haya
cambiado una línea. Sigue siendo una prueba de la suite, y sigue teniendo que fallar si alguien mete una
regresión: lo que no es, es una promesa de versión a versión.

- [ ] **Paso 4: Comprobar**

```bash
grep -c '—' docs/SPEC.md    # 0
grep -n '25 ms\|25 milisegundos' docs/SPEC.md   # la cifra aparece igual en todos los sitios
```

- [ ] **Paso 5: Commit**

```bash
git add docs/SPEC.md
git commit -m "Convierte el arranque rápido en un requisito con cifra y con prueba"
```

---

## Tarea 10: El recuento de comandos

Va después de las tareas 4 y 8 porque son las que añaden comandos. Es una tarea pequeña y propia porque
es la desincronización que este proyecto sufre más a menudo, y merece su propia revisión.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 10 (la frase del recuento)
- Modificar: `docs/SPEC.md`, sección 11 (la tabla de `biso help all` y el bloque de ayuda literal)

- [ ] **Paso 1: Localizar los cuatro sitios**

```bash
grep -n 'Diecinueve\|ocho comandos\|treinta y dos' docs/SPEC.md
grep -n 'init, where, archive, export, config, doctor, board, help' docs/SPEC.md
```

Hoy son: "Diecinueve comandos. Los once primeros son el ciclo de trabajo... los ocho" en la sección 10;
"la lista de los ocho comandos de administración" en la tabla de `biso help all`; "Son treinta y dos
líneas, y no incluyen los ocho comandos de administración" tras el bloque de ayuda; y la lista literal de
los ocho nombres dentro del bloque de ayuda.

- [ ] **Paso 2: Actualizar los cuatro con los comandos nuevos**

Con `biso snapshot` y `biso rename`, los administrativos pasan de ocho a diez y el total de diecinueve a
veintiuno. Actualizar la lista literal de nombres dentro del bloque de ayuda **en el orden que ese bloque
ya usa**, no alfabéticamente si no lo estaba.

- [ ] **Paso 3: Recontar las líneas del bloque de ayuda**

Si los comandos nuevos aparecen en `biso --help`, el número de líneas cambia. Contarlas de verdad sobre
el bloque literal, no estimarlas, y actualizar la frase. Si no aparecen porque son administrativos, decir
por qué el número no cambia.

- [ ] **Paso 4: Comprobar que no queda ningún recuento viejo**

```bash
grep -n 'Diecinueve\|ocho comandos\|treinta y dos' docs/SPEC.md   # ninguna ocurrencia vieja
grep -rn 'diecinueve\|Diecinueve' docs/ CLAUDE.md                 # tampoco fuera de SPEC.md
```

- [ ] **Paso 5: Commit**

```bash
git add docs/SPEC.md
git commit -m "Actualiza el recuento de comandos y la lista de los administrativos"
```

---

## Tarea 11: Cerrar la sección 14 y los documentos que apuntan a la decisión

Es la última porque solo se puede escribir cuando todo lo demás está.

**Ficheros:**
- Modificar: `docs/SPEC.md`, sección 14 (lo que queda fuera)
- Modificar: `docs/SPEC.md`, sección 7.3 (el tercero de los tres mensajes)
- Modificar: `docs/DECISIONES.md` (una sección nueva con el porqué, y el enlace al estado del arte)
- Modificar: `CLAUDE.md`

- [ ] **Paso 1: Reescribir la sección 14**

Quitar de la lista de cosas que quedan fuera todo lo que dependía de esta decisión: el control de
versiones, la pregunta de qué ocurre con una tarea que existe en una versión del proyecto y no en otra, y
el arrendamiento. Y dejar lo que de verdad sigue fuera: la interfaz multiproyecto, exportar al formato de
Backlog.md, y sincronizar entre máquinas, cada uno con su motivo de una frase.

- [ ] **Paso 2: Ajustar el tercer mensaje de la sección 7.3**

Ese mensaje dice hoy que la tarea "was archived and then removed, or it belongs to a version of the
project that this board does not have". La segunda mitad deja de tener sentido, porque el tablero no está
en el árbol de trabajo y no hay versiones del proyecto que lo bifurquen. Reescribir el texto literal del
mensaje quitando esa mitad y **comprobar que su código y su clave `code` no cambian**, porque son
contrato.

- [ ] **Paso 3: Escribir la decisión en DECISIONES.md**

Una sección nueva con el porqué de esta decisión, en el estilo del documento, que registra la evidencia
detrás de cada regla. Tiene que decir por qué no se versiona el tablero con el código, por qué no hay
daemon (con la aritmética, no con el gusto), y por qué el texto es una salida y no un canal de vuelta.
Enlazar `docs/ESTADO-DEL-ARTE.md` como la evidencia, para que no quede huérfano.

- [ ] **Paso 4: Actualizar CLAUDE.md**

Su sección "Lo que falta por decidir, y bloquea" dice que la persistencia es la única decisión que
bloquea escribir código. Ya no lo es. Reescribirla diciendo qué se decidió y dónde está, y dejar el
lenguaje de implementación como lo que sigue abierto sin bloquear. Añadir el enlace al estado del arte.

- [ ] **Paso 5: Comprobar la coherencia del conjunto**

```bash
grep -c '—' docs/SPEC.md docs/DECISIONES.md CLAUDE.md docs/ESTADO-DEL-ARTE.md   # 0 en todos
grep -rn 'ESTADO-DEL-ARTE' docs/ CLAUDE.md      # enlazado desde DECISIONES.md y CLAUDE.md
grep -n 'pendiente de la persistencia\|depende de la decisión de persistencia' docs/SPEC.md docs/DECISIONES.md
```

La última búsqueda tiene que quedar vacía: si algo sigue diciendo que depende de la persistencia, es que
esta tarea no ha terminado. Revisar en particular la tabla de identidad de la sección 3.1, que hoy dice
"arrendamiento... pendiente de la persistencia".

- [ ] **Paso 6: Commit**

```bash
git add docs/SPEC.md docs/DECISIONES.md CLAUDE.md
git commit -m "Cierra la sección 14 y actualiza los documentos que apuntaban a la decisión pendiente"
```

---

## Revisión final, después de la tarea 11

No es una tarea con commit propio: es la comprobación de que el plan hizo su trabajo.

- [ ] Los dieciséis cambios de la sección 13 del diseño tienen su tarea, y ninguno se quedó fuera.
- [ ] La enmienda de la sección 9.2 de `DECISIONES.md` está hecha y registrada como enmienda.
- [ ] `grep -c '—'` da 0 en `SPEC.md`, `DECISIONES.md`, `CLAUDE.md`, `ESTADO-DEL-ARTE.md` y el diseño.
      Este plan es la única excepción, y solo porque el carácter aparece dentro de los propios comandos
      de comprobación.
- [ ] Ninguna búsqueda de "depende de la persistencia" devuelve nada.
- [ ] El recuento de comandos cuadra en los cuatro sitios.
- [ ] Si el texto de `biso prime` cambió, las cifras de bytes se midieron con `wc -c` y están iguales en
      `SPEC.md` y en `DECISIONES.md`.
- [ ] Las seis cosas que este plan cerró están escritas en la especificación con el valor que se fijó, y
      el lenguaje sigue siendo la única decisión abierta.
