[NooBaa Operator](../README.md) /
# NooBaaAccount CRD

NooBaaAccount CRD is used in order to receive new credentials set for accessing different noobaa services (like S3 buckets for example)

This are the main fields which can be provided to this CRD:
- AllowedBuckets - if fullPermission is true - all current and future buckets will be allowed full access by this account.
- AllowedBuckets - if fullPermission is false - all the buckets in the permissionList will be allowed full access by this account.
- AllowBucketCreate - the account will be able to create new buckets on top of the account's default backingstore
- DefaultBackingStore - the name of the backingstore new buckets created by the account will be created on

Constraints:
- If allowed to (allowBucketCreate!=false), NooBaaAccount by default will be able to create new buckets using s3: on the defaultResource
- If no allowedBuckets will be provided, the account won't be able to access any bucket and can only create new ones if is allowed to (allowBucketCreate!=false)
- If allowedBuckets is fullPermission, all current buckets and all future buckets will be allowed access by this account. permissionList if provided, will be overlooked

For more information on using noobaa-account from S3 see [S3 Account](s3-account.md). #TBD

# Definitions

- CRD: [noobaa.io_noobaaaccount_crd.yaml](../deploy/crds/noobaa.io_noobaaaccounts_crd.yaml)
- CR: [noobaa.io_v1alpha1_noobaaaccount_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_noobaaaccount_cr.yaml)


# Reconcile

- The operator will verify that noobaa-account is valid - i.e. the buckets in allowed buckets really exists.
- Changes to a noobaa-account spec will be propagated to noobaa accounts that were instantiated from it.

# Read Status

Here is an example of healthy status:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  name: noobaa-account-1
  namespace: noobaa
spec:
  ...
status:
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa account is ready
    reason: NooBaaAccountPhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa account is ready
    reason: NooBaaAccountPhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-05T13:50:50Z"
    message: noobaa operator completed reconcile - noobaa account is ready
    reason: NooBaaAccountPhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa account is ready
    reason: NooBaaAccountPhaseReady
    status: "True"
    type: Upgradeable
  phase: Ready
```
Once the Account is reconciled by the operator, an account will be created in NooBaa, and the operator will create a Secret with the same name of the noobaa account on the same namespace of the account. For the example above, the Secret will be named `noobaa-account-1`.

The content of the Secret provides all the information needed by the application in order to connect to the system using S3 API. The user should configure its S3 SDK to use the AWS_ACCESS_KEY_ID & AWS_SECRET_ACCESS_KEY credentials as provided by the Secret


# Example

Here are some examples of the cli/YAML usage and NooBaaAccount CRs for the different noobaa-accounts configurations:

account1 will be allowed to create additional buckets, on top of the default backingstore, and to access first.bucket only:
```shell
noobaa -n noobaa account create account1 --allowed_buckets first.bucket --default_resource noobaa-default-backing-store
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  labels:
    app: noobaa
  name: account1
  namespace: noobaa
spec:
  allowedBuckets:
    permissionList: 
    - first.bucket
  allowBucketCreate: true
  defaultResource: noobaa-default-backing-store
```

account1 will not be allowed to create additional buckets, and to access first.bucket and second.bucket only:
```shell
noobaa -n noobaa account create account1 --allowed_buckets first.bucket, second.bucket --allow_bucket_create=false
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  labels:
    app: noobaa
  name: account1
  namespace: noobaa
spec:
  allowedBuckets:
    permissionList:
    - first.bucket
    - second.bucket
  allowBucketCreate: false
```

account1 will be able to access all buckets
```shell
noobaa -n noobaa account create account1 --full_permission
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  labels:
    app: noobaa
  name: account1
  namespace: noobaa
spec:
  allowedBuckets:
    fullPermission: true
```

account1 will be able to access only his own newly created buckets on top of backingstore bs1
```shell
noobaa -n noobaa account create account1 --allow_bucket_create --default_resource bs1
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  labels:
    app: noobaa
  name: account1
  namespace: noobaa
spec:
  allowBucketCreate: true
  defaultResource: bs1
```
