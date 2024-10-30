#!/bin/bash
set -x

# This script aim is to assist users that deploy noobaa on AWS STS cluster.
# You should deploy OCP cluster with AWS STS configurations.
# In the script we would create the role-policy and then create the role in AWS.
# For more information see: https://docs.openshift.com/rosa/authentication/assuming-an-aws-iam-role-for-a-service-account.html

# WARNING: You cannot just run this script! you will need to replace part of the variables below

# ------------------------------------------------------------------------------------------------------------------
# Variables:
# user variables - please REPLACE these values:
ROLE_NAME="shira-28-11" # role name that you pick in your AWS account (replace shira-28-11 with your value)
NAMESPACE="test1" # namespace name where noobaa will be running (replace test1 with your value)

# noobaa variables
SERVICE_ACCOUNT_NAME_1="noobaa" # The service account name of deployment operator
SERVICE_ACCOUNT_NAME_2="noobaa-endpoint" # The service account name of deployment endpoint
SERVICE_ACCOUNT_NAME_3="noobaa-core" # The service account name of statefulset core

# AWS variables
# Please make sure these values are not empty (AWS_ACCOUNT_ID, OIDC_PROVIDER)
# AWS_ACCOUNT_ID is your AWS account number
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
# If you wish to create the role BEFORE using the cluster, please REPLACE this filed as well
# The OIDC provider is in the structure: 
# 1) <OIDC-bucket>.s3.<aws-region>.amazonaws.com. for OIDC bucket configurations are in an S3 public bucket
# 2) `<characters>.cloudfront.net` for OIDC bucket configurations in an S3 private bucket with a public CloudFront distribution URL
# 3) OIDC Endpoint URL for ROSA HCP clusters without https:// (example oidc.os1.devshift.org/<characters>)
OIDC_PROVIDER=$(oc get authentication cluster -ojson | jq -r .spec.serviceAccountIssuer | sed -e "s/^https:\/\///") 
# the permission (S3 full access)
POLICY_ARN_STRINGS="arn:aws:iam::aws:policy/AmazonS3FullAccess"

# ------------------------------------------------------------------------------------------------------------------
# Creating the role (with AWS CLI)

read -r -d '' TRUST_RELATIONSHIP <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Effect": "Allow",
     "Principal": {
       "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
     },
     "Action": "sts:AssumeRoleWithWebIdentity",
     "Condition": {
       "StringEquals": {
        "${OIDC_PROVIDER}:sub": [
          "system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT_NAME_1}",
          "system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT_NAME_2}",
          "system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT_NAME_3}"
          ]
       }
     }
   }
 ]
}
EOF


echo "${TRUST_RELATIONSHIP}" > trust.json


aws iam create-role --role-name "$ROLE_NAME" --assume-role-policy-document file://trust.json --description "role for demo"


while IFS= read -r POLICY_ARN; do
   echo -n "Attaching $POLICY_ARN ... "
   aws iam attach-role-policy \
       --role-name "$ROLE_NAME" \
       --policy-arn "${POLICY_ARN}"
   echo "ok."
done <<< "$POLICY_ARN_STRINGS"
