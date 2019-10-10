package bundle

const Version = "2.0.2-rc1"

const Sha256_deploy_cluster_role_yaml = "f719ff8e0015a73d4e6ff322d2b30efa1cc89fcb3f856c06a5910785cb9e8dd8"

const File_deploy_cluster_role_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: noobaa.noobaa.io
rules:
- apiGroups:
  - noobaa.io
  resources:
  - '*'
  - noobaas
  - backingstores
  - bucketclasses
  - noobaas/finalizers
  - backingstores/finalizers
  - bucketclasses/finalizers
  verbs:
  - '*'
- apiGroups:
  - objectbucket.io
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch
  - create
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
`

const Sha256_deploy_cluster_role_binding_yaml = "15c78355aefdceaf577bd96b4ae949ae424a3febdc8853be0917cf89a63941fc"

const File_deploy_cluster_role_binding_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: noobaa.noobaa.io
subjects:
  - kind: ServiceAccount
    name: noobaa
    namespace: noobaa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: noobaa.noobaa.io
`

const Sha256_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml = "776deb11e769c67bf993ec77762b5375e36607bd4026e4ef20642fde2bd5dc80"

const File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  name: aws1
  labels:
    app: noobaa
  finalizers:
    - noobaa.io/finalizer
spec:
  type: aws-s3
  awsS3:
    targetBucket: noobaa-aws1
    secret:
      name: backing-store-secret-aws1
`

const Sha256_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml = "57e474668ba94f1f0853a04d13f02fff9135caf1428d4963259e8590d8c35c0e"

const File_deploy_crds_noobaa_v1alpha1_backingstore_crd_yaml = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: backingstores.noobaa.io
spec:
  additionalPrinterColumns:
  - JSONPath: .spec.type
    description: Type
    name: Type
    type: string
  - JSONPath: .status.phase
    description: Phase
    name: Phase
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: noobaa.io
  names:
    kind: BackingStore
    listKind: BackingStoreList
    plural: backingstores
    singular: backingstore
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object metadata.
          type: object
        spec:
          description: Specification of the desired behavior of the noobaa BackingStore.
          properties:
            awsS3:
              description: AWSS3Spec specifies a backing store of type aws-s3
              properties:
                region:
                  description: Region is the AWS region
                  type: string
                secret:
                  description: Secret refers to a secret that provides the credentials
                    The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                  type: object
                sslDisabled:
                  description: SSLDisabled allows to disable SSL and use plain http
                  type: boolean
                targetBucket:
                  description: TargetBucket is the name of the target S3 bucket
                  type: string
              required:
              - targetBucket
              - secret
              type: object
            azureBlob:
              description: AzureBlob specifies a backing store of type azure-blob
              properties:
                secret:
                  description: Secret refers to a secret that provides the credentials
                    The secret should define AccountName and AccountKey as provided
                    by Azure Blob.
                  type: object
                targetBlobContainer:
                  description: TargetBlobContainer is the name of the target Azure
                    Blob container
                  type: string
              required:
              - targetBlobContainer
              - secret
              type: object
            googleCloudStorage:
              description: GoogleCloudStorage specifies a backing store of type google-cloud-storage
              properties:
                secret:
                  description: Secret refers to a secret that provides the credentials
                    The secret should define GoogleServiceAccountPrivateKeyJson containing
                    the entire json string as provided by Google.
                  type: object
                targetBucket:
                  description: TargetBucket is the name of the target S3 bucket
                  type: string
              required:
              - targetBucket
              - secret
              type: object
            pvPool:
              description: PVPool specifies a backing store of type pv-pool
              properties:
                numVolumes:
                  description: NumVolumes is the number of volumes to allocate
                  format: int64
                  type: integer
                resources:
                  description: VolumeResources represents the minimum resources each
                    volume should have.
                  type: object
                storageClass:
                  description: StorageClass is the name of the storage class to use
                    for the PV's
                  type: string
              required:
              - numVolumes
              type: object
            s3Compatible:
              description: S3Compatible specifies a backing store of type s3-compatible
              properties:
                endpoint:
                  description: 'Endpoint is the S3 compatible endpoint: http(s)://host:port'
                  type: string
                secret:
                  description: Secret refers to a secret that provides the credentials
                    The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                  type: object
                signatureVersion:
                  description: SignatureVersion specifies the client signature version
                    to use when signing requests.
                  type: string
                targetBucket:
                  description: TargetBucket is the name of the target S3 bucket
                  type: string
              required:
              - targetBucket
              - secret
              - endpoint
              type: object
            type:
              description: Type is an enum of supported types
              type: string
          required:
          - type
          type: object
        status:
          description: Most recently observed status of the noobaa BackingStore.
          properties:
            conditions:
              description: Conditions is a list of conditions related to operator
                reconciliation +patchMergeKey=type +patchStrategy=merge
              items:
                properties:
                  lastHeartbeatTime:
                    format: date-time
                    type: string
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                required:
                - type
                - status
                - lastHeartbeatTime
                - lastTransitionTime
                type: object
              type: array
            phase:
              description: Phase is a simple, high-level summary of where the backing
                store is in its lifecycle
              type: string
            relatedObjects:
              description: RelatedObjects is a list of objects related to this operator.
              items:
                type: object
              type: array
          required:
          - phase
          type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`

const Sha256_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml = "af1411669ca0b29bdb7836e9e1fc44a0ddb7d4a994266abbae793a7116f6499f"

const File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: default
  labels:
    app: noobaa
spec:
  placementPolicy:
    tiers:
    - backingStores:
      - aws1
