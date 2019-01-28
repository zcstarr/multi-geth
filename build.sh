#!/usr/bin/env bash

mkdir -p build/bin
CGO_LDFLAGS="$GOPATH/src/github.com/etclabscore/sputnikvm-ffi/c/libsputnikvm.a -ldl -lm" go build -o build/bin/geth ./cmd/geth

