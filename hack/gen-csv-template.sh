#!/bin/bash

set -e

TMP_VERSION="9999.9999.9999"
CSV_DIR="deploy/olm-catalog/noobaa-operator/$TMP_VERSION"
CSV_FILE="$CSV_DIR/noobaa-operator.v9999.9999.9999.clusterserviceversion.yaml"
CRD_DIR="deploy/crds"
OUT_DIR=build/_output/templates
CSV_OUT_FILE="$OUT_DIR/noobaa-operator.vVERSION.clusterserviceversion.yaml.in"

mkdir -p $OUT_DIR
operator-sdk olm-catalog gen-csv --csv-version=$TMP_VERSION

mv $CSV_FILE $CSV_OUT_FILE
cp -R $CRD_DIR $OUT_DIR/
rm -rf $CSV_DIR
# Make csv-version and image variables

sed -i "s/${TMP_VERSION}/{{.NoobaaOperatorCsvVersion}}/g" $CSV_OUT_FILE
sed -i "s/image\:.*/image: {{.NoobaaOperatorImage}}/g" $CSV_OUT_FILE
