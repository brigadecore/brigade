#!/bin/bash

current_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ -n ${TRAVIS_TAG} ]]; then
    echo "Tag ${TRAVIS_TAG}. building releases..."

    archs=( Linux Darwin Windows )
    for arch in "${archs[@]}"
    do
        VERSION=${TRAVIS_TAG} ostype=${arch} ${current_dir}/build.sh
    done
else
    echo "no tag, skipping release..."
fi