#!/bin/bash

set -euo pipefail

mytmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'yourbase-treesitter')
trap 'rm -rf "$mytmpdir"' EXIT

git clone https://github.com/tree-sitter/tree-sitter.git "$mytmpdir/tree-sitter"
git -C "$mytmpdir/tree-sitter" checkout v0.20.0
git clone https://github.com/tree-sitter/tree-sitter-python.git "$mytmpdir/tree-sitter-python"
git -C "$mytmpdir/tree-sitter-python" checkout v0.19.0

go install modernc.org/ccgo/v3@v3.12.45

mkdir -p internal/lib
ccgo \
  -pkgname=lib \
  -export-defines '' \
  -export-enums '' \
  -export-externs X \
  -export-structs S \
  -export-fields '' \
  -export-typedefs '' \
  -trace-translation-units \
  -o "internal/lib/treesitter_$(go env GOOS)_$(go env GOARCH).go" \
  -I "$mytmpdir/tree-sitter/lib/src" \
  -I "$mytmpdir/tree-sitter/lib/include" \
  "$mytmpdir"/tree-sitter/lib/src/*.c

ccgo \
  -pkgname=lib \
  -export-defines '' \
  -export-enums '' \
  -export-externs X \
  -export-structs S \
  -export-fields '' \
  -export-typedefs '' \
  -trace-translation-units \
  -o "internal/python/python_$(go env GOOS)_$(go env GOARCH).go" \
  -I ./internal/lib \
  -I "$mytmpdir/tree-sitter/lib/include" \
  "$mytmpdir/tree-sitter-python/src/parser.c" \
  internal/python/scanner.c
