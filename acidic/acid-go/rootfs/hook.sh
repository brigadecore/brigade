#!/bin/sh

dest=${DEST_PATH:-src}

echo "Clone ${SSH_URL}#${HEAD_COMMIT_ID} into ${dest}"

git clone $SSH_URL $dest
cd $dest
git checkout $HEAD_COMMIT_ID

ls -lah /hook/data
. /hook/data/main.sh
