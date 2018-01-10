#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

tempdir=$(mktemp -d -t "repos")

export VCS_LOCAL_PATH="${tempdir}/repo"

check_equal() {
  [[ "$1" == "$2" ]] || {
    echo >&2 "Check failed: '$1' == '$2' ${3:+ ($3)}"
    exit 1
  }
}

setup_git_server() {
  git daemon --base-path="${tempdir}" --export-all --reuseaddr "${tempdir}" >/dev/null 2>&1 &

  (
  unset XDG_CONFIG_HOME
  export HOME=/dev/null
  export GIT_CONFIG_NOSYSTEM=1

  mkdir -p "${tempdir}/test.git"
  cd "${tempdir}/test.git"

  git init --bare
  git remote add origin git@github.com:Azure/brigade.git
  git config --add remote.origin.fetch '+refs/heads/*:refs/heads/*'
  git config --add remote.origin.fetch '+refs/pull/1/head:refs/pull/1/head'
  git config --add remote.origin.fetch '+refs/tags/v0.7.0:refs/tags/v0.7.0'
  git fetch -q origin

  # create a branch
  echo "f6c8b7aa3166d390852f0d34638424ca90e8ed6b" > ./refs/heads/hotfix
  )
}

cleanup() {
  pkill -9 git-daemon >/dev/null 2>&1
  rm -rf "${tempdir}"
}
trap 'cleanup' EXIT

test_clone() {
  local revision="$1" sha="$2"

  VCS_REPO="git://127.0.0.1/test.git" VCS_REVISION="${revision}" ./rootfs/clone.sh

  check_equal "${sha}" "$(git -C ${VCS_LOCAL_PATH} rev-parse FETCH_HEAD)"

  rm -rf "${VCS_LOCAL_PATH}"
}

setup_git_server

echo ":: Checkout tag"
test_clone "v0.7.0" "f16f26c22a2057dad44749a6b279e9d6453df9a5"
echo

echo ":: Checkout sha"
test_clone "f16f26c22a2057dad44749a6b279e9d6453df9a5" "f16f26c22a2057dad44749a6b279e9d6453df9a5"
echo

echo ":: Checkout branch"
test_clone "hotfix" "f6c8b7aa3166d390852f0d34638424ca90e8ed6b"
echo

echo ":: Checkout pull request by sha"
test_clone "e397f07262a7027aa2f6083c4883aba197c57196" "e397f07262a7027aa2f6083c4883aba197c57196"

echo
echo "All tests passing"
