#!/bin/sh
extra=""

if [ "" != "${ACID_REPO_KEY}" ]; then
  extra="-i ${ACID_REPO_KEY} -o StrictHostKeyChecking=no"
fi

ssh $extra $@
