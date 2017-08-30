#!/bin/bash

# This script tests that the acid-worker can execute the acid.js in the
# acid project. It is a convenient way to test the acid.js before committing
# it.
#
# NOTE: It does not mock the payload.
#
# To use this script:
# - install an acid-project pointing to deis/acid. That will generate a
#   secret with the ACID_PROJECT_ID shown below
# - run this script

export ACID_EVENT_TYPE=push
export ACID_EVENT_PROVIDER=selftest
export ACID_COMMIT=master
export ACID_PAYLOAD='{}'
export ACID_PROJECT_ID=acid-c0670fe8e03fddf8455c0049eecca2457750291be807712ae9b84d
export ACID_PROJECT_NAMESPACE=default

cd ../acid-worker

# Load the project's acid.js
export ACID_SCRIPT="../acid.js"

echo "running $ACID_EVENT_TYPE on $ACID_SCRIPT for $ACID_PROJECT_ID"
yarn acid
