#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

uuid="$(uuidgen)"
name="brigade-worker-${uuid,,}"

namespace="default"
commit_ref="master"
commit_id="9c75584920f1297008118915024927cc099d5dcc"
event_provider="github"
event_type="push"
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
payload="{}"
script="_cache/github.com/deis/empty-testbed/brigade.js"

while (($# > 0)); do
  case "$1" in
    --namespace) namespace="$2";  shift ;;
    --script)    script="$2";     shift ;;
    -*)          echo "Unrecognized command line argument $1" ;;
    *)           break;
  esac
  shift
done

base64=(base64)
if [[ "$(uname)" != "Darwin" ]]; then
  base64+=(-w 0)
fi

cat <<EOF | kubectl --namespace ${namespace} create -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${name}
  labels:
    heritage: brigade
    project: ${project_id}
    build: ${uuid}
    component: build
type: "brigade.sh/build"
data:
  revision:
    commit: $("${base64[@]}" <<<"${commit_id}")
    ref: $("${base64[@]}" <<<"${commit_ref}")
  event_provider: $("${base64[@]}" <<<"${event_provider}")
  event_type: $("${base64[@]}" <<<"${event_type}")
  project_id: $("${base64[@]}" <<<"${project_id}")
  build_id: $("${base64[@]}" <<<"${uuid}")
  payload: $("${base64[@]}" <<<"${payload}")
  script: $("${base64[@]}" <"${script}")
EOF
