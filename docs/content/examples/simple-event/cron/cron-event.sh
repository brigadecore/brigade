#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace=${NAMESPACE:-default}

event_provider="simple-event"
event_type="my_event"
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit_ref="master"
commit_id="589e15029e1e44dee48de4800daf1f78e64287c0"
uuid="$(uuidgen | tr '[:upper:]' '[:lower:]')"
name="simple-event-$uuid"

payload=$(date)
script=$(cat <<EOF
const { events } = require("brigadier");
events.on("my_event", (e) => {
  console.log("The system time is " + e.payload);
});
EOF
)

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
    commit: $(base64 -w 0 <<<"${commit_id}")
    ref: $(base64 -w 0 <<<"${commit_ref}")
  event_provider: $(base64 -w 0 <<<"${event_provider}")
  event_type: $(base64 -w 0 <<<"${event_type}")
  project_id: $(base64 -w 0 <<<"${project_id}")
  build_id: $(base64 -w 0 <<<"${uuid}")
  payload: $(base64 -w 0 <<<"${payload}")
  script: $(base64 -w 0 <<<"${script}")
EOF
