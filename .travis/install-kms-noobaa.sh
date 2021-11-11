#!/bin/sh
set -o errexit

echo "💬 Enable Vault's noobaa backend"
kubectl exec vault-0 -- vault secrets enable -path=noobaa kv

echo "💬 Set the kms token secret"
secret=kms-token-secret
kubectl create secret generic $secret \
  --from-literal=token=$(kubectl logs vault-0| grep "Root Token" | awk '{print $3}')
kubectl get secret $secret -o yaml

# vault api address
api_address=$(kubectl logs vault-0| grep "Api Address"  | awk '{print $3}')
echo API_ADDRESS=$api_address >> $GITHUB_ENV
echo TOKEN_SECRET_NAME=$secret >> $GITHUB_ENV

echo "💬 Install NooBaa CRD"
./build/_output/bin/noobaa-operator-local crd create

echo "💬 Create NooBaa operator deployment"
./build/_output/bin/noobaa-operator-local operator --operator-image=$OPERATOR_IMAGE install
