[NooBaa Operator](../README.md) /
# BackingStore CRD

The BackingStore CRD represents a storage target to be used as underlying storage for the data in NooBaa buckets.
These storage targets are used to store deduplicated, compressed and encrypted chunks of data. The encryption keys are stored separately.
BackingStores are referred to by name when defining [a BucketClass](bucket-class-crd.md).

Supported BackingStore types: 
- aws-s3
- s3-compatible
- ibm-cos
- google-cloud-storage
- azure-blob
- pv-pool

It is also possible to add new backing store types by providing a GET/PUT key-value store, see [backing stores supported by NooBaa](https://github.com/noobaa/noobaa-core/tree/master/src/agent/block_store_services).


# Definitions

- CRD: [noobaa.io_backingstores_crd.yaml](../deploy/crds/noobaa.io_backingstores_crd.yaml)
- CR: [noobaa.io_v1alpha1_backingstore_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_backingstore_cr.yaml)


# Reconciliation
It is possible to create backingstores by using the NooBaa CLI tool (from here on referred to in shell commands as `noobaa`), or by applying YAMLs.

From here on out, fields that depend on the user's choice and input will be marked with `<>`. In cases where the value might not be inferrable from context, an additional explanation might be included.
Note that the commands will apply to the namespace that's currently active (can be checked with `kubectl config current-context`). A command-specific namespace can be set by adding the `-n/--namespace <NAMESPACE>` flag.
Also note that, if possible and applicable, omitting the `access-key` and `secret-key` flags is recommended in order to avoid leakage of secrets. When both flags are ommitted, the CLI will prompt the user to enter the keys interactively.

If the user opts to use the CLI's `--secret-name` option, or to apply a YAML, please see [Store Secret Creation](store-connection-secrets.md).
In case of YAML application, the value under `secret.namespace` needs to point to the secret's namespace.

## AWS S3
Uses the S3 API for storing encrypted chunks of data in AWS buckets
```shell
noobaa backingstore create aws-s3 <BACKINGSTORE NAME> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
Simlarly to `AWS-S3` this backingstore uses the S3 API for storing encrypted chunks of data in AWS buckets.
However, the difference between the two backingstore types lies in the authentication method. AWS S3 uses a pair of static, user-provided access keys, while AWS S3 STS uses an Amazon Resource Name (ARN) provided by the user to create credentials for every single interaction with AWS by utilizing [AssumeRuleWithWebIdentity](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithWebIdentity.html).
This type of backingstore is useful in cases where the user wishes to limit access to their AWS cloud for a specific amount of time, and for easier management of the cloud's security.

Prior to using this backingstore, an OpenIDConnect provider needs to be set up, which is outside the scope of these docs.

Please note that once the set session duration expires, the backingstore will no longer work, and writing and reading data will not be possible any longer, without editing the backingstore YAML and replacing the ARN manually.
```shell
noobaa backingstore create aws-sts-s3 <BACKINGSTORE NAME> --target-bucket <> --aws-sts-arn <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
Uses the S3 API for storing encrypted chunks of data in buckets that can be interacted with via an S3-compatible API
```shell
noobaa backingstore create s3-compatible <BACKINGSTORE NAME> --endpoint <> --signature-version <> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
Uses the IBM COS API for storing encrypted chunks of data in IBM COS buckets
```shell
noobaa backingstore create ibm-cos bs --endpoint <> --access-key <> --secret-key <> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
Uses the Google Cloud Storage API for storing encrypted chunks of data in Google Cloud buckets
```shell
noobaa backingstore create google-cloud-storage <BACKINGSTORE NAME> --private-key-json-file <PATH TO credentials.json> --target-bucket <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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
Uses the Azure Blob API for storing encrypted chunks of data in an Azure container
```shell
noobaa backingstore create azure-blob <BACKINGSTORE NAME> --account-key <> --account-name <> --target-blob-container <>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
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

## Persistent Volume Pool
A unique kind of backingstore that uses local storage resources instead of ones provided by cloud services.
Creates a StatefulSet with a PVC mounted in each pod. Each resource will connect to the NooBaa core and provide the PV filesystem storage to be used for storing encrypted chunks of data. It is possible to configure the number of pods to be used and their PV size.

Each PV needs at least 17 Gibibytes to function properly. The bare minimum defined in the source code is 16 Gibibytes, but it is better to take some extra space.

*In the YAML, the value of `storage` should adhere to the [K8s Memory Resource Units format](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory).
```shell
noobaa backingstore create pv-pool <BACKINGSTORE NAME> --num-volumes <> --pv-size-gb <INT >= 17> --storage-class <THE SC THAT WILL BE USED FOR PV PROVISIONING>
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: <>
  namespace: <>
spec:
  pvPool:
    numVolumes: <>
    resources:
      requests:
        storage: <SEE NOTE ABOVE*>
    storageClass: STORAGE-CLASS-NAME
  type: pv-pool
```

## Examples
### AWS S3
```shell
noobaa backingstore create aws-s3 aws-backingstore --namespace app-namespace --access-key AKIAIOSFODNN7EXAMPLE --secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --target-bucket personal-bucket
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: aws-backingstore
  namespace: app-namespace
spec:
  awsS3:
    secret:
      name: user-created-aws-s3-secret
      namespace: secret-namespace
    targetBucket: personal-bucket
  type: aws-s3
```

### AWS S3 STS
```shell
noobaa backingstore create aws-sts-s3 aws-sts-backingstore --target-bucket personal-bucket --aws-sts-arn arn:aws:iam::111111111111:role/noobaa_sts
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: aws-sts-backingstore
  namespace: app-namespace
spec:
  awsS3:
    awsSTSRoleARN: arn:aws:iam::111111111111:role/noobaa_sts
    targetBucket: personal-bucket
    secret: {}
  type: aws-s3
```

### Persistent Volume Pool
```shell
noobaa backingstore create pv-pool pv-backingstore --num-volumes 3 --pv-size-gb 17 --storage-class local-path
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  finalizers:
  - noobaa.io/finalizer
  name: pv-backingstore
  namespace: app-namespace
spec:
  pvPool:
    numVolumes: 3
    resources:
      requests:
        storage: 17Gi
    storageClass: local-path
  type: pv-pool
```


## Modifying a Backing Store's Credentials
If a user wishes to change a backing store's credentials, the appropriate secret should be updated by the user, and the operator will propagate the new credentials to the system server.

A backing store's secret can be found in the CR's YAML's `spec` field, under a key named `secret`.

# Resource Status
It is possible to check a resource's status in several ways, including:
- `kubectl get backingstore <NAME> -o yaml`
- `kubectl describe backingstore <NAME>`
- `noobaa backingstore status <NAME>`

Below is an example of a healthy backing store's status, as retrieved with the first command:

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
    mode:
      modeCode: OPTIMAL
      ...
    phase: Ready
```
An example for an unhealthy backing store can be seen at the bottom of the next section.

# Deletion
Backing stores are used for data persistency; therefore, a cleanup process must run before they may be deleted.
The operator will use [the `finalizer` pattern](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers), and set a finalizer on every backing store to mark that external cleanup is needed before it can be deleted.

After marking a backingstore for deletion, the operator will notify the NooBaa server of the deletion, which will enter a *decommissioning* state, in which NooBaa will attempt to rebuild the data to a new backing store location. Once the decommissioning process completes the operator will remove the finalizer and allow the CR to be deleted.

There are cases where the decommissioning cannot be completed due to an inability to read the data from the backing store. For example, in cases where the target bucket was already deleted, the credentials were invalidated, or the system cannot connect to the target storage. In these cases the system status will be used to report the issue and suggest a manual resolution. For example:

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
