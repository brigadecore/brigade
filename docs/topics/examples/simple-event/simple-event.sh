#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace="default"

event_provider="simple-event"
event_type="my_event"

# This is github.com/deis/empty-testbed
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit="master"

b64flags=""
uuidflags=""
system=$(uname)
if [[ $system != "Darwin" ]]; then
  b64flags="-w 0" # Turn off line wrapping
  uuidflags="-t"  # generate UUID v1 for sortability
fi


# This is the brigade script to execute
script=$(cat <<EOF
const { events } = require("brigadier");
events.on("my_event", (e) => {
  console.log("The system time is " + e.payload);
});
EOF
)

# Now we will generate a new event evrey 60 seconds.
while : ; do
  # We'll use a UUID instead of a ULID. But if you want a ULID generator, you
  # can grab one here: https://github.com/technosophos/ulid
  uuid="$(uuidgen $uuidflags | tr '[:upper:]' '[:lower:]')"

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
      commit: ${commit}
      component: build
  type: "brigade.sh/build"
  data:
    commit: $(echo -n "${commit}" | base64 $b64flags)
    event_provider: $(echo -n "${event_provider}" | base64 $b64flags)
    event_type: $(echo -n "${event_type}" | base64 $b64flags)
    project_id: $(echo -n "${project_id}" | base64 $b64flags)
    build_id: $(echo -n "${uuid}" | base64 $b64flags)
    payload: $(echo -n "${payload}" | base64 $b64flags)
    script: $(echo -n "${script}" | base64 $b64flags)
EOF
  sleep 60
done

