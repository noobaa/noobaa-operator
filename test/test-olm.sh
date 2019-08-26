#!/bin/bash
# https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md
# https://github.com/operator-framework/community-operators/blob/master/scripts/ci/test-operator

PACKAGE_NAME=noobaa-operator
PACKAGE_VERSION=$(go run cmd/version/main.go)
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

if false
then
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
fi

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

if false
then
    echo "----> Create the OperatorSource ..."
    yq write deploy/olm-catalog/manual/operator-source.yaml spec.registryNamespace $QUAY_NAMESPACE | kubectl apply -f -

    echo "----> Create my-noobaa-operator namespace ..."
    kubectl create ns my-noobaa-operator

    echo "----> Create OperatorGroup ..."
    kubectl apply -f deploy/olm-catalog/manual/operator-group.yaml

    echo "----> Create Subscription ..."
    kubectl apply -f deploy/olm-catalog/manual/operator-subscription.yaml

    echo "----> Wait for CSV to be ready ..."
    while [ "$(kubectl get csv noobaa-operator.v$PACKAGE_VERSION -n marketplace -o jsonpath={.status.phase})" != "Succeeded" ]
    do
        echo -n '.'
        sleep 1
    done

    echo "----> Wait for operator to be ready ..."
    kubectl wait pod -n marketplace -l noobaa-operator=deployment --for condition=ready
fi

SCORECARD_NS=marketplace

global_manifest=$(mktemp)
cat deploy/crds/noobaa_v1alpha1_noobaa_crd.yaml >> $global_manifest
echo "---" >> $global_manifest
cat deploy/crds/noobaa_v1alpha1_backingstore_crd.yaml >> $global_manifest
echo "---" >> $global_manifest
cat deploy/crds/noobaa_v1alpha1_bucketclass_crd.yaml >> $global_manifest
echo "---" >> $global_manifest
cat deploy/obc/objectbucket_v1alpha1_obc_crd.yaml >> $global_manifest
echo "---" >> $global_manifest
cat deploy/obc/objectbucket_v1alpha1_ob_crd.yaml >> $global_manifest
echo "---" >> $global_manifest
cat deploy/cluster_role.yaml >> $global_manifest
echo "---" >> $global_manifest
yq write deploy/cluster_role_binding.yaml subjects.0.namespace $SCORECARD_NS >> $global_manifest
echo "---" >> $global_manifest
yq write deploy/obc/storage_class.yaml provisioner "noobaa.io/${SCORECARD_NS}.bucket" >> $global_manifest


echo "----> Run Scorecard tests ..."
operator-sdk scorecard \
    --namespace $SCORECARD_NS \
    --csv-path $OPERATOR_DIR/noobaa-operator.v${PACKAGE_VERSION}.clusterserviceversion.yaml \
    --crds-dir $OPERATOR_DIR \
    --global-manifest $global_manifest \
    --cr-manifest deploy/crds/noobaa_v1alpha1_noobaa_cr.yaml \
    --cr-manifest deploy/crds/noobaa_v1alpha1_backingstore_cr.yaml \
    --cr-manifest deploy/crds/noobaa_v1alpha1_bucketclass_cr.yaml

    # --olm-deployed \

rm $global_manifest

echo "----> Done. (for now)"
