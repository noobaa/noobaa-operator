#!/bin/sh
set -o errexit

echo "💬 Install NooBaa CRD"
./build/_output/bin/noobaa-operator-local crd create

echo "💬 Create NooBaa operator deployment"
./build/_output/bin/noobaa-operator-local operator --operator-image=$OPERATOR_IMAGE install
sleep 5
kubectl wait pod -l noobaa-operator  --for condition=ready --timeout=60s

echo "💬 Deploy TLS Vault"
./.travis/deploy-validate-vault.sh # borrowed from rook

# Vault api address
api_address=$(kubectl logs vault-0| grep "Api Address"  | awk '{print $3}')
echo API_ADDRESS=$api_address >> $GITHUB_ENV