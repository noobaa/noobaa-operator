#!/bin/bash

echo "Installing jq ..."
sudo apt-get install jq

echo "Installing yq ..."
curl -LO 'https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64'
sudo chmod +x yq_linux_amd64
sudo mv yq_linux_amd64 /usr/local/bin/yq
yq --version || exit 1
