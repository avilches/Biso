# Reparto de `SPEC.md` en documentos que se citan por su nombre

Diseñado el 2026-09-10. Este documento es el encargo del que se implementa el reparto. Quien lo
ejecute no debería tener que preguntar nada.

**Este documento cita por número a propósito**, con la numeración anterior al reparto, porque describe
el estado del que se parte. Es la única excepción a la convención que él mismo establece, y no hay que
actualizarlo cuando el reparto termine.

## El problema, medido

`docs/SPEC.md` son 5.762 líneas y 317 KB en un solo fichero, y su sección 10 ocupa 3.595 de esas
líneas, el 62 % del total. Toda la documentación del proyecto lo cita por el número de sus secciones,
en la forma "sección 10.4" o "apartado 14.1".

Esas referencias son 646, repartidas así:

| Documento | Referencias | Estado |
|---|---|---|
| `docs/SPEC.md` (a sí mismo) | 215 | vivo |
| `docs/DECISIONES.md` | 97 | vivo |
| `docs/superpowers/plans/2026-09-07-persistencia.md` | 83 | acta cerrada |
| `docs/superpowers/specs/2026-09-06-modelo-de-estados-design.md` | 74 | acta cerrada |
| `docs/superpowers/specs/2026-09-07-persistencia-design.md` | 50 | acta cerrada |
| `INTEGRATION.md` | 35 | vivo |
| `docs/superpowers/plans/2026-09-06-modelo-de-estados.md` | 30 | acta cerrada |
| `bench/sqlite-driver/RESULTADOS.md` | 24 | vivo |
| `docs/ESTADO-DEL-ARTE.md` | 11 | vivo |
| `CLAUDE.md` | 10 | vivo |
| `docs/PENDIENTES.md` | 9 | vivo |
| `bench/sqlite-driver/README.md` | 6 | vivo |
| `docs/superpowers/specs/2026-09-10-tutorial-por-escenarios-design.md` | 2 | vivo |

Ninguna de las 646 es un enlace. Son todas texto plano, así que en el sitio de MkDocs no se puede
hacer clic en ninguna y nada comprueba que sigan apuntando a donde apuntaban.

## Lo que no es el problema, y conviene saberlo antes de tocar nada

**La renumeración no ha ocurrido nunca.** Se han revisado los 89 encabezados numerados que se han
añadido o quitado en toda la historia de `SPEC.md` y los 29 de `DECISIONES.md`, buscando algún título
que se mantuviera cambiando de número. No hay ni uno. Las secciones siempre se han añadido al final de
su grupo, como `4.13` detrás de `4.12` y `10.14` detrás de `10.13`, de modo que "sección 10.4" nunca ha
llegado a apuntar a otro sitio.

Eso no significa que el problema sea imaginario. Significa que el precio se está pagando en otras dos
monedas, y son estas las que justifican el trabajo.

**Primera: la numeración ha congelado la organización.** Nadie inserta una sección en su sitio lógico
porque hacerlo obligaría a revisar 646 referencias, así que el documento solo puede crecer por el final
de cada grupo. El resultado es que `4.13` trata del presupuesto de arranque y está detrás de `4.12`
por orden de llegada y no por afinidad, y que `biso prime` es la sección 9, fuera de la sección 10 que
se llama "Los comandos", siendo un comando.

**Segunda: las frases que cuentan elementos ya han bloqueado trabajo real.** El párrafo de `biso doctor`
dice hoy: "de las dieciocho filas de la tabla, diecisiete son problemas (quince errores y dos avisos) y
una, los huecos en la numeración, no lo es. De esas diecisiete, solo dieciséis llegan a aparecer alguna
vez como una línea del informe". Son seis números encadenados en un párrafo. `docs/PENDIENTES.md`
reconoce que una mejora ya identificada de `doctor` no se hizo porque "obliga a tocar las tres frases
que cuentan las dieciocho filas de su tabla". El mismo fallo aparece en "Veinte comandos. Los once
primeros... los nueve restantes... De esos once, diez son...", en "Las once reglas que no son
adivinables", y en el "los cuatro documentos del proyecto" de `docs/index.md`, que ya hubo que retocar
en `mkdocs.yml` al aparecer un quinto.

## Las decisiones

### 1. La arquitectura de ficheros

`docs/SPEC.md` desaparece y su contenido pasa a `docs/spec/`, con un fichero por comando en
`docs/spec/cmd/` y ficheros temáticos para lo transversal.

