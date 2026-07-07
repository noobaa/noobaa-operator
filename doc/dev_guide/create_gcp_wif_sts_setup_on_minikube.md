# Create GCP WIF (STS) Setup On Minikube

Google Cloud Platform (GCP)  Workload Identity Federation (WIF) is a keyless authentication mechanism that allows workloads running outside of Google Cloud (in our case - Kubernetes) to securely access GCP resources without using long-lived, high-risk Service Account JSON keys. Instead of storing static credentials, WIF establishes a trust relationship between Google Cloud and your external Identity Provider (IdP). It uses the GCP Security Token Service (STS) to dynamically exchange external credentials (like OIDC tokens) for short-lived, automated Google Cloud access tokens.

In this guide, we would create a setup to illustrate the token exchange that is done.

## Requirements
- GCP account (with relevant permissions)
- Minikube installed locally
- *Note: Sections of these instructions require navigating the GCP Web Console.*

## Background
Traditionally, applications running outside of Google Cloud relied on long-lived, downloadable Service Account JSON keys for authentication. However, these static keys present a severe security liability if leaked, mismanaged, or left unrotated. To eliminate this operational risk, external workloads prove their identity via short-lived tokens. Once Google Cloud validates this identity, it dynamically authorizes the external workload to "impersonate" a native GCP Service Account, granting it temporary access tokens on the fly.

## Steps

### 1) Setup Minikube

#### Step 1: Setup Minikube with OIDC Flags
You must tell Minikube what its "Issuer" name is. Since it's local, we use a placeholder URL. OIDC tokens contain an iss (issuer) claim. When Minikube signs a token, it stamps https://my-minikube.local on it. When GCP receives the token, it checks if that string matches the "Issuer" you registered in the Workload Identity Pool.

```bash
minikube start \
  --extra-config=apiserver.service-account-issuer=https://my-minikube.local \
  --extra-config=apiserver.service-account-jwks-uri=https://my-minikube.local/openid/v1/jwks
```

#### Step 2: Create the service account in the cluster

```bash
kubectl create serviceaccount my-sa
```

#### Step 3: Extract the Public Keys (JWKS)
GCP needs to know how to verify the tokens Minikube signs.

```bash
kubectl get --raw /openid/v1/jwks > minikube-jwks.json
```

### 2) Setup GCP Workload Identity - can be done using the UI:
Start by opening the GCP web console, navigate to **IAM & Admin**.

#### Step 1: Create the Workload Identity Pool
1. In the web console in the left-hand sidebar click on **"Workload Identity Federation"**.
2. Click **+ Create Pool**.
3. **Name**: Enter `minikube-pool` (you can choose the name, this is an example).
4. **Description** (optional): "Pool for local Minikube lab."
5. Click **Continue**.

#### Step 2: Add the OIDC Provider
1. In the **"Add a provider to pool"** section, select `OpenID Connect (OIDC)` from the dropdown.
2. **Provider name**: Enter `minikube-provider`.
3. **Issuer (URL)**: Enter `https://my-minikube.local` (This must match your Minikube start command).
4. **JWK File (JSON)**: Click Upload and select your `minikube-jwks.json` file that you extracted from your terminal earlier.
5. **Audiences**: Select **"Default audience"**.
6. Click **Continue**.

#### Step 3: Configure Attributes
This tells GCP how to "read" the Minikube token.
1. In Attribute Mapping, in **google.subject** Enter `assertion.sub`.
2. Click Save.

#### Step 4: Create the Service Account (The "Actor")
1. In the web console in the left-hand sidebar click on **"Service Accounts"**.
2. Click **+ Create Service Account**.
3. **Name**: `minikube-tester`  (you can choose the name, this is an example).
4. **Description** (optional): "service account for WIF local Minikube lab."
5. Click **Create and Continue**.
6. **Permission** (optional): Search for
- **"Storage Object Viewer"** Grants access to view objects and their metadata, excluding ACLs. Can also list the objects in a bucket.
- **"Workload Identity User"** Impersonate service accounts from federated workloads.
- **"Service Account Token Creator"** Impersonate service accounts (create OAuth2 access tokens, sign blobs or JWTs, etc).
7. Click **Done**.

#### Step 5: Connect the Pool to the Service Account
1. Go back to the **"Workload Identity Federation"** page.
2. Click on your pool (`minikube-pool`).
3. Click **+Grant Access** at the top.
4. Choose from the radio-button **"Grant access using service account impersonation"**.
5. **Select Service Account**: Choose `minikube-tester`.
6. **Select Select principals (identities that can access the service account)**: in "subject" add `system:serviceaccount:default:my-sa`
7. Click **Save**.

### 3) Create bucket with object
1. In the GCP web console, navigate to **Google storage**.
2. In the left-hand sidebar click on **"buckets"**.
3. Click **+ Create**.
4. **Name**: `minikube-bucket` (you can choose the name, this is an example).

### 4) Terminal Save Variables
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
export MY_BUCKET="minikube-bucket"

export MY_POOL="minikube-pool"
export MY_PROVIDER="minikube-provider"
export MY_STS_AUDIENCE="//iam.googleapis.com/projects/${MY_PROJECT_NUMBER}/locations/global/workloadIdentityPools/${MY_POOL}/providers/${MY_PROVIDER}"

