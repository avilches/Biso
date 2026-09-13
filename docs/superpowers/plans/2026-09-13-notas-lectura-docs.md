# Notas de lectura sobre la documentación Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Añadir un botón flotante al sitio de MkDocs servido en local (`make docs-serve`) que,
mediante un picker de elementos tipo inspector de DevTools, copie al portapapeles una línea de
nota lista para pegar en `docs/docs-tooling/notas-pendientes.md`, sin que nada de esto aparezca
nunca en `make docs-build`.

**Architecture:** Un hook de Python nuevo (`notas_lectura.py`) que en el evento `on_post_page`
de MkDocs inyecta como texto, antes de `</body>`, el contenido de dos ficheros hermanos
(`notas_lectura.js`, `notas_lectura.css`), pero solo cuando la variable de entorno
`BISO_DOCS_SERVE` está puesta. Esa variable la exporta únicamente el target `docs-serve` del
`Makefile`. El JS implementa el botón, el modo de selección de elementos y el panel que compone
la línea de nota; el CSS lo viste con las variables de color de Material.

**Tech Stack:** Python 3.12 (hook de MkDocs, sin dependencias nuevas), JavaScript vanilla (sin
build, sin dependencias), CSS plano. Pruebas del hook con `pytest` (ya usado en
`docs/docs-tooling/tools/tests/`).

**Spec:** `docs/superpowers/specs/2026-09-12-notas-lectura-docs-design.md`

## Global Constraints

- Nunca em-dash (`—`) en ningún texto generado: ni código, ni comentarios, ni mensajes de commit.
- `docs/docs-tooling/mkdocs/notas_lectura.py`, `.js` y `.css` van enteros en inglés (comentarios
  incluidos), como el resto del código fuente del repositorio.
- El texto que ve la persona que lee la documentación (las etiquetas del panel del botón, el
  propio fichero `notas-pendientes.md`) va en español, como el resto de la documentación.
