#!/usr/bin/env sh

set -e

if [ -z ${IMAGE_VERSION} ]; then
    echo "IMAGE_VERSION env var needs to be set"
    exit 1
fi

REPOSITORY="slok/"
IMAGE="brigadeterm"


docker build \
    --build-arg operator=${OPERATOR} \
    -t ${REPOSITORY}${IMAGE}:${IMAGE_VERSION} \
    -f ./docker/prod/Dockerfile .