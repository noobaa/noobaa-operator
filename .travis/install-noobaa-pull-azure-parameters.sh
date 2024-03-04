#!/bin/sh
set -o errexit


# TODO: Replace it with azure key vault URL once we have Azure key vault
# account is created
echo AZURE_VAULT_URL="https://noobaa-vault.vault.azure.net/" >> $GITHUB_ENV

echo "💬 Install NooBaa CRD"
./build/_output/bin/noobaa-operator-local crd create

echo "💬 Create NooBaa operator deployment"
./build/_output/bin/noobaa-operator-local operator --operator-image=$OPERATOR_IMAGE install
