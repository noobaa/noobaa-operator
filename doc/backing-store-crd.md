[NooBaa Operator](../README.md) /
# BackingStore CRD

BackingStore CRD represents a storage target to be used as underlying storage for the data in NooBaa buckets.
These storage targets are used to store deduped+compressed+encrypted chunks of data (encryption keys are stored separately).
Backing-stores are referred to by name when defining [BucketClass](bucket-class-crd.md).

Multiple types of backing-stores are currently supported: aws-s3, s3-compatible, ibm-cos, google-cloud-storage, azure-blob, pv-pool.
Backing-store type pv-pool is not yet supported by the operator. Instead, the web management console must be used to administer pv-pool backing-stores.
Adding support for a new type of backing-store is rather easy as it requires just GET/PUT key-value store, see [Backing-stores supported by NooBaa](https://github.com/noobaa/noobaa-core/tree/master/src/agent/block_store_services).


# Definitions

- CRD: [noobaa.io_backingstores_crd.yaml](../deploy/crds/noobaa.io_backingstores_crd.yaml)
- CR: [noobaa.io_v1alpha1_backingstore_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_backingstore_cr.yaml)


# Reconcile

#### AWS-S3 type

Create a cloud resource within the NooBaa brain and use S3 API for storing encrypted chunks of data in the AWS cloud.
```shell
noobaa -n noobaa backingstore create aws-s3 bs --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
      name: backing-store-aws-s3-bs
      namespace: noobaa
    targetBucket: BUCKET
  type: aws-s3
```

#### S3-COMPATIBLE type

Create a cloud resource within the NooBaa brain and use S3 API for storing encrypted chunks of data in any S3 API compatible endpoint.
```shell
noobaa -n noobaa backingstore create s3-compatible bs --endpoint ENDPOINT --signature-version v4 --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
      name: backing-store-s3-compatible-bs
      namespace: noobaa
    signatureVersion: v4
    targetBucket: BUCKET
  type: s3-compatible
```

#### IBM-COS type

Create a cloud resource within the NooBaa brain and use IBM COS API for storing encrypted chunks of data in any IBM COS API compatible endpoint.
```shell
noobaa -n noobaa backingstore create ibm-cos bs --endpoint ENDPOINT --access-key KEY --secret-key SECRET --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
      name: backing-store-ibm-cos-bs
      namespace: noobaa
    signatureVersion: v2
    targetBucket: BUCKET
  type: ibm-cos
```

#### GOOGLE-CLOUD-STORAGE type

Create a cloud resource within the NooBaa brain and use Google Cloud Storage API for storing encrypted chunks of data in Google Cloud Storage.
```shell
noobaa -n noobaa backingstore create google-cloud-storage bs --private-key-json-file key.json --target-bucket BUCKET
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
      name: backing-store-google-cloud-storage-bs
      namespace: noobaa
    targetBucket: BUCKET
  type: google-cloud-storage
```

#### AZURE-BLOB type

Create a cloud resource within the NooBaa brain and use BLOB API for storing encrypted chunks of data in Azure cloud.
```shell
noobaa -n noobaa backingstore create azure-blob bs --account-key KEY --account-name NAME --target-blob-container CONTAINER
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
      name: backing-store-azure-blob-bs
      namespace: noobaa
    targetBlobContainer: CONTAINER
  type: azure-blob
```

#### PV-POOL type

Create NooBaa resources StatefulSet with PVC mounted in each pod. Each resource will connect to the NooBaa core and provide the PV filesystem storage to be used for storing encrypted chunks of data. It is possible to configure the number of pods to be used and their PV size.
```shell
noobaa -n noobaa backingstore create pv-pool bs --num-volumes 3 --pv-size-gb 32 --storage-class STORAGE-CLASS-NAME
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  labels:
    app: noobaa
  name: bs
  namespace: noobaa
spec:
  pvPool:
    numVolumes: 3
    resources:
      requests:
        storage: 32Gi
    storageClass: STORAGE-CLASS-NAME
  type: pv-pool
```


#### Credentials change

In case the credentials of a backing-store need to be updated due to a periodic security policy or concern, the appropriate secret should be updated by the user, and the operator will be responsible for watching changes in those secrets and propagating the new credential update to the NooBaa system server.


# Read Status

Here is an example healthy status (see below example of non-healthy status):

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  name: aws-s3
  namespace: noobaa
spec:
  ...
status:
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - backing store is ready
    reason: BackingStorePhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - backing store is ready
    reason: BackingStorePhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-05T13:50:50Z"
    message: noobaa operator completed reconcile - backing store is ready
    reason: BackingStorePhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-06T07:03:46Z"
    message: noobaa operator completed reconcile - backing store is ready
    reason: BackingStorePhaseReady
    status: "True"
    type: Upgradeable
  phase: Ready
```


# Delete

Backing-stores are used for data persistency, therefore there is a cleanup process before they can be deleted.
The operator will use the `finalizer` pattern as explained in the link below, and set a finalizer on every backing-store to mark that external cleanup is needed before it can be delete:

https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers

After marking a backing-store for deletion, the operator will notify the NooBaa server on the deletion which will enter a *decommissioning* state, in which NooBaa will attempt to rebuild the data to a new backing-store location. Once the decomissioning process completes the operator will remove the finalizer and allow the CR to be deleted.

There are cases where the decommissioning cannot complete due to inability to read the data from the backing-store that is already not serving - for example if the target bucket was already deleted or the credentials were invalidated or there is no network from the system to the backing-store service. In such cases the system status will be used to report these issues and suggest manual resolution for example:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
    message: BackingStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: Unknown
    type: Available
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: BackingStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: BackingStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: "True"
    type: Degraded
  - lastHeartbeatTime: "2019-11-06T14:06:35Z"
    lastTransitionTime: "2019-11-06T14:11:36Z"
    message: BackingStore "bs" invalid external connection "INVALID_CREDENTIALS"
    reason: INVALID_CREDENTIALS
    status: Unknown
    type: Upgradeable
  phase: Rejected

```
