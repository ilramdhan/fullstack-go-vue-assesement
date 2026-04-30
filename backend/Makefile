# Variables
GO          ?= go
OPENAPI     := ../openapi.yaml
GEN_PKG     := internal/openapigen
BIN_NAME    := mygolangapp

.DEFAULT_GOAL := help

.PHONY: help dep tool-openapi openapi-gen gen-secret run build clean

help:
	@echo "Available targets:"
	@echo "  make dep          - install Go dependencies"
	@echo "  make tool-openapi - install oapi-codegen tool"
	@echo "  make openapi-gen  - generate types + chi server from ../openapi.yaml"
	@echo "  make gen-secret   - generate JWT_SECRET (helper)"
	@echo "  make run          - run the app locally (uses .env if present)"
	@echo "  make build        - build binary into ./bin/$(BIN_NAME)"
	@echo "  make clean        - remove build artifacts and local sqlite db"

dep:
	@$(GO) mod tidy

tool-openapi:
	@$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1

openapi-gen: tool-openapi
	@echo "generating openapi types + chi server (oapi-codegen v2)"
	@oapi-codegen -generate "types,chi-server,spec" -package openapigen -o $(GEN_PKG)/openapi.gen.go $(OPENAPI)

gen-secret:
	@$(GO) run ./script/gen-secret/main.go

run:
	@CGO_ENABLED=1 $(GO) run main.go

build:
	@mkdir -p bin
	@CGO_ENABLED=1 $(GO) build -o bin/$(BIN_NAME) main.go
	@echo "built ./bin/$(BIN_NAME)"

clean:
	@rm -rf bin data/*.db data/*.db-*
	@echo "cleaned"
