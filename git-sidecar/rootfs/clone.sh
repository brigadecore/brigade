#!/bin/sh
set -euo pipefail

set -x

# The Git SHA1 of the revision.
: "${BRIGADE_COMMIT_ID:=}"

# The Git reference.
#
# Also can be a full hex object name.
# master, v0.1.0, refs/pulls/1/head, 589e15029e1e44dee48de4800daf1f78e64287c0
# If not set, it will try BRIGADE_COMMIT_ID
: "${BRIGADE_COMMIT_REF:=${BRIGADE_COMMIT_ID}}"

# The working directory.
: "${BRIGADE_WORKSPACE:=/src}"

refspec="${BRIGADE_COMMIT_REF}"
if full_ref=$(git ls-remote --exit-code "${BRIGADE_REMOTE_URL}" "${BRIGADE_COMMIT_REF}" | cut -f2); then
  refspec="+${full_ref}:${full_ref}"
fi

git init -q "${BRIGADE_WORKSPACE}"
cd "${BRIGADE_WORKSPACE}"

git fetch -q --force --update-head-ok "${BRIGADE_REMOTE_URL}" "${refspec}"

# reset to $BRIGADE_COMMIT_ID or FETCH_HEAD
git reset -q --hard "${BRIGADE_COMMIT_ID:-FETCH_HEAD}"

git checkout -q --force "${BRIGADE_COMMIT_REF}"

if [ "${BRIGADE_SUBMODULES:=}" = "true" ]; then
    git submodule update --init --recursive
fi
