#!/bin/sh
set -euo pipefail
set -x

: "${VCS_LOCAL_PATH:=/src}"
: "${VCS_REPO:?}"
: "${VCS_REVISION:?}"
: "${INIT_GIT_SUBMODULES:?}"
: "${GITHUB_PULL_REQUEST:=}"
: "${GITHUB_REF:=}"
: "${GITHUB_REF_TYPE:=}"

if [ "$GITHUB_PULL_REQUEST" ]; then
    git clone --depth=50 "${VCS_REPO}" "${VCS_LOCAL_PATH}"
    cd "${VCS_LOCAL_PATH}"
    git fetch origin "+refs/pull/${GITHUB_PULL_REQUEST}/merge:"
    git checkout -qf "${VCS_REVISION}"
elif [ "$GITHUB_REF" ]; then
    git clone --depth=50 --branch="${GITHUB_REF}" "${VCS_REPO}" "${VCS_LOCAL_PATH}"
    cd "${VCS_LOCAL_PATH}"
    if [ "$GITHUB_REF_TYPE" = "tag" ]; then
        git checkout -qf "${GITHUB_REF}"
    else
        git checkout -qf "${VCS_REVISION}"
    fi
else
    if sha=$(git ls-remote --exit-code "${VCS_REPO}" "${VCS_REVISION}" | cut -f1); then
      VCS_REVISION="${sha}"
    fi

    git clone --depth=50 "${VCS_REPO}" "${VCS_LOCAL_PATH}"
    cd "${VCS_LOCAL_PATH}"

    git fetch origin "${VCS_REVISION}"
    git checkout -qf FETCH_HEAD
    git reset --hard -q "${VCS_REVISION}"
fi

if "${INIT_GIT_SUBMODULES}" == "true"; then
    git submodule update --init --recursive
fi
