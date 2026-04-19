SHELL := /usr/bin/env bash

API_ADDR ?= :8080
WEB_DIR := web

.PHONY: api web dev fmt tidy vet lint test fmt-check

init:
	@cd $(WEB_DIR) && bun install 

api:
	@PORT=$${API_ADDR#:} go run ./cmd/api

web:
	@cd $(WEB_DIR) && bun run start

dev:
	@PORT=$${API_ADDR#:} go run ./cmd/api & \
	cd $(WEB_DIR) && bun run start

fmt:
	@go fmt ./...

fmt-check:
	@diff -u <(echo -n) <(gofmt -l . | sed '/^$/d') && echo "format OK" || (echo "Run 'make fmt' to format" && exit 1)

tidy:
	@go mod tidy

vet:
	@go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Run inside nix shell: 'nix develop'"; exit 2; }
	@golangci-lint run

test:
	@go test ./... -race -count=1
