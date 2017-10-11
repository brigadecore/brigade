#!/bin/sh
extra=""

if [ "" != "${BRIGADE_REPO_KEY}" ]; then
  KEY="./id_dsa"
  echo ${BRIGADE_REPO_KEY} | sed 's/\$/\n/g' > $KEY
  chmod 600 $KEY
  extra="-i $KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
fi

ssh $extra $@
