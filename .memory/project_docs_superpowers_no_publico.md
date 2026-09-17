---
name: docs-superpowers-no-publico
description: docs/superpowers/ (actas y planes de diseño) nunca debe construirse en el sitio ni enlazarse desde una página pública
metadata:
  node_type: memory
  type: project
---

`docs/superpowers/` guarda las actas y los planes de las sesiones de diseño internas de este
proyecto. Son documentos de trabajo, no la especificación vigente, y aunque el repositorio esté en
GitHub no deben quedar accesibles como parte del sitio publicado.

Hasta el 2026-09-17 `mkdocs.yml` solo los sacaba del menú lateral (`not_in_nav`) y del buscador (un
hook dedicado, `excluir_superpowers_del_buscador.py`), pero sí los construía como páginas del
sitio, y `docs/estado-del-arte/index.md` enlazaba a uno de ellos
(`superpowers/specs/2026-09-07-persistencia-design.md`). Se corrigió moviendo `superpowers/` a
`exclude_docs` en `mkdocs.yml`, igual que ya estaba `docs-tooling/`, lo que además dejó sin uso el
hook de exclusión del buscador, que se borró. El enlace de `estado-del-arte/index.md` se cambió por
uno a la página pública que recoge esa misma decisión (`docs/decisiones/persistencia.md`).

Si se añade contenido nuevo a `docs/superpowers/` no hace falta marcarlo de ningún modo especial: la
exclusión es de todo el directorio. Lo que sí hay que evitar es volver a enlazar a un fichero de
`superpowers/` desde una página que sí se publica (`docs/spec/`, `docs/decisiones/`,
`docs/estado-del-arte/`, `docs/tutorial/`, `docs/index.md`, `docs/concepts.md`): si hace falta citar
una decisión de diseño, se enlaza a la página pública que la recoge, no al acta original. Los
enlaces entre `docs/docs-tooling/` y `docs/superpowers/` sí son legítimos, porque los dos quedan
fuera de la construcción del sitio (como los tres que hoy tiene
`docs/docs-tooling/tutorial/CLAUDE.md`).
