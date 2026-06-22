#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

WEB_DIR="web"

echo "Generating TypeScript API types..."

cd "$WEB_DIR"
bunx openapi-typescript ../docs/api/projectservice.yaml -o src/app/api/projectservice.ts
bunx openapi-typescript ../docs/api/fieldservice.yaml -o src/app/api/fieldservice.ts

echo "TypeScript API types generated."
