# La urgencia

`urgency` es un `float` derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
criterio de la tupla de orden por defecto de `biso ls`, después de `ordinal` ([`biso ls`](../cmd/ls.md)), y el que ordena
el resumen de `biso prime`.

```
Si el estado de la tarea es el terminal, urgency = 0.0 y no se calcula nada mas.

En cualquier otro caso:

urgency = 6.0  * prioridad         (ver la regla de prioridad, más abajo)
        + 4.0  * activa            (1.0 si el estado es el activo y no hay pregunta abierta, 0.0 si no)
        + 8.0  * bloquea           (1.0 si alguna tarea sin terminar depende de esta)
        - 5.0  * bloqueada         (1.0 si depende de alguna tarea sin terminar)
        + 12.0 * proximidad        (ver la regla siguiente)
        + 1.0  * tiene_criterios   (1.0 si tiene al menos un criterio de aceptacion)
        + 0.5  * min(edad_dias / 30, 4.0)

El resultado se redondea a un decimal, con la regla de redondeo de más abajo.
```

**La regla de `prioridad`, sin ambigüedad:**

```
N = número de niveles en priorities, en el orden configurado, del más urgente al menos urgente
i = posición del nivel de la tarea, 0-indexada desde el más urgente (0 es el nivel más urgente)

si N = 1:  prioridad = 1.0
si N > 1:  prioridad = 1 - i / (N - 1)

si la tarea no tiene ninguna prioridad asignada:  prioridad = 0.3, siempre, sea cual sea el
                                                   vocabulario y su tamaño
```

Para el vocabulario por defecto de tres niveles (`high`, `medium`, `low`), esto da exactamente
`high 1.0, medium 0.5, low 0.0`, los valores que este documento ya citaba. Para un vocabulario de
cuatro niveles, por ejemplo `p0, p1, p2, p3`, del más urgente al menos urgente, da `p0 1.0, p1
0.667, p2 0.333, p3 0.0`.

**La regla de `proximidad`, sin ambigüedad:**

```
dias = fecha_limite - hoy, en dias (puede ser negativo si la fecha ya paso)

si la tarea no tiene fecha limite:  proximidad = 0.0
si la tiene:                        proximidad = clamp((30 - dias) / 30, 0.0, 1.0)
```

Una tarea vencida tiene `dias` negativo, así que `(30 - dias) / 30` supera 1 y el resultado se acota
en **1.0**, el mismo máximo que una tarea que vence hoy. Una tarea vencida no suma más que una que
vence hoy; para distinguirlas está el filtro `--overdue` de `biso ls`, no un término sin tope en la
fórmula.

