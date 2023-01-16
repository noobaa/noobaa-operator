[NooBaa Operator](../README.md) /
# OBC Provisioner

Kubernetes natively supports dynamic provisioning for many types of file and block storage, but lacks support for object bucket provisioning.
In order to provide a native provisioning of object storage buckets, the concept of Object Bucket Claim (OBC/OB) was introduced in a similar manner to Persistent Volume Claim (PVC/PV)

The `lib-bucket-provisioner` repo provides a library implementation and design to unify the implementations:

[OBC Design Document](https://github.com/kube-object-storage/lib-bucket-provisioner/blob/master/doc/design/object-bucket-lib.md)


# StorageClass

The operator creates a default storage class with the name pattern `<noobaa-namespace>.noobaa.io` which helps to identify the target noobaa deployment since storage classes are non-namespaced. 

However the administrator of a noobaa deployment can create additional StorageClasses that refer to a different BucketClass for different data placement policies or reclaim policy, and control its visibility to app-owners using RBAC rules.

For more information see https://github.com/kube-object-storage/lib-bucket-provisioner/blob/master/deploy/storageClass.yaml 

Example:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: noobaa.noobaa.io
provisioner: noobaa.noobaa.io/obc
reclaimPolicy: Delete
parameters:
  bucketclass: noobaa-default-bucket-class
```

# OBC

Applications that require a bucket will create an OBC and refer to a storage class name.

See https://github.com/kube-object-storage/lib-bucket-provisioner/blob/master/deploy/example-claim.yaml 

The operator will watch for OBC's and fulfill the claims by create/find existing bucket in NooBaa, and will share a config map and a secret with the application in order to give it all the needed details to work with the bucket.

Example:

```bash
noobaa obc create my-bucket-claim -n noobaa --app-namespace my-app
```

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-bucket-claim
  namespace: my-app
spec:
  generateBucketName: my-bucket
  storageClassName: noobaa.noobaa.io
```

# OBC with specific BucketClass

Applications that require a bucket from a specific BucketClass can create an OBC with the default StorageClass, but override the BucketClass specifically for that claim using the `spec.additionalConfig.bucketclass` property. The BucketClass can exist either in the same namespace as that of OBC or it can exist in NooBaa system namespace. In case of a conflict, BucketClass that exists in the OBC namespace will take priority over the BucketClass in NooBaa system.

Example:

```bash
noobaa obc create my-bucket-claim -n noobaa --app-namespace my-app --bucketclass custom-bucket-class
```

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-bucket-claim
  namespace: my-app
spec:
  generateBucketName: my-bucket
  storageClassName: noobaa.noobaa.io
  additionalConfig:
    bucketclass: custom-bucket-class
```

# OBC with specific Replication Policy

Applications that require a bucket to have a specific replication policy can create an OBC and add to the claim 
the `spec.additionalConfig.replication-policy` property.

/path/to/json-file.json is the path to a JSON file which defines the replication policy

Example:

```bash
noobaa obc create my-bucket-claim -n noobaa --app-namespace my-app --replication-policy /path/to/json-file.json
```

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-bucket-claim
  namespace: my-app
spec:
  generateBucketName: my-bucket
  storageClassName: noobaa.noobaa.io
  additionalConfig:
    replication-policy: [{ "rule_id": "rule-2", "destination_bucket": "first.bucket", "filter": {"prefix": "bc"}}]
```

# Using the OBC

Once the OBC is provisioned by the operator, a bucket will be created in NooBaa, and the operator will create a Secret and ConfigMap with the same name of the OBC on the same namespace of the OBC. For the example above, the Secret and ConfigMap will both be named `my-bucket-claim`.

The content of the Secret and ConfigMap provides all the information needed by the application in order to connect to the bucket using S3 API, and these can be mounted into the application pods using env or volumes. The application should configure its S3 SDK to use the AWS_ACCESS_KEY_ID & AWS_SECRET_ACCESS_KEY credentials as provided by the Secret, and the BUCKET_HOST:BUCKET_PORT endpoint and BUCKET_NAME as provided by the ConfigMap - see below.

**NOTE on SSL certificates:**
In the last release v2.0.9 the BUCKET_HOST:BUCKET_PORT refers to the s3 service ClusterIP and the https port. However, the SSL certificate of the service is self signed by the cluster CA, and whenever a client application running in the cluster wants to use it and trust the cluster CA it should use the service name instead, which is `s3.<noobaa-namespace>.svc.cluster.local` with the default 443 https port. This was fixed in https://github.com/noobaa/noobaa-operator/pull/189 and will be part of v2.1.0 once it gets released.

Here is an example content:

Secret:
```yaml
apiVersion: v1
data:
  AWS_ACCESS_KEY_ID: UU1odkFSWjNiMWplTmx2UnZpcHU=
  AWS_SECRET_ACCESS_KEY: eTdFdlkva1FSQVEzWElWWXB3L2ZYdVBZdG8wSDBxaVlBSUpraW1iTA==
kind: Secret
metadata:
  creationTimestamp: "2020-01-14T08:05:31Z"
  finalizers:
  - objectbucket.io/finalizer
  labels:
    app: noobaa
    bucket-provisioner: noobaa.noobaa.io-obc
    noobaa-domain: noobaa.noobaa.io
  name: my-bucket-claim
  namespace: my-app
  ownerReferences:
  - apiVersion: objectbucket.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ObjectBucketClaim
    name: my-bucket-claim
    uid: a2852724-36a4-11ea-bb22-0a49da05a082
  resourceVersion: "24953844"
  selfLink: /api/v1/namespaces/my-app/secrets/my-bucket-claim
  uid: a2a6ee13-36a4-11ea-bb22-0a49da05a082
type: Opaque
```

ConfigMap:
```yaml
apiVersion: v1
data:
  BUCKET_HOST: 10.0.169.118
  BUCKET_NAME: my-bucket-9febb742-f14d-4a94-8ed8-3da5d8bd242c
  BUCKET_PORT: "32603"
  BUCKET_REGION: ""
  BUCKET_SUBREGION: ""
kind: ConfigMap
metadata:
  creationTimestamp: "2020-01-14T08:05:32Z"
  finalizers:
  - objectbucket.io/finalizer
  labels:
    app: noobaa
    bucket-provisioner: noobaa.noobaa.io-obc
    noobaa-domain: noobaa.noobaa.io
  name: my-bucket-claim
  namespace: my-app
  ownerReferences:
  - apiVersion: objectbucket.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: ObjectBucketClaim
    name: my-bucket-claim
    uid: a2852724-36a4-11ea-bb22-0a49da05a082
  resourceVersion: "24953845"
  selfLink: /api/v1/namespaces/my-app/configmaps/my-bucket-claim
  uid: a2b28adc-36a4-11ea-bb22-0a49da05a082
```

# Embedding OBC in the Application

An application deployment can claim a bucket and refer to the expected Secret & ConfigMap in a static deployment yaml, since the names of the Secret and ConfigMap follow the same name as the OBC, and pods that mount information from Secret or ConfigMaps will not start until those resources exist, this provides a self contained deployment that will only run the application once the bucket provisioning is complete.

Here is an example yaml that combines the OBC and a Pod that uses it:

```yaml
apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-bucket-claim
spec:
  generateBucketName: my-bucket
  storageClassName: noobaa.noobaa.io
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    app: my-app
spec:
  containers:
  - name: app
    envFrom:
    - secretRef:
        name: my-bucket-claim
    - configMapRef:
        name: my-bucket-claim
    image: banst/awscli
    command:
    - sh
    - "-c"
    - |
      echo "----> Configuring S3 endpoint ...";
      pip install awscli-plugin-endpoint;
      aws configure set plugins.endpoint awscli_plugin_endpoint;
      aws configure set s3.endpoint_url https://s3.noobaa.svc.cluster.local;
      echo "----> Configuring certificates ...";
      aws configure set ca_bundle /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt;
      echo "----> Copying files ...";
      aws s3 cp --recursive /etc s3://$BUCKET_NAME;
      echo "----> List files ...";
      aws s3 ls $BUCKET_NAME;
      echo "----> Done.";
```


# Bucket Permissions and Sharing

The scope of bucket permissions is at the claim scope - this means that the credentials of the OBC are confined to access only that single OBC bucket. Notice that also listing buckets with these S3 credentials will return only that one bucket.

Going forward we would like to have an option to create a single account per application namespace so that all buckets claimed by an application will be visible and accessible to that application.
