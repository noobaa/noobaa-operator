# Create GCP WIF (STS) Setup On Minikube With NooBaa

If you want to understand the GCP WIF and Minikube see: [Create GCP WIF (STS) Setup On Minikube](create_gcp_wif_sts_setup_on_minikube.md).

In this guide, we would create a setup to illustrate the token exchange that is done using installed NooBaa system and getting the token that were projected in the pods.

## Requirements
- GCP account (with relevant permissions)
- Minikube installed locally
- *Note: Sections of these instructions require navigating the GCP Web Console.*

## Steps
Start by opening the GCP web console.

### 1) Create the Public OIDC Storage Bucket on GCP

1. In the GCP web console, navigate to **Google storage**.
2. In the left-hand sidebar click on **"Buckets"**.
3. Click **+ Create**.
4. **Name**: `minikube-oidc-issuer-<your-name>` (you can choose the name, this is an example) and pick a region near you.
5. Under **Choose how to control access to objects**, uncheck **"Enforce public access prevention on this bucket"**.  
Keep Uniform or Fine-grained access, but ensure public access is allowed.
6. Click **Create**.
6. Once created, click on the **Permissions** tab.
7. Click **Grant Access**:
- *New principals*: `allUsers`
- *Role*: `Storage Object Viewer`
8. Click **Save** and confirm **Allow Public Access**.

### 2) Setup Minikube With OIDC Flags

1. Export your bucket URL to environment variables:

```bash
export BUCKET_NAME="minikube-oidc-issuer-<your-name>"
export OIDC_ISSUER_URL="https://storage.googleapis.com/${BUCKET_NAME}"
```

Note: `<your-name>` is a placeholder 

2. Start Minikube with API server arguments configuring the Token Request projection and OIDC features:

```bash
minikube start \
  --extra-config=apiserver.service-account-issuer=${OIDC_ISSUER_URL} \
  --extra-config=apiserver.service-account-key-file=/var/lib/minikube/certs/sa.pub \
  --extra-config=apiserver.service-account-signing-key-file=/var/lib/minikube/certs/sa.key \
  --extra-config=apiserver.api-audiences=openshift
```

### 3) Install NooBaa
Install noobaa (for more information see:  [Deploy NooBaa on Minikube / Rancher Desktop](deply_noobaa_on_minikube_or_rancher_desktop.md)).


### 4) Setup GCP Workload Identity (Web Console UI):
Start by opening the GCP web console, navigate to **IAM & Admin**.  
Now, map Kubernetes Service Accounts to a GCP Service Account using a Workload Identity Pool.

#### Step 1: Create the Workload Identity Pool
1. In the web console in the left-hand sidebar click on **"Workload Identity Federation"**.
2. Click **+ Create Pool**.
3. **Name**: Enter `minikube-oidc-pool` (you can choose the name, this is an example).
4. **Description** (optional): "Pool for local Minikube lab."
5. Click **Continue**.

#### Step 2: Add the OIDC Provider
1. In the **"Add a provider to pool"** section, select `OpenID Connect (OIDC)` from the dropdown.
2. **Provider name**: Enter `minikube-oidc-provider`.
3. **Issuer (URL)**: Enter `https://storage.googleapis.com/minikube-oidc-issuer-<your-name>`   
(This must match your Minikube start command).
4. **Audiences**: Select **"Allowed audiences"** click **Audience**, and type exactly: `openshift`.
5. Click **Continue**.

#### Step 3: Configure Attributes
6. Under **Configure provider attributes**, set up your claims mapping:

| Google Attribute          | CEL Expression (Value)                               |
| ------------------------- | ---------------------------------------------------- |
| google.subject            | assertion.sub                                        |
| attribute.namespace       | assertion['kubernetes.io']['namespace']              |
| attribute.service_account | assertion['kubernetes.io']['serviceaccount']['name'] |

7. Click **Save**

#### Step 4: Create the Service Account (The "Actor")
1. In the web console in the left-hand sidebar click on **"Service Accounts"**.
2. Click **+ Create Service Account**.
3. **Name**: `noobaa-wif-sa`  (you can choose the name, this is an example).
4. **Description** (optional): "service account for WIF local Minikube lab."
5. Click **Create and Continue**.
6. **Permission** (optional): Search for
- **"Storage Admin"** Grants full control of buckets and objects.
- **"Browser"** Access to browse GCP resources.
7. Click **Done**.
8. Once created, click on the Service Account name, go to the **Principals with access** tab, and click **Grant Access**.
9. In **New principals**, construct the principal identifier for your local minikube service account - this is the structure:

