# El contrato de estabilidad

Lo que se promete mientras la versión mayor sea `1`. **Obliga a partir de la versión 1.0**, que
todavía no está publicada: hasta que salga, nada de lo de abajo está roto por cambiar.

**No cambia nunca:**

- Los ["Códigos de salida"](codigos-de-salida.md) y su significado.
- Los identificadores `code` de la sección ["Los identificadores de error"](contrato-json.md#los-identificadores-de-error), con la regla de
  ampliación que allí se dice.
- El nombre y el significado de cada comando y de cada flag. **Un flag nunca cambia de
  semántica**, y en particular ninguna que hoy añade pasará a reemplazar. Si hiciera falta el
  comportamiento contrario, se añade un flag nuevo con otro nombre.
- Las claves de `data` en cada `kind` de JSON. Se pueden añadir claves; las que hay no se quitan ni
  cambian de tipo.
- El ["algoritmo de coincidencia"](vocabularios.md#el-algoritmo-de-coincidencia), idéntico al leer y al escribir.
- La regla de análisis de una etiqueta con ámbito (["Las etiquetas con ámbito"](valores-de-entrada.md#las-etiquetas-con-ámbito)):
  dónde corta la clave, qué separadores hay y qué formas son mal formadas. De ella dependen el
  significado de `clave::valor`, la consulta `--label clave:` y lo que la lista `labels` puede
  declarar, así que cambiarla cambiaría en silencio qué encuentra un filtro ya escrito.
- La simetría entre `biso export` y `biso new --from` sobre todos los campos no derivados, que es una
  prueba de la suite y no una intención. La de `biso snapshot` con `biso init --from` cubre además la
  configuración del tablero entera, campo a campo.
- La estabilidad de las claves de los criterios: una clave asignada no se reasigna nunca.
- **La forma de una clave de orden manual y cómo se comparan dos claves** (["El orden manual y su
  clave"](modelo-de-datos/orden-manual.md)): el alfabeto `0-9a-z`, la prohibición del `0` final y la
  comparación por puntos de código. Es contrato porque la clave viaja en el JSON y en la exportación,
  y porque quien lee dos claves tiene derecho a saber cuál va antes sin preguntárselo al programa.
- El tope de tamaño del mensaje de `biso prime` (sección ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño)).

**Puede cambiar entre versiones menores, y por eso no hay que analizarlo:**

- El texto exacto de los mensajes de error y de los avisos. Lo estable es el `code`, no la prosa.
- La disposición de las columnas de `biso ls` y de la ficha de `biso get`, y para eso está `--json`.
- **Qué clave exacta elige el algoritmo del punto medio dentro de un hueco**
  (["El algoritmo del punto medio"](modelo-de-datos/orden-manual.md#el-algoritmo-del-punto-medio)).
  Está especificado paso a paso y con ejemplos, y una prueba los comprueba, por la misma razón por la
  que están escritos los mensajes de error: para que la implementación sea una sola y no haya que
  adivinarla. Lo que se promete a quien llama es el sitio donde cae la tarea, no la cadena que se
  guarda, así que una versión menor puede elegir otra clave del mismo hueco. Lo que no cambia es su
  forma ni su comparación, que están arriba.
- El texto de `biso prime`, dentro de su tope, que es donde se espera que la herramienta más aprenda
  con el tiempo.
- Los coeficientes por defecto de la urgencia. La estructura de la fórmula, no.
- Los valores por defecto de la configuración, salvo los que este documento fija dentro de un comando.
- **El [presupuesto de arranque](presupuestos.md#el-presupuesto-de-arranque) de 25 milisegundos.** No es de la misma naturaleza
  que el tope de bytes de arriba: los 5.504 bytes son una propiedad del texto, así que cualquiera los
  mide y siempre dan lo mismo, mientras que los 25 milisegundos son una propiedad de la máquina de
  referencia. Congelar en este contrato un número que depende del hardware haría que la herramienta
  incumpliera su propia promesa al ejecutarse en un ordenador más lento, sin que nadie hubiera cambiado
  una línea de código. Sigue siendo una prueba de la suite y sigue teniendo que fallar ante una
  regresión real: lo que no es, es una promesa de versión a versión.

**Cómo se anuncia una retirada.** Nada se quita sin un ciclo completo de aviso: primero la
funcionalidad emite `warning: <x> is deprecated and will be removed in 2.0` durante al menos una
versión menor, y solo entonces desaparece. Una funcionalidad nunca desaparece en silencio entre dos
versiones.

**Migración.** Si cambiara la forma en que los datos se guardan, la herramienta migra sola al
detectarlo, y en cualquier caso el volcado de una versión se puede importar en la siguiente, porque
el formato de `export` lo lee `new --from` entero y los dos están en este contrato.

---

