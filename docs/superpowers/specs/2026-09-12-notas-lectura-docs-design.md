# Notas de lectura sobre la documentación

Fecha: 2026-09-12

## El problema

No hay manera de dejar constancia, mientras se navega el sitio de MkDocs en local, de que
una sección concreta necesita un cambio o merece una nota, sin interrumpir la lectura para
abrir un editor y buscar el fichero fuente correspondiente. Las notas se pierden o se
posponen indefinidamente porque anotarlas cuesta más que el propio hallazgo.

## Alcance

Solo para uso personal, en la máquina donde se corre `make docs-serve`. No hay que compartir
la anotación con nadie más ni publicarla en ningún sitio: por eso queda descartada de raíz
cualquier solución que pase por un servidor, una base de datos o un enlace público.

## Arquitectura

Dos piezas nuevas, ninguna con estado propio en tiempo de ejecución:

1. Un botón flotante, inyectado solo durante `mkdocs serve`, que copia al portapapeles una
   línea de nota lista para pegar.
2. Un fichero de texto versionado, `docs/docs-tooling/notas-pendientes.md`, donde esas líneas
   se van acumulando a mano.

No hay hook de Python que reciba texto del navegador ni que escriba en disco: el navegador
nunca toca el sistema de ficheros directamente, solo el portapapeles. Quien pega la línea en
el fichero, y quien la borra o la marca al procesarla, es siempre una persona o un agente con
acceso al repositorio, nunca el propio sitio servido.

## Por qué solo en `docs-serve`, nunca en `docs-build`

Lo que `mkdocs build --strict` genera en `site/` es, en potencial, lo único que podría acabar
publicado en algún sitio en el futuro (hoy no hay ningún paso de despliegue, pero nada lo
impide). El botón de notas no debe formar parte de ese artefacto bajo ninguna circunstancia,
así que la decisión de incluirlo no puede vivir en `mkdocs.yml` (que `build` y `serve` leen
por igual): tiene que decidirla código que sepa distinguir un comando de otro.

La solución es una variable de entorno, `BISO_DOCS_SERVE`, que exporta únicamente el target
`docs-serve` del `Makefile` antes de invocar `mkdocs serve`. Un hook nuevo de MkDocs,
`docs/docs-tooling/mkdocs/notas_lectura.py`, comprueba esa variable en `on_config` y solo
entonces añade las entradas correspondientes a `config.extra_javascript` y
`config.extra_css`. Seguirá el mismo patrón que el hook existente,
`excluir_superpowers_del_buscador.py`, ya registrado en la clave `hooks:` de `mkdocs.yml`.

`docs-build` y `docs-doctor` no tocan esa variable, así que el sitio que producen es
exactamente el de hoy: sin el botón, sin el JS ni el CSS asociados.

## El fichero de notas

`docs/docs-tooling/notas-pendientes.md`. Vive bajo `docs-tooling/`, que ya está excluido de
la construcción del sitio por la clave `exclude_docs` de `mkdocs.yml`: nunca se publica ni
entra en el buscador, se versiona en git como cualquier otro fichero del repositorio.

Una línea por nota, con este formato:

```
- [ ] <ruta>[#<ancla>][: «<cita>»] | <nota>
```

- `<ruta>` es `location.pathname` de la página donde se pulsó el botón (la URL que sirve
  MkDocs, no la ruta del fichero fuente; ambas coinciden en estructura porque `docs_dir`
  apunta a la raíz de `docs/`).
- `<ancla>` es opcional: el id de la sección más cercana, tomado del enlace activo en la
  tabla de contenidos lateral en el momento de pulsar el botón. Si no hay ninguno resaltado
  (por ejemplo, estás antes del primer encabezado), se omite.
- `«<cita>»` es opcional: el texto del elemento marcado con el picker (ver más abajo), con los
  saltos de línea colapsados a espacios. Si no se marcó ningún elemento, se omite junto con
  los dos puntos que la introducen.
- `<nota>` es el texto libre que se escribe en el panel antes de copiar.

Ejemplos:

```
- [ ] /spec/cmd/ls/#ejemplos: «el ejemplo de columnas asume ancho fijo» | no explica qué pasa si el título mide más que la columna
- [ ] /spec/cmd/prime/ | falta un ejemplo con criterios vacíos
```