`

const Sha256_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml = "2362aac8610c170155687a2d11ba08bca30c7040a7e16c1228ccd94e6c3b7bec"

const File_deploy_crds_noobaa_v1alpha1_bucketclass_crd_yaml = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bucketclasses.noobaa.io
spec:
  additionalPrinterColumns:
  - JSONPath: .spec.placementPolicy
    description: Placement
    name: Placement
    type: string
  - JSONPath: .status.phase
    description: Phase
    name: Phase
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: noobaa.io
  names:
    kind: BucketClass
    listKind: BucketClassList
    plural: bucketclasses
    singular: bucketclass
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object metadata.
          type: object
        spec:
          description: Specification of the desired behavior of the noobaa BucketClass.
          properties:
            placementPolicy:
              description: PlacementPolicy specifies the placement policy for the
                bucket class
              properties:
                tiers:
                  description: Tiers is an ordered list of tiers to use. The model
                    is a waterfall - push to first tier by default, and when no more
                    space spill "cold" storage to next tier.
                  items:
                    properties:
                      backingStores:
                        description: BackingStores is an unordered list of backing
                          store names. The meaning of the list depends on the placement.
                        items:
                          type: string
                        type: array
                      placement:
                        description: Placement specifies the type of placement for
                          the tier If empty it should have a single backing store.
                        type: string
                    type: object
                  type: array
              required:
              - tiers
              type: object
          required:
          - placementPolicy
          type: object
        status:
          description: Most recently observed status of the noobaa BackingStore.
          properties:
            conditions:
              description: Conditions is a list of conditions related to operator
                reconciliation +patchMergeKey=type +patchStrategy=merge
              items:
                properties:
                  lastHeartbeatTime:
                    format: date-time
                    type: string
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                required:
                - type
                - status
                - lastHeartbeatTime
                - lastTransitionTime
                type: object
              type: array
            phase:
              description: Phase is a simple, high-level summary of where the System
                is in its lifecycle
              type: string
            relatedObjects:
              description: RelatedObjects is a list of objects related to this operator.
              items:
                type: object
              type: array
          required:
          - phase
          type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`

const Sha256_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml = "0e1f573b02ad8f9f1e4e8bed5fe47e4651040277e0accf1e30b1369745371483"

const File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  labels:
    app: noobaa
spec: {}
`

const Sha256_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml = "c81b0849d9ab3e16d61b34d85752907f7f3e3cebb322f28a67eaab0cc6a9d701"

const File_deploy_crds_noobaa_v1alpha1_noobaa_crd_yaml = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: noobaas.noobaa.io
spec:
  additionalPrinterColumns:
  - JSONPath: .status.services.serviceMgmt.nodePorts
    description: Management Endpoints
    name: Mgmt-Endpoints
    type: string
  - JSONPath: .status.services.serviceS3.nodePorts
    description: S3 Endpoints
    name: S3-Endpoints
    type: string
  - JSONPath: .status.actualImage
    description: Actual Image
    name: Image
    type: string
  - JSONPath: .status.phase
    description: Phase
    name: Phase
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: noobaa.io
  names:
    kind: NooBaa
    listKind: NooBaaList
    plural: noobaas
    shortNames:
    - nb
    singular: noobaa
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object metadata.
          type: object
        spec:
          description: Specification of the desired behavior of the noobaa system.
          properties:
            coreResources:
              description: CoreResources (optional) overrides the default resource
                requirements for the server container
              type: object
            dbImage:
              description: DBImage (optional) overrides the default image for the
                db container
              type: string
            dbResources:
              description: DBResources (optional) overrides the default resource requirements
                for the db container
              type: object
            dbStorageClass:
              description: DBStorageClass (optional) overrides the default cluster
                StorageClass for the database volume. For the time being this field
                is immutable and can only be set on system creation. This affects
                where the system stores its database which contains system config,
                buckets, objects meta-data and mapping file parts to storage locations.
                +immutable
              type: string
            dbVolumeResources:
              description: 'DBVolumeResources (optional) overrides the default PVC
                resource requirements for the database volume. For the time being
                this field is immutable and can only be set on system creation. This
                is because volume size updates are only supported for increasing the
                size, and only if the storage class specifies ` + "`" + `allowVolumeExpansion:
                true` + "`" + `, +immutable'
              type: object
            image:
              description: Image (optional) overrides the default image for the server
                container
              type: string
            imagePullSecret:
              description: ImagePullSecret (optional) sets a pull secret for the system
                image
              type: object
            pvPoolDefaultStorageClass:
              description: PVPoolDefaultStorageClass (optional) overrides the default
                cluster StorageClass for the pv-pool volumes. This affects where the
                system stores data chunks (encrypted). Updates to this field will
                only affect new pv-pools, but updates to existing pools are not supported
                by the operator.
              type: string
            tolerations:
              description: Tolerations (optional) passed through to noobaa's pods
              items:
                type: object
              type: array
          type: object
        status:
          description: Most recently observed status of the noobaa system.
          properties:
            accounts:
              properties:
                admin:
                  properties:
                    secretRef:
                      type: object
                  required:
                  - secretRef
                  type: object
              required:
              - admin
              type: object
            actualImage:
              description: ActualImage is set to report which image the operator is
                using
              type: string
            conditions:
              description: Conditions is a list of conditions related to operator
                reconciliation +patchMergeKey=type +patchStrategy=merge
              items:
                properties:
                  lastHeartbeatTime:
                    format: date-time
                    type: string
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                required:
                - type
                - status
                - lastHeartbeatTime
                - lastTransitionTime
                type: object
              type: array
            observedGeneration:
              description: ObservedGeneration is the most recent generation observed
                for this noobaa system. It corresponds to the CR generation, which
                is updated on mutation by the API Server.
              format: int64
              type: integer
            phase:
              description: Phase is a simple, high-level summary of where the System
                is in its lifecycle
              type: string
            readme:
              description: Readme is a user readable string with explanations on the
                system
              type: string
            relatedObjects:
              description: RelatedObjects is a list of objects related to this operator.
              items:
                type: object
              type: array
            services:
              properties:
                serviceMgmt:
                  properties:
                    externalDNS:
                      description: ExternalDNS are external public addresses for the
                        service
                      items:
                        type: string
                      type: array
                    externalIP:
                      description: ExternalIP are external public addresses for the
                        service LoadBalancerPorts such as AWS ELB provide public address
                        and load balancing for the service IngressPorts are manually
                        created public addresses for the service https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
                        https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
                        https://kubernetes.io/docs/concepts/services-networking/ingress/
                      items:
                        type: string
                      type: array
                    internalDNS:
                      description: InternalDNS are internal addresses of the service
                        inside the cluster
                      items:
                        type: string
                      type: array
                    internalIP:
                      description: InternalIP are internal addresses of the service
                        inside the cluster https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                      items:
                        type: string
                      type: array
                    nodePorts:
                      description: NodePorts are the most basic network available.
                        NodePorts use the networks available on the hosts of kubernetes
                        nodes. This generally works from within a pod, and from the
                        internal network of the nodes, but may fail from public network.
                        https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                      items:
                        type: string
                      type: array
                    podPorts:
                      description: 'PodPorts are the second most basic network address.
                        Every pod has an IP in the cluster and the pods network is
                        a mesh so the operator running inside a pod in the cluster
                        can use this address. Note: pod IPs are not guaranteed to
                        persist over restarts, so should be rediscovered. Note2: when
                        running the operator outside of the cluster, pod IP is not
                        accessible.'
                      items:
                        type: string
                      type: array
                  type: object
                serviceS3:
                  properties:
                    externalDNS:
                      description: ExternalDNS are external public addresses for the
                        service
                      items:
                        type: string
                      type: array
                    externalIP:
                      description: ExternalIP are external public addresses for the
                        service LoadBalancerPorts such as AWS ELB provide public address
                        and load balancing for the service IngressPorts are manually
                        created public addresses for the service https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
                        https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
                        https://kubernetes.io/docs/concepts/services-networking/ingress/
                      items:
                        type: string
                      type: array
                    internalDNS:
                      description: InternalDNS are internal addresses of the service
                        inside the cluster
                      items:
                        type: string
                      type: array
                    internalIP:
                      description: InternalIP are internal addresses of the service
                        inside the cluster https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                      items:
                        type: string
                      type: array
                    nodePorts:
                      description: NodePorts are the most basic network available.
                        NodePorts use the networks available on the hosts of kubernetes
                        nodes. This generally works from within a pod, and from the
                        internal network of the nodes, but may fail from public network.
                        https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                      items:
                        type: string
                      type: array
                    podPorts:
                      description: 'PodPorts are the second most basic network address.
                        Every pod has an IP in the cluster and the pods network is
                        a mesh so the operator running inside a pod in the cluster
                        can use this address. Note: pod IPs are not guaranteed to
                        persist over restarts, so should be rediscovered. Note2: when
                        running the operator outside of the cluster, pod IP is not
                        accessible.'
                      items:
                        type: string
                      type: array
                  type: object
              required:
              - serviceMgmt
              - serviceS3
              type: object
          required:
          - observedGeneration
          - phase
          - actualImage
          - accounts
          - services
          - readme
          type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`

