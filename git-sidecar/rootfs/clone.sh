#!/bin/sh
set -e #o pipefail
git clone $VCS_REPO $VCS_LOCAL_PATH
cd $VCS_LOCAL_PATH
git checkout $VCS_REVISION
