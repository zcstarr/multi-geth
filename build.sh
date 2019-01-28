#!/usr/bin/env bash

set -e

CGO_LDFLAGS="$GOPATH/src/github.com/etclabscore/sputnikvm-ffi/c/libsputnikvm.a -ldl -lm" go build -o build/bin/geth ./cmd/geth

