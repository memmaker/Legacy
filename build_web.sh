#!/bin/sh
set -e
cd "$(dirname "$0")"
OUT=../fx-games/site/legacy/play
mkdir -p "$OUT"
GOOS=js GOARCH=wasm go build -trimpath -ldflags '-s -w' -o "$OUT/legacy.wasm" .
R=$(go env GOROOT)
cp "$R/lib/wasm/wasm_exec.js" "$OUT/" 2>/dev/null || cp "$R/misc/wasm/wasm_exec.js" "$OUT/"
gzip -9 -k -f "$OUT/legacy.wasm"
ls -l "$OUT"