**Ningún nombre de fichero lleva número.** Un prefijo numérico en el nombre reintroduce el mismo
problema un nivel más arriba: insertar un documento en medio obligaría a renumerar los ficheros y a
romper todos los enlaces, o a dejar el número mintiendo. El orden de lectura vive en dos sitios y solo
en dos, el `nav` de `mkdocs.yml` y `docs/spec/index.md`, que se reordenan sin tocar ningún enlace.

`docs/spec/index.md` es la portada de la especificación: dice en qué orden se lee, para qué sirve cada
pieza, y hereda el papel que hoy tiene la introducción de `SPEC.md`.

El resto del repositorio no cambia de sitio. `DECISIONES.md`, `PENDIENTES.md` y `ESTADO-DEL-ARTE.md`
siguen siendo un fichero cada uno.

### 2. Las mudanzas confirmadas

Estas cinco están decididas, todas por afinidad:

**`biso prime` se une a los comandos**, en `cmd/prime.md`, y es el primero de la lista porque es el
primero que se ejecuta. Hoy es la sección 9, fuera de "Los comandos", sin ninguna razón estructural.

**Los dos presupuestos se juntan** en `presupuestos.md`: el de arranque de 25 milisegundos (hoy 4.13) y
el de tamaño del mensaje de 5.120 bytes (hoy 9.5). Están a setecientas líneas de distancia y la sección
del contrato de estabilidad los discute juntos, porque uno se congela en el contrato y el otro no.

**La resolución del tablero se emancipa** a `resolucion-del-tablero.md`. Hoy es la sección 3.2, una
subsección de "Flags globales", que no es lo que es, y es la sección más citada de toda la
especificación con 61 referencias entrantes.

**La ayuda se reúne** en `cmd/help.md`. Hoy la sección 11 habla de `biso --help` y la 10.13 del comando
`biso help`, que son el mismo tema en dos sitios.

**La sección 4 se reparte por temas** en lugar de seguir siendo un cajón de reglas sin relación entre
sí: terminal, flujos y codificación a un fichero, las formas de pasar un valor a otro, y la
concurrencia con la atomicidad a `garantias.md`.

El criterio para todo lo demás: **un fichero agrupa lo que se lee de una sentada para responder a una
sola pregunta.** La asignación exacta de las subsecciones no nombradas aquí se decide leyéndolas durante
la mudanza, no a ciegas de antemano, y la prueba de identidad de líneas de la decisión 4 garantiza que
ninguna se quede por el camino mientras se decide.

### 3. La forma de citar

Una referencia es un enlace de Markdown cuyo texto es el título de la sección:

```markdown
la sección ["Cómo se elige el tablero"](resolucion-del-tablero.md#cómo-se-elige-el-tablero)
```

Desde otro documento de `docs/`, con la ruta relativa que corresponda. Desde `CLAUDE.md` en la raíz,
`docs/spec/...`. Dentro del mismo fichero, solo el ancla.

**Hay que configurar el slugify de las anclas, y es obligatorio.** Con la configuración actual de
`mkdocs.yml`, el título `Códigos de salida` genera el ancla `codigos-de-salida`, sin acentos, y
`por qué` genera `por-que`. Añadiendo a la extensión `toc` el slugify de `pymdownx`, que ya es
dependencia:

```yaml
markdown_extensions:
  - toc:
      permalink: true
      slugify: !!python/object/apply:pymdownx.slugs.slugify {kwds: {case: lower}}
```

las anclas conservan los acentos, quedan legibles, y coinciden con las que genera GitHub, así que un
enlace funciona igual en el sitio de MkDocs y leyendo el fichero en crudo. Verificado el 2026-09-10
con MkDocs 1.6.1 y pymdown-extensions 11.0.2.

En los dos casos, las comillas de código desaparecen del ancla y los dos puntos y las comas también:
`` ## `biso init` `` genera `biso-init`, y
`` ## Los verbos del ciclo: `start`, `note`, `comment` `` genera
`los-verbos-del-ciclo-start-note-comment`.

**La verificación.** MkDocs 1.6 valida las anclas si se le pide, y con `--strict` termina con código 1.
Hay que añadir a `mkdocs.yml`:

```yaml
validation:
  anchors: warn
  unrecognized_links: warn
```

Comprobado el 2026-09-10: un enlace a un fichero inexistente y un enlace a un ancla inexistente
producen los dos un aviso, y `mkdocs build --strict` aborta con código 1. Eso convierte una referencia
rota en un fallo de build, que es la garantía que la numeración no da.

