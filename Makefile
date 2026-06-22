SHELL := /usr/bin/env bash

# Load environment variables from .env if it exists.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

WEB_DIR := web
COMPOSE_FILES := -f infra/dev/infra.compose.yml -f infra/dev/app.compose.yml

.PHONY: dev down-app migrate generate-api fmt tidy vet lint test fmt-check up down

init:
	@cd $(WEB_DIR) && bun install
	@docker compose --project-directory . -f infra/dev/infra.compose.yml up -d
	@for i in $$(seq 1 30); do docker compose --project-directory . -f infra/dev/infra.compose.yml exec postgres pg_isready -U felter > /dev/null 2>&1 && break || true; sleep 1; done
	@make migrate

generate-api:
	@./scripts/generate-go-api.sh
	@./scripts/generate-ts-api.sh

migrate:
	@go build -buildvcs=false -o build/cli ./cmd/cli && ./build/cli migrate

up:
	@docker compose --project-directory . -f infra/dev/infra.compose.yml up -d

down:
	@docker compose --project-directory . $(COMPOSE_FILES) down -v

dev:
	@docker compose --project-directory . $(COMPOSE_FILES) up --build -d

down-app:
	@docker compose --project-directory . -f infra/dev/app.compose.yml down

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
