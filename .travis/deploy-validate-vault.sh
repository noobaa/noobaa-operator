#!/usr/bin/env bash

# Copyright 2021 The Rook Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -exEuo pipefail

: "${KUBERNETES_AUTH:=true}"

#############
# VARIABLES #
#############
SERVICE=vault
NAMESPACE=default
VAULT_SA=noobaa
VAULT_POLICY_NAME=noobaa
SECRET_NAME=vault-server-tls
TMPDIR=$(mktemp -d)
#############
# FUNCTIONS #
#############

function create_secret_generic {
  kubectl create secret generic ${SECRET_NAME} \
  --namespace ${NAMESPACE} \
  --from-file=vault.key="${TMPDIR}"/vault.key \
  --from-file=vault.crt="${TMPDIR}"/vault.crt \
  --from-file=vault.ca="${TMPDIR}"/vault.ca

  # for noobaa operator
  kubectl create secret generic vault-ca-cert --namespace ${NAMESPACE} --from-file=cert="${TMPDIR}"/vault.ca
  kubectl create secret generic vault-client-cert --namespace ${NAMESPACE} --from-file=cert="${TMPDIR}"/vault.crt
  kubectl create secret generic vault-client-key --namespace ${NAMESPACE} --from-file=key="${TMPDIR}"/vault.key
}

function vault_helm_tls {

cat <<EOF >"${TMPDIR}/"custom-values.yaml
global:
  enabled: true
  tlsDisable: false

server:
  extraEnvironmentVars:
    VAULT_CACERT: /vault/userconfig/vault-server-tls/vault.ca

  extraVolumes:
  - type: secret
    name: vault-server-tls # Matches the ${SECRET_NAME} from above

  standalone:
    enabled: true
    config: |
      listener "tcp" {
        address = "[::]:8200"
        cluster_address = "[::]:8201"
        tls_cert_file = "/vault/userconfig/vault-server-tls/vault.crt"
        tls_key_file  = "/vault/userconfig/vault-server-tls/vault.key"
        tls_client_ca_file = "/vault/userconfig/vault-server-tls/vault.ca"
      }

      storage "file" {
        path = "/vault/data"
      }
EOF

}

function deploy_vault {
  # TLS config
  scriptdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  bash "${scriptdir}"/generate-tls-config.sh "${TMPDIR}" ${SERVICE} ${NAMESPACE}
  create_secret_generic
  vault_helm_tls

  # Install Vault with Helm
  helm repo add hashicorp https://helm.releases.hashicorp.com
  helm install vault hashicorp/vault --values "${TMPDIR}/"custom-values.yaml
  #sleep 5 # give vault's pod a chance
  #kubectl wait pod vault-0 --for condition=ready --timeout=300s
  until kubectl get pods -l app.kubernetes.io/name=vault --field-selector=status.phase=Running|grep vault-0; do sleep 5; done

  # Unseal Vault
  VAULT_INIT_TEMP_DIR=$(mktemp)
  kubectl exec -ti vault-0 -- vault operator init -format "json" -ca-cert /vault/userconfig/vault-server-tls/vault.crt | tee -a "$VAULT_INIT_TEMP_DIR"
  for i in $(seq 0 2); do
    kubectl exec -ti vault-0 -- vault operator unseal -ca-cert /vault/userconfig/vault-server-tls/vault.crt "$(jq -r ".unseal_keys_b64[$i]" "$VAULT_INIT_TEMP_DIR")"
  done
  kubectl get pods -l app.kubernetes.io/name=vault

  # Wait for vault to be ready once unsealed
  while [[ $(kubectl get pods -l app.kubernetes.io/name=vault -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting vault to be ready" && sleep 1; done

  # Configure Vault
  ROOT_TOKEN=$(jq -r '.root_token' "$VAULT_INIT_TEMP_DIR")
  kubectl exec -it vault-0 -- vault login -ca-cert /vault/userconfig/vault-server-tls/vault.crt "$ROOT_TOKEN"
  kubectl exec -ti vault-0 -- vault secrets enable -ca-cert /vault/userconfig/vault-server-tls/vault.crt -path=noobaa kv
  #kubectl exec -ti vault-0 -- vault secrets enable -ca-cert /vault/userconfig/vault-server-tls/vault.crt -path=noobaa kv-v2
  kubectl exec -ti vault-0 -- vault kv list -ca-cert /vault/userconfig/vault-server-tls/vault.crt noobaa || true # failure is expected

  # Configure Vault Policy for NooBaa
  echo '
  path "noobaa/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
  }
  path "sys/mounts" {
    capabilities = ["read"]
  }'| kubectl exec -i vault-0 -- vault policy write -ca-cert /vault/userconfig/vault-server-tls/vault.crt "$VAULT_POLICY_NAME" -

  # Configure Kubernetes auth
  if [[ "${KUBERNETES_AUTH}" == "true" ]]; then
    set_up_vault_kubernetes_auth
  else
    set_up_vault_token_auth
  fi
}

function set_up_vault_kubernetes_auth {
  # get the service account common.yaml created earlier
  VAULT_SA_SECRET_NAME=$(kubectl -n "$NAMESPACE" get sa "$VAULT_SA" -o jsonpath="{.secrets[*]['name']}")

  # Set SA_JWT_TOKEN value to the service account JWT used to access the TokenReview API
  SA_JWT_TOKEN=$(kubectl -n "$NAMESPACE" get secret "$VAULT_SA_SECRET_NAME" -o jsonpath="{.data.token}" | base64 --decode)

  # Set SA_CA_CRT to the PEM encoded CA cert used to talk to Kubernetes API
  SA_CA_CRT=$(kubectl -n "$NAMESPACE" get secret "$VAULT_SA_SECRET_NAME" -o jsonpath="{.data['ca\.crt']}" | base64 --decode)

  # get kubernetes endpoint
  # Point to the internal API server hostname
  # https://kubernetes.io/docs/tasks/run-application/access-api-from-pod/
  K8S_HOST=https://kubernetes.default.svc

  # enable kubernetes auth
  kubectl exec -ti vault-0 -- vault auth enable kubernetes

  # configure the kubernetes auth
  kubectl exec -ti vault-0 -- vault write auth/kubernetes/config \
    token_reviewer_jwt="$SA_JWT_TOKEN" \
    kubernetes_host="$K8S_HOST" \
    kubernetes_ca_cert="$SA_CA_CRT" \
    issuer="https://kubernetes.default.svc.cluster.local"

  # configure a role for noobaa
  kubectl exec -ti vault-0 -- vault write auth/kubernetes/role/noobaa \
    bound_service_account_names="$VAULT_SA" \
    bound_service_account_namespaces="$NAMESPACE" \
    policies="$VAULT_POLICY_NAME" \
    ttl=1440h
}

# Create a token for noobaa
function set_up_vault_token_auth {
  echo "💬 Get the vault token secret"
  TOKEN=$(kubectl exec vault-0 -- vault token create -policy=$VAULT_POLICY_NAME -format json -ca-cert /vault/userconfig/vault-server-tls/vault.crt|jq -r '.auth.client_token')
  echo Token=\"$TOKEN\"
  secret=vault-token-secret
  kubectl create secret generic $secret \
    --from-literal=token=$TOKEN
  kubectl get secret $secret -o yaml
  echo TOKEN_SECRET_NAME=$secret >> $GITHUB_ENV
}

########
# MAIN #
########

deploy_vault
