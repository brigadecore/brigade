#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

tempdir=$(mktemp -d -t "repos")

export BRIGADE_WORKSPACE="${tempdir}/repo"

check_equal() {
  [[ "$1" == "$2" ]] || {
    echo >&2 "Check failed: '$1' == '$2' ${3:+ ($3)}"
    exit 1
  }
}

setup_git_server() {
  local srvroot="${root_dir}/tmp"

  git ls-remote git://localhost/test.git >/dev/null 2>&1 || git daemon --base-path="${srvroot}" --export-all --reuseaddr "${srvroot}" >/dev/null 2>&1 &

  sleep 5

  [[ -d  "${srvroot}" ]] && return

  (
  unset XDG_CONFIG_HOME
  export HOME=/dev/null
  export GIT_CONFIG_NOSYSTEM=1

  repo_root="${root_dir}/tmp/test.git"

  git clone --mirror https://github.com/deis/empty-testbed.git "${repo_root}"
  )
}

cleanup() {
  pkill -9 git-daemon >/dev/null 2>&1
  rm -rf "${tempdir}"
}
trap 'cleanup' EXIT

test_clone() {
  local revision="$1" want="$2"

  BRIGADE_REMOTE_URL="git://127.0.0.1/test.git" BRIGADE_COMMIT_REF="${revision}" ./rootfs/clone.sh

  got="$(git -C ${BRIGADE_WORKSPACE} rev-parse --short FETCH_HEAD)"

  check_equal "${want}" "${got}"

  rm -rf "${BRIGADE_WORKSPACE}"
}

setup_git_server

echo ":: Checkout tag"
test_clone "v0.1.0" "ddff78a"
echo

echo ":: Checkout branch"
test_clone "hotfix" "589e150"
echo

echo ":: Checkout pull request by reference"
test_clone "refs/pull/1/head" "5c4bc10"
echo

echo ":: Checkout sha"
test_clone "589e15029e1e44dee48de4800daf1f78e64287c0" "589e150"
echo

echo "All tests passing"
