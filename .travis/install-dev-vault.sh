#!/bin/sh
set -o errexit

# Based on https://www.vaultproject.io/docs/platform/k8s/helm/run
helm repo add hashicorp https://helm.releases.hashicorp.com
helm search repo hashicorp/vault -l
helm install vault hashicorp/vault \
    --set "server.dev.enabled=true"
sleep 5 # give the vault pod chance to start
kubectl get pods
kubectl wait pod vault-0 --for condition=ready --timeout=60s
kubectl logs vault-0
