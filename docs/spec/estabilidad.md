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
- La simetría entre `biso export` y `biso new --from` sobre todos los campos no derivados, que es una
  prueba de la suite y no una intención. La de `biso snapshot` con `biso init --from` cubre además la
  configuración del tablero entera, campo a campo.
- La estabilidad de las claves de los criterios: una clave asignada no se reasigna nunca.
- El tope de tamaño del mensaje de `biso prime` (sección ["El presupuesto de tamaño"](presupuestos.md#el-presupuesto-de-tamaño)).

**Puede cambiar entre versiones menores, y por eso no hay que analizarlo:**

- El texto exacto de los mensajes de error y de los avisos. Lo estable es el `code`, no la prosa.
- La disposición de las columnas de `biso ls` y de la ficha de `biso get`, y para eso está `--json`.
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
el formato de `export` es el de `new --from` y los dos están en este contrato.

---

