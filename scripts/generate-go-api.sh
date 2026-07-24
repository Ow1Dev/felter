#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Generating Go API types..."

oapi-codegen --config internal/projectservice/api/oapi-codegen.yaml docs/api/projectservice.yaml
oapi-codegen --config internal/fieldservice/api/fieldvalue/oapi-codegen.yaml docs/api/fieldvalue.yaml
oapi-codegen --config internal/fieldservice/api/oapi-codegen.yaml docs/api/fieldservice.yaml

echo "Go API types generated."