**La salvedad que hay que decir.** `mkdocs build --strict` solo valida los enlaces de los ficheros que
están dentro de `docs/`. Los de `CLAUDE.md`, `INTEGRATION.md`, `bench/sqlite-driver/README.md` y
`bench/sqlite-driver/RESULTADOS.md` quedan fuera de esa red y necesitan un comprobador propio, que
resuelva cada enlace relativo a un fichero de `docs/` y compruebe que el ancla existe en él.

### 4. La red de seguridad: cómo se demuestra que no se pierde nada

El encargo dice "sin que se pierda nada", y eso se demuestra, no se promete. Se consigue separando la
mudanza de la reescritura en dos pasos con pruebas distintas, y el orden no es negociable.

**Paso uno: la mudanza no cambia ni una letra del contenido.** Se parte y se reordena sin tocar ninguna
referencia, que al terminar este paso siguen diciendo "sección 10.4". Eso habilita una prueba exacta:
se toman todas las líneas de contenido del `SPEC.md` viejo, se toman todas las de los ficheros nuevos, y
se comprueba que los dos conjuntos son idénticos línea a línea, cada una apareciendo el mismo número de
veces. Si sobra o falta una sola línea, el script falla y dice cuál.

Las únicas diferencias admitidas son los encabezados, que pierden su número, y la introducción de
`SPEC.md` que pasa a `docs/spec/index.md` y deja de hablar de "este documento" en singular. Cada
excepción se declara una a una en el script, con su texto viejo y su texto nuevo, para que nadie pueda
colar un cambio de contenido dentro de una mudanza.

**Paso dos: la reescritura de las referencias.** Aquí las líneas cambian por diseño, así que la prueba
es otra: `mkdocs build --strict` en verde, el comprobador de enlaces de los ficheros de fuera de `docs/`
en verde, y un recuento que confirme que no queda ni un "sección N.N" ni un "apartado N.N" fuera de los
documentos históricos congelados.

### 5. La regla contra los recuentos

**La regla: en la prosa no se escribe un número que cuente elementos que el propio documento enumera.**
Se dice "las reglas de abajo" o "cada fila de la tabla", no "siete reglas" ni "dieciocho filas".

**La excepción, que es lo que hace la regla usable: sí se escribe el número cuando el número es la regla
y hay una prueba que lo comprueba.** Los 5.120 bytes del mensaje de arranque, los 25 milisegundos del
presupuesto, el recorte del título a 100 celdas, las ocho columnas fijas de `biso ls`, el mínimo de tres
estados. La diferencia es si el número forma parte del contrato del programa o es una descripción del
contenido del documento. Lo primero se queda y no caduca porque una prueba lo sostiene. Lo segundo se va.

**El comprobador.** Un script que busca un número cardinal, en cifra o en palabra, seguido de uno de una
lista corta de sustantivos que nombran partes del documento: reglas, filas, comandos, códigos, flags,
secciones, apartados, documentos, principios, garantías, mensajes, campos, verbos, criterios, columnas,
entradas, tipos, familias, errores, avisos, estados. Hoy encuentra unas cuarenta frases. Las normativas
se declaran en una lista de excepciones indexada por el texto de la frase, no por el número de línea, para
que la lista no se invalide al editar el fichero alrededor.

El primer cliente es el párrafo de `biso doctor`, y arreglarlo desbloquea la mejora que `PENDIENTES.md`
dejó sin hacer.

### 6. Los números de las salidas de ejemplo se quedan como están

En los bloques de salida de ejemplo no se ponen placeholders. Dos razones.

La primera es que no caducan: son datos del tablero de ejemplo, que es un fixture fijo definido en la
especificación, así que un `To Do 54 | In Progress 4 | Done 190` solo cambia si alguien cambia el
fixture, y entonces cambiarlo es lo correcto.

La segunda es que romperían la alineación. El algoritmo de columnas dice que "el ancho de esa columna es
la anchura del valor más largo", así que los anchos se calculan sobre el contenido, y sustituir un `54`
por un `<N>` de otra anchura desalinea la tabla entera. Además esas salidas están declaradas como
literales y se comparan carácter a carácter: el `CLAUDE.md` del proyecto dice que se generan y no se
escriben a mano, y los fixtures del tutorial ya marcan cada una con `origen: literal SPEC <sección>`
para poder ejecutarlas contra el binario el día que exista.

