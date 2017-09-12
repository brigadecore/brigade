#!/bin/sh
set -e #o pipefail

# From the SSH key, generate the known hosts file
if [ "" != "${BRIGADE_REPO_KEY}" ]; then
  mkdir -p $HOME/.ssh
  KEY="$HOME/.ssh/id_rsa"
  echo ${BRIGADE_REPO_KEY} | tr '$' '\n' > $KEY
  chmod 600 $KEY
  ssh-keygen -y -f $HOME/.ssh/id_rsa > $HOME/.ssh/id_rsa.pub
  echo "github.com ssh-rsa $(cat $HOME/.ssh/id_rsa.pub)" >> $HOME/.ssh/known_hosts
  echo "bitbucket.com ssh-rsa $(cat $HOME/.ssh/id_rsa.pub)" >> $HOME/.ssh/known_hosts
  # Might need to add more.
fi

# If we want to force SSH:
# git config --global url."ssh://git@github.com".insteadOf "https://github.com" || true

git clone $VCS_REPO $VCS_LOCAL_PATH
cd $VCS_LOCAL_PATH
git checkout $VCS_REVISION
