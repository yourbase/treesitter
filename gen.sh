#!/bin/bash

set -euo pipefail

ensure_upstream() {
  local dest="upstream/$1"
  local url="$2"
  local rev="$3"

  if [[ -d "$dest" ]]; then
    git -C "$dest" fetch
  else
    git clone "$url" "$dest"
  fi
  git -C "$dest" checkout --force --quiet "$rev"
}

ensure_upstream tree-sitter https://github.com/tree-sitter/tree-sitter.git v0.20.0
ensure_upstream tree-sitter-python https://github.com/tree-sitter/tree-sitter-python.git v0.19.0

go install modernc.org/ccgo/v3@v3.12.45

# Verify that gcc isn't secretly clang. This is a problem on macOS.
CC="${CC:-gcc}"
CCGO_CPP="${CCGO_CPP:-cpp}"
if "$CC" --version 2> /dev/null | grep clang &> /dev/null || \
   "$CCGO_CPP" --version 2> /dev/null | grep clang &> /dev/null; then
  echo "$CC and/or $CCGO_CPP are actually clang." 1>&2
  echo "Set CC to a gcc compiler and CCGO_CPP to a gcc preprocessor." 1>&2
  echo 1>&2
  echo "On macOS:" 1>&2
  echo 'brew install gcc && '\\ 1>&2
  # shellcheck disable=SC2016
  echo 'export CC="$(which gcc-11)" && '\\ 1>&2
  # shellcheck disable=SC2016
  echo 'export CCGO_CPP="$(which cpp-11)"' 1>&2
  exit 1
fi

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
  -I upstream/tree-sitter/lib/src \
  -I upstream/tree-sitter/lib/include \
  upstream/tree-sitter/lib/src/*.c

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
  -I upstream/tree-sitter/lib/include \
  upstream/tree-sitter-python/src/parser.c \
  internal/python/scanner.c
