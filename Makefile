DOCS_TOOLING := docs/docs-tooling
UV_DOCS := uv run --with-requirements $(DOCS_TOOLING)/mkdocs/docs-requirements.txt --no-project

.PHONY: docs-serve docs-build docs-doctor

docs-serve:
	$(UV_DOCS) mkdocs serve -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml

docs-build:
	$(UV_DOCS) python $(DOCS_TOOLING)/tutorial/generate.py
	$(UV_DOCS) mkdocs build --strict -f $(DOCS_TOOLING)/mkdocs/mkdocs.yml

docs-doctor:
	$(UV_DOCS) python $(DOCS_TOOLING)/tools/doctor.py
