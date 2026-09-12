# El estado del trabajo en curso

Este documento es volátil: dice qué se está haciendo ahora mismo, qué sigue después y en qué orden, y qué
conviene tener en cuenta mientras dura el desarrollo. No es documentación del proyecto ni un registro de
decisiones: eso vive en [`docs/DECISIONES.md`](DECISIONES.md) o en el fichero de `docs/` que corresponda.
Cuando no quede nada por hacer, este fichero se borra.

## AHORA EN CURSO

**Revisar y mezclar el PR de la rama `worktree-tutorial`.** Ya está hecho: el campo `origen` de los
fixtures del tutorial se separó en `source_kind` (`literal` o `derived`) y `source`, una lista de citas
por fichero y título, según el diseño de
[`docs/superpowers/specs/2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md`](superpowers/specs/2026-09-11-origen-combinado-y-tutorial-en-ingles-design.md),
y el tutorial entero (nombres de campos, contenido de los trece escenarios, `tutorial/conceptos.md` y las
cadenas de página de `tutorial/generate.py`) nació en inglés. Los comandos de `tutorial/CLAUDE.md` y
el build estricto de MkDocs pasan. Falta la revisión y la mezcla del PR.

## SIGUIENTE

1. **Repasar y retomar el plan del almacén.** El plan del primer paso de implementación (el almacén
   SQLite, sus migraciones, la transacción `WithTx` y la primera medida real del presupuesto de arranque)
   está escrito en la rama `worktree-implementacion-almacen`, en
   `docs/superpowers/plans/2026-09-11-almacen.md`, junto con el documento de arquitectura que lo motiva en
   `docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md`. Los dos citan `docs/SPEC.md`
   por número de sección, y ese fichero ya no existe: hay que releerlos y actualizar sus referencias a los
   ficheros y anclas de `docs/spec/` antes de ejecutarlo. Ya no está bloqueado una vez se mezcle el PR de
   `worktree-tutorial`.

2. **Añadir a `biso doctor` la comprobación del fichero de exclusión.** ["`biso init`"](spec/cmd/init.md)
   declara que si alguien cambia la clave `vcs` después de crear el tablero, el fichero de exclusión se
   queda con el nombre del sistema anterior y hay que arreglarlo a mano. `biso doctor` podría comprobar
   ese desajuste y no lo hace; añadirlo implica tocar la tabla de comprobaciones de
   [`biso doctor`](spec/cmd/doctor.md).

3. **Fijar el sistema operativo de referencia del presupuesto de arranque.**
   ["El presupuesto de arranque"](spec/presupuestos.md#el-presupuesto-de-arranque) lo amarra a "la máquina
   que ejecuta la suite de integración continua", pero el margen medido cambia mucho según la plataforma
   (en macOS el suelo de arranque es de 8,1 ms, en Linux de 0,37 ms). Cuando exista integración continua,
   hay que fijar esa plataforma en el documento o decir explícitamente que da igual y por qué.

4. **Escribir la receta de otro sistema de control de versiones distinto de `git` (opcional).** El
   catálogo de sistemas de control de versiones solo tiene a `git` documentado: cuál es la raíz de un
   repositorio, si un directorio está dentro de uno y si ignora una carpeta. Sin esa receta, Mercurial,
   Jujutsu o Subversion pasan por `custom` y pierden el identificador de la revisión. No bloquea la
   versión 1.0.

5. **Diseñar la exportación al formato de Backlog.md.** Es para que quien ya usa Backlog.md pueda probar
   `biso` sin salto al vacío, no para uso propio. Ese formato tiene bugs abiertos que se heredarían, entre
   ellos que al editar una tarea pierde las claves de frontmatter que no conoce. La compatibilidad con
   Backlog.md que declara [`docs/spec/index.md`](spec/index.md) es de modelo de datos, no de formato de
   fichero.

6. **Diseñar la interfaz multiproyecto.** Fuera de alcance a propósito y con su propio brainstorming
   pendiente. De ella solo se recogió el requisito que afecta al almacenamiento: los tableros se enumeran
   por convención, con una raíz por defecto más una lista explícita de raíces adicionales, sin ningún
   registro que se actualice solo.

## TENER EN CUENTA

- Ramas abiertas ahora mismo: `worktree-tutorial` (ver AHORA EN CURSO) y
  `worktree-implementacion-almacen` (el plan del almacén de SIGUIENTE #1, con referencias a `docs/SPEC.md`
  todavía sin actualizar).
- El mensaje del commit `bd0c874` describe cuatro cambios que en realidad entraron en `aaa5b96`. El árbol
  es correcto, solo el mensaje se adelantó. No se enmienda sin que se pida explícitamente reescribir
  historia.
