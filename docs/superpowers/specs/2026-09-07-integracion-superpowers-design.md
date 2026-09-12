# Integrar `biso` con las skills de Superpowers

> **Estado: análisis, no decisión.** Este documento recoge entero el análisis de una sesión del
> 2026-09-07 que se iba a perder. Nada de lo que hay aquí está implementado, ni añadido a
> `docs/spec/`, ni acordado. La decisión que lo cierra está al final, en "La decisión abierta".
>
> Se escribe con el detalle suficiente para que una sesión fría lo retome sin volver a investigar
> nada: cada afirmación sobre Superpowers lleva el fichero de donde sale, y cada afirmación sobre
> `biso` lleva su sección de `docs/spec/`.

## Índice

1. [El problema que se estaba resolviendo](#el-problema-que-se-estaba-resolviendo)
2. [El material de partida: qué es Superpowers y qué hay instalado](#el-material-de-partida-qué-es-superpowers-y-qué-hay-instalado)
3. [Cómo se activa una skill sin que nadie la invoque](#cómo-se-activa-una-skill-sin-que-nadie-la-invoque)
4. [Dos hallazgos que condicionan todo el diseño](#dos-hallazgos-que-condicionan-todo-el-diseño)
5. [Problema 1: la tarea como registro duradero](#problema-1-la-tarea-como-registro-duradero)
6. [Problema 2: el despacho automático](#problema-2-el-despacho-automático)
7. [Problema 3: por dónde se le dice esto al agente](#problema-3-por-dónde-se-le-dice-esto-al-agente)
8. [La decisión abierta](#la-decisión-abierta)
9. [Riesgos y cabos sueltos](#riesgos-y-cabos-sueltos)
10. [Si se aprueba, qué hay que tocar](#si-se-aprueba-qué-hay-que-tocar)
11. [Referencias](#referencias)

---

## El problema que se estaba resolviendo

Tres preguntas, planteadas en este orden.

**Primera, y es la que duele.** Las sesiones se llenan y hay que cerrarlas con un documento de
handoff, escrito por una skill propia (`~/.claude/skills/handoff/SKILL.md`). Eso no gusta. Lo ideal
sería hacer un brainstorming de algo, que de ahí salga una tarea, y que a esa tarea se le vayan
añadiendo cosas según pasan: la spec, el plan, las decisiones que se toman. Si la sesión se corta,
la siguiente continúa solo con la información de la tarea. La duda concreta era si tiene sentido
guardar además las rutas de los ficheros involucrados. El objetivo declarado es dejar de depender
tanto de compactar y de los handoffs.

**Segunda.** Cuando exista un despacho automático de tareas (todavía no existe), se quiere que las
skills de Superpowers se activen solas. La sospecha de partida era que eso no tiene nada que ver
con `biso` sino con el arnés que revisa las tareas pendientes y levanta al agente.

**Tercera, que es la que da título a este documento.** Analizar cómo funcionan las skills de
Superpowers para que, cuando se ejecuten, actualicen el estado de la tarea en `biso`. Como no se
pueden modificar skills de terceros, la vía que se intuía era sugerirlo desde `biso prime`.

---

## El material de partida: qué es Superpowers y qué hay instalado

### De dónde viene y dónde está

Superpowers no se instala desde el repositorio de su autor directamente. En esta máquina está
declarado en `~/.claude/settings.json` como `superpowers@claude-plugins-official`, o sea que lo
sirve el marketplace oficial de Anthropic (`anthropics/claude-plugins-official`), que a su vez
apunta al repositorio original y lo fija a un commit concreto:

```json
{
  "name": "superpowers",
  "source": { "source": "url", "url": "https://github.com/obra/superpowers.git",
              "sha": "b36e0829c6d0140e93cfef2ca599b1b07d4a7797" },
  "homepage": "https://github.com/obra/superpowers.git"
}
```

La copia descargada vive en
`~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/`, ocupa 2,1 MB, y de todo eso
Claude Code solo consume tres piezas: `skills/` (las catorce skills, 472 KB),
`hooks/hooks.json` con su script `hooks/session-start`, y `.claude-plugin/plugin.json` con los
metadatos. El resto (`docs/`, `tests/`, `scripts/` y los manifiestos para Codex, Cursor, Gemini,
Kimi, OpenCode, Devin, Hermes y Pi) viaja en la caché pero no se lee nunca desde Claude Code.

**Consecuencia práctica para este diseño:** la versión es 6.3.0 y está fijada por el marketplace, no
por nosotros. Superpowers reorganiza sus skills entre versiones mayores, así que cualquier acuerdo
que dependa de un nombre de fichero interno suyo es frágil.

### El único hook

`hooks/hooks.json` declara exactamente un hook, y es lo único del plugin que se ejecuta solo:

```json
{ "hooks": { "SessionStart": [ { "matcher": "startup|clear|compact",
  "hooks": [ { "type": "command",
               "command": "\"${CLAUDE_PLUGIN_ROOT}/hooks/run-hook.cmd\" session-start",
               "shell": "bash", "async": false } ] } ] } }
```

Lo que inyecta es el contenido íntegro de la skill `using-superpowers`, en cada arranque de sesión,
cada `/clear` y cada compactación.

### Las catorce skills

| Skill | Qué es | Cómo se dispara |
|---|---|---|
| `using-superpowers` | La regla de despacho. No hace trabajo: obliga a comprobar si alguna otra skill aplica antes de responder o tocar nada | Sola, por el hook `SessionStart` |
| `brainstorming` | Convierte una idea en un diseño aprobado. Clasifica en spike, acotada o arquitectónica, y en los tres casos exige aprobación humana antes de implementar | Antes de cualquier trabajo creativo |
| `writing-plans` | Convierte la spec en un plan de tareas con ficheros exactos, tests y verificación, sin placeholders | Cuando ya hay spec o requisitos para trabajo de varios pasos |
| `executing-plans` | Ejecuta un plan escrito en la sesión, con puntos de control y parando ante cualquier bloqueo | Cuando existe un plan escrito y se elige esta vía |
| `subagent-driven-development` | Ejecuta el plan despachando un subagente por tarea, con doble revisión por tarea y revisión final de toda la rama. Lleva bitácora en disco | Alternativa recomendada a la anterior si hay subagentes |
| `systematic-debugging` | Cuatro fases obligatorias antes de proponer arreglo. Ley de hierro: ningún fix sin causa raíz investigada | Ante cualquier bug o comportamiento inesperado |
| `test-driven-development` | Rojo, verde, refactor. Ley de hierro: ningún código de producción sin un test que haya fallado antes | Antes de escribir código de una función o de un arreglo |
| `verification-before-completion` | Exige el comando ejecutado y su salida antes de admitir cualquier frase de éxito | Antes de decir que algo está listo, o de commitear |
| `using-git-worktrees` | Garantiza aislamiento: detecta si ya lo hay, prefiere la herramienta nativa, y solo si no la hay usa `git worktree` | Al empezar trabajo que necesita aislarse de la rama actual |
| `requesting-code-review` | Despacha un subagente revisor con contexto acotado, sin heredar el historial de la sesión | Tras cada tarea, al completar algo importante, antes de mezclar |
| `receiving-code-review` | Cómo digerir feedback: sin acuerdo teatral, verificando contra el código antes de implementar | Al recibir comentarios de revisión |
| `dispatching-parallel-agents` | Reparte problemas independientes entre subagentes en paralelo | Ante dos o más problemas que no comparten estado |
| `finishing-a-development-branch` | Cierre estándar: verifica tests, detecta worktree, presenta un menú de tres opciones y deja la decisión de integrar a la persona | Cuando la implementación está completa y en verde |
| `writing-skills` | TDD aplicado a la documentación de skills | Al crear o editar una skill |

Hay un mapa visual interactivo de estas catorce y de cómo se llaman entre sí, publicado como
artifact: <https://claude.ai/code/artifact/ef1c633e-5292-4602-87c0-5ec39be84976>

### Los artefactos que produce una sesión con Superpowers

Esto es lo que importa para la integración, porque es lo que hay que llevar a la tarea:

| Artefacto | Dónde lo deja | ¿Sobrevive a la sesión? |
|---|---|---|
| Spec de diseño | `docs/superpowers/specs/YYYY-MM-DD-<tema>-design.md`, commiteada | Sí, en git |
| Plan de implementación | `docs/superpowers/plans/YYYY-MM-DD-<nombre>.md`, commiteado | Sí, en git |
| Bitácora de ejecución | `.superpowers/sdd/<nombre-del-plan>/progress.md`, ignorada por git | **No: se borra con `rm -rf` al terminar** |
| Briefs y reports por tarea | el mismo directorio de la bitácora | **No** |
| Paquetes de revisión (diffs) | el mismo directorio | **No** |
| Commits | git | Sí |
| Las decisiones (`Ruling:`) | la bitácora, y un resumen en el mensaje final | **No: mueren con la sesión** |

Este repositorio ya tiene los dos primeros, del trabajo del modelo de estados:
`docs/superpowers/specs/2026-09-06-modelo-de-estados-design.md` y
`docs/superpowers/plans/2026-09-06-modelo-de-estados.md`.

---

## Cómo se activa una skill sin que nadie la invoque

Esto se analizó aparte porque de ello depende todo el problema 2, y la conclusión no es la
intuitiva.

`using-superpowers` no dice "usa brainstorming o systematic-debugging". Dice que si la descripción
de cualquiera de las catorce encaja con lo que estás a punto de hacer, aunque sea con un 1% de
probabilidad, hay que invocarla. Cita esas dos solo como las que más veces encajan.

Lo que hace que en la práctica solo esas dos sean puertas de entrada es otra cosa: **la descripción
de casi todas las demás describe un estado que solo existe después de que otra skill lo haya
producido.** `executing-plans` se dispara cuando "ya tienes un plan escrito";
`finishing-a-development-branch` cuando "la implementación está completa y los tests pasan";
`requesting-code-review` cuando "acabas de terminar una tarea". Nada de eso existe en el primer
mensaje de una conversación. Solo `brainstorming` ("crear algo, cambiar un comportamiento") y
`systematic-debugging` ("cualquier bug o comportamiento inesperado") describen algo que se puede
decir en frío.

Hay tres formas de llegar a las otras doce:

1. **Bajando por la cadena** desde una de esas dos puertas.
2. **Trayendo la precondición puesta.** Si el mensaje ya dice "aquí tienes `plan.md`, ejecútalo",
   eso encaja literalmente con la descripción de `executing-plans` sin que `writing-plans` haya
   intervenido nunca. Lo mismo con `receiving-code-review` si se pega el comentario de un revisor.
   **Este mecanismo es la clave del problema 2.**
3. **Nombrándola explícitamente**, o con `/nombre-de-la-skill`.

Un caso especial: `test-driven-development` tiene una descripción tan amplia como la de
`brainstorming` ("al implementar cualquier función o arreglo, antes de escribir código"), así que
técnicamente encajaría en frío. Lo que la frena es la regla de prioridad de `using-superpowers`:
las skills de proceso van primero y las de implementación después.

---

## Dos hallazgos que condicionan todo el diseño

### Superpowers documenta la fuga que no puede tapar

En `subagent-driven-development`, sección "Setup":

> Conversation memory does not survive compaction. In real sessions, controllers that lost their
> place have re-dispatched entire completed task sequences, the single most expensive failure
> observed. Track progress in a ledger file, not only in todos.

Esa bitácora vive en `<raíz-del-repo>/.superpowers/sdd/<nombre-del-plan>/progress.md`, la crea el
script `scripts/sdd-workspace PLAN_FILE` del propio plugin, y su formato está fijado:

- Primera línea de identidad: `# SDD ledger — plan: <ruta del plan>`
- Una línea por tarea completada: `Task <N>: complete (commits <base7>..<head7>, review clean)`
- Una línea por ronda de arreglo: `Task <N>: fix round <R>/5 (<X> addressed, <Y> open; commits ...)`
- Una línea por decisión tomada en nombre del humano:
  `Ruling: <qué decidí> — <por qué> — <qué cuesta si me equivoco>`
- Hallazgos aparcados: `Task <N>: parked — <hallazgo> — Ruling: <por qué el código se queda>`

Los separadores de esas líneas llevan el em-dash real (—) que usa el plugin, y son la única
excepción a la regla de este repositorio de no escribir nunca un em-dash: ese texto no lo genera
`biso` ni lo redacta este documento, sino que reproduce literalmente una convención ajena, y
cambiarlo por un guion normal haría que la cita no coincidiera con un `progress.md` de verdad.

Y en su sección "Finish", dos órdenes consecutivas:

> Before you delete anything, collect every ledger line containing `Ruling:` ... into your final
> message under "Rulings I made" ... That list is the only place the decisions you took on your
> human partner's behalf reach them ... **A ruling that dies with the workspace was a decision made
> in secret.**

> When the final whole-branch review is clean and its fixes are merged, delete this plan's workspace
> (`rm -rf <workspace>`), the git history is the record now.

O sea: el propio autor identifica que las decisiones tomadas en tu nombre solo te llegan por un
mensaje final, y que ese mensaje muere con la sesión. **Los `comments` de `biso` son inmutables,
con autor y con fecha (sección ["Los comentarios"](../../spec/modelo-de-datos.md#los-comentarios)). El encaje es exacto, y viene ya justificado
desde el otro lado.** Este es el argumento más fuerte a favor de toda la integración.

### `DECISIONES.md` ya descarta el canal fácil

La sección ["El presupuesto del mensaje de arranque"](../../DECISIONES.md#el-presupuesto-del-mensaje-de-arranque) de `docs/DECISIONES.md` dice, literalmente:

> `biso prime` sustituye por completo a las guías de instrucciones y a cualquier inyección de texto
> en los ficheros de convenciones del proyecto.

Y su tabla comparativa presume de la fila "contexto gastado en sesiones que no tocan tareas: 0",
contra la inyección permanente en el fichero de convenciones de la herramienta estudiada.

**Así que la vía fácil, meter el protocolo en el `CLAUDE.md` de este repositorio, está descartada
por el propio diseño de `biso`.** Conviene tenerlo presente porque es la solución que cualquiera
propondría primero, y es además la que Superpowers bendice explícitamente en su sección "User
Instructions" ("User instructions (CLAUDE.md, AGENTS.md, etc) take precedence over skills"). Aquí
no se puede usar sin contradecir la decisión que justifica la existencia de `biso prime`.

---

## Problema 1: la tarea como registro duradero

### No hace falta añadir ni un campo

El modelo lógico de la sección ["El modelo de datos de una tarea"](../../spec/modelo-de-datos.md) ya tiene sitio para todo lo que produce una
sesión con Superpowers. Inventario de lo que se usa:

| Campo de `biso` | Bandera | Qué guardaría de una sesión con Superpowers |
|---|---|---|
| `description` | `-d --desc` | El qué y el porqué, salido de la fase de entender de `brainstorming` |
| `acceptanceCriteria` | `--ac` | Los criterios de éxito acordados en el diseño |
| `definitionOfDone` | `--dod` | Lo que la spec exija como condición de cierre |
| `plan` | `--plan` | La posición actual y el puntero al plan, no su contenido |
| `documentation` | `--doc` | Las rutas de la spec y del plan commiteados |
| `notes` | `--note` | El diario: hipótesis descartadas, hallazgos, dónde se ha quedado |
| `comments` | `--comment` / `biso comment` | Cada `Ruling:`, inmutable, con autor y fecha |
| `question` | `biso ask` / `biso answer` | Las puertas donde Superpowers exige una persona |
| `summary` | `--summary` | El resumen final, al cerrar |
| `references` | `--ref` | Enlaces externos si los hay |
| `modifiedFiles` | `--file` | Ver ["Qué NO meter en la tarea, que es la mitad del diseño"](#qué-no-meter-en-la-tarea-que-es-la-mitad-del-diseño): en general **no** se usa |
| `parent` | `-p --parent` | Hallazgos aparcados que se convierten en tarea hija |

### La correspondencia, momento a momento

| Momento del flujo de Superpowers | Lo que produce | Escritura en `biso` |
|---|---|---|
| `brainstorming` clasifica y entiende | La idea con sus criterios de éxito | `biso new "TÍTULO" -d "qué y por qué" --ac "..."` |
| `brainstorming` presenta el diseño y espera el sí (HARD-GATE) | Un diseño bloqueado a la espera de una persona | `biso ask <ref> "<lo que hay que decidir>"`. El `biso answer` de la persona desaparca y deja pregunta y respuesta en los comentarios |
| `brainstorming` guarda y commitea la spec | `docs/superpowers/specs/...-design.md` | `biso set <ref> --doc <ruta>` |
| `writing-plans` guarda y commitea el plan | `docs/superpowers/plans/....md` | `biso set <ref> --doc <ruta>` |
| `using-git-worktrees` crea el aislamiento | Rama y ruta del worktree | `biso start <ref> --plan "rama X, worktree Y, plan en <ruta>"` |
| `subagent-driven-development` completa una tarea del plan | `Task N: complete (commits a..b)` | `biso note <ref> "Tarea N de M hecha, commits a..b"` |
| Cada `Ruling:` de la bitácora | Una decisión tomada en tu nombre | `biso comment <ref> "Ruling: qué, por qué, y qué cuesta si me equivoco"` |
| `systematic-debugging` descarta una hipótesis (fases 1 a 3) | Lo que más se pierde al compactar | `biso note <ref> "Descartado X porque Y"` |
| `systematic-debugging` acumula tres arreglos fallidos (fase cuatro y media) | Hay que cuestionar la arquitectura con una persona | `biso ask <ref> "..."` |
| Una revisión deja un hallazgo aparcado | `Task N: parked - ...` | Tarea hija con `--parent <ref>`, o `--note` si es menor |
| `verification-before-completion` da el visto bueno | La evidencia de que está en verde | Es la puerta previa al `finish`, no una escritura propia |
| `finishing-a-development-branch` presenta su menú | Tres opciones, decide la persona | Si no hay persona: `biso ask`. Si la hay y elige: `biso finish <ref> --check all --summary "..."` |

### Qué NO meter en la tarea, que es la mitad del diseño

Son tres cosas, y saltárselas es lo que convierte un tablero en un archivo muerto.

**El contenido de la spec y del plan no se copia.** Están commiteados en git y `--doc` guarda la
ruta. Copiarlos crea una segunda fuente de verdad que se desincroniza a la primera edición. Matiz
importante: conviene apuntar también **la rama** donde viven, porque una ruta sin rama no se
resuelve si el trabajo todavía no está mezclado.

**La lista de ficheros tocados tampoco**, y esto responde a la duda original sobre las rutas. El
campo `modifiedFiles` existe y `--file` lo escribe, pero `git diff --name-only` ya contesta eso
mejor y siempre actualizado. Lo que git no sabe, y sí merece el apunte, es **la rama, el worktree y
el rango de commits**, que es el puntero desde el que se reconstruye todo lo demás. La regla es:
apunta el puntero, no la copia. La excepción razonable sería un fichero tocado y todavía no
commiteado, que git no puede recuperar solo.

**La bitácora de Superpowers entera tampoco.** Está diseñada para el bucle interno, es de un plan
concreto, y se borra sola al acabar. Lo único que debe cruzar a `biso` es lo que muere con ella: las
líneas `Ruling:`, los hallazgos aparcados y la posición ("voy por la tarea 4 de 9"). Intentar
sustituir esa bitácora con `biso` sería duplicar dos registros con dos ciclos de vida distintos.

### Por qué esto le gana a un handoff, y no es por dónde se guarda

Un handoff es **una sola escritura al final**, en el peor momento posible (contexto casi lleno) y
solo si alguien se acuerda de pedirla. Si la sesión se corta antes, no hay absolutamente nada.

La tarea son **muchas escrituras pequeñas** en momentos que ya existen en el flujo de todas formas:
al empezar, en cada decisión, en cada puerta, al cerrar. Cada una es una llamada barata a un CLI
cuya salida por defecto no es un eco (principio 4 de la sección ["Los principios"](../../spec/principios.md)). Journalear una
tarea entera cuesta uno o dos miles de tokens, contra los diez kilobytes de un documento de handoff.

Y hay un tercer punto que el propio `prime` ya dice en su párrafo de cierre: `work, biso note <ref>
"..." as you go`. El bucle de escritura continua **ya está especificado**; lo que falta es que los
documentos y las decisiones entren en él.

---

## Problema 2: el despacho automático

### Lo que sí es del arnés

Confirmado: `biso` es un CLI y no puede inyectar nada en el contexto de nadie. Quién levanta al
agente y con qué prompt inicial es responsabilidad del arnés (un cron, un script, `claude -p`, un
hook `SessionStart`). La sección ["Lo que se deja fuera a propósito"](../../spec/fuera-de-alcance.md) ya lo deja fuera a propósito: "No hay servidor
de integración ni protocolo de herramientas".

### El matiz que cambia el diseño: el texto de la tarea es prompt

Cuando el agente ejecuta `biso get TASK-7`, lo que haya en `description`, `plan` y `notes` entra en
su contexto y funciona como instrucción. Y, por lo visto en la sección ["Cómo se activa una skill sin que nadie la invoque"](#cómo-se-activa-una-skill-sin-que-nadie-la-invoque) de este documento, las
skills se disparan por coincidencia con su descripción.

**Conclusión: una tarea cuyo texto se lee como un informe de fallo dispara `systematic-debugging`
sola, sin que el despachador nombre ninguna skill.** Eso que se quería sale gratis, con una
condición: que las tareas estén escritas con la forma del trabajo que son. Es el mismo mecanismo que
la skill `handoff` propia usa en su sección "Suggested skills", pero sin necesidad de un fichero
aparte.

### El conflicto con la puerta de `brainstorming`

`brainstorming` lleva un bloque `HARD-GATE` que prohíbe implementar nada, invocar ninguna skill de
implementación y escribir ningún código hasta que una persona apruebe. Y aplica a los tres caminos,
incluido el acotado: "the ceremony scales with the task; the approval gate never does".

**Un agente desatendido no puede cruzar esa puerta.** Si el despachador le da una tarea que todavía
necesita diseño, lo correcto es que se pare, y se pierde el turno.

### La política de despacho que se deduce

1. Despachar solo tareas listas: `biso ls --not-blocked --not-waiting` (los dos filtros existen,
   sección ["`biso ls`"](../../spec/cmd/ls.md), junto con `--blocked`, `--active`, `--not-active`, `--mine` y
   `--unassigned`).
2. Exigir además que **la puerta de diseño esté cerrada**, que en la práctica significa que la tarea
   ya lleva un `--doc` con su plan. Una tarea sin plan es trabajo de diseño y va a una persona.
3. Cuando el agente choque con una puerta a mitad de faena, la salida no es adivinar sino
   `biso ask`, que deja la tarea aparcada y visible en el bloque `NEEDS ANSWER` del `prime`
   sin que nadie tenga que abrir nada.
4. Las cuatro cosas que paran a `subagent-driven-development` (una operación irreversible o
   destructiva, una acción sensible de seguridad, un efecto fuera del worktree como mezclar o
   publicar, y un plan tan roto que todo camino sea una conjetura) son exactamente los cuatro casos
   en los que el agente debe llamar a `biso ask` en vez de decidir.

Esto ya está medio dicho en la regla 11 del mensaje de arranque: "Ask instead of guessing. A task
assigned to you is one a person decided you should do".

---

## Problema 3: por dónde se le dice esto al agente

### Los cuatro canales

| Canal | Cuándo llega al agente | Qué debería llevar | Coste fijo |
|---|---|---|---|
| El texto de la propia tarea | Al hacer `biso get`, justo cuando toca | Lo específico de esa tarea | Ninguno |
| `biso prime` | Al arrancar la sesión | La regla general, una o dos líneas | Fijo y con tope duro |
| Una skill propia | Cuando su descripción encaja con la situación | La tabla de correspondencia entera | Ninguno hasta que dispara |
| Hooks de Claude Code | `SessionStart` y `Stop` | Ejecución, no instrucción | Ninguno |

Queda fuera el `CLAUDE.md` del repositorio, por lo dicho en ["`DECISIONES.md` ya descarta el canal fácil"](#decisionesmd-ya-descarta-el-canal-fácil).

### Recomendación: no mencionar Superpowers en el SPEC

Ni el nombre. Tres motivos:

1. Es un plugin de terceros, fijado por un marketplace ajeno, que va por la 6.3.0 y reorganiza sus
   skills entre versiones. Atar la especificación a su vocabulario envejece mal.
2. La sección ["Lo que se deja fuera a propósito"](../../spec/fuera-de-alcance.md) ya declara que no hay sincronización con ningún sistema externo.
3. No hace falta. La regla que resuelve el problema es general y no nombra a nadie: **la tarea es el
   registro, los documentos van en `--doc`, las decisiones en `--note`, y se pregunta en vez de
   adivinar.** Un agente que tenga eso y las skills en contexto tiende el puente solo, porque el
   puente es obvio: la ruta de una spec es un documento y un `Ruling:` es una decisión.

Comprobado además que hoy la especificación no contiene ni una sola aparición de las palabras
`superpower`, `skill`, `handoff` ni `compact`. Es territorio virgen a propósito.

### El presupuesto del mensaje de arranque, con los números

De la sección ["El presupuesto de tamaño"](../../spec/presupuestos.md#el-presupuesto-de-tamaño) y de la sección ["El presupuesto del mensaje de arranque"](../../DECISIONES.md#el-presupuesto-del-mensaje-de-arranque) de `docs/DECISIONES.md`:

| Parte | Tope | Ocupado hoy | Libre |
|---|---:|---:|---:|
| Parte fija (título, `COMMANDS`, `FIELD FLAGS`, `RULES`, cierre) | 3.456 | 3.327 | **129** |
| Resumen del tablero (`BOARD` y los cuatro bloques) | 1.664 | 1.491 | **173** |
| **Total** | **5.120** | **4.818** | **302** |

**El hueco de la parte fija se ha encogido y eso cambia lo que se puede proponer.** Las banderas
`--check-dod` y `--uncheck-dod` entraron en la rejilla de `FIELD FLAGS`, y con ellas la parte fija
pasó de 3.255 a 3.327 bytes: de los 201 libres que había quedan 129. El resumen del tablero no se ha
movido, así que el total baja de 374 a 302 libres. Los 302 no sirven para medir texto nuevo del
bloque fijo, porque las dos partes tienen tope propio y el resumen no cede el suyo.

Dos restricciones que hay que respetar al proponer texto:

- El tope total de 5.120 bytes es de las pocas cosas que **no cambian nunca** según el ["contrato de
  estabilidad"](../../spec/estabilidad.md). El texto dentro del tope sí puede cambiar entre versiones menores, y de
  hecho esa sección dice que es "donde se espera que la herramienta más aprenda con el tiempo".
- `docs/DECISIONES.md` es explícito sobre qué cede cuando aprieta: "un tope que se sube cada vez que
  aprieta deja de ser un tope ... si los números no cupieran, lo que se recorta es contenido, no el
  tope".

### La frase propuesta, que hoy ya no cabe

Su sitio natural es el párrafo de cierre, que es donde ya vive el bucle:

```
Everything a later session needs lives in the task: --doc for a design or plan
file, --note for a decision, ask when a person must choose. Keep no notes elsewhere.
```

Ese texto mide **164 bytes** con su salto de línea final, y la parte fija tiene hoy **129 libres**.
Cuando este apartado se escribió había 201 y por eso decía que cabía; con las banderas nuevas de
`FIELD FLAGS` ya no cabe, y le faltan 35 bytes. No es un número que corregir: es que la propuesta, tal
como está escrita, no entra.

La regla del proyecto sobre qué cede cuando el presupuesto aprieta ya está citada en ["El presupuesto del mensaje de arranque, con los números"](#el-presupuesto-del-mensaje-de-arranque-con-los-números), y viene de
`docs/DECISIONES.md`: lo que se recorta es contenido, no el tope. Así que subir los 3.456 de la parte
fija para hacerle sitio a esta frase no es una de las opciones. Quedan tres, y **elegir entre ellas es
decisión de la sección ["La decisión abierta"](#la-decisión-abierta), no de este apartado**:

1. **Recortar la frase** hasta 129 bytes o menos, aceptando que pierde detalle. Los 35 bytes de más
   salen, por ejemplo, de dejar de nombrar los dos usos de `--doc` ("a design or plan file") o de
   suprimir la última oración, que es la que cierra la puerta a escribir notas en otro sitio y
   probablemente sea la parte que más aporta.
2. **Recortar otra cosa del bloque fijo** para hacerle sitio: alguna de las reglas, o una línea
   de `COMMANDS`. Esto es exactamente el caso que `docs/DECISIONES.md` describe como la prueba de
   fuego del tope, y solo vale si esa reducción sale limpia; si para meter la frase hay que quitar
   algo que un agente necesita para arrancar, la que sobra es la frase.
3. **Dejarla fuera del mensaje**, y quedarse con las dos alternativas que la sección ["La decisión abierta"](#la-decisión-abierta) ya lista: la
   skill propia, o el texto de las tareas.

Una advertencia sobre la variante que se proponía antes: meterla como regla numerada, la duodécima
del bloque `RULES`, no ahorra nada, porque hay que sumarle el prefijo de numeración y la sangría. Con
129 bytes libres, esa variante está aún más lejos de caber que la del cierre.

**Lo que no cabe, y no debe caber, es la tabla de correspondencia de la sección ["La correspondencia, momento a momento"](#la-correspondencia-momento-a-momento).**

### La tabla de correspondencia va en una skill propia

Es política del agente, no comportamiento de la herramienta, y una skill no gasta contexto hasta que
se dispara. Es el reemplazo natural de la skill `handoff`: la misma función (que el trabajo
sobreviva a la sesión) pero escribiendo durante en vez de al final.

Su descripción tiene que encajar con los momentos del ciclo para que dispare sola. Algo del estilo
de: `Use when starting, journaling, parking or finishing work on a task in a project with a biso
board`. Nótese que, por lo visto en la sección ["Cómo se activa una skill sin que nadie la invoque"](#cómo-se-activa-una-skill-sin-que-nadie-la-invoque), una descripción que resume el flujo de trabajo es
contraproducente: la propia `writing-skills` documenta que un agente puede seguir la descripción en
vez de leer la skill entera. La descripción dice **cuándo**, nunca **qué hace**.

### El hook es el único mecanismo coercitivo

Ninguna instrucción, ni en el `prime` ni en una skill, garantiza que se escriba nada. Dos hooks sí
aportan garantía real:

- **`SessionStart`** que ejecute `biso prime` quita el paso manual del arranque y asegura que el
  estado del tablero está siempre en contexto. En esta máquina ya hay un `SessionStart` que inyecta
  la salida de un script como `systemMessage`, así que el patrón está probado.
- **`Stop`** puede devolver un bloqueo con su motivo, e impedir que una sesión termine en silencio
  con una tarea activa y sin resumen. **Este detalle hay que confirmarlo en la documentación de
  hooks antes de construir encima: no se verificó en la sesión que escribió este documento.**

---

## La decisión abierta

**¿Gana el mensaje de `biso prime` la frase de ["La frase propuesta, que hoy ya no cabe"](#la-frase-propuesta-que-hoy-ya-no-cabe)?**

A favor: el criterio de diseño de la sección ["Qué resuelve este comando"](../../spec/cmd/prime.md#qué-resuelve-este-comando) dice que quien lea el mensaje y no haya visto nunca
la herramienta tiene que poder completar un ciclo de trabajo entero sin leer nada más. Hoy ese ciclo
es completo para trabajo pequeño, pero no dice dónde viven los documentos ni las decisiones, que es
exactamente lo que se pierde cuando la sesión muere.

En contra: el `prime` enseña la herramienta, no el proceso del proyecto, y "la tarea es el registro
duradero" es proceso.

**Y hay algo más que ha cambiado desde que se escribió esta pregunta: ya no es solo que la frase
tenga que merecer el sitio, es que el sitio no existe.** Cuando este documento planteó la duda, la
parte fija tenía 201 bytes libres y la frase medía unos 160, así que la única cuestión era si valía la
pena gastarlos. Hoy la parte fija tiene 129 libres y la frase mide 164: **la propuesta, con el tamaño
que tiene, no entra**. La pregunta abierta pasa a tener dos mitades, y la primera hay que contestarla
antes que la segunda:

1. ¿Se le hace sitio? Recortando la frase a 129 bytes o menos, o recortando algo del bloque fijo para
   que quepa entera. El apartado ["La frase propuesta, que hoy ya no cabe"](#la-frase-propuesta-que-hoy-ya-no-cabe) detalla las dos vías y lo que cuesta cada una.
2. Si se le hace sitio, ¿lo merece? Que es la pregunta original, con sus argumentos a favor y en
   contra intactos.

Subir el tope no está entre las salidas: la regla del proyecto, citada en ["El presupuesto del mensaje de arranque, con los números"](#el-presupuesto-del-mensaje-de-arranque-con-los-números), es que lo que se
recorta es contenido y no el tope.

Las dos alternativas si la respuesta es que no:

1. Solo la skill propia, sin tocar la especificación. Se puede probar mañana mismo y no compromete
   nada.
2. Meterlo en el texto de las tareas, tarea por tarea, que es el canal de coste cero pero exige que
   alguien escriba esa línea en cada tarea.

---

## Riesgos y cabos sueltos

- **El riesgo que esta lista llamaba "el más serio de este diseño" está cerrado, y conviene saber
  cómo.** Decía que si el tablero acabara viviendo dentro del repositorio y versionado, una tarea
  creada desde un worktree podría no verse desde `main`, y que entonces el registro duradero dejaría
  de serlo justo cuando más falta hace. Estaba bien identificado: es la peor propiedad de las
  herramientas que guardan las tareas como ficheros del árbol de trabajo, donde el estado se bifurca
  con la rama y una incidencia cerrada en una rama reaparece abierta en la principal. La sección ["La decisión de persistencia"](../../DECISIONES.md#la-decisión-de-persistencia)
  de `docs/DECISIONES.md` no lo mitiga, lo elimina: un tablero es una base de datos SQLite en un
  directorio propio **fuera** del proyecto, localizada por el fichero puntero versionado
  `.biso.json`, así que no hay ninguna rama que contenga la tarea y un worktree ve exactamente el
  mismo tablero que `main`. Esa es la propiedad de la que depende todo este documento, y ya no hay
  que esperar a nadie para darla por buena. Lo que `CLAUDE.md` señala hoy como lo primero que hay que
  resolver antes de escribir código no es la persistencia, sino cómo habla el programa con SQLite, y
  eso no afecta a nada de lo que se propone aquí.
- **El riesgo nuevo que aparece al cerrar esa decisión: el tablero ya no viaja, y los documentos
  sí.** El precio declarado de sacar el tablero del árbol de trabajo es que no se clona con el
  proyecto, y lo único que cruza es la exportación de texto de `biso snapshot` (sección ["`biso snapshot`"](../../spec/cmd/snapshot.md)).
  Eso toca de lleno al argumento de la sección ["Superpowers documenta la fuga que no puede tapar"](#superpowers-documenta-la-fuga-que-no-puede-tapar): los `Ruling:` que se rescatan
  de la bitácora antes del `rm -rf` pasan a vivir en una base de datos local a una máquina, así que
  la frase de Superpowers "la historia de git es el registro ahora" solo vuelve a ser cierta cuando
  alguien ejecuta `biso snapshot`. Sin ese paso, la integración cambia perder
  las decisiones al morir la sesión por perderlas al morir la máquina, que es mucho mejor pero no es
  lo que promete la sección ["Superpowers documenta la fuga que no puede tapar"](#superpowers-documenta-la-fuga-que-no-puede-tapar). Tanto la política de despacho de ["La política de despacho que se deduce"](#la-política-de-despacho-que-se-deduce) como la skill propia de ["La tabla de correspondencia va en una skill propia"](#la-tabla-de-correspondencia-va-en-una-skill-propia)
  deberían contar el `snapshot` como parte del cierre.
- **Y hay una configuración en la que ese riesgo desaparece del todo, que conviene tener presente
  aquí**: si el tablero se crea dentro del proyecto con `biso init --at` y el proyecto versiona esa
  carpeta, la revisión de la instantánea va al repositorio del código y viaja con su remoto, sin que
  nadie configure nada (sección ["`biso snapshot`"](../../spec/cmd/snapshot.md)). La base de datos sigue quedando fuera,
  porque el fichero de exclusión que `init` escribe dentro del tablero la excluye siempre, así que la
  propiedad de la que depende este documento, que un worktree vea el mismo tablero que `main`, se
  conserva igual. En esa configuración, "la historia de git es el registro ahora" vuelve a ser cierta
  en cuanto alguien ejecuta `biso snapshot`, y lo que cruza a otra máquina cruza con el proyecto.
- **El arrendamiento con caducidad existe precisamente para el fallo que motiva este documento**:
  detectar una tarea que un agente cogió y cuya sesión murió sin liberarla. La sección ["Saber si alguien está trabajando de verdad"](../../DECISIONES.md#saber-si-alguien-está-trabajando-de-verdad) de
  `docs/DECISIONES.md` ya lo resuelve, dentro de un capítulo titulado "El modelo de estados: cuatro
  requisitos, cerrados": se guardan `leaseExpiresAt` y `leaseHolder`, un plazo vencido no saca la
  tarea del estado activo por sí solo, y quien reclama una tarea vencida lo hace con el mismo
  `biso start` de siempre. Lo que este documento todavía no aprovecha es que el latido no es un
  comando aparte: cualquier escritura del agente renueva el plazo, así que el bucle de `biso note`
  de la sección ["Por qué esto le gana a un handoff, y no es por dónde se guarda"](#por-qué-esto-le-gana-a-un-handoff-y-no-es-por-dónde-se-guarda) ya sirve de latido sin añadir ni una llamada.
- **Superpowers puede cambiar bajo los pies.** La versión la fija el marketplace oficial, no
  nosotros. Cualquier acuerdo que dependa de rutas internas suyas (`.superpowers/sdd/...`) hay que
  darlo por frágil. Las rutas de `docs/superpowers/specs` y `docs/superpowers/plans` son más
  estables porque están escritas en las skills como convención de salida y ya se usan en este repo.
- **El bloqueo del hook `Stop`** está sin verificar (ver ["El hook es el único mecanismo coercitivo"](#el-hook-es-el-único-mecanismo-coercitivo)).
- **Doble registro.** Mientras la bitácora de Superpowers y el tablero convivan, hay dos sitios
  donde mirar. La regla de ["Qué NO meter en la tarea, que es la mitad del diseño"](#qué-no-meter-en-la-tarea-que-es-la-mitad-del-diseño) lo acota, pero conviene comprobar en uso real que no se degrada.

---

## Si se aprueba, qué hay que tocar

En este orden, y ninguno de los pasos depende de que exista una línea de código de `biso`:

1. **La skill propia.** Crear `home/.claude/skills/<nombre>/SKILL.md` en el repo de dotfiles con la
   tabla de la sección ["La correspondencia, momento a momento"](#la-correspondencia-momento-a-momento), siguiendo el procedimiento de su `CLAUDE.md` (incluida la excepción en
   `home/.claude/.gitignore`, que es el paso que se olvida). Esto no toca la especificación de
   `biso` y se puede probar en cuanto exista el CLI.
2. **La frase del `prime`**, si la decisión de la sección ["La decisión abierta"](#la-decisión-abierta) es que sí: primero dejarla en 129 bytes o
   menos, o recortar antes lo que haga falta del bloque fijo, porque con su tamaño de hoy no cabe;
   luego editar la salida literal de la sección ["La salida literal"](../../spec/cmd/prime.md#la-salida-literal), **recalcular los bytes de las
   dos partes** y actualizar los números de la sección ["El presupuesto de tamaño"](../../spec/presupuestos.md#el-presupuesto-de-tamaño) y de la sección ["El presupuesto del mensaje de arranque"](../../DECISIONES.md#el-presupuesto-del-mensaje-de-arranque) de `docs/DECISIONES.md`, que hoy
   dicen 3.327 y 4.818.
3. **La entrada en `docs/DECISIONES.md`** explicando por qué esa frase entra y por qué no se nombra
   a Superpowers. Sin eso, la siguiente sesión que vea la frase la puede quitar por parecer ajena al
   resto del mensaje.
4. **Los hooks**, en el repo de dotfiles: `SessionStart` para el `prime` y, si se confirma que puede
   bloquear, `Stop` para el cierre.
5. **La política de despacho** de la sección ["La política de despacho que se deduce"](#la-política-de-despacho-que-se-deduce), cuando exista el despachador.

Cuidado con una cosa al tocar el `prime`: `CLAUDE.md` de este repo avisa de que los ejemplos de
salida del documento se generan y no se escriben a mano, así que cualquier cambio obliga a
regenerarlos y a comprobar que coinciden carácter a carácter.

---

## Referencias

**En este repositorio:**

- `docs/spec/`, sección ["El modelo de datos de una tarea"](../../spec/modelo-de-datos.md), el modelo de datos y sus campos.
- `docs/spec/`, secciones ["Qué resuelve este comando"](../../spec/cmd/prime.md#qué-resuelve-este-comando), ["El presupuesto de tamaño"](../../spec/presupuestos.md#el-presupuesto-de-tamaño), ["Qué entra en el mensaje y qué se relega a `--help`"](../../spec/cmd/prime.md#qué-entra-en-el-mensaje-y-qué-se-relega-a---help) y ["La salida literal"](../../spec/cmd/prime.md#la-salida-literal), el mensaje de arranque, su presupuesto y su texto
  literal.
- `docs/spec/`, sección ["`biso ls`"](../../spec/cmd/ls.md), los filtros de `biso ls`, incluidos `--not-blocked` y `--waiting`.
- `docs/spec/`, sección ["El contrato de estabilidad"](../../spec/estabilidad.md), el contrato de estabilidad, que congela el tope de 5.120 bytes.
- `docs/spec/`, sección ["Lo que se deja fuera a propósito"](../../spec/fuera-de-alcance.md), lo que se deja fuera a propósito.
- `docs/DECISIONES.md` sección ["El presupuesto del mensaje de arranque"](../../DECISIONES.md#el-presupuesto-del-mensaje-de-arranque), el presupuesto del mensaje de arranque y la decisión de sustituir
  las guías de instrucciones.
- `docs/DECISIONES.md` sección ["Saber si alguien está trabajando de verdad"](../../DECISIONES.md#saber-si-alguien-está-trabajando-de-verdad), el arrendamiento con caducidad, ya decidido.

**Fuera de este repositorio:**

- `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/skills/`, las catorce skills.
- `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/hooks/hooks.json`, el único
  hook.
- `~/.claude/skills/handoff/SKILL.md`, la skill que este diseño quiere hacer innecesaria.
- <https://claude.ai/code/artifact/ef1c633e-5292-4602-87c0-5ec39be84976>, el mapa visual de las
  catorce skills y de cómo se llaman entre sí.
