# Integrar `biso` con las skills de Superpowers

> **Estado: análisis, no decisión.** Este documento recoge entero el análisis de una sesión del
> 2026-09-07 que se iba a perder. Nada de lo que hay aquí está implementado, ni añadido a
> `docs/SPEC.md`, ni acordado. La decisión que lo cierra está al final, en "La decisión abierta".
>
> Se escribe con el detalle suficiente para que una sesión fría lo retome sin volver a investigar
> nada: cada afirmación sobre Superpowers lleva el fichero de donde sale, y cada afirmación sobre
> `biso` lleva su sección de `docs/SPEC.md`.

## Índice

1. [El problema que se estaba resolviendo](#1-el-problema-que-se-estaba-resolviendo)
2. [El material de partida: qué es Superpowers y qué hay instalado](#2-el-material-de-partida-qué-es-superpowers-y-qué-hay-instalado)
3. [Cómo se activa una skill sin que nadie la invoque](#3-cómo-se-activa-una-skill-sin-que-nadie-la-invoque)
4. [Dos hallazgos que condicionan todo el diseño](#4-dos-hallazgos-que-condicionan-todo-el-diseño)
5. [Problema 1: la tarea como registro duradero](#5-problema-1-la-tarea-como-registro-duradero)
6. [Problema 2: el despacho automático](#6-problema-2-el-despacho-automático)
7. [Problema 3: por dónde se le dice esto al agente](#7-problema-3-por-dónde-se-le-dice-esto-al-agente)
8. [La decisión abierta](#8-la-decisión-abierta)
9. [Riesgos y cabos sueltos](#9-riesgos-y-cabos-sueltos)
10. [Si se aprueba, qué hay que tocar](#10-si-se-aprueba-qué-hay-que-tocar)
11. [Referencias](#11-referencias)

---

## 1. El problema que se estaba resolviendo

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

## 2. El material de partida: qué es Superpowers y qué hay instalado

### 2.1. De dónde viene y dónde está

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

### 2.2. El único hook

`hooks/hooks.json` declara exactamente un hook, y es lo único del plugin que se ejecuta solo:

```json
{ "hooks": { "SessionStart": [ { "matcher": "startup|clear|compact",
  "hooks": [ { "type": "command",
               "command": "\"${CLAUDE_PLUGIN_ROOT}/hooks/run-hook.cmd\" session-start",
               "shell": "bash", "async": false } ] } ] } }
```

Lo que inyecta es el contenido íntegro de la skill `using-superpowers`, en cada arranque de sesión,
cada `/clear` y cada compactación.

### 2.3. Las catorce skills

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

### 2.4. Los artefactos que produce una sesión con Superpowers

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

## 3. Cómo se activa una skill sin que nadie la invoque

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

## 4. Dos hallazgos que condicionan todo el diseño

### 4.1. Superpowers documenta la fuga que no puede tapar

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

Y en su sección "Finish", dos órdenes consecutivas:

> Before you delete anything, collect every ledger line containing `Ruling:` ... into your final
> message under "Rulings I made" ... That list is the only place the decisions you took on your
> human partner's behalf reach them ... **A ruling that dies with the workspace was a decision made
> in secret.**

> When the final whole-branch review is clean and its fixes are merged, delete this plan's workspace
> (`rm -rf <workspace>`), the git history is the record now.

O sea: el propio autor identifica que las decisiones tomadas en tu nombre solo te llegan por un
mensaje final, y que ese mensaje muere con la sesión. **Los `comments` de `biso` son inmutables,
con autor y con fecha (sección 5.2 de `docs/SPEC.md`). El encaje es exacto, y viene ya justificado
desde el otro lado.** Este es el argumento más fuerte a favor de toda la integración.

### 4.2. `DECISIONES.md` ya descarta el canal fácil

La sección 3 de `docs/DECISIONES.md` dice, literalmente:

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

## 5. Problema 1: la tarea como registro duradero

### 5.1. No hace falta añadir ni un campo

El modelo lógico de la sección 5 de `docs/SPEC.md` ya tiene sitio para todo lo que produce una
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
| `modifiedFiles` | `--file` | Ver 5.3: en general **no** se usa |
| `parent` | `-p --parent` | Hallazgos aparcados que se convierten en tarea hija |

### 5.2. La correspondencia, momento a momento

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
| `systematic-debugging` acumula tres arreglos fallidos (fase 4.5) | Hay que cuestionar la arquitectura con una persona | `biso ask <ref> "..."` |
| Una revisión deja un hallazgo aparcado | `Task N: parked — ...` | Tarea hija con `--parent <ref>`, o `--note` si es menor |
| `verification-before-completion` da el visto bueno | La evidencia de que está en verde | Es la puerta previa al `finish`, no una escritura propia |
| `finishing-a-development-branch` presenta su menú | Tres opciones, decide la persona | Si no hay persona: `biso ask`. Si la hay y elige: `biso finish <ref> --check all --summary "..."` |

### 5.3. Qué NO meter en la tarea, que es la mitad del diseño

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

### 5.4. Por qué esto le gana a un handoff, y no es por dónde se guarda

Un handoff es **una sola escritura al final**, en el peor momento posible (contexto casi lleno) y
solo si alguien se acuerda de pedirla. Si la sesión se corta antes, no hay absolutamente nada.

La tarea son **muchas escrituras pequeñas** en momentos que ya existen en el flujo de todas formas:
al empezar, en cada decisión, en cada puerta, al cerrar. Cada una es una llamada barata a un CLI
cuya salida por defecto no es un eco (principio 4 de la sección 1 de `docs/SPEC.md`). Journalear una
tarea entera cuesta uno o dos miles de tokens, contra los diez kilobytes de un documento de handoff.

Y hay un tercer punto que el propio `prime` ya dice en su párrafo de cierre: `work, biso note <ref>
"..." as you go`. El bucle de escritura continua **ya está especificado**; lo que falta es que los
documentos y las decisiones entren en él.

---

## 6. Problema 2: el despacho automático

### 6.1. Lo que sí es del arnés

Confirmado: `biso` es un CLI y no puede inyectar nada en el contexto de nadie. Quién levanta al
agente y con qué prompt inicial es responsabilidad del arnés (un cron, un script, `claude -p`, un
hook `SessionStart`). La sección 14 de `docs/SPEC.md` ya lo deja fuera a propósito: "No hay servidor
de integración ni protocolo de herramientas".

### 6.2. El matiz que cambia el diseño: el texto de la tarea es prompt

Cuando el agente ejecuta `biso get TASK-7`, lo que haya en `description`, `plan` y `notes` entra en
su contexto y funciona como instrucción. Y, por lo visto en la sección 3 de este documento, las
skills se disparan por coincidencia con su descripción.

**Conclusión: una tarea cuyo texto se lee como un informe de fallo dispara `systematic-debugging`
sola, sin que el despachador nombre ninguna skill.** Eso que se quería sale gratis, con una
condición: que las tareas estén escritas con la forma del trabajo que son. Es el mismo mecanismo que
la skill `handoff` propia usa en su sección "Suggested skills", pero sin necesidad de un fichero
aparte.

### 6.3. El conflicto con la puerta de `brainstorming`

`brainstorming` lleva un bloque `HARD-GATE` que prohíbe implementar nada, invocar ninguna skill de
implementación y escribir ningún código hasta que una persona apruebe. Y aplica a los tres caminos,
incluido el acotado: "the ceremony scales with the task; the approval gate never does".

**Un agente desatendido no puede cruzar esa puerta.** Si el despachador le da una tarea que todavía
necesita diseño, lo correcto es que se pare, y se pierde el turno.

### 6.4. La política de despacho que se deduce

1. Despachar solo tareas listas: `biso ls --ready --not-waiting` (los dos filtros existen, sección
   10.4 de `docs/SPEC.md`, junto con `--blocked`, `--active`, `--not-active`, `--mine` y
   `--unassigned`).
2. Exigir además que **la puerta de diseño esté cerrada**, que en la práctica significa que la tarea
   ya lleva un `--doc` con su plan. Una tarea sin plan es trabajo de diseño y va a una persona.
3. Cuando el agente choque con una puerta a mitad de faena, la salida no es adivinar sino
   `biso ask`, que deja la tarea aparcada y visible en el bloque `WAITING ON A PERSON` del `prime`
   sin que nadie tenga que abrir nada.
4. Las cuatro cosas que paran a `subagent-driven-development` (una operación irreversible o
   destructiva, una acción sensible de seguridad, un efecto fuera del worktree como mezclar o
   publicar, y un plan tan roto que todo camino sea una conjetura) son exactamente los cuatro casos
   en los que el agente debe llamar a `biso ask` en vez de decidir.

Esto ya está medio dicho en la regla 11 del mensaje de arranque: "Ask instead of guessing. A task
assigned to you is one a person decided you should do".

---

## 7. Problema 3: por dónde se le dice esto al agente

### 7.1. Los cuatro canales

| Canal | Cuándo llega al agente | Qué debería llevar | Coste fijo |
|---|---|---|---|
| El texto de la propia tarea | Al hacer `biso get`, justo cuando toca | Lo específico de esa tarea | Ninguno |
| `biso prime` | Al arrancar la sesión | La regla general, una o dos líneas | Fijo y con tope duro |
| Una skill propia | Cuando su descripción encaja con la situación | La tabla de correspondencia entera | Ninguno hasta que dispara |
| Hooks de Claude Code | `SessionStart` y `Stop` | Ejecución, no instrucción | Ninguno |

Queda fuera el `CLAUDE.md` del repositorio, por lo dicho en 4.2.

### 7.2. Recomendación: no mencionar Superpowers en el SPEC

Ni el nombre. Tres motivos:

1. Es un plugin de terceros, fijado por un marketplace ajeno, que va por la 6.3.0 y reorganiza sus
   skills entre versiones. Atar la especificación a su vocabulario envejece mal.
2. La sección 14 de `docs/SPEC.md` ya declara que no hay sincronización con ningún sistema externo.
3. No hace falta. La regla que resuelve el problema es general y no nombra a nadie: **la tarea es el
   registro, los documentos van en `--doc`, las decisiones en `--note`, y se pregunta en vez de
   adivinar.** Un agente que tenga eso y las skills en contexto tiende el puente solo, porque el
   puente es obvio: la ruta de una spec es un documento y un `Ruling:` es una decisión.

Comprobado además que hoy `docs/SPEC.md` no contiene ni una sola aparición de las palabras
`superpower`, `skill`, `handoff` ni `compact`. Es territorio virgen a propósito.

### 7.3. El presupuesto del mensaje de arranque, con los números

De la sección 9.5 de `docs/SPEC.md` y de la 3 de `docs/DECISIONES.md`:

| Mitad | Tope | Ocupado hoy | Libre |
|---|---:|---:|---:|
| Parte fija (título, `COMMANDS`, `FIELD FLAGS`, `RULES`, cierre) | 3.456 | 3.255 | **201** |
| Resumen del tablero (`BOARD` y los cuatro bloques) | 1.664 | 1.441 | **223** |
| **Total** | **5.120** | **4.696** | **424** |

Dos restricciones que hay que respetar al proponer texto:

- El tope total de 5.120 bytes es de las pocas cosas que **no cambian nunca** según el contrato de
  estabilidad (sección 13). El texto dentro del tope sí puede cambiar entre versiones menores, y de
  hecho la sección 13 dice que es "donde se espera que la herramienta más aprenda con el tiempo".
- `docs/DECISIONES.md` es explícito sobre qué cede cuando aprieta: "un tope que se sube cada vez que
  aprieta deja de ser un tope ... si los números no cupieran, lo que se recorta es contenido, no el
  tope".

### 7.4. La frase propuesta

Cabe en el párrafo de cierre, que es donde ya vive el bucle. Unos 160 bytes de los 201 libres de la
mitad fija:

```
Everything a later session needs lives in the task: --doc for a design or plan
file, --note for a decision, ask when a person must choose. Keep no notes elsewhere.
```

Alternativa, si se prefiere como regla numerada en vez de en el cierre: sería la duodécima del
bloque `RULES`, y entonces hay que contar también el prefijo de numeración y su sangría.

**Lo que no cabe, y no debe caber, es la tabla de correspondencia de la sección 5.2.**

### 7.5. La tabla de correspondencia va en una skill propia

Es política del agente, no comportamiento de la herramienta, y una skill no gasta contexto hasta que
se dispara. Es el reemplazo natural de la skill `handoff`: la misma función (que el trabajo
sobreviva a la sesión) pero escribiendo durante en vez de al final.

Su descripción tiene que encajar con los momentos del ciclo para que dispare sola. Algo del estilo
de: `Use when starting, journaling, parking or finishing work on a task in a project with a biso
board`. Nótese que, por lo visto en la sección 3, una descripción que resume el flujo de trabajo es
contraproducente: la propia `writing-skills` documenta que un agente puede seguir la descripción en
vez de leer la skill entera. La descripción dice **cuándo**, nunca **qué hace**.

### 7.6. El hook es el único mecanismo coercitivo

Ninguna instrucción, ni en el `prime` ni en una skill, garantiza que se escriba nada. Dos hooks sí
aportan garantía real:

- **`SessionStart`** que ejecute `biso prime` quita el paso manual del arranque y asegura que el
  estado del tablero está siempre en contexto. En esta máquina ya hay un `SessionStart` que inyecta
  la salida de un script como `systemMessage`, así que el patrón está probado.
- **`Stop`** puede devolver un bloqueo con su motivo, e impedir que una sesión termine en silencio
  con una tarea activa y sin resumen. **Este detalle hay que confirmarlo en la documentación de
  hooks antes de construir encima: no se verificó en la sesión que escribió este documento.**

---

## 8. La decisión abierta

**¿Gana el mensaje de `biso prime` la frase de 7.4?**

A favor: el criterio de diseño de la sección 9.1 dice que quien lea el mensaje y no haya visto nunca
la herramienta tiene que poder completar un ciclo de trabajo entero sin leer nada más. Hoy ese ciclo
es completo para trabajo pequeño, pero no dice dónde viven los documentos ni las decisiones, que es
exactamente lo que se pierde cuando la sesión muere.

En contra: el `prime` enseña la herramienta, no el proceso del proyecto, y "la tarea es el registro
duradero" es proceso. Y son 160 de los 201 bytes libres de la mitad fija, que es la que no se puede
recortar sola.

Las dos alternativas si la respuesta es que no:

1. Solo la skill propia, sin tocar la especificación. Se puede probar mañana mismo y no compromete
   nada.
2. Meterlo en el texto de las tareas, tarea por tarea, que es el canal de coste cero pero exige que
   alguien escriba esa línea en cada tarea.

---

## 9. Riesgos y cabos sueltos

- **La decisión de persistencia está abierta y afecta a todo esto.** `CLAUDE.md` dice que es lo
  único que bloquea escribir código, y que al tomarla hay que responder qué ocurre con una tarea que
  existe en una versión del proyecto y no en otra. Si el tablero acabara viviendo dentro del
  repositorio y versionado, **una tarea creada desde un worktree podría no verse desde `main`**, y
  entonces el registro duradero deja de serlo justo cuando más falta hace. Es el riesgo más serio
  de este diseño y no se puede cerrar hasta esa decisión.
- **El arrendamiento con caducidad, aplazado en la sección 9.2 de `docs/DECISIONES.md`, existe
  precisamente para el fallo que motiva este documento**: detectar una tarea que un agente cogió y
  cuya sesión murió sin liberarla. Cuando se retome, conviene mirarlo a la vez que esto.
- **Superpowers puede cambiar bajo los pies.** La versión la fija el marketplace oficial, no
  nosotros. Cualquier acuerdo que dependa de rutas internas suyas (`.superpowers/sdd/...`) hay que
  darlo por frágil. Las rutas de `docs/superpowers/specs` y `docs/superpowers/plans` son más
  estables porque están escritas en las skills como convención de salida y ya se usan en este repo.
- **El bloqueo del hook `Stop`** está sin verificar (ver 7.6).
- **Doble registro.** Mientras la bitácora de Superpowers y el tablero convivan, hay dos sitios
  donde mirar. La regla de 5.3 lo acota, pero conviene comprobar en uso real que no se degrada.

---

## 10. Si se aprueba, qué hay que tocar

En este orden, y ninguno de los pasos depende de que exista una línea de código de `biso`:

1. **La skill propia.** Crear `home/.claude/skills/<nombre>/SKILL.md` en el repo de dotfiles con la
   tabla de la sección 5.2, siguiendo el procedimiento de su `CLAUDE.md` (incluida la excepción en
   `home/.claude/.gitignore`, que es el paso que se olvida). Esto no toca la especificación de
   `biso` y se puede probar en cuanto exista el CLI.
2. **La frase del `prime`**, si la decisión de la sección 8 es que sí: editar la salida literal de
   la sección 9.7 de `docs/SPEC.md`, **recalcular los bytes de las dos mitades** y actualizar los
   números de la 9.5 y de la sección 3 de `docs/DECISIONES.md`, que hoy dicen 3.255 y 4.696.
3. **La entrada en `docs/DECISIONES.md`** explicando por qué esa frase entra y por qué no se nombra
   a Superpowers. Sin eso, la siguiente sesión que vea la frase la puede quitar por parecer ajena al
   resto del mensaje.
4. **Los hooks**, en el repo de dotfiles: `SessionStart` para el `prime` y, si se confirma que puede
   bloquear, `Stop` para el cierre.
5. **La política de despacho** de la sección 6.4, cuando exista el despachador.

Cuidado con una cosa al tocar el `prime`: `CLAUDE.md` de este repo avisa de que los ejemplos de
salida del documento se generan y no se escriben a mano, así que cualquier cambio obliga a
regenerarlos y a comprobar que coinciden carácter a carácter.

---

## 11. Referencias

**En este repositorio:**

- `docs/SPEC.md` sección 5, el modelo de datos y sus campos.
- `docs/SPEC.md` secciones 9.1, 9.5, 9.6 y 9.7, el mensaje de arranque, su presupuesto y su texto
  literal.
- `docs/SPEC.md` sección 10.4, los filtros de `biso ls`, incluidos `--ready` y `--waiting`.
- `docs/SPEC.md` sección 13, el contrato de estabilidad, que congela el tope de 5.120 bytes.
- `docs/SPEC.md` sección 14, lo que se deja fuera a propósito.
- `docs/DECISIONES.md` sección 3, el presupuesto del mensaje de arranque y la decisión de sustituir
  las guías de instrucciones.
- `docs/DECISIONES.md` sección 9.2, el arrendamiento aplazado.

**Fuera de este repositorio:**

- `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/skills/`, las catorce skills.
- `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/hooks/hooks.json`, el único
  hook.
- `~/.claude/skills/handoff/SKILL.md`, la skill que este diseño quiere hacer innecesaria.
- <https://claude.ai/code/artifact/ef1c633e-5292-4602-87c0-5ec39be84976>, el mapa visual de las
  catorce skills y de cómo se llaman entre sí.
