#!/bin/bash

function usage() {
    echo ""
    echo "Usage: $0 [-h] <namespace>"
    echo ""
    echo "  -h       	Print this help"
    echo "  namepsace   namespace to remove"
    echo ""
    exit 0
}

function echoerr() {
    echo -e "$@" 1>&2
}

function sighandler() {
    [ -n "${TMP_DIR}" ] && [ -d ${TMP_DIR} ] && {
        rm -rf ${TMP_DIR}
    }
    return 0
}


set -eo pipefail

trap 'sighandler' EXIT

SCRIPT_NAME=$0
SCRIPT_PATH_NAME=$(cd ${0%/*} && echo $PWD/${0##*/})
SCRIPT_PATH=`dirname "${SCRIPT_PATH_NAME}"`

[ -z "$1" ] || [ "-h" = "$1" ] && {
    usage
}

kubectl get ns $1
echo "ns $1 exist"

echo "Removing ns $1"
kubectl delete --timeout=5s ns $1 > /dev/null 2>&1 || true

TMP_DIR=$(mktemp -d -t remove-ns)

echo "Finding related ns $1 objects. It may take a while"
kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n $1 > ${TMP_DIR}/namespace.objects
echo "Removing finalizers from the related objects"
cat ${TMP_DIR}/namespace.objects | egrep -v "packagemanifest|NAME" | awk '{print $1}' | 
while read res
do 
	echo "Removing finalizers from $res"
	kubectl patch -n $1 $res --type json --patch='[ { "op": "remove", "path": "/metadata/finalizers" } ]' || true
done

echo "Done"