**Sí se añade una nota de "este valor puede variar" donde el contrato de estabilidad ya declara que el
valor cambia**, que son los valores de urgencia, porque los coeficientes por defecto pueden cambiar
entre versiones menores, y la cadena de versión del programa.

### 7. `DECISIONES.md` se reordena, no se parte

Sus 1.374 líneas se quedan en un fichero, y adopta la misma forma de citar. Lo que cambia es el orden,
porque hoy es el de llegada y no el de afinidad: las secciones 9, 10 y 11 son todas del modelo de estados
y podrían ser una, y la 13 y la 14 hablan las dos de rendimiento. Se reordena con la misma prueba de
identidad de líneas del paso uno de la decisión 4.

**Condensar, solo donde haya dos párrafos diciendo literalmente lo mismo.** El valor de este documento
es la evidencia que guarda, y resumir evidencia es perderla. El propio `CLAUDE.md` del proyecto dice que
si una regla parece arbitraria, su razón está aquí antes de cambiarla.

Es un documento vivo, no un histórico: cambia en 27 de los 84 commits que tocan alguno de los dos
documentos grandes, 21 a la vez que `SPEC.md` y 6 por su cuenta.

### 8. Lo que queda fuera

**Los documentos de `docs/superpowers/` anteriores a este no se tocan.** Son actas fechadas de sesiones cerradas, y
editarlas a posteriori les quita su valor como registro de lo que se acordó aquel día. Llevan una nota al
principio que avisa de que sus referencias usan la numeración anterior al reparto, con la fecha del
reparto para que se pueda reconstruir a qué apuntaban.

**El tutorial a medias necesita una pasada propia al final.** En el worktree `worktree-tutorial` hay
trabajo sin commitear, y el formato de sus fixtures declara el origen de cada salida como
`origen: literal SPEC 4.3`, citando por número. Ese campo hay que reconvertirlo, y el escenario 01, que
ya está escrito, con él. Esa pasada va al final, cuando los nombres de los ficheros nuevos ya son
definitivos.

## Cómo se comprueba que el trabajo está bien

Al terminar, todas estas tienen que cumplirse a la vez:

1. El script de identidad de líneas no encuentra ninguna línea perdida ni duplicada entre el `SPEC.md`
   viejo y los ficheros de `docs/spec/`.
2. `mkdocs build --strict` termina con código 0.
3. El comprobador de enlaces de los ficheros de fuera de `docs/` termina con código 0.
4. No queda ninguna referencia de la forma "sección N.N" ni "apartado N.N" en ningún documento vivo.
5. El comprobador de recuentos termina con código 0, con su lista de excepciones normativas declarada.
6. El párrafo de `biso doctor` ya no encadena seis números, y la mejora que `PENDIENTES.md` dejó
   aplazada se puede hacer sin tocar ninguna frase que cuente.
7. Ningún fichero de `docs/spec/` pasa de unas 700 líneas.
8. Ningún nombre de fichero de `docs/spec/` empieza por un número.

## Evidencia medida el 2026-09-10, para no repetirla

- 646 referencias por número en 13 ficheros, ninguna en forma de enlace.
- `docs/SPEC.md`: 5.762 líneas, 317 KB. Sección 10: 3.595 líneas, el 62 %.
- Secciones de nivel 2 de `SPEC.md`, en líneas: vocabulario 34, principios 28, códigos de salida 46,
  flags globales 368, reglas transversales 362, modelo de datos 303, vocabularios 101, referencias 81,
  familias de flags 142, `prime` 381, comandos 3.595, ayuda 42, contrato JSON 129, estabilidad 49,
  fuera de alcance 44, por dónde empezar 26.
- Los catorce comandos de la sección 10, en líneas: `init` 423, `where` 136, `new` 220, `ls` 330,
  `get` 244, `set` 175, verbos del ciclo 673, `archive` 89, `export` 109, `config` 272, `doctor` 339,
  `board` 88, `help` 102, `snapshot` 344.
- Secciones más citadas: 3.2 con 61 referencias, luego 5 con 42, 8 y 12 con 29, 14 y 13 con 25.
- Ningún encabezado se ha renumerado nunca, en 89 cambios de encabezado en `SPEC.md` y 29 en
  `DECISIONES.md`.
- MkDocs 1.6.1 valida anclas con `validation.anchors` y aborta con código 1 bajo `--strict`.
- El slugify por defecto quita los acentos del ancla; el de `pymdownx` los conserva.
- `mkdocs build --strict` no valida los enlaces de los ficheros de fuera de `docs_dir`.
