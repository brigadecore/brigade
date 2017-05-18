#!/bin/sh
extra=""

if [[ "" != $ACID_REPO_KEY ]]; then
  extra="-i ${ACID_REPO_KEY}"
fi

echo ssh $extra $@
