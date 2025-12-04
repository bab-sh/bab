#!/bin/sh
set -e
mkdir -p schema
go run ./internal/schema/generate.go > schema/babfile.schema.json
echo "Generated schema/babfile.schema.json"
