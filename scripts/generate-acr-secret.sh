#!/usr/bin/env bash

SECNAME="acid-registry"
REGISTRY="acidic.azurecr.io"
REG_USER=$(echo ${REGISTRY} | docker-credential-osxkeychain get | jq -r .Username)
REG_SECRET=$(echo ${REGISTRY} | docker-credential-osxkeychain get | jq -r .Secret )

# I don't believe this is used on ACR
EMAIL=example@example.com

kubectl create secret docker-registry $SECNAME \
  --docker-server=$REGISTRY \
  --docker-username=$REG_USER \
  --docker-password="${REG_SECRET}" \
  --docker-email=$EMAIL

