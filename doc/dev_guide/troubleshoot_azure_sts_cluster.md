# TroubleShoot Azure STS Cluster
Here are common errors encountered during Azure STS installations and how to investigate/fix them. These issues typically occur when trying to install NooBaa and the system is stuck in phase 'Configuring' while NooBaa attempts to create the default backingstore that matches the Azure STS platform or when backingstore and namespacestore stuck in creating state. 


### Basic Sanity Check 

Before installing NooBaa with Azure STS or creating a BackingStore/NamespaceStore that utilizes Azure STS credentials, you must verify that the client ID, tenant ID, subscription ID, and resource group have access to the necessary storage resources. This access must be granted in addition to the ServiceAccount token. The following steps will guide you through this verification process.


1. Fetch the Projected service account token and save it in `WEB_FEDERATED_TOKEN`.

```bash
WEB_FEDERATED_TOKEN=$(kubectl exec noobaa-core-0 -- cat /var/run/secrets/openshift/serviceaccount/token)
```

2. Fetch Access Token using Federated Identity Token(JWT Token) and use that access token to access the resource (In our case, it's the Storage Account and Container storage).

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

Make Sure you are able to access the non-public container blob data.
If this test succeeds, the Azure Workload Identity Federation is configured correctly, and the issue may be elsewhere in the NooBaa configuration.

### Main issues:

#### 1) Wrong Managed Identity or Missing Federated Credentials:

```
time="2024-03-19T10:15:30Z" level=warning msg="⏳ Temporary Error: could not authenticate with Azure using Workload Identity Federation with client ID <client-id> and web identity token file /var/run/secrets/tokens/azure-identity-token, AADSTS70021: No matching federated identity record found for presented assertion subject 'system:serviceaccount:noobaa:noobaa-core'. Check your federated identity credential Subject, Audience and Issuer against the presented assertion.
```

##### Solution:
Verify that the Managed Identity has the correct Federated Credentials configured in the Azure portal.

You can check the federated credentials in Azure Console:
- Go to Managed Identities → Select your Managed Identity → Settings → Federated credentials

Make sure you have three separate federated credentials configured for:
1. Service Account: `noobaa` (operator)
2. Service Account: `noobaa-core` (core)
3. Service Account: `noobaa-endpoint` (endpoint)

Each federated credential should have:
- **Cluster Issuer URL**: Your OIDC container URL (e.g., `https://<storage-account>.blob.core.windows.net/<container>`)
- **Namespace**: The namespace where NooBaa is installed (e.g., `openshift-storage` or `noobaa`)
- **Service Account**: The respective service account name
- **Audience**: `openshift`

You also need to verify that the token is projected - for example in the operator pod:
```bash
MY_TOKEN_OPERATOR=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep operator | awk '{ print $1}') -c noobaa-operator -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

```bash
echo ${MY_TOKEN_OPERATOR}
```

To verify the token content, decode the JWT token:
```bash
echo ${MY_TOKEN_OPERATOR} | cut -d '.' -f 2 | base64 -d | jq .
```

You should see the service account name in the token payload under `kubernetes.io.serviceaccount.name`.

Note: In core you might see the error starts with `got error on listContainers with params`

#### 2) Missing Service Account Name in Federated Credentials

If the federated credentials are not configured for all three service accounts (noobaa, noobaa-core, noobaa-endpoint), you'll encounter authentication errors when those specific pods try to access Azure resources.

##### Solution:
Verify that every pod that needs to access Azure resources has its token properly configured. Test each service account:

For the operator pod:
```bash
MY_TOKEN_OPERATOR=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep operator | awk '{ print $1}') -c noobaa-operator -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

