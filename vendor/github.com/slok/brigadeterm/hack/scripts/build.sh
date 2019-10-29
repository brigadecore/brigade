#!/usr/bin/env bash

set -o errexit
set -o nounset

src=./cmd/brigadeterm
out=./bin/brigadeterm

goarch=amd64

if [ $ostype == 'Linux' ]; then 
    echo "Building linux release..."
    goos=linux
    binary_ext=-linux-amd64
elif [ $ostype == 'Darwin' ]; then 
    echo "Building darwin release..."
    goos=darwin
    binary_ext=-darwin-amd64
elif [ $ostype == 'Windows' ]; then 
    echo "Building windows release..."
    goos=windows
    binary_ext=-windows-amd64.exe
else
    echo "ostype env var required"
    exit 1
fi

final_out=${out}${binary_ext}
ldf_cmp="-w -extldflags '-static'"
f_ver="-X main.Version=${VERSION:-dev}"

echo "Building binary at ${final_out}"
GOOS=${goos} GOARCH=${goarch} CGO_ENABLED=0 go build -o ${final_out} --ldflags "${ldf_cmp} ${f_ver}"  ${src}