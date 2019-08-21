#!/bin/bash

PACKAGE_NAME=noobaa-operator
PACKAGE_VERSION=1.1.0
VENV=./build/_output/venv
OPERATOR_DIR=./build/_output/olm
QUAY_LOGIN_FILE=$HOME/.quay/login.json

[ -f $QUAY_LOGIN_FILE ] || { 
    echo "----> No quay login file ($QUAY_LOGIN_FILE)"
    echo "Lets create it so you don't have to lookup your credentials next time ..."
    echo -n "Enter quay username: "
    read USERNAME
    echo -n "Enter quay password: "
    read -s PASSWORD
    echo
    echo "{ \"user\": { \"username\": \"$USERNAME\", \"password\": \"$PASSWORD\" } }" > $QUAY_LOGIN_FILE
    echo "Written $QUAY_LOGIN_FILE."
}

echo "----> Logging in to quay ..."
QUAY_NAMESPACE=$(jq -r .user.username $QUAY_LOGIN_FILE)
QUAY_TOKEN=$(curl -H "Content-Type: application/json" -X POST https://quay.io/cnr/api/v1/users/login -sd @$HOME/.quay/login.json | jq -r .token)

echo "----> Preparing operator-courier ..."
python3 -m venv $VENV || exit 1
. $VENV/bin/activate
pip3 install operator-courier || exit 1

echo "----> Running operator-courier verify ..."
operator-courier verify --ui_validate_io "$OPERATOR_DIR" || exit 1

echo "----> Deleting quay release $QUAY_NAMESPACE/$PACKAGE_NAME/$PACKAGE_VERSION"
curl -X DELETE \
    -H "Content-Type: application/json" \
    -H "Authorization: $QUAY_TOKEN" \
    https://quay.io/cnr/api/v1/packages/$QUAY_NAMESPACE/$PACKAGE_NAME/$PACKAGE_VERSION/helm

echo "----> Pushing "$OPERATOR_DIR" to quay $QUAY_NAMESPACE/$PACKAGE_NAME/$PACKAGE_VERSION"
operator-courier push \
    "$OPERATOR_DIR" \
    "$QUAY_NAMESPACE" \
    "$PACKAGE_NAME" \
    "$PACKAGE_VERSION" \
    "$QUAY_TOKEN" || exit 1

echo "----> Install OLM ..."
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/olm.yaml

echo "----> Install the Operator Marketplace ..."
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/01_namespace.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/02_catalogsourceconfig.crd.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/03_operatorsource.crd.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/04_service_account.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/05_role.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/06_role_binding.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/07_upstream_operatorsource.cr.yaml
kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/08_operator.yaml

echo "----> Create the OperatorSource ..."
yq write deploy/olm-catalog/manual/operator-source.yaml spec.registryNamespace $QUAY_NAMESPACE | kubectl apply -f -

echo "----> Create my-noobaa-operator namespace ..."
kubectl create ns my-noobaa-operator

echo "----> Create OperatorGroup ..."
kubectl apply -f deploy/olm-catalog/manual/operator-group.yaml

echo "----> Create Subscription ..."
kubectl apply -f deploy/olm-catalog/manual/operator-subscription.yaml

echo "----> Done. (for now)"