For the core pod:
```bash
MY_TOKEN_CORE=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep core | awk '{ print $1}') -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

For the endpoint pod:
```bash
MY_TOKEN_ENDPOINT=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep endpoint | awk '{ print $1}') -n <your-namespace> -- /var/run/secrets/openshift/serviceaccount/token)
```

Decode each token to verify the service account name:
```bash
echo ${MY_TOKEN_CORE} | cut -d '.' -f 2 | base64 -d | jq .
```

The token should show:
```json
{
  "kubernetes.io": {
    "namespace": "openshift-storage",
    "serviceaccount": {
      "name": "noobaa-core"
    }
  },
  "sub": "system:serviceaccount:openshift-storage:noobaa-core"
}
```

Ensure all three federated credentials are configured in the Managed Identity with the correct subject identifiers:
- `system:serviceaccount:<namespace>:noobaa`
- `system:serviceaccount:<namespace>:noobaa-core`
- `system:serviceaccount:<namespace>:noobaa-endpoint`

#### 3) Cluster OIDC Configuration Issues

```
time="2024-03-19T10:20:45Z" level=warning msg="⏳ Temporary Error: could not authenticate with Azure using Workload Identity Federation, AADSTS50027: JWT token is invalid or malformed" sys=noobaa/noobaa
```

##### Solution:
Verify that the cluster's OIDC configuration is properly set up.

For OpenShift clusters, check the service account issuer:
```bash
oc get authentication cluster -o jsonpath --template='{ .spec.serviceAccountIssuer }'
```

The output should be your OIDC container URL:
```
https://<storage-account>.blob.core.windows.net/<container>
```

For Minikube clusters, ensure you started Minikube with the correct flag:
```bash
minikube start --extra-config=apiserver.service-account-issuer=https://<storage-account>.blob.core.windows.net/<container>
```

Verify that the OIDC configuration files are publicly accessible:
- `https://<storage-account>.blob.core.windows.net/<container>/keys.json`
- `https://<storage-account>.blob.core.windows.net/<container>/.well-known/openid-configuration`

You can test these URLs in your browser or with curl:
```bash
curl https://<storage-account>.blob.core.windows.net/<container>/keys.json
curl https://<storage-account>.blob.core.windows.net/<container>/.well-known/openid-configuration
```

#### 4) Wrong Audience Configuration

```
time="2024-03-19T10:25:30Z" level=warning msg="⏳ Temporary Error: could not authenticate with Azure, AADSTS50013: Assertion audience claim does not match the expected audience" sys=noobaa/noobaa
```

##### Solution:
Ensure the audience in all the three federated credential matches the audience in the projected service account token.

The federated credential in Azure should have:
- **Audience**: `openshift`

And the service account token projection should specify the same audience:
```yaml
volumes:
  - name: bound-sa-token
    projected:
      sources:
        - serviceAccountToken:
            path: token
            expirationSeconds: 3600
            audience: openshift
```

#### 5) Missing Azure Credentials in Secret

```
time="2024-03-19T10:30:15Z" level=info msg="identified azure region eastus" sys=noobaa/noobaa
time="2024-03-19T10:30:15Z" level=warning msg="⏳ Temporary Error: could not read Azure credentials, missing client_id or tenant_id" sys=noobaa/noobaa
```

##### Solution:
Check that the secret contains the required Azure STS credentials:

```bash
kubectl get secret noobaa-azure-cloud-creds-secret -n <your-namespace> -o json | jq -r '.data'
```

The secret should contain:
- `azure_client_id`: The Managed Identity's Client ID (base64 encoded)
- `azure_tenant_id`: The Azure Tenant ID (base64 encoded)
- `AccountName`: The storage account name (base64 encoded)

To decode and verify:
```bash
kubectl get secret noobaa-azure-cloud-creds-secret -n <your-namespace> -o json | jq -r '.data.azure_client_id' | base64 -d
kubectl get secret noobaa-azure-cloud-creds-secret -n <your-namespace> -o json | jq -r '.data.azure_tenant_id' | base64 -d
```

If the secret is missing or incorrect, you may need to recreate it with the correct values.

#### 6) Insufficient Permissions on Managed Identity

```
time="2024-03-19T10:35:20Z" level=warning msg="⏳ Temporary Error: Azure operation failed, AuthorizationFailed: The client '<client-id>' with object id '<object-id>' does not have authorization to perform action 'Microsoft.Storage/storageAccounts/write'" sys=noobaa/noobaa
```

