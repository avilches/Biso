# Reorganización de la documentación y su tooling en `docs-tooling/`

> Diseño para TASK-18. Antes de esta reorganización, `docs/`, `tutorial/`, `tools/`, `mkdocs.yml`
> y `docs-requirements.txt` eran cinco cosas sueltas en la raíz del repositorio, sin que se
> distinguiera a simple vista qué es contenido editado a mano, qué es generado y qué son los
> scripts que lo generan o lo verifican.

## El problema

La raíz del repositorio mezclaba:

- Documentación fuente, editada a mano: `docs/spec/`, `docs/DECISIONES.md`,
  `docs/ESTADO-DEL-ARTE.md`, `docs/PENDIENTES.md`, `docs/index.md`.
- Documentación generada: `docs/TUTORIAL.md` (por `tutorial/generate.py`), `site/` (por MkDocs).
- El generador del tutorial y sus fixtures: `tutorial/`.
- Los scripts que verifican la documentación: `tools/`, con su propia carpeta de pruebas y,
  mezclados con scripts todavía vigentes, ficheros de una migración ya terminada (el reparto de
  `docs/SPEC.md`, un fichero que ya no existe).
- La configuración del motor que renderiza `docs/`: `mkdocs.yml` y `docs-requirements.txt`,
  sueltos en la raíz.
- Un documento de análisis sin decisión, `INTEGRATION.md`, suelto en la raíz junto a `CLAUDE.md`.

`docs/` no se puede anidar bajo ninguna carpeta paraguas porque la skill de `superpowers` escribe
siempre en la ruta literal `docs/superpowers/specs/` y `docs/superpowers/plans/`, sin que se le
pueda indicar otra.

## La disposición final

```
biso/
├── CLAUDE.md                              (comandos actualizados)
├── bench/                                 (intacto: evidencia de una decisión de implementación,
│                                            no documentación; convivirá con cmd/ e internal/
│                                            cuando exista el código Go)
├── site/                                  (generado, sigue en la raíz, gitignored)
├── docs/
│   ├── index.md
│   ├── DECISIONES.md
│   ├── ESTADO-DEL-ARTE.md
│   ├── PENDIENTES.md
│   ├── TUTORIAL.md                        (generado)
│   ├── spec/                              (34 ficheros, intacto)
│   ├── superpowers/
│   │   ├── plans/                         (intacto)
│   │   └── specs/
│   │       ├── ...(los seis existentes, intactos)
│   │       └── 2026-09-07-integracion-superpowers-design.md   <- INTEGRATION.md
│   └── docs-tooling/
│       ├── mkdocs/
│       │   ├── mkdocs.yml                 ← raíz
│       │   └── docs-requirements.txt      ← raíz
│       ├── tutorial/                      ← raíz, completo
│       │   ├── CLAUDE.md
│       │   ├── conceptos.md
│       │   ├── continuity.py
│       │   ├── urgency.py
│       │   ├── generate.py
│       │   ├── tablero.yaml
│       │   ├── escenarios/                (13 .yaml)
│       │   └── lagunas/                   (4 .md)
│       └── tools/                         ← raíz, sin los ficheros muertos (ver abajo)
│           ├── doctor.py                  (nuevo)
│           ├── comprobar_enlaces.py
│           ├── comprobar_recuentos.py
│           ├── comprobar_referencias.py
│           ├── requirements.txt
│           ├── mapa-de-secciones.txt
│           ├── recuentos-normativos.txt
│           └── tests/
│               ├── test_comprobar_enlaces.py
│               └── test_comprobar_recuentos.py
```

## Lo que se borra

Tres scripts y ficheros de datos de la migración de `docs/SPEC.md` a `docs/spec/`, ya terminada,
que no se pueden volver a ejecutar porque `docs/SPEC.md` ya no existe:

- `tools/generar_mapa.py` (lee `docs/SPEC.md` directamente)
- `tools/manifiesto-del-reparto.txt` (rangos de línea de ese fichero borrado)
- `tools/excepciones-de-encabezado.txt` (excepciones de esa comparación concreta)

Y un cuarto, reusable en abstracto pero sin ningún uso vivo hoy:

- `tools/verificar_mudanza.py` y su prueba `tools/tests/test_verificar_mudanza.py`

`tools/mapa-de-secciones.txt` **no se borra**: lo sigue consumiendo `comprobar_referencias.py` y lo
citan por nombre `tutorial/tablero.yaml` y los cuatro ficheros de `tutorial/lagunas/`, como parte
del contenido del propio tutorial.