const Sha256_deploy_internal_ceph_objectstore_user_yaml = "9b60853f585d771e484854e34fc59adf56a6edb6acb1315eb2ea02b59d213755"

const File_deploy_internal_ceph_objectstore_user_yaml = `apiVersion: ceph.rook.io/v1
kind: CephObjectStoreUser
metadata:
  name: CEPH_OBJ_USER_NAME
spec:
  store: STORE_NAME
  displayName: my display name
`

const Sha256_deploy_internal_cloud_creds_aws_cr_yaml = "8e4159bc3470c135b611b6d9f4338612be0e6ea381d5061cc79e84a7eec0ab6a"

const File_deploy_internal_cloud_creds_aws_cr_yaml = `apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: CRED-REQ-NAME
spec:
  secretRef:
    name: CRED-SECRET-NAME
    namespace: CRED-SECRET-NAMESPACE
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    statementEntries:
      - effect: Allow
        action:
          - s3:*
        resource: "arn:aws:s3:::BUCKET"
      - effect: Allow
        action:
          - s3:*
        resource: "arn:aws:s3:::BUCKET/*"
      - effect: Allow
        action:
          - s3:ListAllMyBuckets
        resource: "*"
`

const Sha256_deploy_internal_prometheus_rules_yaml = "31412ea08c2c489c6cccdb28acdc1817f7ed97b9f3672b1abf80ab4f4129c39f"

