#!/bin/sh
set -euo pipefail
set -x

: "${VCS_LOCAL_PATH:=/src}"
: "${VCS_REPO:?}"
: "${VCS_REVISION:?}"

if sha=$(git ls-remote --exit-code "${VCS_REPO}" "${VCS_REVISION}" | cut -f1); then
  VCS_REVISION="${sha}"
fi

git clone --depth=50 "${VCS_REPO}" "${VCS_LOCAL_PATH}"
cd "${VCS_LOCAL_PATH}"

git fetch origin "${VCS_REVISION}"
git checkout -qf FETCH_HEAD
git reset --hard -q "${VCS_REVISION}"
