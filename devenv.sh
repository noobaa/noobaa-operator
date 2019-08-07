# source me to your dev shell
export GOROOT="$(go env GOROOT)"
export GO111MODULE=on
alias nb="build/_output/bin/noobaa-operator-local"

if minikube status &> /dev/null
then
  eval $(minikube docker-env)
else
  echo "WARNING: minikube is not started - cannot change docker-env"
fi