```bash
principal://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/<POOL_NAME>/subject/system:serviceaccount:<NOOBAA_NAMESPACE>:<NOOBAA_SERVICE_ACCOUNT_NAME>
```

Note: We used placeholders:
- <NOOBAA_NAMESPACE> is where NooBaa is deployed.
- <NOOBAA_SERVICE_ACCOUNT_NAME> is a placeholder that we will have for noobaa (noobaa operator), noobaa-core and noobaa-endpoint.
- <POOL_NAME> is the pool name (in this example `minikube-oidc-pool`)

10. Assign the role: **Workload Identity User**
11. Click **Save**.

#### Step 5: Extract and Upload OIDC Discovery Files
GCP needs to fetch the cluster’s public signing keys from your bucket. We will extract them from the running Minikube api-server and upload them to your GCP bucket.
1. Extract Discovery Documents: Create a local directory structure matching the OIDC specification

```bash
mkdir -p .well-known
```

2. Fetch openid configuration

```bash
kubectl get --raw /.well-known/openid-configuration > .well-known/openid-configuration
```

3. Fetch keys

```bash
kubectl get --raw /openid/v1/jwks > keys.json
```

4. Edit the `openid-configuration` file so that `jwks_uri` would be `"jwks_uri":"https://storage.googleapis.com/minikube-oidc-issuer-<your-name>/keys.json"` (replace `-<your-name>`).
5. Go to your bucket in the GCP Console and create a folder named `.well-known`.
6. Upload the files to the bucket (replace `-<your-name>`):

```bash
# Upload openid-configuration while explicitly forcing the JSON content-type and breaking edge cache policies
gcloud storage cp .well-known/openid-configuration gs://minikube-oidc-issuer-<your-name>/.well-known/openid-configuration \
  --content-type="application/json" \
  --cache-control="no-store, no-cache, must-revalidate, max-age=0"

# Upload keys.json with the proper content-type
gcloud storage cp keys.json gs://minikube-oidc-issuer-<your-name>/keys.json \
  --content-type="application/json" \
  --cache-control="no-store, no-cache, must-revalidate, max-age=0"
```

Troubleshooting: Google caches OIDC discovery endpoints. If you update your configuration (`.well-known/openid-configuration`) or keys file (`keys.json`) and still see "Error connecting to the given credential's issuer.", force a cache bust check in your terminal by adding a dummy query parameter:

```bash
curl -i "https://storage.googleapis.com/minikube-oidc-issuer-<your-name>/.well-known/openid-configuration?nocache=$(date +%s)"
```

Verify that the returned "jwks_uri" explicitly points to your public `keys.json` file on GCP before retrying the token exchange loop.

### 5) Create bucket with object
1. In the GCP web console, navigate to **Google storage**.
2. In the left-hand sidebar click on **"Buckets"**.
3. Click **+ Create**.
4. **Name**: `my-test-bucket` (you can choose the name, this is an example).

### 6) Terminal Save Variables
Save the details of the created resources in variable names:
Notes:
- The `MY_PROJECT_ID` and `MY_PROJECT_NUMBER` are variables that you need to add the information, in the below lines it is a placeholder.
- The project number you would find in the GCP dashboard.

```bash
# --- SET YOUR PROJECT DETAILS HERE ---
export MY_PROJECT_ID="<your-project-id>"
export MY_PROJECT_NUMBER="<your-project-number>"
#------------------------------------- 
# These derived variables will handle the rest of the commands
export MY_POOL="minikube-oidc-pool"
export MY_PROVIDER="minikube-oidc-provider"
export MY_STS_AUDIENCE="//iam.googleapis.com/projects/${MY_PROJECT_NUMBER}/locations/global/workloadIdentityPools/${MY_POOL}/providers/${MY_PROVIDER}"

export MY_GSA="noobaa-wif-sa"
export MY_GSA_EMAIL="${MY_GSA}@${MY_PROJECT_ID}.iam.gserviceaccount.com"
```