- El separador de las tres partes de una línea de nota es `|`, nunca em-dash (spec, sección "El
  fichero de notas").
- Ningún `alert`, `confirm` ni `prompt` nativo: toda confirmación visual vive dentro del propio
  panel.
- Ningún commit ni mensaje lleva coautoría ni menciona haber sido generado por un agente.
- El trabajo vive en el worktree `.claude/worktrees/notas-lectura-docs`
  (rama `worktree-notas-lectura-docs`); no se edita el checkout principal.

---

## File Structure

- `docs/docs-tooling/mkdocs/notas_lectura.py` (nuevo): el hook. Expone `notes_enabled()`,
  `build_snippet()` y el evento `on_post_page()` que usa MkDocs.
- `docs/docs-tooling/mkdocs/tests/test_notas_lectura.py` (nuevo): pruebas de ese hook con
  `pytest`.
- `docs/docs-tooling/mkdocs/notas_lectura.js` (nuevo): el botón, el picker de elementos y el
  panel.
- `docs/docs-tooling/mkdocs/notas_lectura.css` (nuevo): su estilo.
- `docs/docs-tooling/mkdocs/mkdocs.yml` (modificado): añade `notas_lectura.py` a `hooks:`.
- `Makefile` (modificado): el target `docs-serve` exporta `BISO_DOCS_SERVE=1`.
- `docs/docs-tooling/notas-pendientes.md` (nuevo): el fichero de notas, vacío salvo su
  cabecera explicativa.

---

### Task 1: El hook que decide cuándo inyectar el botón

**Files:**
- Create: `docs/docs-tooling/mkdocs/notas_lectura.py`
- Test: `docs/docs-tooling/mkdocs/tests/test_notas_lectura.py`

**Interfaces:**
- Produces: `notas_lectura.notes_enabled(env: dict[str, str] | None = None) -> bool`,
  `notas_lectura.build_snippet(js_text: str, css_text: str) -> str`,
  `notas_lectura.on_post_page(output: str, page, config) -> str`, y las constantes de módulo
  `notas_lectura.JS_PATH` y `notas_lectura.CSS_PATH` (rutas `pathlib.Path` a los ficheros
  hermanos `.js`/`.css`, que Task 2 crea; este task no necesita que existan para pasar sus
  pruebas, porque las pruebas de `on_post_page` sustituyen esas rutas por ficheros temporales).

- [ ] **Step 1: Crear el directorio de pruebas y escribir la primera prueba, que falla**

Crea `docs/docs-tooling/mkdocs/tests/test_notas_lectura.py`:

```python
"""Tests for the reading-notes injection hook."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import notas_lectura


def test_notes_disabled_by_default(monkeypatch):
    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    assert notas_lectura.notes_enabled() is False


def test_notes_enabled_when_env_var_set(monkeypatch):
    monkeypatch.setenv("BISO_DOCS_SERVE", "1")
    assert notas_lectura.notes_enabled() is True
```

- [ ] **Step 2: Ejecutar las pruebas y comprobar que fallan**

Run: `uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project pytest docs/docs-tooling/mkdocs/tests -q`
Expected: FAIL, `ModuleNotFoundError: No module named 'notas_lectura'`.

- [ ] **Step 3: Escribir `notes_enabled` para que pase**

Crea `docs/docs-tooling/mkdocs/notas_lectura.py`:

```python
"""MkDocs hook that injects the reading-notes button only for `mkdocs serve` runs.

`mkdocs build` and `mkdocs serve` read the same `mkdocs.yml`, and anything under
`docs/docs-tooling/` (where this file lives) is excluded from the site by `exclude_docs`, so it
is never copied into `site/` and `extra_javascript`/`extra_css` cannot reach it. Instead, this
hook reads its sibling `notas_lectura.js` and `notas_lectura.css` from disk and inlines their
content into each rendered page, right before `</body>`, only when the `BISO_DOCS_SERVE`
environment variable is set. The `docs-serve` target of the Makefile sets it before invoking
`mkdocs serve`; `docs-build` and `docs-doctor` never do, so the built site never contains a
trace of the button.
"""

from __future__ import annotations

import os
from pathlib import Path

NOTES_ENV_VAR = "BISO_DOCS_SERVE"
HERE = Path(__file__).resolve().parent
JS_PATH = HERE / "notas_lectura.js"
CSS_PATH = HERE / "notas_lectura.css"


def notes_enabled(env: dict[str, str] | None = None) -> bool:
    source = env if env is not None else os.environ
    return bool(source.get(NOTES_ENV_VAR))
```

- [ ] **Step 4: Ejecutar las pruebas y comprobar que pasan**

Run: `uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project pytest docs/docs-tooling/mkdocs/tests -q`
Expected: PASS (2 passed).

- [ ] **Step 5: Añadir las pruebas de `build_snippet` y `on_post_page`, que fallan**

Añade al final de `docs/docs-tooling/mkdocs/tests/test_notas_lectura.py`:

```python
def test_build_snippet_wraps_css_and_js():
    snippet = notas_lectura.build_snippet("console.log(1)", "body{color:red}")
    assert "<style>body{color:red}</style>" in snippet
    assert "<script>console.log(1)</script>" in snippet


def test_on_post_page_leaves_output_untouched_when_disabled(monkeypatch):
    monkeypatch.delenv("BISO_DOCS_SERVE", raising=False)
    output = "<html><body>hola</body></html>"
    assert notas_lectura.on_post_page(output, page=None, config=None) == output


def test_on_post_page_injects_snippet_before_closing_body(monkeypatch, tmp_path):
    js_file = tmp_path / "notas_lectura.js"
    css_file = tmp_path / "notas_lectura.css"
    js_file.write_text("console.log('hola')", encoding="utf-8")
    css_file.write_text("body{color:red}", encoding="utf-8")
    monkeypatch.setattr(notas_lectura, "JS_PATH", js_file)
    monkeypatch.setattr(notas_lectura, "CSS_PATH", css_file)
    monkeypatch.setenv("BISO_DOCS_SERVE", "1")

    output = "<html><body>hola</body></html>"
    result = notas_lectura.on_post_page(output, page=None, config=None)

    assert "console.log('hola')" in result
    assert "body{color:red}" in result
    assert result.endswith("</body></html>")
    assert result.index("<script>") < result.index("</body>")
```

- [ ] **Step 6: Ejecutar las pruebas y comprobar que fallan**

Run: `uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project pytest docs/docs-tooling/mkdocs/tests -q`
Expected: FAIL, `AttributeError: module 'notas_lectura' has no attribute 'build_snippet'` (y
sin llegar a ejecutarse `on_post_page`, que tampoco existe todavía).

- [ ] **Step 7: Implementar `build_snippet` y `on_post_page`**

Añade al final de `docs/docs-tooling/mkdocs/notas_lectura.py`:

```python
def build_snippet(js_text: str, css_text: str) -> str:
    return f"<style>{css_text}</style>\n<script>{js_text}</script>\n"


def on_post_page(output, page, config):
    if not notes_enabled():
        return output
    snippet = build_snippet(
        JS_PATH.read_text(encoding="utf-8"), CSS_PATH.read_text(encoding="utf-8")
    )
    return output.replace("</body>", snippet + "</body>", 1)
```

- [ ] **Step 8: Ejecutar las pruebas y comprobar que pasan**

Run: `uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project pytest docs/docs-tooling/mkdocs/tests -q`
Expected: PASS (5 passed).

- [ ] **Step 9: Commit**

```bash
git add docs/docs-tooling/mkdocs/notas_lectura.py docs/docs-tooling/mkdocs/tests/test_notas_lectura.py
git commit -m "Anade el hook que decide cuando inyectar el boton de notas de lectura"
```

---

### Task 2: El botón, el picker de elementos y su conexión a `docs-serve`

**Files:**
- Create: `docs/docs-tooling/mkdocs/notas_lectura.js`
- Create: `docs/docs-tooling/mkdocs/notas_lectura.css`
- Create: `docs/docs-tooling/notas-pendientes.md`
- Modify: `docs/docs-tooling/mkdocs/mkdocs.yml`
- Modify: `Makefile`

**Interfaces:**
- Consumes: `notas_lectura.JS_PATH`, `notas_lectura.CSS_PATH`, `notas_lectura.on_post_page`
  (Task 1); ahora esos ficheros existen de verdad, así que `on_post_page` los lee tal cual.
- Produces: el sitio construido con `BISO_DOCS_SERVE=1` contiene el marcador
  `id="notas-lectura-button"`; construido sin esa variable, no lo contiene. Esta comprobación
  la usa Task 3 para verificar en el navegador y Task 4 para revisar el conjunto.

- [ ] **Step 1: Escribir el CSS del botón, el resaltado y el panel**

Crea `docs/docs-tooling/mkdocs/notas_lectura.css`:

```css
/* Reading-notes button, panel and element-picker overlay. Injected only by notas_lectura.py
   during `mkdocs serve`; never part of a `mkdocs build` output. */

#notas-lectura-button {
  position: fixed;
  right: 1.2rem;
  bottom: 1.2rem;
  z-index: 10000;
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  font-size: 1.3rem;
  background: var(--md-primary-fg-color);
  color: var(--md-primary-bg-color);
  box-shadow: var(--md-shadow-z2);
}

#notas-lectura-button.is-picking {
  background: var(--md-accent-fg-color);
}

body.notas-lectura-picking,
body.notas-lectura-picking * {
  cursor: crosshair !important;
}

#notas-lectura-overlay {
  position: fixed;
  z-index: 9999;
  pointer-events: none;
  border: 2px solid var(--md-accent-fg-color);
  border-radius: 0.2rem;
  display: none;
}

#notas-lectura-panel {
  position: fixed;
  right: 1.2rem;
  bottom: 4.6rem;
  z-index: 10000;
  width: min(24rem, calc(100vw - 2.4rem));
  padding: 0.9rem;
  border-radius: 0.4rem;
  background: var(--md-default-bg-color);
  color: var(--md-default-fg-color);
  box-shadow: var(--md-shadow-z2);
  display: none;
  font-size: 0.7rem;
}

#notas-lectura-panel.is-open {
  display: block;
}

#notas-lectura-panel .notas-lectura-field {
  margin-bottom: 0.6rem;
}

#notas-lectura-panel label {
  display: block;
  font-weight: bold;
  margin-bottom: 0.2rem;
}

#notas-lectura-panel textarea,
#notas-lectura-panel input {
  width: 100%;
  box-sizing: border-box;
  font: inherit;
  color: inherit;
  background: transparent;
  border: 1px solid var(--md-default-fg-color--lightest);
  border-radius: 0.2rem;
  padding: 0.3rem;
}

#notas-lectura-panel .notas-lectura-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

#notas-lectura-panel button {
  font: inherit;
  cursor: pointer;
  border-radius: 0.2rem;
  border: none;
  padding: 0.35rem 0.7rem;
}

#notas-lectura-panel .notas-lectura-cancel {
  background: transparent;
  color: var(--md-default-fg-color--light);
}

#notas-lectura-panel .notas-lectura-copy {
  background: var(--md-primary-fg-color);
  color: var(--md-primary-bg-color);
}
```

- [ ] **Step 2: Escribir el JS del botón, el picker y el panel**

Crea `docs/docs-tooling/mkdocs/notas_lectura.js`:

```javascript
// Reading-notes button: an element picker (like a browser inspector) that copies a
// ready-to-paste line for docs/docs-tooling/notas-pendientes.md to the clipboard.
// Injected only during `mkdocs serve` by notas_lectura.py; never ships in `mkdocs build`.
(function () {
  "use strict";

  var BLOCK_SELECTOR =
    "p, li, h1, h2, h3, h4, h5, h6, pre, blockquote, table, img, dt, dd";

  var picking = false;
  var highlighted = null;
  var picked = null;

  var button, overlay, panel, refField, quoteField, noteField, copyButton;

  function init() {
    if (document.getElementById("notas-lectura-button")) {
      return;
    }

    button = document.createElement("button");
    button.id = "notas-lectura-button";
    button.type = "button";
    button.title = "Anadir nota de lectura";
    button.textContent = "📝";
    button.addEventListener("click", toggleForButton);

    overlay = document.createElement("div");
    overlay.id = "notas-lectura-overlay";

    panel = buildPanel();

    document.body.appendChild(button);
    document.body.appendChild(overlay);
    document.body.appendChild(panel);

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("scroll", onScroll, true);
    document.addEventListener("click", onDocumentClick, true);
    document.addEventListener("keydown", onKeyDown);
  }

  function buildPanel() {
    var el = document.createElement("div");
    el.id = "notas-lectura-panel";

    el.innerHTML =
      '<div class="notas-lectura-field">' +
      "<label>Referencia</label>" +
      '<input type="text" id="notas-lectura-ref" readonly>' +
      "</div>" +
      '<div class="notas-lectura-field">' +
      "<label>Cita</label>" +
      '<input type="text" id="notas-lectura-quote">' +
      "</div>" +
      '<div class="notas-lectura-field">' +
      "<label>Nota</label>" +
      '<textarea id="notas-lectura-note" rows="3"></textarea>' +
      "</div>" +
      '<div class="notas-lectura-actions">' +
      '<button type="button" class="notas-lectura-cancel">Cancelar</button>' +
      '<button type="button" class="notas-lectura-copy">Copiar</button>' +
      "</div>";

    refField = el.querySelector("#notas-lectura-ref");
    quoteField = el.querySelector("#notas-lectura-quote");
    noteField = el.querySelector("#notas-lectura-note");
    copyButton = el.querySelector(".notas-lectura-copy");

    el.querySelector(".notas-lectura-cancel").addEventListener("click", closePanel);
    copyButton.addEventListener("click", copyNote);

    return el;
  }

  function toggleForButton() {
    if (picking) {
      stopPicking();
      return;
    }
    if (panel.classList.contains("is-open")) {
      closePanel();
      return;
    }
    startPicking();
  }

  function startPicking() {
    picking = true;
    button.classList.add("is-picking");
    document.body.classList.add("notas-lectura-picking");
  }

  function stopPicking() {
    picking = false;
    highlighted = null;
    button.classList.remove("is-picking");
    document.body.classList.remove("notas-lectura-picking");
    overlay.style.display = "none";
  }

  function onMouseMove(event) {
    if (!picking) {
      return;
    }
    var block = event.target.closest ? event.target.closest(BLOCK_SELECTOR) : null;
    if (!block) {
      overlay.style.display = "none";
      highlighted = null;
      return;
    }
    highlighted = block;
    positionOverlay(block);
  }

  function onScroll() {
    if (picking && highlighted) {
      positionOverlay(highlighted);
    }
  }

  function positionOverlay(el) {
    var rect = el.getBoundingClientRect();
    overlay.style.display = "block";
    overlay.style.left = rect.left + "px";
    overlay.style.top = rect.top + "px";
    overlay.style.width = rect.width + "px";
    overlay.style.height = rect.height + "px";
  }

  function onDocumentClick(event) {
    if (event.target === button || button.contains(event.target)) {
      return;
    }
    if (!picking) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    if (highlighted) {
      pick(highlighted);
    }
  }

  function onKeyDown(event) {
    if (picking && event.key === "Escape") {
      stopPicking();
    }
  }

  function pick(el) {
    picked = el;
    stopPicking();
    openPanel();
  }

  function openPanel() {
    refField.value = buildReference();
    quoteField.value = picked ? collapseWhitespace(picked.textContent) : "";
    noteField.value = "";
    panel.classList.add("is-open");
    noteField.focus();
  }

  function closePanel() {
    panel.classList.remove("is-open");
    picked = null;
  }

  function buildReference() {
    var ref = window.location.pathname;
    var activeLink = document.querySelector(".md-nav--secondary .md-nav__link--active");
    if (activeLink) {
      var href = activeLink.getAttribute("href") || "";
      var hashIndex = href.indexOf("#");
      if (hashIndex !== -1) {
        ref += href.slice(hashIndex);
      }
    }
    return ref;
  }

  function collapseWhitespace(text) {
    return text.replace(/\s+/g, " ").trim();
  }

  function copyNote() {
    var line = buildLine();
    navigator.clipboard.writeText(line).then(function () {
      var original = copyButton.textContent;
      copyButton.textContent = "Copiado";
      setTimeout(function () {
        copyButton.textContent = original;
        closePanel();
      }, 900);
    });
  }

  function buildLine() {
    var quote = quoteField.value.trim();
    var note = noteField.value.trim();
    var line = "- [ ] " + refField.value;
    if (quote) {
      line += ": «" + quote + "»";
    }
    line += " | " + note;
    return line;
  }

  if (window.document$) {
    window.document$.subscribe(init);
  } else {
    document.addEventListener("DOMContentLoaded", init);
  }
})();
```

- [ ] **Step 3: Registrar el hook en `mkdocs.yml`**

En `docs/docs-tooling/mkdocs/mkdocs.yml`, dentro de la clave `hooks:` ya existente:

```yaml
hooks:
  - excluir_superpowers_del_buscador.py
  - notas_lectura.py
```

- [ ] **Step 4: Hacer que `docs-serve` exporte la variable de entorno**

En `Makefile`, cambia la línea del target `docs-serve`:

```makefile
docs-serve: ## Sirve la documentacion en local con recarga automatica al editar
	BISO_DOCS_SERVE=1 $(UV_DOCS) mkdocs serve -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml
```

- [ ] **Step 5: Crear el fichero de notas, vacío salvo su cabecera**

Crea `docs/docs-tooling/notas-pendientes.md`:

```markdown
<!--
Notas de lectura pendientes de procesar. Una linea por nota:
  - [ ] <ruta>[#<ancla>][: «<cita>»] | <nota>
Se genera con el boton flotante del sitio servido en local (make docs-serve). Al procesar una
nota, marca su casilla o borra la linea.
-->
```

- [ ] **Step 6: Verificar que `mkdocs build` (sin la variable) no incluye el botón**

Run:
```bash
rm -rf /tmp/biso-notas-build-sin
uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
  mkdocs build --strict -f docs/docs-tooling/mkdocs/mkdocs.yml -d /tmp/biso-notas-build-sin
grep -rl "notas-lectura-button" /tmp/biso-notas-build-sin || echo "ausente, como se espera"
```
Expected: `ausente, como se espera` (el `grep -rl` no encuentra nada y devuelve 1, así que se
ejecuta el `echo` del `||`).

- [ ] **Step 7: Verificar que `mkdocs build` con la variable puesta sí lo incluye**

Run:
```bash
rm -rf /tmp/biso-notas-build-con
BISO_DOCS_SERVE=1 uv run --with-requirements docs/docs-tooling/mkdocs/docs-requirements.txt --no-project \
  mkdocs build --strict -f docs/docs-tooling/mkdocs/mkdocs.yml -d /tmp/biso-notas-build-con
grep -l "notas-lectura-button" /tmp/biso-notas-build-con/index.html
```
Expected: imprime la ruta del fichero (coincidencia encontrada).

- [ ] **Step 8: Commit**

```bash
git add docs/docs-tooling/mkdocs/notas_lectura.js docs/docs-tooling/mkdocs/notas_lectura.css \
  docs/docs-tooling/mkdocs/mkdocs.yml Makefile docs/docs-tooling/notas-pendientes.md
git commit -m "Implementa el boton de notas de lectura y lo conecta a docs-serve"
```

---

### Task 3: Verificación en el navegador

**Files:** ninguno (task de verificación, sin cambios de código).

**Interfaces:**
- Consumes: el sitio servido por `make docs-serve` con el botón de Task 2.

- [ ] **Step 1: Arrancar `docs-serve` en segundo plano**

Run (en segundo plano, puerto por defecto 8000):
```bash
make docs-serve
```

- [ ] **Step 2: Navegar con Claude in Chrome a una página con varias secciones**

Abre `http://127.0.0.1:8000/spec/cmd/ls/` (o la que esté sirviendo el puerto real).

- [ ] **Step 3: Comprobar el botón y el modo de selección**

Comprueba visualmente: el botón circular aparece en la esquina inferior derecha; al pulsarlo,
mover el ratón sobre un párrafo o un bloque de código lo resalta con un contorno.

- [ ] **Step 4: Comprobar que un clic abre el panel con los datos correctos**

Haz clic sobre un párrafo bajo un encabezado conocido. Comprueba que el panel muestra en
"Referencia" la ruta de la página con el ancla de esa sección, y en "Cita" el texto de ese
párrafo.

- [ ] **Step 5: Comprobar el botón "Copiar"**

Escribe una nota de prueba, pulsa "Copiar", y comprueba (leyendo el portapapeles con
`navigator.clipboard.readText()` desde la consola del navegador, o pegando en cualquier campo
de texto) que la línea copiada sigue el formato `- [ ] ruta#ancla: «cita» | nota`.

- [ ] **Step 6: Comprobar Escape y el cierre sin acción**

Pulsa el botón para entrar en modo de selección y pulsa Escape: el resaltado desaparece y no se
abre ningún panel. Repite y pulsa "Cancelar" en el panel: se cierra sin copiar nada.

- [ ] **Step 7: Parar `docs-serve`**

Detén el proceso en segundo plano.

Si cualquiera de estas comprobaciones falla, vuelve a Task 2 y corrige antes de continuar: no
hay un test automático que cubra esta parte, así que esta verificación manual es la única red
de seguridad del picker y el panel.

---

### Task 4: Revisión con un agente y cierre

**Files:** ninguno directamente; puede tocar cualquiera de los anteriores si la revisión pide
cambios.

- [ ] **Step 1: Pedir una revisión de todo el diff de la rama a un agente fresco**

Usa la skill `code-review` (o despacha un agente `general-purpose` con el diff completo de
`worktree-notas-lectura-docs` frente a `main`) pidiendo revisión de correctitud y de que se
cumple la spec `docs/superpowers/specs/2026-09-12-notas-lectura-docs-design.md`.

- [ ] **Step 2: Aplicar los hallazgos que procedan**

Corrige lo que la revisión confirme como problema real; si algo se descarta, anota por qué.

- [ ] **Step 3: Volver a ejecutar las pruebas y el build tras cualquier cambio**

Run: `uv run --with-requirements docs/docs-tooling/tools/requirements.txt --no-project pytest docs/docs-tooling/mkdocs/tests -q`
Run: `make docs-doctor`
Expected: ambos en verde.

- [ ] **Step 4: Cerrar la tarea de Backlog.md (TASK-31)**

Sigue `backlog instructions task-finalization`: comprueba cada criterio de aceptación con
evidencia, márcalos, y mueve TASK-31 a su estado terminal.

- [ ] **Step 5: Mezclar a `main` y borrar el worktree**

Desde el worktree, sin abrir PR (pedido explícitamente por el usuario): mezcla
`worktree-notas-lectura-docs` en `main` y borra el worktree.
