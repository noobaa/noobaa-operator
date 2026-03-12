# Create Azure STS Setup
This document provides step-by-step instructions for setting up Azure Workload Identity Federation with NooBaa on Azure openshift and minikube clusters for local development and testing.Azure Openshift cluster set up will allow user to install NooBaa STS enabled and create default backingstore using Azure federated identity token.


### Background
What is Azure Workload Identity Federation? Azure Workload Identity Federation allows Kubernetes workloads to access Azure resources without storing credentials. allows applications running outside Azure (e.g., GitHub Actions, Kubernetes, AWS) to securely access Microsoft Entra-protected resources without managing secrets or certificates. It uses OpenID Connect(OIDC) establish trust between your Kubernetes cluster and Azure, And exchange short-lived tokens from external identity providers for Azure access tokens, eliminating risks associated with long-lived credentials. 

## Creating Azure resources

This steps are common for Openshift and Minikube.
1. Create Managed Identities: Managed Identities &rarr; Click on  `Create` &rarr; select Subscription and Resource group
 from dropdown, And also give name and Region(Isolation Scope can leave as it is) &rarr; click on `Review + create` &rarr; Click on Create.

Note: Make sure this Managed Identity has the privilege to create a Storage Account(`Storage Account Contributor`) and Container Blob(`Storage Blob Data Contributor`). You can check this in the console

Go to Managed Identity  console &rarr; Select New Managed identity &rarr; Access control &rarr; Click on `View my access`.

### Create Federated credentials under Managed Identity
2. Click on newly created Managed Identity &rarr; Settings &rarr; Federated credentials &rarr; Click on `Add Credential` &rarr; From `Add Federated Credential` console select `Kubernetes accessing Azure resources` &rarr; Add Federated Credential details &rarr;  Click on `Add`

**Where:**
- Cluster Issuer URL: {OPENID_CONTAINER_URL}
- Namespace: Namespace where NooBaa is installed
- Service Account: ServiceAccount name
    - Operator: noobaa
    - Core: noobaa-core
    - Endpoint: noobaa-endpoint
- Subject identifier: Auto populated with above details
- Name: Name of the federated credential.
- Audience: `openshift`

NooBaa uses three distinct ServiceAccounts they are `noobaa`, `noobaa-core`, and `noobaa-endpoint`, operator, core, and endpoint components respectively,

When the serviceAccountToken is mounted for each NooBaa component, the tokens will have different subjects:
- Operator: system:serviceaccount:noobaa:noobaa
- Core: system:serviceaccount:noobaa:noobaa-core
- Endpoint: system:serviceaccount:noobaa:noobaa-endpoint

Because of this, it is necessary to configure three separate Federated credentials within the managed identity, each with a different subject to accommodate these distinctions.

## Create Azure STS Setup on Openshift cluster

#### Requirements:
- Azure account with necessary permissions
- Valid Resource Group 
- Azure CLI installed and configured
- An Azure subscription with appropriate permissions
- Azure based openshift cluster(ClusterBot: `launch 4.21 azure`)
- openssl

The setup involves:
1. Create Container
2. Creating OIDC configuration files
3. Setting up Azure Managed Identities
4. Configuring Kubernetes Service Accounts
5. Establishing federated credentials
6. Verify the Federate token

## Initial Setup; Create OIDC Container
1. Create the Azure Container (for hosting the OIDC configurations). Please make it public access after creation (see notes below).

Note: Before starting with these steps, make sure the user has permission to create a storage account and a Container inside that Storage Account.

Create storage account
```bash
az storage account create --name <oidc_storage_account> --resource-group <resource_group> --location <azure_region> --sku Standard_LRS
```

Create a Container using the storage account name and key. 
```bash
az storage container create --account-name <storage_account> --name <oidc_container> --account-key $ACCOUNT_KEY
```
Where `ACCOUNT_KEY` is the account key from the previously created Storage Account.

#### Notes:
- We will refer to this container name as <oidc_container> in the next instructions.
- This Container should be publicly accessible\
In Azure Container Storage browser Console:\
&nbsp;&nbsp; (1) Click on the `Change access level`.\
&nbsp;&nbsp;
(2) Select `Blob (anonymous read access for blobs only)` and save.

2. Save the storage account name and container name as a variable (replace `<oidc_container>` with the name that you chose for the container):

```bash
OPENID_STORAGE_ACCOUNT='<oidc_storage_account>'
OPENID_CONTAINER_NAME='<oidc_container>'
```

### General Setup:
3. Create a directory for files to save the files in one-place, for example: `~/Documents/azure_sts` and make sure that it is empty.

```bash
mkdir -p ~/Documents/azure_sts
```

```bash
cd ~/Documents/azure_sts
```

```bash
ls -al
```

4. Save this variable (on every terminal that we use):

```bash
OPENID_CONTAINER_URL='https://<oidc_storage_account>.blob.core.windows.net/<oidc_container>'
```

