#!/usr/bin/env bash

set -euo pipefail

for os in $OSES; do
  for arch in $ARCHS; do 
    echo "building $os-$arch"
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
      go build \
      -ldflags "-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
      -o ../bin/brig-$os-$arch \
      ./cli
  done
  if [ $os = 'windows' ]; then
    mv ../bin/brig-$os-$arch ../bin/brig-$os-$arch.exe
  fi
done
