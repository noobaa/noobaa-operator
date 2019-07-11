# source me to your dev shell
export GOROOT=$(go env GOROOT)
export GO111MODULE=on
[[ ":${PATH}:" =~ ":build/_output/bin:" ]] || PATH=$PATH:build/_output/bin
eval $(minikube docker-env)
alias nb='noobaa-operator-local'
