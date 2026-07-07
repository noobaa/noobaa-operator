# Troubleshoot GCP WIF (STS) Cluster

Common errors during GCP Workload Identity Federation (WIF) installations and how to investigate and fix them.

Open the operator logs:

```bash
kubectl logs <operator-pod> -n <your-namespace> -c noobaa-operator -f
```

## Main issues:

### 1) Wrong principal or missing Workload Identity User binding

Partial output from the operator logs:
Note: removed the project name and replace this place-holder `<project-name>`.

```shell
time="2026-06-11T08:04:05Z" level=error msg="got error when trying to create bucket noobaabucketdvxd1. error: Post \"https://storage.googleapis.com/storage/v1/b?alt=json&prettyPrint=false&project=<project-name>\": credentials: status code 403: {\n  \"error\": {\n    \"code\": 403,\n    \"message\": \"Permission 'iam.serviceAccounts.getAccessToken' denied on resource (or it may not exist).\",\n    \"status\": \"PERMISSION_DENIED\",\n    \"details\": [\n      {\n        \"@type\": \"type.googleapis.com/google.rpc.ErrorInfo\",\n        \"reason\": \"IAM_PERMISSION_DENIED\",\n        \"domain\": \"iam.googleapis.com\",\n        \"metadata\": {\n          \"permission\": \"iam.serviceAccounts.getAccessToken\"\n        }\n      }\n    ]\n  }\n}\n" sys=openshift-storage/noobaa
time="2026-06-11T08:04:05Z" level=info msg="SetPhase: temporary error during phase \"Configuring\"" sys=openshift-storage/noobaa
time="2026-06-11T08:04:05Z" level=warning msg="⏳ Temporary Error: Post \"https://storage.googleapis.com/storage/v1/b?alt=json&prettyPrint=false&project=<project-name>\": credentials: status code 403: {\n  \"error\": {\n    \"code\": 403,\n    \"message\": \"Permission 'iam.serviceAccounts.getAccessToken' denied on resource (or it may not exist).\",\n    \"status\": \"PERMISSION_DENIED\",\n    \"details\": [\n      {\n        \"@type\": \"type.googleapis.com/google.rpc.ErrorInfo\",\n        \"reason\": \"IAM_PERMISSION_DENIED\",\n        \"domain\": \"iam.googleapis.com\",\n        \"metadata\": {\n          \"permission\": \"iam.serviceAccounts.getAccessToken\"\n        }\n      }\n    ]\n  }\n}\n" sys=openshift-storage/noobaa
time="2026-06-11T08:04:05Z" level=info msg="UpdateStatus: Done generation 1" sys=openshift-storage/noobaa
```

This error happens during default backing store creation. The operator uses the CCO secret built from the noobaa operator service account token, so fixing the `noobaa` principal binding usually resolves this specific failure. You still need bindings for `noobaa-core` and `noobaa-endpoint` for normal system operation.

#### Solution:
In GCP Console → **IAM & Admin** → **Service Accounts** → select your service account → **Principals with access**, verify a **Workload Identity User** binding exists for each NooBaa service account:

```text
principal://iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/subject/system:serviceaccount:<your-namespace>:noobaa
principal://iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/subject/system:serviceaccount:<your-namespace>:noobaa-core
principal://iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/subject/system:serviceaccount:<your-namespace>:noobaa-endpoint
```

Decode the token to confirm the `sub` claim matches the principal:

```bash
MY_TOKEN_OPERATOR=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep operator | awk '{ print $1}') -c noobaa-operator -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)

echo ${MY_TOKEN_OPERATOR} | cut -d '.' -f 2 | base64 -d | jq .
```

### 2) Missing service account in principal bindings

If only the operator binding exists, core or endpoint pods will fail when they use their own tokens.

#### Solution:
Verify that every pod that accesses GCP (operator, core, and endpoint) has a matching principal binding. Run the steps from issue 1 for each pod — for example, in the core pod:

```bash
MY_TOKEN_CORE=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep core | awk '{ print $1}') -c core -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
echo ${MY_TOKEN_CORE} | cut -d '.' -f 2 | base64 -d | jq .sub
```

Repeat for the operator and endpoint pods. Each token has a different `sub` claim; all three must be bound in GCP.

### Additional Resources

- [Create GCP WIF (STS) Setup On Minikube With NooBaa](create_gcp_wif_sts_setup_on_minikube_with_noobaa.md)
- [Create GCP WIF (STS) Setup On Minikube](create_gcp_wif_sts_setup_on_minikube.md)
