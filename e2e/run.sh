#!/bin/bash
# custom script for e2e testing

# kudos to https://elder.dev/posts/safer-bash/
set -o errexit # script exits when a command fails == set -e
set -o nounset # script exits when tries to use undeclared variables == set -u
#set -o xtrace # trace what's executed == set -x (useful for debugging)
set -o pipefail # causes pipelines to retain / set the last non-zero status

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

TEST_PROJECT_NAME=brigade-5b55ed522537b663e178f751959d234fd650d626f33f70557b2e82

echo "-----Creating a test project-----"
go run "${DIR}"/../brig/cmd/brig/main.go project create -f "${DIR}"/testproject.yaml -x

echo "-----Checking if the test project secret was created-----"
PROJECT_NAME=$(kubectl get secret -l app=brigade,component=project,heritage=brigade -o=jsonpath='{.items[0].metadata.name}' -n $BRIGADE_NAMESPACE)
if [ "$PROJECT_NAME" != "$TEST_PROJECT_NAME" ]; then
    echo "Wrong secret name. Expected $TEST_PROJECT_NAME, got $PROJECT_NAME"
    exit 1
fi

echo "-----Running a Build on the test project-----"
go run "${DIR}"/../brig/cmd/brig/main.go run e2eproject -f "${DIR}"/test.js

# get the worker pod name
WORKER_POD_NAME=$(kubectl get pod -l component=build,heritage=brigade,project=$TEST_PROJECT_NAME -o=jsonpath='{.items[0].metadata.name}' -n $BRIGADE_NAMESPACE)

# get the number of lines the expected log output appears
LOG_LINES=$(kubectl logs $WORKER_POD_NAME -n $BRIGADE_NAMESPACE | grep "==> handling an 'exec' event" | wc -l)

if [ $LOG_LINES != 1 ]; then
    echo "Did not find expected output on worker Pod logs"
    exit 1
fi
