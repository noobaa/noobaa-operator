[NooBaa Operator](../README.md) /
# Vector Buckets

Vector buckets provide native vector-database storage backed by [NSFS NamespaceStores](namespace-store-crd.md). They are provisioned through the standard OBC (ObjectBucketClaim) flow and are managed with the [AWS S3 Vectors CLI](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-vectors.html).

Currently supported vector DB type: **Lance**.

# Prerequisites

- A running NooBaa system
- An NSFS NamespaceStore (the backing storage for vector data)
- AWS CLI v2.34.23+ (for `s3vectors` subcommand support)

# Setup

## 1. Create NSFS Storage (PV and PVC)

Create a PersistentVolume and PersistentVolumeClaim to provide local filesystem storage for NSFS:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nsfs-vol
spec:
  storageClassName: nsfs-local
  volumeMode: Filesystem
  persistentVolumeReclaimPolicy: Retain
  local:
    path: /nsfs/
  capacity:
    storage: 1Ti
  accessModes:
    - ReadWriteMany
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/os
              operator: Exists
```

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nsfs-vol
spec:
  storageClassName: nsfs-local
  resources:
    requests:
      storage: 1Ti
  accessModes:
    - ReadWriteMany
```

```shell
kubectl apply -f nsfs-local-pv.yaml
kubectl apply -f nsfs-local-pvc.yaml
```

If the OBC specifies a `path` in `additionalConfig`, make sure that directory exists on the node's filesystem under the PV mount path (e.g. `/nsfs/my-data-path/`).

## 2. Create an NSFS NamespaceStore

Create the NamespaceStore backed by the PVC:

```shell
noobaa namespacestore create nsfs nss1 --pvc-name nsfs-vol
```

## 3. Create a Vector BucketClass

A vector bucket class requires a `vectorPolicy` specifying the NSFS NamespaceStore resource and the vector DB type:

```shell
noobaa bucketclass create vector-bucketclass my-vector-bc --resource nss1 --vector-db-type=lance
```

```yaml
apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  labels:
    app: noobaa
  name: my-vector-bc
  namespace: app-namespace
spec:
  vectorPolicy:
    resource: nss1
    vectorDBType: lance
```

### Constraints

- `vectorPolicy` cannot be combined with `placementPolicy` or `namespacePolicy`.
- Updates to a vector bucket class are not supported at this time.

## 4. Create a Vector OBC

Create an ObjectBucketClaim with `bucketType: vector` in `additionalConfig`, referencing the vector bucket class:

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-vector-obc
  namespace: app-namespace
spec:
  bucketName: my-vector-bucket
  storageClassName: app-namespace.noobaa.io
  additionalConfig:
    bucketclass: my-vector-bc
    bucketType: vector
    path: "my-data-path"
```

### `additionalConfig` Fields

| Field         | Required | Description                                                                                     |
|---------------|----------|-------------------------------------------------------------------------------------------------|
| `bucketclass` | Yes      | Name of the vector BucketClass to use.                                                          |
| `bucketType`  | Yes      | Must be `vector`.                                                                               |
| `path`        | No       | Subdirectory path inside the NSFS PVC where vector data is stored. Must exist on the filesystem. |

Apply the OBC:

```shell
kubectl apply -f my-vector-obc.yaml
```

Verify the OBC is bound:

```shell
kubectl get obc
```

```
NAME             STORAGE-CLASS          PHASE   AGE
my-vector-obc    app-namespace.noobaa.io   Bound   1m
```

## 5. Retrieve Credentials

Once the OBC is bound, the operator creates a Secret and ConfigMap with the same name as the OBC. Retrieve the access keys:

```shell
noobaa obc status my-vector-obc
```

Set up a shell alias for convenience:

```shell
alias s3vectors='AWS_ACCESS_KEY_ID=<access_key> AWS_SECRET_ACCESS_KEY=<secret_key> aws s3vectors --no-verify-ssl --endpoint-url https://localhost:14443'
```

<!-- If connecting from inside the cluster, use the service endpoint `https://vectors.<noobaa-namespace>.svc.cluster.local` instead. -->

# Usage

## Port-Forward (local development)

