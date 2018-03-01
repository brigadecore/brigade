#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace="default"

event_provider="simple-event"
event_type="my_event"

# This is github.com/deis/empty-testbed
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit_ref="master"
commit_id="589e15029e1e44dee48de4800daf1f78e64287c0"

base64=(base64)
uuidgen=(uuidgen)
if [[ "$(uname)" != "Darwin" ]]; then
  base64+=(-w 0)
  uuidgen+=(-t) # generate UUID v1 for sortability
fi

# This is the brigade script to execute
script=$(cat <<EOF
const { events } = require("brigadier");
events.on("my_event", (e) => {
  console.log("The system time is " + e.payload);
});
EOF
)

# Now we will generate a new event every 60 seconds.
while :; do
  # We'll use a UUID instead of a ULID. But if you want a ULID generator, you
  # can grab one here: https://github.com/technosophos/ulid
  uuid="$("${uuidgen[@]}" | tr '[:upper:]' '[:lower:]')"

  # We can use the UUID to make sure we get a unique name
  name="simple-event-$uuid"

  # This will just print the system time for the system running the script.
  payload=$(date)

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
    script: $("${base64[@]}" <<<"${script}")
EOF
  sleep 60
done
