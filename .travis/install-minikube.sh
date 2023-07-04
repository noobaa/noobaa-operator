#!/bin/bash


export PS4='\e[36m+ ${FUNCNAME:-main}@${BASH_SOURCE}:${LINENO} \e[0m'
set -e # exit immediately when a command fails
set -o pipefail # only exit with zero if all commands of the pipeline exit successfully
set -u # error on unset variables
set -x # print each command before executing it

#
# NOTE: This script was originally copied from the Cojaeger-operator build
# https://github.com/jaegertracing/jaeger-operator/blob/master/.travis/setupMinikube.sh

# socat is needed for port forwarding
sudo apt-get update && sudo apt-get install socat && sudo apt-get install conntrack

MINIKUBE_DEBUG=""
#MINIKUBE_DEBUG="--alsologtostderr --v=5"
export MINIKUBE_VERSION=v1.30.1
export KUBERNETES_VERSION=v1.27.3
export CRICTL_VERSION=v1.27.0 # check latest version in /releases page

sudo mount --make-rshared /
sudo mount --make-rshared /proc
sudo mount --make-rshared /sys

curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz --output crictl-${CRICTL_VERSION}-linux-amd64.tar.gz && \
    sudo tar zxvf crictl-$CRICTL_VERSION-linux-amd64.tar.gz -C /usr/local/bin && \
    rm -f crictl-$CRICTL_VERSION-linux-amd64.tar.gz

curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x kubectl &&  \
    sudo mv kubectl /usr/local/bin/

curl -Lo minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
    chmod +x minikube &&  \
    sudo mv minikube /usr/local/bin/minikube

mkdir "${HOME}"/.kube || true
# touch "${HOME}"/.kube/config

# minikube config
minikube config set WantNoneDriverWarning false
minikube config set vm-driver none

minikube version
sudo minikube start --kubernetes-version=$KUBERNETES_VERSION ${MINIKUBE_DEBUG}
# sudo chown -R travis: /home/travis/.minikube/

minikube update-context || true

# Following is just to demo that the kubernetes cluster works.
kubectl cluster-info
# Wait for kube-dns to be ready.
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lk8s-app=kube-dns -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for kube-dns to be available"; kubectl get pods --all-namespaces; done
