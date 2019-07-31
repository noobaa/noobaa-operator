# source me to your dev shell
export GOROOT=$(go env GOROOT)
export GO111MODULE=on
[[ ":${PATH}:" =~ ":build/_output/bin:" ]] || PATH=$PATH:build/_output/bin

alias nb='noobaa-operator-local'

minikube status &> /dev/null
rc=$?
if [ ${rc} -eq 0 ]; then
  echo "minikube is started. setting docker env to work with minikube's docker"
  eval $(minikube docker-env)
else
  echo "minikube is not started. working with local docker"
fi

