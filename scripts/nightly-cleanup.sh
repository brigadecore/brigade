#!/bin/bash

set -eou pipefail

az login --service-principal --username $AZ_USERNAME --password $AZ_PASSWORD --tenant $AZ_TENANT
repo_list=$(az acr repository list --name unstablebrigade -o table | tail -n +3)

for repo in ${repo_list}
do
    az acr repository delete --name unstablebrigade --repository ${repo} --yes
done