#!/bin/sh
extra=""

if [ "" != "${BRIGADE_REPO_KEY}" ]; then
  extra="-i ${BRIGADE_REPO_KEY} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
fi

ssh $extra $@
