[NooBaa Operator](../README.md) /
# NamespaceStore CRD
The NamespaceStore CRD represents a storage target to be used as underlying storage for the data in NooBaa buckets (i.e. the buckets that are created by using the product)
These storage targets are used to store and read plain data.
NamespaceStores are referred to by name when defining [a BucketClass](bucket-class-crd.md).

Supported NamespaceStore types:
- aws-s3
- s3-compatible
- ibm-cos
- google-cloud-storage
- azure-blob

# Definitions
- CRD: [noobaa.io_NamespaceStores_crd.yaml](../deploy/crds/noobaa.io_namespacestores_crd.yaml)
- CR: [noobaa.io_v1alpha1_NamespaceStore_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_namespacestore_cr.yaml)


# Reconciliation
It is possible to create namespace stores by using the NooBaa CLI tool (from here on referred to in shell commands as `noobaa`), or by applying YAMLs.

From here on out, fields that depend on the user's choice and input will be marked with `<>`. In cases where the value might not be inferrable from context, an additional explanation might be included.
Note that the commands will apply to the namespace that's currently active (can be checked with `kubectl config current-context`). A command-specific namespace can be set by adding the `-n <NAMESPACE>` flag.
Also note that, if possible and applicable, omitting the `access-key` and `secret-key` flags is recommended in order to avoid leakage of secrets. When both flags are ommitted, the CLI will prompt the user to enter the keys interactively.

If you choose to use the CLI's `--secret-name` option, or to apply a YAML, please see [Store Secret Creation](store-connection-secrets.md).

## AWS S3
Uses the S3 API for IO operations on plain data in AWS buckets
```shell
noobaa namespacestore create aws-s3 <NAMESPACESTORE NAME> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  awsS3:
    secret:
      name: <>
      namespace: <>
    targetBucket: <>
  type: aws-s3
```

## AWS-S3 Security Token Service (STS)
Simlarly to `AWS-S3` this namespacestore uses the S3 API for storing plain data in AWS buckets.
However, the difference between the two namespacestore types lies in the authentication method. AWS S3 uses a pair of static, user-provided access keys, while AWS S3 STS uses an Amazon Resource Name (ARN) provided by the user to create credentials for every single interaction with AWS by utilizing [AssumeRuleWithWebIdentity](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html).
This type of namespacestore is useful in cases where the user wishes to limit access to their AWS cloud for a specific amount of time, and for easier management of the cloud's security.

Prior to using this namespacestore, an OpenIDConnect provider needs to be set up, which is outside the scope of these docs.

Please note that once the set session duration expires, the namespacestore will no longer work, and writing and reading data will not be possible any longer, without editing the namespacestore YAML and replacing the ARN manually.
```shell
noobaa namespacestore create aws-sts-s3 <NAMESPACESTORE NAME> --target-bucket <> --aws-sts-arn <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  awsS3:
    awsSTSRoleARN: <>
    targetBucket: <>
    secret:
      name: <>
      namespace: <>
  type: aws-s3
```

## S3 Compatible
Uses the S3 API for IO operations on plain data in buckets that can be interacted with via an S3-compatible API
```shell
noobaa namespacestore create s3-compatible <NAMESPACESTORE NAME> --endpoint <> --signature-version <> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  s3Compatible:
    endpoint: <>
    secret:
      name: <>
      namespace: <>
    signatureVersion: <>
    targetBucket: <>
  type: s3-compatible
```

## IBM COS
Uses the IBM COS API for IO operations on plain data in IBM COS buckets
```shell
noobaa namespacestore create ibm-cos bs --endpoint <> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  s3Compatible:
    endpoint: <>
    secret:
      name: <>
      namespace: <>
    signatureVersion: <>
    targetBucket: <>
  type: ibm-cos
```

## Google Cloud Storage
Uses the Google Cloud Storage API for IO operations on plain data in Google Cloud buckets
```shell
noobaa namespacestore create google-cloud-storage <NAMESPACESTORE NAME> --private-key-json-file <PATH TO credentials.json> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  googleCloudStorage:
    secret:
      name: <>
      namespace: <>
    targetBucket: <>
  type: google-cloud-storage
```

## Azure Blob Storage
Uses the Azure Blob API for IO operations on plain data in an Azure container
```shell
noobaa namespacestore create azure-blob <NAMESPACESTORE NAME> --account-key <> --account-name <> --target-blob-container <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  azureBlob:
    secret:
      name: <>
      namespace: <>
    targetBlobContainer: <>
  type: azure-blob
```

## Examples
### AWS S3
Note that the keys below are example keys from the AWS documentation, and will not work.
```shell
noobaa namespacestore create aws-s3 aws-namespacestore --access-key EXAMPLEAKIAIOSFODNN7 --secret-key EXAMPLEKEYwJalrXUtnFEMI/K7MDENG/bPxRfiCY --target-bucket personal-bucket
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: aws-namespacestore
  namespace: app-namespace
spec:
  awsS3:
    secret:
      name: user-created-aws-s3-secret
      namespace: secret-namespace
    targetBucket: personal-bucket
  type: aws-s3
```

## Modifying a Namespace Store's Credentials
If a user wishes to change a namespace store's credentials, the appropriate secret custom resource should be edited and updated by the user, and the operator will be propagate the new credentials to the system server.


# Resource Status
Below is an example of a healthy namespace store's status:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  name: aws-s3
  namespace: noobaa
spec:
  ...
status:
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - namespace store is ready
    reason: NamespaceStorePhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - namespace store is ready
    reason: NamespaceStorePhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-05T13:50:50Z"
    message: noobaa operator completed reconcile - namespace store is ready
    reason: NamespaceStorePhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - namespace store is ready
    reason: NamespaceStorePhaseReady
    status: "True"
    type: Upgradeable
    mode:
      modeCode: OPTIMAL
      ...
  phase: Ready
```


# Deletion
Namespace stores are used for data persistency; therefore, a cleanup process must run before they may be deleted.
The operator will use [the `finalizer` pattern](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers), and set a finalizer on every namespace store to mark that external cleanup is needed before it can be deleted.

After marking a namespace store for deletion, the operator will notify the NooBaa server of the deletion, which will enter a *decommissioning* state, in which NooBaa will attempt to rebuild the data to a new namespace store location. Once the decommissioning process completes the operator will remove the finalizer and allow the CR to be deleted.

There are cases where the decommissioning cannot be completed due to an inability to read the data from the namespace store. For example, in cases where the target bucket was already deleted, the credentials were invalidated, or the system cannot connect to the target storage. In these cases the system status will be used to report the issue and suggest a manual resolution. For example, this unhealthy namespacestore:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  name: aws-s3
  namespace: noobaa
  finalizers:
    - noobaa.io/finalizer
spec:
  ...
status:
    conditions:
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: NamespaceStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: Unknown
    type: Available
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: NamespaceStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: NamespaceStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: "True"
    type: Degraded
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: NamespaceStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: Unknown
    type: Upgradeable
  phase: Rejected

```