**El coeficiente de este término, `urgency.due` (12.0), es el mayor de los siete.** En un tablero
donde ninguna tarea usa `due`, `proximidad` vale siempre 0.0 en todas las tareas, así que el término
de mayor peso de la fórmula no aporta nada a la urgencia de ninguna: queda inerte sin que la fórmula
cambie ni haya que retirar nada. El 2026-09-21 se midió que `due` está en 0 de las 535 tareas de los
seis tableros de Backlog.md de esta máquina, con el flag `--due-date` (y su opuesto
`--clear-due-date`) disponible en el CLI de Backlog.md 1.52.0: es un cero de decisión, no de
imposibilidad. El campo se conserva por dos motivos: `due` está en cuatro de las siete herramientas
comparadas en ["Estado del arte"](../../estado-del-arte/compatibilidad-de-modelos.md#resumen) y es,
después del título y del estado, el campo más universal del espacio; y este cero dice más sobre estos
seis tableros, de una sola persona sin plazos externos, que sobre el campo, porque es la clase de
cero que cambia el día que el tablero lo use otra persona con un compromiso de fecha.

**El redondeo, `hoy` y las unidades de tiempo, sin ambigüedad:**

```
El redondeo a un decimal es "round half away from zero": un empate exacto en el segundo decimal se
redondea alejándose de cero, nunca hacia el par más cercano. Ejemplo: 1.45 redondea a 1.5, y 15.35
redondea a 15.4.

"hoy" es la fecha de calendario en UTC en el instante en que se calcula, nunca la zona horaria de
quien llama: el mismo tablero, leído en el mismo instante real desde dos zonas distintas, da la
misma urgencia en las dos. Ejemplo: a las 23:00 UTC del 2026-09-06, "hoy" es 2026-09-06 en
cualquier máquina, aunque localmente ya sea 2026-09-07 en una zona más al este.

edad_dias, el numerador del término de antigüedad, es un entero: la diferencia entre la fecha de
calendario de hoy y la fecha de calendario de createdAt, sin mirar la hora de ninguna de las dos.
Ejemplo: una tarea creada el 2026-08-07T23:50:00Z y leída el 2026-08-08T00:10:00Z, veinte minutos
después, tiene edad_dias = 1, no una fracción de día.
```

Con esa misma unidad de calendario, una tarea que vence hoy tiene `dias = 0`, que no es negativo:
**no es `--overdue`**. El filtro `--overdue` de [`biso ls`](../cmd/ls.md) solo marca vencidas las
tareas con `dias < 0`, el mismo umbral que ya usa `proximidad` arriba para saturar en 1.0.

Un ejemplo completo, que es el que imprime `biso get --explain-urgency` en la sección [`biso get`](../cmd/get.md): una tarea
de prioridad alta, en el estado activo, de la que depende otra tarea sin terminar, sin fecha límite,
con dos criterios y creada hoy, suma `6.0 + 4.0 + 8.0 + 0.0 + 0.0 + 1.0 + 0.0`, es decir **19.0**.

El valor de urgencia del ejemplo sale de los coeficientes por defecto, que el contrato de estabilidad
permite cambiar entre versiones menores, así que la cifra exacta puede no ser esta.

**Los coeficientes configurables son exactamente siete, bajo `urgency.`, uno por término de la
fórmula**: `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`,
`urgency.criteria` y `urgency.age`, con los valores de arriba (6.0, 4.0, 8.0, -5.0, 12.0, 1.0 y 0.5)
como valores por defecto. **El peso de cada prioridad no es una de esas siete claves configurables:
se deriva de la posición en `priorities`, nunca del nombre**, según la regla de arriba, y esa regla
en sí es parte de la estructura fija de la fórmula, que no cambia en la versión 1.0.

**Una tarea archivada sin terminar cuenta como terminada para `bloquea` y `bloqueada`.** Los dos
términos miran si "alguna tarea depende de esta" o "esta depende de alguna tarea" está **sin
terminar**, y "sin terminar" aquí significa ni en el estado terminal ni archivada: una tarea
archivada, aunque su `status` guardado no sea el terminal, no cuenta como la que bloquea ni como la
que deja bloqueada a otra. **El campo derivado `blocks` de la sección ["El modelo de datos de una
tarea"](index.md#los-campos-derivados) es la lista de identificadores que hace cierto `bloquea`**, así
que sigue la misma regla: un dependiente archivado sin terminar no sale en `blocks`, igual que no
suma al término `bloquea`. Consecuencia directa: archivar una tarea que no había terminado
desbloquea de inmediato a quien dependía de ella, en la misma lectura, sin ninguna escritura
adicional. Es una definición distinta de la que usa el aviso de "subtareas sin terminar" de [`biso
finish`](../cmd/verbos-del-ciclo.md#biso-finish), que sí sigue contando una subtarea archivada como
sin terminar; el porqué de la diferencia está en [`biso archive`](../cmd/archive.md).

**El `ordinal` no forma parte de la urgencia.** Es un orden manual que se aplica aparte, según la
regla de orden completa de la sección [`biso ls`](../cmd/ls.md), y es una clave de texto y no un
número (["El orden manual y su clave"](orden-manual.md)), así que tampoco podría entrar en una suma.
