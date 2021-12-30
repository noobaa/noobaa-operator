#!/bin/sh
set -o errexit

echo "💬 Set the kms token secret"
if [ -z "$IBM_SERVICE_API_KEY"  -o -z "$IBM_CUSTOMER_ROOT_KEY"  ]; then
    echo "Please define IBM KP API key and customer root key"
    exit 0
fi

secret=kms-token-secret
kubectl create secret generic $secret \
  --from-literal=IBM_SERVICE_API_KEY=$IBM_SERVICE_API_KEY \
  --from-literal=IBM_CUSTOMER_ROOT_KEY=$IBM_CUSTOMER_ROOT_KEY
echo TOKEN_SECRET_NAME=$secret >> $GITHUB_ENV

echo "💬 Install NooBaa CRD"
./build/_output/bin/noobaa-operator-local crd create

echo "💬 Create NooBaa operator deployment"
./build/_output/bin/noobaa-operator-local operator --operator-image=$OPERATOR_IMAGE install
