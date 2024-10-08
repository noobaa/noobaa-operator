[NooBaa Operator](../README.md) /

# Bucket Replication
Bucket replication is a NooBaa feature that allows a user to set a replication policy for all or some objects. The goal of replication policies is simple - to define a target for objects to be copied to.

When applied to a bucketclass, the policy will apply to all future bucket claims that'll utilize the bucketclass.

In general, a replication policy is a JSON-compliant string which defines an array containing at least one rule  -
  - Each rule is an object containing a `rule_id`, a `destination bucket`, and an optional `filter` key that contains a `prefix` field.
  - When a filter with prefix is provided - only objects keys that match the prefix will be replicated

The main mechanism behind bucket replication lists all objects across the two buckets, compares the differences, and copies the missing objects from the source to the target bucket. Log-based optimizations utilize AWS S3 server access logging or Azure Monitor to optimize the replication process by copying objects that have been created or modified since the feature was turned on - effectively allowing users to get up to speed with up-to-date recent objects, as the rest replicate in the background with the classic method.

## Replication Policy Parameters
As stated above, a replication policy is a JSON-compliant array of rules (examples are provided at the bottom of this section)
  - Each rule is an object that contains the following keys:
    - `rule_id` - which identifies the rule
    - `destination_bucket` - which dictates the target NooBaa buckets that the objects will be copied to
    - (optional) `{"filter": {"prefix": <>}}` - if the user wishes to filter the objects that are replicated, the value of this field can be set to a prefix string
    - (optional, log-based optimization, see below) `sync_deletions` - can be set to a boolean value to indicate whether deletions should be replicated
    - (optional, log-based optimization, see below) `sync_versions` - can be set to a boolean value to indicate whether object versions should be replicated

In addition, when the bucketclass is backed by namespacestores, each policy can be set to optimize replication by utilizing logs (configured and supplied by the user, currently only supports AWS S3 and Azure Blob):
  - <sup><sub>(optional, only supported on namespace buckets)</sub></sup> `log_replication_info` - an object that contains data related to log-based replication optimization -
    - <sup><sub>(necessary on Azure)</sub></sup> `endpoint_type` - this field can be set to an appropriate endpoint type (currently, only AZURE is supported)
    - <sup><sub>(necessary on AWS)</sub></sup> `{"logs_location": {"logs_bucket": <>}}` - this field should be set to the location of the AWS S3 server access logs

## Examples
Note that the example poicies below can also be saved as files and passed to the NooBaa CLI. In that case, it's necessary to omit the outer single quotes.
### AWS replication policy with log optimization:

`'{"rules":[{"rule_id":"aws-rule-1", "destination_bucket":"first.bucket", "filter": {"prefix": "a."}}], "log_replication_info": {"logs_location": {"logs_bucket": "logsarehere"}}}'`

### Azure replication policy with log optimization:

`'{"rules":[{"rule_id":"azure-rule-1", "sync_deletions": true, "sync_versions": false, "destination_bucket":"first.bucket"}], "log_replication_info": {"endpoint_type": "AZURE"}}'`

### Namespace bucketclass with replication to first.bucket:

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
  replicationPolicy: '[{ "rule_id": "rule-1", "destination_bucket": "first.bucket", "filter": {"prefix": "ba"}}]'
```

### OBC with specific Replication Policy

Applications that require a bucket to have a specific replication policy can create an OBC and add to the claim 
the `spec.additionalConfig.replication-policy` property.

/path/to/json-file.json is the path to a JSON file which defines the replication policy

Example:

```bash
noobaa obc create my-bucket-claim -n appnamespace --replication-policy /path/to/json-file.json
```

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-bucket-claim
  namespace: appnamespace
spec:
  generateBucketName: my-bucket
  storageClassName: noobaa.noobaa.io
  additionalConfig:
    replicationPolicy: '[{ "rule_id": "rule-2", "destination_bucket": "first.bucket", "filter": {"prefix": "bc"}}]'
```
