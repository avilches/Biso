# La urgencia

`urgency` es un `float` derivado que se recalcula en cada lectura y **nunca se guarda**. Es el segundo
criterio de la tupla de orden por defecto de `biso ls`, después de `ordinal` ([`biso ls`](../cmd/ls.md)), y el que ordena
el resumen de `biso prime`.

```
Si el estado de la tarea es el terminal, urgency = 0.0 y no se calcula nada mas.

En cualquier otro caso:

urgency = 6.0  * prioridad         (high 1.0, medium 0.5, low 0.0, sin prioridad 0.3)
        + 4.0  * activa            (1.0 si el estado es el activo y no hay pregunta abierta, 0.0 si no)
        + 8.0  * bloquea           (1.0 si alguna tarea sin terminar depende de esta)
        - 5.0  * bloqueada         (1.0 si depende de alguna tarea sin terminar)
        + 12.0 * proximidad        (ver la regla siguiente)
        + 1.0  * tiene_criterios   (1.0 si tiene al menos un criterio de aceptacion)
        + 0.5  * min(edad_dias / 30, 4.0)

El resultado se redondea a un decimal.
```

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

Un ejemplo completo, que es el que imprime `biso get --explain-urgency` en la sección [`biso get`](../cmd/get.md): una tarea
de prioridad alta, en el estado activo, de la que depende otra tarea sin terminar, sin fecha límite,
con dos criterios y creada hoy, suma `6.0 + 4.0 + 8.0 + 0.0 + 0.0 + 1.0 + 0.0`, es decir **19.0**.

El valor de urgencia del ejemplo sale de los coeficientes por defecto, que el contrato de estabilidad
permite cambiar entre versiones menores, así que la cifra exacta puede no ser esta.

**Los coeficientes configurables son exactamente siete, bajo `urgency.`, uno por término de la
fórmula**: `urgency.priority`, `urgency.active`, `urgency.blocking`, `urgency.blocked`, `urgency.due`,
`urgency.criteria` y `urgency.age`, con los valores de arriba (6.0, 4.0, 8.0, -5.0, 12.0, 1.0 y 0.5)
como valores por defecto. **Los pesos por prioridad no son configurables**: `high 1.0, medium 0.5,
low 0.0, sin prioridad 0.3` son parte de la estructura fija de la fórmula, que no cambia en la
versión 1.0.

**El `ordinal` no forma parte de la urgencia.** Es un orden manual que se aplica aparte, según la
regla de orden completa de la sección [`biso ls`](../cmd/ls.md).