### 7) Validate Projected Tokens From Noobaa Pods
1. Take the projected token:

```bash
MY_TOKEN_OPERATOR=$(kubectl exec $(kubectl get pods -n <NOOBAA_NAMESPACE> | grep operator | awk '{ print $1}') -c noobaa-operator -n <NOOBAA_NAMESPACE> -- cat /var/run/secrets/openshift/serviceaccount/token)

MY_TOKEN_CORE=$(kubectl exec $(kubectl get pods -n <NOOBAA_NAMESPACE> | grep core | awk '{ print $1}') -c core -n <NOOBAA_NAMESPACE> -- cat /var/run/secrets/openshift/serviceaccount/token)

MY_TOKEN_ENDPOINT=$(kubectl exec $(kubectl get pods -n <NOOBAA_NAMESPACE> | grep endpoint | awk '{ print $1}') -c endpoint -n <NOOBAA_NAMESPACE> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

To verify you can print the token (`echo $MY_TOKEN_OPERATOR`, `echo $MY_TOKEN_CORE`, `echo $MY_TOKEN_ENDPOINT` each one).

2. In this example we will continue with noobaa-operator token, but it should work with noobaa-core and noobaa-endpoint as well.

```bash
export SUBJECT_TOKEN="${MY_TOKEN_OPERATOR}"
echo "$SUBJECT_TOKEN" | cut -d'.' -f2 | base64 --decode
```

### 8) Token Exchange
Exchange for the Federated Token (STS)

```bash
STS_RESPONSE=$(curl -s -X POST https://sts.googleapis.com/v1/token \
  --data-urlencode "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  --data-urlencode "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  --data-urlencode "scope=https://www.googleapis.com/auth/cloud-platform" \
  --data-urlencode "audience=${MY_STS_AUDIENCE}" \
  --data-urlencode "subject_token=${SUBJECT_TOKEN}" \
  --data-urlencode "subject_token_type=urn:ietf:params:oauth:token-type:jwt")

FEDERATED_TOKEN=$(echo $STS_RESPONSE | tr -d '\n' | grep -o '"access_token":[^,]*' | awk -F'"' '{print $4}')

```

To verify, you can print the `STS_RESPONSE` (`echo $STS_RESPONSE`) and see the JSON:
```json
{
  "access_token": "<JWT-Token>",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3479
}
```

### 9) Swap Token

```bash
IMPERSONATED_RESPONSE=$(curl -s -X POST "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/${MY_GSA_EMAIL}:generateAccessToken" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${FEDERATED_TOKEN}" \
  -d '{
    "scope": ["https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/cloud-platform"]
  }')

GCS_ACCESS_TOKEN=$(echo $IMPERSONATED_RESPONSE | tr -d '\n' | grep -o '"accessToken":[^,]*' | awk -F'"' '{print $4}')
```

To verify, you can print the `IMPERSONATED_RESPONSE` (`echo $IMPERSONATED_RESPONSE`) and see the JSON:
```json
{
  "accessToken": "ya29.c.c0AY...",
  "expireTime": "2026-05-17T08:20:47Z"
}
```

### 10) Test the token
In this example - list objects in the bucket:

```bash
export MY_BUCKET="my-test-bucket"

curl -i -H "Authorization: Bearer ${GCS_ACCESS_TOKEN}" \
"https://storage.googleapis.com/storage/v1/b/${MY_BUCKET}/o"
```

You would see the output of the list objects inside the bucket.
partial output:
```json
{
  "kind": "storage#objects",
  "items": [
    {
      "kind": "storage#object",
      "id": "my-test-bucket/hello.txt/1778689193324041",
      "selfLink": "",
      "mediaLink": "",
      "name": "hello.txt",
      "bucket": "my-test-bucket",
      "generation": "1778689193324041",
      "metageneration": "1",
      "contentType": "text/plain",
      "storageClass": "STANDARD",
      "size": "6",
      "md5Hash": "",
      "crc32c": "",
      "etag": "",
      "timeCreated": "2026-05-13T16:19:53.385Z",
      "updated": "2026-05-13T16:19:53.385Z",
      "timeStorageClassUpdated": "2026-05-13T16:19:53.385Z",
      "timeFinalized": "2026-05-13T16:19:53.385Z"
    }
  ]
}
```