`<oidc_container>` is the container you created in the initial step and the `<oidc_storage_account>` is the storage account for that container (replace them with your details).

### Openshift Setup:

After the openshift cluster up and running, update the system `kubeconfig` with cluster's kube config.

6. Patch the cluster authentication config with openid container URl, setting `spec.serviceAccountIssuer`.

```bash
oc patch authentication cluster --type=merge -p "{\"spec\":{\"serviceAccountIssuer\":\"${OPENID_CONTAINER_URL}\"}}"
```

Wait for the kube-apiserver pods to be updated with the new config. This process can take several minutes.
```bash
oc adm wait-for-stable-cluster
```


Note: Every time you create a new OpenShift cluster, you will need to patch the `serviceAccountIssuer` (and the 2 files that are described in the next steps).

### Build an OIDC configuration

Note: those steps were taken from [OCP 4.7 doc](https://docs.openshift.com/container-platform/4.7/authentication/managing_cloud_provider_credentials/cco-mode-sts.html#sts-mode-installing-manual-config_cco-mode-sts), and [CCO Azure AD Workload Identity](https://github.com/openshift/cloud-credential-operator/blob/master/docs/azure_workload_identity.md#steps-to-in-place-migrate-an-openshift-cluster-to-azure-ad-workload-identity). There you can read the full explanations for each command.

7. fetch the Service Account signing public key for Openshift cluster:

```bash
cd ~/Documents/azure_sts 
```

Extract the cluster's ServiceAccount public signing key.
```bash
oc get secret/next-bound-service-account-signing-key \
--namespace openshift-kube-apiserver-operator \
-o jsonpath='{ .data.service-account\.pub }' \
| base64 -d \
> sa-signer.pub
```

You'll be able to see the a new created file  `sa-signer.pub`.

```bash
ls -al
```

8. Create a file named `keys.json` that contains the following information:

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

If your OpenSSL library is `LibreSSL` Then you will get an error(`Invalid cipher '-outform' error`) for the above command. In that case, please try with the below command
```bash
openssl pkey -pubin -in sa-signer1.pub -outform DER | openssl dgst -binary -sha256 | openssl base64 | tr '/+' '_-' | tr -d '='
```

- `<public_signing_key_modulus>` is generated from the public key with:

```bash
openssl rsa -pubin -in sa-signer.pub -modulus -noout | sed  -e 's/Modulus=//' | xxd -r -p | base64 | tr '/+' '_-' | tr -d '='
```

- `<public_signing_key_exponent>` is generated from the public key with:

```bash
printf "%016x" $(openssl rsa -pubin -in sa-signer.pub -noout -text | grep Exponent | awk '{ print $2 }') |  awk '{ sub(/(00)+/, "", $1); print $1 }' | xxd -r -p | base64  | tr '/+' '_-' | tr -d '='
```

Note: In the commands above was piping to `base64 -w0``, since I don't have this flag on MAC I removed it.

9. Create a file named `openid-configuration` that contains the following information:

```json
{
    "issuer": "${OPENID_CONTAINER_URL}",
    "jwks_uri": "${OPENID_CONTAINER_URL}/keys.json",
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
- You need the paste the value of `${OPENID_CONTAINER_URL}`.

10. Upload the OIDC configuration:

```bash
 az storage blob upload --account-name <account-name> --account-key <account-key> --container-name ${OPENID_CONTAINER_NAME}  --file ./keys.json --name keys.json
```

```bash
 az storage blob upload --account-name <account-name> --account-key <account-key> --container-name ${OPENID_CONTAINER_NAME}  --file ./openid-configuration --name '.well-known/openid-configuration'
```


11. You can verify that the configuration are public available by using your browser (Chrome, Firefox, etc.) and enter the URL of: 
```
{OPENID_CONTAINER_URL}/keys.json
{OPENID_CONTAINER_URL}/.well-known/openid-configuration
```

### Federated Token Test:
We would create an nginx pod and fetch the ServiceAccount token from it and then validate and see that we can get the credentials.

12. Create a nginx pod

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
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
             audience: openshift
EOF
```

Note: For testing this you need to create Federated credentials with Namespace `default` and Service Account `default`. You can follow the same steps mentioned here [Create Federated credentials under Managed Identity](#create-federated-credentials-under-managed-identity).

13. Fetch the Projected service account token and save it in `WEB_FEDERATED_TOKEN`.

```bash
WEB_FEDERATED_TOKEN=$(kubectl exec nginx -- cat /var/run/secrets/tokens/oidc-token)
```

14. Fetch Access Token using Federated Identity Token(JWT Token) and use that access token to access the resource (In our case, it's the Storage Account and Container storage).

Get Access Token using Federated Identity

```bash
curl -L -X POST 'https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token' \ 
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "client_id=<client_id>" \ 
     -d "scope=https://storage.azure.com/.default" \
     -d "grant_type=client_credentials" \
     -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
     -d "client_assertion=$WEB_FEDERATED_TOKEN"
```
Extract the access token from the previous response
```
ACCESS_TOKEN="<access_token_from_previous_step>"
```

**Where:**
- tenant_id: Tenant ID from Azure Entra ID
- client_id: Managed Identity's Client ID

```bash
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "x-ms-version: 2023-11-03" \
    "https://<storage-account>.blob.core.windows.net/<container-name>/<key>"
```

Make Sure you are able to access the container blob key.

15. Test Azure STS with Noobaa Installation

    Install NooBaa in Azure cluster
```bash
noobaa install  --azure-tenant-id <azure-tenant-id> --azure-client-id <azure-client-id> --azure-subscription-id <azure-subscription-id> --azure-resourcegroup <azure-resourcegroup>
```

**Where:**
- azure-tenant-id: Tenant ID from Azure Entra ID
- azure-client-id: Managed Identity's Client ID
- azure-subscription-id: Resource Manager Subscriptions ID
- azure-resourcegroup: Managed Identity created resource group name

Verify default backingstore created with Azure STS credentials is in ready state.

```bash
noobaa-operator % nb backingstore status noobaa-default-backing-store
```

16. Create backingstore using Azure STS credentials

- Create secret with STS credentials

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: <backingstore-secret>
  namespace: <namespace>a
  labels:
    app: noobaa
data:
  azure_tenant_id: <tenant_id>
  AccountName: <account-name>
  azure_client_id: <client-id>
type: Opaque
```

- Create backing store
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  name: <name>
  namespace: <namespace>
  labels:
    app: noobaa
spec:
  azureBlob:
    clientId: <client-id>
    secret:
      name: <backingstore-secret>
      namespace: <namespace>
    targetBlobContainer: <target-container>
  type: azure-blob
```

Verify backingstore created with Azure STS credentials is in ready state.

```bash
noobaa-operator % nb backingstore status <name>
```


## Create Azure STS Setup on Minikube cluster

#### Requirements:
- AWS account (with Admin permission)
- AWS CLI
- minikube
- openssl

1. Create OIDC container as mentioned steps 1 to 4 in  [Initial Setup - Create OIDC Container](#initial-setup-create-oidc-container)

### Minikube Setup:
2. Start Minikube with the following flags:

```bash
minikube start --extra-config=apiserver.service-account-issuer=${OPENID_CONTAINER_URL}
```

Note: every time you do `minikube delete` and `minikube start` you will need to create this new file (and the 2 files that are described in next steps).


### Build an OIDC configuration for Minikube

Note: those steps were taken from [OCP 4.7 doc](https://docs.openshift.com/container-platform/4.7/authentication/managing_cloud_provider_credentials/cco-mode-sts.html#sts-mode-installing-manual-config_cco-mode-sts), and [CCO Azure AD Workload Identity](https://github.com/openshift/cloud-credential-operator/blob/master/docs/azure_workload_identity.md#steps-to-in-place-migrate-an-openshift-cluster-to-azure-ad-workload-identity). There you can read the full explanations for each command.

4. After the minikube node has started, fetch the Service Account signing public key:

```bash
cd ~/Documents/azure_sts 
```

```bash
minikube ssh sudo cat /var/lib/minikube/certs/sa.pub > sa-signer.pub
```

You'll be able to see the new created file  `sa-signer.pub`.

```bash
ls -al
```

5. Create OIDC configuration and push those files to OIDC container, As mentioned here [Build an OIDC configuration](#build-an-oidc-configuration) step 8 to 12

### Federated Token Test:
We would create an nginx pod and fetch the ServiceAccount token from it and then validate and see that we can get the credentials.

6. Create a nginx pod

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
             audience: openshift
EOF
```

Note: For testing this you need to create Federated credentials with Namespace `default` and Service Account `default`. You can follow the same steps mentioned in step 14.

17. Fetch the Projected service account token and save it in `WEB_FEDERATED_TOKEN`.

```bash
WEB_FEDERATED_TOKEN=$(kubectl exec nginx -- cat /var/run/secrets/tokens/oidc-token)
```

18. Fetch Access Token using Federated Identity Token(JWT Token) and use that access token to access the resource (In our case, it's the Storage Account and Container storage).

Get Access Token using Federated Identity

```bash
curl -L -X POST 'https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token' \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "client_id=<client_id>" \
     -d "scope=https://storage.azure.com/.default" \
     -d "grant_type=client_credentials" \
     -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
     -d "client_assertion=$WEB_FEDERATED_TOKEN"
```
Note: You might need to remove spaces at the end of each line.

Extract the access token from the previous response
```
ACCESS_TOKEN="<access_token_from_previous_step>"
```

**Where:**
- tenant_id: Tenant ID from Azure Entra ID
- client_id: Managed Identity's Client ID

```bash
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "x-ms-version: 2023-11-03" \
    "https://<storage-account>.blob.core.windows.net/<container-name>/key"
```

Make Sure you are able to access the container blob key.
