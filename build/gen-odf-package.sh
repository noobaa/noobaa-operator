#!/bin/bash

if [ "${MANIFESTS}" == "" ] || [ "${CSV_NAME}" == "" ] || [ "${CORE_IMAGE}" == "" ] || [ "${DB_IMAGE}" == "" ] || [ "${OPERATOR_IMAGE}" == "" ]
then
  echo "gen-odf-package.sh: not all required envs were supplied"
  exit 1
fi

./build/_output/bin/noobaa-operator-local olm catalog -n openshift-storage --dir ${MANIFESTS} --odf --csv-name ${CSV_NAME} --noobaa-image ${CORE_IMAGE} --db-image ${DB_IMAGE} --operator-image ${OPERATOR_IMAGE}

temp_csv=$(mktemp)

grep -v "status: {}" ${MANIFESTS}/${CSV_NAME} > ${temp_csv}
cat >> ${temp_csv} << EOF
  relatedImages:
  - image: ${CORE_IMAGE}
    name: noboaa-core
  - image: ${DB_IMAGE}
    name: noobaa-db
  - image: ${OPERATOR_IMAGE}
    name: noobaa-operator
EOF

cp -f ${temp_csv} ${MANIFESTS}/${CSV_NAME}



