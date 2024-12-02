[NooBaa Operator](../README.md) /
# BucketClass CRD
The BucketClass CRD represents a structure that defines bucket policies relating to data placement, [namespace](namespace-store-crd.md) properties, replication policies and more.

Note that placement-bucketclass and namespace-bucketclass both use the same CR, and the difference lies inside the bucket class' `spec` section, more specifically the presence of either the `placementPolicy` or `namespacePolicy` key. 

# Definitions
- CRD: [noobaa.io_bucketclasses.yaml](../deploy/crds/noobaa.io_bucketclasses.yaml)
- CR: [noobaa.io_v1alpha1_bucketclass_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_bucketclass_cr.yaml)

## Placement Policy
A placement bucket class defines a policy for standard buckets - i.e. NooBaa buckets that are backed by [backingstores](backing-store-crd.md).
The data placement capabilities are built as a multi-layer structure, here are the layers bottom-up:
- Spread Layer - list of backing-stores, aggregates the storage of multiple stores.
- Mirroring Layer - list of spread-layers, async-mirroring to all mirrors, with locality optimization (will allocate on the closest region to the source endpoint), mirroring requires at least two backing-stores.
- Tiering Layer - list of mirroring-layers, push cold data to next tier.

## Namespace Policy
A namespace bucket class defines a policy for namespace buckets - i.e. NooBaa buckets that are backed by [namespacestores](namespace-store-crd.md).
There are several types of namespace policies:
- Single - a single namespace store is used for both read and write operations on the target bucket
- Multi - a single namespace store is used for write operations, and a list of namespace stores can be used for read operations
- Cache - functions similarly to `Single`, except with an additional `TTL` key, which dictates the time-to-live of the cached data

### Time-to-live (TTL)
Cache bucketclasses work by saving read objects in a chosen backingstore, which leads to faster access times in the future. In order to make sure that the cached object is not out of sync with the one in the remote target, an ETag comparison might be run upon read, depending on the TTL that the user chooses. The TTL can fall in one of three categories:
- Negative (e.g. `-1`) - when the user knows there are no out of band writes, they can use a negative TTL, which means no revalidations are done; if the object is in the cache - it is returned without an ETag comparison. This is the most performant option.
- Zero (`0`) - the cache will always compare the object's ETag before returning it. This option has a performance cost of getting the ETag from the remote target on each object read. This is the least performant option.
- Positive (denoted in milliseconds, e.g. `3600000` equals to an hour) - once an object was read and saved in the cache, the chosen amount of time will have to pass prior to the object's ETag being compared again.

## Constraints:
- A backing store name may appear in more than one bucket class but may not appear more than once in a certain bucket class.
- The operator CLI currently only supports a single tier placement policy for a bucket class.
- Thus, YAML must be used to create a bucket class with a placement policy that has multiple tiers.
- Upon creating standard buckets, the user will first need to create a placement bucketclass which contains a placemant policy.
- Upon creating namespace buckets, the user will first need to create a namespace bucketclass which contains a namespace policy.
- A namespace bucket class of type cache must contain both a placement and a namespace policy.
- A namespace bucket class of type single/multi must contain a namespace policy.
- Placement policy is case sensitive and should be of value `Mirror` or `Spread` when more than one backingstore is provided.
- Namespace policy is case sensitive and should be of values `Single`, `Multi` or `Cache`.


# Reconciliation
- The operator will verify that the bucket class is valid - i.e. that the backingstores and namespacestores exist and can be accessed and used.
- Changes to a bucket class spec will be propagated to buckets that were instantiated from it.
- Other than that the bucket class is passive, just waiting there for new buckets to use it.

# Resource Status
It is possible to check a resource's status in several ways, including:
- `kubectl get bucketclass -A <NAME> -o yaml` (will retrieve bucketclasses from all cluster namespaces)
- `kubectl describe bucketclass <NAME>`
- `noobaa bucketclass status <NAME>`

Below is an example of a healthy bucket class' status, as retrieved with the first command:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: noobaa-default-class
  namespace: app-namespace
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


# Examples
Please note that CLI (`noobaa`) examples need NooBaa to run under `app-namespace`, despite the fact bucketclasses are supported in all namespaces

Single tier, single backing store, Spread placement:
```shell
noobaa -n app-namespace bucketclass create placement-bucketclass bc --backingstores bs --placement Spread
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs
      placement: Spread
```

Single tier, two backing stores, Spread placement:
```shell
noobaa -n app-namespace bucketclass create placement-bucketclass bc --backingstores bs1,bs2 --placement Spread
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      - bs2
      placement: Spread
```

Single tier, two backing stores, Mirror placement:
```shell
noobaa -n app-namespace bucketclass create placement-bucketclass bc --backingstores bs1,bs2 --placement Mirror
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - bs1
      - bs2
      placement: Mirror
```

Two tiers (only achievable by applying a YAML at the moment) - single backing stores per tier, Spread placement in tiers:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
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

Two tiers (only achievable by applying a YAML at the moment) - two backing stores per tier, Spread placement in first tier and Mirror in second tier:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
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

Namespace bucketclass, a single read and write resource in Azure:
```shell
noobaa -n app-namespace bucketclass create namespace-bucketclass single bc --resource azure-blob-ns
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  namespacePolicy:
    type: Single
    single: 
      resource: azure-blob-ns
```

Namespace bucketclass, a single write resource in AWS, multiple read resources in AWS and Azure:
```shell
noobaa -n app-namespace bucketclass create namespace-bucketclass multi bc --write-resource aws-s3-ns --read-resources aws-s3-ns,azure-blob-ns 
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  namespacePolicy:
    type: Multi
    multi:
      writeResource: aws-s3-ns 
      readResources:
      - aws-s3-ns
      - azure-blob-ns
```

Namespace bucketclass, cache stored in `noobaa-default-backing-store`, objects are read from and written to IBM COS:
```shell
noobaa -n app-namespace bucketclass create namespace-bucketclass cache bc --hub-resource ibm-cos-ns --ttl 36000 --backingstores noobaa-default-backing-store
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
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
noobaa -n app-namespace bucketclass create namespace-bucketclass single bc --resource azure-blob-ns --replication-policy=/path/to/json-file.json
```
```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: bc
  namespace: app-namespace
spec:
  namespacePolicy:
    type: Single
    single: 
      resource: azure-blob-ns
  replicationPolicy: [{ "rule_id": "rule-1", "destination_bucket": "first.bucket", "filter": {"prefix": "ba"}}]
```

Bucket class in a namespace other than the NooBaa system namespace. `<TARGET-NOOBAA-SYSTEM-NAMESPACE>` is the namespace where the NooBaa system is deployed:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    noobaa-operator: <TARGET-NOOBAA-SYSTEM-NAMESPACE>
    app: noobaa
  name: bc
  namespace: app-namespace
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - noobaa-test-backing-store
```
