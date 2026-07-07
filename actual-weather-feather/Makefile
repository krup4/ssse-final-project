.PHONY: deps deps-windows tidy

deps:
	./scripts/install_deps.sh

deps-windows:
	powershell -ExecutionPolicy Bypass -File ./scripts/install_deps.ps1

tidy:
	go mod tidy
