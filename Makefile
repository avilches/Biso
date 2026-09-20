DOCS_TOOLING := docs/docs-tooling
UV_DOCS := uv run --with-requirements $(DOCS_TOOLING)/mkdocs/docs-requirements.txt --no-project

BIN := bin/biso

.PHONY: all build test vet fmt-check check docs-serve docs-build docs-doctor help

all: build docs-build ## Genera todo: el binario y la documentacion

build: ## Compila el binario en bin/biso
	go build -o $(BIN) ./cmd/biso

test: ## Ejecuta la suite de Go con el detector de carreras
	go test -race ./...

vet: ## Ejecuta go vet sobre todo el codigo Go
	go vet ./...

fmt-check: ## Falla si algun fichero Go no esta formateado con gofmt
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "Sin formatear con gofmt:"; echo "$$out"; exit 1; fi

check: build vet fmt-check test docs-doctor ## Comprueba todo antes de dar algo por terminado

docs-serve: ## Sirve la documentacion en local con recarga automatica al editar
	BISO_DOCS_SERVE=1 $(UV_DOCS) mkdocs serve -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml -a 127.0.0.1:8100

docs-build: ## Regenera el tutorial y construye el sitio de documentacion en site/
	$(UV_DOCS) python $(DOCS_TOOLING)/tutorial/generate.py
	$(UV_DOCS) mkdocs build --strict -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml

docs-doctor: ## Comprueba enlaces, recuentos y referencias, y luego hace docs-build
	$(UV_DOCS) python $(DOCS_TOOLING)/tools/doctor.py

help: ## Lista los objetivos disponibles
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*##"}; {printf "  %-12s %s\n", $$1, $$2}'
