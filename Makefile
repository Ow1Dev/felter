SHELL := /usr/bin/env bash

# Load environment variables from .env if it exists.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

API_ADDR ?= :8080
WEB_DIR := web

.PHONY: fieldservice userservice proxy migrate web fmt tidy vet lint test fmt-check up down

init:
	@cd $(WEB_DIR) && bun install

web:
	@cd $(WEB_DIR) && bun start

fieldservice:
	@go build -o build/fieldservice ./cmd/fieldservice && PORT=$${API_ADDR#:} ./build/fieldservice

userservice:
	@go build -o build/userservice ./cmd/userservice && ./build/userservice

proxy:
	@go build -o build/proxy ./cmd/proxy && ./build/proxy

migrate:
	@go build -o build/migrate ./cmd/migrate && ./build/migrate

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
