SHELL := /usr/bin/env bash

API_ADDR ?= :8080
WEB_DIR := web

.PHONY: api web dev fmt tidy

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

tidy:
	@go mod tidy
