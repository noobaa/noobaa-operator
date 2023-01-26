[NooBaa Operator](../README.md) /
# NamespaceStore CRD

NamespaceStore CRD represents an underlying storage to be used as read or write target for the data in NooBaa 
namespace buckets.
These storage targets are used to store plain data.
Namespace-Store are referred to by name when defining [BucketClass](bucket-class-crd.md).
Multiple types of Namespace-Store are currently supported: aws-s3, s3-compatible, ibm-cos, google-cloud-storage, azure-blob.

# Definitions

- CRD: [noobaa.io_NamespaceStores_crd.yaml](../deploy/crds/noobaa.io_namespacestores_crd.yaml)
- CR: [noobaa.io_v1alpha1_NamespaceStore_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_namespacestore_cr.yaml)


# Reconcile

#### AWS-S3 type

Create a cloud resource within the NooBaa brain and use S3 API for reading or writing data in the AWS cloud.
```shell
noobaa -n noobaa namespacestore create aws-s3 bs --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  awsS3:
    secret:
      name: namespace-store-aws-s3-bs
      namespace: noobaa
    targetBucket: BUCKET
  type: aws-s3
```

#### S3-COMPATIBLE type

Create a cloud resource within the NooBaa brain and use S3 API for reading or writing data in any S3 API compatible endpoint.
```shell
noobaa -n noobaa namespacestore create s3-compatible bs --endpoint ENDPOINT --signature-version v4 --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  s3Compatible:
    endpoint: ENDPOINT
    secret:
      name: namespace-store-s3-compatible-bs
      namespace: noobaa
    signatureVersion: v4
    targetBucket: BUCKET
  type: s3-compatible
```

#### IBM-COS type

Create a cloud resource within the NooBaa brain and use IBM COS API for reading or writing data in any IBM COS API compatible endpoint.
```shell
noobaa -n noobaa namespacestore create ibm-cos bs --endpoint ENDPOINT --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  s3Compatible:
    endpoint: ENDPOINT
    secret:
      name: namespace-store-ibm-cos-bs
      namespace: noobaa
    signatureVersion: v2
    targetBucket: BUCKET
  type: ibm-cos
```

#### GOOGLE-CLOUD-STORAGE type

Create a cloud resource within the NooBaa brain and use Google Cloud Storage API for reading or writing data in any Google Cloud Storage API compatible endpoint.
```shell
noobaa -n noobaa namespacestore create google-cloud-storage ns --private-key-json-file key.json --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  googleCloudStorage:
    secret:
      name: namespace-store-google-cloud-storage-bs
      namespace: noobaa
    targetBucket: BUCKET
  type: google-cloud-storage
```

#### AZURE-BLOB type

Create a cloud resource within the NooBaa brain and use BLOB API for reading or writing data in Azure cloud.
```shell
noobaa -n noobaa namespacestore create azure-blob bs --account-key KEY --account-name NAME --target-blob-container CONTAINER
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  azureBlob:
    secret:
      name: namespace-store-azure-blob-bs
      namespace: noobaa
    targetBlobContainer: CONTAINER
  type: azure-blob
```

#### Credentials change

In case the credentials of a Namespace-store need to be updated due to periodic security policy or concern, the appropriate secret should be updated by the user, and the operator will be responsible for watching changes in those secrets and propagating the new credential update to the NooBaa system server.


# Read Status

Here is an example of healthy status (see below an example of non-healthy status):

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
  phase: Ready
```


# Delete

Namespace-Store is used for data persistency, therefore there is a cleanup process before they can be deleted.
The operator will use the `finalizer` pattern as explained in the link below, and set a finalizer on every Namespace-store to mark that external cleanup is needed before it can be deleted:

https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers

After marking a Namespace-store for deletion, the operator will notify the NooBaa server on the deletion which will enter a *decommissioning* state, in which NooBaa will attempt to rebuild the data to a new Namespace-store location. Once the decommissioning process completes the operator will remove the finalizer and allow the CR to be deleted.

There are cases where the decommissioning cannot complete due to the inability to read the data from the Namespace-store that is already not serving - for example, if the target bucket was already deleted or the credentials were invalidated or there is no network from the system to the Namespace-store service. In such cases the system status will be used to report these issues and suggest a manual resolution for example:

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
