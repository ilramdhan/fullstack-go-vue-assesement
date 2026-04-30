# Variables
go_bin ?= go
GO_PACKAGES ?= $(shell go list ./... | grep -v 'examples\|qtest\|mock\|config\|cmd')
OPENAPI=../openapi.yaml
GEN_PKG=internal/openapigen

# Default target
.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@echo "  make dep        	- install depedencies
	@echo "  make openapi-gen	- generate openapigen result"
	@echo "  make gen-secret	- generate JWT_SECRET"
	@echo "  make run        	- Run the app locally (with .env)"
	@echo "  make build      	- Build binary into ./bin/mygolangapp"

dep:
	@go mod tidy
	@go mod vendor

run:
	CGO_ENABLED=1 $(go_bin) run main.go

build:
	CGO_ENABLED=1 $(go_bin) build -o bin/mygolangapp main.go

tool-openapi:
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

openapi-gen: tool-openapi
	@echo "generating openapi types/server (oapi-codegen v2)"
	@oapi-codegen -generate "types,chi-server,spec" -package openapigen -o $(GEN_PKG)/openapi.gen.go $(OPENAPI)

gen-secret:
	@echo "🔐 Generating JWT secret..."
	@go run ./script/gen-secret/main.go
