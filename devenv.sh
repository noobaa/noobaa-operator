# source me to your dev shell

export GO111MODULE=on
export GOPROXY="https://proxy.golang.org"
export GOROOT="$(go env GOROOT)"

alias nb="build/_output/bin/noobaa-operator-local"

if minikube status &> /dev/null
then
  eval $(minikube docker-env)
else
  echo "WARNING: minikube is not started - cannot change docker-env"
fi
