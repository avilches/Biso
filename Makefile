DOCS_TOOLING := docs/docs-tooling
UV_DOCS := uv run --with-requirements $(DOCS_TOOLING)/mkdocs/docs-requirements.txt --no-project

.PHONY: all docs-serve docs-build docs-doctor help

all: docs-build ## Genera todo: hoy solo la documentacion; cuando exista codigo Go, tambien el binario

docs-serve: ## Sirve la documentacion en local con recarga automatica al editar
	$(UV_DOCS) mkdocs serve -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml

docs-build: ## Regenera el tutorial y construye el sitio de documentacion en site/
	$(UV_DOCS) python $(DOCS_TOOLING)/tutorial/generate.py
	$(UV_DOCS) mkdocs build --strict -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml

docs-doctor: ## Comprueba enlaces, recuentos y referencias, y luego hace docs-build
	$(UV_DOCS) python $(DOCS_TOOLING)/tools/doctor.py

help: ## Lista los objetivos disponibles
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*##"}; {printf "  %-12s %s\n", $$1, $$2}'