## El nuevo `doctor.py`

`docs/docs-tooling/tools/doctor.py` ejecuta, en este orden, todo el pipeline de comprobación y
generación de la documentación, y no se detiene en el primer fallo: acumula el resultado de cada
paso y termina con código distinto de cero si alguno falló, imprimiendo un resumen final, igual
que hace `biso doctor` con las comprobaciones del programa.

1. `comprobar_enlaces.py` sobre `CLAUDE.md`, `bench/sqlite-driver/README.md` y
   `bench/sqlite-driver/RESULTADOS.md` (los `.md` de fuera de `docs/`; ya no hace falta pasarle
   `INTEGRATION.md`, que se muda dentro de `docs/`).
2. `comprobar_recuentos.py` sobre `docs/spec/*.md`, `docs/spec/cmd/*.md` y `docs/*.md`.
3. `comprobar_referencias.py` sobre `docs/spec/*.md` y `docs/spec/cmd/*.md`.
4. Regenera `docs/TUTORIAL.md` ejecutando `tutorial/generate.py`.
5. `mkdocs build --strict` con `docs/docs-tooling/mkdocs/mkdocs.yml`.

Se invoca así, porque el último paso necesita MkDocs instalado:

```
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
    python docs/docs-tooling/tools/doctor.py
```

## Cambios mecánicos que exige la mudanza

- **`mkdocs.yml`**: añadir `docs_dir: ../..` (para seguir apuntando a `docs/` desde su nueva
  ubicación, dos niveles más arriba), `site_dir: ../../../site` (para que el sitio generado siga
  apareciendo en la raíz) y `exclude_docs: | docs-tooling/` (para que MkDocs no trate los scripts,
  los `.yaml` de los escenarios ni los `.pyc` como contenido a publicar). La exclusión existente
  `not_in_nav: | superpowers/` no cambia.
- **`CLAUDE.md`** (raíz) y **`tutorial/CLAUDE.md`**: cada comando `uv run --with-requirements
  docs-requirements.txt --no-project mkdocs ...` pasa a usar
  `docs/docs-tooling/mkdocs/docs-requirements.txt` y añade `-f docs/docs-tooling/mkdocs/mkdocs.yml`;
  las llamadas a `tutorial/generate.py`, `tutorial/urgency.py` y `tutorial/continuity.py` pasan a
  `docs/docs-tooling/tutorial/...`; la llamada a `pytest tools/tests` pasa a
  `docs/docs-tooling/tools/tests`.
- **`tutorial/generate.py`**: la cabecera que escribe al principio de `docs/TUTORIAL.md` (el aviso
  de "no editar a mano, regenerar con...") cita ese mismo comando y hay que actualizarla ahí.
- **`comprobar_referencias.py`**: su constante `MAPA_PATH = Path("tools/mapa-de-secciones.txt")`
  pasa a `Path("docs/docs-tooling/tools/mapa-de-secciones.txt")` (sigue asumiendo que se invoca
  desde la raíz del repositorio, como hoy).
- **`comprobar_recuentos.py`**: ya resuelve `recuentos-normativos.txt` con
  `Path(__file__).resolve().parent`, así que no necesita ningún cambio: sigue encontrándolo junto
  a sí mismo tras la mudanza.

## Lo que no se toca

- `docs/superpowers/plans/` y `docs/superpowers/specs/` existentes son actas de sesiones pasadas
  y citan rutas como estaban entonces (`tools/mapa-de-secciones.txt`, `tools/verificar_mudanza.py`,
  etc.); no se reescriben, igual que no se edita una memoria vieja.
- `docs/PENDIENTES.md` solo nombra ficheros en prosa, sin comandos que dependan de una ruta; no
  necesita cambios.
- `bench/` no se mueve: no es tooling de documentación, es el banco de pruebas de una decisión de
  implementación, y en el futuro convivirá con el código Go real.

## Pruebas

Después de mover, `docs/docs-tooling/tools/tests/test_comprobar_enlaces.py` y
`test_comprobar_recuentos.py` tienen que seguir pasando desde su nueva ubicación:

```
uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project \
    pytest docs/docs-tooling/tools/tests -q
```

Y el propio `doctor.py` es la prueba de integración de todo el pipeline: si termina con código 0,
la mudanza no ha roto nada.
