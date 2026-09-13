// Reading-notes button: an element picker (like a browser inspector) that saves a
// ready-to-process line to notas-pendientes.md, at the repository root, through the
// /__notas_lectura__/save endpoint that notas_lectura.py adds to the dev server.
// Injected only during `mkdocs serve` by notas_lectura.py; never ships in `mkdocs build`.
(function () {
  "use strict";

  var BLOCK_SELECTOR =
    "p, li, h1, h2, h3, h4, h5, h6, pre, blockquote, table, img, dt, dd";
  var SAVE_PATH = "/__notas_lectura__/save";

  var picking = false;
  var highlighted = null;
  var picked = null;

  var button, overlay, panel, refField, quoteField, noteField, saveButton;

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
      '<textarea id="notas-lectura-note" rows="5"></textarea>' +
      "</div>" +
      '<div class="notas-lectura-actions">' +
      '<button type="button" class="notas-lectura-cancel">Cancelar</button>' +
      '<button type="button" class="notas-lectura-save">Guardar</button>' +
      "</div>";

    refField = el.querySelector("#notas-lectura-ref");
    quoteField = el.querySelector("#notas-lectura-quote");
    noteField = el.querySelector("#notas-lectura-note");
    saveButton = el.querySelector(".notas-lectura-save");

    el.querySelector(".notas-lectura-cancel").addEventListener("click", closePanel);
    saveButton.addEventListener("click", saveNote);

    return el;
  }

  function toggleForButton() {
    if (picking) {
      stopPicking();
      return;
    }
    if (panel.classList.contains("is-open")) {
      hidePanel();
      return;
    }
    if (picked) {
      showPanel();
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
    showPanel();
  }

  function showPanel() {
    panel.classList.add("is-open");
    focusNoteField();
  }

  function hidePanel() {
    panel.classList.remove("is-open");
  }

  function focusNoteField() {
    window.requestAnimationFrame(function () {
      noteField.focus();
    });
  }

  function closePanel() {
    hidePanel();
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

  function saveNote() {
    var line = buildLine();
    var original = saveButton.textContent;

    fetch(SAVE_PATH, { method: "POST", body: line })
      .then(function (response) {
        if (!response.ok) {
          throw new Error("save failed with status " + response.status);
        }
        saveButton.textContent = "Guardado";
        setTimeout(function () {
          saveButton.textContent = original;
          closePanel();
        }, 900);
      })
      .catch(function () {
        copyToClipboardAsFallback(line, original);
      });
  }

  function copyToClipboardAsFallback(line, original) {
    navigator.clipboard.writeText(line).then(
      function () {
        saveButton.textContent = "No se pudo guardar, copiado";
        setTimeout(function () {
          saveButton.textContent = original;
          closePanel();
        }, 1800);
      },
      function () {
        saveButton.textContent = "Error al guardar";
        setTimeout(function () {
          saveButton.textContent = original;
        }, 1500);
      }
    );
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