const File_deploy_internal_prometheus_rules_yaml = `apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    prometheus: k8s
    role: alert-rules
  name: prometheus-noobaa-rules
spec:
  groups:
  - name: noobaa-telemeter.rules
    rules:
    - expr: |
        sum(NooBaa_num_unhealthy_buckets + NooBaa_num_unhealthy_bucket_claims)
      record: job:noobaa_total_unhealthy_buckets:sum
    - expr: |
        sum(NooBaa_num_buckets + NooBaa_num_buckets_claims)
      record: job:noobaa_bucket_count:sum
    - expr: |
        sum(NooBaa_num_objects + NooBaa_num_objects_buckets_claims)
      record: job:noobaa_total_object_count:sum
    - expr: |
        NooBaa_accounts_num
      record: noobaa_accounts_num
    - expr: |
        NooBaa_total_usage
      record: noobaa_total_usage
  - name: bucket-state-alert.rules
    rules:
    - alert: NooBaaBucketErrorState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is in error state for
          more than 6m
        message: A NooBaa Bucket Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_status{bucket_name=~".*"} == 0
      for: 6m
      labels:
        severity: warning
    - alert: NooBaaBucketReachingQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is using {{ printf
          "%0.0f" $value }}% of its quota
        message: A NooBaa Bucket Is In Reaching Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_quota{bucket_name=~".*"} > 80
      labels:
        severity: warning
    - alert: NooBaaBucketExceedingQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is exceeding its quota
          - {{ printf "%0.0f" $value }}% used
        message: A NooBaa Bucket Is In Exceeding Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_quota{bucket_name=~".*"} >= 100
      labels:
        severity: warning
    - alert: NooBaaBucketLowCapacityState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is using {{ printf
          "%0.0f" $value }}% of its capacity
        message: A NooBaa Bucket Is In Low Capacity State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_capacity{bucket_name=~".*"} > 80
      labels:
        severity: warning
    - alert: NooBaaBucketNoCapacityState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is using all of its
          capacity
        message: A NooBaa Bucket Is In No Capacity State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_capacity{bucket_name=~".*"} > 95
      labels:
        severity: warning
  - name: resource-state-alert.rules
    rules:
    - alert: NooBaaResourceErrorState
      annotations:
        description: A NooBaa resource {{ $labels.resource_name }} is in error state
          for more than 6m
        message: A NooBaa Resource Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_resource_status{resource_name=~".*"} == 0
      for: 6m
      labels:
        severity: warning
  - name: system-capacity-alert.rules
    rules:
    - alert: NooBaaSystemCapacityWarning85
      annotations:
        description: A NooBaa system is approaching its capacity, usage is more than
          85%
        message: A NooBaa System Is Approaching Its Capacity
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_system_capacity > 85
      labels:
        severity: warning
    - alert: NooBaaSystemCapacityWarning95
      annotations:
        description: A NooBaa system is approaching its capacity, usage is more than
          95%
        message: A NooBaa System Is Approaching Its Capacity
        severity_level: critical
        storage_type: NooBaa
      expr: |
        NooBaa_system_capacity > 95
      labels:
        severity: critical
    - alert: NooBaaSystemCapacityWarning100
      annotations:
        description: A NooBaa system approached its capacity, usage is at 100%
        message: A NooBaa System Approached Its Capacity
        severity_level: critical
        storage_type: NooBaa
      expr: |
        NooBaa_system_capacity == 100
      labels:
        severity: critical
`

const Sha256_deploy_internal_secret_empty_yaml = "d63aaeaf7f9c7c1421fcc138ee2f31d2461de0dec2f68120bc9cce367d4d4186"

const File_deploy_internal_secret_empty_yaml = `apiVersion: v1
kind: Secret
metadata:
  labels:
    app: noobaa
type: Opaque
data: {}
`

const Sha256_deploy_internal_service_mgmt_yaml = "14afee25800613df32344779addc74507951fbc2a31f639cfc190a0e56a5f29e"

const File_deploy_internal_service_mgmt_yaml = `apiVersion: v1
kind: Service
metadata:
  name: SYSNAME-mgmt
  labels:
    app: noobaa
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/scheme: http
    prometheus.io/port: "8080"
spec:
  type: LoadBalancer
  selector:
    noobaa-mgmt: SYSNAME
  ports:
    - port: 80
      name: mgmt
      targetPort: 8080
    - port: 443
      name: mgmt-https
      targetPort: 8443
    - port: 8444
      name: md-https
    - port: 8445
      name: bg-https
    - port: 8446
      name: hosted-agents-https
`

const Sha256_deploy_internal_service_monitor_yaml = "224b1ce993c390fa80898dedfee8380f2d0209d3702eb0b8b514dd380ade453c"

const File_deploy_internal_service_monitor_yaml = `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: noobaa-mgmt-service-monitor
  labels:
    app: noobaa
spec:
  endpoints:
  - port: mgmt
  namespaceSelector: {}
  selector:
    matchLabels:
      app: noobaa
`

const Sha256_deploy_internal_service_s3_yaml = "bfd8c420ca27482a996a9adfa0c8890fdd596850a3dbfd8c10d3ebaae1b5cc89"

const File_deploy_internal_service_s3_yaml = `apiVersion: v1
kind: Service
metadata:
  name: s3
  labels:
    app: noobaa
spec:
  type: LoadBalancer
  selector:
    noobaa-s3: SYSNAME
  ports:
    - port: 80
      targetPort: 6001
      name: s3
    - port: 443
      targetPort: 6443
      name: s3-https
`

const Sha256_deploy_internal_statefulset_core_yaml = "1f9a19c4b334c3ef2854c8317d62ebbf49aba52f8221c8a434c5b1d808c9be2e"

const File_deploy_internal_statefulset_core_yaml = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: noobaa-core
  labels:
    app: noobaa
spec:
  replicas: 1
  selector:
    matchLabels:
      noobaa-core: noobaa
  serviceName: noobaa-mgmt
  updateStrategy:
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      name: db
      labels:
        app: noobaa
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 50Gi
  template:
    metadata:
      labels:
        app: noobaa
        noobaa-core: noobaa
        noobaa-mgmt: noobaa
        noobaa-s3: noobaa
    spec:
      serviceAccountName: noobaa
      volumes:
      - name: logs
        emptyDir: {}
      initContainers:
#----------------#
# INIT CONTAINER #
#----------------#
      - name: init
        image: NOOBAA_CORE_IMAGE
        imagePullPolicy: IfNotPresent
        command:
        - /noobaa_init_files/noobaa_init.sh
        - init_mongo
        volumeMounts:
        - name: db
          mountPath: /mongo_data
      containers:
#----------------#
# CORE CONTAINER #
#----------------#
      - name: core
        image: NOOBAA_CORE_IMAGE
        imagePullPolicy: IfNotPresent
        volumeMounts:
        - name: logs
          mountPath: /log
        readinessProbe:
          tcpSocket:
            port: 6001 # ready when s3 port is open
          timeoutSeconds: 5
        resources:
          requests:
            cpu: "500m"
            memory: "1Gi"
          limits:
            cpu: "8"
            memory: "16Gi"
        ports:
        - containerPort: 6001
        - containerPort: 6443
        - containerPort: 8080
        - containerPort: 8443
        - containerPort: 8444
        - containerPort: 8445
        - containerPort: 8446
        - containerPort: 60100
        env:
        - name: CONTAINER_PLATFORM
          value: KUBERNETES
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: noobaa-server
              key: jwt
        - name: SERVER_SECRET
          valueFrom:
            secretKeyRef:
              name: noobaa-server
              key: server_secret
        - name: AGENT_PROFILE
          value: VALUE_AGENT_PROFILE
        - name: DISABLE_DEV_RANDOM_SEED
          value: "true"
        - name: OAUTH_AUTHORIZATION_ENDPOINT
          value: ""
        - name: OAUTH_TOKEN_ENDPOINT
          value: ""
        - name: container_dbg
          value: "" # any non-empty value will set the container to dbg mode
        - name: CONTAINER_CPU_REQUEST
          valueFrom:
            resourceFieldRef:
              resource: requests.cpu
        - name: CONTAINER_MEM_REQUEST
          valueFrom:
            resourceFieldRef:
              resource: requests.memory
        - name: CONTAINER_CPU_LIMIT
          valueFrom:
            resourceFieldRef:
              resource: limits.cpu
        - name: CONTAINER_MEM_LIMIT
          valueFrom:
            resourceFieldRef:
              resource: limits.memory
        # - name: ENDPOINT_FORKS_NUMBER
        #   value: "1"
