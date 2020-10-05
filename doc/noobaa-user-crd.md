[NooBaa Operator](../README.md) /
# NooBaaUser CRD

NooBaaUser CRD is used in order to receive new user credantials for accessing different noobaa services (like S3 buckets for example)

This are the main fields which can be provided to this CRD:
- Email - an optional email address to be associated with this user 
- AllowedBuckets - if fullPermission is true - all current and future buckets will be allowed full access by this user.
- AllowedBuckets - if fullPermission is false - all the buckets in the permissionList will be allowed full access by this user.
- AllowBucketCreate - the user will be able to create new buckets on top of the user default backing store
- DefaultBackingStore - the name of the backingstore new buckets created by the user will be created on

Constraints:
- If allowed to (allowBucketCreate!=false), NooBaaUser by default will be able to create new buckets using s3: on the defaultBackingStore if provided or on the noobaa-default-backing-store if not.
- If no allowedBuckets will be provided, the user won't be able to access any bucket and can only create new ones if is allowed to (allowBucketCreate!=false)
- If allowedBuckets is fullPermission, all current buckets and all future buckets will be allowed access by this user. permissionList if provided, will be overlooked

For more information on using noobaa-user from S3 see [S3 Account](s3-account.md). #TBD

# Definitions

- CRD: [noobaa.io_noobaauser_crd.yaml](../deploy/crds/noobaa.io_noobaausers_crd.yaml)
- CR: [noobaa.io_v1alpha1_noobaauser_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_noobaauser_cr.yaml)


# Reconcile

- The operator will verify that noobaa-user is valid - i.e. the buckets in allowed buckets really exists.
- Changes to a noobaa-user spec will be propagated to noobaa accounts that were instantiated from it.

# Read Status

Here is an example of healthy status:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaUser
metadata:
  name: noobaa-user-1
  namespace: noobaa
spec:
  ...
status:
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa user is ready
    reason: NooBaaUserPhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa user is ready
    reason: NooBaaUserPhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-05T13:50:50Z"
    message: noobaa operator completed reconcile - noobaa user is ready
    reason: NooBaaUserPhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - noobaa user is ready
    reason: NooBaaUserPhaseReady
    status: "True"
    type: Upgradeable
  phase: Ready
```


# Example

Here are some examples of the cli/YAML usage and NooBaaUser CRs for the different noobaa-users configurations:

user1 will be allowed to create additional buckets, on top of the default backingstore, and to access first.bucket only:
```shell
noobaa -n noobaa noobaauser create user1 --allowed_buckets first.bucket --default_backingstore noobaa-default-backing-store
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaUser
metadata:
  labels:
    app: noobaa
  name: user1
  namespace: noobaa
spec:
  allowedBuckets:
    permissionList: 
    - first.bucket
  allowBucketCreate: true
  defaultBackingStore: noobaa-default-backing-store
```

user1 will not be allowed to create additional buckets, and to access first.bucket and second.bucket only:
```shell
noobaa -n noobaa noobaauser create user1 --allowed_buckets first.bucket, second.bucket --allow_bucket_create=false
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaUser
metadata:
  labels:
    app: noobaa
  name: user1
  namespace: noobaa
spec:
  allowedBuckets:
    permissionList:
    - first.bucket
    - second.bucket
  allowBucketCreate: false
```

user1 will be able to access all buckets
```shell
noobaa -n noobaa noobaauser create user1 --allowed_buckets all_buckets
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaUser
metadata:
  labels:
    app: noobaa
  name: user1
  namespace: noobaa
spec:
  allowedBuckets:
    fullPermission: true
```

user1 will be attached with an email address and will be able to only create new buckets on top backingstore bs1
```shell
noobaa -n noobaa noobaauser create user1 --allow_bucket_create --default_backingstore bs1
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaaUser
metadata:
  labels:
    app: noobaa
  name: user1
  namespace: noobaa
spec:
  email: user1@redhat.com
  allowBucketCreate: true
  defaultBackingStore: bs1
```
