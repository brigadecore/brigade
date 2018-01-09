#!/bin/sh
set -euo pipefail
set -x

: "${VCS_LOCAL_PATH:=/src}"
: "${VCS_REPO:?}"
: "${VCS_REVISION:?}"

git clone "${VCS_REPO}" "${VCS_LOCAL_PATH}"
cd "${VCS_LOCAL_PATH}"

git fetch origin "${VCS_REVISION}"
git checkout -qf FETCH_HEAD
git reset --hard -q "${VCS_REVISION}"
