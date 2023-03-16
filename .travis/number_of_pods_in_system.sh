#!/bin/sh

# This script checks the number of pods in a given namespace.
# It waits until we have the minimum number of pods to start working on the system.
CURRENT_PODS_NUMBER=0
THESHOLD_PODS_NUMBER=5
NAMESPACE=default
NUMBER_OF_ITERATIONS=15 # arbitrary number

while true; do
    case "${1}" in
        --pods) THESHOLD_PODS_NUMBER=${2}
                       shift 2 ;;
        --namespace)  NAMESPACE=${2}
                        shift 2 ;;
    esac
    if [ -z ${1} ]; then
      break
    fi
done

echo "Check status of noobaa pods:"
echo "Namespace: ${NAMESPACE}"
echo "Number of pods to start using the system is ${THESHOLD_PODS_NUMBER}"

i=1
while [ "$i" -le ${NUMBER_OF_ITERATIONS} ];
do
    PODS_NUMBER=$(kubectl get pods -n ${NAMESPACE} --no-headers | wc -l)
    echo ${PODS_NUMBER}
    echo -ne "\033[1A\033[2K\033[1A" # move the cursor up 1 line and clear the line
    if [ ${PODS_NUMBER} -ge ${THESHOLD_PODS_NUMBER} ]; then
        echo "All needed pods were created!"
        break
    fi

    echo "waiting for noobaa pods ${PODS_NUMBER}/${THESHOLD_PODS_NUMBER}..."
    sleep 1
done
