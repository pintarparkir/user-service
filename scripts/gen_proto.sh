#!/usr/bin/env bash
# Generate Go code from .proto files via buf.
# Prereqs: buf v1.32+, protoc-gen-go, protoc-gen-go-grpc on PATH.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/api/proto"

echo "→ buf lint"
buf lint

echo "→ buf generate"
buf generate

echo "✓ proto code regenerated under api/proto/**/*.pb.go"