To access the S3 Vectors endpoint locally, port-forward the NooBaa endpoint pod:

```shell
kubectl port-forward noobaa-endpoint-<suffix> 14443:14443
```

## Create an Index

```shell
s3vectors create-index \
    --vector-bucket-name my-vector-bucket \
    --index-name my-index \
    --data-type float32 \
    --dimension 1536 \
    --distance-metric cosine
```

## Put Vectors

```shell
s3vectors put-vectors \
    --vector-bucket-name my-vector-bucket \
    --index-name my-index \
    --vectors '[
        {
            "key": "vector-1",
            "data": {
                "float32": [0.12, 0.45, 0.78, 0.90]
            },
            "metadata": {
                "category": "research",
                "tag": "sample"
            }
        }
    ]'
```

## Verify Data on Disk

The vector data is stored in the NSFS PVC under the configured path:

```shell
kubectl exec -it noobaa-endpoint-<suffix> -- ls -la /nsfs/nss1/my-data-path/my-vector-bucket_my-index.lance/
```

```
total 0
drwxrwxrwx 6 root root 120 Apr  6 08:59 .
drwxr-xr-x 3 root root  60 Apr  6 08:59 ..
drwxrwxrwx 5 root root 100 Apr  6 08:59 _indices
drwxrwxrwx 2 root root 120 Apr  6 08:59 _transactions
drwxrwxrwx 2 root root 120 Apr  6 08:59 _versions
drwxrwxrwx 2 root root  60 Apr  6 08:59 data
```

## Verify in the Database

You can inspect the NooBaa PostgreSQL database to confirm that the vector bucket and its associated account were created.

Connect to the database:

```shell
kubectl exec -it noobaa-db-pg-cluster-1 -c postgres -- psql nbcore
```

Query the `vector_buckets` table:

```sql
SELECT data->>'name' AS name,
       data->>'vector_db_type' AS vector_db_type,
       data->'namespace_resource'->>'path' AS path,
       data->'bucket_claim'->>'bucket_class' AS bucket_class,
       data->>'creation_time' AS creation_time
FROM vector_buckets
WHERE data->>'deleted' IS NULL;
```

```
      name       | vector_db_type |    path     | bucket_class |  creation_time
-----------------+----------------+-------------+--------------+----------------
 my-vector-bucket | lance          | my-data-path | my-vector-bc | 1775465823580
(1 row)
```

Query the OBC account in the `accounts` table:

```sql
SELECT data->>'name' AS name,
       data->>'bucket_claim_owner' AS bucket_claim_owner
FROM accounts
WHERE data->>'name' LIKE 'obc-account.my-vector-bucket%'
  AND data->>'deleted' IS NULL;
```

```
                    name                     |      bucket_claim_owner
---------------------------------------------+--------------------------
 obc-account.my-vector-bucket.69d3755f@noobaa.io | 69d3755ff4811c00224a30c7
(1 row)
```

# Cleanup

Delete the OBC to remove the vector bucket and its associated account:

```shell
kubectl delete obc my-vector-obc -n app-namespace
```

The vector bucket and its OBC account are soft-deleted in the NooBaa database (marked with a `deleted` timestamp).

You can verify the deletion by querying the database:

```sql
SELECT data->>'name' AS name,
       data->>'deleted' AS deleted
FROM vector_buckets;
```

```
      name        |          deleted
------------------+----------------------------
 my-vector-bucket | 2026-04-06T09:35:35.815Z
(1 row)
```

```sql
SELECT data->>'name' AS name,
       data->>'deleted' AS deleted
FROM accounts
WHERE data->>'name' LIKE 'obc-account.my-vector-bucket%';
```

```
                    name                     |          deleted
---------------------------------------------+----------------------------
 obc-account.my-vector-bucket.69d3755f@noobaa.io | 2026-04-06T09:35:35.826Z
(1 row)
```

# Limitations

- Only NSFS NamespaceStores are supported as the backing resource.
- Only `lance` is supported as the `vectorDBType`.
- Bucket tagging and quota are not yet supported for vector buckets.
- Vector bucket class updates (spec changes) are denied by the admission webhook.
