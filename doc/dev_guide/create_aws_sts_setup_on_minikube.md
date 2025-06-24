# Create AWS STS Setup On Minikube
AWS Security Token Service (STS) is a web service that enables users to request temporary, limited-privilege credentials. In this guide, we would create a setup that eventually we will able to get those credentials (see the Test part). We would use OpenID Connection (OIDC) which is an identity layer that allows the verification of the identity.

#### Requirements:
- AWS account (with Admin permission)
- AWS CLI
- minikube
- openssl

### Background
AWS STS allows applications to provide a Json Web Token (JWT) that can be used to assume an IAM role. The JWT includes an Amazon Resource Name (ARN) for `sts:AssumeRoleWithWebIdentity` IAM action to allow temporarily granted permission for the `ServiceAccount` to do this. The JWT contains the signing keys for the `ProjectedServiceAccountToken` that can be validated by AWS IAM.

### Initial Setup - Create S3 Bucket (the OIDC Bucket)
1. Create the AWS S3 bucket (for hosting the OIDC configurations). Please make it public access and ACL enabled after creation (see notes below).

```bash
aws s3api create-bucket --bucket <bucket_name> --region <aws_region> --create-bucket-configuration LocationConstraint=<aws_region>
```

#### Notes:
- If the value of <aws_region> is us-east-1, do not specify the LocationConstraint parameter.
- We will refer to this bucket name as <oidc_bucket_name> in the next instructions.
- In AWS Console:\
&nbsp;&nbsp; (1) under "Block Public Access settings for this bucket" remove all the check marks;\
&nbsp;&nbsp; (2) mark "ACLs enabled".
- Can be installed using AWS Console UI: S3 &rarr; Buckets &rarr; Create Bucket &rarr; Fill the bucket name + region &rarr;  ACLs enabled &rarr; Remove the V from Block all public access &rarr; Use the rest of the defaults.

2. Save the bucket name as a variable (replace `<oidc_bucket_name>` with the name that you chose for the bucket):

```bash
OPENID_BUCKET_NAME='<oidc_bucket_name>'
```

### General Setup:
3. Create a directory for files to save the files in one-place, for example: `~/Documents/my_sts` and make sure that it is empty.

```bash
mkdir -p ~/Documents/my_sts
```

```bash
cd ~/Documents/my_sts
```

```bash
ls -al
```

4. Save this variable (on every terminal that we use):

```bash
OPENID_BUCKET_URL='https://<oidc_bucket_name>.s3.<aws_region>.amazonaws.com'
```

`<oidc_bucket_name>` is the bucket you created in the initial step and the `<aws_region>` is the region of this bucket (replace them with your details).

### Minikube Setup:
5. Start Minikube with the following flags:

```bash
minikube start --extra-config=apiserver.service-account-issuer=${OPENID_BUCKET_URL} --kubernetes-version=v1.23.12
```

I added the flag of `--kubernetes-version=v1.23.12` because iI had errors of timeout without it.

6. After the minikube node has started, fetch the Service Account signing public key:

```bash
cd ~/Documents/my_sts 
```

```bash
minikube ssh sudo cat /var/lib/minikube/certs/sa.pub > sa-signer.pub
```

You'll be able to see the new created file  `sa-signer.pub`.

```bash
ls -al
```

Note: every time you do `minikube delete` and `minikube start` you will need to create this new file (and the 2 files that are described in next steps).

