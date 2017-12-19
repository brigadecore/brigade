#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

uuid="$(uuidgen)"
name="brigade-worker-${uuid,,}"

namespace="default"
commit="9c75584920f1297008118915024927cc099d5dcc"
event_provider="github"
event_type="push"
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
payload="{}"
script="_cache/github.com/deis/empty-testbed/brigade.js"

while (( "$#" > 0 )); do
  case "$1" in
    --namespace) namespace="$2";  shift ;;
    --script)    script="$2";     shift ;;
    -*)          echo "Unrecognized command line argument $1" ;;
    *)           break;
  esac
  shift
done

cat <<EOF | kubectl --namespace ${namespace} create -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${name}
  labels:
    heritage: brigade
    project: ${project_id}
    build: ${uuid}
    commit: ${commit}
    jobname: ${name}
    component: build
type: Opaque
data:
  commit: $(echo -n "${commit}" | base64)
  event_provider: $(echo -n "${event_provider}" | base64)
  event_type: $(echo -n "${event_type}" | base64)
  project_id: $(echo -n "${project_id}" | base64)
  build_id: $(echo -n "${uuid}" | base64)
  payload: $(echo -n "${payload}" | base64)
  script: $(base64 < "${script}")
EOF
