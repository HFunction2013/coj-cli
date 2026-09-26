#!/usr/bin/env bash
# 构建 coj CLI
set -euo pipefail
cd "$(dirname "$0")"

GO=${GO:-go}
OUT=${OUT:-coj-cli}

echo "==> 格式化"
$GO fmt ./...

echo "==> 静态检查"
$GO vet ./... || true

echo "==> 构建 $OUT"
$GO build -ldflags "-s -w" -o "$OUT" .

echo "==> 完成: $(pwd)/$OUT"
"./$OUT" version