#--------------------#
# DATABASE CONTAINER #
#--------------------#
      - name: db
        image: NOOBAA_DB_IMAGE
        imagePullPolicy: IfNotPresent
        command:
        - bash
        - -c
        - /opt/rh/rh-mongodb36/root/usr/bin/mongod --port 27017 --bind_ip 127.0.0.1 --dbpath /data/mongo/cluster/shard1
        resources:
          requests:
            cpu: "500m"
            memory: "1Gi"
          limits:
            cpu: "4"
            memory: "16Gi"
        volumeMounts:
        - name: db
          mountPath: /data
`

const Sha256_deploy_internal_text_system_status_readme_progress_tmpl = "d26aa1028e4a235018cc46e00392d3209d3e09e8320f3692be6346a9cfdf289a"

const File_deploy_internal_text_system_status_readme_progress_tmpl = `

	NooBaa operator is still working to reconcile this system.
	Check out the system status.phase, status.conditions, and events with:

		kubectl -n {{.NooBaa.Namespace}} describe noobaa
		kubectl -n {{.NooBaa.Namespace}} get noobaa -o yaml
		kubectl -n {{.NooBaa.Namespace}} get events --sort-by=metadata.creationTimestamp

	You can wait for a specific condition with:

		kubectl -n {{.NooBaa.Namespace}} wait noobaa/noobaa --for condition=available --timeout -1s

	NooBaa Core Version:     {{.CoreVersion}}
	NooBaa Operator Version: {{.OperatorVersion}}
