#!/bin/bash

if [ "${MANIFESTS}" == "" ] || [ "${CSV_NAME}" == "" ] || [ "${CORE_IMAGE}" == "" ] || [ "${PSQL_12_IMAGE}" == "" ] || [ "${DB_IMAGE}" == "" ] || [ "${OPERATOR_IMAGE}" == "" ] || [ "${COSI_SIDECAR_IMAGE}" == "" ]
then
  echo "gen-odf-package.sh: not all required envs were supplied"
  exit 1
fi

echo "--obc-crd=${OBC_CRD}"

./build/_output/bin/noobaa-operator-local olm catalog -n openshift-storage \
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
--obc-crd=${OBC_CRD} 

temp_csv=$(mktemp)

# remove status property and everything after it
status_line_number=$(grep -n "status:" ${MANIFESTS}/${CSV_NAME} | cut -f1 -d:)
n=$((status_line_number-1))
head -n ${n} ${MANIFESTS}/${CSV_NAME} > ${temp_csv}

# add relatedImages to the final CSV
cat >> ${temp_csv} << EOF
  relatedImages:
  - image: ${CORE_IMAGE}
    name: noobaa-core
  - image: ${DB_IMAGE}
    name: noobaa-db
  - image: ${PSQL_12_IMAGE}
    name: noobaa-psql-12
  - image: ${OPERATOR_IMAGE}
    name: noobaa-operator
EOF

cp -f ${temp_csv} ${MANIFESTS}/${CSV_NAME}



