[NooBaa Operator](../README.md) /
# BucketClass CRD

NooBaaBucket CRD represents a class for buckets that defines policies for data placement and more.

Data placement capabilities are built as a multi-layer structure, here are the layers bottom-up:
- Spread Layer - list of backing-stores, aggregates the storage of multiple stores.
- Mirroring Layer - list of spread-layers, async-mirroring to all mirrors, with locality optimization (will allocate on the closest region to the source endpoint), mirroring requires at least two backing-stores.
- Tiering Layer - list of mirroring-layers, push cold data to next tier.

Constraints:
A backing-store name may appear in more than one bucket-class but may not appear more than once in a single bucket-class.
The operator cli currently only supports a single tier placement-policy for a bucket-class. 
YAML must be used to create a bucket-class with a placement-policy that has multiple tiers.
Placement-policy is case sensitive and should be of value (Mirror|Spread).

For more information on using bucket-classes from S3 see [S3 Account](s3-account.md).

# Definitions

- CRD: [noobaa_v1alpha1_bucketclass_crd.yaml](../deploy/crds/noobaa_v1alpha1_bucketclass_crd.yaml)
- CR: [noobaa_v1alpha1_bucketclass_cr.yaml](../deploy/crds/noobaa_v1alpha1_bucketclass_cr.yaml)


# Reconcile

- The operator will verify that bucket-class is valid - i.e. that the backing-stores exist and can be used.
- Changes to a bucket-class spec will be propagated to buckets that were instantiated from it.
- Other than that the bucket-class is passive, just waiting there for new buckets to use it.

# Read Status

Here is an example of healthy status:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: noobaa-default-class
  namespace: noobaa
spec:
  ...
status:
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - bucket class is ready
    reason: BucketClassPhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - bucket class is ready
    reason: BucketClassPhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-05T13:50:50Z"
    message: noobaa operator completed reconcile - bucket class is ready
    reason: BucketClassPhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:50Z"
    lastTransitionTime: "2019-11-07T07:03:58Z"
    message: noobaa operator completed reconcile - bucket class is ready
    reason: BucketClassPhaseReady
    status: "True"
    type: Upgradeable
  phase: Ready
```


# Example

Here are some examples of the cli/YAML usage and BucketClass CRs for the different bucket-class configurations:

Single tier, single backing-store, placement Spread:
```shell
noobaa -n noobaa bucketclass create bc --backingstores bs --placement Spread
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs
      placement: Spread
```

Single tier, two backing-stores, placement Spread:
```shell
noobaa -n noobaa bucketclass create bc --backingstores bs1,bs2 --placement Spread
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      - bs2
      placement: Spread
```

Single tier, two backing-store, placement Mirror:
```shell
noobaa -n noobaa bucketclass create bc --backingstores bs1,bs2 --placement Mirror
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      - bs2
      placement: Mirror
```

Two tiers (not yet supported via operator cli), single backing-store per tier, placement Spread in tiers:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      placement: Spread
    - backingStores:
      - bs2
      placement: Spread
```

Two tiers (not yet supported via operator cli), two backing-store per tier, placement Spread in first tier and Mirror in second tier:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      - bs2
      placement: Spread
    - backingStores:
      - bs3
      - bs4
      placement: Mirror
```