##### Solution:
Verify that the Managed Identity has the required permissions:

Required Azure roles:
- **Storage Account Contributor**: To create and manage storage accounts
- **Storage Blob Data Contributor**: To create and manage blob containers and data

Check the Managed Identity's role assignments in Azure Console:
1. Go to Managed Identities → Select your Managed Identity
2. Click on "Azure role assignments"
3. Verify the roles are assigned at the appropriate scope (subscription or resource group)

To add missing roles via Azure CLI or from console:
```bash
# Get the Managed Identity's principal ID
PRINCIPAL_ID=$(az identity show --name <managed-identity-name> --resource-group <resource-group> --query principalId -o tsv)

# Assign Storage Account Contributor role
az role assignment create --assignee ${PRINCIPAL_ID} --role "Storage Account Contributor" --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group>

# Assign Storage Blob Data Contributor role
az role assignment create --assignee ${PRINCIPAL_ID} --role "Storage Blob Data Contributor" --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group>
```

#### 7) Token Not Projected in Pod

```
time="2024-03-19T10:40:25Z" level=warning msg="⏳ Temporary Error: could not read Azure identity token from path /var/run/secrets/openshift/serviceaccount/token, open /var/run/secrets/openshift/serviceaccount/token: no such file or directory" sys=noobaa/noobaa
```

##### Solution:
Verify that the service account token is properly projected into the pod.

Check the pod's volume mounts:
```bash
kubectl get pod <pod-name> -n <your-namespace> -o yaml | grep -A 10 volumeMounts
```

You should see a volume mount for the Azure identity token:
```yaml
volumeMounts:
  - mountPath: /var/run/secrets/openshift/serviceaccount
    name: azure-identity-token
```

And the corresponding volume:
```yaml
volumes:
  - name: bound-sa-token
    projected:
      sources:
        - serviceAccountToken:
            path: token
            expirationSeconds: 3600
            audience: openshift
```

If the volume is not configured, the NooBaa operator should recreate the pods with the correct configuration. Try deleting the pod:
```bash
kubectl delete pod <pod-name> -n <your-namespace>
```

#### 8) Network Connectivity Issues

```
time="2024-03-19T10:45:30Z" level=warning msg="⏳ Temporary Error: could not connect to Azure, Post \"https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token\": dial tcp: lookup login.microsoftonline.com: no such host" sys=noobaa/noobaa
```

##### Solution:
Verify network connectivity from the cluster to Azure endpoints:

Test connectivity to Azure login endpoint:
```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl -v https://login.microsoftonline.com
```

Test connectivity to Azure storage endpoint:
```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl -v https://<storage-account>.blob.core.windows.net
```

If connectivity fails, check:
- Network policies in the namespace
- Firewall rules
- Proxy configurations
- DNS resolution



#### 10) Other Issues

If you've verified all the above and still encounter issues:

1. Check the NooBaa operator logs for more detailed error messages:
```bash
kubectl logs -n <your-namespace> $(kubectl get pods -n <your-namespace> | grep operator | awk '{print $1}') -c noobaa-operator --tail=100
```

2. Check the NooBaa core logs:
```bash
kubectl logs -n <your-namespace> $(kubectl get pods -n <your-namespace> | grep core | awk '{print $1}') --tail=100
```

3. Verify the NooBaa system status:
```bash
kubectl get noobaa -n <your-namespace> -o yaml
```

4. If the issue persists after restart, check for any recent changes to:
   - Azure Managed Identity configuration
   - Federated credentials
   - OIDC configuration files
   - Cluster authentication settings

### Additional Resources

- [Create Azure STS Setup Guide](create_azure_sts_setup.md)
- [Azure Workload Identity Documentation](https://learn.microsoft.com/en-us/azure/active-directory/workload-identities/workload-identity-federation)
- [OpenShift Cloud Credential Operator - Azure Workload Identity](https://github.com/openshift/cloud-credential-operator/blob/master/docs/azure_workload_identity.md)