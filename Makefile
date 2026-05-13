SHELL := /usr/bin/env bash

# Load environment variables from .env if it exists.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

API_ADDR ?= :8080
WEB_DIR := web

.PHONY: dev dev-app down-app start-fieldservice start-userservice start-proxy start-web fieldservice userservice proxy migrate web fmt tidy vet lint test fmt-check up down

init:
	@cd $(WEB_DIR) && bun install
	@docker compose up -d
	@for i in $$(seq 1 30); do docker exec felter-postgres-1 pg_isready -U felter > /dev/null 2>&1 && break || true; sleep 1; done
	@make migrate

web:
	@cd $(WEB_DIR) && bun start

fieldservice:
	@go build -buildvcs=false -o build/fieldservice ./cmd/fieldservice && PORT=$${API_ADDR#:} ./build/fieldservice

userservice:
	@go build -buildvcs=false -o build/userservice ./cmd/userservice && ./build/userservice

proxy:
	@go build -buildvcs=false -o build/proxy ./cmd/proxy && ./build/proxy

migrate:
	@go build -buildvcs=false -o build/migrate ./cmd/migrate && ./build/migrate

up:
	@docker compose up -d

down:
	@docker compose down -v

dev: dev-app

dev-app:
	@process-compose up -f process-compose.yml -d

down-app:
	@process-compose down -f process-compose.yml

start-fieldservice:
	@process-compose up -f process-compose.yml -d fieldservice

start-userservice:
	@process-compose up -f process-compose.yml -d userservice

start-proxy:
	@process-compose up -f process-compose.yml -d proxy

start-web:
	@process-compose up -f process-compose.yml -d web

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
