#!/bin/bash
# This script builds an operator registry as desribed here:
#   https://github.com/operator-framework/operator-registry
# then it adds a catalog source to the cluster and create a subscription in order to install noobaa using OLM.

set -e # exit immediately when a command fails
set -o pipefail # only exit with zero if all commands of the pipeline exit successfully
set -u # error on unset variables
set -x # print each command before executing it

if [ -z ${1+x} ]
then
    echo "Usage: $0 <catalog-image> (but you probably want to use 'make test-olm')"
    exit 1
fi

CATALOG_IMAGE=$1
VERSION=$(go run cmd/version/main.go)

if [ -z "${OPERATOR_SDK}" ]; then
    OPERATOR_SDK=operator-sdk
fi

function install_olm() {
    echo "----> Install OLM and Operator Marketplace ..."
    ${OPERATOR_SDK} olm install --version 0.16.1 || true
    # kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/crds.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/olm.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/01_namespace.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/02_catalogsourceconfig.crd.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/03_operatorsource.crd.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/04_service_account.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/05_role.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/06_role_binding.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/07_upstream_operatorsource.cr.yaml
    # kubectl apply -f https://github.com/operator-framework/operator-marketplace/raw/master/deploy/upstream/08_operator.yaml
}

function create_catalog() {
    echo "----> Creating CatalogSource ..."
    # kubectl delete -n olm catalogsource operatorhubio-catalog
    yq write deploy/olm/catalog-source.yaml spec.image $CATALOG_IMAGE | kubectl apply -f -
    # yq eval ".spec.image = \"$CATALOG_IMAGE"\" deploy/olm/catalog-source.yaml
}

function create_subscription() {
    # echo "----> Create my-noobaa-operator namespace ..."
    # kubectl create ns my-noobaa-operator

    echo "----> Create OperatorGroup ..."
    kubectl apply -f deploy/olm/operator-group.yaml

    echo "----> Create Subscription ..."
    kubectl apply -f deploy/olm/operator-subscription.yaml
}

function wait_for_operator() {
    echo "----> Wait for CSV to be ready ..."
    while [ "$(kubectl get csv noobaa-operator.v$VERSION -o jsonpath={.status.phase})" != "Succeeded" ]
    do
        echo -n '.'
        sleep 1
    done

    echo "----> Wait for operator to be ready ..."
    # kubectl wait pod -l noobaa-operator=deployment --for condition=ready
    kubectl rollout status deployment noobaa-operator
}

function test_operator() {
    MINI_RESOURCES='{"requests":{"cpu":"10m","memory":"128Mi"}}'
    ${OPERATOR_SDK} run --local --operator-flags "system create --core-resources $MINI_RESOURCES --db-resources $MINI_RESOURCES --endpoint-resources ${MINI_RESOURCES}"
    while [ "$(kubectl get noobaa/noobaa -o jsonpath={.status.phase})" != "Ready" ]
    do
        echo -n '.'
        sleep 3
        ${OPERATOR_SDK} run --local --operator-flags "status"
    done
    ${OPERATOR_SDK} run --local --operator-flags "status"
}

function main() {
    echo "----> Starting OLM test"
    install_olm
    create_catalog
    create_subscription
    wait_for_operator
    test_operator
    echo "----> Finished OLM test"
}

main
