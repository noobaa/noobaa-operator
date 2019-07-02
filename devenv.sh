# source me to your dev shell
export GOROOT=$(go env GOROOT)
export GO111MODULE=on
export PATH=$PATH:build/_output/bin
eval $(minikube docker-env)
