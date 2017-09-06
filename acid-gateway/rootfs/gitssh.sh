#!/bin/sh
extra=""

if [ "" != "${ACID_REPO_KEY}" ]; then
  extra="-i ${ACID_REPO_KEY} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
fi

ssh $extra $@
