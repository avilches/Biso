# El menú lateral del sitio pasa a tener opciones en Inicio, Tutorial, Decisiones y Estado del arte

Diseño para TASK-29 y sus subtareas TASK-29.1 y TASK-29.2.

## El problema

El tema Material de MkDocs solo muestra opciones en el menú de la izquierda dentro de un tab del nav
cuando ese tab agrupa varias páginas. Hoy solo "Especificación" lo hace: "Inicio" apunta a un único
`index.md`, "Tutorial" a un único `docs/TUTORIAL.md` generado, y "Decisiones de diseño" y "Estado del
arte" a un único fichero cada uno. Los cuatro tabs quedan con el menú vacío, aunque sus páginas sean
largas y tengan estructura interna (el tutorial tiene trece escenarios y una sección de conceptos;
`DECISIONES.md` tiene treinta apartados; `ESTADO-DEL-ARTE.md` tiene tres partes).

## Las decisiones

**El tutorial pasa a ser una página por escenario.** `docs-tooling/tutorial/generate.py` deja de
escribir un único `docs/TUTORIAL.md` y escribe `docs/tutorial/index.md` (la portada, con la lista
enlazada de los trece escenarios) y `docs/tutorial/NN-nombre.md` por cada fixture de
`tutorial/escenarios/`, con el mismo nombre de fichero que su `.yaml` de origen. Sigue fallando sin
escribir nada si algún fixture es inválido. `mkdocs.yml` lista cada página del tutorial como una
entrada de nav propia, con su título en inglés.

**Conceptos se extrae del tutorial y pasa a vivir bajo Inicio.** El contenido de
`tutorial/conceptos.md` dejaba de ser la introducción incrustada en `docs/TUTORIAL.md` y pasa a
generarse como página independiente, `docs/concepts.md`, listada bajo el tab "Inicio". La página del
tutorial (`docs/tutorial/index.md`) enlaza a ella para quien empieza por ahí sin haber leído Inicio.
Se queda en inglés: se decidió extender la excepción de idioma que hoy cubre solo
`docs/TUTORIAL.md` y los fixtures del tutorial, en vez de traducirla al español al sacarla del
generador. La razón práctica es que el contenido nace del mismo fixture en inglés
(`tutorial/conceptos.md`) que ya usa el tutorial, y traducirlo habría exigido mantener dos versiones
o bifurcar la fuente.

**Decisiones de diseño y Estado del arte se agrupan en unas pocas páginas temáticas, no una por
apartado.** Partir por cada uno de los treinta apartados de `DECISIONES.md` habría exigido reescribir
cada enlace cruzado del repositorio que hoy cita un apartado suelto dentro del fichero grande (varios,
desde `docs/spec/` y desde este mismo `CLAUDE.md`). Agrupando en 6-8 páginas por área (modelo de
estados, persistencia, lenguaje y rendimiento, comandos y banderas, formato de datos, y el resto de
decisiones de detalle) la mayoría de esos enlaces se quedan apuntando al mismo fichero, y solo hace
falta corregir los que cambien de página. `ESTADO-DEL-ARTE.md` se parte en tres, una por cada una de
sus partes actuales (las herramientas, el catálogo de problemas, y lo que la gente quiere conservar).
El reparto exacto de apartados a página es trabajo de TASK-29.2, no de este documento.

**El reparto que implementó TASK-29.2**, bajo `docs/decisiones/` y `docs/estado-del-arte/`, cada uno
con su propio `index.md` de portada:

- `principios-y-mantenimiento.md`: la evidencia de los siete principios, la advertencia sobre cómo se
  mantiene la especificación, lo que se deja fuera y por qué, y los cuatro requisitos aprendidos de
  otras herramientas.
- `modelo-de-estados.md`: los cuatro requisitos del modelo de estados con sus subapartados, el
  criterio de estado frente a campo, y los riesgos conocidos y aceptados.
- `persistencia.md`: la decisión de persistencia con sus subapartados, y el origen de la cifra de 25
  milisegundos.
- `lenguaje-y-rendimiento.md`: por qué Go, y por qué `modernc.org/sqlite` sin `cgo`.
- `vocabulario-y-mensaje-de-arranque.md`: la regla de coincidencia de vocabulario, el presupuesto del
  mensaje de arranque, y el grid completo de banderas de campo.
- `comandos-y-banderas.md`: el porqué de reglas concretas, y `<command>` repetible en `biso help`.
- `detalles.md`: el resto de decisiones de detalle que cuesta reconstruir, del juego de caracteres de
  un token a la retirada de `project` y `milestone`.
- `estado-del-arte/herramientas.md`, `catalogo-de-problemas.md` y `lo-que-se-conserva.md`: las tres
  partes actuales, sin más división.

Cada enlace cruzado que citaba un apartado de `DECISIONES.md` o `ESTADO-DEL-ARTE.md` por su ancla se
reescribió al fichero y ancla nuevos; los que citaban el documento entero sin ancla pasaron a apuntar
al `index.md` correspondiente. Los documentos de diseño y planes anteriores a esta tarea no se
reescriben en su prosa, salvo el enlace mecánico roto que había que corregir para que el sitio
siguiera construyendo.

## Lo que no cambia

- El nav de "Especificación" no se toca: ya tenía sus opciones en el menú.
- `urgency.py` y `continuity.py` operan solo sobre los fixtures YAML (`escenarios/*.yaml`,
  `tablero.yaml`), nunca sobre la página generada, así que no dependen de esta forma de salida.
- Los documentos de diseño y actas anteriores que citan `docs/TUTORIAL.md` como fichero único
  (`2026-09-10-tutorial-por-escenarios-design.md`, `2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md`,
  `2026-09-12-reorganizacion-docs-tooling-design.md`, y el plan `2026-09-10-reparto-de-la-spec.md`) no
  se reescriben: son actas de sesiones pasadas, igual que no se edita una memoria vieja.

## Comprobado antes de implementar

Ningún fichero vivo del repositorio enlazaba a `docs/TUTORIAL.md` ni a un ancla suya
(`#escenario-NN`, `#concepts`, etc.) desde fuera del propio fichero: el único uso de esas anclas era
el índice interno de la propia página, que desaparece al partirla. Por eso este cambio no necesita
una pasada de corrección de enlaces cruzados, al revés que el de Decisiones de diseño y Estado del
arte.