`

const Sha256_deploy_internal_text_system_status_readme_ready_tmpl = "d2d8a51e85e4d75ee15f70b4e0baaf514149ae1a3678475a057a5ce04c6a0157"

const File_deploy_internal_text_system_status_readme_ready_tmpl = `

	Welcome to NooBaa!
	-----------------
	NooBaa Core Version:     {{.CoreVersion}}
	NooBaa Operator Version: {{.OperatorVersion}}

	Lets get started:

	1. Connect to Management console:

		Read your mgmt console login information (email & password) from secret: "{{.SecretAdmin.Name}}".

			kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq '.data|map_values(@base64d)'

		Open the management console service - take External IP/DNS or Node Port or use port forwarding:

			kubectl port-forward -n {{.ServiceMgmt.Namespace}} service/{{.ServiceMgmt.Name}} 11443:8443 &
			open https://localhost:11443

	2. Test S3 client:

		kubectl port-forward -n {{.ServiceS3.Namespace}} service/{{.ServiceS3.Name}} 10443:443 &
		NOOBAA_ACCESS_KEY=$(kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
		NOOBAA_SECRET_KEY=$(kubectl get secret {{.SecretAdmin.Name}} -n {{.SecretAdmin.Namespace}} -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
		alias s3='AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY aws --endpoint https://localhost:10443 --no-verify-ssl s3'
		s3 ls

`

const Sha256_deploy_internal_text_system_status_readme_rejected_tmpl = "32d46b0a1eadbe10501b2b3a6529503c76c0c77e25464f56f4ee9fd9115100c4"

const File_deploy_internal_text_system_status_readme_rejected_tmpl = `
	ERROR: NooBaa operator cannot reconcile this system spec.

	Check out the system status.phase, status.conditions, and events with:

		kubectl -n {{.NooBaa.Namespace}} describe noobaa
		kubectl -n {{.NooBaa.Namespace}} get noobaa -o yaml
		kubectl -n {{.NooBaa.Namespace}} get events --sort-by=metadata.creationTimestamp

	In order to retry, edit the system spec and the operator is watching and will be notified.

	NooBaa Core Version:     {{.CoreVersion}}
	NooBaa Operator Version: {{.OperatorVersion}}
`

const Sha256_deploy_namespace_yaml = "303398323535d7f8229cb1a5378ad019cf4fa7930891688e3eea55c77e7bf69a"

const File_deploy_namespace_yaml = `apiVersion: v1
kind: Namespace
metadata:
  name: noobaa
  labels:
    openshift.io/cluster-monitoring: "true"
`

const Sha256_deploy_obc_objectbucket_v1alpha1_objectbucket_cr_yaml = "0246e12a1337b2f68d408ff688b55fd6116bc7cd8f877e06d36e00d7255a81f9"

const File_deploy_obc_objectbucket_v1alpha1_objectbucket_cr_yaml = `apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucket
metadata:
  name: my-obc
spec:
  storageClassName: object-bucket-class
  reclaimPolicy: Delete
  claimRef:
    name: my-obc
    namespace: my-app
  endpoint:
    bucketHost: xxx
    bucketPort: 80
    bucketName: my-obc-1234-5678-1234-5678
    region: xxx
    subRegion: xxx
    additionalConfig: {}
  additionalState: {}
`

const Sha256_deploy_obc_objectbucket_v1alpha1_objectbucket_crd_yaml = "57432d2fe37757af1fe1e263e2426abd2e9b5afcbaf48eee543f94e2d30626ff"

const File_deploy_obc_objectbucket_v1alpha1_objectbucket_crd_yaml = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.objectbucket.io
spec:
  version: v1alpha1
  versions:
    - name: v1alpha1
      served: true
      storage: true
  group: objectbucket.io
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    singular: objectbucket
    shortNames:
      - ob
      - obs
  scope: Cluster
  subresources:
    status: {}
  additionalPrinterColumns:
  - JSONPath: .spec.storageClassName
    description: StorageClass
    name: Storage-Class
    type: string
  - JSONPath: .spec.claimRef.namespace
    description: ClaimNamespace
    name: Claim-Namespace
    type: string
  - JSONPath: .spec.claimRef.name
    description: ClaimName
    name: Claim-Name
    type: string
  - JSONPath: .spec.reclaimPolicy
    description: ReclaimPolicy
    name: Reclaim-Policy
    type: string
  - JSONPath: .status.phase
    description: Phase
    name: Phase
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object metadata.
          type: object
        spec:
          description: Specification of the desired behavior of the bucket.
          properties:
            storageClassName:
              description: StorageClass names the StorageClass object representing the 
                desired provisioner and parameters
              type: string
            reclaimPolicy:
              description: Describes a policy for end-of-life maintenance of ObjectBucket.
              enum:
                - "Delete"
                - "Retain"
                - "Recycle"
              type: string
            claimRef:
              description: ObjectReference to ObjectBucketClaim
              type: object
            endpoint:
              description: Endpoint contains all connection relevant data that an app may
                require for accessing the bucket
              properties:
                bucketHost:
                  description: Bucket address hostname
                  type: string
                bucketPort:
                  description: Bucket address port
                  type: integer
                bucketName:
                  description: Bucket name
                  type: string
                region:
                  description: Bucket region
                  type: string
                subRegion:
                  description: Bucket sub-region
                  type: string
                additionalConfig:
                  description: AdditionalConfig gives providers a location to set
                    proprietary config values (tenant, namespace, etc)
                  additionalProperties:
                    type: string
                  type: object
              type: object
            additionalState:
              description: additionalState gives providers a location to set
                proprietary config values (tenant, namespace, etc)
              additionalProperties:
                type: string
              type: object
          required:
            - storageClassName
          type: object
        status:
          description: Most recently observed status of the bucket.
          properties:
            phase:
              description: ObjectBucketStatusPhase is set by the controller to save the 
                state of the provisioning process
              enum:
                - "Bound"
                - "Released"
                - "Failed"
              type: string
          type: object
`

const Sha256_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml = "1a4391ac6d7393a2d3fba47f18c1097506a3f1f27bf6309c18897e30de9ec8c8"

const File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_cr_yaml = `apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: my-obc
  namespace: my-app
spec:
  storageClassName: object-bucket-class
  generateBucketName: my-obc
  additionalConfig: {}
`

const Sha256_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_crd_yaml = "edc3d10012249f73a226cca91830ae8b2751bcc10b41735db5a4a0f535e3ecb7"

const File_deploy_obc_objectbucket_v1alpha1_objectbucketclaim_crd_yaml = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.objectbucket.io
spec:
  version: v1alpha1
  versions:
    - name: v1alpha1
      served: true
      storage: true
  group: objectbucket.io
  names:
    kind: ObjectBucketClaim
    listKind: ObjectBucketClaimList
    plural: objectbucketclaims
    singular: objectbucketclaim
    shortNames:
      - obc
      - obcs
  scope: Namespaced
  subresources:
    status: {}
  additionalPrinterColumns:
  - JSONPath: .spec.storageClassName
    description: StorageClass
    name: Storage-Class
    type: string
  - JSONPath: .status.phase
    description: Phase
    name: Phase
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object metadata.
          type: object
        spec:
          description: Specification of the desired behavior of the claim.
          properties:
            storageClassName:
              description: StorageClass names the StorageClass object representing the 
                desired provisioner and parameters
              type: string
            bucketName:
              description: BucketName (not recommended) the name of the bucket. Caution!
                In-store bucket names may collide across namespaces.  If you define
                the name yourself, try to make it as unique as possible.
              type: string
            generateBucketName:
              description: GenerateBucketName (recommended) a prefix for a bucket name to be
                followed by a hyphen and 5 random characters. Protects against
                in-store name collisions.
              type: string
            additionalConfig:
              description: AdditionalConfig gives providers a location to set
                proprietary config values (tenant, namespace, etc)
              additionalProperties:
                type: string
              type: object
          required:
            - storageClassName
          type: object
        status:
          description: Most recently observed status of the claim.
          properties:
            phase:
              description: ObjectBucketClaimStatusPhase is set by the controller to save the state of the provisioning process
              enum:
                - "Pending"
                - "Bound"
                - "Released"
                - "Failed"
              type: string
          type: object
`

const Sha256_deploy_obc_storage_class_yaml = "9d03664a6263d8e54a00cdd604cc3cb6b5ce04ff9d73187b7fcb122bcbbdd1d2"

const File_deploy_obc_storage_class_yaml = `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: noobaa.noobaa.io
provisioner: noobaa.noobaa.io/obc
reclaimPolicy: Delete
`

const Sha256_deploy_olm_catalog_catalog_source_config_yaml = "c90eee41215b6a9810b320ea8dd6dc90a552d8f5f692464de964bf037df7ca78"

const File_deploy_olm_catalog_catalog_source_config_yaml = `apiVersion: operators.coreos.com/v1
kind: CatalogSourceConfig
metadata:
  name: noobaa-operator
  namespace: marketplace
spec:
  targetNamespace: olm
  source: noobaa-operator
  packages: noobaa-operator
`

const Sha256_deploy_olm_catalog_csv_config_yaml = "7902c00f83ed852ecb10b9ba2602e5c0271fc4f94afdc81dc757198942c63217"

const File_deploy_olm_catalog_csv_config_yaml = `role-paths:
- deploy/role.yaml
- deploy/cluster_role.yaml
`

const Sha256_deploy_olm_catalog_description_md = "1579500c3a3a89680441e60f9ffec59c51ad1b28f282e1c2e55de3d15a88ff62"

const File_deploy_olm_catalog_description_md = `The noobaa operator creates and reconciles a NooBaa system in a Kubernetes/Openshift cluster.

NooBaa provides an S3 object-store service abstraction and data placement policies to create hybrid and multi cloud data solutions.

For more information on using NooBaa refer to [Github](https://github.com/noobaa/noobaa-core) / [Website](https://www.noobaa.io) / [Articles](https://noobaa.desk.com). 

## How does it work?

- The operator deploys the noobaa core pod and two services - Mgmt (UI/API) and S3 (object-store).
- Both services require credentials which you will get from a secret that the operator creates - use describe noobaa to locate it.
- The service addresses will also appear in the describe output - pick the one that is suitable for your client:
    - minikube - use the NodePort address.
    - remote cluster - probably need one of the External addresses.
    - connect an application on the same cluster - use Internal DNS (though any address should work)
    
- Feel free to email us or open github issues on any question.

## Getting Started

### Notes:
- The following instructions are for **minikube** but it works on any Kubernetes/Openshift clusters.
- This will setup noobaa in the **my-noobaa-operator** namespace.
- You will need **jq**, **curl**, **kubectl** or **oc**, **aws-cli**.

### 1. Install OLM (if you don't have it already):
` + "`" + `` + "`" + `` + "`" + `
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/install.sh | bash -s 0.10.0
` + "`" + `` + "`" + `` + "`" + `

### 2. Install noobaa-operator:
` + "`" + `` + "`" + `` + "`" + `
kubectl create -f https://operatorhub.io/install/noobaa-operator.yaml
` + "`" + `` + "`" + `` + "`" + `
Wait for it to be ready:
` + "`" + `` + "`" + `` + "`" + `
kubectl wait pod -n my-noobaa-operator -l noobaa-operator --for=condition=ready
` + "`" + `` + "`" + `` + "`" + `

### 3. Create noobaa system:
` + "`" + `` + "`" + `` + "`" + `
curl -sL https://operatorhub.io/api/operator?packageName=noobaa-operator | 
    jq '.operator.customResourceDefinitions[0].yamlExample | .metadata.namespace="my-noobaa-operator"' |
    kubectl create -f -
` + "`" + `` + "`" + `` + "`" + `
Wait for it to be ready:
` + "`" + `` + "`" + `` + "`" + `
kubectl wait pod -n my-noobaa-operator -l noobaa-core --for=condition=ready
kubectl get noobaa -n my-noobaa-operator -w
# NAME     PHASE   MGMT-ENDPOINTS                  S3-ENDPOINTS                    IMAGE                    AGE
# noobaa   **Ready**   [https://192.168.64.12:31121]   [https://192.168.64.12:32557]   noobaa/noobaa-core:4.0   19m
` + "`" + `` + "`" + `` + "`" + `

### 4. Get system information to your shell:
` + "`" + `` + "`" + `` + "`" + `
NOOBAA_SECRET=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.accounts.admin.secretRef.name' )
NOOBAA_MGMT=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.services.serviceMgmt.nodePorts[0]' )
NOOBAA_S3=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.services.serviceS3.nodePorts[0]' )
NOOBAA_ACCESS_KEY=$(kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
NOOBAA_SECRET_KEY=$(kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
` + "`" + `` + "`" + `` + "`" + `

### 5. Connect to Mgmt UI:
` + "`" + `` + "`" + `` + "`" + `
# show email/password from the secret:
kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq '.data|map_values(@base64d)'

# open mgmt UI login:
open $NOOBAA_MGMT
` + "`" + `` + "`" + `` + "`" + `

### 6. Connect to S3 with aws-cli:
` + "`" + `` + "`" + `` + "`" + `
alias s3='AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY aws --endpoint $NOOBAA_S3 --no-verify-ssl s3'
s3 ls
s3 sync /var/log/ s3://first.bucket
s3 ls s3://first.bucket
` + "`" + `` + "`" + `` + "`" + `
`

const Sha256_deploy_olm_catalog_noobaa_operator_clusterserviceversion_yaml = "a903bea4780d1cd6a9c02c12dab5e52f6e2ef0b70b787e510ccbf7dd96a675a1"

const File_deploy_olm_catalog_noobaa_operator_clusterserviceversion_yaml = `apiVersion: operators.coreos.com/v1alpha1
kind: ClusterServiceVersion
metadata:
  annotations:
    categories: Storage,Big Data
    capabilities: Basic Install
    repository: https://github.com/noobaa/noobaa-operator
    containerImage: placeholder
    createdAt: 2019-07-08T13:10:20.940Z
    certified: "false"
    description: NooBaa is an object data service for hybrid and multi cloud environments.
    support: Red Hat
    alm-examples: |-
      [
        {
          "apiVersion": "noobaa.io/v1alpha1",
          "kind": "NooBaa",
          "metadata": {
            "name": "noobaa",
            "namespace": "my-noobaa-operator"
          },
          "spec": {}
        },
        {
          "apiVersion": "noobaa.io/v1alpha1",
          "kind": "BackingStore",
          "metadata": {
            "name": "aws1",
            "namespace": "my-noobaa-operator"
          },
          "spec": {
            "type": "aws-s3",
            "bucketName": "noobaa-aws1",
            "secret": {
              "name": "backing-store-secret-aws1",
              "namespace": "my-noobaa-operator"
            }
          }
        },
        {
          "apiVersion": "noobaa.io/v1alpha1",
          "kind": "BucketClass",
          "metadata": {
            "name": "default",
            "namespace": "my-noobaa-operator"
          },
          "spec": {
            "placementPolicy": {
              "tiers": [{
                "tier": {
                  "mirrors": [{
                    "mirror": {
                      "spread": ["aws1"]
                    }
                  }]
                }
              }]
            }
          }
        }
      ]
  name: placeholder
  namespace: placeholder
spec:
  displayName: NooBaa Operator
  version: "999.999.999-placeholder"
  maturity: alpha
  provider:
    name: NooBaa
  links:
  - name: Github
    url: https://github.com/noobaa/noobaa-core
  - name: Website
    url: https://www.noobaa.io
  - name: Articles
    url: https://noobaa.desk.com
  maintainers:
  - email: gmargali@redhat.com
    name: Guy Margalit
  - email: etamir@redhat.com
    name: Eran Tamir
  - email: nbecker@redhat.com
    name: Nimrod Becker
  keywords:
  - noobaa
  - kubernetes
  - openshift
  - cloud
  - hybrid
  - multi
  - data
  - storage
  - s3
  - tiering
  - mirroring
  labels:
    app: noobaa
  apiservicedefinitions: {}
  customresourcedefinitions:
    required: []
    owned: []
  installModes:
  - supported: true
    type: OwnNamespace
  - supported: false
    type: SingleNamespace
  - supported: false
    type: MultiNamespace
  - supported: false
    type: AllNamespaces
  install:
    strategy: deployment
    spec: {}
  description: placeholder
  icon:
  - mediatype: image/png
    base64data: placeholder`

const Sha256_deploy_olm_catalog_noobaa_icon_base64 = "4684eb3f4be354c728e210364a7e5e6806b68acb945b6e129ebc4d75fd97073c"

const File_deploy_olm_catalog_noobaa_icon_base64 = `iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAIAAAD2HxkiAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAA25pVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDIxIDc5LjE1NTc3MiwgMjAxNC8wMS8xMy0xOTo0NDowMCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtcE1NOk9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDpBNjQyRDdGQUIxMDkxMUU0QURFMENEMjA1QUJCMENEMyIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDoxOTA3OEQwNDAyRjAxMUU1QjdFQkI4RTFBMzY3NkQxRiIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDoxOTA3OEQwMzAyRjAxMUU1QjdFQkI4RTFBMzY3NkQxRiIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNCAoV2luZG93cykiPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDo5NWU4ZDg3YS1mNGU4LTRlMTYtOGIwYi1hZGIzYzY2OThkOGUiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6QTY0MkQ3RkFCMTA5MTFFNEFERTBDRDIwNUFCQjBDRDMiLz4gPC9yZGY6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`

const Sha256_deploy_olm_catalog_operator_group_yaml = "94f9961868713e06338bf0a2bbae73ad4983a5a52765b74a72ce0f5a62480af5"

const File_deploy_olm_catalog_operator_group_yaml = `apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: noobaa-operator-group
  namespace: my-noobaa-operator
spec:
  targetNamespaces:
  - my-noobaa-operator
`

const Sha256_deploy_olm_catalog_operator_source_yaml = "2f5cc3b1bec5332087fd6f3b80f0769c404a513d061ef822604fb87b6f301f30"

const File_deploy_olm_catalog_operator_source_yaml = `apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: noobaa-operator-source
  namespace: marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: noobaa
  displayName: "NooBaa Operator"
  publisher: "NooBaa"
`

const Sha256_deploy_olm_catalog_operator_subscription_yaml = "9eb0139d4fa841a97fe33e48ac5586bed89a3fa49ed26e20767108e531fe9ab0"

const File_deploy_olm_catalog_operator_subscription_yaml = `apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: noobaa-operator-subscription
  namespace: my-noobaa-operator
spec:
  channel: alpha
  name: noobaa-operator
  source: noobaa-operator-source
  sourceNamespace: marketplace
`

const Sha256_deploy_operator_yaml = "cca96157c1b8890cf472e7fadec04d3fb70b98741c25826d74fb6b1abd8f916a"

const File_deploy_operator_yaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: noobaa-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      noobaa-operator: deployment
  template:
    metadata:
      labels:
        app: noobaa
        noobaa-operator: deployment
    spec:
      serviceAccountName: noobaa
      containers:
        - name: noobaa-operator
          image: noobaa/noobaa-operator:2.0.2-rc1
          imagePullPolicy: IfNotPresent
          resources:
            limits:
              cpu: "250m"
              memory: "256Mi"
          env:
            - name: OPERATOR_NAME
              value: noobaa-operator
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
`

const Sha256_deploy_role_yaml = "c505b808b24f44248bb52f96464e89ac839ee2968daa84167155091206e4ce97"

const File_deploy_role_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: noobaa
rules:
- apiGroups:
  - noobaa.io
  resources:
  - '*'
  - noobaas
  - backingstores
  - bucketclasses
  - noobaas/finalizers
  - backingstores/finalizers
  - bucketclasses/finalizers
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - configmaps
  - secrets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  resourceNames:
    - noobaa
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  - replicasets
  - statefulsets
  verbs:
  - '*'
- apiGroups:
  - monitoring.coreos.com
  resources:
  - servicemonitors
  - prometheusrules
  verbs:
  - get
  - create
  - update
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - services/finalizers
  verbs:
  - update
- apiGroups:
  - apps
  resourceNames:
  - noobaa-operator
  resources:
  - deployments/finalizers
  verbs:
  - update
- apiGroups:
  - cloudcredential.openshift.io
  resources:
  - credentialsrequests
  verbs:
  - get
  - create
  - update
  - list
  - watch
- apiGroups:
  - ceph.rook.io
  resources:
  - cephobjectstores
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ceph.rook.io
  resources:
  - cephobjectstoreusers
  verbs:
  - get
  - create
  - update
  - list
  - watch
`

const Sha256_deploy_role_binding_yaml = "59a2627156ed3db9cd1a4d9c47e8c1044279c65e84d79c525e51274329cb16ff"

const File_deploy_role_binding_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: noobaa
subjects:
  - kind: ServiceAccount
    name: noobaa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: noobaa
`

const Sha256_deploy_service_account_yaml = "51241cd291100562ccd8bec1625c3779e212a58a0a21d4042937a98c73245d66"

const File_deploy_service_account_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa
`

