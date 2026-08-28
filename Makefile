.PHONY: help validate test check example

help:
	@echo "make validate  Valida catálogo, Recipes, referencias y documentos"
	@echo "make test      Ejecuta pruebas del resolver y curador"
	@echo "make check     Ejecuta validación y pruebas"
	@echo "make example   Resuelve el ejemplo school-requests sin escribir"

validate:
	python3 scripts/validate_platform.py

test:
	python3 -m unittest discover -s tests -v

check: validate test

example:
	./eng new --from examples/intakes/school-requests.json --output /tmp/school-requests-blueprint --dry-run