El separador es `|`, nunca em-dash. Es plano y sin más estado que la casilla de la lista de
tareas: no es una tarea formal de Backlog.md (no lleva milestone ni pasa por el flujo de
`backlog task create`), es una bandeja de entrada de lectura. Si una nota concreta merece
convertirse en tarea, eso es un paso aparte y explícito.

## El botón y el picker de elementos

Botón circular flotante, esquina inferior derecha, con un icono de nota, con estilo acorde a
la paleta de Material (variables de color de `mkdocs-material`, para que funcione en modo
claro y oscuro sin duplicar la paleta a mano).

Al pulsarlo entra en modo de selección, igual que el inspector de elementos de las
herramientas de desarrollo del navegador: el cursor cambia, y al mover el ratón por la página
se resalta con un contorno superpuesto el bloque de contenido más cercano bajo el puntero
(párrafo, ítem de lista, encabezado, bloque de código, fila de tabla, imagen, cita en bloque;
la lista concreta de selectores CSS que cuentan como "bloque" vive en el propio script). El
resaltado se recalcula tanto al mover el ratón como al hacer scroll mientras el modo sigue
activo, para no quedarse desalineado.

Un clic sobre el bloque resaltado lo fija como el elemento de la nota y abre el panel; mientras
el modo de selección está activo, los clics no disparan su comportamiento normal (no se sigue
ningún enlace, no se activa el botón de copiar código de un bloque `pre`), así que hace falta
interceptarlos en fase de captura. La tecla Escape, o pulsar otra vez el botón flotante, cancela
el modo de selección sin abrir el panel.

El panel muestra:

- La referencia detectada: `location.pathname` más el ancla de la sección activa (leída de
  `.md-nav--secondary .md-nav__link--active`, el marcador que ya pinta
  `navigation.tracking`, activo en `mkdocs.yml`; no hace falta reimplementar un scrollspy).
- La cita: el `textContent` del elemento marcado, con espacios y saltos de línea colapsados,
  en un campo editable por si hace falta recortarla.
- Un `<textarea>` para la nota.
- Un botón "Copiar" que compone la línea final con el formato de arriba y la copia con
  `navigator.clipboard.writeText`. La confirmación es visual dentro del propio panel (cambia
  el texto del botón un instante), nunca un `alert` ni ningún otro diálogo nativo bloqueante.

Como `navigation.instant` está activo en el tema, la página nunca se recarga al navegar entre
secciones del sitio; el script se registra sobre el observable `document$` que expone
Material (el patrón documentado por el propio tema para este caso), no sobre
`DOMContentLoaded`, para seguir funcionando después de cada navegación instantánea.

## El procesado

Manual, sin comando ni skill nuevos: se abre `docs/docs-tooling/notas-pendientes.md`, y cada
línea se resuelve a mano o pidiéndole a un agente de Claude Code que la procese (lee la
referencia, sigue la ruta hasta el `.md` correspondiente, aplica o descarta el cambio). Al
resolver una nota se marca la casilla o se borra la línea directamente; no hay un estado
intermedio de "procesada pero no borrada".

## Ficheros que toca esta spec

- `docs/docs-tooling/mkdocs/mkdocs.yml`: registra el nuevo hook en la clave `hooks:`.
- `docs/docs-tooling/mkdocs/notas_lectura.py`: hook nuevo, en inglés (comentarios incluidos),
  seleccionado según `BISO_DOCS_SERVE`, añade `extra_javascript`/`extra_css`.
- `docs/docs-tooling/mkdocs/notas_lectura.js`, `notas_lectura.css`: el botón, el panel y su
  estilo. Código en inglés.
- `Makefile`: el target `docs-serve` exporta `BISO_DOCS_SERVE=1` antes de invocar
  `mkdocs serve`.
- `docs/docs-tooling/notas-pendientes.md`: fichero nuevo, vacío al crearlo, con un comentario
  de una línea arriba explicando su formato para quien lo abra sin contexto.

## Fuera de alcance

- Persistir el resaltado del elemento marcado más allá del momento de crear la nota (el
  contorno del picker desaparece al abrir el panel o al cancelar; no queda ninguna marca en
  la página después).
- Subir por el árbol del DOM con las flechas para afinar la selección, como sí permite el
  inspector de DevTools: se marca directamente el bloque de contenido más cercano bajo el
  cursor, sin paso intermedio.
- Convertir automáticamente cada nota en una tarea de Backlog.md.
- Cualquier forma de publicar o compartir las notas fuera de este repositorio.
