#!/bin/bash

# This script tests that the brigade-worker can execute the brigade.js in the
# brigade project. It is a convenient way to test the brigade.js before committing
# it.
#
# NOTE: It does not mock the payload.
#
# To use this script:
# - install a brigade-project pointing to Azure/brigade. That will generate a
#   secret with the BRIGADE_PROJECT_ID shown below
# - run this script

export BRIGADE_EVENT_TYPE="${BRIGADE_EVENT_TYPE:-push}"
export BRIGADE_EVENT_PROVIDER="${BRIGADE_EVENT_PROVIDER:-selftest}"
export BRIGADE_COMMIT_REF="${BRIGADE_COMMIT_REF:-master}"
export BRIGADE_PAYLOAD="${BRIGADE_PAYLOAD:-'{}'}"
export BRIGADE_PROJECT_ID="${BRIGADE_PROJECT_ID:-brigade-f1ccf2ad3495a1c68b82c78789061036240a9eb5ea2f0b0c4c20f9}"
export BRIGADE_PROJECT_NAMESPACE="${BRIGADE_PROJECT_NAMESPACE:-default}"

cd ../brigade-worker

# Load the project's brigade.js
export BRIGADE_SCRIPT="../brigade.js"

echo "running $BRIGADE_EVENT_TYPE on $BRIGADE_SCRIPT for $BRIGADE_PROJECT_ID"
yarn brigade
