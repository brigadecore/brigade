#!/bin/bash

az login --service-principal --username $AZ_USERNAME --password $AZ_PASSWORD --tenant $AZ_TENANT
repo_list=$(az acr repository list --name unstablebrigade)

# remove leading and trailing []
repo_list=${repo_list::-1}
clean_repo_list=${repo_list:1}

for repo in ${clean_repo_list}
do
    repo=$(echo $repo | tr -d '"')
    az acr repository delete --name unstablebrigade --repository ${repo} --yes
done