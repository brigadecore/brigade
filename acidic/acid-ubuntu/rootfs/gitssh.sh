#!/bin/sh
extra=""

if [ "" != "${ACID_REPO_KEY}" ]; then
  KEY="./id_dsa"
  echo ${ACID_REPO_KEY} | sed 's/\$/\n/g' > $KEY
  chmod 600 $KEY
  extra="-i $KEY -o StrictHostKeyChecking=no"
fi

ssh $extra $@
