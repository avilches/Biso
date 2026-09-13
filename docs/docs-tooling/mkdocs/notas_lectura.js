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
