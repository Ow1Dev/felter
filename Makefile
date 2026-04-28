SHELL := /usr/bin/env bash

# Load environment variables from .env if it exists.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

API_ADDR ?= :8080
WEB_DIR := web

.PHONY: fieldservice userservice proxy web dev fmt tidy vet lint test fmt-check migrate up down

init:
	@cd $(WEB_DIR) && bun install

fieldservice:
	@PORT=$${API_ADDR#:} go run ./cmd/fieldservice

userservice:
	@go run ./cmd/userservice

proxy:
	@go run ./cmd/web

web:
	@cd $(WEB_DIR) && bun run start

dev:
	@go run ./cmd/web &
	@PORT=$${API_ADDR#:} go run ./cmd/fieldservice &
	cd $(WEB_DIR) && bun run start

migrate:
	@go run ./cmd/migrate

up:
	@docker compose up -d

down:
	@docker compose down

fmt:
	@gofumpt -w .

fmt-check:
	@diff -u <(echo -n) <(gofumpt -l . | sed '/^$/d') && echo "format OK" || (echo "Run 'make fmt' to format" && exit 1)

tidy:
	@go mod tidy

vet:
	@go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Run inside nix shell: 'nix develop'"; exit 2; }
	@golangci-lint run

test:
	@go test ./... -race -count=1
