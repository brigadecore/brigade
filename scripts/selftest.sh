#!/bin/bash

# This script tests that the brigade-worker can execute the brigade.js in the
# brigade project. It is a convenient way to test the brigade.js before committing
# it.
#
# NOTE: It does not mock the payload.
#
# To use this script:
# - install an brigade-project pointing to deis/brigade. That will generate a
#   secret with the BRIGADE_PROJECT_ID shown below
# - run this script

export BRIGADE_EVENT_TYPE=push
export BRIGADE_EVENT_PROVIDER=selftest
export BRIGADE_COMMIT=master
export BRIGADE_PAYLOAD='{}'
export BRIGADE_PROJECT_ID=brigade-c0670fe8e03fddf8455c0049eecca2457750291be807712ae9b84d
export BRIGADE_PROJECT_NAMESPACE=default

cd ../brigade-worker

# Load the project's brigade.js
export BRIGADE_SCRIPT="../brigade.js"

echo "running $BRIGADE_EVENT_TYPE on $BRIGADE_SCRIPT for $BRIGADE_PROJECT_ID"
yarn brigade