### Build an OIDC configuration
Note: those steps were taken from [OCP 4.7 doc](https://docs.openshift.com/container-platform/4.7/authentication/managing_cloud_provider_credentials/cco-mode-sts.html#sts-mode-installing-manual-config_cco-mode-sts), there you can read the full explanations for each command).

7. Create a file named `keys.json` that contains the following information:

```json
{
    "keys": [
        {
            "use": "sig",
            "kty": "RSA",
            "kid": "<public_signing_key_id>",
            "alg": "RS256",
            "n": "<public_signing_key_modulus>",
            "e": "<public_signing_key_exponent>"
        }
    ]
}
```

**Where:**
- `<public_signing_key_id>` is generated from the public key with:

```bash
openssl rsa -in sa-signer.pub -pubin --outform DER | openssl dgst -binary -sha256 | openssl base64 | tr '/+' '_-' | tr -d '='
```

- `<public_signing_key_modulus>` is generated from the public key with:

```bash
openssl rsa -pubin -in sa-signer.pub -modulus -noout | sed  -e 's/Modulus=//' | xxd -r -p | base64 | tr '/+' '_-' | tr -d '='
```

- `<public_signing_key_exponent>` is generated from the public key with:

```bash
printf "%016x" $(openssl rsa -pubin -in sa-signer.pub -noout -text | grep Exponent | awk '{ print $2 }') |  awk '{ sub(/(00)+/, "", $1); print $1 }' | xxd -r -p | base64  | tr '/+' '_-' | tr -d '='
```

Note: in the commands above were piping to `base64 -w0``, since I don't have this flag on MAC I removed it.

8. Create a file named `openid-configuration` that contains the following information:

```json
{
	"issuer": "${OPENID_BUCKET_URL}",
	"jwks_uri": "${OPENID_BUCKET_URL}/keys.json",
    "response_types_supported": [
        "id_token"
    ],
    "subject_types_supported": [
        "public"
    ],
    "id_token_signing_alg_values_supported": [
        "RS256"
    ],
    "claims_supported": [
        "aud",
        "exp",
        "sub",
        "iat",
        "iss",
        "sub"
    ]
}
```

**Where:**
- You need the paste the value of `${OPENID_BUCKET_URL}`.

9. Upload the OIDC configuration:

```bash
aws s3api put-object --bucket ${OPENID_BUCKET_NAME} --key keys.json --body ./keys.json
```

```bash
aws s3api put-object --bucket ${OPENID_BUCKET_NAME} --key '.well-known/openid-configuration' --body ./openid-configuration
```

10. Allow the AWS IAM OIDC identity provider to read these files:

```bash
aws s3api put-object-acl --bucket ${OPENID_BUCKET_NAME} --key keys.json --acl public-read
```

```bash
aws s3api put-object-acl --bucket ${OPENID_BUCKET_NAME} --key '.well-known/openid-configuration' --acl public-read
```

11. You can verify that the configuration are public available by using your browser (Chrome, Firefox, etc.) and enter the URL of: 
https://<oidc-bucket-name>.s3.<region>.amazonaws.com/keys.json
https://<oidc-bucket-name>.s3.<region>.amazonaws.com/.well-known/openid-configuration

### Creating AWS resources manually - Using the Amazon web console:
12. Create s3 bucket (you already did it in the step Initial Setup - Create S3 Bucket).

13. Create Identity Provider: IAM &rarr; Identity providers &rarr; Add provider &rarr; Provider type: choose OpenID Connect &rarr; Provider URL: paste the value of OPENID_BUCKET_URL &rarr; click on `Get thumbprint`` &rarr; Audience: api (type api in the field) &rarr; Click on add provider.

14. Create role: IAM &rarr; Roles &rarr; Create Role &rarr; Trusted entity type: Web Identity &rarr; Identity Provider should be the name of the provider that we added (with structure: https://<oidc_bucket_name>.s3.<aws_region>.amazonaws.com ) &rarr; Add the permission: `AmazonS3FullAccess`.

When you finish, check in the Trusted entities that you see:

```json
    "Principal": {
        "Federated": "arn:aws:iam::<account-id>:oidc-provider/<oidc-bucket-name>.s3.<region>.amazonaws.com"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
        "StringEquals": {
            "<oidc-bucket-name>.s3.<region>.amazonaws.com:aud": "api"
        }
```

15. In later steps you will need to provide the ARN of the role (you can easily copy it from AWS console, it looks like `arn:aws:iam::<id-account>:role/<role-name>` you can create a variable in the terminal:

```bash
OIDC_ROLE_ARN='<paste here you role ARN>'
```

### Test:
We would create an nginx pod and fetch the Service Account token from it and then run `assume-role-with-web-identity` and see that we can get the credentials.

16. Create a nginx pod

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - image: nginx:alpine
      name: oidc
      volumeMounts:
        - mountPath: /var/run/secrets/tokens
          name: oidc-token
  volumes:
    - name: oidc-token
      projected:
        sources:
          - serviceAccountToken:
             path: oidc-token
             expirationSeconds: 7200
             audience: api
EOF
```

17. Fetch the Projected service account token and save it in `WEB_IDENTITY_TOKEN`.

```bash
WEB_IDENTITY_TOKEN=$(kubectl exec nginx -- cat /var/run/secrets/tokens/oidc-token)
```

18. Use assume-role-with-web-identity

```bash
aws sts assume-role-with-web-identity --role-arn ${OIDC_ROLE_ARN} --role-session-name "test" --web-identity-token ${WEB_IDENTITY_TOKEN}
```

You should see in the output the credentials (which includes the `AccessKeyId`, `SecretAccessKey`, and `SessionToken`), output example:

```json
{
    "Credentials": {
        "AccessKeyId": "EXAMPLE",
        "SecretAccessKey": "EXAMPLE",
        "SessionToken": "EXAMPLE",
        "Expiration": "2023-11-08T10:03:57+00:00"
    },
    "SubjectFromWebIdentityToken": "system:serviceaccount:default:default",
    "AssumedRoleUser": {
        "AssumedRoleId": "EXAMPLE:<session-name>",
        "Arn": "arn:aws:sts::<account-id>:assumed-role/<role-name>/<session-name>"
    },
    "Provider": "arn:aws:iam::<account-id>:oidc-provider/<oidc_bucket_name>.s3.<aws_region>.amazonaws.com",
    "Audience": "api"
}
```

#### Create backing store using STS

Noobaa WEB_IDENTITY_TOKEN Audience value is `openshift` because of that You shouls update the Audience in Identity Provider to opeshift.

```bash
TARGET_BUCKET='<target bucket name>'
```

```
nb backingstore create aws-sts-s3 {backing-store} --target-bucket ${TARGET_BUCKET} --aws-sts-arn ${OIDC_ROLE_ARN}
``` 