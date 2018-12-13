#!/bin/sh
set -euo pipefail

set -x

# retry solution discovered here: https://unix.stackexchange.com/a/137639

function fail {
  echo $1 >&2
  exit 1
}

function retry {
  local n=1
  local max=5
  local delay=5
  while true; do
    "$@" && break || {
      if test "$n" -lt "$max" ; then
        echo "Command failed. Attempt $n/$max. Waiting for $(($delay*$n)) seconds before retrying."
        sleep $(($delay*$n));
        n=$((n+1))
      else
        fail "The command has failed after $n attempts."
      fi
    }
  done
}

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

retry git fetch -q --force --update-head-ok "${BRIGADE_REMOTE_URL}" "${refspec}"

retry git checkout -q --force "${BRIGADE_COMMIT_REF}"

if [ "${BRIGADE_SUBMODULES:=}" = "true" ]; then
    retry git submodule update --init --recursive
fi
