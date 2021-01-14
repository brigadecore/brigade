#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

tempdir=$(mktemp -d)

export BRIGADE_WORKSPACE="${tempdir}/repo"

check_equal() {
  [[ "$1" == "$2" ]] || {
    echo >&2 "Check failed: '$1' == '$2' ${3:+ ($3)}"
    exit 1
  }
}

cleanup() {
  rm "${tempdir}/event.json"
}
trap 'cleanup' EXIT

test_clone() {
  local revision="$1" want="$2"
  local eventjson="${tempdir}/event.json"

  local key="ref"
  if [[ ${#revision} == 40 ]]; then
    key="commit"
  elif [[ ${#revision} == 0 ]]; then
    # switch key to something bogus so that no ref or commit is included
    key="foo"
  fi

  jq -n \
    --arg ref "${revision}" \
    --arg cloneURL "https://github.com/brigadecore/empty-testbed.git" \
    '{worker: {git: {'"${key}"': $ref, cloneURL: $cloneURL}}}' > "${eventjson}"

  cat "${eventjson}" | jq

  ../bin/git-initializer \
    -p "${eventjson}" \
    -w "${BRIGADE_WORKSPACE}"

  got="$(git -C ${BRIGADE_WORKSPACE} rev-parse --short FETCH_HEAD)"

  check_equal "${want}" "${got}"

  rm -rf "${BRIGADE_WORKSPACE}"
}

echo ":: Checkout sha"
test_clone "99f3efa2b70c370d4ee0833c213c085a6ec146ab" "99f3efa"
echo

echo ":: Checkout tag"
test_clone "v0.1.0" "ddff78a"
echo

echo ":: Checkout branch"
test_clone "hotfix" "589e150"
echo

echo ":: Checkout pull request by reference"
test_clone "refs/pull/1/head" "5c4bc10"
echo

echo "All tests passing"
