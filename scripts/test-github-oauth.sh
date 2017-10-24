#!/bin/bash

# This is a quick test you can use to make sure your GitHub Personal Access
# Token is correctly configured for talking to Azure/brigade.
# Usage: TOKEN=YOURTOKEN ./test-github-oauth.sh

PROJECT=https://api.github.com/repos/Azure/brigade

curl -i -H "Authorization: token ${TOKEN}" $PROJECT
