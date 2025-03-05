#!/bin/bash

missing_envs=""
[ "${MANIFESTS}" == "" ] && missing_envs="${missing_envs} MANIFESTS"
[ "${CSV_NAME}" == "" ] && missing_envs="${missing_envs} CSV_NAME"
[ "${CORE_IMAGE}" == "" ] && missing_envs="${missing_envs} CORE_IMAGE"
[ "${PSQL_12_IMAGE}" == "" ] && missing_envs="${missing_envs} PSQL_12_IMAGE"
[ "${DB_IMAGE}" == "" ] && missing_envs="${missing_envs} DB_IMAGE"
[ "${OPERATOR_IMAGE}" == "" ] && missing_envs="${missing_envs} OPERATOR_IMAGE"
[ "${COSI_SIDECAR_IMAGE}" == "" ] && missing_envs="${missing_envs} COSI_SIDECAR_IMAGE"

if [ "${missing_envs}" != "" ]
then
  echo "gen-odf-package.sh: missing required environment variables:${missing_envs}"
  exit 1
fi

echo "--obc-crd=${OBC_CRD}"

# Build the command base
cmd="./build/_output/bin/noobaa-operator-local olm catalog -n openshift-storage \
--dir ${MANIFESTS} \
--odf \
--csv-name ${CSV_NAME} \
--skip-range "${SKIP_RANGE}" \
--replaces "${REPLACES}" \
--noobaa-image ${CORE_IMAGE} \
--db-image ${DB_IMAGE} \
--psql-12-image ${PSQL_12_IMAGE} \
--operator-image ${OPERATOR_IMAGE} \
--cosi-sidecar-image ${COSI_SIDECAR_IMAGE} \
--obc-crd=${OBC_CRD}"

# Add CNPG flags only if CNPG_IMAGE is set and not empty
if [ -n "${CNPG_IMAGE}" ]; then
    cmd+=" --include-cnpg"
    cmd+=" --cnpg-image ${CNPG_IMAGE}"
fi


# Execute the command
eval $cmd

temp_csv=$(mktemp)

# remove status property and everything after it
status_line_number=$(grep -n "status:" ${MANIFESTS}/${CSV_NAME} | cut -f1 -d:)
n=$((status_line_number-1))
head -n ${n} ${MANIFESTS}/${CSV_NAME} > ${temp_csv}

cp -f ${temp_csv} ${MANIFESTS}/${CSV_NAME}



