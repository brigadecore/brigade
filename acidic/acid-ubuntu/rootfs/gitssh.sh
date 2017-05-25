#!/bin/sh

dest=${DEST_PATH:-src}

url=$CLONE_URL
if [ "" != "${ACID_REPO_KEY}" ]; then
  url=$SSH_URL
fi

echo "Clone ${url}#${HEAD_COMMIT_ID} into ${dest}"

git clone $url $dest
cd $dest
git checkout $HEAD_COMMIT_ID

ls -lah /hook/data
. /hook/data/main.sh