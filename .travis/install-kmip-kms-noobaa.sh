#!/bin/bash
set -o errexit

container=$(docker run -d \
    -e "DOCKER_STEPCA_INIT_NAME=The Authority" \
    -e "DOCKER_STEPCA_INIT_DNS_NAMES=localhost,$(hostname -f)" \
    smallstep/step-ca)
echo "💬 Generate certificates in container $container"
while [ "$(docker inspect -f {{.State.Running}} $container)" != "true" ]; do
    echo "💬 Wait for container $container to start running, state running $(docker inspect -f {{.State.Running}} $container)"
    echo "💬 container $container full inspect $(docker inspect $container)"
    echo "💬 container $container logs $(docker logs $container)"
    sleep 1
done

while [ "$(docker exec $container step ca health)" != "ok" ]; do
    echo "💬 Wait for CA's health to become ok, actual $(docker exec $container step ca health)"
    sleep 1
done

docker exec $container step ca certificate --provisioner-password-file /home/step/secrets/password localhost localhost.crt localhost.key
docker exec $container step ca certificate --provisioner-password-file /home/step/secrets/password kmip.ciphertrustmanager.local kmip.ciphertrustmanager.local.crt kmip.ciphertrustmanager.local.key

echo "💬 Build and deploy $PYKMIP_IMAGE"
pykmp_certs=./pkg/util/kms/test/kmip/pykmip/certs
rm -rf $pykmp_certs; mkdir -p $pykmp_certs
docker cp $container:/home/step/certs/root_ca.crt $pykmp_certs/ca.crt
docker cp $container:/home/step/kmip.ciphertrustmanager.local.crt $pykmp_certs/server.crt
docker cp $container:/home/step/kmip.ciphertrustmanager.local.key $pykmp_certs/server.key
docker build ./pkg/util/kms/test/kmip/pykmip -t $PYKMIP_IMAGE --progress=plain
docker push $PYKMIP_IMAGE
kubectl create -f pkg/util/kms/test/kmip/pykmip/k8s


secret_certs=$(mktemp -d)
echo "💬 Create the kms cert secret for NooBaa in $secret_certs"
docker cp $container:/home/step/certs/root_ca.crt $secret_certs/CA_CERT
docker cp $container:/home/step/localhost.crt $secret_certs/CLIENT_CERT
docker cp $container:/home/step/localhost.key $secret_certs/CLIENT_KEY
secret=kms-kmip-certs
kubectl create secret generic $secret \
  --from-file=$secret_certs
kubectl get secret $secret -o yaml

echo "💬 Remove temp cert directories and step ca container"
rm -rf $pykmp_certs $secret_certs
docker stop $container; docker rm $container

# Pass endpoint, certs secret in env to the running test
echo KMIP_ENDPOINT="kmip:5696" >> $GITHUB_ENV
echo KMIP_CERTS_SECRET=$secret >> $GITHUB_ENV

echo "💬 Install NooBaa CRD"
./build/_output/bin/noobaa-operator-local crd create

echo "💬 Create NooBaa operator deployment"
./build/_output/bin/noobaa-operator-local operator --operator-image=$OPERATOR_IMAGE install
