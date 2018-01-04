#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace=${NAMESPACE:-default}

event_provider="simple-event"
event_type="my_event"
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit="master"
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
    commit: ${commit}
    component: build
type: Opaque
data:
  commit: $(echo -n "${commit}" | base64 -w 0)
  event_provider: $(echo -n "${event_provider}" | base64 -w 0)
  event_type: $(echo -n "${event_type}" | base64 -w 0)
  project_id: $(echo -n "${project_id}" | base64 -w 0)
  build_id: $(echo -n "${uuid}" | base64 -w 0)
  payload: $(echo -n "${payload}" | base64 -w 0)
  script: $(echo -n "${script}" | base64 -w 0)
EOF