export MY_GSA="minikube-tester"
export MY_GSA_EMAIL="${MY_GSA}@${MY_PROJECT_ID}.iam.gserviceaccount.com"
```

### 5) Minikube Generate Service Account Token
Minikube uses its private key to digitally sign the service account token and output it to `k8s_token.txt`

```bash
kubectl create token my-sa --audience=$MY_STS_AUDIENCE > k8s_token.txt
```

This is a JWT token, if you decode it using online decoder or terminal command:

```bash
cat k8s_token.txt | \
  cut -d'.' -f2 | \
  sed 's/$/==/' | \
  base64 --decode 2>/dev/null | \
  jq
```

partial output:
```json
{
  "aud": [
    "//iam.googleapis.com/projects/<project-id>/locations/global/workloadIdentityPools/<pool-name>/providers/<provider-name>"
  ],
  "exp": 1779004239,
  "iat": 1779000639,
  "iss": "https://my-minikube.local",
  "jti": "<identifier>",
  "kubernetes.io": {
    "namespace": "default",
    "serviceaccount": {
      "name": "my-sa",
      "uid": "5cdf4bc7-816e-47b0-af8c-8b8a738cb541"
    }
  },
  "nbf": 1779000639,
  "sub": "system:serviceaccount:default:my-sa"
}
```

Optional: To enforce a tight security window, reduce the token duration (e.g., 10 minutes)

```bash
export MY_DURATION=600s
kubectl create token my-sa --audience=${MY_STS_AUDIENCE} --duration=${MY_DURATION} > k8s_token.txt
```

### 6) Token Exchange
Exchange for the Federated Token (STS)

```bash
RESPONSE=$(curl -s -X POST https://sts.googleapis.com/v1/token \
    --data-urlencode "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
    --data-urlencode "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
    --data-urlencode "scope=https://www.googleapis.com/auth/cloud-platform" \
    --data-urlencode "audience=${MY_STS_AUDIENCE}" \
    --data-urlencode "subject_token=$(cat k8s_token.txt)" \
    --data-urlencode "subject_token_type=urn:ietf:params:oauth:token-type:id_token")

FEDERATED_TOKEN=$(echo $RESPONSE | jq -r .access_token)
```

To verify, you can print the `RESPONSE` (`echo $RESPONSE`) and see the JSON:
```json
{
  "access_token": "<JWT-Token>",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3479
}
```

### 7) Swap Token

```bash
FINAL_TOKEN_JSON=$(curl -s -X POST "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/${MY_GSA_EMAIL}:generateAccessToken" \
    -H "Authorization: Bearer ${FEDERATED_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{ "scope": ["https://www.googleapis.com/auth/cloud-platform"] }')

GCP_TOKEN=$(echo $FINAL_TOKEN_JSON | jq -r .accessToken)
```

To verify, you can print the `FINAL_TOKEN_JSON` (`echo $FINAL_TOKEN_JSON`) and see the JSON:
```json
{
  "accessToken": "ya29.c.c0AY...",
  "expireTime": "2026-05-17T08:20:47Z"
}
```

Optional: Use `tokeninfo` for checking if GCP recognizes your active impersonation identity:

```bash
curl -s "https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=${GCP_TOKEN}"
```

### 8) Test the token

#### Step 1: action that the service account has permission to do
In this example - list objects in the bucket:

```bash
curl -H "Authorization: Bearer ${GCP_TOKEN}" \
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
      "id": "minikube-bucket/hello.txt/1778689193324041",
      "selfLink": "",
      "mediaLink": "",
      "name": "hello.txt",
      "bucket": "minikube-bucket",
      "generation": "",
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

#### Step 2: action that the service account does not have permission to do
In this example - try to put object `denied-test.txt` (the service account only has the role of "Storage Object Viewer").

```bash
curl -X POST \
  -H "Authorization: Bearer ${GCP_TOKEN}" \
  -H "Content-Type: text/plain" \
  -d "This write attempt should fail" \
  "https://storage.googleapis.com/upload/storage/v1/b/${MY_BUCKET}/o?uploadType=media&name=denied-test.txt"
```

partial output:
```json
{
  "error": {
    "code": 403,
    "message": "<MY_GSA> does not have storage.objects.create access to the Google Cloud Storage object. Permission 'storage.objects.create' denied on resource (or it may not exist).",
    "errors": [
      {
        "message": "<MY_GSA> does not have storage.objects.create access to the Google Cloud Storage object. Permission 'storage.objects.create' denied on resource (or it may not exist).",
        "domain": "global",
        "reason": "forbidden"
      }
    ]
  }
}
```

The resulting 403 Forbidden response confirms that identity federation completed successfully (Google clearly identifies the actor token as `minikube-tester`). However, security boundaries are enforced correctly because the account is prevented from modifying bucket storage assets.

#### Step 3: action after the token expired
Rerun an action that you have the permission to do, like list objects in the bucket and see "Invalid Credentials" error.

```bash
curl -H "Authorization: Bearer ${GCP_TOKEN}" \
"https://storage.googleapis.com/storage/v1/b/${MY_BUCKET}/o"
```

output:
```json
{
  "error": {
    "code": 401,
    "message": "Invalid Credentials",
    "errors": [
      {
        "message": "Invalid Credentials",
        "domain": "global",
        "reason": "authError",
        "locationType": "header",
        "location": "Authorization"
      }
    ]
  }
}
```
