#!/bin/sh
extra=""

if [ "" != "${BRIGADE_REPO_KEY}" ]; then
  KEY="./id_dsa"
  printf "%s" "$BRIGADE_REPO_KEY" | sed 's/\$/\n/g' > $KEY

# checking for presence of the ssh certificate
# see https://github.blog/2019-08-14-ssh-certificate-authentication-for-github-enterprise-cloud/ for more details
  if [ "" != "${BRIGADE_REPO_SSH_CERT}" ]; then
    CERT="./id_dsa-cert.pub"
    printf "%s" "$BRIGADE_REPO_SSH_CERT" | sed 's/\$/\n/g' > $CERT
  fi

  chmod 600 id_dsa*

  extra="-i $KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
fi

ssh $extra $@
