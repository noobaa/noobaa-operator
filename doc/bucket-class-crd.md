[NooBaa Operator](../README.md) /
# BucketClass CRD

NooBaaBucket CRD represents a class for buckets that defines policies for data placement, namespace, replication policy and more.

Data placement capabilities are built as a multi-layer structure, here are the layers bottom-up:
- Spread Layer - list of backing-stores, aggregates the storage of multiple stores.
- Mirroring Layer - list of spread-layers, async-mirroring to all mirrors, with locality optimization (will allocate on the closest region to the source endpoint), mirroring requires at least two backing-stores.
- Tiering Layer - list of mirroring-layers, push cold data to next tier.

Namespace policy:
A namespace bucket-class will define a policy for namespace buckets.
Namespace policy will require a type, the type's value can be one of the following: single, multi, cache.


Namespace policy of type single will require the following configuration:
  - Resource - a single namespace-store, defines the read and write target of the namespace bucket.

Namespace policy of type multi will require the following configuration:
- Read Resources - list of namespace-stores, defines the read targets of the namespace bucket.
- Write Resource - a single namespace-store, defines the write target of the namespace bucket.

Namespace policy of type cache will require the following configuration:
  - Hub Resource - a single namespace-store, defines the read and write target of the namespace bucket.
  - TTL - defines the TTL of the cached data.


Replication policy:
A bucket-class will define a replication policy when an admin would like to replicate objects 
within noobaa bucket to another noobaa bucket (source/destination buckets can be regular/namespace buckets).

Replication policy will require the content of a JSON file which defines array of rules.
  - Each rule is an object contains rule_id, destination bucket and optional filter object that contains prefix field.
  - When a filter with prefix is defined - only objects keys that match prefix, will be replicated.

Constraints:
A backing-store name may appear in more than one bucket-class but may not appear more than once in a single bucket-class.
The operator cli currently only supports a single tier placement-policy for a bucket-class.
Upon creating regular buckets, the user will first need to create a placement-bucketclass which contains placemant policy.
Upon creating namespace buckets, the user will first need to create a namespace-bucketclass which contains namespace policy.
A namespace bucket class of type cache must contain both Placement-policy or Namespace-policy.
A namespace bucket class of type single/multi must contain Namespace-policy.
YAML must be used to create a bucket-class with a placement-policy that has multiple tiers.
Placement-policy is case sensitive and should be of value (Mirror|Spread).
Namespace-policy is case sensitive.

For more information on using bucket-classes from S3 see [S3 Account](s3-account.md).

# Definitions

- CRD: [noobaa.io_bucketclasses_crd.yaml](../deploy/crds/noobaa.io_bucketclasses_crd.yaml)
- CR: [noobaa.io_v1alpha1_bucketclass_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_bucketclass_cr.yaml)


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
noobaa -n noobaa bucketclass create placement-bucketclass bc --backingstores bs --placement Spread
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
noobaa -n noobaa bucketclass create placement-bucketclass bc --backingstores bs1,bs2 --placement Spread
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
noobaa -n noobaa bucketclass create placement-bucketclass bc --backingstores bs1,bs2 --placement Mirror
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

Namespace bucketclass:
```shell
noobaa -n noobaa bucketclass create namespace-bucketclass single bc --resource azure-blob-ns
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
  namespacePolicy:
    type: Single
    single: 
      resource: azure-blob-ns
```

Namespace bucketclass:
```shell
noobaa -n noobaa bucketclass create namespace-bucketclass multi bc --write-resource aws-s3-ns --read-resources aws-s3-ns,azure-blob-ns 
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
  namespacePolicy:
    type: Multi
    multi:
      writeResource: aws-s3-ns 
      readResources:
      - aws-s3-ns
      - azure-blob-ns
```

Namespace bucketclass:
```shell
noobaa -n noobaa bucketclass create namespace-bucketclass cache bc --hub-resource ibm-cos-ns --ttl 36000 --backingstores noobaa-default-backing-store
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
  namespacePolicy:
    type: Cache
    cache:
      caching:
        ttl: 36000
      hubResource: ibm-cos-ns 
  placementPolicy:
    tiers:
    - backingStores:
      - noobaa-default-backing-store
```


Namespace bucketclass with replication to first.bucket:

/path/to/json-file.json is the path to a JSON file which defines the replication policy
```shell
noobaa -n noobaa bucketclass create namespace-bucketclass single bc --resource azure-blob-ns --replication-policy=/path/to/json-file.json
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
  namespacePolicy:
    type: Single
    single: 
      resource: azure-blob-ns
  replicationPolicy: [{ "rule_id": "rule-1", "destination_bucket": "first.bucket", "filter": {"prefix": "ba"}}]
```

BucketClass in namespace other than NooBaa System namespace, here `<TARGET-NOOBAA-SYSTEM-NAMESPACE>` is the namespace where NooBaa system is deployed:

```shell
TODO
```

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    noobaa-operator: <TARGET-NOOBAA-SYSTEM-NAMESPACE>
    app: noobaa
  name: bc
  namespace: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - noobaa-test-backing-store
```
