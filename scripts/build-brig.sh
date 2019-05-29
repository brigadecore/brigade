#!/usr/bin/env bash

set -euo pipefail

oses="linux windows darwin"
archs="amd64"

for os in $oses; do
  for arch in $archs; do 
    echo "building $os-$arch";
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o ./bin/brig-$os-$arch ./brig/cmd/brig; \
  done; \
  if [ $os = 'windows' ]; then
    mv ./bin/brig-$os-$arch ./bin/brig-$os-$arch.exe; \
  fi; \
done
