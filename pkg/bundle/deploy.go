package bundle

const Version = "5.18.0"

const Sha256_deploy_cluster_role_yaml = "3f8118853db73926c4f9d14be84ac8f81833c3a7a94a52ecf1e9ebcf712eee93"

const File_deploy_cluster_role_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: noobaa.noobaa.io
rules:
  - apiGroups:
      - noobaa.io
    resources:
      - "*"
      - noobaas
      - backingstores
      - bucketclasses
      - noobaas/finalizers
      - backingstores/finalizers
      - bucketclasses/finalizers
    verbs:
      - "*"
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
      - persistentvolumes
    verbs:
      - "*"
  - apiGroups:
      - ""
    resources:
      - namespaces
      - services
      - pods
    verbs:
      - get
      - list
      - watch
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
      - delete
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  - apiGroups: # from system:auth-delegator
      - authentication.k8s.io
    resources:
      - tokenreviews
    verbs:
      - create
  - apiGroups: # from system:auth-delegator
      - authorization.k8s.io
    resources:
      - subjectaccessreviews
    verbs:
      - create
  - apiGroups:
      - admissionregistration.k8s.io
    resources:
      - validatingwebhookconfigurations
    verbs:
      - get
      - update
      - list
  - apiGroups:
      - security.openshift.io
    resources:
      - securitycontextconstraints
    verbs:
      - '*'
  - apiGroups:
      - keda.sh
    resources:
      - clustertriggerauthentications
      - scaledobjects
      - triggerauthentications
      - scaledjobs
    verbs: 
      - get
      - list
      - watch
      - create
      - update
      - delete
  - apiGroups:
      - monitoring.coreos.com
    resources:
      - prometheuses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterrolebindings
    verbs:
      - create
      - get
      - list
      - watch
      - delete
  - apiGroups:
      - apiregistration.k8s.io
    verbs:
      - create
      - get
      - list
      - watch
      - delete
    resources:
      - 'apiservices'
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - rolebindings
    verbs:
      - create
      - get
      - list
      - watch
      - delete
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
    verbs:
      - create
      - get
      - list
      - watch
      - delete
  - apiGroups:
    - monitoring.coreos.com
    resources:
    - prometheuses
    verbs:
    - get
    - list
    - watch
  - apiGroups: 
      - coordination.k8s.io
    resources: 
      - leases
    verbs: 
      - get
      - watch
      - list
      - delete
      - update
      - create
  - apiGroups: 
      - objectstorage.k8s.io
    resources: 
      - buckets
      - bucketaccesses
      - bucketclaims
      - bucketaccessclasses
      - buckets/status
      - bucketaccesses/status
      - bucketclaims/status
      - bucketaccessclasses/status
    verbs: 
      - get
      - watch
      - list
      - delete
      - update
      - create
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

const Sha256_deploy_cosi_bucket_access_claim_yaml = "d2cc909826860644165d6baf660f765f7deae6e16b2bbb5d60fc55af0c8ff43c"

const File_deploy_cosi_bucket_access_claim_yaml = `apiVersion: objectstorage.k8s.io/v1alpha1
kind: BucketAccess
metadata:
  name: my-cosi-bucket-access
spec: 
  bucketClaimName: my-cosi-bucket-claim
  bucketAccessClassName: my-cosi-bucket-access-class
  credentialsSecretName: my-cosi-bucket-creds

`

const Sha256_deploy_cosi_bucket_access_class_yaml = "b0fec82ffd0214bc551199f27e6fe8d4132e3e1d717fdb25d38d6b917edfb3f0"

const File_deploy_cosi_bucket_access_class_yaml = `apiVersion: objectstorage.k8s.io/v1alpha1
kind: BucketAccessClass
metadata:
  name: my-cosi-bucket-access-class
driverName: noobaa.objectstorage.k8s.io
authenticationType: KEY
`

const Sha256_deploy_cosi_bucket_claim_yaml = "3780873c20475baea80775346e9e9214ae514ce7f0d62036d40be573c7911415"

const File_deploy_cosi_bucket_claim_yaml = `apiVersion: objectstorage.k8s.io/v1alpha1
kind: BucketClaim
metadata:
  name: my-cosi-bucket-claim
spec:
  bucketClassName: my-cosi-bucket-class
  protocols:
    - "S3"
`

const Sha256_deploy_cosi_bucket_class_yaml = "df623fd1a41c71246a658ac4b7b13255578210dc4d2756aac60391bc69e489c9"

const File_deploy_cosi_bucket_class_yaml = `apiVersion: objectstorage.k8s.io/v1alpha1
kind: BucketClass
metadata:
  name: my-cosi-bucket-class
driverName: noobaa.objectstorage.k8s.io
deletionPolicy: delete
parameters:
  placementPolicy: '{"tiers":[{"backingStores":["noobaa-default-backing-store"]}]}'
  replicationPolicy: '"{\"rules\":[{\"rule_id\":\"rule-1\",\"destination_bucket\":\"first.bucket\",\"filter\":{\"prefix\":\"a\"}}]}"'

`

const Sha256_deploy_cosi_cosi_bucket_yaml = "23af88f7164889958027390904c4b7d7a411593947e008dbf367f2c4be21f0ee"

const File_deploy_cosi_cosi_bucket_yaml = `apiVersion:  objectstorage.k8s.io/v1alpha1
kind:         Bucket
metadata:
  name: my-cosi-bucket-class-xxx
spec:
  protocols:
      - S3
  bucketClaim:
    name:             my-cosi-bucket-claim
    namespace:        my-app
  bucketClassName:    my-cosi-bucket-class
  deletionPolicy:     delete
  driverName:         noobaa.objectstorage.k8s.io
  parameters: {}

`

const Sha256_deploy_crds_noobaa_io_backingstores_yaml = "3c59eda2da91cf4ec6025491d8d1067b4626d3cad5fbd43bb0846ebd6d3126bf"

const File_deploy_crds_noobaa_io_backingstores_yaml = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.3
  name: backingstores.noobaa.io
spec:
  group: noobaa.io
  names:
    kind: BackingStore
    listKind: BackingStoreList
    plural: backingstores
    singular: backingstore
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: Type
      jsonPath: .spec.type
      name: Type
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: BackingStore is the Schema for the backingstores API
        properties:
          apiVersion:
            description: |-
              APIVersion defines the versioned schema of this representation of an object.
              Servers should convert recognized schemas to the latest internal value, and
              may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
            type: string
          kind:
            description: |-
              Kind is a string value representing the REST resource this object represents.
              Servers may infer this from the endpoint the client submits requests to.
              Cannot be updated.
              In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
            type: string
          metadata:
            type: object
          spec:
            description: Specification of the desired behavior of the noobaa BackingStore.
            properties:
              awsS3:
                description: AWSS3Spec specifies a backing store of type aws-s3
                properties:
                  awsSTSRoleARN:
                    description: AWSSTSRoleARN allows to Assume Role and use AssumeRoleWithWebIdentity
                    type: string
                  region:
                    description: Region is the AWS region
                    type: string
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  sslDisabled:
                    description: SSLDisabled allows to disable SSL and use plain http
                    type: boolean
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - targetBucket
                type: object
              azureBlob:
                description: AzureBlob specifies a backing store of type azure-blob
                properties:
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AccountName and AccountKey as provided by Azure Blob.
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  targetBlobContainer:
                    description: TargetBlobContainer is the name of the target Azure
                      Blob container
                    type: string
                required:
                - secret
                - targetBlobContainer
                type: object
              googleCloudStorage:
                description: GoogleCloudStorage specifies a backing store of type
                  google-cloud-storage
                properties:
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define GoogleServiceAccountPrivateKeyJson containing the entire json string as provided by Google.
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - secret
                - targetBucket
                type: object
              ibmCos:
                description: IBMCos specifies a backing store of type ibm-cos
                properties:
                  endpoint:
                    description: 'Endpoint is the IBM COS compatible endpoint: http(s)://host:port'
                    type: string
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  signatureVersion:
                    description: SignatureVersion specifies the client signature version
                      to use when signing requests.
                    type: string
                  targetBucket:
                    description: TargetBucket is the name of the target IBM COS bucket
                    type: string
                required:
                - endpoint
                - secret
                - targetBucket
                type: object
              pvPool:
                description: PVPool specifies a backing store of type pv-pool
                properties:
                  numVolumes:
                    description: NumVolumes is the number of volumes to allocate
                    type: integer
                  resources:
                    description: VolumeResources represents the minimum resources
                      each volume should have.
                    properties:
                      limits:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: |-
                          Limits describes the maximum amount of compute resources allowed.
                          More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                        type: object
                      requests:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: |-
                          Requests describes the minimum amount of compute resources required.
                          If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                          otherwise to an implementation-defined value. Requests cannot exceed Limits.
                          More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                        type: object
                    type: object
                  secret:
                    description: |-
                      Secret refers to a secret that provides the agent configuration
                      The secret should define AGENT_CONFIG containing agent_configuration from noobaa-core.
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  storageClass:
                    description: StorageClass is the name of the storage class to
                      use for the PV's
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
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  signatureVersion:
                    description: SignatureVersion specifies the client signature version
                      to use when signing requests.
                    type: string
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - endpoint
                - secret
                - targetBucket
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
                  reconciliation
                items:
                  description: |-
                    Condition represents the state of the operator's
                    reconciliation functionality.
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
                      description: ConditionType is the state of the operator's reconciliation
                        functionality.
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              mode:
                description: Mode specifies the updating mode of a BackingStore
                properties:
                  modeCode:
                    description: ModeCode specifies the updated mode of backingstore
                    type: string
                  timeStamp:
                    description: TimeStamp specifies the update time of backingstore
                      new mode
                    type: string
                type: object
              phase:
                description: Phase is a simple, high-level summary of where the backing
                  store is in its lifecycle
                type: string
              relatedObjects:
                description: RelatedObjects is a list of objects related to this operator.
                items:
                  description: ObjectReference contains enough information to let
                    you inspect or modify the referred object.
                  properties:
                    apiVersion:
                      description: API version of the referent.
                      type: string
                    fieldPath:
                      description: |-
                        If referring to a piece of an object instead of an entire object, this string
                        should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
                        For example, if the object reference is to a container within a pod, this would take on a value like:
                        "spec.containers{name}" (where "name" refers to the name of the container that triggered
                        the event) or if no container name is specified "spec.containers[2]" (container with
                        index 2 in this pod). This syntax is chosen only to have some well-defined way of
                        referencing a part of an object.
                      type: string
                    kind:
                      description: |-
                        Kind of the referent.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
                      type: string
                    name:
                      description: |-
                        Name of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                      type: string
                    namespace:
                      description: |-
                        Namespace of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
                      type: string
                    resourceVersion:
                      description: |-
                        Specific resourceVersion to which this reference is made, if any.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
                      type: string
                    uid:
                      description: |-
                        UID of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
                      type: string
                  type: object
                  x-kubernetes-map-type: atomic
                type: array
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_crds_noobaa_io_bucketclasses_yaml = "303a0b43c30509718a314dd4a0f733679229416cddc52daffe08434d7d4ea652"

const File_deploy_crds_noobaa_io_bucketclasses_yaml = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.3
  name: bucketclasses.noobaa.io
spec:
  group: noobaa.io
  names:
    kind: BucketClass
    listKind: BucketClassList
    plural: bucketclasses
    singular: bucketclass
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: Placement
      jsonPath: .spec.placementPolicy
      name: Placement
      type: string
    - description: NamespacePolicy
      jsonPath: .spec.namespacePolicy
      name: NamespacePolicy
      type: string
    - description: Quota
      jsonPath: .spec.quota
      name: Quota
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: BucketClass is the Schema for the bucketclasses API
        properties:
          apiVersion:
            description: |-
              APIVersion defines the versioned schema of this representation of an object.
              Servers should convert recognized schemas to the latest internal value, and
              may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
            type: string
          kind:
            description: |-
              Kind is a string value representing the REST resource this object represents.
              Servers may infer this from the endpoint the client submits requests to.
              Cannot be updated.
              In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
            type: string
          metadata:
            type: object
          spec:
            description: Specification of the desired behavior of the noobaa BucketClass.
            properties:
              namespacePolicy:
                description: NamespacePolicy specifies the namespace policy for the
                  bucket class
                properties:
                  cache:
                    description: Cache is a namespace policy configuration of type
                      Cache
                    properties:
                      caching:
                        description: Caching is the cache specification for the ns
                          policy
                        properties:
                          prefix:
                            description: Prefix is prefix of the future cached data
                            type: string
                          ttl:
                            description: TTL specifies the cache ttl
                            type: integer
                        type: object
                      hubResource:
                        description: HubResource is the read and write resource name
                          to use
                        type: string
                    type: object
                  multi:
                    description: Multi is a namespace policy configuration of type
                      Multi
                    properties:
                      readResources:
                        description: ReadResources is an ordered list of read resources
                          names to use
                        items:
                          type: string
                        type: array
                      writeResource:
                        description: WriteResource is the write resource name to use
                        type: string
                    type: object
                  single:
                    description: Single is a namespace policy configuration of type
                      Single
                    properties:
                      resource:
                        description: Resource is the read and write resource name
                          to use
                        type: string
                    type: object
                  type:
                    description: Type is the namespace policy type
                    type: string
                type: object
              placementPolicy:
                description: PlacementPolicy specifies the placement policy for the
                  bucket class
                properties:
                  tiers:
                    description: |-
                      Tiers is an ordered list of tiers to use.
                      The model is a waterfall - push to first tier by default,
                      and when no more space spill "cold" storage to next tier.
                    items:
                      description: Tier specifies a storage tier
                      properties:
                        backingStores:
                          description: |-
                            BackingStores is an unordered list of backing store names.
                            The meaning of the list depends on the placement.
                          items:
                            type: string
                          type: array
                        placement:
                          description: |-
                            Placement specifies the type of placement for the tier
                            If empty it should have a single backing store.
                          enum:
                          - Spread
                          - Mirror
                          type: string
                      type: object
                    type: array
                type: object
              quota:
                description: Quota specifies the quota configuration for the bucket
                  class
                properties:
                  maxObjects:
                    description: limits the max total quantity of objects per bucket
                    type: string
                  maxSize:
                    description: limits the max total size of objects per bucket
                    type: string
                type: object
              replicationPolicy:
                description: ReplicationPolicy specifies a json of replication rules
                  for the bucketclass
                type: string
            type: object
          status:
            description: Most recently observed status of the noobaa BackingStore.
            properties:
              conditions:
                description: Conditions is a list of conditions related to operator
                  reconciliation
                items:
                  description: |-
                    Condition represents the state of the operator's
                    reconciliation functionality.
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
                      description: ConditionType is the state of the operator's reconciliation
                        functionality.
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              mode:
                description: Mode is a simple, high-level summary of where the System
                  is in its lifecycle
                type: string
              phase:
                description: Phase is a simple, high-level summary of where the System
                  is in its lifecycle
                type: string
              relatedObjects:
                description: RelatedObjects is a list of objects related to this operator.
                items:
                  description: ObjectReference contains enough information to let
                    you inspect or modify the referred object.
                  properties:
                    apiVersion:
                      description: API version of the referent.
                      type: string
                    fieldPath:
                      description: |-
                        If referring to a piece of an object instead of an entire object, this string
                        should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
                        For example, if the object reference is to a container within a pod, this would take on a value like:
                        "spec.containers{name}" (where "name" refers to the name of the container that triggered
                        the event) or if no container name is specified "spec.containers[2]" (container with
                        index 2 in this pod). This syntax is chosen only to have some well-defined way of
                        referencing a part of an object.
                      type: string
                    kind:
                      description: |-
                        Kind of the referent.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
                      type: string
                    name:
                      description: |-
                        Name of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                      type: string
                    namespace:
                      description: |-
                        Namespace of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
                      type: string
                    resourceVersion:
                      description: |-
                        Specific resourceVersion to which this reference is made, if any.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
                      type: string
                    uid:
                      description: |-
                        UID of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
                      type: string
                  type: object
                  x-kubernetes-map-type: atomic
                type: array
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_crds_noobaa_io_namespacestores_yaml = "427b53370d424315e81fdca907fb51f8106c56e5a2a7b186384348f794ad330d"

const File_deploy_crds_noobaa_io_namespacestores_yaml = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.3
  name: namespacestores.noobaa.io
spec:
  group: noobaa.io
  names:
    kind: NamespaceStore
    listKind: NamespaceStoreList
    plural: namespacestores
    singular: namespacestore
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: Type
      jsonPath: .spec.type
      name: Type
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: NamespaceStore is the Schema for the namespacestores API
        properties:
          apiVersion:
            description: |-
              APIVersion defines the versioned schema of this representation of an object.
              Servers should convert recognized schemas to the latest internal value, and
              may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
            type: string
          kind:
            description: |-
              Kind is a string value representing the REST resource this object represents.
              Servers may infer this from the endpoint the client submits requests to.
              Cannot be updated.
              In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
            type: string
          metadata:
            type: object
          spec:
            description: Specification of the desired behavior of the noobaa NamespaceStore.
            properties:
              accessMode:
                description: AccessMode is an enum of supported access modes
                type: string
              awsS3:
                description: AWSS3Spec specifies a namespace store of type aws-s3
                properties:
                  awsSTSRoleARN:
                    description: AWSSTSRoleARN allows to Assume Role and use AssumeRoleWithWebIdentity
                    type: string
                  region:
                    description: Region is the AWS region
                    type: string
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  sslDisabled:
                    description: SSLDisabled allows to disable SSL and use plain http
                    type: boolean
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - targetBucket
                type: object
              azureBlob:
                description: AzureBlob specifies a namespace store of type azure-blob
                properties:
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AccountName and AccountKey as provided by Azure Blob.
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  targetBlobContainer:
                    description: TargetBlobContainer is the name of the target Azure
                      Blob container
                    type: string
                required:
                - secret
                - targetBlobContainer
                type: object
              googleCloudStorage:
                description: GoogleCloudStorage specifies a namespace store of type
                  google-cloud-storage
                properties:
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define GoogleServiceAccountPrivateKeyJson containing the entire json string as provided by Google.
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - secret
                - targetBucket
                type: object
              ibmCos:
                description: IBMCos specifies a namespace store of type ibm-cos
                properties:
                  endpoint:
                    description: 'Endpoint is the IBM COS compatible endpoint: http(s)://host:port'
                    type: string
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  signatureVersion:
                    description: SignatureVersion specifies the client signature version
                      to use when signing requests.
                    type: string
                  targetBucket:
                    description: TargetBucket is the name of the target IBM COS bucket
                    type: string
                required:
                - endpoint
                - secret
                - targetBucket
                type: object
              nsfs:
                description: NSFS specifies a namespace store of type nsfs
                properties:
                  fsBackend:
                    description: FsBackend is the backend type of the file system
                    enum:
                    - CEPH_FS
                    - GPFS
                    - NFSv4
                    type: string
                  pvcName:
                    description: PvcName is the name of the pvc in which the file
                      system resides
                    type: string
                  subPath:
                    description: SubPath is a path to a sub directory in the pvc file
                      system
                    type: string
                required:
                - pvcName
                type: object
              s3Compatible:
                description: S3Compatible specifies a namespace store of type s3-compatible
                properties:
                  endpoint:
                    description: 'Endpoint is the S3 compatible endpoint: http(s)://host:port'
                    type: string
                  secret:
                    description: |-
                      Secret refers to a secret that provides the credentials
                      The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                    properties:
                      name:
                        description: name is unique within a namespace to reference
                          a secret resource.
                        type: string
                      namespace:
                        description: namespace defines the space within which the
                          secret name must be unique.
                        type: string
                    type: object
                    x-kubernetes-map-type: atomic
                  signatureVersion:
                    description: SignatureVersion specifies the client signature version
                      to use when signing requests.
                    type: string
                  targetBucket:
                    description: TargetBucket is the name of the target S3 bucket
                    type: string
                required:
                - endpoint
                - secret
                - targetBucket
                type: object
              type:
                description: Type is an enum of supported types
                type: string
            required:
            - type
            type: object
          status:
            description: Most recently observed status of the noobaa NamespaceStore.
            properties:
              conditions:
                description: Conditions is a list of conditions related to operator
                  reconciliation
                items:
                  description: |-
                    Condition represents the state of the operator's
                    reconciliation functionality.
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
                      description: ConditionType is the state of the operator's reconciliation
                        functionality.
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              mode:
                description: Mode specifies the updating mode of a NamespaceStore
                properties:
                  modeCode:
                    description: ModeCode specifies the updated mode of namespacestore
                    type: string
                  timeStamp:
                    description: TimeStamp specifies the update time of namespacestore
                      new mode
                    type: string
                type: object
              phase:
                description: Phase is a simple, high-level summary of where the namespace
                  store is in its lifecycle
                type: string
              relatedObjects:
                description: RelatedObjects is a list of objects related to this operator.
                items:
                  description: ObjectReference contains enough information to let
                    you inspect or modify the referred object.
                  properties:
                    apiVersion:
                      description: API version of the referent.
                      type: string
                    fieldPath:
                      description: |-
                        If referring to a piece of an object instead of an entire object, this string
                        should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
                        For example, if the object reference is to a container within a pod, this would take on a value like:
                        "spec.containers{name}" (where "name" refers to the name of the container that triggered
                        the event) or if no container name is specified "spec.containers[2]" (container with
                        index 2 in this pod). This syntax is chosen only to have some well-defined way of
                        referencing a part of an object.
                      type: string
                    kind:
                      description: |-
                        Kind of the referent.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
                      type: string
                    name:
                      description: |-
                        Name of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                      type: string
                    namespace:
                      description: |-
                        Namespace of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
                      type: string
                    resourceVersion:
                      description: |-
                        Specific resourceVersion to which this reference is made, if any.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
                      type: string
                    uid:
                      description: |-
                        UID of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
                      type: string
                  type: object
                  x-kubernetes-map-type: atomic
                type: array
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_crds_noobaa_io_noobaaaccounts_yaml = "4317a1b539d6a491f6afd9f508e75103c50797b7280ad53941f0d2b546f0f6c1"

const File_deploy_crds_noobaa_io_noobaaaccounts_yaml = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.3
  name: noobaaaccounts.noobaa.io
spec:
  group: noobaa.io
  names:
    kind: NooBaaAccount
    listKind: NooBaaAccountList
    plural: noobaaaccounts
    singular: noobaaaccount
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: NooBaaAccount is the Schema for the NooBaaAccounts API
        properties:
          apiVersion:
            description: |-
              APIVersion defines the versioned schema of this representation of an object.
              Servers should convert recognized schemas to the latest internal value, and
              may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
            type: string
          kind:
            description: |-
              Kind is a string value representing the REST resource this object represents.
              Servers may infer this from the endpoint the client submits requests to.
              Cannot be updated.
              In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
            type: string
          metadata:
            type: object
          spec:
            description: Specification of the desired behavior of the NooBaaAccount.
            properties:
              allow_bucket_creation:
                description: AllowBucketCreate specifies if new buckets can be created
                  by this account
                type: boolean
              default_resource:
                description: DefaultResource specifies which backingstore this account
                  will use to create new buckets
                type: string
              force_md5_etag:
                description: ForceMd5Etag specifies whether MD5 Etags should be calculated
                  for the account or not
                type: boolean
              nsfs_account_config:
                description: NsfsAccountConfig specifies the configurations on Namespace
                  FS
                nullable: true
                properties:
                  distinguished_name:
                    type: string
                  gid:
                    type: integer
                  new_buckets_path:
                    type: string
                  nsfs_only:
                    type: boolean
                  uid:
                    type: integer
                required:
                - new_buckets_path
                - nsfs_only
                type: object
            required:
            - allow_bucket_creation
            type: object
          status:
            description: Most recently observed status of the NooBaaAccount.
            properties:
              conditions:
                description: Conditions is a list of conditions related to operator
                  reconciliation
                items:
                  description: |-
                    Condition represents the state of the operator's
                    reconciliation functionality.
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
                      description: ConditionType is the state of the operator's reconciliation
                        functionality.
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              phase:
                description: Phase is a simple, high-level summary of where the noobaa
                  user is in its lifecycle
                type: string
              relatedObjects:
                description: RelatedObjects is a list of objects related to this operator.
                items:
                  description: ObjectReference contains enough information to let
                    you inspect or modify the referred object.
                  properties:
                    apiVersion:
                      description: API version of the referent.
                      type: string
                    fieldPath:
                      description: |-
                        If referring to a piece of an object instead of an entire object, this string
                        should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
                        For example, if the object reference is to a container within a pod, this would take on a value like:
                        "spec.containers{name}" (where "name" refers to the name of the container that triggered
                        the event) or if no container name is specified "spec.containers[2]" (container with
                        index 2 in this pod). This syntax is chosen only to have some well-defined way of
                        referencing a part of an object.
                      type: string
                    kind:
                      description: |-
                        Kind of the referent.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
                      type: string
                    name:
                      description: |-
                        Name of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                      type: string
                    namespace:
                      description: |-
                        Namespace of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
                      type: string
                    resourceVersion:
                      description: |-
                        Specific resourceVersion to which this reference is made, if any.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
                      type: string
                    uid:
                      description: |-
                        UID of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
                      type: string
                  type: object
                  x-kubernetes-map-type: atomic
                type: array
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_crds_noobaa_io_noobaas_yaml = "e862d263d097ed43f774784eaaf9a616967746b67608fadbe4ca71d93b220ab6"

const File_deploy_crds_noobaa_io_noobaas_yaml = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.16.3
  name: noobaas.noobaa.io
spec:
  group: noobaa.io
  names:
    kind: NooBaa
    listKind: NooBaaList
    plural: noobaas
    shortNames:
    - nb
    singular: noobaa
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: S3 Endpoints
      jsonPath: .status.services.serviceS3.nodePorts
      name: S3-Endpoints
      type: string
    - description: STS Endpoints
      jsonPath: .status.services.serviceSts.nodePorts
      name: Sts-Endpoints
      type: string
    - description: Syslog Endpoints
      jsonPath: .status.services.serviceSyslog.nodePorts
      name: Syslog-Endpoints
      type: string
    - description: Actual Image
      jsonPath: .status.actualImage
      name: Image
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: NooBaa is the Schema for the NooBaas API
        properties:
          apiVersion:
            description: |-
              APIVersion defines the versioned schema of this representation of an object.
              Servers should convert recognized schemas to the latest internal value, and
              may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
            type: string
          kind:
            description: |-
              Kind is a string value representing the REST resource this object represents.
              Servers may infer this from the endpoint the client submits requests to.
              Cannot be updated.
              In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
            type: string
          metadata:
            type: object
          spec:
            description: Specification of the desired behavior of the noobaa system.
            properties:
              affinity:
                description: Affinity (optional) passed through to noobaa's pods
                properties:
                  nodeAffinity:
                    description: Describes node affinity scheduling rules for the
                      pod.
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          The scheduler will prefer to schedule pods to nodes that satisfy
                          the affinity expressions specified by this field, but it may choose
                          a node that violates one or more of the expressions. The node that is
                          most preferred is the one with the greatest sum of weights, i.e.
                          for each node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions, etc.),
                          compute a sum by iterating through the elements of this field and adding
                          "weight" to the sum if the node matches the corresponding matchExpressions; the
                          node(s) with the highest sum are the most preferred.
                        items:
                          description: |-
                            An empty preferred scheduling term matches all objects with implicit weight 0
                            (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                          properties:
                            preference:
                              description: A node selector term, associated with the
                                corresponding weight.
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  items:
                                    description: |-
                                      A node selector requirement is a selector that contains values, a key, and an operator
                                      that relates the key and values.
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          Represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                        type: string
                                      values:
                                        description: |-
                                          An array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. If the operator is Gt or Lt, the values
                                          array must have a single element, which will be interpreted as an integer.
                                          This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  items:
                                    description: |-
                                      A node selector requirement is a selector that contains values, a key, and an operator
                                      that relates the key and values.
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          Represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                        type: string
                                      values:
                                        description: |-
                                          An array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. If the operator is Gt or Lt, the values
                                          array must have a single element, which will be interpreted as an integer.
                                          This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                              type: object
                              x-kubernetes-map-type: atomic
                            weight:
                              description: Weight associated with matching the corresponding
                                nodeSelectorTerm, in the range 1-100.
                              format: int32
                              type: integer
                          required:
                          - preference
                          - weight
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          If the affinity requirements specified by this field are not met at
                          scheduling time, the pod will not be scheduled onto the node.
                          If the affinity requirements specified by this field cease to be met
                          at some point during pod execution (e.g. due to an update), the system
                          may or may not try to eventually evict the pod from its node.
                        properties:
                          nodeSelectorTerms:
                            description: Required. A list of node selector terms.
                              The terms are ORed.
                            items:
                              description: |-
                                A null or empty node selector term matches no objects. The requirements of
                                them are ANDed.
                                The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  items:
                                    description: |-
                                      A node selector requirement is a selector that contains values, a key, and an operator
                                      that relates the key and values.
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          Represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                        type: string
                                      values:
                                        description: |-
                                          An array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. If the operator is Gt or Lt, the values
                                          array must have a single element, which will be interpreted as an integer.
                                          This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  items:
                                    description: |-
                                      A node selector requirement is a selector that contains values, a key, and an operator
                                      that relates the key and values.
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          Represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                        type: string
                                      values:
                                        description: |-
                                          An array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. If the operator is Gt or Lt, the values
                                          array must have a single element, which will be interpreted as an integer.
                                          This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                              type: object
                              x-kubernetes-map-type: atomic
                            type: array
                            x-kubernetes-list-type: atomic
                        required:
                        - nodeSelectorTerms
                        type: object
                        x-kubernetes-map-type: atomic
                    type: object
                  podAffinity:
                    description: Describes pod affinity scheduling rules (e.g. co-locate
                      this pod in the same node, zone, etc. as some other pod(s)).
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          The scheduler will prefer to schedule pods to nodes that satisfy
                          the affinity expressions specified by this field, but it may choose
                          a node that violates one or more of the expressions. The node that is
                          most preferred is the one with the greatest sum of weights, i.e.
                          for each node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions, etc.),
                          compute a sum by iterating through the elements of this field and adding
                          "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              properties:
                                labelSelector:
                                  description: |-
                                    A label query over a set of resources, in this case pods.
                                    If it's null, this PodAffinityTerm matches with no Pods.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      items:
                                        description: |-
                                          A label selector requirement is a selector that contains values, a key, and an operator that
                                          relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: |-
                                              operator represents a key's relationship to a set of values.
                                              Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: |-
                                              values is an array of string values. If the operator is In or NotIn,
                                              the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This array is replaced during a strategic
                                              merge patch.
                                            items:
                                              type: string
                                            type: array
                                            x-kubernetes-list-type: atomic
                                        required:
                                        - key
                                        - operator
                                        type: object
                                      type: array
                                      x-kubernetes-list-type: atomic
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: |-
                                        matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions, whose key field is "key", the
                                        operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                  x-kubernetes-map-type: atomic
                                matchLabelKeys:
                                  description: |-
                                    MatchLabelKeys is a set of pod label keys to select which pods will
                                    be taken into consideration. The keys are used to lookup values from the
                                    incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key in (value)` + "`" + `
                                    to select the group of existing pods which pods will be taken into consideration
                                    for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                    pod labels will be ignored. The default value is empty.
                                    The same key is forbidden to exist in both matchLabelKeys and labelSelector.
                                    Also, matchLabelKeys cannot be set when labelSelector isn't set.
                                    This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                mismatchLabelKeys:
                                  description: |-
                                    MismatchLabelKeys is a set of pod label keys to select which pods will
                                    be taken into consideration. The keys are used to lookup values from the
                                    incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key notin (value)` + "`" + `
                                    to select the group of existing pods which pods will be taken into consideration
                                    for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                    pod labels will be ignored. The default value is empty.
                                    The same key is forbidden to exist in both mismatchLabelKeys and labelSelector.
                                    Also, mismatchLabelKeys cannot be set when labelSelector isn't set.
                                    This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                namespaceSelector:
                                  description: |-
                                    A label query over the set of namespaces that the term applies to.
                                    The term is applied to the union of the namespaces selected by this field
                                    and the ones listed in the namespaces field.
                                    null selector and null or empty namespaces list means "this pod's namespace".
                                    An empty selector ({}) matches all namespaces.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      items:
                                        description: |-
                                          A label selector requirement is a selector that contains values, a key, and an operator that
                                          relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: |-
                                              operator represents a key's relationship to a set of values.
                                              Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: |-
                                              values is an array of string values. If the operator is In or NotIn,
                                              the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This array is replaced during a strategic
                                              merge patch.
                                            items:
                                              type: string
                                            type: array
                                            x-kubernetes-list-type: atomic
                                        required:
                                        - key
                                        - operator
                                        type: object
                                      type: array
                                      x-kubernetes-list-type: atomic
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: |-
                                        matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions, whose key field is "key", the
                                        operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                  x-kubernetes-map-type: atomic
                                namespaces:
                                  description: |-
                                    namespaces specifies a static list of namespace names that the term applies to.
                                    The term is applied to the union of the namespaces listed in this field
                                    and the ones selected by namespaceSelector.
                                    null or empty namespaces list and null namespaceSelector means "this pod's namespace".
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                topologyKey:
                                  description: |-
                                    This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
                                    the labelSelector in the specified namespaces, where co-located is defined as running on a node
                                    whose value of the label with key topologyKey matches that of any node on which any of the
                                    selected pods is running.
                                    Empty topologyKey is not allowed.
                                  type: string
                              required:
                              - topologyKey
                              type: object
                            weight:
                              description: |-
                                weight associated with matching the corresponding podAffinityTerm,
                                in the range 1-100.
                              format: int32
                              type: integer
                          required:
                          - podAffinityTerm
                          - weight
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          If the affinity requirements specified by this field are not met at
                          scheduling time, the pod will not be scheduled onto the node.
                          If the affinity requirements specified by this field cease to be met
                          at some point during pod execution (e.g. due to a pod label update), the
                          system may or may not try to eventually evict the pod from its node.
                          When there are multiple elements, the lists of nodes corresponding to each
                          podAffinityTerm are intersected, i.e. all terms must be satisfied.
                        items:
                          description: |-
                            Defines a set of pods (namely those matching the labelSelector
                            relative to the given namespace(s)) that this pod should be
                            co-located (affinity) or not co-located (anti-affinity) with,
                            where co-located is defined as running on a node whose value of
                            the label with key <topologyKey> matches that of any node on which
                            a pod of the set of pods is running
                          properties:
                            labelSelector:
                              description: |-
                                A label query over a set of resources, in this case pods.
                                If it's null, this PodAffinityTerm matches with no Pods.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  items:
                                    description: |-
                                      A label selector requirement is a selector that contains values, a key, and an operator that
                                      relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          operator represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: |-
                                          values is an array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. This array is replaced during a strategic
                                          merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: |-
                                    matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions, whose key field is "key", the
                                    operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                              x-kubernetes-map-type: atomic
                            matchLabelKeys:
                              description: |-
                                MatchLabelKeys is a set of pod label keys to select which pods will
                                be taken into consideration. The keys are used to lookup values from the
                                incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key in (value)` + "`" + `
                                to select the group of existing pods which pods will be taken into consideration
                                for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                pod labels will be ignored. The default value is empty.
                                The same key is forbidden to exist in both matchLabelKeys and labelSelector.
                                Also, matchLabelKeys cannot be set when labelSelector isn't set.
                                This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            mismatchLabelKeys:
                              description: |-
                                MismatchLabelKeys is a set of pod label keys to select which pods will
                                be taken into consideration. The keys are used to lookup values from the
                                incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key notin (value)` + "`" + `
                                to select the group of existing pods which pods will be taken into consideration
                                for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                pod labels will be ignored. The default value is empty.
                                The same key is forbidden to exist in both mismatchLabelKeys and labelSelector.
                                Also, mismatchLabelKeys cannot be set when labelSelector isn't set.
                                This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            namespaceSelector:
                              description: |-
                                A label query over the set of namespaces that the term applies to.
                                The term is applied to the union of the namespaces selected by this field
                                and the ones listed in the namespaces field.
                                null selector and null or empty namespaces list means "this pod's namespace".
                                An empty selector ({}) matches all namespaces.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  items:
                                    description: |-
                                      A label selector requirement is a selector that contains values, a key, and an operator that
                                      relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          operator represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: |-
                                          values is an array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. This array is replaced during a strategic
                                          merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: |-
                                    matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions, whose key field is "key", the
                                    operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                              x-kubernetes-map-type: atomic
                            namespaces:
                              description: |-
                                namespaces specifies a static list of namespace names that the term applies to.
                                The term is applied to the union of the namespaces listed in this field
                                and the ones selected by namespaceSelector.
                                null or empty namespaces list and null namespaceSelector means "this pod's namespace".
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            topologyKey:
                              description: |-
                                This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where co-located is defined as running on a node
                                whose value of the label with key topologyKey matches that of any node on which any of the
                                selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
                          required:
                          - topologyKey
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                    type: object
                  podAntiAffinity:
                    description: Describes pod anti-affinity scheduling rules (e.g.
                      avoid putting this pod in the same node, zone, etc. as some
                      other pod(s)).
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          The scheduler will prefer to schedule pods to nodes that satisfy
                          the anti-affinity expressions specified by this field, but it may choose
                          a node that violates one or more of the expressions. The node that is
                          most preferred is the one with the greatest sum of weights, i.e.
                          for each node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling anti-affinity expressions, etc.),
                          compute a sum by iterating through the elements of this field and adding
                          "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              properties:
                                labelSelector:
                                  description: |-
                                    A label query over a set of resources, in this case pods.
                                    If it's null, this PodAffinityTerm matches with no Pods.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      items:
                                        description: |-
                                          A label selector requirement is a selector that contains values, a key, and an operator that
                                          relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: |-
                                              operator represents a key's relationship to a set of values.
                                              Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: |-
                                              values is an array of string values. If the operator is In or NotIn,
                                              the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This array is replaced during a strategic
                                              merge patch.
                                            items:
                                              type: string
                                            type: array
                                            x-kubernetes-list-type: atomic
                                        required:
                                        - key
                                        - operator
                                        type: object
                                      type: array
                                      x-kubernetes-list-type: atomic
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: |-
                                        matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions, whose key field is "key", the
                                        operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                  x-kubernetes-map-type: atomic
                                matchLabelKeys:
                                  description: |-
                                    MatchLabelKeys is a set of pod label keys to select which pods will
                                    be taken into consideration. The keys are used to lookup values from the
                                    incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key in (value)` + "`" + `
                                    to select the group of existing pods which pods will be taken into consideration
                                    for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                    pod labels will be ignored. The default value is empty.
                                    The same key is forbidden to exist in both matchLabelKeys and labelSelector.
                                    Also, matchLabelKeys cannot be set when labelSelector isn't set.
                                    This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                mismatchLabelKeys:
                                  description: |-
                                    MismatchLabelKeys is a set of pod label keys to select which pods will
                                    be taken into consideration. The keys are used to lookup values from the
                                    incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key notin (value)` + "`" + `
                                    to select the group of existing pods which pods will be taken into consideration
                                    for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                    pod labels will be ignored. The default value is empty.
                                    The same key is forbidden to exist in both mismatchLabelKeys and labelSelector.
                                    Also, mismatchLabelKeys cannot be set when labelSelector isn't set.
                                    This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                namespaceSelector:
                                  description: |-
                                    A label query over the set of namespaces that the term applies to.
                                    The term is applied to the union of the namespaces selected by this field
                                    and the ones listed in the namespaces field.
                                    null selector and null or empty namespaces list means "this pod's namespace".
                                    An empty selector ({}) matches all namespaces.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      items:
                                        description: |-
                                          A label selector requirement is a selector that contains values, a key, and an operator that
                                          relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: |-
                                              operator represents a key's relationship to a set of values.
                                              Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: |-
                                              values is an array of string values. If the operator is In or NotIn,
                                              the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This array is replaced during a strategic
                                              merge patch.
                                            items:
                                              type: string
                                            type: array
                                            x-kubernetes-list-type: atomic
                                        required:
                                        - key
                                        - operator
                                        type: object
                                      type: array
                                      x-kubernetes-list-type: atomic
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: |-
                                        matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions, whose key field is "key", the
                                        operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                  x-kubernetes-map-type: atomic
                                namespaces:
                                  description: |-
                                    namespaces specifies a static list of namespace names that the term applies to.
                                    The term is applied to the union of the namespaces listed in this field
                                    and the ones selected by namespaceSelector.
                                    null or empty namespaces list and null namespaceSelector means "this pod's namespace".
                                  items:
                                    type: string
                                  type: array
                                  x-kubernetes-list-type: atomic
                                topologyKey:
                                  description: |-
                                    This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
                                    the labelSelector in the specified namespaces, where co-located is defined as running on a node
                                    whose value of the label with key topologyKey matches that of any node on which any of the
                                    selected pods is running.
                                    Empty topologyKey is not allowed.
                                  type: string
                              required:
                              - topologyKey
                              type: object
                            weight:
                              description: |-
                                weight associated with matching the corresponding podAffinityTerm,
                                in the range 1-100.
                              format: int32
                              type: integer
                          required:
                          - podAffinityTerm
                          - weight
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: |-
                          If the anti-affinity requirements specified by this field are not met at
                          scheduling time, the pod will not be scheduled onto the node.
                          If the anti-affinity requirements specified by this field cease to be met
                          at some point during pod execution (e.g. due to a pod label update), the
                          system may or may not try to eventually evict the pod from its node.
                          When there are multiple elements, the lists of nodes corresponding to each
                          podAffinityTerm are intersected, i.e. all terms must be satisfied.
                        items:
                          description: |-
                            Defines a set of pods (namely those matching the labelSelector
                            relative to the given namespace(s)) that this pod should be
                            co-located (affinity) or not co-located (anti-affinity) with,
                            where co-located is defined as running on a node whose value of
                            the label with key <topologyKey> matches that of any node on which
                            a pod of the set of pods is running
                          properties:
                            labelSelector:
                              description: |-
                                A label query over a set of resources, in this case pods.
                                If it's null, this PodAffinityTerm matches with no Pods.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  items:
                                    description: |-
                                      A label selector requirement is a selector that contains values, a key, and an operator that
                                      relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          operator represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: |-
                                          values is an array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. This array is replaced during a strategic
                                          merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: |-
                                    matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions, whose key field is "key", the
                                    operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                              x-kubernetes-map-type: atomic
                            matchLabelKeys:
                              description: |-
                                MatchLabelKeys is a set of pod label keys to select which pods will
                                be taken into consideration. The keys are used to lookup values from the
                                incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key in (value)` + "`" + `
                                to select the group of existing pods which pods will be taken into consideration
                                for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                pod labels will be ignored. The default value is empty.
                                The same key is forbidden to exist in both matchLabelKeys and labelSelector.
                                Also, matchLabelKeys cannot be set when labelSelector isn't set.
                                This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            mismatchLabelKeys:
                              description: |-
                                MismatchLabelKeys is a set of pod label keys to select which pods will
                                be taken into consideration. The keys are used to lookup values from the
                                incoming pod labels, those key-value labels are merged with ` + "`" + `labelSelector` + "`" + ` as ` + "`" + `key notin (value)` + "`" + `
                                to select the group of existing pods which pods will be taken into consideration
                                for the incoming pod's pod (anti) affinity. Keys that don't exist in the incoming
                                pod labels will be ignored. The default value is empty.
                                The same key is forbidden to exist in both mismatchLabelKeys and labelSelector.
                                Also, mismatchLabelKeys cannot be set when labelSelector isn't set.
                                This is a beta field and requires enabling MatchLabelKeysInPodAffinity feature gate (enabled by default).
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            namespaceSelector:
                              description: |-
                                A label query over the set of namespaces that the term applies to.
                                The term is applied to the union of the namespaces selected by this field
                                and the ones listed in the namespaces field.
                                null selector and null or empty namespaces list means "this pod's namespace".
                                An empty selector ({}) matches all namespaces.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  items:
                                    description: |-
                                      A label selector requirement is a selector that contains values, a key, and an operator that
                                      relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: |-
                                          operator represents a key's relationship to a set of values.
                                          Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: |-
                                          values is an array of string values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the operator is Exists or DoesNotExist,
                                          the values array must be empty. This array is replaced during a strategic
                                          merge patch.
                                        items:
                                          type: string
                                        type: array
                                        x-kubernetes-list-type: atomic
                                    required:
                                    - key
                                    - operator
                                    type: object
                                  type: array
                                  x-kubernetes-list-type: atomic
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: |-
                                    matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions, whose key field is "key", the
                                    operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                              x-kubernetes-map-type: atomic
                            namespaces:
                              description: |-
                                namespaces specifies a static list of namespace names that the term applies to.
                                The term is applied to the union of the namespaces listed in this field
                                and the ones selected by namespaceSelector.
                                null or empty namespaces list and null namespaceSelector means "this pod's namespace".
                              items:
                                type: string
                              type: array
                              x-kubernetes-list-type: atomic
                            topologyKey:
                              description: |-
                                This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where co-located is defined as running on a node
                                whose value of the label with key topologyKey matches that of any node on which any of the
                                selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
                          required:
                          - topologyKey
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                    type: object
                type: object
              annotations:
                additionalProperties:
                  additionalProperties:
                    type: string
                  description: Annotations are annotation for a given daemon
                  type: object
                description: The annotations-related configuration to add/set on each
                  Pod related object.
                nullable: true
                type: object
                x-kubernetes-preserve-unknown-fields: true
              autoscaler:
                description: Configuration related to autoscaling
                properties:
                  autoscalerType:
                    description: Type of autoscaling (optional) for noobaa-endpoint,
                      hpav2(default) and keda - Prometheus metrics based
                    enum:
                    - hpav2
                    - keda
                    type: string
                  prometheusNamespace:
                    description: Prometheus namespace that scrap metrics from noobaa
                    type: string
                type: object
              bucketLogging:
                description: BucketLogging sets the configuration for bucket logging
                properties:
                  bucketLoggingPVC:
                    description: |-
                      BucketLoggingPVC (optional) specifies the name of the Persistent Volume Claim (PVC) to be used
                      for guaranteed logging when the logging type is set to 'guaranteed'. The PVC must support
                      ReadWriteMany (RWX) access mode to ensure reliable logging.
                      For ODF: If not provided, the default CephFS storage class will be used to create the PVC.
                    type: string
                  loggingType:
                    description: |-
                      LoggingType specifies the type of logging for the bucket
                      There are two types available: best-effort and guaranteed logging
                      - best-effort(default) - less immune to failures but with better performance
                      - guaranteed - much more reliable but need to provide a storage class that supports RWX PVs
                    type: string
                type: object
              bucketNotifications:
                description: BucketNotifications (optional) controls bucket notification
                  options
                properties:
                  connections:
                    description: |-
                      Connections - A list of secrets' names that are used by the notifications configrations
                      (in the TopicArn field).
                    items:
                      description: |-
                        SecretReference represents a Secret Reference. It has enough information to retrieve secret
                        in any namespace
                      properties:
                        name:
                          description: name is unique within a namespace to reference
                            a secret resource.
                          type: string
                        namespace:
                          description: namespace defines the space within which the
                            secret name must be unique.
                          type: string
                      type: object
                      x-kubernetes-map-type: atomic
                    type: array
                  enabled:
                    description: Enabled - whether bucket notifications is enabled
                    type: boolean
                  pvc:
                    description: |-
                      PVC (optional) specifies the name of the Persistent Volume Claim (PVC) to be used
                      for holding pending notifications files.
                      For ODF - If not provided, the default CepthFS storage class will be used to create the PVC.
                    type: string
                required:
                - enabled
                type: object
              cleanupPolicy:
                description: CleanupPolicy (optional) Indicates user's policy for
                  deletion
                properties:
                  allowNoobaaDeletion:
                    type: boolean
                  confirmation:
                    description: CleanupConfirmationProperty is a string that specifies
                      cleanup confirmation
                    type: string
                type: object
              coreResources:
                description: CoreResources (optional) overrides the default resource
                  requirements for the server container
                properties:
                  claims:
                    description: |-
                      Claims lists the names of resources, defined in spec.resourceClaims,
                      that are used by this container.

                      This is an alpha field and requires enabling the
                      DynamicResourceAllocation feature gate.

                      This field is immutable. It can only be set for containers.
                    items:
                      description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                      properties:
                        name:
                          description: |-
                            Name must match the name of one entry in pod.spec.resourceClaims of
                            the Pod where this field is used. It makes that resource available
                            inside a container.
                          type: string
                        request:
                          description: |-
                            Request is the name chosen for a request in the referenced claim.
                            If empty, everything from the claim is made available, otherwise
                            only the result of this request.
                          type: string
                      required:
                      - name
                      type: object
                    type: array
                    x-kubernetes-list-map-keys:
                    - name
                    x-kubernetes-list-type: map
                  limits:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Limits describes the maximum amount of compute resources allowed.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                  requests:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Requests describes the minimum amount of compute resources required.
                      If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                      otherwise to an implementation-defined value. Requests cannot exceed Limits.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                type: object
              dbConf:
                description: DBConf (optional) overrides the default postgresql db
                  config
                type: string
              dbImage:
                description: DBImage (optional) overrides the default image for the
                  db container
                type: string
              dbResources:
                description: DBResources (optional) overrides the default resource
                  requirements for the db container
                properties:
                  claims:
                    description: |-
                      Claims lists the names of resources, defined in spec.resourceClaims,
                      that are used by this container.

                      This is an alpha field and requires enabling the
                      DynamicResourceAllocation feature gate.

                      This field is immutable. It can only be set for containers.
                    items:
                      description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                      properties:
                        name:
                          description: |-
                            Name must match the name of one entry in pod.spec.resourceClaims of
                            the Pod where this field is used. It makes that resource available
                            inside a container.
                          type: string
                        request:
                          description: |-
                            Request is the name chosen for a request in the referenced claim.
                            If empty, everything from the claim is made available, otherwise
                            only the result of this request.
                          type: string
                      required:
                      - name
                      type: object
                    type: array
                    x-kubernetes-list-map-keys:
                    - name
                    x-kubernetes-list-type: map
                  limits:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Limits describes the maximum amount of compute resources allowed.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                  requests:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Requests describes the minimum amount of compute resources required.
                      If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                      otherwise to an implementation-defined value. Requests cannot exceed Limits.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                type: object
              dbStorageClass:
                description: |-
                  DBStorageClass (optional) overrides the default cluster StorageClass for the database volume.
                  For the time being this field is immutable and can only be set on system creation.
                  This affects where the system stores its database which contains system config,
                  buckets, objects meta-data and mapping file parts to storage locations.
                type: string
              dbType:
                description: |-
                  DBType (optional) overrides the default type image for the db container.
                  The only possible value is postgres
                enum:
                - postgres
                type: string
              dbVolumeResources:
                description: |-
                  DBVolumeResources (optional) overrides the default PVC resource requirements for the database volume.
                  For the time being this field is immutable and can only be set on system creation.
                  This is because volume size updates are only supported for increasing the size,
                  and only if the storage class specifies ` + "`" + `allowVolumeExpansion: true` + "`" + `,
                properties:
                  limits:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Limits describes the maximum amount of compute resources allowed.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                  requests:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Requests describes the minimum amount of compute resources required.
                      If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                      otherwise to an implementation-defined value. Requests cannot exceed Limits.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                type: object
              debugLevel:
                description: DebugLevel (optional) sets the debug level
                enum:
                - all
                - nsfs
                - warn
                - default_level
                type: integer
              defaultBackingStoreSpec:
                description: 'Deprecated: DefaultBackingStoreSpec is not supported
                  anymore, use ManualDefaultBackingStore instead.'
                properties:
                  awsS3:
                    description: AWSS3Spec specifies a backing store of type aws-s3
                    properties:
                      awsSTSRoleARN:
                        description: AWSSTSRoleARN allows to Assume Role and use AssumeRoleWithWebIdentity
                        type: string
                      region:
                        description: Region is the AWS region
                        type: string
                      secret:
                        description: |-
                          Secret refers to a secret that provides the credentials
                          The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      sslDisabled:
                        description: SSLDisabled allows to disable SSL and use plain
                          http
                        type: boolean
                      targetBucket:
                        description: TargetBucket is the name of the target S3 bucket
                        type: string
                    required:
                    - targetBucket
                    type: object
                  azureBlob:
                    description: AzureBlob specifies a backing store of type azure-blob
                    properties:
                      secret:
                        description: |-
                          Secret refers to a secret that provides the credentials
                          The secret should define AccountName and AccountKey as provided by Azure Blob.
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      targetBlobContainer:
                        description: TargetBlobContainer is the name of the target
                          Azure Blob container
                        type: string
                    required:
                    - secret
                    - targetBlobContainer
                    type: object
                  googleCloudStorage:
                    description: GoogleCloudStorage specifies a backing store of type
                      google-cloud-storage
                    properties:
                      secret:
                        description: |-
                          Secret refers to a secret that provides the credentials
                          The secret should define GoogleServiceAccountPrivateKeyJson containing the entire json string as provided by Google.
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      targetBucket:
                        description: TargetBucket is the name of the target S3 bucket
                        type: string
                    required:
                    - secret
                    - targetBucket
                    type: object
                  ibmCos:
                    description: IBMCos specifies a backing store of type ibm-cos
                    properties:
                      endpoint:
                        description: 'Endpoint is the IBM COS compatible endpoint:
                          http(s)://host:port'
                        type: string
                      secret:
                        description: |-
                          Secret refers to a secret that provides the credentials
                          The secret should define IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      signatureVersion:
                        description: SignatureVersion specifies the client signature
                          version to use when signing requests.
                        type: string
                      targetBucket:
                        description: TargetBucket is the name of the target IBM COS
                          bucket
                        type: string
                    required:
                    - endpoint
                    - secret
                    - targetBucket
                    type: object
                  pvPool:
                    description: PVPool specifies a backing store of type pv-pool
                    properties:
                      numVolumes:
                        description: NumVolumes is the number of volumes to allocate
                        type: integer
                      resources:
                        description: VolumeResources represents the minimum resources
                          each volume should have.
                        properties:
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: |-
                              Limits describes the maximum amount of compute resources allowed.
                              More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: |-
                              Requests describes the minimum amount of compute resources required.
                              If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                              otherwise to an implementation-defined value. Requests cannot exceed Limits.
                              More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                            type: object
                        type: object
                      secret:
                        description: |-
                          Secret refers to a secret that provides the agent configuration
                          The secret should define AGENT_CONFIG containing agent_configuration from noobaa-core.
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      storageClass:
                        description: StorageClass is the name of the storage class
                          to use for the PV's
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
                        description: |-
                          Secret refers to a secret that provides the credentials
                          The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                      signatureVersion:
                        description: SignatureVersion specifies the client signature
                          version to use when signing requests.
                        type: string
                      targetBucket:
                        description: TargetBucket is the name of the target S3 bucket
                        type: string
                    required:
                    - endpoint
                    - secret
                    - targetBucket
                    type: object
                  type:
                    description: Type is an enum of supported types
                    type: string
                required:
                - type
                type: object
              denyHTTP:
                description: DenyHTTP (optional) if given will deny access to the
                  NooBaa S3 service using HTTP (only HTTPS)
                type: boolean
              disableLoadBalancerService:
                description: DisableLoadBalancerService (optional) sets the service
                  type to ClusterIP instead of LoadBalancer
                nullable: true
                type: boolean
              endpoints:
                description: |-
                  Endpoints (optional) sets configuration info for the noobaa endpoint
                  deployment.
                properties:
                  additionalVirtualHosts:
                    description: |-
                      AdditionalVirtualHosts (optional) provide a list of additional hostnames
                      (on top of the builtin names defined by the cluster: service name, elb name, route name)
                      to be used as virtual hosts by the the endpoints in the endpoint deployment
                    items:
                      type: string
                    type: array
                  maxCount:
                    description: |-
                      MaxCount, the number of endpoint instances (pods)
                      to be used as the upper bound when autoscaling
                    format: int32
                    type: integer
                  minCount:
                    description: |-
                      MinCount, the number of endpoint instances (pods)
                      to be used as the lower bound when autoscaling
                    format: int32
                    type: integer
                  resources:
                    description: Resources (optional) overrides the default resource
                      requirements for every endpoint pod
                    properties:
                      claims:
                        description: |-
                          Claims lists the names of resources, defined in spec.resourceClaims,
                          that are used by this container.

                          This is an alpha field and requires enabling the
                          DynamicResourceAllocation feature gate.

                          This field is immutable. It can only be set for containers.
                        items:
                          description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                          properties:
                            name:
                              description: |-
                                Name must match the name of one entry in pod.spec.resourceClaims of
                                the Pod where this field is used. It makes that resource available
                                inside a container.
                              type: string
                            request:
                              description: |-
                                Request is the name chosen for a request in the referenced claim.
                                If empty, everything from the claim is made available, otherwise
                                only the result of this request.
                              type: string
                          required:
                          - name
                          type: object
                        type: array
                        x-kubernetes-list-map-keys:
                        - name
                        x-kubernetes-list-type: map
                      limits:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: |-
                          Limits describes the maximum amount of compute resources allowed.
                          More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                        type: object
                      requests:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: |-
                          Requests describes the minimum amount of compute resources required.
                          If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                          otherwise to an implementation-defined value. Requests cannot exceed Limits.
                          More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                        type: object
                    type: object
                type: object
              externalPgSSLRequired:
                description: ExternalPgSSLRequired (optional) holds an optional boolean
                  to force ssl connections to the external Postgres DB
                type: boolean
              externalPgSSLSecret:
                description: ExternalPgSSLSecret (optional) holds an optional secret
                  with client key and cert used for connecting to external Postgres
                  DB
                properties:
                  name:
                    description: name is unique within a namespace to reference a
                      secret resource.
                    type: string
                  namespace:
                    description: namespace defines the space within which the secret
                      name must be unique.
                    type: string
                type: object
                x-kubernetes-map-type: atomic
              externalPgSSLUnauthorized:
                description: ExternalPgSSLUnauthorized (optional) holds an optional
                  boolean to allow unauthorized connections to external Postgres DB
                type: boolean
              externalPgSecret:
                description: ExternalPgSecret (optional) holds an optional secret
                  with a url to an extrenal Postgres DB to be used
                properties:
                  name:
                    description: name is unique within a namespace to reference a
                      secret resource.
                    type: string
                  namespace:
                    description: namespace defines the space within which the secret
                      name must be unique.
                    type: string
                type: object
                x-kubernetes-map-type: atomic
              image:
                description: Image (optional) overrides the default image for the
                  server container
                type: string
              imagePullSecret:
                description: ImagePullSecret (optional) sets a pull secret for the
                  system image
                properties:
                  name:
                    default: ""
                    description: |-
                      Name of the referent.
                      This field is effectively required, but due to backwards compatibility is
                      allowed to be empty. Instances of this type with an empty value here are
                      almost certainly wrong.
                      More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                    type: string
                type: object
                x-kubernetes-map-type: atomic
              joinSecret:
                description: |-
                  JoinSecret (optional) instructs the operator to join another cluster
                  and point to a secret that holds the join information
                properties:
                  name:
                    description: name is unique within a namespace to reference a
                      secret resource.
                    type: string
                  namespace:
                    description: namespace defines the space within which the secret
                      name must be unique.
                    type: string
                type: object
                x-kubernetes-map-type: atomic
              labels:
                additionalProperties:
                  additionalProperties:
                    type: string
                  description: Labels are label for a given daemon
                  type: object
                description: The labels-related configuration to add/set on each Pod
                  related object.
                nullable: true
                type: object
                x-kubernetes-preserve-unknown-fields: true
              loadBalancerSourceSubnets:
                description: |-
                  LoadBalancerSourceSubnets (optional) if given will allow access to the NooBaa services
                  only from the listed subnets. This field will have no effect if DisableLoadBalancerService is set
                  to true
                properties:
                  s3:
                    description: S3 is a list of subnets that will be allowed to access
                      the Noobaa S3 service
                    items:
                      type: string
                    type: array
                  sts:
                    description: STS is a list of subnets that will be allowed to
                      access the Noobaa STS service
                    items:
                      type: string
                    type: array
                type: object
              logResources:
                description: LogResources (optional) overrides the default resource
                  requirements for the noobaa-log-processor container
                properties:
                  claims:
                    description: |-
                      Claims lists the names of resources, defined in spec.resourceClaims,
                      that are used by this container.

                      This is an alpha field and requires enabling the
                      DynamicResourceAllocation feature gate.

                      This field is immutable. It can only be set for containers.
                    items:
                      description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                      properties:
                        name:
                          description: |-
                            Name must match the name of one entry in pod.spec.resourceClaims of
                            the Pod where this field is used. It makes that resource available
                            inside a container.
                          type: string
                        request:
                          description: |-
                            Request is the name chosen for a request in the referenced claim.
                            If empty, everything from the claim is made available, otherwise
                            only the result of this request.
                          type: string
                      required:
                      - name
                      type: object
                    type: array
                    x-kubernetes-list-map-keys:
                    - name
                    x-kubernetes-list-type: map
                  limits:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Limits describes the maximum amount of compute resources allowed.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                  requests:
                    additionalProperties:
                      anyOf:
                      - type: integer
                      - type: string
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    description: |-
                      Requests describes the minimum amount of compute resources required.
                      If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
                      otherwise to an implementation-defined value. Requests cannot exceed Limits.
                      More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
                    type: object
                type: object
              manualDefaultBackingStore:
                description: |-
                  ManualDefaultBackingStore (optional - default value is false) if true the default backingstore/namespacestore
                  will not be reconciled by the operator and it should be manually handled by the user. It will allow the
                  user to  delete DefaultBackingStore/DefaultNamespaceStore, user needs to delete associated buckets and
                  update the admin account with new BackingStore/NamespaceStore in order to delete the DefaultBackingStore/DefaultNamespaceStore
                nullable: true
                type: boolean
              pvPoolDefaultStorageClass:
                description: |-
                  PVPoolDefaultStorageClass (optional) overrides the default cluster StorageClass for the pv-pool volumes.
                  This affects where the system stores data chunks (encrypted).
                  Updates to this field will only affect new pv-pools,
                  but updates to existing pools are not supported by the operator.
                type: string
              region:
                description: |-
                  Region (optional) provide a region for the location info
                  of the endpoints in the endpoint deployment
                type: string
              security:
                description: Security represents security settings
                properties:
                  kms:
                    description: KeyManagementServiceSpec represent various details
                      of the KMS server
                    properties:
                      connectionDetails:
                        additionalProperties:
                          type: string
                        type: object
                      enableKeyRotation:
                        type: boolean
                      schedule:
                        type: string
                      tokenSecretName:
                        type: string
                    type: object
                type: object
              tolerations:
                description: Tolerations (optional) passed through to noobaa's pods
                items:
                  description: |-
                    The pod this Toleration is attached to tolerates any taint that matches
                    the triple <key,value,effect> using the matching operator <operator>.
                  properties:
                    effect:
                      description: |-
                        Effect indicates the taint effect to match. Empty means match all taint effects.
                        When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                      type: string
                    key:
                      description: |-
                        Key is the taint key that the toleration applies to. Empty means match all taint keys.
                        If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                      type: string
                    operator:
                      description: |-
                        Operator represents a key's relationship to the value.
                        Valid operators are Exists and Equal. Defaults to Equal.
                        Exists is equivalent to wildcard for value, so that a pod can
                        tolerate all taints of a particular category.
                      type: string
                    tolerationSeconds:
                      description: |-
                        TolerationSeconds represents the period of time the toleration (which must be
                        of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default,
                        it is not set, which means tolerate the taint forever (do not evict). Zero and
                        negative values will be treated as 0 (evict immediately) by the system.
                      format: int64
                      type: integer
                    value:
                      description: |-
                        Value is the taint value the toleration matches to.
                        If the operator is Exists, the value should be empty, otherwise just a regular string.
                      type: string
                  type: object
                type: array
            type: object
          status:
            description: Most recently observed status of the noobaa system.
            properties:
              accounts:
                description: Accounts reports accounts info for the admin account
                properties:
                  admin:
                    description: UserStatus is the status info of a user secret
                    properties:
                      secretRef:
                        description: |-
                          SecretReference represents a Secret Reference. It has enough information to retrieve secret
                          in any namespace
                        properties:
                          name:
                            description: name is unique within a namespace to reference
                              a secret resource.
                            type: string
                          namespace:
                            description: namespace defines the space within which
                              the secret name must be unique.
                            type: string
                        type: object
                        x-kubernetes-map-type: atomic
                    required:
                    - secretRef
                    type: object
                required:
                - admin
                type: object
              actualImage:
                description: ActualImage is set to report which image the operator
                  is using
                type: string
              beforeUpgradeDbImage:
                description: BeforeUpgradeDbImage is the db image used before last
                  db upgrade
                type: string
              conditions:
                description: Conditions is a list of conditions related to operator
                  reconciliation
                items:
                  description: |-
                    Condition represents the state of the operator's
                    reconciliation functionality.
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
                      description: ConditionType is the state of the operator's reconciliation
                        functionality.
                      type: string
                  required:
                  - status
                  - type
                  type: object
                type: array
              endpoints:
                description: |-
                  Endpoints reports the actual number of endpoints in the endpoint deployment
                  and the virtual hosts list used recognized by the endpoints
                properties:
                  readyCount:
                    format: int32
                    type: integer
                  virtualHosts:
                    items:
                      type: string
                    type: array
                required:
                - readyCount
                - virtualHosts
                type: object
              lastKeyRotateTime:
                description: LastKeyRotateTime is the time system ran an encryption
                  key rotate
                format: date-time
                type: string
              observedGeneration:
                description: |-
                  ObservedGeneration is the most recent generation observed for this noobaa system.
                  It corresponds to the CR generation, which is updated on mutation by the API Server.
                format: int64
                type: integer
              phase:
                description: Phase is a simple, high-level summary of where the System
                  is in its lifecycle
                type: string
              postgresUpdatePhase:
                description: Upgrade reports the status of the ongoing postgres upgrade
                  process
                type: string
              readme:
                description: Readme is a user readable string with explanations on
                  the system
                type: string
              relatedObjects:
                description: RelatedObjects is a list of objects related to this operator.
                items:
                  description: ObjectReference contains enough information to let
                    you inspect or modify the referred object.
                  properties:
                    apiVersion:
                      description: API version of the referent.
                      type: string
                    fieldPath:
                      description: |-
                        If referring to a piece of an object instead of an entire object, this string
                        should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
                        For example, if the object reference is to a container within a pod, this would take on a value like:
                        "spec.containers{name}" (where "name" refers to the name of the container that triggered
                        the event) or if no container name is specified "spec.containers[2]" (container with
                        index 2 in this pod). This syntax is chosen only to have some well-defined way of
                        referencing a part of an object.
                      type: string
                    kind:
                      description: |-
                        Kind of the referent.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
                      type: string
                    name:
                      description: |-
                        Name of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                      type: string
                    namespace:
                      description: |-
                        Namespace of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
                      type: string
                    resourceVersion:
                      description: |-
                        Specific resourceVersion to which this reference is made, if any.
                        More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
                      type: string
                    uid:
                      description: |-
                        UID of the referent.
                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids
                      type: string
                  type: object
                  x-kubernetes-map-type: atomic
                type: array
              services:
                description: Services reports addresses for the services
                properties:
                  serviceMgmt:
                    description: ServiceStatus is the status info and network addresses
                      of a service
                    properties:
                      externalDNS:
                        description: ExternalDNS are external public addresses for
                          the service
                        items:
                          type: string
                        type: array
                      externalIP:
                        description: |-
                          ExternalIP are external public addresses for the service
                          LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
                          IngressPorts are manually created public addresses for the service
                          https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
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
                        description: |-
                          InternalIP are internal addresses of the service inside the cluster
                          https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                        items:
                          type: string
                        type: array
                      nodePorts:
                        description: |-
                          NodePorts are the most basic network available.
                          NodePorts use the networks available on the hosts of kubernetes nodes.
                          This generally works from within a pod, and from the internal
                          network of the nodes, but may fail from public network.
                          https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                        items:
                          type: string
                        type: array
                      podPorts:
                        description: |-
                          PodPorts are the second most basic network address.
                          Every pod has an IP in the cluster and the pods network is a mesh
                          so the operator running inside a pod in the cluster can use this address.
                          Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
                          Note2: when running the operator outside of the cluster, pod IP is not accessible.
                        items:
                          type: string
                        type: array
                    type: object
                  serviceS3:
                    description: ServiceStatus is the status info and network addresses
                      of a service
                    properties:
                      externalDNS:
                        description: ExternalDNS are external public addresses for
                          the service
                        items:
                          type: string
                        type: array
                      externalIP:
                        description: |-
                          ExternalIP are external public addresses for the service
                          LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
                          IngressPorts are manually created public addresses for the service
                          https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
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
                        description: |-
                          InternalIP are internal addresses of the service inside the cluster
                          https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                        items:
                          type: string
                        type: array
                      nodePorts:
                        description: |-
                          NodePorts are the most basic network available.
                          NodePorts use the networks available on the hosts of kubernetes nodes.
                          This generally works from within a pod, and from the internal
                          network of the nodes, but may fail from public network.
                          https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                        items:
                          type: string
                        type: array
                      podPorts:
                        description: |-
                          PodPorts are the second most basic network address.
                          Every pod has an IP in the cluster and the pods network is a mesh
                          so the operator running inside a pod in the cluster can use this address.
                          Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
                          Note2: when running the operator outside of the cluster, pod IP is not accessible.
                        items:
                          type: string
                        type: array
                    type: object
                  serviceSts:
                    description: ServiceStatus is the status info and network addresses
                      of a service
                    properties:
                      externalDNS:
                        description: ExternalDNS are external public addresses for
                          the service
                        items:
                          type: string
                        type: array
                      externalIP:
                        description: |-
                          ExternalIP are external public addresses for the service
                          LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
                          IngressPorts are manually created public addresses for the service
                          https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
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
                        description: |-
                          InternalIP are internal addresses of the service inside the cluster
                          https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                        items:
                          type: string
                        type: array
                      nodePorts:
                        description: |-
                          NodePorts are the most basic network available.
                          NodePorts use the networks available on the hosts of kubernetes nodes.
                          This generally works from within a pod, and from the internal
                          network of the nodes, but may fail from public network.
                          https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                        items:
                          type: string
                        type: array
                      podPorts:
                        description: |-
                          PodPorts are the second most basic network address.
                          Every pod has an IP in the cluster and the pods network is a mesh
                          so the operator running inside a pod in the cluster can use this address.
                          Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
                          Note2: when running the operator outside of the cluster, pod IP is not accessible.
                        items:
                          type: string
                        type: array
                    type: object
                  serviceSyslog:
                    description: ServiceStatus is the status info and network addresses
                      of a service
                    properties:
                      externalDNS:
                        description: ExternalDNS are external public addresses for
                          the service
                        items:
                          type: string
                        type: array
                      externalIP:
                        description: |-
                          ExternalIP are external public addresses for the service
                          LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
                          IngressPorts are manually created public addresses for the service
                          https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
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
                        description: |-
                          InternalIP are internal addresses of the service inside the cluster
                          https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
                        items:
                          type: string
                        type: array
                      nodePorts:
                        description: |-
                          NodePorts are the most basic network available.
                          NodePorts use the networks available on the hosts of kubernetes nodes.
                          This generally works from within a pod, and from the internal
                          network of the nodes, but may fail from public network.
                          https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
                        items:
                          type: string
                        type: array
                      podPorts:
                        description: |-
                          PodPorts are the second most basic network address.
                          Every pod has an IP in the cluster and the pods network is a mesh
                          so the operator running inside a pod in the cluster can use this address.
                          Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
                          Note2: when running the operator outside of the cluster, pod IP is not accessible.
                        items:
                          type: string
                        type: array
                    type: object
                required:
                - serviceMgmt
                - serviceS3
                type: object
              upgradePhase:
                description: Upgrade reports the status of the ongoing upgrade process
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml = "84ca6f2a35a413e74a51375bd0ec31c33bb76a00de8e0ef8d02a7798e02ec460"

const File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: BackingStore
metadata:
  name: default
spec:
`

const Sha256_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml = "fc6047d603b8275240b1d2dc12efa32a977a83edfff4ab565e92c6523a5d8b70"

const File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: BucketClass
metadata:
  name: default
spec:
`

const Sha256_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml = "0938c22769bd9f2759d0ffd33b04a4650ec84dcd73508d9ef368f5908c1caec4"

const File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: NamespaceStore
metadata:
  name: default
spec:
`

const Sha256_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml = "498c2013757409432cfd98b21a5934bccf506f1af1b885241db327024aa450fd"

const File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
spec: {}
`

const Sha256_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml = "69085515e8d16eaa9f320a32f2881cbd93d232bfbb072eef8692896a86f7b6dd"

const File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml = `apiVersion: noobaa.io/v1alpha1
kind: NooBaaAccount
metadata:
  name: default
spec: {}
`

const Sha256_deploy_internal_admission_webhook_yaml = "6ac4c09a3923e2545fe484dbf68171d718669cf03e874889f44e005ed5f8529c"

const File_deploy_internal_admission_webhook_yaml = `apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: admission-validation-webhook
webhooks:
  - name: admissionwebhook.noobaa.io
    matchPolicy: Equivalent
    rules:
    - apiGroups:   ["noobaa.io"]
      apiVersions: ["v1alpha1"]
      operations:  
      - "CREATE" 
      - "UPDATE"
      - "DELETE"
      resources:   
      - "backingstores"
      - "namespacestores"
      scope: "Namespaced"
    - apiGroups:   ["noobaa.io"]
      apiVersions: ["v1alpha1"]
      operations:  
      - "CREATE" 
      resources:   
      - "bucketclasses"
      scope: "Namespaced"
    - apiGroups:   ["noobaa.io"]
      apiVersions: ["v1alpha1"]
      operations:  
      - "CREATE" 
      - "UPDATE"
      resources:   
      - "noobaaaccounts"
      scope: "Namespaced"
    - apiGroups:   ["noobaa.io"]
      apiVersions: ["v1alpha1"]
      operations:  
      - "DELETE"
      - "CREATE"
      - "UPDATE"
      resources:   
      - "noobaas"
      scope: "Namespaced"
    sideEffects: None
    clientConfig:
      service:
        name: admission-webhook-service
        namespace: placeholder
        path: "/validate"
      caBundle:
    admissionReviewVersions: ["v1", "v1beta1"]
    failurePolicy: Ignore
    timeoutSeconds: 5
`

const Sha256_deploy_internal_ceph_objectstore_user_yaml = "655f33a1e3053847a298294d67d7db647d26fd11d1df7e229af718a8308bbd8e"

const File_deploy_internal_ceph_objectstore_user_yaml = `apiVersion: ceph.rook.io/v1
kind: CephObjectStoreUser
metadata:
  name: CEPH_OBJ_USER_NAME
spec:
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

const Sha256_deploy_internal_cloud_creds_azure_cr_yaml = "e9a8455b8657869be6e8a107519f3d1cfab36a536c479d6688eef6981262946a"

const File_deploy_internal_cloud_creds_azure_cr_yaml = `apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: CRED-REQ-NAME
spec:
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AzureProviderSpec
    roleBindings:
      - role: "Storage Account Contributor"
      - role: "Storage Blob Data Contributor"
  secretRef:
    name: CRED-SECRET-NAME
    namespace: CRED-SECRET-NAMESPACE
`

const Sha256_deploy_internal_cloud_creds_gcp_cr_yaml = "f4415e851da03426e8c31a7cb5b904b4438d958a5297c70b967ca6c2881d360f"

const File_deploy_internal_cloud_creds_gcp_cr_yaml = `apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: CRED-REQ-NAME
  namespace: CRED-REQ-NAMESPACE
spec:
  secretRef:
    name: CRED-SECRET-NAME
    namespace: CRED-SECRET-NAMESPACE
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: GCPProviderSpec
    predefinedRoles:
    - roles/storage.admin
    skipServiceCheck: true
`

const Sha256_deploy_internal_configmap_ca_inject_yaml = "fac2305a04146c6b553398b1cb69b3ee2f32c5735359f5102590d43d33ccecba"

const File_deploy_internal_configmap_ca_inject_yaml = `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    config.openshift.io/inject-trusted-cabundle: "true"
  name: ocp-injected-ca-bundle
data: {}
`

const Sha256_deploy_internal_configmap_empty_yaml = "6405c531c6522ecd54808f5cb531c1001b9ad01a73917427c523a92be44f348f"

const File_deploy_internal_configmap_empty_yaml = `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: noobaa
data: {}
`

const Sha256_deploy_internal_configmap_postgres_db_yaml = "afe8a865abf2b033229df9dcea392abc1cb27df965d5ff0181f6d931504dce4e"

const File_deploy_internal_configmap_postgres_db_yaml = `apiVersion: v1
kind: ConfigMap
metadata:
  name: noobaa-postgres-config
  labels:
    app: noobaa
data:
  noobaa-postgres.conf: |
    # disable huge_pages trial
    # see https://bugzilla.redhat.com/show_bug.cgi?id=1946792
    huge_pages = off

    # postgres tuning
    max_connections = 600
    shared_buffers = 1GB
    effective_cache_size = 3GB
    maintenance_work_mem = 256MB
    checkpoint_completion_target = 0.9
    wal_buffers = 16MB
    default_statistics_target = 100
    random_page_cost = 1.1
    effective_io_concurrency = 300
    work_mem = 1747kB
    min_wal_size = 2GB
    max_wal_size = 8GB
    shared_preload_libraries = 'pg_stat_statements'
`

const Sha256_deploy_internal_deployment_endpoint_yaml = "21b206c9119e37c4ebba84d5c1e2b1d45b06c716b4def69db9ba9268ef75e1e1"

const File_deploy_internal_deployment_endpoint_yaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: noobaa
  name: noobaa-endpoint
spec:
  replicas: 1
  selector:
    matchLabels:
      noobaa-s3: noobaa
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        noobaa-s3: noobaa
        app: noobaa
      annotations:
        noobaa.io/configmap-hash: ""
    spec:
        # Notice that changing the serviceAccountName would need to update existing AWS STS role trust policy for customers
      serviceAccountName: noobaa-endpoint
      volumes:
        - name: mgmt-secret
          secret:
            secretName: noobaa-mgmt-serving-cert
            optional: true
        - name: s3-secret
          secret:
            secretName: noobaa-s3-serving-cert
            optional: true
        - name: sts-secret
          secret:
            secretName: noobaa-sts-serving-cert
            optional: true
        # This service account token can be used to provide identity outside the cluster.
        # For example, this token can be used with AssumeRoleWithWebIdentity to authenticate with AWS using IAM OIDC provider and STS.
        - name: bound-sa-token
          projected:
            sources:
              - serviceAccountToken:
                  path: token
                  # For testing purposes change the audience to api
                  audience: openshift
        - name: noobaa-auth-token
          secret:
            secretName: noobaa-endpoints
            optional: true
        - name: noobaa-server
          secret:
            secretName: noobaa-server
            optional: true
      containers:
        - name: endpoint
          image: NOOBAA_CORE_IMAGE
          command:
            - /noobaa_init_files/noobaa_init.sh
            - init_endpoint
          resources:
            requests:
              cpu: "999m"
              memory: "2Gi"
            limits:
              cpu: "999m"
              memory: "2Gi"
          securityContext:
            fsGroupChangePolicy: "OnRootMismatch"
            seLinuxOptions:
              type: "spc_t"
            capabilities:
              add: ["SETUID", "SETGID"]
          ports:
            - containerPort: 6001
            - containerPort: 6443
            - containerPort: 7443
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: NOOBAA_DISABLE_COMPRESSION
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_DISABLE_COMPRESSION
            - name: NOOBAA_LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_LOG_LEVEL
            - name: NOOBAA_LOG_COLOR
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_LOG_COLOR
            - name: MGMT_ADDR
            - name: SYSLOG_ADDR
            - name: BG_ADDR
            - name: MD_ADDR
            - name: HOSTED_AGENTS_ADDR
            - name: DB_TYPE
            - name: POSTGRES_HOST
            - name: POSTGRES_PORT
            - name: POSTGRES_DBNAME
              value: nbcore
            - name: POSTGRES_USER
            - name: POSTGRES_PASSWORD
            - name: POSTGRES_CONNECTION_STRING
            - name: POSTGRES_SSL_REQUIRED
            - name: POSTGRES_SSL_UNAUTHORIZED
            - name: VIRTUAL_HOSTS
            - name: REGION
            - name: ENDPOINT_GROUP_ID
            - name: LOCAL_MD_SERVER
            - name: LOCAL_N2N_AGENT
            - name: NOOBAA_ROOT_SECRET
            - name: NODE_EXTRA_CA_CERTS
            - name: GUARANTEED_LOGS_PATH
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
          volumeMounts:
            # curently ssl_utils expects both secrets to be configured in order to use
            # certificates. TODO: Allow each secret to be configured by intself.
            - name: mgmt-secret
              mountPath: /etc/mgmt-secret
              readOnly: true
            - name: s3-secret
              mountPath: /etc/s3-secret
              readOnly: true
            - name: noobaa-auth-token
              mountPath: /etc/noobaa-auth-token
              readOnly: true
            - name: noobaa-server
              mountPath: /etc/noobaa-server
              readOnly: true
            - name: sts-secret
              mountPath: /etc/sts-secret
              readOnly: true
            # used for aws sts endpoint type
            - name: bound-sa-token
              mountPath: /var/run/secrets/openshift/serviceaccount
              readOnly: true
          readinessProbe: # must be configured to support rolling updates
            tcpSocket:
              port: 6001 # ready when s3 port is open
            timeoutSeconds: 5
      securityContext:
        runAsUser: 0
        runAsGroup: 0
`

const Sha256_deploy_internal_hpa_keda_scaled_object_yaml = "c8137738713103bb55d7867051b2d5cce27c2f37501d5f0c5e48ae97506ac2bb"

const File_deploy_internal_hpa_keda_scaled_object_yaml = `apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  labels:
    deploymentName: noobaa-endpoint
    origin: keda
    app: noobaa
    scaledobject.keda.sh/name: prometheus-scaledobject
spec:
  cooldownPeriod: 150
  pollingInterval: 30
  scaleTargetRef:
    name: noobaa-endpoint
    apiVersion: apps/v1
    kind: Deployment
  triggers:
    - type: cpu
      metricType: Utilization
      metadata:
        value: "80"
`

const Sha256_deploy_internal_hpa_keda_secret_yaml = "bb193ad94aef5d5a4398c9d50f7fba4ae56a9db5da0080c1211d34f6fc88a682"

const File_deploy_internal_hpa_keda_secret_yaml = `apiVersion: v1
kind: Secret
metadata:
  name: prometheus-k8s-secret
  labels:
    origin: keda
    app: noobaa
type: Opaque
data: {}
`

const Sha256_deploy_internal_hpa_keda_trigger_authentication_yaml = "55e1a26af5761d17c32ae145623f78b639e44abc2364fdc0002269e343013d23"

const File_deploy_internal_hpa_keda_trigger_authentication_yaml = `apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: keda-prom-creds
  labels:
    origin: keda
spec:
  secretTargetRef: []
`

const Sha256_deploy_internal_hpav2_apiservice_yaml = "87f414bb02c6cf05700fdf2733fb253f970290633b658bfa2b2ea6799cf565c9"

const File_deploy_internal_hpav2_apiservice_yaml = `apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  labels:
    app: prometheus-adapter
  name: v1beta1.custom.metrics.k8s.io
spec:
  service:
    name: custom-metrics-prometheus-adapter
  group: custom.metrics.k8s.io
  version: v1beta1
  insecureSkipTLSVerify: true
  groupPriorityMinimum: 100
  versionPriority: 100
`

const Sha256_deploy_internal_hpav2_autoscaling_yaml = "5af69e55a40026f5a01d102232fbecb1ecbc2c5482f60b1226baf5fe2afc07e6"

const File_deploy_internal_hpav2_autoscaling_yaml = `kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2
metadata:
  labels:
    app: noobaa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: noobaa-endpoint
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
`

const Sha256_deploy_internal_hpav2_configmap_adapter_yaml = "8f857756f46511c8763fbc03e9373cb3eec11c2251d7a844ae4990d55208336b"

const File_deploy_internal_hpav2_configmap_adapter_yaml = `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: prometheus-adapter
  name: adapter-config
data:
  config.yaml: |
    rules:
    - seriesQuery: 'container_cpu_usage_seconds_total{namespace="placeholder",pod=~"noobaa-endpoint-.*"}'
      resources:
        overrides:
          namespace:
            resource: namespace
          pod:
            resource: pod
      name:
        matches: "^(.*)_total"
        as: "${1}_per_second"
      metricsQuery: (sum(irate(<<.Series>>{<<.LabelMatchers>>}[5m])) by (<<.GroupBy>>))
`

const Sha256_deploy_internal_hpav2_deployment_adapter_yaml = "097a81580b56da76caee3022d682d7eee1fd58acd46fed383039188026102429"

const File_deploy_internal_hpav2_deployment_adapter_yaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: metrics-adapter
    app.kubernetes.io/name: prometheus-adapter
    app.kubernetes.io/version: 0.10.0
    app: prometheus-adapter
  name: prometheus-adapter
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: metrics-adapter
      app.kubernetes.io/name: prometheus-adapter
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/component: metrics-adapter
        app.kubernetes.io/name: prometheus-adapter
        app.kubernetes.io/version: 0.10.0
    spec:
      automountServiceAccountToken: true
      containers:
      - args:
        - --v=6
        - --config=/etc/adapter/config.yaml
        - --logtostderr=true
        - --metrics-relist-interval=1m
        - --secure-port=6443
        - --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
        image: registry.k8s.io/prometheus-adapter/prometheus-adapter:v0.10.0
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /livez
            port: https
            scheme: HTTPS
          initialDelaySeconds: 30
          periodSeconds: 5
        name: prometheus-adapter
        ports:
        - containerPort: 6443
          name: https
        readinessProbe:
          failureThreshold: 5
          httpGet:
            path: /readyz
            port: https
            scheme: HTTPS
          initialDelaySeconds: 30
          periodSeconds: 5
        resources:
          requests:
            cpu: 102m
            memory: 180Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /tmp
          name: tmpfs
          readOnly: true
        - mountPath: /var/run/serving-cert
          name: volume-serving-cert
          readOnly: false
        - mountPath: /etc/adapter
          name: config
          readOnly: true
        - name: prometheus-adapter-prometheus-config
          mountPath: /etc/prometheus-config
        - name: serving-certs-ca-bundle
          mountPath: /etc/ssl/certs
          readOnly: true
        - mountPath: /var/run/empty/serving-cert
          name: volume-empty-serving-cert
          readOnly: false
      nodeSelector:
        kubernetes.io/os: linux
      securityContext: {}
      serviceAccountName: custom-metrics-prometheus-adapter
      volumes:
      - emptyDir: {}
        name: tmpfs
      - name: volume-serving-cert
        secret:
          secretName: prometheus-adapter-serving-cert
          defaultMode: 420
          optional: true
      - name: config
        configMap:
          name: adapter-config
          optional: true
      - name: prometheus-adapter-prometheus-config
        configMap:
          name: prometheus-adapter-prometheus-config
          optional: true
          defaultMode: 420
      - name: serving-certs-ca-bundle
        configMap:
          name: serving-certs-ca-bundle
          items:
            - key: service-ca.crt
              path: service-ca.crt
          optional: true
          defaultMode: 420
      - name: volume-empty-serving-cert
        emptyDir: {}
`

const Sha256_deploy_internal_hpav2_prometheus_auth_config_yaml = "191fea97f0e4552cca83902cff8fc7ff1c3d13f5d900499a6e5aac51a50c1a16"

const File_deploy_internal_hpav2_prometheus_auth_config_yaml = `kind: ConfigMap
apiVersion: v1
metadata:
  name: prometheus-adapter-prometheus-config
  labels:
    app: prometheus-adapter
data:
  prometheus-config.yaml: |
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority: /etc/ssl/certs/service-ca.crt
        server: prometheus-url-placeholder
      name: prometheus-k8s
    contexts:
    - context:
        cluster: prometheus-k8s
        user: prometheus-k8s
      name: prometheus-k8s
    current-context: prometheus-k8s
    kind: Config
    preferences: {}
    users:
    - name: prometheus-k8s
      user:
        tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
`

const Sha256_deploy_internal_hpav2_service_adapter_yaml = "0d08c1a3835ccf284f29fdf1b4081275847537c882f08feeccf1af9e58c5d513"

const File_deploy_internal_hpav2_service_adapter_yaml = `apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: metrics-adapter
    app.kubernetes.io/name: prometheus-adapter
    app.kubernetes.io/version: 0.10.0
    app: prometheus-adapter
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: 'prometheus-adapter-serving-cert'
    service.alpha.openshift.io/serving-cert-secret-name: 'prometheus-adapter-serving-cert'
  name: custom-metrics-prometheus-adapter
spec:
  ports:
  - name: https
    port: 443
    targetPort: 6443
  selector:
    app.kubernetes.io/component: metrics-adapter
    app.kubernetes.io/name: prometheus-adapter
`

const Sha256_deploy_internal_hpav2_serving_certs_ca_bundle_yaml = "be21595af7052a8191a131fb62fdcf21fd4dfd4d2435823f7a15f01d3efee8f3"

const File_deploy_internal_hpav2_serving_certs_ca_bundle_yaml = `kind: ConfigMap
apiVersion: v1
metadata:
  name: serving-certs-ca-bundle
  labels:
    app: prometheus-adapter
  annotations:
    service.beta.openshift.io/inject-cabundle: 'true'
data: {}
`

const Sha256_deploy_internal_nsfs_pvc_cr_yaml = "6dd65ca7d324991b813f209ec6a8a6bcf6c2c9a9f45c519ad3fba51e25042f07"

const File_deploy_internal_nsfs_pvc_cr_yaml = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: noobaa-default-nsfs-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 30Gi
`

const Sha256_deploy_internal_pod_agent_yaml = "0d3d438a85024b605e1d1b3587c0bf9522f7e30f187fdd0f1d607337e3df90d1"

const File_deploy_internal_pod_agent_yaml = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: noobaa
  name: noobaa-agent
spec:
  containers:
    - name: noobaa-agent
      image: NOOBAA_CORE_IMAGE
      imagePullPolicy: IfNotPresent
      resources:
        # https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
        requests:
          cpu: "999m"
          memory: "1Gi"
        limits:
          cpu: "999m"
          memory: "1Gi"
      env:
        # Insert the relevant config for the current agent
        - name: CONTAINER_PLATFORM
          value: KUBERNETES
        - name: AGENT_CONFIG
        - name: NOOBAA_LOG_LEVEL
        - name: NOOBAA_LOG_COLOR
      command: ["/noobaa_init_files/noobaa_init.sh", "agent"]
      # Insert the relevant image for the agent
      ports:
        # This should change according to the allocation from the NooBaa server
        - containerPort: 60101
      volumeMounts:
        - name: noobaastorage
          mountPath: /noobaa_storage
        - name: tmp-logs-vol
          mountPath: /usr/local/noobaa/logs
      securityContext:
        runAsNonRoot: true
        allowPrivilegeEscalation: false
  automountServiceAccountToken: false
  securityContext:
    runAsUser: 10001
    runAsGroup: 0
    fsGroup: 0
    fsGroupChangePolicy: "OnRootMismatch"
  volumes:
    - name: tmp-logs-vol
      emptyDir: {}
    - name: noobaastorage
      persistentVolumeClaim:
        claimName: noobaa-pv-claim
`

const Sha256_deploy_internal_prometheus_rules_yaml = "a6c6475935673a77c31f3d6bd66b284a5bf6c9b62c05778456795cfee50394ab"

const File_deploy_internal_prometheus_rules_yaml = `apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    prometheus: k8s
    role: alert-rules
  name: prometheus-noobaa-rules
  namespace: default
spec:
  groups:
  - name: noobaa-telemeter.rules
    rules:
    - expr: |
        sum(NooBaa_num_unhealthy_buckets + NooBaa_num_unhealthy_bucket_claims)
      record: job:noobaa_total_unhealthy_buckets:sum
    - expr: |
        NooBaa_num_unhealthy_namespace_buckets
      record: job:noobaa_total_unhealthy_namespace_buckets
    - expr: |
        sum(NooBaa_num_buckets + NooBaa_num_buckets_claims)
      record: job:noobaa_bucket_count:sum
    - expr: |
        NooBaa_num_namespace_buckets
      record: job:noobaa_namespace_bucket_count
    - expr: |
        sum(NooBaa_num_objects + NooBaa_num_objects_buckets_claims)
      record: job:noobaa_total_object_count:sum
    - expr: |
        NooBaa_accounts_num
      record: noobaa_accounts_num
    - expr: |
        NooBaa_total_usage
      record: noobaa_total_usage
  - name: noobaa-odf.rules
    rules:
    - expr: |
        NooBaa_odf_health_status
      labels:
        system_type: OCS
        system_vendor: Red Hat
      record: odf_system_health_status
    - expr: |
        NooBaa_total_usage
      labels:
        system_type: OCS
        system_vendor: Red Hat
      record: odf_system_raw_capacity_used_bytes
    - expr: |
        sum by (namespace, managedBy, job, service) (rate(NooBaa_providers_ops_read_num[5m]) + rate(NooBaa_providers_ops_write_num[5m]))
      labels:
        system_type: OCS
        system_vendor: Red Hat
      record: odf_system_iops_total_bytes
    - expr: |
        sum by (namespace, managedBy, job, service) (rate(NooBaa_providers_bandwidth_read_size[5m]) + rate(NooBaa_providers_bandwidth_write_size[5m]))
      labels:
        system_type: OCS
        system_vendor: Red Hat
      record: odf_system_throughput_total_bytes
    - expr: |
        sum(NooBaa_num_buckets + NooBaa_num_buckets_claims)
      record: odf_system_bucket_count
      labels:
        system_type: OCS
        system_vendor: Red Hat
    - expr: |
        sum(NooBaa_num_objects + NooBaa_num_objects_buckets_claims)
      record: odf_system_objects_total
      labels:
        system_type: OCS
        system_vendor: Red Hat
  - name: noobaa-replication.rules
    rules:
    - expr: |
        sum_over_time(sum by (replication_id) (NooBaa_replication_last_cycle_writes_size)[1y:6m])
      record: noobaa_replication_total_writes_size
    - expr: |
        sum_over_time(sum by (replication_id) (NooBaa_replication_last_cycle_writes_num)[1y:6m])
      record: noobaa_replication_total_writes_num
    - expr: |
        sum_over_time(sum by (replication_id) (NooBaa_replication_last_cycle_error_writes_size)[1y:6m])
      record: noobaa_replication_total_error_writes_size
    - expr: |
        sum_over_time(sum by (replication_id) (NooBaa_replication_last_cycle_error_writes_num)[1y:6m])
      record: noobaa_replication_total_error_writes_num
    - expr: |
        count_over_time(count by (replication_id) (NooBaa_replication_last_cycle_writes_size)[1y:6m])
      record: noobaa_replication_total_cycles
  - name: bucket-state-alert.rules
    rules:
    - alert: NooBaaBucketErrorState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is in error state for
          more than 5m
        message: A NooBaa Bucket Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_status{bucket_name=~".*"} == 0
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaNamespaceBucketErrorState
      annotations:
        description: A NooBaa namespace bucket {{ $labels.bucket_name }} is in error
          state for more than 5m
        message: A NooBaa Namespace Bucket Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_namespace_bucket_status{bucket_name=~".*"} == 0
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaBucketReachingSizeQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is using {{ printf
          "%0.0f" $value }}% of its quota
        message: A NooBaa Bucket Is In Reaching Size Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_size_quota{bucket_name=~".*"} > 80
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaBucketExceedingSizeQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is exceeding its size quota
          - {{ printf "%0.0f" $value }}% used
        message: A NooBaa Bucket Is In Exceeding Size Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_size_quota{bucket_name=~".*"} >= 100
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaBucketReachingQuantityQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is using {{ printf "%0.0f" $value }}% of its quantity quota
        message: A NooBaa Bucket Is In Reaching Quantity Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_quantity_quota{bucket_name=~".*"} > 80
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaBucketExceedingQuantityQuotaState
      annotations:
        description: A NooBaa bucket {{ $labels.bucket_name }} is exceeding its quantity quota - {{ printf "%0.0f" $value }}% used
        message: A NooBaa Bucket Is In Exceeding Quantity Quota State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_bucket_quantity_quota{bucket_name=~".*"} >= 100
      for: 5m
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
      for: 5m
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
      for: 5m
      labels:
        severity: warning
  - name: resource-state-alert.rules
    rules:
    - alert: NooBaaResourceErrorState
      annotations:
        description: A NooBaa resource {{ $labels.resource_name }} is in error state
          for more than 5m
        message: A NooBaa Resource Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_resource_status{resource_name=~".*"} == 0
      for: 5m
      labels:
        severity: warning
    - alert: NooBaaNamespaceResourceErrorState
      annotations:
        description: A NooBaa namespace resource {{ $labels.namespace_resource_name
          }} is in error state for more than 5m
        message: A NooBaa Namespace Resource Is In Error State
        severity_level: warning
        storage_type: NooBaa
      expr: |
        NooBaa_namespace_resource_status{namespace_resource_name=~".*"} == 0
      for: 5m
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
      for: 5m
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
      for: 5m
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
      for: 5m
      labels:
        severity: critical
`

const Sha256_deploy_internal_pvc_agent_yaml = "c76fd98867e2e098204377899568a6e1e60062ece903c7bcbeb3444193ec13f8"

const File_deploy_internal_pvc_agent_yaml = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    app: noobaa
  name: noobaa-pv-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 30Gi
`

const Sha256_deploy_internal_route_mgmt_yaml = "1d462d165da5a660b85900e46a11e4d1a53e1498bf9d086b4b68afdceab08394"

const File_deploy_internal_route_mgmt_yaml = `apiVersion: route.openshift.io/v1
kind: Route
metadata:
  labels:
    app: noobaa
  name: noobaa-mgmt
spec:
  port:
    targetPort: mgmt-https
  tls:
    termination: reencrypt
    insecureEdgeTerminationPolicy: Redirect
  to:
    kind: Service
    name: noobaa-mgmt
    weight: 100
  wildcardPolicy: None
`

const Sha256_deploy_internal_route_s3_yaml = "51a2eeee88436d97847f2911a4d05077885dd94bc69c537ad67417bc823a8e20"

const File_deploy_internal_route_s3_yaml = `apiVersion: route.openshift.io/v1
kind: Route
metadata:
  annotations:
    haproxy.router.openshift.io/disable_cookies: 'true'
  labels:
    app: noobaa
  name: s3
spec:
  port:
    targetPort: s3-https
  tls:
    insecureEdgeTerminationPolicy: Allow
    termination: reencrypt
  to:
    kind: Service
    name: s3
    weight: 100
  wildcardPolicy: None
`

const Sha256_deploy_internal_route_sts_yaml = "a593179d9e3864dbc953e61cae744cd747ddd41aeb524248597f8f932680854e"

const File_deploy_internal_route_sts_yaml = `apiVersion: route.openshift.io/v1
kind: Route
metadata:
  labels:
    app: noobaa
  name: sts
spec:
  port:
    targetPort: sts-https
  tls:
    termination: reencrypt
  to:
    kind: Service
    name: sts
    weight: 100
  wildcardPolicy: None
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

const Sha256_deploy_internal_service_db_yaml = "ad9f76ccec1a38c07af34d0251e9e3f3d64bfad48ebaebbdfeef653af1e6eafc"

const File_deploy_internal_service_db_yaml = `apiVersion: v1
kind: Service
metadata:
  name: noobaa-db
  labels:
    app: noobaa
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: "noobaa-db-serving-cert"
    service.alpha.openshift.io/serving-cert-secret-name: "noobaa-db-serving-cert"
spec:
  clusterIP: None # headless service
  selector:
    noobaa-db: SYSNAME
  ports:
    - port: 5432
      targetPort: 5432
      name: postgres
`

const Sha256_deploy_internal_service_mgmt_yaml = "fa5f052fb360e6893fc446a318413a6f494a8610706ae7e36ff985b3b3a5c070"

const File_deploy_internal_service_mgmt_yaml = `apiVersion: v1
kind: Service
metadata:
  name: SYSNAME-mgmt
  labels:
    app: noobaa
    noobaa-mgmt-svc: "true"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/scheme: http
    prometheus.io/port: "8080"
    service.beta.openshift.io/serving-cert-secret-name: noobaa-mgmt-serving-cert
    service.alpha.openshift.io/serving-cert-secret-name: noobaa-mgmt-serving-cert
spec:
  type: ClusterIP
  selector:
    noobaa-mgmt: SYSNAME
  ports:
    - port: 80
      name: mgmt
      targetPort: 8080
    - port: 443
      name: mgmt-https
      targetPort: 8443
    - port: 8445
      name: bg-https
    - port: 8446
      name: hosted-agents-https
`

const Sha256_deploy_internal_service_s3_yaml = "df7d8c8ee81b820678b7d8648b26c6cf86da6be00caedad052c3848db5480c37"

const File_deploy_internal_service_s3_yaml = `apiVersion: v1
kind: Service
metadata:
  name: s3
  labels:
    app: noobaa
    noobaa-s3-svc: "true"
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: 'noobaa-s3-serving-cert'
    service.alpha.openshift.io/serving-cert-secret-name: 'noobaa-s3-serving-cert'
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
    - port: 8444
      name: md-https
    - port: 7004
      name: metrics

`

const Sha256_deploy_internal_service_sts_yaml = "79224e49aed0b4466014599fad98dce701cff56f819c9fe5accf71144fffb404"

const File_deploy_internal_service_sts_yaml = `apiVersion: v1
kind: Service
metadata:
  name: sts
  labels:
    app: noobaa
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: 'noobaa-sts-serving-cert'
    service.alpha.openshift.io/serving-cert-secret-name: 'noobaa-sts-serving-cert'
spec:
  type: LoadBalancer
  selector:
    noobaa-s3: SYSNAME
  ports:
    - port: 443
      targetPort: 7443
      name: sts-https

`

const Sha256_deploy_internal_service_syslog_yaml = "f9f17c0ed02a78b01ff78d80b3881fbb9defbe7b7f22c4a88467566574738cf2"

const File_deploy_internal_service_syslog_yaml = `apiVersion: v1
kind: Service
metadata:
  name: SYSNAME-syslog
  labels:
    app: noobaa
    noobaa-syslog-svc: "true"
spec:
  type: ClusterIP
  selector:
    noobaa-mgmt: SYSNAME
  ports:
    - protocol: UDP
      port: 514
      name: syslog
      targetPort: 5140
`

const Sha256_deploy_internal_service_admission_webhook_yaml = "810a70b263d44621713864aa6e6e72e6079bbdc02f6e2b9143ba9ebf4ab52102"

const File_deploy_internal_service_admission_webhook_yaml = `apiVersion: v1
kind: Service
metadata:
  name: admission-webhook-service
spec:
  ports:
  - name: webhook
    port: 443
    targetPort: 8080
  selector:
    noobaa-operator: deployment
`

const Sha256_deploy_internal_servicemonitor_mgmt_yaml = "172b25b71872e74fb32ecf32b9c68d41cc60d155cb469ed5ecf7ad282f3e597a"

const File_deploy_internal_servicemonitor_mgmt_yaml = `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: noobaa-mgmt-service-monitor
  labels:
    app: noobaa
spec:
  endpoints:
  - port: mgmt
    path: /metrics/web_server
  - port: mgmt
    path: /metrics/bg_workers
  - port: mgmt
    path: /metrics/hosted_agents
  namespaceSelector: {}
  selector:
    matchLabels:
      noobaa-mgmt-svc: "true"
`

const Sha256_deploy_internal_servicemonitor_s3_yaml = "e3940bdfdfbaf5cacefa51f92623ffb00e5360e58640c67558b5cf5135edd57f"

const File_deploy_internal_servicemonitor_s3_yaml = `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: noobaa-s3-service-monitor
  labels:
    app: noobaa
spec:
  endpoints:
  - port: metrics
    path: /
  namespaceSelector: {}
  selector:
    matchLabels:
      noobaa-s3-svc: "true"
`

const Sha256_deploy_internal_statefulset_core_yaml = "50e5b11d8e0a2f2bb8a6db8d154b34b6569e160fa7ad2b1fb154001b36c8a152"

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
  template:
    metadata:
      labels:
        app: noobaa
        noobaa-core: noobaa
        noobaa-mgmt: noobaa
      annotations:
        noobaa.io/configmap-hash: ""
    spec:
    # Notice that changing the serviceAccountName would need to update existing AWS STS role trust policy for customers
      serviceAccountName: noobaa-core
      volumes:
        - name: logs
          emptyDir: {}
        - name: mgmt-secret
          secret:
            secretName: noobaa-mgmt-serving-cert
            optional: true
        - name: noobaa-server
          secret:
            secretName: noobaa-server
            optional: true
        # This service account token can be used to provide identity outside the cluster.
        # For example, this token can be used with AssumeRoleWithWebIdentity to authenticate with AWS using IAM OIDC provider and STS.
        - name: bound-sa-token
          projected:
            sources:
              - serviceAccountToken:
                  path: token
                  # For testing purposes change the audience to api
                  audience: openshift
      securityContext:
        runAsUser: 10001
        runAsGroup: 0
      containers:
        #----------------#
        # CORE CONTAINER #
        #----------------#
        - name: core
          image: NOOBAA_CORE_IMAGE
          volumeMounts:
            - name: logs
              mountPath: /log
            - name: mgmt-secret
              mountPath: /etc/mgmt-secret
              readOnly: true
            - name: noobaa-server
              mountPath: /etc/noobaa-server
              readOnly: true
            - name: bound-sa-token
              mountPath: /var/run/secrets/openshift/serviceaccount
              readOnly: true
          resources:
            requests:
              cpu: "999m"
              memory: "4Gi"
            limits:
              cpu: "999m"
              memory: "4Gi"
          ports:
            - containerPort: 8080
            - containerPort: 8443
            - containerPort: 8444
            - containerPort: 8445
            - containerPort: 8446
            - containerPort: 60100
          env:
            - name: NOOBAA_DISABLE_COMPRESSION
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_DISABLE_COMPRESSION
            - name: DISABLE_DEV_RANDOM_SEED
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: DISABLE_DEV_RANDOM_SEED
            - name: NOOBAA_LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_LOG_LEVEL
            - name: RESTRICT_RESOURCE_DELETION
              value: "false"
            - name: NOOBAA_LOG_COLOR
              valueFrom:
                configMapKeyRef:
                  name: noobaa-config
                  key: NOOBAA_LOG_COLOR
            - name: POSTGRES_HOST
              value: "noobaa-db-pg-0.noobaa-db-pg"
            - name: POSTGRES_PORT
            - name: POSTGRES_DBNAME
              value: nbcore
            - name: POSTGRES_USER
            - name: POSTGRES_PASSWORD
            - name: POSTGRES_CONNECTION_STRING
            - name: POSTGRES_SSL_REQUIRED
            - name: POSTGRES_SSL_UNAUTHORIZED
            - name: GUARANTEED_LOGS_PATH
            - name: DB_TYPE
              value: postgres
            - name: CONTAINER_PLATFORM
              value: KUBERNETES
            - name: NOOBAA_ROOT_SECRET
            - name: NODE_EXTRA_CA_CERTS
            - name: AGENT_PROFILE
              value: VALUE_AGENT_PROFILE
            - name: OAUTH_AUTHORIZATION_ENDPOINT
              value: ""
            - name: OAUTH_TOKEN_ENDPOINT
              value: ""
            - name: NOOBAA_SERVICE_ACCOUNT
              valueFrom:
                fieldRef:
                  fieldPath: spec.serviceAccountName
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
          securityContext:
            runAsNonRoot: true
            allowPrivilegeEscalation: false
        - name: noobaa-log-processor
          image: NOOBAA_CORE_IMAGE
          command:
            [
              "/root/node_modules/noobaa-core/src/deploy/NVA_build/noobaa_logs.sh",
            ]
          volumeMounts:
            - name: logs
              mountPath: /log
          resources:
            requests:
              cpu: "200m"
              memory: "500Mi"
            limits:
              cpu: "200m"
              memory: "500Mi"
          ports:
            - containerPort: 5140
            - containerPort: 6514
          env:
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
`

const Sha256_deploy_internal_statefulset_postgres_db_yaml = "37a6c36928ba426ca04fd89e1eb2685e10d1a5f65c63ebb40c68a4f5c37645de"

const File_deploy_internal_statefulset_postgres_db_yaml = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: noobaa-db-pg
  labels:
    app: noobaa
spec:
  replicas: 1
  selector:
    matchLabels:
      noobaa-db: noobaa
  serviceName: noobaa-db-pg
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: noobaa
        noobaa-db: postgres
    spec:
      serviceAccountName: noobaa-db
      containers:
        #--------------------#
        # Postgres CONTAINER #
        #--------------------#
        - name: db
          image: NOOBAA_DB_IMAGE
          env:
            - name: POSTGRESQL_DATABASE
              value: nbcore
            - name: LC_COLLATE
              value: C
            - name: POSTGRESQL_USER
              valueFrom:
                secretKeyRef:
                  key: user
                  name: noobaa-db
            - name: POSTGRESQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: password
                  name: noobaa-db
          imagePullPolicy: "IfNotPresent"
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
          ports:
            - containerPort: 5432
          resources:
            requests:
              cpu: "500m"
              memory: "4Gi"
            limits:
              cpu: "500m"
              memory: "4Gi"
          volumeMounts:
            - name: db
              mountPath: /var/lib/pgsql
            - name: noobaa-postgres-config-volume
              mountPath: /opt/app-root/src/postgresql-cfg
      volumes:
        - name: noobaa-postgres-config-volume
          configMap:
            name: noobaa-postgres-config
      securityContext:
        runAsUser: 10001
        runAsGroup: 0
        fsGroup: 0
        fsGroupChangePolicy: "OnRootMismatch"
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

const Sha256_deploy_internal_text_system_status_readme_ready_tmpl = "224e606bf5eee299b2b2070a193a6579f9cb685e78e47d7b92fe9714c9eee63f"

const File_deploy_internal_text_system_status_readme_ready_tmpl = `

	Welcome to NooBaa!
	-----------------
	NooBaa Core Version:     {{.CoreVersion}}
	NooBaa Operator Version: {{.OperatorVersion}}

	Lets get started:

	Test S3 client:

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

const Sha256_deploy_job_analyze_resource_yml = "c80810baeda94fd9dd97a6c62241be5c582e08009bdbb1f2a13992c99d90ea33"

const File_deploy_job_analyze_resource_yml = `apiVersion: batch/v1
kind: Job
metadata:
  name: noobaa-analyze-resource
  labels:
    app: noobaa
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 0
  activeDeadlineSeconds: 60
  ttlSecondsAfterFinished: 10
  template:
    spec:
      volumes:
      - name: cloud-credentials
        secret:
          secretName: SECRET_NAME_PLACEHOLDER
          optional: true
      containers:
      - name: noobaa-analyze-resource
        image: NOOBAA_CORE_IMAGE_PLACEHOLDER
        env:
          - name: RESOURCE_TYPE
          - name: RESOURCE_NAME
          - name: BUCKET
          - name: ENDPOINT
          - name: REGION
          - name: S3_SIGNATURE_VERSION
          - name: HTTP_PROXY
          - name: HTTPS_PROXY
          - name: NO_PROXY
          - name: NODE_EXTRA_CA_CERTS
        command: 
            - /bin/bash
            - -c
            - "cd /root/node_modules/noobaa-core/; node ./src/tools/diagnostics/analyze_resource/analyze_resource.js"
        volumeMounts:
            - name: cloud-credentials
              mountPath: "/etc/cloud-credentials"
              readOnly: true
      restartPolicy: Never
`

const Sha256_deploy_namespace_yaml = "303398323535d7f8229cb1a5378ad019cf4fa7930891688e3eea55c77e7bf69a"

const File_deploy_namespace_yaml = `apiVersion: v1
kind: Namespace
metadata:
  name: noobaa
  labels:
    openshift.io/cluster-monitoring: "true"
`

const Sha256_deploy_obc_lib_bucket_provisioner_package_yaml = "26eed5792ad7e75fa7e02329e648efff0be25f33595dcc1b4671fb99758f7cc0"

const File_deploy_obc_lib_bucket_provisioner_package_yaml = `packageName: lib-bucket-provisioner
channels:
  - name: alpha
    currentCSV: lib-bucket-provisioner.v1.0.0
defaultChannel: alpha
`

const Sha256_deploy_obc_lib_bucket_provisioner_v1_0_0_clusterserviceversion_yaml = "aee3bfbb7be1965fbe6ec0741802d84fc81f3b47ea213c1c8bb1bb2c3eb130b6"

const File_deploy_obc_lib_bucket_provisioner_v1_0_0_clusterserviceversion_yaml = `apiVersion: operators.coreos.com/v1alpha1
kind: ClusterServiceVersion
metadata:
  name: lib-bucket-provisioner.v1.0.0
  namespace: placeholder
  annotations:
    categories: Storage,Big Data
    capabilities: Basic Install
    repository: https://github.com/kube-object-storage/lib-bucket-provisioner
    containerImage: kubernetes/pause
    createdAt: 2014-07-19T07:02:32.267701596Z
    certified: "false"
    description: Library for the dynamic provisioning of object store buckets to be used by object store providers.
    support: Red Hat
    alm-examples: |-
      [
        {
          "apiVersion": "objectbucket.io/v1alpha1",
          "kind": "ObjectBucketClaim",
          "metadata": {
            "name": "my-obc",
            "namespace": "my-app"
          },
          "spec": {
            "storageClassName": "object-bucket-class",
            "generateBucketName": "my-obc",
            "additionalConfig": {}
          },
          "status": {}
        },
        {
          "apiVersion": "objectbucket.io/v1alpha1",
          "kind": "ObjectBucket",
          "metadata": {
            "name": "my-obc"
          },
          "spec": {
            "storageClassName": "object-bucket-class",
            "reclaimPolicy": "Delete",
            "claimRef": {
              "name": "my-obc",
              "namespace": "my-app"
            },
            "endpoint": {
              "bucketHost": "xxx",
              "bucketPort": 80,
              "bucketName": "my-obc-1234-5678-1234-5678",
              "region": "xxx",
              "subRegion": "xxx",
              "additionalConfig": {}
            },
            "additionalState": {}
          },
          "status": {}
        }
      ]
spec:
  displayName: lib-bucket-provisioner
  version: "1.0.0"
  minKubeVersion: 1.10.0
  maturity: alpha
  provider:
    name: Red Hat
  links:
    - name: Github
      url: https://github.com/kube-object-storage/lib-bucket-provisioner
  maintainers:
    - email: jcope@redhat.com
      name: Jon Cope
    - email: jvance@redhat.com
      name: Jeff Vance
    - email: gmargali@redhat.com
      name: Guy Margalit
    - email: dzaken@redhat.com
      name: Danny Zaken
    - email: nbecker@redhat.com
      name: Nimrod Becker
  keywords:
    - kubernetes
    - openshift
    - object
    - bucket
    - storage
    - cloud
    - s3
  installModes:
    - supported: true
      type: OwnNamespace
    - supported: true
      type: SingleNamespace
    - supported: true
      type: MultiNamespace
    - supported: true
      type: AllNamespaces
  description: |
    ### CRD-only Operator

    This operator package is **CRD-only** and the operator is a no-op operator.

    Instead, bucket provisioners using this library are using these CRD's and using CSV [required-crds](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md#required-crds) them so that OLM can install it as a dependency.

    ### Generic Bucket Provisioning

    Kubernetes natively supports dynamic provisioning for many types of file and block storage, but lacks support for object bucket provisioning. 

    This repo is a placeholder for an object store bucket provisioning library, very similar to the Kubernetes [sig-storage-lib-external-provisioner](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/controller/controller.go) library.

    ### Known Provisioners
    - https://github.com/noobaa/noobaa-operator (NooBaa)
    - https://github.com/rook/rook (Rook-Ceph)
    - https://github.com/yard-turkey/aws-s3-provisioner (AWS-S3)

  customresourcedefinitions:
    owned:
      - name: objectbucketclaims.objectbucket.io
        kind: ObjectBucketClaim
        version: v1alpha1
        displayName: ObjectBucketClaim
        description: Claim a bucket just like claiming a PV.
          Automate you app bucket provisioning by creating OBC with your app deployment.
          A secret and configmap (name=claim) will be created with access details for the app pods.
        resources:
          - name: secrets
            kind: Secret
            version: v1
          - name: configmaps
            kind: ConfigMap
            version: v1
      - name: objectbuckets.objectbucket.io
        kind: ObjectBucket
        version: v1alpha1
        displayName: ObjectBucket
        description: Used under-the-hood. Created per ObjectBucketClaim and keeps provisioning information
        resources:
          - name: secrets
            kind: Secret
            version: v1
          - name: configmaps
            kind: ConfigMap
            version: v1
  install:
    strategy: deployment
    spec:
      deployments:
        - name: lib-bucket-provisioner
          spec:
            replicas: 1
            selector:
              matchLabels:
                name: lib-bucket-provisioner
            template:
              metadata:
                labels:
                  name: lib-bucket-provisioner
              spec:
                serviceAccountName: lib-bucket-provisioner
                containers:
                  - name: lib-bucket-provisioner
                    image: kubernetes/pause
                    imagePullPolicy: Always
                    env:
                      - name: WATCH_NAMESPACE
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.namespace
                      - name: POD_NAME
                        valueFrom:
                          fieldRef:
                            fieldPath: metadata.name
                      - name: OPERATOR_NAME
                        value: "lib-bucket-provisioner"
      permissions:
        - serviceAccountName: lib-bucket-provisioner
          rules: []
      clusterPermissions:
        - serviceAccountName: lib-bucket-provisioner
          rules:
            - apiGroups:
                - objectbucket.io
              resources:
                - "*"
              verbs:
                - "*"
  icon:
    - mediatype: image/png
      base64data: iVBORw0KGgoAAAANSUhEUgAAAUAAAAFACAYAAADNkKWqAAAAAXNSR0IArs4c6QAAAAlwSFlzAAALEwAACxMBAJqcGAAAAmhpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IlhNUCBDb3JlIDUuNC4wIj4KICAgPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4KICAgICAgPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIKICAgICAgICAgICAgeG1sbnM6ZXhpZj0iaHR0cDovL25zLmFkb2JlLmNvbS9leGlmLzEuMC8iCiAgICAgICAgICAgIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIj4KICAgICAgICAgPGV4aWY6UGl4ZWxZRGltZW5zaW9uPjU0MDwvZXhpZjpQaXhlbFlEaW1lbnNpb24+CiAgICAgICAgIDxleGlmOlBpeGVsWERpbWVuc2lvbj41NDA8L2V4aWY6UGl4ZWxYRGltZW5zaW9uPgogICAgICAgICA8ZXhpZjpDb2xvclNwYWNlPjE8L2V4aWY6Q29sb3JTcGFjZT4KICAgICAgICAgPHRpZmY6UmVzb2x1dGlvblVuaXQ+MjwvdGlmZjpSZXNvbHV0aW9uVW5pdD4KICAgICAgICAgPHRpZmY6T3JpZW50YXRpb24+MTwvdGlmZjpPcmllbnRhdGlvbj4KICAgICAgPC9yZGY6RGVzY3JpcHRpb24+CiAgIDwvcmRmOlJERj4KPC94OnhtcG1ldGE+CqBAHCwAAEAASURBVHgB7F0HgFXF1f5e3b5LLwtIB0GaoCgKihRL1EhsUWONRmOJ3Rg09o69gYqKLcYSewGlKBZUFGtQQVFRkCJt2b6v7P999+1ZLs/XFlh/ojvwdu6de+bMmXNnzj0zc86MZ+3atbUMSBeKiorg9XohWI/HkxRcz/Vbv369EycFrHsg2EAggIKCgnr4VPiVraamBmVlZfX0pCojGo0iLy8P2dnZGdGussvLy1FdXZ0RftFfWFgIv9+fEX7RKt6IrkRB+Kz+tbVR+PxZKCrMIajPAa+tjaCqKoyALwB/UO9N74Lvpe4qXFWDkvL18Pp8iEai9bgMp4Ok7o/KEl9yc3Mzol04KisrnZ+1BTe++Gvh13vV+3XXKx5O9/ZcvAmHwxnxXvSoXSq2/IlwW1okEnF4b/epYuELBoPIz893cAs2EQ+VbmWrzajtZMIbvX/hzsrKqs8vXImC4VebV9vPBL/wqF362A4sfyLcStNzhZKSkvprJyHJH8GrvQu/5U3GG0MRCoVQWlqaEe3CmZOT4/x0nQq3Pa+oqGC/qMoYv/VZvxHYFG8dHFDj3iAcPQgEPVi3vgRfffklf4vwwYcfYVV5CbJ8eWiWlYPeAwai8zbboF3zfHTo3hF5zQoRDFPghKOopWBQUCcQThMUW0dNm6ho4sD/PweaBOD//zvYiAIJKn3VAgFplB68Mf1V3Hb3ZEx9aepGcHpx4Y1SgD5Dd8XRh/8RQ7cfiN49e1BzjH2hpZmYhhqXJe7W9Mi45KbbJg78SjnQJAC3ohfraGgUgD6fH2EOXx956EGcceZZDoUTjxiMoe2iyCrIRxWHv8HqMIKeKNYhiKXrqrFkdRVe/GwBxp91ugP/+/1+hyOPOho77zwMrdu04fCgGpFI2MHNQQ9hYsPG2HWMCRK4dUpjLKHpbxMHfuUcaBKAW9ELjnLOr9broxCK4Jkn/+MIv+NGbocjdyhEvxYhVIezEI6G0TxagnCAc0deoAPK0a1dEKG2udi3dy98O7YHZnxThQnPvYznX3wZY8eMxkkn/RXDRwxHXn4Baqor4WFGL4UnLxwt01jg4X0sJJ/jNdimuIkDvwYONAnAregtSvxkc8FgxowXccKJJ2H/XQbg3N3y0Dw7gpIqCkdUUzh6EakNUHjxnmPgEBdHqqJK540/iP7Nq9F3aDYO6DMcU7+swpXPzsT0GTNxwrFH47TTTkOXPv2oXbKg6hAivhriNKG3YTJ/K2JJEylNHGhUDlCHaApbAwc075fFoW/Jyp9w9WU3kKQALtvDjyJ/COsqqa1pRZgamg8RZ5jqrP3yj4asWT4Pcigca6jZreMCSFVVJdrnVeMvQ/2Yfd4wHDNiW9z7wEMYtMNQPHLffahevwL+3AAXRqQJbhCAoHDVMLgpNHHgt8KBX70ATLWEvrW95FquAH/0xeeY+/EHmHjUDijO9qAqRJMDr8dZyY3NEUp92xA8iFL0+VHqa4as2mrk1ZY78NGaCGqqIuhaGMGFYwtx17GDKFPb4LTT/4ar/vEP/LDwIxRkZ1EIxsyWhFFC2LGq2YC+6aqROLC1tUvn3TdSXbdmtFuFAGxM5jcm7s19sT/rBBRG77/7roN2cOsKlHtpu0iFLFC3MszHFFDBuGIpHGkbmB9ZDz+HwmFfNij7EPDS7IXD28oItTzav+2/bS5mnNYeB+/SHxMfeRx7bL8rZr0+i3aGAWqRsQURL+Nfq/6ndvAzfsdxMtXt5uRNhHdra5dbun6J6pxp2i/JG6+7MDEh2U9wBmvX8bEqqPwG565wMryW7oaNx2v3BqP7+HIMT6LY8ik2XIliN5zBWloivJZmMJYnGW7B65kF97WXix8hLlB8u/h77N8JaNWyOTzVNPamcIu6pZIWL+JClEPXKn8uaigcQ5KYnAus8vg5oo0gK1LGYXMUpeUhbNvSj6vG5OHSA7fHSuLYb7/f4/F//9spV/aHVVx5ruXEougycxw3jVZsfD2MD/GxwVssXMl+bpzua+WNx6t7C4ZP93YdH+uZ8mywr1RKLCTCbWkGozgep/ve6FGawToX/GO4EsUGY3ncON3Xhl9wurZy7D4VboN143NfGw6D072FRHgtTTDuPG6c7mvDpbRMaRecO7jxxV8brNIVLNa10ZooNtj6RRBDrAeJgiFJ9MzShMPgdK0GZ1brbsIMXrDucnWfKrjxC7e7rC2FX+ULl9EeT6ObPj1zd6pU9Bvtxg+j3/BpBXY9vQg+X7oUQzvmIita4wgzzdGFyBap6jHuOGqgZXNichEgfKx8zRdqmKxUD0LeLF5rGB1BaVUt8rNCOGmQB32bbYdDXw7huD//GV/M/y/Ov/hytCjMRrimGr5AroPLeODmra7lUaF6pOONYC1vKt6oEoIzfFaupRsO3ccH5ckUt+EXDrtOhFvPRIM9yxS/8Brt6fALd0PxC168l2G7ghuHk1D3x8rWbaa0Wx6jPx1uPU+HW+W74YRbeawsPYsP8c/SlRGPP5N2KToUhNtvzDRC3AiVZvdifCKCLZ8bVgUIr4ixDu+Gc18Lp2CFPz4YMyw2WhTLsFfp9iw+r90bQxLhNxjDYfh1L/zK6w723J0m2lXfZPiFS8GYrmvlUbo9E14nje2htCqE/CBn9bxhlHkDyAJdnyTg6kOyj4Q7fcO1Ixzr8ub6alHCxeIczhnu2t2Dt44I4tLZfTHhxptQsnoVrrz2WjRr0x6hUEyYqv6ql+h0112ubfHv1f3cSFWdlB7PG8MnOLsWnILyKBhvnBv+icevezd9BqfYcFqa5bV2qfR4+g3WYj03/JZmseG32PDrubUbe2Z54mPhVsiUN8KnPIbfjc9dvjvd3eYsPZ4uy2u8sXorPVnQM72neNoFb/gtNvyKf6k+a2WLHitf1xZEu+or+v3y1VSGVEFI5KOpDIbc4vh8StevIT6a8v2UH6IbZ6JrpYlw+a8Kv+hSUHqqIP9V4dfLVX43bnc+PVcdxRP5D2eKXz6O8nU0vBa7ceva6HT7VRuMngWDWRjSuze+nTEb1R76/0ZLJRFIB/Ma4GbEIQrS7Fo2XA/9eauz0bFlFm7Zy4NLcgbgbq4Sf/X9ctw2aSI6F7ehL/TGPqeiT7yT/6r5DosUq1MysuS/Kv9Yyy8eG1/defRc6eKNCc50uEWP2+dc+Q1PItzCK/wKBuuGi7/WO03VLg3e6ib/1WbNmtXXLx398l81/KpLMt5Yu5T/qrXLdLhVP7VL9S3jicVGt8VK10+8MRjFqYLwivcKKkvB8jo3dX+UprrJ51n43bBuuPhr+fU2pM+qXTaE99ZnOe+9oaIiLtHPDaNrvRDF8T97gQYvXO7rRLhVcYPRtXDE4zcYg3PHdp0It9Isr13H41Z++wnG8Fk+u0+F33ALNhl+pRtOxZbHriNsUMFgAP16dsaTPwCrK73Ip6alDQ0ILTAG6XMba6Wx9Mz+ejkbWMuhsI/eI/z+0bE+hMJsL64YXYDjRvXHrFmv4p9nn4bvv/8BOfzIqOF6ktBtPFPJqXhjzwUvAWT54mPBKc0N79zwTyL8ehYPn4z3BqfYcFma3cfH7rJ1LdyZ4BescDcEv2BT8Ua0GT7Dr1ghnm67F7yuFXSdjHalG353Xicj/1iaOzacSlNoCG8sr2I3Tve1nim48asOiX4Go2cKBqNrN073tZ5Z8OuBO7M9SBQbnBUaDxOfLnjDb3ndeQw+HsbS3bC6Njh7brFwp8KvvO7nlk/pqYKVJxh3/vg87vKT4Y5PVx43frYiB+2uu+3uxF+vDqFdJy9qOGQN0DOEq1UUWkH4aBJDubRJIeKhAKoNcVElQCEYIo4oKmlUXeStxNm75tPOsD/ufWkawrQnvPmWm1Hcvh3nBEP0Od7AX6NbBBhPLHYTZfXVM/dzS3fDxl8L3s0bd/54WN3b82S43emG2/BbXjfeeHiDcafHwxuM0g23rt3puldw43E/d6fHIBP/TYffcgm34U+G251u8A3Bb2W58ViaYjcuS7c0o83SDd7u3c+T4RdsPJzdW2z4FLvx6Hn9IogbKNm1O3MyGCHNBE75jcCG5ElWbqJ0w69noilTuhLhSpXWENyJaHDT2adXL6eoZz5cjp1bFSKLK7uyVa7l6m+YdjBBZ7EjNixORVOiZx65jlCDVMxZWl5x2BXlFlEUeC2zKnHOLjShCQ/EQy+/jEigADfdcB06tm1OdxPCSxt0vVvVw013fHn2zOprcTyc+z4TGDe8XWeaz+Dc9TAc7thot9jyuWESXbvhdG3542Hd6YJz54uH3Zz7THGLnk2hQXkyzZsp/k3hTab1FC/j8W/6eCrFm3EXkgLsV/NoS9RXOLSFVavWrTF5yj147N2v8fl67tUXpFlKpJpDX37NvbEdYjJtTD9nsE13EBcxCqePZjMBby09SIC8YBQX7BrEQcMHYepzj+PSf5yL5atKaFWjleTYR8QE4Zao88/p27SUrYmWhtbgf5n2hta1seA3h4eNIgA3vYM2Fov+N/BSEXPC78bu68QT3/ke62pp+ByQlkB/X8e8hauksamXza4UB9Oo5kpzLQ0NOeOHMrrG5eT4cenILOw0qA8ef+o/uHz8WShdu8pZwXM2aOU8ngUqAE1hMznQ1Fc2k4Gbmb1RBOBm0vTbza4J6XA1ijsU4957J+GF95Zgztec+ctuw70PaALBoauX2pptdLq5jOIAmKvClZxj5JA6mgU/t8uqZvktsqpw+9hs9OvdHQ8//hQuvuhiZwfuQCBYP1SS8KPS2hSaOPA/zYEmAbgVvT69jFp6cGhw+oc/HIKddtwRxz8wFwtXVMKXy5VbqojeiIfrt7TJdAkf93VDq8OZPS5ycI6PE40UgxRqXpTV+NC5yIdb922Grr164/4HH8KNN9/MVWPaanIBJhLlQsqWUkMbSnATfBMHtiAHmgTgFmTmlkIV5qJDixYtccONNzoo756zGqig0OPos9wf5coVjcB1Q+HlhE0ciip3zKxGs4EUac5uMDJy5a4y3Eihd5ss3DY6m355PXHj9dfj0ccehTeLQ2bCafFExTYN4WKvoOnv/yYHGiQAM51szBTOWNZYnciN12iy2Mp2x4meuXG4YTf1OlEZhktlmb2WDLJ32XVXXHnN5Xji3S/x5Od+GoDTVpDGuYhyXtARejEB6FIGDdUmxULpp3YpTTOL8rWqvAzbd8zG5D21CNIRZ3InmanTpsEvc5o6s51kBSXiW6q6C0+658nKyiRduA1/ItrcOOy5xXpmed1wdp3qmcG443i8btrccKmu3TiSwTWUrmR4kqU3Bn53vQy/xYnoSPTMjSNRHnfaRobQ7gfx10KqFUDFqX4GY4RZHI/PfS+YVDgTPVN+5UuF3/3M6LI4EU49089COvwGlwhXsjThN7osNjy613MzLFV84smnYffddsbfn3oPLy/MQkFOAYeqNay86IzRuokKoBVbHzsaIZEFuLlqNZ3w5JwYra7C2L7NcfWhLXhXiHH77YcZr0ylAKSEJGx8HQyZO111Ej8s3hTeGF53rDIyweuGUX7lc9PnxmnPLTaaLU5Eu57pZyFT/IK3vKnwW5lGczr8RoflyzR24zccyWLBZkKzlW140tFuNAje8FtsuNyxnuln+dLhNzoMh4fH7GXUf3QcnzJlEkSEjhTMJAhWFdDBPVaJVPlEg6zmhT8TeoRTbjvmqpYKt54Jp3CrjEzxi3b3S0hXhvAnq6vwSPCp7Aivc2iCMnfuXOw0bGei7YCXzuiIIa15LChlYGwxRMac8utwZhDritZ72jS9UMNbFosg0XmiVVjvaY7mtWu5bVYerptbjVtf/Iy4I/jsv/PRb7u+3GK/grwK0kaRXnucR3S3ENVR7cZcKOuISxml4k2ijOJ9JkH8FF+FP5Mg2kW36E/2rtx4hF++rnYEqPtZomvhVJs0V7VEMO404W9Iu1TehvRZwctdLZMg2tVOhT+TINobu8+K7+K/ykoXRL9od/rs6tWr0+cgxubNuT0TM2YSRATPG84E1CFYxMvPMdNgPpqZCClVUn6C8h/ONMh3Vf7DmeBXXeWDKNhMw7p16xzmp4MXbnXaZuT9g1Me4M4tx6FP3/741x/y0CZQjoraLD7njGCoHCF/dp2ZjLDaK83sfSWnQ/aCMeGaRcFYzr0F/zazFDPnfI7dedjSlH89gmatWsIbqkLEm8diJYw2lCney/9T7zfTIP/STD5W4o3aY4sW0kwzC6JHvM8kCFYCQX7hmQYJY/mYZtJuhF+4MxU6okF+1SrDPpCp6BJ/1GcFm2lQn1W+dEEwEjbmV50OXs8b2mfl8yzf6kyD/Kr1y4T3ol+0qw7kT2K/XjUu988YY3EywvRcP8NrsRtXomvhs7ypcBtcJngFo5+FTPELPhP8Vg/hVcgEv2AsX7oy9LzW0USBI484GJecfSK++Pwz3PZeJWq8RTwVjvN1NFuJalNTx0bQIYN/JIQ2CCJLbXgsjY4exHSfq6Ge2cxfgSv2aI1te3bH7Hffwc233orykBoRf7XUlJwDnWLtRrTrZyET3ghWvLG8xqdEscFIkKTDLbwGkwl+wy1Yd17nJsEf4bagvInodacZfstjtNl9fOzG78aT6lplWD6L4/HavZ6Lj4YvXR3sufIrbyr87meWz8pJFAtGPwtbGr+Vafg3lGQpWzh2M2BzUYv4hoREzGsojoaUJ9iG4E/HG/WrLPrsRkPUrLLycco5/8SIngPxwIxP8MC89QhK61Rj4XkhjRcksDXIrkFFyItueeW4bix3VGnRE7ffdhv+/fCDHC/nsQNtvF1aorplwptE7yxV3YQzE7xumHRlJKI9FQ0NfZaofDd9DcW3JeDd5aerf7rnbnrceN3pya4T4W4ojmS4E6U3ugBMVGhTWmYckLwPcbMCr5+WejwbpE1xJ9x4351O5sufW4LnvqhEXpZ8OGTBF9NC5OImA+cNOklmZaWCquWqsIfSOEBhW1oFDO6Ui+tG6GNUjL+fdSrefvttBLI5XJHEdoVEjdn1+Be/bMyO9ItXpqnALcKBJgG4RdjYSEg0vKBw0zkg2fQACVPL2nHErnj2+WdZ4BqcPLMSXywr4QaqnGML0zyG5inaRzCHhspcy60jKjYk2xwKtVu1BGtsVxguzlSV4Q87tMUJu3Pej+GCvx2LH5ct51xfsH7BQ8LPPZTZnPIbklflun/Kq+GdxXrWJAgddjT9IQeaBOBW3AwkdrwcAnslzGic7ONChMLv99sXl1x8EbDiG5zzRhjfl1HcZeXAE67ihlk8QN0bdOz5NH+noE6/OYGzbKSEItCZgajl8Zv0HY5W4IRdW6N9z974aOESTJx4F4/jrHImliVwNBltgmdzym5oXquryta1VnIVdG/0KDYhKBjL09CymuD/9znQJAC34ncYm/KU7kUBVifDJMs0J3fq6X/FqLF74ZNPvsSEOaV8zJ2es2gjqMUIDoKjnBf0sKOz62/xGkqzq+a2+Z3zwrh/n3wK52647ZYb8eqrr9SbJ0nw/H8IFgk2E3TuCXWl6T5e8One0rY4o5oQbvUcaBKAW/Er2kiASPviT0KxhkZ3rVsW486br3Go/8/s+VwUKUUeV4UlIHVKXIjDYZ9zNKbyKOMWrCiFay1PnisLRTGgVS2u2lP2YIU48sgjKZA/dUyORLsEzkZ12IIkJEKlsvST9qmfdtmOSgNkmtmI6bl5sYgvlicRvqa0Xz8HfvUC0Br51vgqM9E8BFNfhzotUDvChHhmx7bbbY+Zs2c5VbvsxQrMXFCGPNpO1URqY6vHElTUAJ1OTqi67FuAFRyaE3eE5cj85oCh3bDvjm0dvFdfeTloW4pAnaF6JnWMJ2hT8qh2yiehu3DBAtx93/348xnn4vC/noZLr52Ad+bMIR+oBUrocdebWBlbjiPxdUh3v2l1TIe16XlDOVB/LrBeSKqf+0tp14li4VB6fEiFW7DuPInwWprBxpeTDL/g9cyC4UkUG4zFglFIhtvSBWOwifBamuANzp3H8CSK3XAxPKSHZHl99CTgeSGjdtsDE++cSOeMb3H0m2F8uIwGsFoU4YKJj+Vx6wLaDNJ31zlcOH44vIEvKifToCF2SNvqU9hEORTWRqp/35WrwNv0wIsvT8Wzjz/qKJzcvEuMcdAaDxRrOGr3di0gpVmw60Q8qU9jKfKWiUbopcSyauhZ8a+H7kfvbbfF4yccj96LbsX2392NtQ9d4PhV//3Cy1BZtg7+bJ61wnxUBeverfSAGC/qcbvajJsmo9sdu5/r2mjXtRuf+1rP3MGNL/7aDadrw+/GF39teQw2Hqf7Xnkdflimujgep/teIIbbrt043df2XPndedz43NeCdwc3rvhrgzO8FrvxJbpWPsE6Z4LYjTuzrpXRYkNiBaaK9RW2hm3DIOFx47f8li786YI7v+aYhNtoNDzCYWmKjQ5LT1WG4Vcs/BpGKbjTde+m1Z650wSTLBhvjH7BuRufm3bh1r3htlg4tH6lc0IU/vznY3mQ0be49rrrcd7LXfHvQ1rRy6AQ6zlEzQnX8BjM2ApxmAbLLMzJoz8ybdmUrbQkbIKca4xyLrIMPEOYB7r3apuDu3YH/vpwMc74+3j0H9gfQ3bZvV7rTFQPh4a69251Vezmje71cwe1FHLGcQGUDWSEw3EvPwaT77wNp511LiYdPQz79qHLmw6G59p1iNu9HrTGhwuvuwZHLPkJD95yFQr5kZB3QqwN2VA9xm+VpXeiYbOC8d25SfDH6BOctRtdG+323LLqWfw7t2eJYsuvWPli7z/WLu2ZladYQem6tvtEeN1p1i6FX+1e+e0nODd+pcf41vA+KzzKKxzxPFCanitd11auc5Hkj8HpsfFe15ZuOJVmwZ6pLA/dRzZuXQYVFzvAzKCQCKmlOwD8Y5VTvkyCvVTBJsJvRNtzu28o/kS4DadiBcEkg4tBbPz3F+ONGkVdAyGBzlyWXnpWVhDLlq/ACccdjVdnvIZDR/TBVaPzKSArqasV8hyRKgoBzodx92e9Dgk+BZ0JIpvBTQrEU0uh4/NRm6IQkiF0ViAHN7y+ErfPWoTdBvfGI89OR3FxseNWKHc4tYl076sh7cZ5T6xBdk4enn/6aRx48MHctGFnHDM4F+U1HIbTZjFCIca9rpHtpytndgGuuHQ6OvzjKpz/97NJj4QtjxzgV0B0eXnglNFn79/apcMx8TyuPQvOgkNP3X08nMG4Y8EYnJXnfq7rRPgtTzxs/H1D8TeE9yrLeKPrRPQnol2wmdJv+BPhtjIVKwgmGVwMYuO/xhv2Bdcb3Bhmozv5UKqzKaMVFJ/Vnolw+SFmGoRX+C2/8rmvDY/wClZ+vfLvzTTIr1f+vcqfjHbh0nM5p8tPMFOneeWT/6qcq1PhV30sNMSvWo3S/FeN3yqnvrGK5ppqFBbl4715n2Lk8F1ZTDYmHNQZhw9phxra7FE+0ZOE2jL3Gdwyu0nz601h6ouG4edPRy2FfdksFVhVEcKQp8qBRQtx++2347jjjnP8V43vBFFr5QuORc49/4g/qpN4Y5q3PUsWcwqS/rGlWPHdl+g9YCiG9u2I+w/pzMUgLsrIRdBbgDC11CyUoibqp1F5ENNX5OPMydPwPjeY2IEbzv48SKDF3lWIbaGE7TL2XiW8TVPcOJeeq13qvGT9Mg3mv5qq3QiX4ZfPuWmlmZTxv9xnZVIl3+dMeLM5fbb+VDjrXIkYq8apn4hxd+R4WMOhBqzGbPkEZ8/i8wjGOofwpwqCE4wxRbBGTyr8grF86fCLdjdNgk+GW88Eq5/lU1qiYDgaQrvhEX4F5VUwPihdHU82gOV0kt9pUF888+w0/GHc3tw+az2y81rgoN4UkHLSj/CAJckdTiDyzThDX+oz9cNoB3HGfzj8DFO79GWh0pdDAxwOtaPcQCKajXbNcvCv4ZX406J2+Nvf/oYBAwdgyJAh3DWGw01ntxiWzXI2fA6cCjlp1rlVP6uz8S0haax7lPOb02e95jy+YGRHtA6GsK46wt1scqn9VcMX4eHg3D8x6KvB6gruGFKxCocPa4Y/nvAXHHvMcehe3A47UxC2bduGh47zo1rH61pqtFEKQM1zglouuc5NYLl44vv5jiPu92H0pqPfnjekXQq3fpbXykrEG8FYWzP4RHCGo6F9VriEPxVuwQi/1dHKsjx2Lzh3sOeWz/0s/lo4RLuCri1vMtyCE4x+ypfRIogyWRDiZD/BCLERYteWbgW7Y3umOBleSxeMguG3a8VunO5rPbNgeJLFbjjhEJyCG1/8tTtPMryGR7D1mlsD8Suv4de1gu6dDkoVz8Ot8ivZ8ceOGYFJkybx6TKc/uJKzF/BTVT9nOxlB46ozDrRo4WUmChyUDX4jzRAqZbaKMHDuJqLIgGfVqe5kWvPPJw4KjYCuJQbuq5ZvYKNTcd7UjPjsDPM800YxfjLVeQYl+vuSUky3qu+4p9iCX4fh7drfvoREx5+EuP6BrBtew/Wc9t+PysX9pAuDtHZzCnA5FLoQXFeBeG8uHq/fnhoVAjNXzoTEw4/DN17dEd+n0G4/Y6JePe9eagor3Q2dggEqdMyn5fapEfHAFADVLkKRodoUbBYtMffx7cZg3EA6/Iqf7JfPD53WYlwNwS/cAverbCkw688FpLRbOkGZ/h13xj4hVP1SIfbzRvBpla5jPq62JDHJf/sNlO4n2VsxITGpKkxcadjibqfXqr6XYRSRfGxxx6Lk086kd5yi/DP18qxotKPPGchRHN+GwKtaTY71HV3Chpqo7Q91LEhWRR0J++Qi6Ju3fHa1JmY/uzTznxlVZRCk8LJ5zjVxSgxgSxCUvFRz6zxKpZ2oPDeW+/j+3kf4MDhO6KIQ1+vDn5nszZR7+CngPZSeNVwJbyKJ81HuIdh15aFOGTsbnjmqlGYcc7OuHZIFKeffiqG7bwDfn/gEZg0+X58sXA+h/YsKzuX+TmXSMGtso0Wi42uVPQ7xCb4syl5EqBJmCTcjY0/YcFbKLExaTcSGyQALVNTvDEHrANsnPrL3amhaH5KQkH7xWmOdPxFl6DXtv0x97OFuGUO5+QodoJxJ8pJC9xSQYJQX/ksDjXLazxo08yLW4ZrP7fWOP708/DJ/M+Rm5PLhZMQzXOcQXhd0TERmo4OEzyy41M5quvKVStw1fUT0LW1Bzu05ao0DcR1qFS8ZmsdSXkpZjls9lNTjKKygvPOVT+hZ/MojhraHp9fMgJPnDwULb9/FqeceDz69emH6y75J955ew75WoOiLBr20CVR5QuPvjax63TUNz3fWjnQJAC31jezCXSpo0swaFK4E4/WvPu+uxwsD8yuwNMLuGDBTVPZY6kVxLQtnQS3pYI0LW3dVenNpv4V4NZZVditRz6OH9XaKeKR++9GVWmJM58mHXDD4CMzASgkjhB0hFiM6pmvzsTnn36MS/cfihbZHpRwoSO2wq16xYSdItEm7ThWbx0GH+K2/zWklrOXXlJN4VbGSfcgD5zauYMXNxy1C6afPRzn/W4gbrztTuy111j844zjMfuNN525TH1ghEudx/n48caphYu2GIVNf7d2DjQJwK39DWVIX72WQwEYm9COYOQuu+DJZ54nhiU4Y3oZPllei4IgOz3PF3bEA4esWy5wiElhEoxUOAssoVo/cilkThjMHZU79sTd99yD6TNnIpCTT2EUE0r1QqqOiFSadExU8y/NVlS/pT8uwRF/+hP692yP3dtXcWNWDYs5/pbA82Q5c3ZgHOVScTbnJrP9XLiha6D2tOHOYgx+ij8el8CPQJTzk9l0IwR31q7kHF9VTQW2bVmLv+6SjXfHD8UV+w3GlMdfwNjfjcP488/Dfz/9lB8T8o5zkBHuwqOV9Sh/mgu10CQLjRNbd9wkALfu99Mg6kyAWCwx84dx++PCCy+ifcoiXDSnGj/VBJDLlUzN2dVyr0FHc6krRXJhUwyjY9k5LOWCRlRClUIlSESltL/uWBShgbRWUYHrrr4EP/6wGMEANUCusrqFhAnwDbTHsNpfaXa0ZKTGxiEoFyWmPPiA8+ic0T2QR8W2msQHKNi1pCJcnBFwtD6vPwc/hbOxvJzWgJECHjLfEoW0mwz4tYJMuAjNeSi4OGp3Zg59WuQg/aHKMpoNVaJNThTHDQ/gzfMG4x/7DcB9PJpgZ35YJt52M9YsX4qC/EIHB0+0Z30oDMUD/mLCnfyUqtgUtloO+C655JJLkzU6N9WaW3IalrvVugFc18KnYYLhtdgFstGl8Ap/OjjLJHMJt51eunwaEmZyzoTht8NndJ8Ot2DqD1jJgDeC3xTeKF8mQdqRzpkQ3WF2Zj8FRv/+/fHJB3Px1pz3EM1pj6GduYkz+2hYc3HOMDj2HZQwdAvETMpLBeMcsMT+3725F4vLCjD7k6/RtnUr7DxsWL1g0JBd799i0a5rhXjeaxgboPB+9+13uPHC0fjTmAE4YaAHFVzYALcA4w4NyPFUs16sXNTHOccIvlrrx/Dr3sHkt5djyuzFWPb6dygv5hkuNBFqyzOAsunG5xhES3mj8A6Qb9KLq0QX8hHgZtxltCNsUxDA0Ha5GDewGNW5Rbjl/v/gRQ7Be/Trhy6dWiIQpbbpITxxepyviBamYj/VJVFQm1TbjK9nIliliTd6vwqZ5Nma+qzobuw+a2ZUmfDG+qzT0jL9StkLTVeAPU+H155bnO7FGl6nBWTwJx6v0Z8sazx+5XfjSJbP0tPhF1x8GZY3kzhV3vhnUjy8tJELcYjWmkLn9kn3OEXcPe1LzPi2CrnUwlg56lTUBJ1J/UwoaBhM7LQ6Nnx6gZwxnJunFnbFRRdfgvfff9/pzFpAkFlJIoEXrzdFKVSCXID4aenXOP6MKx1CThzMQ384xK3l6myQdoi14FA4Stc8CkK/L5fzkLn4003v4Mabbsd/P/sS9z/xDFpfcDlOfWg+Rl31Os564l1MXViNVfQY0UdJw9oqytIwV4uz+XHwgYdNZVWhgHjKK/wop5bXq1UEE3ZviadO2QX55V/igD3H4tY7bkZJVQW1Swo80kJJmJJR7jZlbSb+/bkRxD/LtF26ccfjcOPf1GvDaXEiPKmeJYJvbN7E4/ddypAJkbLMthUvewHxsSpkBehEJ+FNhdueKY+8NXQfj9N9L/yiwTTATPHrK6uvoYIbX/y1ngu/fWnT4Re8gnCrDIV4nPH3gsmEN4KzII8B4bEQj9OeiXa3Bih4KVNy92rdpg0GDRyExx9/CC+tb4ExnfzoUMAdpClA/DT1IOX8RzMPlcP3sCVCLbVLDzWxWgq5ls04J8jT1t/8cgVt91ZgzNg964Wg6BftijfwRhRsoEMklZeXciPYy/DSC0/gtmO2x/C2IZ5Ux7k/ahfa6UV1l9gKUFh6Al5c9cIcYMjvccVFF6BtqyL0790De48dg7/8+S/UirfDtM/X4u4X5mLyWz+ifZvWaEUtr20BtUd6jFRyQUXHfGpn7VpqlEEOb70R+j5TwHlRiR4tIth7YHd0KW6LC299HHM//xS77bAd+dzBmQ+MrTgn5qK1e2mA0kQUEr1TS9Nz8Ue80cci03apPqV8CoYrPrZnwin8Ckafc5Pij/ArCKfFifCLhsbus/qA6f2n443VTX1WH18PXcQ29CynGon/CHlDghmMZpJHRJkWkAm8mGwvNhN4azSZwApGuO2lZpKnofgbwhuV3xDeu3ljdZCfro8nx9WGa3D91Vfhsmuvx8G79MO1e3iRk8W5urDOQWZXj1RxN2l2srpFkkzqnhJGiwIURtqVhmNXrKGmtd9jJVi2aCEee+wJHHLogfzw8ZhHx4Qn9rEUbzQ/qF2tCzSfFqlGBb1NCgIeTOb828knnYTD99gel43KcbS+CAWSh8dxhun54WO+EM1gWgeiePb7AI6/5118+OF7FPxDaCSuHWNIB705gllygPJg5fKf8Oabb+BOGkC/NedtpyoTjxnCj4MPBdTm1pNfWeEIIkFO0dSyPE3tOXsscq6P85s+bvkV4pksS1cD51//JuaiK95590kMHTyACylVPEaKO3M7CyMa5sYEl5tfDW03DW2XDWk3oqsh7bKx+2xD8TeUN8Z7j84FVkcxyeh+Qe7rhvivCp/7XOBU+PXst3gusBqbeJ6ON2rE8gHNNAiv+VW78zgaEg83X75uFcafsC+enfEZLjtkEP48MBeVXPUEPTsk+Chn2HEpjDL6LLpLiL8mIuKjFTG36uf8Medtc7no8OwiP0558D20Lu6Nzz6cRRe0Ysin11FCiaKkZD1lFA9T5zkongj9t/10Twtk461pT+D3hxyJYK9tMWdcAdpyRF3Kw5q00UMNh7uBaDVXd8mrQAhfrS/A7hRIt998I04582ypKBSyMZu9WLskbQwa8oq/a9aswdtvvY4pk6ZgOuMxA9vigj23RZ9m1Vxd5uHoPA8538s60BukRnRphUXV00IOV5OLOFG4NNoMN7y2FI+8/g1en/0WdtxxMBlZzvlITgpSkOu9hGWoznLV+dRhN+VcYI3ElD9dULtSn80E1nC5zwVO1y6l0W3KucCZ0CPeNOa5wMJvPue+8ePHX6qOmO5nE+upGCNG6rl+ptqrwqlwK48aofArn4LgEwUrW43JtjIy+ERlGLMlYPXCMsUv3GKS5U+E252myV3BZoJf9GbKG5UhvOKNBaUlCla26E7MG9LHerVuno9ug0bj/nsm4fXPV2Bojzbo3TpmC8f9WqhBUWAx3hJBtEojy5aQkoClDtalMIJloRb44PMvsE27jthhp52clWenSxO+RkeAsvggqxniewsEs/DZe3Ow534HkKQWePGgIvRsX0B7vAryxk9N0Ys8HQilATzNWSL0+z3j8XfgbbEdbrz1JuRzibiWgkd8FG/Ee11r5xfdqy2Jv/0GDMKIMSPRt9d2uH3K03hwzjcobtMOfVtREw2GKQi5kUSA84ssx0LM+DxCTTGLyyUh7N27BTVRL06+byqGdu2EH77/AavX0Sc7p4CbJORwqiE2irJ3qHbTkHYp2hvSLlWvTNul2s+mtEtrd1Yn443Feq5nydvlxrJH9Cqoz+qXKX61eb1Ly68yU/2MN+k/JVaTTYytApuYfaNsyZi8EZDrRmXHl99QHC50CS83B3983vgC7HkmNNvLFg7LtzE+DkX9Naguq8GQHp3w6FOPCRJnz16Pn0q4kwmHheUeqlWOCcfGOTf9rs4H2dm5z4dSrpQWctX16IEagrbE6eeeja8WLqDvLjtI3cdP8iVC9ZOznsin9vfFvDdxxOhRhPfhtiO2wZBOWVxwoGZJYeflRqgFYe70wmG7zG9yOU93zwcVeGNRFFPuux3tWrfmEHrjVVbxxn4mHBRrt6BWrYtx9DGHY+EHr+Hc0/+Gcx77AGc+swir5EqYFUC1Fjg2CvJrpmbr04eD8480/m5b3AO7FYYxjltz7TvuQIwcuRt2OfhIPPrvx1BaWlon8GIfV/d7yuQdb1R0A2/S4Xc/d9OVqJh0z9153Hjd6cmuE+FuKI5kuBOlx7/RRDBNab8KDnBek8OxanqDRGjwO2o0T5a7YgKWLfgK935KuzfaBGZxfisWpOXENB1O4dVdNZwJWlTR0FsbJ2RHymiiQk8Rzvn37xDESXt2dhDef9ddPNEztjGpla2FmPycAN6f9x4O321vLOGDy/7QF+P65KCCQigYrSReCk3SVlObS+HDLcy8NZi1JAvXvfgxrr7mOozgtmBhzvfJtCU+qEPZT8JP2ok2i6jlfGN5jQ/tt90e/+Rq9dTnnsXz31Sj/1Xv4uv1+Rzqau5ZE4GxIA7Jp0VDb9FUQ9rGda/GAwcXYOElw/H5xbvgiVN2wm6hj3HyX0/CDXTbkxDU6rxMlJxxdB2uTedyPYKmi03gQJMA3ASm/W9mkbmL/B+q6a6Ww/k4D46ntrMPV2Nve/kDzFwUcubPtCYZ5jxXgBsaVFNQBDke1WrupgY6oTE/hSCHwlSROL/odRYG/tSPTa9TD1x/2214e85bDvoQh786YyQQ8OPNWbOx+27D8S3LP2e//vjT4GYIhssQ0jCQc3DVFCLO4gqHwQXUKheWF+Coe97G8F2G4cQT/8LhsIZPFHQsL5kGIcFnIUr7Qi+Hyj56r1RyRwfqjdhjr99hxZfzcPLxJ2Hk9bPxzrIACnPzHIEpHnFKk/DafIG15MpxiPWUJ4wWY7Jpj5hFzXDnDtW45qD+eOCEPXDLrbfhuWee5cJTgAsxHJ5ypVqCL0rXGCqqxMtrfTDq6FKcSCMympvizedAkwDcfB7+72Dg25bNn156mAsThUWFuP6GCejXCjj+wf/SaNiDZn76yVIQhX006qX2VMXVz5h/7aZVMzZjSe2KuGQYrRPZZILQvYUH1+wUM7u44aorsJYbjwZoRVxJU5NHn3wOe++7n1PgGWN74eSduRFouJyLHkBlVhFyqKlmc+GmivZ+wUAlSrjXwz9fX+fA337nnWjZorljliQB6Qg/xu4goRIvWFRHubQJkpvlMx+1Sy44tCnuRn/g6/HQYw/jkDvfwts/yBwlH34KOi3y6GwUbpKFWtIkTxTNEYYpeKspzCIU5uVVNKuiPWGPttREe/XBKaedisefeQ4/Lv4WJcuXOULPx8UYDz1TIvzoOIvGpMGEoJvupustz4EmAbjledoAjBt3zAZk3CRQLQaYaYSjbdB2cVt6Mkx4dCbxVeHGOVWce8tGYTa3teL40h+p5NZWmXvopCYqpkXKJtDDjVS1Ivr7HkC/vtvixemv4ZUXnsPalT9i8m034K/HH0NUhTj/Dz1w1rDmFMjadVo+twUoCq+n3KG5TlYusr1yPcvGvfNq8frc+bjrnrswaND2znGYokXCTxPjBEpNWtxTrdXqzfjoW1zNvDks66g/Hklf5lk4aOK7+GGdB82z6QecU0Qowy3RF3ufGvpzJtPZlzAgbZVnprQLluKNA1ti1qkDsPy+P+GMI/dBm469cfHV1+LV12Zj+Y8rHC8XH429FUxAKzYNNpZm5TlgTX82kwNNAnAzGbg52dVlUgW3eLROkAo+3TOb71KpWmGLsHOtr+Lq5egRuP6aq/GfOfMx/VsP3lwRwLTF3HSAw70qqSRavWSmGLWiyk1ZulI3fh6gvUs1hWtVuJaGx7k4c6CGoQGcc/SxGMXh6z8vu4Lj9E645aieOGuH1hyO11A0c9hLwRzlPF+EQ/cyX3PkVFdxVTUHT3wK3DrtA4y/4AIcdeyRFBzUcEmv1XVT+aa6yqjaq+F7mG6FFMJjRu2BN2e/hl0mvIl5q4L4YiWFI+cfTfBtqCkFKHuWI3epITqbLbAW3bPXo387L44evSvuObAzpp/WE97nLsReo0aiQ/99cPPtd+H7xYsd2vWhitcCY1zfdN5voK/pyjhQvyO0JWxuvKkNbnPLTZY/nh77siaD/+XSNYkfs1w3DcUxlq0jQHQ7AkdjMQa3JlAHkjSKr3M8oGzidGi4zDL8NH72cD7t+JNPxogRu+GUKe/iipdWYnT7IGrpgRLgXFctjYFFizYNkGajebpN7Yay+dNqrUxuKihYhvVsi8P6F+BHzjZ+vOh7jOndBi+e2hGH9AlgPVdM/NT9dAZxkPNqAUqUqDRTbWSVHcSjC4Hz/vM+d8HeC+edfyHd+zQ/R1odxsU0QAlCC8bDdPyph6cmp3WhWk4LhPkNiPCUveG7jcRbb7+FPW9eQJvFtRS2XFRigeKHzHx0pXJ0rfm8Wi6QRGSuE/FxASeAVTSZKeUZLhHyvW/7fJz6h50w/wqaJR3owU2nn4bOXbrgLi4MrVyx3FkxdtpB3cJLLd+T0yrq6md0bm7slGFM21xkSfJvTr/bnLxJyKlPrm8dYkKyn6BFhBFi14lig3XHuk6GW+nxIRFeSxOsrt2xrpPhj4eze8PnjvVMoaH4Y7lif9344q834Bb9/ElbYfVlUKsgkw2l6V47iji7irCuNhkudzsN6ew+GX5Ld5Dyj/FG6SYQNBcXu48tFjQvaoYHptwPtOqKQ7arRZt8Tu5zbiqPniI51BZz6WoUpCubhyu6EXppbPLhSnQrizK/8NRwUaOQJo6HDutCKrX80hJ/37MYQ9tHUFohg18/U7NYHneTYUvVZgjc7ZC7v2Rh+hIfzn+Yrm4M1024Fs0LuRrsrCbHTr5TPa2uDpDrj+ptwXiTOOaUAYWOl/Rm8WMlGhT69+2D3fam1pYv7Y+LIZ4Kfkg45yfxRI1R5aqIWCnUpKnt1rIO2nhCi0sUiQ7/tFuO9iEsCK/E2G75mHb5aEw+enuczI9R23Yd6PL3Io2wiZWNJMRY84o5XG12ekx9f9X73fCOrR4xSjf8tTaRKN4ARZoz5I07j64T4bU0N2wm+N3wdm244mP3c103FD+X5mLB1G0hEBPdwZgan+6GcV8Lzjqpro1owcTjdz9Lh1+wCoZfDc2NPx63YC2P5VOcLLhhRb/NlyndfqloTPVMZQpHjN6YsAuyQ1WGuPoY1IIDTTGc1VauW5JnYRr3quGHJRz1PphX9CQrw2iP4d/AJ6Pb6mxwurd3bji7dOmCA/bZA+0q74cvqxuWrKjArC9+wJLVOvIS2K5zCwwtzkWrXM7JSZUTXQ0M8qSVzZ9WciVTouFV2G6bQhyze1c8OPtbvL+0I/q24eoz604ohxdeGlIrRKlNZXN7+rcXleLP965w0ubQhW377Qc5dYm9r43bbzJ+xITUxh9144MQu/lm6WxuTvj6q6/wxrRnUDhwN4oy7U+YR+FMI2sOybOrf6Ifch4FN7Vracqsh+j2yHdOgTzT2xHnfJpf5Wq1RH+kssI5YP6AXj4MvmgkJr+5Avv9fn+cP/6fGH/eWShqXkThGeHwv5AacUwDr7edZH5S7PxN9MfoT/TM0oxPqrf7Ws8tzWB1b30jHW7BWlB72xr6rJsmvx0vmaoiyiDHZ8uo2JhklXPHeiY3HwupcAtGWo2OwFOwMpybuj/xZYn5cg8z2HT4pRnI/ioej5WhdHs5wqn5sUzxC16rmnJRUtB9qnL0XEcnSgNTpymgj2xlRTnm0zPi2yUraRrBuS6uwvbu1hUdOxQjv5CnedD0w7Q/wx+rM7sR/ztpdZVRAyssLKy7U39LLqT0TLwpKSlxGnQ159VKymmCTK1s7nel2PeOudzQrzMu56amzXJonFwawXfcV681T370sROGOb8V6871xaW9EK21cpFz8uZzbq2KNnxhHDeoPQXgUlz49BKM7LYNT5ejZR7n3fJDpVjna0cBUo6i3Cy88nUYx963kuWswOuzZmLo0J0c+lUXtYsY/g2dzk2QYDa9XYrPNcil90Zl3bvmEXM8Fj6EhXS9a1aYjVahZagK0v2M9dNKsjPVQGNtabsqW7RtCNRmqe3T189p/z4KSi/f+1qK1A6+NbhkVC526bcHjr7mSnz5348xgVpurx49uach7SC1jyERCWeyoLLUJq1MxT+nIZbb8JgsUKqlJcOv/pKqz8bna2ifVXtvzD4rrxfJNL86eyZBlvISVMaYRMxUmhijyrqZmQ6/8sUMROs+sUkyCE40aNcKuRBlGtTJJaSsgySiXbgMv9xkGoJfjDQXK8NjDc9odJfp8IZtV3sSf/jxp7h6wg146t+PGGh9vOPIPXHkuN9h4MCB6NW7tyM4VYcq1oUtm/SS5joRJMEne7KgNvvM8J2qIPFGtKu+igtqynDR0+Cc3HfYt187TNh/Gw6H2VGjtG/jzsrrw7lcGJHWw/I36tD1ZKe8oO6gUnmesHbnyUaIK8Ll3KK5b/My/G2v7rj9lS/w7PxWOG1YGy4bVFH4teBuztyWilrV059X49SHlzLfckx/+XnsvscorFq92hE0KlT0i8/xweEN26Xq2BDeCI+1S4kuzX2WU8uLcIcbhUCgJaZ+v44fJw/pX0WzGbZ9Z+cYbrdPrU8+zJqzBOf7NCe4IYhGLZQwFi8p9LxUhyOaEuGu2us4rA7QD3p02zJMO39P7H3di5j532/w2dTH0KV3H756jQRS9xWVlWmfNeEo/9tE/NtA94Yr9UP1qXTwei7Y2HZjmfdZCcBfos/6rfLxHXZDVWNXqoipr/YsvvKGyxqcnhtMMvx6rmfq2AZr+ONjwblpsPIElwq/8Fo+K8NiK8Pyiw4Lht+eWbo7Fh794unfGL+0HglY/jinFOZcX4C7s3z00TwMGbwDDt4GePr0odiO81iyB1tBW7ivl5Th3e++xhlnnukUd+hhh+OIPx6Cbbftg47t2xCOc2M0LnY8Cmhg7OxEzI4k1SCqD1VdB9MHSbRY91PXs2A0SihEOKTNzy/C7qNG44UXX+RKRRQnjeqMZnnVnKcqpVF0M3i48Wi2Zy21G8ILvxuZIU0TaxbPw87OwT3LpD8vnX65xoIKGj8f0j+PAjAL10/9GgfQ66MzbQWjEZpiR7Iw5ZMKXP3Uhw72GbNex+g9dqfw5kCTgl/1UD2TfbTsPaoN2Lu0utt9PNkOz1ztUlUN0dMjx8ezQzwxbf/SD3zoGSnBWTsVcvdrrtpmNediTbVzCHtOMIC3FlejS5EXbZrzjOIaeoA4QsveBWlWm6AAlKYoOiJ+zidypJzjXed4uKynYB1YsAqvnzeShtiv49gzLsZTj09BS87XmjISq4eoszccq4nSrb5WVz1xX+ve6q/2a9cGY/eCcwd7bn3K/Sz+WjgEp5+C7i1/pvgN3mIrw/Ibbjd+e2aw7lh49FOd63eEtsRksaSxGlm6oPwKtuuxVTgZXsEKxoaQuk8XRHj87rKp8Otrkkw7iC9LtAi3ylBIhtfSBSPcapBJg4Y6YjbbaA3n9PLoGF+yei267f0nHNWvBJceviv6aNGB80R0rUerQAR9aJw8okceDt+xG/p0KMILz8/ExEcexx2PvYxQyTIulPAF0h7Ox2GZeO2li5s/QM2BtGsDUke70It2BITVQ3oIG6AzL6U0GkSzviEeWi6jYW0Q0KljO8ycMRN9c8tw9oh2zk4omqwjFGHYSXUtvLHXnLTKyR7Euj87P/Ue0aHdkyUQNdxtWZDFLeazMHvBT9yOKgu79mqP1aXluOKdMCa+/JGD8s0338TI3Tn3Rn7LpUzt0tqY4lRB76yh7VL4LcSwcwjK84IXLVyA516eilsP6kq3vSjnauXhQloInEO/6q9+CuPlhaXYa1t+LMIS0uQ3a03u16Gro5VC0dI83N1G2+rLaybMXXQCbA8hfh3aFXoxZvuuuOrhV7kg5cXI3fZwylHHl9lP7F2Ijxt+KkR9amvps6K1sfusfQDdfEh0Ld5Yn81cJ1WuDIMKbQouDrCRO+KRf4IUhBruvPTqK8DX7+FcGsQW1vyAtZwWo8csF0KCKGHjz6J9np++rB24P8EhAwsxstdwfL0myg1Ff8D1N93m/FTC4OFjMW7UcBS3akF7My+69uiBTl26UBDSt5e/MIfFuTnZyOMvRCGTlZ1DrUliiJ2NNnkellFLjVTbVcmcI0ij3y6du6H068VYuK7KOTujkHi4but0OletNuvSsKnz+6P8gJAeL+N9e+fhcqbd+uaPiNDrY94P5Xhn/iLsOmQIJk2Z4mzvrw+aPlDxnbsx253EFdnmzM82a9GK/sIDcNm6aejYLIfmOuX8EFVRg8ulBliFcE0I176+GuN3a01tjuY83K1ap89RRBKLalwn/DbiIJFziC1vGfEiZp7ENsHksuowerfy4qEThuDoK67BXmP3xbAR9HV2+EBsjhDcMHLZCO1v4GZz3nujCMB0X+LfwDvZuIrOB4EN34k9WFdSiolPv4xJ+9PqJLsSpeXcCozCK8K5HZ76gFyah0jnimg+iMMgDzfXbOUPo30xz6Uo3gbHDm2LpaXAwpU1WLjiY9x/+XTO2MUHrmRQoMp2I6smVkCsAABAAElEQVRHB2zDxZS1XGjaqX8f9O/Tx5nIl2eIbNU0nxSSRknNcflPP+GlaS9RUhfhvvdX46AB7dE6h7TRFS2Tead4KtLdSyvSeR41FP9+zj92KMrDBfv2wtUvLcAdM7jYEV6Oy7gxwUl/O4O7Osdc3KRNSPhtTsNPR1ei51p1DWTn47uvPsfdN0/A3DO3o20gFwfJPz9XcyPcxzBAbfy9JdTkOZTt3LqQ0wfr6ROsQ5sc3ZvCL3lgrWhvGNP8tAN1Jf2LS1nXLHrk1LCNjOyUjQMGAWdedDlmvfgU8vJ1wl5suiM51qYnqTjQKAIwVYG/xWcamWmiWwMgzf2Vc0X6yxXr0Gtgb07zcyWKQ6domPM6dJeKhOlbSh/XCsf0jTZ4dP2KcOK9kqeX1VCz8HIc3TzXi2LuEzq4Dd2s+nXHOWP6csdjbkJbEcLayhDKKunSFtbh314KtBVYX7aAJVPYFVDQzZ+Pr16lZsGFYk4jUqVhzPOE2rUDirlwv2O+D0edPBjFLQJolSOjYy5FUDuNzV81ztsLURAHuBIcYWfX6vK+dE6+eh7n2ZYvxm133ImTTzzBGd5L45HQ008fWfeH1q4bUyiyZE4EVuH5aa9gDFnRrGVrThGs5/vTjtE8WrN2LWc2s/Est9o/oE8ePz30YfFzSzDuUagNEjiHQJoTi0DphWojbAj4JlyEguoSFLfMQRHt/nTYUzWnOvI5QfiPfcdg2FWv4vMvPuemq0M5P7rBJKVx3s6vG2uTAPzF3i8XP9i5fdTqaijUBvpK0apZAV26wjyBjNu5ezlJHqIAYCcIcUEgl3aAmnINcxfn+kETNweVo32UixFc9nDmwPw+bvHOucA8vsk2zTmP10JHXnJOkKuRylcTac85pU7smLG5Vs23SStUcISJJCDhpFVRCeWVXL84l0UVJlxDjYSwMeEnuMYILFsygeYiXMWgsI2iPVedLxtSgEuoiFZRa/VQQkf43Ecty4KGwBJ6JgwbQ/CZUHWXWbF+LSY99Rz+cUgn5Gh/RfJIrnc1FOJeDtnL+bV4eEE5hrWmHzVt9ioov6q500yALnPOFLoj/ygIOZ8aOz0uhp0p3FiB+j+nAbpmVWDiZyF08i9Dc08pzyfhfGlBO/Ruwbnvwih6MMsTTz7lCEDtnBMmz7yOcK3TivnBNd4Y7XZvsaX/1uMmAfgLtAA1bhkA59VWsBOwM+QVoEfnjjzO8WPUZBdx+EkhpkZLsVZDgefhoko1BZVMTXxSH9VZKJzkvuaVBkHVTRqdHtXQPUsPKyXU2BF8tK+TJ4meSUh4NFTkAFPWGOwjTJfojAVpNMKjoI7hXOtWQ3VH06oTlI6QdMAa6Y8EgioYo1v+t7tvy+H7a+3w93+cj733GYP+AwazXiTOIY3wqmBdaAzhJ9Ru4aohd5AfoAXffIel77+L3sN3pgULF0i4YCEXtaBW4unx4dd3Yq2256c6TRL9tNnTK6OIIj4tWgjzBtp1Z0Gp0rm9fM/HbZeDflfPBzoMoKpJ3syfxxkNxjS1UXj7DW4gMf1VnncyEG3btHXSlNuZFyQ9ol0023SBO3bzri7jbzZqEoC/wKtn+3cE3zpvEZr5uPsyzT+++rEUa3M5zGnRGr6aSq6u0l6KZ1v4aQzr51e9VFs9cR5Ie+iF6EMqc4kAV3md1Wln3kfaj5o8O5i0Qq00axWZq4YytYk6TvzUqiRYCCe9kDkIV9f5lFlpylcX1EGdZEv4xWJpsyybGrKf6mCYH4AezYFL9mqFy55ZjhkzpqJff05+1RH3S3VgdzkSKNqppbQ8ZrDfIY/vi3OXEb4f6vbOT+TpMCdUl3IRqp2zmFHLld0sDl+rGDseHHofDG7tz9gsTVgafoQfsY5Z5bjniG1w97o+OPvQQ+hHXEZXuGz8tPY7lK1ex91jVmLvPffC4IEDsMeoPfiR2AeDth+MVq1aO+hEu/1MGCp2Bz2PT3M//y1cx97GFq7pb52p8eyUDgbahxVxzm1NWRjnX3YVeq5/Af06dUW0ch1XP7kKS6NZH3ckjtAwuIKuGC14HPd7y/glD7aiH24+PHltsKIqFwtWefHJ6ly8vzwL7y314/0V2fh0dR4Wrc/GmkghQsFmNC6WPSH9d/O4KslVXR+3j3cc8zUPpU5Q1xFiq4exOTW9s427R3wtGvM+1lmlJTkedhTKGi7u2VWTlG1w9jn/xOJvvnE6q3Vaxb90kAb+048/OsU6R2ZymsAZlPP9yTxIQ1itBB80sAVmfrGOx6l7ucVVbLVa5kha/U8VHA8ZbqIqc6i1bAeDO7fEvJcfxdJl38FHz6osav492/TEgO13xNi99+Epef/CgYcdhQXfLcXYPffGTkN3xuTJk7F06VKHVzI7kRZoPNM71r2Fpn4qxWALBzHbGL6FUTvoGrvhbwr+tHnYoT2aq6kso0vTrVg6cRJuuWoshzorUaWdhbnvXJQmKvINDrOBFgaqsLg8H4fd+y7+sSvw+WxAVnA/pGFo97ZeDNimK4dhIQzsWIDubWQb6KUdWQDtaF+XRY1SXzwZTcuAOsQzdaURSvDVyURd8d/PhUtjvlNSQTpIBU1GnFVpCoswNdn2zXNw8u/aY9LLKzHztVk4vrtmv0R/TGM0vjcmbU6B/CNhoeHlmlVrnaQqcFft2jXU2Lg4Jc8MauwBavKtsqMYQfvFM5/8DKfu0RI9WxZxAYs8Vr8wZIpVX8f3253INMonbc/v4Wl0zfIKcci2wJyvF+LAzt15eFMlqiL8uNF7J8wdZeQ2Wcz0P3bvhf15aNTCBV9wN+wTHYR33nkH9tnndzx1r63j7SOPH7OTc5eo61+Cf/Flbsl7awebgtMvV59MEMhNRsOUdMwyXOvXc+NKvvRMgoZ1mfreCp8aoo5+NHpSlSEaZPyqY/AyoV0NXQakOow8E/wqW7698ZPyVpZi/Th+xW23PoA7bpyAdy4aw5PMVtPEhcahnPdzdl+OsGPxUCIZhIS4KPL4h4vQo98wNNv/EOx/bAscU8SdkGkc7OE8WCV9dpdzN+Ec2fRxx5aysnL8SM1k2qvT8cz79N9leJJTRu5w5I7N0LljFwzs1BKt6XHarnk2cjTxKLmjITUFonYw1pb12r3Zx0WYWvqpxgbOfE6+yFVLRtFRpVPrSSQo3WVmfK2hN4GlAdVdUaOi7aKnEuN6BDGJT6+mKczIUWPQrq5D27uRb28mmoy9B/k9ZxIEL4Fh7cbuu3br5GSX2VIuBVU5V59akmuV1P681dSwqb3v0k0gHsz4vALdhjfn3KvSOTfLd+3n1IQ+crXU0HlZH9RTtN2YdsrhUpnjCkfLTXTtsR0m3Pcivl+0DH3a5aNHn37IzS9A69ZtHLMlGUqXltE7JZiLQUOGYcpDT2DBgvk49dTTHNzPPvs09uJQWUdY1lAIynNGWqDxTLF+qp/6bKZB/G/MPit3VOO90ZqINtGt5+qzDZEJkmf6+WVRLyTpgiqsXyZB+NyeEcnwi3B7GZnitvLddKfDrzwNwW/0K06GWzit4STijeqldNGpjvTum7Nx7plnYMqpu6MTbf+0WgguePhq1tM4ej1PNsvlxuketA76MG9ZFLdMX4FpL96PvfYaQzjOK6UJ+iicfsYZKOMHrYQN+ZtFX+EHbq65tqQMs96ag0dm/Rd4/+N6LKN7AcN7dEG/9oVo27IQLbgLdCEN1nI1D0fBI28PLcRUsYNpbVrLONJhZJjr5wYFtRxWK21LBGt+EoKUFM5fH4VhKacwe7apxWHDe+Oxt77EG7Nfx6GH/nGjYZx4nKqDOMjq/uid6H0YfLJ3a+1S783ajfIqtGpX7MQLVpQjp0UOPlxciW+WVaJ1uxrs1MqDwmY+rKLZC8UZrn7le+zcrTVGFIewUg4l/Ahy0ZimUOSls63VhiGx8wEgiKYAaL7ubLWvrbMKaIQu3X/nwEos+bAalz3GW4bRY/ZBv0GD0L17Twoi2kfyw1hBn3TNU/bhfOk9HB7PffdtjOPJdKedeiouvuQSCs3WTv1VJ32w1WaMf+KF1VH4U/FGz8Qj443g0wXBN2afVfmGPxntgrF3a/R7eBgy4dMLQH1BVGGrvJAlCnquX6YaoGDli6ovua4VRFyioOd6JjcW7URh9CSCtTTlkYYmLTBT/HIilwtUpvi1+4o6i9GnstWY9FP60qVL0LHjABy2UwDX792VZhE8+4EtXfZ+qrGHB+iEObTJDVZjTU0OBl75Dk49+nj884arkM0vYQFXjZ1AtmiAKryiTS9csQVngYQ3jhM8O4Tm+DTxvm5dCVZy4ryUmuLXX3yGn1atwWtz3nM2FLC8ivfhb9hoYNB2w+mWls1jKaO0BaRNNOcvQzS1YWHsmtqbLla3xG/JjXHTriVYZRsZojmP9iN856sqHPzAp/jdmNGYdN/9jmagNqD6ql1ae3HzP75kPRPfMtVyBK85NG1coWsJZi0SrVj+A445YD+8MvcrptFfsVkJjt25CyLU/r5bvhzNuW/Y8x978I/fF+H1dxfj3cpszDujD4o5fC3n1vjwF3DY7OdHhhIxrtvp3XKbWvK7wjGI1pzhjTOXY0lJCHce3oU93Iv7P6zBJU+9h+7tm2HRsnVONf94+FHYgbvitGnX3vGrrqaZlMxjcvgxXbH8R4y/8FLUlK2E3AiHDx/u1Mfddqydm3Ycq6+DOuEfPVe7Vrs3WHsHCTMwUUPw+o0lHH4mg4wJX214ol86/HqusjVikxZodUmOPYbf+uwWnwNMVfDmPkvH5ET4jYH2bFNwWN6GxDpzQ8MNhYcfeph/1+Jvu49kw1lFLwxJFYoPCpVaTnbX1FId97HD87S2u99b4+Q5+uSjUMQGUMFhsuMs76TyD1+2XrLqYQJP1yZwJRjUGGgY5sD5KQSbFRWgVcvmXBD2Yuft+/ADUouj/3w8StZei1DZeny1eAm+/PpbzlV58dnS5biYhwBZ2IMXbVoAXYcPQMecMNpy26fuHEYX0GFfRrsSyY4Qr+/NsUHz5gyPNSeow4YCFLyV1fwgdMzDfjvy7JAZM/Hxxx9T+6FWrE6UpiNZHRRv/nuPiftCTqW079GH+4V9ismHZWOXXt3QggsXtbQJPO8/hfjXR4sx85zB6NsyG/v2zMGIG7/DOU98iWsPG4iOPEektno9txHL5yq9DJ+lU8eC+BXmSrEMwjXFEOD84DpuAjFxXilu3LsTzSQ5fOZQO1xK75guffDCS487xuGzZ7+B+++bjMf/zUObDjvC2SGnoJAHtNMkZ31pJVq05hzqnbdgxivTuOP3CDzyyCP44x//6AiwRPOC8f2ljrz/12jz311y8v+nBGDyamxdT2LCSEPfIObNm+ecV3HxuO3RvihMLw26fLGR0y6WNn/UTHlerpeGrgF2jjeXBDBx+gLceNMN6N1/KCr4RQv6Y3M09TVkx3FC3ZfPGqwaiQSjhKLzcybYtWefXjE9RqgVUGI4nU7G0Pm5QR7xyI0TcrthILUHTb6rO1ZRS7nl4jOdPfaqeDxkkCe1rVz6PUaNpdBhuO64PTCUZ3OEuYFnLTdg0OKFX0a+XMxRfo9jFkJtyzH30LCZwkeCyqGnjnYHU+o/UXpWeDR3xhmxID1S9uemAi++D0x9+SXn6EsfebtBfMRwNVZHMS1Dpcx67R088OjjmHzcIBzQOwfrtEkI6//J0giF37e4/4RhGNQijFW0gN6mZR7ePasHDnhyEY56YD4ePqY3tqGgpHMIa8VpIuLjZ8qR41Ht8Uf28C+HyTx0iWcQf/P9aqrzq9GlVTtusBDEd2uLcMWrH+OG685AZ/pryw3u8MMP45EAozFnzjs45a9/wZOPPYqzzj6XdpM7OD7glVU0q+J8476/H4du3bvhyCOPxKefforx48c7c3jSBNVeVMdYW1J707fFRHOMt7/WvxvGT7/WGv4/1EtNJ8jhRzlX6u6adKdDwe/60neNUsYrYcHGXsO5Gm+Yq3qeHORzmmfx+iwcedcbGL377jjk4EO5KCGDZWoWEhxskda5JUJ0HdvlZYM2KOGnYHBayGBp/KdhMO3UmFGmGzI41gJklPhrOL8X5k4w6oHayl0NPycnF8Xti9Fn2z7YfkA/bNenJ/ZQB3vjDQf/j6tKON9UzhVJ+r5yYp9T+USmzkNapBFq3kzGITLGZQeNsEydq+FxZvzFGcKnDdKOtKEo94TjcFELMgM7Umtu1wF33X0PFsz/L1fVJXxj83Jp0W0GgAkG8XV9aRnuvHsy+tNGcZdtaMBOH+1q8rCKVXrqU2pmvgIM78ihHke3fL382FVySy9g2ondsE+fNjj9lrlYVqF5Rdpzcn4vSN5wVM4PItlFXkVpCFjLj04uebu6OoDxjy/B6AHtsF1xkWNk/eL82OhgND9GWkiRa2Q1319zbs5wwLg/4L0PPsY111yDm/kBvebay1CybgV3HqJpDt9PeWWYc4NDcN1Nt9MSYQJO+MufsXrtakf4SQg6opdtQyMXC78FIRjrNVbjpniLcCDWhDz4+I1Xce99U3D9oYPRMZ8NXFs3cfFAK6l+nf/BXV8k4MoYP/oBv/YM555/Plq0aOHMmWhFVNqkghqjCTcnYRP/xMSPCVWWIMEpyceyRLc+/NbpnZjlq7MN4/DpqSeewq0vfIhJH9LzhOcGZ/OwoCg7cpiaIDemoiZIgcqhVw2H9SG5b1EIBqgR1rJjxyzmKLgpkNMH1pXDQO5dw5VgzsdSWHeje9yVu3dwsr789jzOoVY6NnjWSS1Oj7vhEIZ7Oedypz77JI4euR1tM7nySGFVwHlSTR3c99YyXHtAW7rHUYOj4I5QmDSnm+Iaat5dqIQP61CLuc1a8DyTHH58uBDD+c0QD0qq4cp+NfcX1I44PtYzK5cDXb6EyW8vxoKyEozfow1aBCP4fp0XV7zwGa6ccAu6dOtBg3d+uOqCFjM07VFcXIxTTjkFc+fOxbY9e+CMU/+KT7jnZDYXUryUyCWlFdxQowvuvOs+PPWfZ7g4cgpWc0NZaYAyO3LmjH9jEuFXX90tITSsoWUSq7P4KVTKStfjqhtvd7KM6U5NjDudaPdf7RAsIaMJ7ih3Dymg8d+8H6pw+6wFuPifF2LosJ2dRR7RbVpdJuVuKRjJQpVd/2NdnPkpbvO074HjMIUa7Q3PzsWdb/zELby4gVeWNNkKdvqAs3tJDjuaL1TGHUy4aEKR6rjmUZRRChIPm5sqnyY4opi8imiRiEK0jJpVlB10dAfhbI2HrjkPq5YtpWUJ9wpzxHYahJv4WO9SfLCPkAyMFdrle5Ed5iax3KQih9MZn6/SB4THG3DOVstEXnpseOi9U0Hf7iKap3xT5sHh9y/EAyM5FxvgkJmDX/2yvDwHhFqiJ1TEc47pJUSbzdUVuZgwdTFumVWOl08eyn0h+THhR+Vfc5c5ZR80bn+2i9jqsbUPayua/5Uw3G677Xiy3CTcccdtuPG6K/Eqd88OcNShYXQFD5vKK2iJSfdMwWP/fsKxG1yzZo2z4KP62k+FCe+vPfzqBaBe6C8ZrLy3uco6ddp0XHfwYHQqrEUJFzNotEeBIA2IWhJ/ORR+ayo9uOa1nxwSDz7kUB4+FLPeV+OObZG+YeXdcDdWfZI1eA3nPPRhziJNhxx7DP714L245ZUFOJlHaH65hts35dJGkZpPvo/aHjWeWq7qB3l4UQVt+cp4cLmXY2AdEOTMA6btUzFhIg1Kx2BWc/OIHPpQl0SzuLuyD2ftXYzv+S155533OMwUL2MIk9FuvBLv0sEYrMWCl0DRarCE3/XXXes88sqigPTJFtLLOdaf1kswc+dn7ujNQnglbbiM54Nw+3zOl85frudR9Cpuj3L6a1P0sR1E+MHgTtE0hSoKcit/nrv58JcRDLr6Q3zJrc7eOKcPBrQVzRG8/mMebp39Pf79xNPYpn1bejzSTEmfJdbJLQStftIGc2njduqpf8Oct9/Co49MwZP/fpgLKeXO1ExlVQi5BS0wcfIDePrpp3H22Wc5plNa2W3MD4rDvAz+NHY7d5Pws3OBjYkCcl+LKCPMruNjy2Nw7oISXQu/uww3TDxuN067ttidz32dCL/lUez+KV88LQab6Fl8muFSg9SX+F+cKFcY0dWPMm6X7qPwCzuGrrx2hoUUhmxvs5dk4aOvfsTkibeja5dtUM1Ja5uUFj1Gg8XCaWW5Y6UrKM2dL5b687/xdbW8FhtuImRn03CWug3nqXiDA8eNw+szZ2KRrw9G3/QOJs74Gv9dXosllTyxjPsWhin4Zn0TwivzS2jzxo0b6OERE/mUpMSTOqg8CTVqOZqXZPnaD08rzhEOtUf1lHsc8PLLL6KsZLVTV2loDr3MqX8W6uvABNXXNDl7brGeuflh9yzaWS2V6cxxxx2HqdNnMEs7CsMVRBgghWGUU7Mt0iYFHJBnc9t7TuShksN3GvJwo9tK8gyY93UUhw/O5dkqFHTc1YaDXmfbM3CBqZT7BT7w8Roc9sh8vD//K7x0Mk/H+2MP9CrQ6XJ+fFHWEn+a9BqOP/4v2HPUHqjg8FmrxLIMsDoZvYrV/mRWpo0jtMo7bJdd8QlXzl966Xlcd83lzv6F2dncko1bpuXlN8M99z6IBx98CNdy7lDCXm1PfBIea0vGJ8VWlq713B2M3+7Ynhuu+DzC5w5u/JZuedx4Lc3y273FypsIl8HruWB/1hrdCOxasSEzBO57uzZ43eva7lVYomAwBufGY9fu2A1neRPhtTSDsXxKN9rceC3N4Cyf4VFsz1KlWQf78KOPaPoyBZcdxLm/5hSIdDkLysCYmbVqqsbF/Ze5UhjE9c+/CbTtQYf2vanRaGEh1pmdcup46KZVdLjv7dpNXyL63XTr2mDc+YwPhlMxC3PEka51kLqc+GWgPHToEMyZ/jwuv/IaXEdbtX3veB8nv7QOF766DEdN+RRXvb6ShtZFtGlbzcUWznlyo0/t+Kyt/NMHDqNlB8j5UX0sanjspBZz5MnQq0UtDhjSA489+Qw+//JLked02Pr6aBK/rl9aPayOujc4Nw3xaXZPVjvhqf/8B9OnT8ett99FOZeFxaE8am+UbORFFuNuRdJEQ1jKBY6KAA/son7PzxiiPPFPG83O/eEbtGzWxulsOklPO29HaVGeE1qDFpxCaN+yJbXaWlz5h8HYoSM/GNzXsSrYFj9UFGDva6dj9J6jnbOPc2jPKu2PTKxvj+46Gd32QdAzuTwO4I4x3377LXp064rxF4/n4k0ZV4jpc15RzTnHItx65z249tpr8fDDj7BtxsynJDzVTq1NG7+sDMXGX6PBfW/XgrNgdNm9Yvdzu7cyDC4Z/ni4ZLgMjxu/roWXdeQcDn9GcKLYvgqqgDHECo+PDcbwWpwIbx0B9Yx2Myger+4Nt64Nr8XJ8It2d8WVN/7eyjL8brqEPxluS1c+Bd2r4cziNkUKY7pxMwLOFdHgjxPdEiR0feNOMOXUkJpxgvyTZREsXl2LKZefi9bcGKGC2l+AwxDDq1jlG12Z0C4Yy5eON1Y3q38y/Fa+YpJEYcYFAG7D1bpNG5xx+mkURAtx0cWX4aPPPsPDS9pjHFeyXz2mLYes1N4o+Jp5uZMJhUWI+/lRhJp8cniU/E9M8JKrXMXmx5Rl1tJusjnnsg7gZgMKM6a9Qn7HNkkVbdJ6VHdHD1Seup/Rr3vjicVuXruvhV8r+ctoSPzn44/HeedfhDZtO2LPkX1x5/QvsKqkAvnZ3LE5koOWXDkvKizA1G+5uW11BXIouD1U7yMaIlOgvLOO+z7S7tNDwRWm9pZN4a41+moOgf3EsZbnnowsDnLumEbrbAN5+dlYWpaLI69l/UjH3Vx5bsmFMe0+7aVG7SVue3dumt3Xeq6gz42M+rt06YKHHnoQh+63Dy648BKeA8P9CjknWE5Tq+Yt2+DSK67FmfRUmv7qq47jgJOXvEzGJ6W7+Sp447c7dsMYLovd9Lqv1Wd178aTDr/gDa9iN774a+E33H7zpRRQsqBKyPNCsREieCGJD1aY4RWMpcXD2r2EhlnoG85E+C1NczLmJ5gJfh1bKT9By2/lxsd6rjqan6CulaZfsqDyZeGunzNXtGQJLrzoYhyxWw+u/FLw8YQ35de6n9dZCFBHoNtSbQ5e+3qVg3af/faiuQIFNSfCHeNpwlu9NCxpCG80jyMfTQXDkYp+dY6G8EZeNc2bt6BWx7bAIbFc1vr07o4Lxp9Lk5AwJlx9BZas8mFRc26+4KuggMjiTsZ52KZAK8S0fWQnl+1bQwJbEHnGzOz8pfS13aF9CL27tsU1PE70T0cfiy5dOjsmIcKpulq9da9rC2r48hxRMJhEvLE8ejZt6r8d+B49e3F11k+ToL3x6vSpeHhuFKeNzUWhbwX3YszD6SOKccVLX+Av/fphSLdmqC5bxc0K6NfLxZviAh9W019XZk0BCUG2A5Whemkjhde+WYO9euXT/IW2jbyft5y7Yt/2BgYOGYF5kyehFY2ZZbJSlJtPQUr7SObVEJiVdWhL9kfv1eqp/iuvlgsuvBCLv/8BN994E879+98558zFq6oadO/VF+eedyEOOvhgx06wf//+zlSOeCaB4g7iS0PbpYblDemz5ttr79Ndvvtaz63PCn+mfVbeUnXeZDFp6UYafy0GCLF+1jjiYexeMAoiTD+TxvY8USwY2SK5cbuv3XmsgkpLh9+Nw41f+RIFwetnzzOh3fBYWYsWLXKSjh1MNynKtOqoVgZpzEvfzzBXBnXmQx6NYb/nbi9T3liE2++4A21pmuDsiEw+WHd10xDPdyvLyrbYeKO8+qWi33AILhPeqAw3bzi6I37ykWY9scOWsnD5JRdh1qzXUF68PUbd/B6O/M9a3Pv+Kqyq4UeAmpvOfIw6HwGjOLNYosLPfI5pDSrAhVEc2j/mkzvn7bdJl4bBzieGsd7fhmGxSlAdrb52/3/sfQVgluX6/rXuMUaMbTQMGN3doqKiCKIIIioiYiBiHfTY3a2IiR4LQRRBFBSkU+nu7his+39dz7tnvHxu44MDHs//dx749tbTcT/3c2dJfWPjaFP7Zuw49O1/HULCw6hrm0ysNxatm7fG27NW47kZ+3EoLx5R1KPu1zgKDQmUe0xJxtwDVKGjfq4fDaiWJmf35oYR+HjuQR6R6fOFWiPpPAHQnAu5sj7YdywV3y/ORyUeg9dt34F73phL4DcPj3ND+WbC16hTvQrV7FLZ+SSgUHKALeHVNErVLDaovQJeAlS27WnpmYiMiMTrr71C15p++G7cVzSxReBG7DyNuumNm7XEZT164fHHH+PxON1s6O717C7Mc97omy3HHU/3dl7q3tt5qbjezEuVqZ/yVTjduNq4uqpepzBB9LK4oALsT3FOF9fmU1I8+01XW3HbEJvefVU893c9u/Nwx3XfK407f5vGHcfee+Zv3xeVRr2lbtfOLrPo+ezQqVOnozrfVSP256jDketHOpCsfASQIygfDv7kbG7b73B+WzRr7hTBSWj3WXe79FF1V3DXzbxw/VEa93fPPFxR/3TrTd945q/5ZurFuvkR05GYSyChYpcunfHRO+/huaefwrbtO/HP7nXQtgJV2ijzlk1ObqAxolD8PPtT5QpesHuI9VCwWgCO9vLaVKNgNMOvkybgBC28+JL25kMao9NXDllH3wUwVHd3KKpv7Dt7Vfz91O39mSpk9eo1FrJrvOzt2buLAs7H8Pyzz+LTGUlo9Piv+GLORuxLyUTz6gRse3fg6jdmYMziZCw7FIgNyX7Yvn0jc8vCTxuzkEz5yWgercMD6erAPxzzdzMNHb33eHcJus2LR9m+d2Lu7N8w4r5/ICY6gnRPR+edM4j/HMzPzDmPNqm+7qC5oAWuE4HuBQx1FU0wNi4eb3OMJv0wAXNmz+Dpg0LlfJ9BDv4VvfqQM/wdxoz5xGSn/lTv2T60V310zxsTuYg/in8+56WK9MzfVsNdV/vOfVX9RZcvDO6MCl+6bk6XoaIqDzfEdyX/060tT1dv8rYZFBXX5mXj6Op+pzRFpXPH131xcdx52TSaiCzEAEH51T1ModLxC5bg1suouiROWzYZH6T3gDu+jnA8BBogl0fLLxv2rqcufQ3Urk2zLCYfYeLmxnl2/bV1slfXpz/dehNHidztURpv0rnjGc0PYYA2aAyJ/GcRswkkQLxx0M148KWP8P2CjRjUshwZIFq8lBckFiztkVNBks2k6KtKIfwxgtbZjsEwVKM2Rq+ODfHN5Km4c9t2NGxMk/n6pxOIq1ruclR/227P9uq9fWevmzZtMhWKjY3nnGb5xObEUR1w/UAMHz4CF13cHcuWL8W7E3/BH0tOoHVgOXz87ghs3b4HD774UmFj2l10MdpGZOLZCTNx6GA0rmtel4yQXPyy4zCe/341nnz4IVzQtQsqVamKqLI8ORDTE1ZN/UXeEkCy3vxDLFpZOo2z7SgspIgbd5uczxpn0QSzaUWmJmbOnInOnTsTq62AOokNcYLHwjAes59/6XWK0NxBnyPNjd8RYWLa4BXc5dp+cvIu+a+3cT3z9yad4hQVz52XrZ37ndKcAgBtpH/3qkKUubuwfzfPvyr9mdZZUzOXgqrBJGbv4ELc98c8dH6gM7GR49yxKe8lmg/FOcQFTiXmF8Tjbxo9uE395TgeI/NDNClnw3CA75mW/1f1i8rxrJvnGAsAyeudjl0VYmPx05fv4JLLLkODyrHoUDEXx7KDuCFIxc/iut7X3mwzZsJmIZWYZNnADFxaIxjfzQYX8iw0oWn4DNLPJLN4anBBw1M//OnJbtwOFgmsW7cO5eNq0Kq2HFkRBJPuOWHcv9Cty9tmzGomUFyFvysv7Ub/LtTbDi5Pk1ORlNhJx1Dq4x45fBhlwyjHGBOHXB6fp/zwHe4a9AI+mDvXKTuyBl56/gVcf9N1CIssT8AkOmkqsVzpkUscRaeuAvinFG5o/qfa//mFJ1Aw2Bz7UEE0tubNm+O7Cd+iV++r8MbbHyCidFnSAzNRqXJV3DR4KG66cRDmL5hfaPXFc/z/XOJ/5s2/U68/TZdz0QTPjj8Xef6d8xAtTDyBffv2oh4rWjHSIRJL91bmjfy46B0miOTaSBPhrr6E8Vq1bGHk65y2eb9Q/1N94TmuRU08AREtNF07de2KG66/CX1GzcHWNApK8wibTzm5swkCBKQeEHQ6x8Ac+k5pGOOP6KhAPPfym9i7eycBENXMhKqdZbDt0VUE8l/Jzb+waydHdIkAyaqf1ahRg+2gkLZ0cbPS4B9ehvyrCoiODGJcwilqqMTHlEfDenURV7USSocEoiyBy8AbBmH9vkVYsX4dNmxch50rp2PwkFto7CGSfn+TaGcxheQ4mspSWcK62I+62nqdZbMKk1mamsZRR2PJq3Yl5vnEE49h+J23EHBTJY/vM8nhb9OuA9asXYOvv/7apPcc+8JM/8tvzm42/pc3+lxX34/mPVK5e0+aOQP9LqSlYIq40OYHsR1Zd6amBHHBHB4hgkijCqcIgwwTKMhKh4LoL2ard5/dzJf/sj8EHFqs+gkQhVBj4q4Rt5tGTF5D7QX2gQxBOEc6vuaZjk/eBWLSRqCaCYRFptCEWGxoHm7q0ATpxw9gPYGKjIGKVnYSVVJHez/F3YBGXMJt27YgukxZHn05TmzPkcM0eMCQkJBAQCcmA+svYE+AEUC94HTKLOqwLnP2efTvks166sgv0kCOMT5Bkc+YMmhYuw4xxzooF0sHRtw5c40oD+mWFHw2NE72n9lE2FaJ0Zwr4KM87c8BhuwpHrdlRr/nFZdh0sTveUIhAKSr1oCgEDz1/Mu49dZbsXHjRqc+7AO1WzrDqpPuRZc07wj59U/3GlTP7dzdt6YT/yZ/vJ8df5MK/12qoQlgJqYmN4GdT+pBfD1lPprWSqRaEyX9JerAaRBASa5sTmLqs9NsUjB2pPjSgZFkA0kCLBtjrvlUns+nzJg4mScXr/n0X/VHohlaYJrswiQUanOxP/7MC3jm+6XYSJ3ZSDJK8snRlAP0DB+6AOVVWLGOxZ6Lxt14ffflwsvjgjV9xHRSsetYiYkZpk+diiya6MojmUH9rrFhNXjvHUZoxlI5awEzCIBnUUTIT0BJC5rtOnHiuKEZycyZGSvG9aVgszFsxbpQ3J114o+A3pfMngB+C+TpQEYh/CnCE8i8WIDJWzWTPQPajJCwACEd6X40NCumrMo29WccxTvboHFwAx5jV5KASvlrnNTh2bRYJLGvpynQPnHCWKxbswwRlEPM4oZeuXICOnbpjtdff5VA2hGMdjRQyDzg+AaT1htCqzyhBJbBgZRjZPv8eBVX2Z/HdxbkOuGcbSvOb7rzQgM8v1X+z+ZuF4hqocml9eJD6yeaTbFlQjmhQ0jA5rFI3E7JrmmPZJw0annwJdbuTsZn8yjWwHDw4Db+rUMPbjy6cSIK0ZAcmBbu/w9BSzeAi+GGXt3x+Dtv4WnKyX3Qvza1IITp+COcTqF8CbAyGNHsxEpQYtsVwQnseXM8rF7GBxfVLYc3R0+giafBqFqnvlm84s84QMSmKPnqjOVJurXBAGkoNpyWygXstPAlr6oRLQwaKA6u0kp32QG9HHED0S1TqzA2bwSQTo6vTgJiJjkw3ZbttNENuNw5nMm9BaQG2DGhnlUH9Ys83PkQUMuuoi+Fs9NSaUSX4b333sHjTzxHXeEo0799+lyFu+64BZ1pgqt1yzbYuGIdfqc9wSwKZedSrEbMrZCoMDRv0wqVYygSREdOgWWjiA3zWyYBrYC5/mkR/A3DGQFAbwfF23h/ZX+cbZ3MInJBJJuP3psfAVwAJ//GA0nYt3oZtjZrRwseFNEgN9QuZk5x4gYUl02nRZNqtCJMW3JLdoTSr+vlePn5F3HbbUMRRE2CLIohaAP9q4Nt0+nK9Yzn2Tee6QXL0kjzi60YhzEvv44b+/fBr7uD0LtaOg7RkkwQV740RET3MnxeJfAyaEgyiJqVD/HDhQ0rY9raP7B29VpUq93A5CDz9RbTUT0VbP3ts3np8cd+s3bxHExW4FZAQxi6E0xeBqAUPPNSuMR1IyBoJ4ATpfCv/WKuarNuznGwY6OrBYQShdFPQYLJYoSojjmkZY4dOxY33HAz0tNSqGc9Cf37X0/6ZiZtDZbDLbfdjb69rwYZ72hRlf2+nTcMkcxKhxYRBv5h3gCXX9EL/fteg44dOhrB5zTSSKU+aYFwQTSvLna8vIrsimTb7npV7K2/t4XYjvQmvrsC3sRX7ZTGm+CO523eytcuBqVx5+Eus6hvRZVxavv8kLx/E0Y8/jxe7FkOV9XmUSYzhTRA2mArmNma47m0mJxE5kdE9mH40QbcJfXDsfSJS9F05APYsWsdXn7pVQSERBl6oBU5UN3M+nAt4OLqrrjub0XVW3GKCjZdUe238Yv65k0ZOWxAMoXBL27fDF2u6I2hH01Aq380RxTdnCRnhXHzoB8HMybaJrybA06diDERc8nmptE43tnHZ86ag47dL+dxjIYniOFosTtYj9M37nbadnle7UKVTw0FAQoFB0459VM+9mc+8o/N28Rle0rqG+Vlg9LZuO487Hd7PfWbOwcbw3V1zRdpJ6lNBgiyf/ft22dsB8pizPoNG3DwwCEcOHCAx9zX8PwLr6IsRWIG39Qfbdu2QZVq1akql4VmzZthXoNENI2LRqvK5WgCLJvWfyiYThKBrOLk8NQyiIAw6XgGluz6A/2u+w4XtOuGl994GQl1ahpAK5qjgm2rq7bF3p7JmnXn674vNnN+UJ/6a6J4k0CTyU6OkjLVN+VnG+we4OLSqaF2ohUXx763+SmNfgol1d9+U/723uZV1FVtVBk2f880+qY4tt9yWYcn3/oIpaaMR7/nu9Ck/HGSwSkczDNYwTzktKP5y+wMnAjOx4H0KGzfkYl1B47TCxxlx1o0xVvvfIJaia0x6MYBxpm5qOa5pAcZDjLnuqWn2baqDkUFWzfL7bPxi4prvymN8vdsZ1Fp7PgX1zd/SkMsSJzfNDIESsfG4YHb78BvFL6dtN0XgxtQI4J9ZChoZA7xsMTkbBfnjneBQIb9pJSVQvJwdauqeO+TTzB0xL2oVbM6CfliIDhjadum59PNS7VNc0UcUoUw2vgTk8pXxLmCoP6yfWb7XOlUju17W6ZN43nVd/1sPp7fPZ9Vjn5O/nmGbijPfeqvfGGmwjh5nNWzZFLVi6mpyTzaZkDyjAsXL8JxcrZnLlyM5XNnF2bfhHd+cQ6wT+GRtibVHO+4czhepPvW1159k5sIaZdBEbjs0j40B/YUPmf86vxdc1kbVI2JppHfJOSRtBNMmncMjUJcUaoemtVqhIe/mIQmzRvTO+FmigbF8LjtrBlTsKptamie/vTH9t25WrN/KoAvNBeUv9d+gaVfahdAURm632mg3HqIei4u6JsqIi9N3gbRYujNrnCCF5dOHamJKSKv1QEtLq77vUtP0Ew69zflqTorX3mtWrpsOd599QUse6wzDYGeoO4vDQAYzp2oRXYxy64dNSVI9M7hRI2k8+y65fxxPIO+h2NSEdumNobdeSu+njwZQ/r1QYvWbRFPOTrkp1P+LBtlqOOooHL1s8cY89LjjwCz+t4ufo/PhY+2b+Qt70z6RnrV3vS9CmJVOa4RKB1KZgf7rUvnthgyeDAe/fBDXFSpFeKjqGOeFUjbyFzQchVAIKjNQj+Hm1tY3WJuCETJlqD/d1xQKwrjFgGrl/+BRg3qUdHfMZulcvVTEPCQ5zP7rL4sKghr0hxQyCAQiSxdntxNHuOIBAhUSM9dc8ozWL1qb/teurmn63vVUfXVVaI5BjAT880hh5luQygcTpoiTehrrmWcSDNaMatXrcT+Azswa/YcfPXFuMJqJvKua4d66HV1J6OGF0W9bvklzCF9OvO9H2jgQtogxPhatce8eQuxbMkStG3fCbSmj2q1atOxVhC9yl2GhMR6eF7GVI/uQ9vEOjhKY7nBhMEnuJ2F5qUgNiwAd195MV7/fip+ocGKIbRSnUpf1sfNvPQno4qbheVNaWhcw6C2am1J5/x0fVPYMN4Ioy3Q7TV95f7mea++VN5mI7OTwTNScc92QEr67o7jTf42jtIp2GfPMmy+9lpcPHc6dxxv87d1cKe1eeqd8hEQEiF58e+/43LC7mT6x/DN4aD5UtyDA5jDeBb8mbSS7yJHONo/A+XLaaemS0FEoBuJKGIC3N6xHRav+5V0mEnckqvjzZfvRM/LL6GoBD2CceGqXFsfTZCSNiN3XFvvoq7ueOeib9xlKG+LDem9AHNQUDAG0qDq+wSAX6/NwoOtDd+UmwO9oLEPBPicv64V4c7U415GAUJoeCqdgtENYh3VuHkLFuIqGpYVELPBzhc92za739l4uuq9frZ/Mw0myL4ntmraUxDHxlV+7rxs/vpeUlA8G5Rewf3O/c3mr6vETPz92G46YpdzdtHsDm7bjY2bd2H2r1Mx/pOPsZfvFLpWK4eHel1AWl0e5SVpUkycoWxitpI4IEKbl5pOXXUyK0Ij6Gguku5SD/IEQ7P8VCu8ouflePLxh1G/fmMEh5dCUFg4+g8YhoXz56DvgIH4qG59jCWtMGzXCjSuFIcM5hNEcgb1gBBMJykxIdzty0bjVmqUXEqHTLFx5ThSbCN/ht9/svnOS1W4ILj7oaS+UXTbN7r3tu8V14aTeL19c46vtgHnIlt3x3ibn2f5Z5OHLUt56efs8FL7yqEM2jFMOgH8tj6J/isyaMOOCv8cXPf4Kr2Oa2L3ZdPbWSrnYAppVymUt5I/CT/KgtWNyMNA2thb/RgFU3uGUH7uHrTvNQjrNm42O5XysG35d9qgfIoL5zJf1dWdn71v3rIl7uRR+PUpy7BmH/Vrg7gxUH5OFlLOOHBToeCIwQLjueD6tKqJUaNGGevNzhg54iTufO0Yut8VdW8BqDY5AQ6l8yNmL0qWHYei0nn7zjMP2z9FpdeGp+/adMOFTRML37NzBybSr8dDD/0TDRs0RJ9el+Hzd95EuwaV8EjfLnhzcA/06dQCVUr70PCB6HXEHjOTiV1RNpX55fKkIstD/twoQojWVqlaF7/MmkYnWRQlIog6eGi/qcr8eXM4/zjXOV/rNWyGxUsWUUNmDe0IRqJP96ux0ycWhzOOU/hb3G3mS+kH+ZDJJUCMjeEphmH6jOmk81L8iTmz98x/88HLPyX1jZdZFBvtLGZdsXn9n/ggwCa6nx9lGAK4A+89ehztaJ6uXz06tM4nAKTSfwGF5tT+4AQ2RzxeJVPmzwVFg+hGxowSVgSmGTx+pKCU3wkMaRqCn+/tjF2rFqBZk8accGvN5HcW4X/PkNlFbq/CoGTy/7obrjd98/lmCg+TSyLT9vJZQTCj5cGr5/ZxalfaJ2GAcizlQ58k/lyAFyY6ZsDWrlltoogTfLaLR8LP5cuVNqphKkYhgM6FVDPbHvOSf862DJu+yCsLVTkC5DqqKegIOW3aVDz36tuoW78BTwvX4+MPPsLATo3xdL/ueOr67rikXiVUDwtEGFXqwngaofVAGpX1oxkt0vLYV6CLgjTO0WTKKx4mPDpIo6jHCVDDI0thLw3Mbt++3ehcz/ptJjp1aE8/wh9h954tRhg8gP5N+l1/PVb9vpTC3+mIoCvXpOAwYo0iD3BryKVzLG4YPiF0IH84E3V52hnJ4/ZHpM0mJaVS3MuxCmQY5EU2+q9/6fTsX1/uf1WJmohalMYsPO9pzJfyUyFYTvrfqDdfx5T7u/CYkYwTOcG0diIhUGoBEDvRgrbBLGxhgQ4Y1CqnIDAXP2k58hMiwWmZlkql7xB/+sBoEJ2KhSM7oc3rs9Dpop5Yseg3HiMqGqwzgJjAf0MQYDB9Z69czNr+GzRqgIGDB+HjDz/GVYnN0LgcbdIR8/CjTKBsBWYwmj9pbidRBduPeucKXEkiN9DxIzHIHDSJdeh+n40Zg+7dLyHAIoXLyGI6nFyl9AZYqc7CjsIj6PqSKneqtoShBbxVE9Ga3DYXvc1X8WzwrIfKFGEsjwLeEhSXYy21NofH1U08BcybMwevv/UW1tDorELvto1QK7YMKpUORgjN7OcRi84kwymL9ieTqZmSQyC3++B+rP9tNaabFJxevIpQIIk/9UiVhLqoXrs20g+cQHR8GcQk1sTnH43G0DvuxOa1q3HPJ5+jQZM2+PnHCRh0y70GgDVp2gIPjLgL3Xv1plrfFuxbOAWlL2xv3LhKBjaA4l/q8vXMM6FMCBJKReD5cZOpVrcSjZs2RHYayTmaBiawzepQj2Et+PiXXP4HAL3sZuEmPgRSkmkKCsqietBmdOjQHW/2rYL65XyRShEAHzo6F57HIeY/jezJcPJZMZwRtzuhgzFyF+VEzqXBT03+ZIpoVQlPx5Rh3XDJS7/is08/xz33P2DoP//hOXOyUV7cCYPRYrdXKfqHBYViOIWWPyMAHL8iA/W7kaHADSWHCz+Am4eA4KlB/XVqf+q7RGyZNY90NBJKVbPokGAMaFMNn3/3PZ7eugW16yQaAOhjOppjosinCYqjI6cYFNJiOZ50zIylsNdSZH4oiKtasWJFQxP0FZYp2tpZBAE9lWUC72UmIoCboASkk5OPYdXqNfhm3Dd447W3TJS6FEMZdmkrVC0TwZOCgHmOEbpP56aZSzpfKI+Z+05k45HvZpn4Lbpfis7PDUSb0BBUrVoD0RXjKYJFrQ2KYQnDFVMnmKbx87Iky8ojakoSRj37DB5+eCQqVoo3UgiNmjTF52++hF0XX0WxmATqNJdBp/Yd8NrrbyI0ZSd6t29HsSPSsUnyy6ZTrAjaONyZlIlfVm/ERf06Mt98dGZtJk35lkf1OhwP8f0dbN9U8uy6ziQ9F3/+BwC97EUtjHzuzsE06340KR33PXwnLo9JQo8GTanIngxf2niT06BAg7l4mekp0bjAhTXSzywJhcQkqXvKid064gheG9AYIx56EH2u7Ysa1arwCM6EBYv6lCz+Gx4IMETIb9C4GYYPuRFvvD8Gveo1QfO4MPrQlfA4ifB5oSSmE40oBFhaJfqdCgRlaSePFqa5muknmA7FA3LQonpZfL5gG1avXE4AWIddSdKEF4vMAj7LYBIAbN68BX79dQY6XnAJy/ZFmXLlTQ/v3LnLXJVGv7MNButjYhlW8KMoifDX5GPHMZt0t88++4jA73uT9UVN66NN1XKIj5QeOel5ApacBDIum0esT4ZmA6l+GRgSg5lrJmHQPQ+h982DUKp0FKLLlzHYK5OY7uMUNV1p4C6rnsX3nNLG3mpg6Ujc9tyLyA2JwL/efoVMvTwElwrHjfc9wLqMx900ASYGSZv2XfDr809SbOtiJibWSawun75Nwoh5ZpKbNWnrdvRqUgeRBK7ZPIpf0KsTHnn2Ldx2w22IJjMkR4WefbeZPjlXfzy32nOV798mn3MxQU/mIaMGwPivv8cvU2Zi2A2d6LHsENV9SMQn58OIinFRSt3prALT+pEWmOYfRZEZCRXk4wiFRLpUdkSE5s+c72RLQHw+ZpBdkGdVdy8TkVTOxUvfFiS+X3f9rSbVmFWkndJHRaivBGulPc192QCWUwGeuwh9EUlCm5KAlg8BnXyGNC7v9P20KVOQRsfiksf0BJzufOy9bbswVQVhfJUrV8aqgiOnUfi3mKmBJoxUfPVMHiX9UXn6qTwBv2Ok7/34w/fof11v9Lj8cgP8bujWGi/3vQC9G8eAyB83Ph5zKU+aTbGTbBrakK+UYPaXgFJQcDSOZfph+kag55W9kVC3BgJopeZ4Sg5dXmaTW0x6K39ZVE/Lpp5vHoFQDjWTArnZ+tOQQz7fpaTTY110KQwcdoepui832QzaJGzUtCnSj+3Axk1rESjzbYHBiK5EGUlO+HxDIiAgpixiIE36r9mXjEUr9qBjzTLE0GUaLRuxpYU552Dzlu0FPViA9ZbUQX/Rt//vAaCdaP9uf2qRBfLotnb1agy/bzge7VkXiaVSycWlYjx70Y+GDoSMkN1lns+mPNFG/MgVppISjxPhJF5zgnHLLhuShRspsTp95q8GUMjFpl2wZ1NOcWlOAvriYpyL95r8lGNjuxo2a4gbBt+GCXPWYUWSD0JpMCCXNKwcX/rPUGcUoglKc+qiEVjTpiPSqYzNEjeiOJE/YqPyMbh1ZbxPksHhw0dI0Od4WIDFNMUF9afab3+KV6VyFdL7TiCd7i0NoKK4Tjdigxs3bTQL25dMsLMJedzodHQO5nFU8oYTaJPvCgK9Hj17YfJPM3D7JZ3wer+L0K5SGI/inFSaV1kE7AZtI4GAx8gA+mEODpDHwVweOdMwmSz113+ZQaftlRBSKQZGtVcGeVUOAayOetp8tCGYbNRvTJ/BZx8CMvFog6jRkZWSh6gy5dChc0dib8TImUZkmSG3P4jxn37Auo7DOMr21U5sQlVE2jyiSqOMvYXSB/LO5Gz6i16Kuy5s7pgGo068pB5I0EWv6nQBO38JsUptbhrLgt3j39hEmMm/HQpN4tuBL+7qBiT2vqir0uu9ZyguX71XcKcpKl/7zsb1LKek/G0ZNq3Ny/Oq7zYojaHVaVGwOfLWlknB5S8+/cxEuSKR5GShg2yrKBqG8muOpbzlkeZsgnGXyHL8JKpALFK0K+MMiMecHi2a4dPPxmDXzt0ma9VdQNndBn3Qs7d941lHd16e9zau3ivYq+6L63t9cweTJ3vLj0d9P9Y9iN7UBvS7xkT5YbPEisho8KUKG9/I8svJ4MyRk88n79TlKj+XnGSe2FAqIButazuC4wsWLTAROYKF/aS47mDrbmmU9ptETqpUqWwe9+3aZTj+eRSDaUUVsWdpDv/QIfrt1RzR+Bf8FFnv9Gzz8U4cUgAAQABJREFUNVfznpiXpgkhdji5qXR8iV9mUOf22gG46qo+mDt3Lm7r3h4v9bsALeMCEBZAmToK/PtkcGPl0TKLPmXyyIQJlAN6kkmOpQfgt01H8M7MlXh6wlxM3J+MbWQ8PPL5u4iqVJEGWmUOX5Z3VB9qYRD30r2AXz6PqdnkutNbqxkHGW8QHVqklcAwYnMEaH7Z/th/9KgBtOSqUEymOpp37IL1H7+Hm2OTUFcO2nUMZ9vEgU8i93fUuN/QqmECasWXI21S60LyqoxA+mCDOhUxd9ZcHD12iBIN3MC5RtQ3EjHSiLj7i4+nBNu/RV1tRH1TsFd3fkXd27iUxjg1oZ6VwDPYTDzfF/WsyWQXp+6Vpy3HM779VlSZRcXVO8XVEcVOWpuHZ3w923ro/nRl2DrqKsFdPwKjPA6yhkhp16xZg3feew9PXd0CFWj8MoPbrBiy2Vys+Qb4mRluBlTlnWkwC9VMG04MTh5NoVzRFIktxUc5nMxjR48UZqv222DrbvtG19P1/dn0jcpTOm/73tbvZN8TGzH7hjPH2rRsjqsJAD78djyNJDRGg4qByEjnijnZtIIsipqT/FRAc1VvKUkWxWqqlw81aRYtWkjl/N70ges866Xq7g6233S1ddS9tI3CSAfs3asXDlJ/tmZCohmL0nRepGCNJdh0tt81byS2YsvRGOq/hEOlnOJPmtnKjasx+v2P8NHo0SavIVQrqxsdjgjakWTRdFpOxoqkCfjLzE/mHKMOOUoZbYqVx1OwZOc+LFh1gGlzERAVbewTijFz9z8fR5tO3Xi0JQ1Z04YxVC/lKX1rihzwDY+8ZGmrr/Q+n4DQl8fXfIr4BAfRe93WbfiSHuNmzpuB2g2qU5e3HvLL0H3rvl2YP2sWHrjxIoggk06NjwxihoHktPvQ7cOkJRtxvFQ0HmlRHQFUB6VuCNcBsVWui1we2WNj6mDmlJ+wcfsWtCzXimuG5A6uYdVH4kqmMnxS+HfnpfKw46p7z2C/qW/8rcqQZyTPZ00IO0H0zX1v46riTodzAAok8YuKZ+PrqspoMVndS/vNM53i6af3KkcqXJ5xbFr31aZR/e29vXrGs8+qj9TchIFp1zTHWpY9bvwEE6VLVWmBcJFyN5YjcwnIKijfcxFOLnUOECeIjnblI4LQiJkvWbGGNJkWhZiHBXIsXINS2Deqx+n6R/VV+jPpG+Xpbd+rDtqoFGyfa+xsnfVNQKZv/34YRwD42x4/JMb7kXOYSYzF6VOT2Ns/rFs6AU3VsBx0qRmBt7/5kYT7e8nRrGLaqGw0L0/XL4qnepan3+OOnTri1TdGoU2HTswjjy4qK/BrBH6nBlBMTAx1jundjdxnxddaUjvVLmGQKkcAUVziAB4RU3kk/3TSWAwfOlxFoG+3VnQYVRoR1JHM5iaXSxpevj/R4AD6AskLJqAiMPWLIODLwuKDezB21U4cOXAMTSrE4KYm0fAtUwFr0oOwcN4i3DTsblxL4Xlf0uf8KGCfLVaypq4zLcwpxpAL2D9+nE96kamNm7AnJIJkHLrtnPPdNNx1TU/DjLmnZxNsXLUeR+sfpJzgcTw66g3cXS8SEcS0T5Dl60NZsAC2K8SHR/hlezBz3R481rcrGSGkN7JteSw/WCQKrp8car2HB8SgJ99v37EH7dpKS8fBlNVP6jvPoL6z89J+8xw393rTtzOdl0rvX5ROoy3QfZV+qQZWBdnJbK+K575XHPno9DZoklgdTeWj4C5Hz3qvhaM6qKHi0nkbpL+q/G1ne+btmb/0kg0AJ5puMAuWK3eXb7/1JoZcWI8Ov+nfg+ct8TpkD83BBk5iEN7Wy5t4sjgspkAId9FuLYCf51LV68qeBju16W3fy1KvHBJFnIFetfRXPfvG9of6yd7bvle/n0nfS2bOraOpfDThlbeu2vha0tpIVwKYlybOQo/arVAz3PEr4hIYs00t8cqZyX8gp94H17Wvjd/G/E65ueU06VTWWCTRmJ6JfqkKq1ixMnZuI82P9ZTtQtEme1x5CebPm2s84MloxfbtxJq+/BKdO3dGt27dTNs0TzUu0mlV2EDtiYcffw7jv/kCXZvUJ4CugPgwAcgsHnUFjEhvY1w/6kaHUDY0LyifQsrAjt1HMHr2OkKUFFxauzyqVE/g/KcqXGhprNt3HAuXLsc9jz6DAXcNQx71rtNTSfdjH5t+1t5TsJuqX3wMQCJ2K6yMMzs4nPKG9Gezc/kajHn7LYz/eDSublkb7etWRXmyhsN9kzDiiRtRPikP917QgptTKDWYeKLjEVr0Qn8CwylrD2HSim147KoWqBpOph2P5r70dudHcaZ8ac6wf/zIqMmiz5R6F9XAzFmzcV2/fkaX2s4B1dUzyFRXUfPSHc8NE6SfbfvaHae4e/naFoAVbdQEZVZcUEX1U0V1LSloUrsBjY1fXP72u82/pLxtg1UPm59Nb5890+u7zVtX1a24YPNXPN2rpTbfhfPmmWQX1ylHQedkTibu8CQCqzu0w9t4xeV9tu99iRmQ+kh/F4FoXL8RXvnkYzxxz52oWTPB0IicXZ6V4PDlGoh8Zn2jetlxPaO+UcMZSmq3+lFB+eunoPi2PF21+ZUuU54OzgdixpxZWLA5BQktaIyTG4wzUsqj+LlpMi38I4xdW1YgqkURQ2eYv2CREdtQ2+y46r2tW0n117dGFNpW2Ld3HypVrcnG+KBps6aYPX0KTeQfQnzFSpg8+UdDF/zmm3G47757cdNNN3ID5XmXQQv5++++Q7/+/c3zsO6dUb9CGMFPKjF7tlD1JUAKIHANIZDI5zH0cEYQVm44hM/nrKLZsFxc1agiDT6URTgxTB9ucGmce9/9uATbmeNrX3+FjldczRM2gU0qFzSBTjbrKMzPYH/qdt4L05P6m3rHP9AXUWTg7Fq3EbPIeX72kX+AeyueufZSGjLIMX5OjpMGGBsVhFe7tSETJZLygzTvnyE7R2R8kDmXR73un1ZtowGKDXikVxtUjgpHOq0mhPjRiASPxsTzSZOlMCuP7z5ZBIDBWYitUBkTx3yMQ08/g4rxFRwAJBEm0cydqVI4P1idwnlyPualxl8/QwO0Dyr0dKGkCWPTKo7N031vv3te3XE9vxX1bPPU1d4rj9MFxfUm2Dwp1WRoJ3Ij+MXYL1GViStHpCNNXDju0maHU1tNvqcv35uy3XFMbbl7SlskkxyCGrGOOMyCeXNQq3YCzZQT8GrDoShIAAnLWeQgq04+3H1zqPYkdTsOROFid+dt722/nWnf2D6y6W1+nlf73Z2/3jnptXmwzhTvaNuxPWd8Vby+4AA61Yqgiat8nGA/S/fCbDTkwGrxlhhES+JpKo9c4bhwPzQhuW7iz7Mw6OZb6dsjunCuOK4HnM3c1q+ofA1wpizd3XffheXEJKvXrEXn6JkGC5v5yGzs2bsXEaR7TZ82Hff941E0adkKI0eOJJb3GXW5R1KOsA3efvsNPPvM0+hG01AX1Y1FeTJ50qm6J/k97Qk5/BPgl8Ux9KOKWij+2LwfHy9aSyJbNi5JrIhatDIeyOMmbQmRERyK9asX4ad9wC13j8CVQ25DfI0Ezg0BJm4sxNrEYZZWWj7pfdmUJQ0g1ietmixyXv3owD2UZR7es49GTyfj8TuHmGbffnknNGQ5kj5I4TGfxu5ZN2KhxE5DKd/nR0CWRqdWErsJJPaZyeP5lCXbMGXFBjzZpxsqhftQv51rhbRPabGINimI5isBdZaXRZU432waCuHJt0nDMhg/cRKG0hNeMMlMeTpJEWArFDUW7nljIhXzx85Hey0qL8+kilOIAXp+LOrZZl7UN/tOmXpbaZtGV2/ydsc/0/uzyT+HZtsDiTmsWbsUP02ZipE96qJ8qNB5Ujk4gbliOWgcYa268xA0LXI06din9KSJuNBs3NaxCm6jvbbGTVqiYdNGnOzcsWkZJI8qU6GUr+OsRy7pQFoQAhfegGVv+8ZzXL1J55lG3WTfmX2Dz7K0HE+61ksv3IX7778HC3dWRJVE0pi4AElNIx2MojGMczqxE+UrulQGjzay49ezcxM8/u1CbN26BeVjyhuMwzl+qxanD4obRisonTt3xpWUret+8aXcYMjBpo28O4c/gFHvvstFH0hu7jQ0adPGAPNhtw3BA/cMw68zrjAWlI+xmFu6d0STmHCClSyK1GQgV2boJe9IO4lRBBqZOeFYfvgY3llGNbcDaWhXowIax0UiilhWHmX8ktmeDcuW4hfyv9r2vAajhwxC6/adSdAMomwfAZ9GW9OQQeI1GvT8AsCSS0CYww1SQDSdnOvp06Zg5OAbTdybL2iOxEplqcbJBJTDzBTmqH+cR9nUSMlnO8XlJmuGZAWWQ0HzrPRAfLuAXu24Kb/ctxNN4JPmx43YOG/iacU944TgMrHRuwvgp2xioHXJVBlxxxD89uuPeINGgKvWqG5IIabeSs25boM388vGPZur8j8jAOhtIe5GeJvm7xjPRwqgHMRtu7aZ6tWpWIrOjSiewMHnqiSmxQFjJxYwIs9LE8SIySWQy+dC8SOmdGOr8hg1ewfaXHoDXnn0TgSRXlY2LAotW7RG6XI0q8+JGkSswo+qYcZA5nmplZPpuRhnrVsdzQJ5LOvapbPJ+JuVR3FRQkWEIZniLZIZJBZoGE0C6cUHLR1hdz7EhIQ1dqxE5X8GGeVs2qxZoVEB89KLPzqiJyenoFmz5nQknkBryrt4DK5BQErrPYn1cfsbL9J2YxtuRo1JI8xERlo6YirE03H6IMyZ+QN69BmKuTOmo0lZmv2iYYIU0dCpMeRDwOdLTCmcqpO7k1PxzeotWL5hN9rGBaFxx0ZkSlA1kPXfTY7//MUrsYl1vbj3dfh40M2o164VTVNRU4anyzwdSXmUlfiU6HvcJ8xeLKAlYx0KrCoNIfhhw6K56N/5Mr5J4bUxNWYqIkqiNfmpoGw1MdJAcmt5kuCklvFXYdtBmtvc7AMpL+hLTPDQMR88OHkFutSJwj+aJiKCG0Q26et5HLuAXGKfLNLZdlUyabyEfnQNxU1MtPscQ6IpT0MJF7Zsgh++m4iY0Gi8Mvp1hIdFnHcESDUqKhTsG0V9Ovt3doc/+xz+HikDuGvmErPas3OvqVB1iqLIybl2O26rzm51XnrQtp/HVy1of2J03MVTKOJRJZRGEh7qhgfb5uDde+/CG7fehmsH9EP1Hj2wa98xlJaMnRaE6sjr3z2ohlrE6WQyVKOa33082i1YvRPLd2XwyEYNAq5OHb10vNPyKjlw+RFgSnhcpIEK4XmoQaDw/eQphhYnjM7J4ySWUVJ+Kk1E/KioUnQdeTPmzZ1nMCR/cv937d1pkg659Q60bNORjJLtpr8JetC0SXO6OThM95eNcek1fbFm6xJCGNk8DEQQsaUwAr5M4oPTNx7AQxOXI+V4Lm7s2pzWwRvzeJmE+T/Oxys//44vFp/A5c+8hK/n/Y7nPvwYLbp2odpaMMWv5EWQYIUYmzB/yWPrFCmsT8wNf81b9lUeNxY/7s7BjLDh9z9Y3xQ8fsMVuLROHOuRimN02ZlJQJxD+qGYbcKeFcSBNxIQNFARQr3hDOppz92chgfHz8LwVnG4ukU8wrOTcZyC1vmBYaYMVeDUda/1QfVQ1kSM7VzS+nwp+iJ3sRs2LMMrr72JD774Ab8vWkRQ6ZnWVOMv+XNel+9f0oLzWQgHMCUlHe+98znaVwJiqfSdxSOGWPvSydSAn9zxzkdFtPC1e5KbSDUTIns4jChUDDqCW9vF4rcnOmL6U62w4bkrMGDTGkz8aTJls6Ufy4XB+nm3zM9Hvc8sTwEa7idU5wqkMdOrTeKFW2jhl4zRAAJyP44DCZ6GA+s+YnmWovYKxjE2sXSKMlEY7srO1TFl8iQjuGyI6SzHyvB5pvd81sJ0CPA+uPDCizB1ykQkJ51g/vQBTQspQ4bcjqjS5UgXzMLOHdvM8TMrKxuleNyuWK0mVq6chaZ1EjA9sz4Bmw7zPKKTkbE5OZNY/HJ8MX8lulQNQJvy+7Btxu94deIifDSXZunvug8fTJ6KaTvW4Mbh9yGRdiIzycFOlfoagZzmhI6X0tPw9eWkoBpbLkkyPiTwhRI4pxyhkDavIRRqDmC5OoX6hkr0BJi5dD12H0liPNIDSdvzJVYqWnEggZNEusTlNpgg8w/iBrQ3zRf/mrMas5f+gSd6dUD92hUQRBU6GsMic0T6yNSDJ1Mkm+V7noR8KGntIO4SFyPnmkf2Tbt2IL7RhYiPq4rrhwzALz/9aAwLC9vWpvVXh/8BwNP0eD4ZCuvWL0YZCpzm5Kfw+MLdiivNoPu8avGezyDAEEaq9o4TAVh5mHRA+r/IJdE7JyMZKRRD0BElGofQfWAiHnn4DR5TaCSex2ZzND+fFTuHebNHuQiJ7BGjqN+wAa679lq8NWs99iYTFwkIIvAgECTOxO2gxP5mq80RTsCfJHwE+WWgafUYU9P5NPEeYDiOEsERiDx9MKQHLlzRH2vWrIkHHrgf69etwIkjh7H6j6VIbNDIHIcDqQlx/MQRqo6xjmxHIFX6rrr8UtKO12PmzFkon7KLsnWOR7Zle4/hmfGzsX4vuaUMB0MrYl/FHmhPr3lfTp+P6RsWYMQzz6L1xRfRkEGM0ejISBEBjX1BbNYorQnSE3AJpGUTA/TnzhhC23xHdm/DV6+9hE5U4Rvz0gvYuYF2JMlVFvJ85MBh9K9Kj4TVKmHkd/Ox5/AJgylmE5jmEfhIrTDHV7KvtCojZkpWMGZu3IMHx06jLm8pDLvyAlSKkgVqqrfxlCHJBMJLLgAaQOV8kziYquUO4m6LVJ7FI3UQLcXI/cGnMzahdavW5GYHoxp1rWctmo8k+ltWcGvVuLFJ9707/3Nx/z8A6GUvxsXx2MBBE5etMJxn6CfAkE1z79kEArnkHF72xg7MpkOlE0EVyFELQgSFsVN0hOHxatPRdDRu34AUF9VPELrgWljZv++NaInCtCQTKBnDATS6qbBkq7ilwmRJfDdcbZmOdfW/R5PcGHkwV146F2diWR2dQUX8rTwG0wcLMaPTcpML8lWdVDfJ9Clce20/fDn2K6xYuRRdu12EMtSZlRxlGWqHWJeZouHnEGMvHVuVwO8XzBjzIq5pmUBMKAc/r9+DN39ejKtbN0brmBRcN+hWjJrwE577cjyuHjYcdTu0QZlKVQ3DK5kiJSmi8XFnEBZljqSko+lwKywwgEBItLkI0tT2bFmPb14ehe41quO5h/+BixpUwbuPPYjL6BvlAzIaVpCOuGrNBgRRNLd+jC+eurwz/rlwH9YeTUUU+1VuSSXWUpp0SjkZ+GN7Em7/djHm7z+Cx69shV714xFJUkA2fTf4kTkioCtgZYMn4Ct8z87IIU06n8A7hHN47f7d8E9shIRa1ZFCs/2VaXJr/pwl3FTWmySWpmyvNp/zeTVMEE0cbwpVHBuvJKjsTRw1ypbrzktp3c/uxpf0zR3P3tv89ay0p6tXUflnyb0YQxAXg+jKOk7YATfXk/PAxDuXf7SnyueFzD7tTjYHHlw5ahm6VAKu71ANrWqURZnwYByjwOwTk7fjvQ8fQzT9MMiMu44yOnKdrO2pNTubvrH9Z3Ny52Hf2avnN5tW7z2D3uknFTJdW7ZuTUOddWk2fzW6JzSiLFwwESD1vHpE20LRwZRBrMMUwfjZhBwVgjNwS+faeOWVFzDkVpqIiiIUMJuYHcU/56U62Lmgsqy8Yo0aNXDPXXfQXt7DlOsbyHzYwyxPjp9k9UTkEJUtsZFQX0dQ/44rutNMVy7eWrwZy9buw10t6yC4bDjGLQTu7NMH5YmRHSHzPkeGQlkv0RB9heoxI9HkCD8o3sR86aZTx+gQuhCIJOqXSubJ3vXbMefHKXjtkQdMI4Ze1AqNKkWTbpyF7o0rYxN9Vb/7+H38OW286eKm3Cyp41zaH893row3f9uK/MbZqF+5Ao0n5GDJvkMYtfogcojh3tetHupULMNqhBn/wKB2Tj61WXLInAlkv2ocStpI5LHOn32SRcAX5h+GY8fT8PYvK/HgI09SGD2KVmoyEBEWgt5XX0etmiVo166dqaT63jPYueP5Xs92nOw3m764NHZsbdqz9gtsC7IF2wzt++Iq4BnfphOB2tAB2AE2D3dcdzz7/XRluL87BHAnR5venb/i6r3iudNxnE3gZ8Myp6gqv/PVn8fJnd05uddyl1CXjOd/t3w1ul55Dfpc3AVz58zFoC8/Rb34bRjRoy0WrtqERBqubN+2NXLo9CYwVEK4mqTFB3cbT9c3ykXx3fHc6Ysqxf3dnc6z7z3jKW50NA1/3jIQI+4bibUngtAukn1AgJZr2JymV4oqUiPD7qJFYgoV+xCzkZHUfB6/OlaLxAczgT37D6NWQi32jNOeIjNxfSukSxk6rINRCwscN/4TYoXivvKdjDDwiCyNDgEFMQMCSedLSjkCLelQMqKnbjqEZat344621RBBucJvJ8/FlQNvRa1WrXCC9DT/bNI6RXvjTzLBYlyITik93SxeZWtSTY8gwE06sB9rVy2j6ayJGPuBo0t8+8WtUa9CGZI+MowLzwxia6XJDGoVH4HEGzthS1I+Fqzcjk8WHUUIDXgkVIhADK0139u9Dh7/cj4NGBzH77vp0ZDH4OEtKqNq6ToIDwlHKvuS2whFwbQupBdPlwbCoCUIrSPvnyaYFoVe8pvWEo/S/uyXHPbHj/Ql0rprd7osTaTAtMNIzCYmK6OzM6ZPx+DBtxjtEKttxkxYRvHYvr4r2Hlp55HGrKRg49m17u94hy8pifPNW51hm5O3+dr43qrk2fjnK391jFRkjDMczjq5Q1QoRYKvhIwNDsg4DgQkTYqzwDg8UqRzHlQOTRbRhtvmFTlo3iaUx61K6HZZP7Rq1ZYK63MwePTnptSpUz5GvcTaSKXqWSaFdWUsU3PBAhy7uRRVxfPZ9zpGni5/96TUkVJ62G0KMIJp65LQqiL1XskEcvpaLVC/FBX4ngufS4fAhAR4CkRLvKNqtN6B2hu/Ug+1DW3iUbOGfcOtzuT0p3VsYvMPOe+hdH+p4RY2LeaDT34pxMVUIw0uyxE+p5xJKI+Pu/ccRFpGKspHVsA2GgR99qsxeOSyFthFTG0sGR7D21ZHSHR5/DR/LnTgG/XwP6mHSrElMjB0yM8gsBMHN4DzLZN05kCKnYRwumWznuknUnFg+1Ysnj0Tzz/5LLKT9pMVBtx1WVvUjo4yGGYWmRg+1AH280lDMAc+mXVndlShDES96GBU7ZCIboe2Y+Hu/Vi1KwXtEqJQJ7YsBl7SGO/8uBz39WhKwEd3nwR2WcQ0U3MyTC8EUkVPTChfHpMJ0owFGJ0u+L+IYWA/c5x8aCIuXxtDXhot/oRhxa59+GVbFl4Z3oOyhKTmUshbfoK1YVSsXAkvv/4GtmzejAYNGtDiEjcS5c1wJiqXApZnAhPkitf4BZY/TbtInGKL/qtd2U7UomOcfKv85DvWhuLyV36quHQ0z9QvsPQELbQ/Xf7qSC2q0wXlozqdSE6mTFcawuiBawoFoBX28HiQmc0drUDMQu/MMixuLSrCvxlUnyAeIZYeBZZxPg+u24GcwHTjuzWSSvlX0H5ckwZ18dqjD+GVN16jjFplJCYm4sgR9T2nF1e2Fq8jLnHqbmr7Xv1yOgDlbob0qmXDztu+l+6tJpq3QeOaTvWxKtVroX/f/hhDDZybmjZDAo9tKWSSqE3FBX2S2Seq0XNs+MSFm81jWLmoYLQvB5qYX4a9+w/Rs1oQj3LkehYwQ2yWJo1SFszLEHKlQ8MJAIWFEbOTi8y33xuFn6b+ij79r0UmsfNc1im4TBh5osBbH75H5+AxWDh3KkZ0aoTqNAv19YKl6Ne8JoJ4TB5FzK9MQgP8+OtXiKJJ/dQ0AlHWV5uExkmAS8i7hN6P0w7h7t17sGbRPMz9eRp+/m6cafYF9auhSceuqMJjdCh9x+TkphBTo0EBpvEjdkawaWQCA6SqSYZGtk86jSykIIJjEFOxFiLK5eD9ecvw/I/r0KhaZWQH5mFox1poFF8KyZRjTOY0EQZaAIMIyCj6pT7hZuJgd8xXlfUIAmacojz+ywzqcaTmRbPMUtjA8/07U5fhrnseRGz5SkimwLU/j/EkI/K0QmMYFDTP5nF42/YtSKhZgxon1LkmHA2hrURv1qythuCYfurLkjBHO7ayVaC41JsmFlNEg2zG9mqBg73a955XfdfPLhB913NRQZVRsFcbzz57ptF3fbP527oXF9/zfUn565vQb1Nv3odQMX36r7/g5kGDTDV+T6JaDwfXh0RpTQWZF5KMHivkLDbPyp6DZ3UbzXyS4+tMx5WrNqBF+3pGyyGN3ryEkFarVQ9Pj/oEn378AerSV+v4cWNptJMK+fwoupnoalLpchbZyXGwfWOvJfWNmqLvNq766Fz2vc1fV5WhWRFBzKtHr6vIdPgSK/dkGwc7Bvs2X0+tj9LZUDCjCurLseKCDqeNwIs6NMaj336PB+4fiZgG9YmtiaFAYMkEvlzwouU5vaONw8klh0dOH2JSEgvhORSffTUOj/zzETza+0LM2r0TOceSEUaaYvKRXIpsA1eE+CDe/wAuvao9aoTmYj99Y6xK80cbAso52/Zh2MuvoOeNg2lwNJLWl8ndp24Y3XogjxipP8HHvp3bcIiY3vpVW7Bm2TJMGvuZaVYc/95ArY16caURRXqiZEJBffQsSTkTiFPahy2hYQnmolbpjO+bT4YGTwGBdFQUyjP2sSwfzNm6Ez/SeEHbqtEYOiCR9cvFa5NX48oEOn6nxZccmtwKYjrNbxtOAjunT+x7z6t6L5ReETPYl6nUHY6izOGetFJ4gf5Jel47EI1atiGGRl1lYpJiEmmUZRQ3hJou1w64BpuJAXbu3NXYFpTlazsaZzovNXZmDXtWsODZjq3N1zkbFBP5331tCykun9N990xnK+/5vrhn5e9ZRkl52EUtazMLFy5Ez969ce/IR7Bn81p8Pf5b7EmvQF1gHsUICCV9qgVT8rQormbevVddRRcKoA4nYqvhw/eexa/zO2HIgN6Iq0JuIfWDk1MorBpWiqpZ99BlYh30ubovRr83Gtf0vYaTicuK9Cm3fTpbsme/6H1JfWPTeXs9m/xt+epX6T83bVjLFDdhXTK6JJCYTq5nDulbdpy8qwuXGrG9FpWFowG7dmxHvbqJZoOwi9uTmG/rLnlPCRz7U3Vt7cYtGHLjAIzs3QWJsdTZXfsLRY6upcxLKM1lfU7tiqroWqs8gojBZFLk42BGIKKDc0mHi8a3s1eiKUkWTVu0JFa3E1u2ZVJzglxekiv27tiJQ3y3bcNGfPfll4VNqsu7Wy+kkQFil6VDaCGGsy2XNM1M+okR1NTmyIpx/kkWj3OSwJp4FcWJ+Ey6m18ABZQ5d45lZuCPPSn4ctYmdE2MxpAu9VE9gnMq6zgqVCiF+y+vhdkUyE5oXx2+RD19yUAhisYdprAq3t2wr0i8MHKo4aSPHqd65kNjf0TLdp1w5YU9aKKLJmAJIOX50Je0RvW5tFWyOD8rU2xn3Ddjcf31N9CkmByLaV2dXFl2XnhXkZJj2bG1sc4rALSF/LdchToLWMj0VadOnTB8xP0cwC74PSPFNGHJ7kwk1KcusLF2qzkiDIIDZWaj00pvgaLwRsV19sKie0h55/A4WLFMPH69IZIC0LEYPXcN/vHALAyjg5omLdoaG3SZFL7N5RHn4ksuJwOhNG4deiuPkekYeMNAAouTRwLl5zkBii75P/2WWASJX5XiYzDy4Sfx/NOPYkuXeDQvT9NRROks4DJrpJiFahcNT64GC4wJ46bFMH/xH7ig24UEgGSlSqVRwycEmwtYY+IO5J/wu2hzgCy9NOa1SnmanEo+jBaJ7fHJxG9w7NAB9Krhi1aV63CsUpAmUQFiMHKOlcLjY5fqZVA5tBlmbD+I3p06uLM/5b45YU6/dnVRMz6WoilAZJiYDsTyiddJ2yONgE+iMBLCzuPRjaWYE4iAn575kcfPHMo60loMAdEhWnHevHcfPl2wCs1qxOHBKxuiamQwj6ey9pKPVDKJQghQK4UH4RfSIfdROLsCPaRn0rWDsQgtQtyp3XFKff/8QGybSbRJHTmeTK2RBaieWBc333Qz6X6BNABBaQquLZIG1aWqrlpgyAtx8fFEWf0NDS+QnGbTuDMr/M/V8fLN/1kAKEBggYEWi0ND4QQg4HjkkYfRqcsFaN6iHYU06TO1UjXTnfNX70WPWlWMYrg4f1mkXRqFfU6WHP44Bcllo6AodzktGi0osxC5K8vEezYnlx93b4nppvuEk/PHceeOK73VPCqe67glgKjgACuuJRLxS+UfQinahwvjPO/TPAGvTT+Mt0jz69VnNy67vLehlaRTGyGDnLoWLTvg3vv9acHkbtLeAjHw+gE8AhMrENdOi/NvH9SL6iuaSeLxrVuntgSAwB9bjtB/cBSPUOwrs4B4DOdi0hrV8dUJunq2kRRBAqNSoT64pU04vp05H3cPO46IKDntYXz+lyK+ke90VqXJSjnlEED6+IbixNGjmDB2LJ0VNUdgehrFUgJQgXTF/hl7ERgTiVh6i8sh/SmH6X3Vx6w7oRPrR7vIrFu92NKoTi5tZpNKhizhlCUyAo+v3LiCxeUl0LTO4YnkcswIMIgl5hJQ68QhHV0dd0Vhk0OkALWd7wMp3C0OsgTm5fB8/dFkTF9PVcIt+9AkoSoev6oLaH2LiBfnNgXofSTTSEAlcJ9Js/lhFKOqQVnJwymZiKW4US43HvmiYcTCfnCwMWcum3XDTvNjJXP9wth36Wwq+5hAONA/F5uP5eP1CQsQUz0RI+6+35xOUrmJ+8o8P+sryy9WhtCuwbDQcCxauBh79uxB2bLlSctj2/80jqY65/wPW/q/YAAhAaKWzte0r6bd/vqbBlP0grQIMh1i4iriaqpojV91COsO8l0QCbciXxN4yXiQTBkFi2vGBZPkT8VuZiTMw46hchYH07ww2CLvOYGWHRIXmVL15IzlclWLH6gVqahGqV+7O+ltmdQBJk5hrHP8Rjtx9dpfiLFffYXvxo/DU08+iqSjB2kkkxL5BMrp3M2bNGuNu+6+F3fccTu+puBugAw3sEJWoJeF/K2DgI84uDKZnlCnLurUqIaJS7bgKMUn5HheHSROpOmrU1qiPj41aBPiQLL3eKSuXg07/phHp0abuTk4mLGO09btgTulFr2FqyuWLycDZTlFR8pyQ8oxrmByuWlVpMpbDGmAmQSKPPUZnVxDJ+YYC0jL2b3ccooGG0RsTuagyoWRGUFAExMeiHIUYo6kgHQA547SSRMmg9i8jvni4ufS/JWZLtz4MsidlZ3sfHJ7OcM4Z3j65phn86i5/lgqflq3C3d9OROv/LgUNWhf8eErO2BQu9qoouMu56XylTphLkk34n7TpikBMDm1FFJuULUCVm47SOzPESMSYHasHDlT2PRhQedoY9a/fOqn+/hys5fJK45JJN+v3Z+H5ybMRDrpeg8/dD8iyHxMpzVrOZmXPrdGxwI/Zeds8moLGxNQmr62N9lSCq7n//J/GgA6g8n1wcHx48TYuH4dbiTK/sQzL9OEOQVAORF9eIwJ4ABGl6VxOYaXFxzHcdpGC/cnFZuTzxDPudOnBpVFGAGWyHVacNzu+L/gmCsMgz/xybS4JVoTSdWlFGKXE1efIH09glZmUrmTOxigAcicUGahc2Hk8rhSijbjdh8PxcPfb8aA7l1wDdXFVqxYhh1b6AvijiE4dmQ/OWryqpZLAn8emrdsh5soW3XLLbdi9tw5xBK9cyGgNv5ng/qMWAgxnjxiQaERpXDv8Duw4kAW9qbSdDv70QAtARlTUfVocYF9yDHRxpLLfmxMi8YKu/bsNQtc4y+yh/Izq/OUbE6eENIzhJEwijYpYXbKlsfNfBLzCzVATAZa4u7ARc95oCCgI8CWRSCXTaCTzc1Kz/pp83PGnDQ9YnwCvLLx6JsjRobu6U+EQCxEtvmCaLXPJwRJ9KI3Y91O3D5xBp6fuADrNyVhyAVN8EK/1uhenxgn5WSC6JNXdOJMngzYa+QSE6tmfVRL9ZoBZDwdlNPRl1vE/qRUitRwsyU2ajmpbgChOtpfHjFxQlOUkngS85q+9RDe+HEWru0/AMNuuxWyuOyUxLKUjnGKCipHDse6dm5NrRpK2TKi6YuiIp+Hd/9nj8DqS7MD8SqObhbNGb1Bk/cXdb+CDnBqUySAdtuI3YXQ1+m61SswetR7pvvnLN+OSZWBga3jaQJJ04ODRmKRiPabMkshnsDQn89ZQgMVNKIcfj1qUuexLIlfZNOfaocqQXjgx6OIC8tHJ5o7z6AQs/wCi3MpsQvNGqUpF5SGg3SU88i8EybLy3v2MNeGDRsTm9lA3cpWuGfYUMpTvYPyFShekUq3mnQv2bZDF2zZtAVDh96OaVOnIS4+zsg4qt0Kpmrm7u/1R9VzDBbIpaMPGjRtZSo4c+NR1GnHjSgrmSNDgKL+cZpSZAO07BwaqI64VE+jYF0Nxly6fAUuu/QSQyKwC92sUI+8zBgwfrnylKFhyBftjRi7dITlXlL5K1/LlFFetm8V330vqGNwJ1cZWugmjhrMewFCPfsLyHLcAwKJGRE45/JImkPd3B17T2B78nHqEe/DUpIEfLgBX9OoCmpUKYvK0TJoyjIImDPJ8ZWNSJOWYMifm7vkFp0pebICmr06d4Rys69dxh87DiWTTliO64Ft4kbjAC7N34LAOjqzmeQ87vFBQWHYl5aECTOXY/H+FAy9ezjakGzkTyD/66zfUCo61shKlnTyUBlyPhbDPhY2KmzxrwSA6oP/r8Mpk5At9excO2nnzpuPd999D1f26mV2TTl6jqC07MHdFGp96mH07Hk1Jk35GV3bNcU/f9iOsVtIV6IljGAuCklJhPJodOBYCj5eTmsh1NM1A8nJp11QU06cL5mp0jLIkygAj0Whgfn4R7dK6DdmO8b8cYLObyQAG4hwYofaiUOkyE6LHEuPhWLYtExMmfcHxnzyEerVr89FQRok5dISatbC4sW/ox7pPffdfQcOH9pHS7u0Ds38AwJC0eeafti5cycZJyMN11GYrg1CfNzBs2/c3/6yewMMtAMQ0FCjQzJsVRLq4Gqqjb00ZQ2OkFblT+CjI7Iv+1641clwaoPU78LCAxnfh+QM9feQXjXwFp3/pNDOn0SDND9Ek/OcJ3xhxo2QmOJFDegd7mrMmz+Fm1u4UQXTLmWAJ4FFIRAVHFOZ/BUe9dQe807jzs2P9eF/E8e8V9kGK8tDCEWsQmjNxzANaJhgL08aa/cexy+rtuL2L37Cs1Pm4cs5a1G5VDkecemsqX8rXNK4EmqVJmZLgCkB/myexQW8RE/2F+GEzualmqeDrzuoVqwR6yFOrC/qUjtk4/40GZZhOmJ1rJO+qaY65Zg7btwKQZQukM3FpQePYeTXc7E5MBLPPvcyOrW+gMwrYpPk5NaoVQtbN6wzTEWTiH8K+1gdxKCLNhnJA6anpZJ009TUSGb93XPRfW8SnsM/hW4xCytXROaqgP3pc0kVMhOKk8YzlJS/Z9zT5W+/26vSF5e/4rjjKa6dsLpq906jcO+H77+P/tffSOJ4NBW1eSAggyOL9tDGf/mpkuDee4cZ7YQYEuIHXNIb9300FweubI7r6pdCVDAXEw2ltq9aChPXnaDKUSoGtIhCKZ/jOJpL14GS8qfAqnZl6fWGcDf25SQSI2Tuli3MPR0PjV+Gh3j34sAONOXOYxGt8qaQyDxr/Qm89dNSVQEfvP+B4ewaOMopaWX8atSoga++nYSGtKQy6p03yAR5kMAz3EjVy8n1E8+8gMf++Q+K9fTAtX37URWJtvYICIvamT37yhTs+mO/26s+edv37jSuLAtvzXeNF9tm9GK5iEUjDaMV7u60jjJu/HjsT/FDRXJ0T/D4qUUtoKKgxaT0XE+nBNEKfYlRscWIJs22Wrmy9DS0iE7Tj9J4aQWziRjhdpODk9S0R53MsjWTQymdPOSWQeh+6ThE1ExBc2qWhPIoKq9qkgU1s111ZVz55mBC/iPDgf90TFZ+OmqZQ7tQVva94vrxvXgNudwcpUh3nG5Wk2kAYf2BdM6jbchJdjD+8JhoOnlqTL8boShbKohyjY5eujbBVGKjOt4Kqkq7Ipf1UJeoGwRoxeI2GjSe/cJPqmEej8X+tA0YQqH/49kHyA3OQDU6a0ohgFPLgigaI6aeaJ/iSAdRr3c/6/jTinWYs3YPfRr3RbfulyEigvKQFOuR3Ua5aoirWA17Fs8jA+mQYThpQ9bY5LN+8gstrCGNfSU66Jq1G+k35VtcO/B6WgPKMAwe0wjW0QTNCbXFow0FXwuiKIIzB3Qtbk6642i+FAJAPZgJqBgeQZnZn8enIh8NjYbAxUzIApTWAh13AuXpLrekSiudrZ/iWeBl8yguf/d7m787je5Xr1qFr77+Gu99MIY0Gg4URyqQ0qXzFizE70t/x6effoomjRvj2NEk1CE2MG7aBLz5ymt46YuxeOl70gX71qQSenmKO2TjisRgXPv+Kmw6FI9b2lWh5y9xEqlK5F+asl+0ZsIJkOxTDhuP5eLZ12ZiEds1nzKHpUMj8dbbL+KB98eoqaeE/tf2pWjLULpo7Fy4TLV4FAz9itcGFO5duHA+WrduS7t1k3B5r2sM503+TKpWq0l64FAMImOnfr2GqFMn0dBbOAtP6VOTYTF/3H0vwGmxJ71397FNbvtYVwV7td89r8rHptG9fipDzsGFZdet18AkmbM1BU3LhZEuywXMOS9wqSAyhjniOY/mnf6odMnFiaOcQ4AVV5ZHQ4ZVq1aidp1appwcHhs9g1M+6XAFaHLb9u1wDRlhH48bhwU1KqMprYMnViiPEH4P41yRupiYNtpYVKYPmR92HciidT59gRA6GBqg9IYzqLZ2jBZfdh1JwdZDJ3CUrgw27dplqlElpDQurRWDcqVr0idIGP0ak8PKU4WOuDw6IIPziJJYLIeYsNotQMw5a/pQpbv7wOn+U16ZQvRHY0MAasgzBOhtapbD8p37ENuYcpKUZwwmgyZN3BZipcGURpCF9OWHjuHNHxczcTj+ScMGtSh8n8PKZBCJ8BNmyLKzSU4KDAtGrVq1sZ509Zat22vxGoDsS2Aqs6gywBpN+vrWHbvx87/uw8s3XoZne12JFlu20YtfOE9MZltx5k3BHCqst+tGbbbBPS81J93fbBw7x/Ssex/qXp7MwcbyuCojiy0UlamNrm92ogs7scG+s8+eV1VW+bsr5xnH/ax4yt9dnvu7573yVhkn8+dkYSTtSOIAPv7EYxRO3YveffoijQAjiGIJxw7vxT3Dh+G6667D6NGjzWLUlJMCPOcicrlTLlxC4Pj5Zxjz6XjmRi4wQ3TlBLoVjMcPPy/nUxJu6lAZNaOoKkUz59JVPEornz9M2wi6vcGAa3vh2SefQgWKDEg+S8T2bXRYs2XzdmwjcyMurgod8VRDbFw57vyRNEfEiU+MRkfAU2c528EjodSHJk78AX36XI2hd95NjLUj86R/DL5PPnEMt996Az2W3YS33nrLOZqwTzTNzBGQbStubN39rEVttUpON67M2ozrqX2vt6cGd/7ueaNYSitVunRyWu+77x68//5HWP5IR0T7n6DYD3EfLYASFogG2pfAJ4MW+ehnicfqfAx47HfUJXNo1NtvmHkgQKJy1DZbF7VN80akCl3lcnHGjBno3r07qlSpgh07dphGlKWfqib0OyybgAEE1OobBePsiXXLJplC/SsufjJ1fNfSbFnSHictpeYYMxX1qsfxWBtKDnEwEirGUQ6QWhWcYz5ED2UMgWCO9vk0h9kWznvlpvqdnM+myDP+Y9YA6yhgpE0/mUyI7xcepKWh6hTmPsq5HoZylEDIoUXo7SQZTP5jC01l7UWtejXpl2YkZU4lskLcmnNSTD9xg53hcPoukNIHK5b9QYMHidQJDmV+BILcKPxoz9+XTLnMlKO4/Z5n8NglMTRbVhYfvDcZ1/1rHG4ccBWtXpOzzvhmjnG6q64Cru5gx0rv1O/u8XPHK+pe42rS849HtkVF51Iuwi+wO6atjK6qyJn4BdYEk+5wSQ2w+Sru2eivSurej8YNtFtypEzjAzhg27ZtQ5OmTfHsS68jjvJ+Uk4P4I43Y+ok/Otfn2IRTXa3bNmS3DrxEp0jlll4ZHRQqoGTJhMHt2+nv4j9nJ05dBITgXKx8TiedBTfTfgaIx95lbEc4Gj7a+SIEbjo4u4Usm5FOaxS5LpxYXJ0/TnD/QsWEAXLSIxxuJZH6BtC7ZcaUT4nqpyy2wWgPPXN7OKsoWiPr7z2Kp58kgLEtAVXIa6aUUEK5eJav2YZ6ZmPYdov03AhhYEF/bKIjYhjV1LfqwyVp0kj/6vSlPE2qN8F+E+Xv/JTO6KiogoBrN455Uobww9f/GsMBgy8CWMGt8DF1fOQQiMR7HSBbkVlMMvEuS38q68EbtozaAI+IMgH/5q9CJP9r8SP//rYzFOJHKl+LK0wlW6yyJSyOud+xN6ETY987DF8M3oU7rv5crpK2I9NO9Ox8sRRJO3be0raUx9IL44MRe0KUYjm5lo+MgxR4aGIpDhMBGm+oZyHoQSgEoVhJ1CeMI+eB9lmcm/9dI5V1fQTkOFPolfnJLAsAa8cdg5LZ1/n48t5q9GydiwakTyQQssxSRTuXkWs+4vFS9GgYX307XsT+5DWqWnPr3rNOgTwsg3InhMMVP9rCJgvO5SIRCA2rlnF43UINw3OQ9L6xI0PIwaRnpaMES8/jmtig9CxZlWWnYeVW6irXaUnxrz7pqmLdPK1IdoNtKg223kpnXPjy7uoSEW805wXDZ1C9aqt6mwuRUR1Xhm2NiuvAosLykM/VVoGCBTXxi8uf9uAZDbWxj1d/lqAVlHapikpfy1AKUo7k5xt5WjJX0IYmQxffPkFBg8Zik8+/4YCrpwIJL4nJx3CiGFDcPddw/HiSy8aYrvKMWXwqkmZR3Un0XfkFSyNWJ3ENkj64DfK7ZHGF0wAKVP2ew8eJU3xhJG7yiGUCqFaU00aLfAPoktANjTf0HB4PKIsocjOmukSqg7k7i9ucR4HSdhPLoGfsB0tZgVhLe7Aoih9k8uFFknrxEno0rUrJxtoUupB1jHYHP/IG8G7b75OmmcSZs6YRVeRZQicMtk3J40buPN036vt+slwgp1ohX3ijui613cBQBHn7Ti5Phfe2rFTHAFYCyxtBH3XDr9j22ZUrZ6AB7rXw7B2lDEjJ92X43VSgFz95+4X9RcxMubrTzWsXNJjA0L9sfZQIC55bTZ+p4mmZs25uRGwO5ZemLYgueqSwfzTiInolcZa+uELly0kI6wLnux7MR24Z1HjlZQxHqI05o6JLJ6WmFZgWbRHh+RHkgqPhwKi0mIRLFN7jOkyvpPOto7neYyssoyqGPMQ6FY9xFDRVa1Ti4ggmlDYbwLejHPawDyUg+aRgJb6xcinMr0P5fWC6f9j2c5jWHH0GHrVTcCaTQfx4eJlQOl4jBh0PRJI/onkus7lRjB12k9o174tGW704ax/nGxGnk91ZebErwwAPHb4ELZv2ojGdB4lQWxx9TO44bz58ShUWrcK1/VsyfmdakS91C+3fTYNG9avRY2EWiQ5HTPYv2mn2mfqf2or9U0/wRtrdMP0leIXE/Rd8MaMe2En8qU+FPdTXraw4q6Ko/QWYuv+dPnbNMXl6X6vuAp6p7ztva4l1dvGs3mJAOvHWXQiNQWj3/+QamX3cUJSBUiDSGn2tWt0fAWuoxZFAHcx5a1FqUkrDEse4Xyl1cGjgfbOXBLXUwhkswlMMoyLLRJ4OUmyiDXGxpRBrerVUa1GNdSsURUV4+Ipy0XgpzZwQxFW6iPn0aS1+BGYSiQgSHQk0kf8CQj9qc0huUER2w32qnHlz7bFXs1CI9CVpd2y5crxmPsm1q1djSW0JhJMrrK42mw1rrjqKqxauQY///yTaaPh9rkmizBsm6f7qsj22d33xfW7jeNOZ9N7Xk1F+EfzRsHmqT63/a730RSVGHjdDXjx5zVIyvQjN9ICB301oEE3ruAAEV+Oqxz/GIDCfihL4WCFw0eOmKtjDEFz/2TZ+kCQxUqps/meWJK0hBrTG9wD99+LR8dOJdOAgsXEaiRqE8Y+Ls1fWTJsyoX6oUwwT0F8jqTlE/nfMD6aibnnccOThZoMHjdFG0wT8DNQT8d51oE//ieQ1ECreM4T3ejKnwV+emX7Sd+8CcLwjJA16XCiR5q5zHJ8eAQJDpALThp3pYbM7GVbMeKLqQb43XzzULzz/LNoTLnSQM7b5FRal+EcrVW3HlasWk6/ITT8SkCvI7TWj8aWPB2uD3F3sxBZqjRFuI4hhxaEQoJDWX4mvv5wFEqtnYXeV3SmgPZx1oDpOb+DKMcqSu+yVesM917ts/NR3eE5b/Rsg+7dc872TVFXdxpnJtg35+jqrsg5yvKcZqOFJixmw4YNWPL775T7q2WI7dqlU7kzfPDJGFw/YADqU9xEQe2xnV90RTh5Gcd0tpPCeea9sB+h2nKWY+45KaQSpPhMUHR2Z/lW7bK00fbt21Ol7xF8MPpdHNq/j4Y5aQKKiz8urhI6d+2Gl198mcLTR82uqXrbiSaA4234K8dZbYug0YceV1xmqreZFljkMEnaFtKaYGcaAPfnumthEJAI6+emIhm+0v5ZGFCXNvnGfMNjbaqYvQa4nFxOTi4Cqe6g45svuaCDbxpgXi/dm2VMvfvkUj2OR8FsAtlsLmQBAgE3+zO0Ly5Qm5uZJ6yU6qVfUaEwblEfz/CdGScWpI1Afq51+jGBzJ8IbbKU59tBT3NTVm7FM99OM59uJtPstbc/QMcLLqY6W4Rhcgh4a37kcE5XoT8PHX8P7N/FPOmGlP/MnHbVTT6txRipWLoMT0lJpClk4JOPR2PWnJnod9ElCMraQ6FuAmOpDgpn5ObboBYwf/5Cs1aUlZ2PJ3vPVcA5uPV+tp9BYW7IfAbJ/rKotn4rVqwyZZYqHc3JqeMvrWfQHDgo/tKrV29D67JYSUmVK24SK83JyV4AIEuKXFIhp/lmd0BdHUAYQCu7g02q6b9M5ZWMJsqu+XGydrvwYixbuQIL5ZKQ9dHCFnb79w7Ooq1fr7ap5pIdScRoZU1PhAMHkthxPbUdWjr8Ltov77IJBCIpD1irAjGNbTsL3gsA8rvH2GhRu4MWow/t2ZWuXAsff/guPvh1MY7ksA7E2AVY5IDc5GGwEd4ysf258/kr7zW+rI7ZdDN4alGNwmQ0lUfRbJJGVlN3+PN5S/HEt7/ht525JAfdhpdfew+du11Ks13lkUodc2mtOPPYwci1WYrJU6NGbaxdu4xAirRp5mt62tWHjE3k0B+V4qpizqo/8OGnH2L+7Nl4us8lCA+i7cF8OnPK1glHAjmcocy3UtUmmD53Dvbv3282c7sxOz157nvuvADAc1/Nc5ujsD8xXT78YDSG3HYnj51UMOcgS/J+Gd3/KTSlUKYNRS8s+/XvcbV1NMd0LkaJj1TmLv3++6MxedJ3OEAifTC5wVKNqlipKtvXEu++844xbirAb9P/PVpTRC1URwLq6gk1cSllAkdP2UBfGjQgKvES0UW1yosIZlkSM/Mnl1wmKoQFZhHYNazdHDsXzcCyFWu4uEV/FR32/5H3HoBVVVkf70rvJLR0IKH33osEUHpTQQUs2LvO2HV01NHRGR1ndMbeG+rYRbEjUgQpAoJI770kAUJIT97vv09OuGQIoqPf9733NuSee8/ZZ9e11159H+i4jA0AAEAASURBVKOAarfK8Z+VrHU4niRDhg+3r5d+Z+WRCYqc4Gw8PQnf0YizWhG/yc+jazzSEcnmNDZyN4sPxSgaZcuW0hj7ZN1+u/WFz+2BDxfazrAEAkRcZn++417re/IwAhIk4cNbBHVM2C4n7gH9gNi0sepPMCb4Sk9rhEH5Pv7ynehGrKwjGPy5QI4tMUUelOY7cFVzvvrE7j3rFEuqRdRpeRcSAKQEF89QWHLtP5rfxIRYW7HwW+w09zkE6CFw0GsAYv01B/D/FwjQX9z+JOqazQB/v+wHawCScAJcdvF8DE8//3w6rmOXMbnpbpyV1yfDf82B/7XLUhvVTwGgrj5FN3qkxzJ+M2cWz5BBAWlBANxgjFc/+nga/sTfWyRnNui93wrIfp2+ir7gKEji71104SUm6d3OPNANi7Fqfh0dwWLVanLYrJL6Q0YlWW25ZFVgiiKQVdMkkJbK2LXHXZ3gzft23M9y2O5Q3BWja9W1m264wT78fqftyTnkZH048iA3lPG2WDoveW07gpD8+8e7+v1xeWBRHXVFEfLwUAoRyw9cynJAQsEg/kr0WzmZexmPy21PLnA6RClEIao49GgJgTye/3yO3fnaF/bWrO+s7+jT7J57/mo3XHebdT5pOFGr61gBSp8SWF0lj6X12q42CT58GHFHRQBzTZq0gpjY60yAdPi7WogukfaFoQSJtuVrFtvt199pMYkxdt9ZJ1uDcGIaYnhdAScSgeJIcyIFkA6CkqVFHBG4lVauWu0Qt2TUUh45xOqe/Lofv5I+/ddt1G9RWiASEyAtWYx2i5ScnIpMwzOz2IaTfN4BztjF1iuEnU/kt49I/In/Ldr2a5TpA6jKcoDKVcL2JExyXnj+WTv/gous/6CBBHVIBgBLrXGTpq5aBX7t0KGDQ/I+kP1f7CvMlGNfZQDduKnXdgVIaJcIAhT1BmaTUomlRP89ROE6qA+oP6Sj5BN9JteLCoJLFNpp4MD1aCgV9cfFB+Q5o1f1WuBXd5NHageSMNkPOeP4m2691V65/367+fyhaO4x8QX3CglKdyIk7MZSRXp4xBXzUx/++LvW4GMuZRitptlCrkSYIQCBUE0Ybm9uzkAeUWhOJMQAB0kngfwxGOPqIsvJr7AZ8xfbwt35rtp+WYPs92d0QSnXzGJrJbjN0EURIiy9KDghm8CmBn6v3m7wIUdZ1LW1a36Aq2hI/V6/RdFFohicid3ks888aR0bx9i5fftZLajnPKhnnS9iOoidPhAsmzaI/ENRB+INQXnUn4r2EoJM4++UdBpHkO3PGsTqja3hdzVIqSHX/0duV1E5DOzW7VutQ6fuji2UkFqUwaFK16PWRAz+f3uiO5Xrt9wGDjrZdWfzhg3s1J4QO5yDfM6aeA4A+vR/sBv/F/vOEmARiLIqh+rIwMd5on0yexHCcrS7IAV3CBTUgjNxEgWk1VmV9F0lKHkKnwjCxA8d1tRufPENzoDR+TJeHpelhg/5FSv5XjhSOE1GWbaOe4u3ZVsd2kETaA/2ZZXUkkNmx8MirsRjf7j2wp4julXvQTD8SVSjb9yU/DMCsQ2h/uAmiTDNpr0NV8EZG/fbcx9/Z7e8PssemDrbIpp3tUsR9Tz4t0dt8nkXW5fuvS0G1zX2fWeoLVMsabmlzT7RJNGCzHfiOeWuiIKK0A5ryIM5Ta64YJ+99uo/HfIb26MLZ45kESIfuzvYcZkl6V0fOvnhkjdezBJl1m4eb/MXLLA9uxXmLbqSOznRlv28fAzdiSd/Z/qpN040n1/OUSS/f/NXuPpUkdrjIz+ZnUi+8PlnH1tWVhY7DDMAAmT7sYULcElLiGLQMcP//0ByiIHFo5Djd999lz333MvWsVMXZCsAOsDetkMne+O1VzhEKdsSEzkXAgpKFG9N81F9Xv3xPdGh0vs1la0yqpd/VLlMkXufeYqtFW9dO3ewW958jWMKIi09FpctXMsU40+aRA9xVK6sowrxfsgbQcqQxDiiD6/4jgOICqwWxwUHrkvlDGyr3zZ3j36Io5BWv1GjDJvy6ss26exzrf35I/DgwOYUMlABcL13RHeeWApssepxfyAM2aYq+ow6poPSdeKaTKcOE/Vl98FCwoQdti07c+2rZWur6ho4aIBd2rKdJTVpaSl1ki2OOG0KxFCMLV7+YWxY+echfdpGf9zYah2osYENqaHpyqKIPdH4EAt56jTCZLxYlqxZYlOeegkj8TV2w+j+uPLJBKsA1z3GmhAywZCoFVC1gZX4fVW9sulMSGlqL742xWZ/9ol9/MV0a4bctwTricB3amiWu63y/Pk6Xj49+8XnAh+v4MAGnEhDqgb/eIVWPlPZfvqpsv3n/gAzzyxwjgzE1GX2nG+tX9ZwxtRjofRsxvTPsfG6wdJSUx0yCGSb/TqPdfXLP9azY93z2+Vfj5XHv6c8gX327x/rGphP7wWWP2DAQLvzzrts757dlpjaCGUI7HFyiitmxowZ7jQ55Xcs1bEK517gs8Cya8heddtvl3+telDtS+Dz6uVrfkSfyJZR1wEn9XFvL9+Rb41QDIcRMFTRkkNADiEsMLFXYlgd5ciTwKRwT/KlzajnyZt0zm7DBg1cLqqpSmqD3yb/6rdL4hH9yYukTbuO7p1P12y1vk3SLJ5ApWqDQFXspceSe7JZr3zvU0io6jdf9N39ubmTxpU5pJAyNulDINWDcPoF2OztwV1y677t9tX6PJzivNQWb6bTOEWvCd4Z9ZPSLK5uPB4mUi6gsCAazCGOoqxAFCD49+BafROSrmSjK8upbJD/q8YrzKsztNd2IxHK2rUrbNXGH+ypf/6T86pj7KrT+ltKAsIAKNXDKEPCZO9K20UdgzFduR6u9dazYMvNoDaxwjzr0u8kC4IyvwE560svvYgBvoLDenmrN0rv+vPiX6vnOdZvlRcqZHAiSR4AJ1q4Cj7RcpVXVIdcoAI7qO811SfAk2dKTc8D+6NyZCEu1zxZfut8Y12VklNSnTZLhwcdzNvv7qWmpTukqPJ9mzr3oIYPlS/PFPmKnmjSsZKBfT3ee+rjzzkyVMDgj40/hgL49PR0S01J5vStVbj8gQDxXojAH7PvgEG2e/duACyKscEmC4A+3riK6vHLP1679Uz1a7x/juucvHYCEa1fh8qSUL6U/mk+E4W8Exra0h25Nqx5XTx9MUZHlhfFAveQn1CJEMyRRaN+UQzlyLc9CEPlCJNwYOb0mdatS0eLTUjGQJkILEF4u4iXxbo8PiHeo4r8hnD1x0dtEowsJWK00s6gBLvx5U9scEOzpLTmlorhdj1c4ILEYlKvc8Ukn7Sdrj8yH1GDdI8rzUOBwDnQyCj3cq5G7v4DHCQUbHs5+Orr5ZtcPvdB0NSe/frY8I71CW7RzhLq1HUGxwqYEMTYlLG5ETUfB0ypWgnAwSbvhQ4T++mJAAQTXtXHXmfeWAlFQyWBgIWeFE2ayaEERKA8wG4ZUUOF7dm7zx7Hr1ppfFZ3y2qUYHEg3QPqN+McQV1laIJdbEZmR3tTOIhZEoViFDgSbKimcBwBsg8espfmr7E/3D3Z0pKT8F+/kFD5OzgHu7uzp62JKPm5cOl7k4UKGfiT4HpQw4cQYE2VV39F5fnl+gNZPY//26/bZ7184PKfH+uqd9RhbxI9ADpWPt3TYtKCUfl++3UEn1IUfdLbEZjB7JTQldQASkAIVuXr+lPtUVs0Nir/RJPGxt+1/P4f6109U7laZCeaVLbarnbrT/WoDCHRM848y9549x3r259DgSRBovO9evS0j6e+ZzfeeCMKoeTKhSgQP3ZS2Sc69mq/XBZ/TvvlslhT+eqPyiymT3GwwXdeM9nuxuf5yj7JFku4JsVSrMCGswLEEoj4/J54EbpBpCCVchZuUEiejb+wk13697/bQ89Ot3em3GdZfbKAicN2MKjQooLiQag/PfaJ9QnSGh9nl118uY0jCs8STGOW/LDcXvngG7/qX3aF6mnWuIXVjq1vZ57Vj0hErd2GEkPkoHhC8UvO6WtPNS4y5aoA+QmTekSW5lHGP/SV50oe6+tRYDXBtj/OKlPfFQVbCDAIsYH0FQr6EI5L4c49+zizeKq9O22q9WiRZgPaN7EmnH9cRsDabFhZAsGB7DwkqjHXN0VG0jyUQo3qe6h2BihDxdCUMfS/3phro089HeeElixeESpxtmrlSocAfRdM9aN6krNBTXBTPa/6JYJFcHnCLLBfiD8o/u/qVz0PzFPTIAe+5w+4EI6PpPz3/Gd+fv+3rn4e/1n1q59X94UI1C4hA1EwaQ2aOTezAhf9JcwO4RanJHLeRxo/Vb57gQ+VqxTYb3ej2oefz2/7T5Xv5/PfU3E/9Y6fJ3AchRQ14U3Rnu7ashXfS0IX4dGg5ZCKa96KH3+0veziPgL0y9DVT4F989vlPzuRq9+HmtofWL7afqx8yqNn6o8Qq/qjtKcADiJeh/4QdII1Xx19++95z4REZfWHDSHTlhmpQBVN7ObJKTZ61AhCkj1pZ086ladHOJ6a2u63uQOyVINiK8W8Jjktw4YRefuUwYMtJ3u/ZbOxivopxBPiUKX/qezr8oiuUiuujts8S6FcdUB4NBGDgtFo6DCrMMxE4mvX4nQ4UaCMhzZC5xYJBuJ/Me+4TVR2fswjndJQeMiPtXEkgXSO/Dihb1X9Zay1S+q0N5neREAJhsZgQ7s/29YsXM5YKdCH2UUDOlv7hvEWhytmCcETpIiKoL2Oxa5WuzTUTC72mHqT9U6ZwbwHmWrvzP3Rgtp1taGjxjmXvSj6O4EjYN95522CMJyJ33Hkf6wxfw5U2i+By5/e3lTyf5H8wTxeEUJ8yledUtA9N8mVE3qsRXG8cgMHRwtHZSnl5uZYk8aNsJOCBcE2TnZjimGmVKdOnap87sZv+VEJszVWUQnUrt8+gNeY+T8faFwDd80Mwjgpyck8KaUWQIYDPP6dSjqPQSlwzNyNX/HjROfvWG3w7/mbk363bI7fFGnxeg6K6sIZLiw2T/t79MCqXkd96DmPJGuSb3VZaYSlo/Qy+8ZOHTnUJo671lqfdBmBDQ7bRWefL+KlMh0Dq/LExzMJtePtvMmTbPF3S6zPSf3tIEoG+XXXTUxlnFM8lptgF44CrZxHmSiJsBdbKEQhFlXEkMNlsKxwwSA5kA8wK8WACzAA4nTueKJeaVIIiMNjLaGuZCfzKyTXZT4kf1S7VFEo7YkJU3TuYpv/3Xx79YN3LHf1KhvaqakNbFEfChV7Ip7hKm1FIK1wkHIYHSgRy1wNbl35zIP8kuWS5+wJESHNWrHZPlu1yx7+158sLCqGYEj4HMO5NWvR0v50x622desWa1Y5379CN6uK+M0RYFVNNXwRIMsQVwhKkUN24wIjOVAyMp669eoyUB4b54CY0fOBTu/p76eS8vgLT1dRDzOwT2pLaB/f0VwL4uABTwaoEtUWv+zA74F1ee05dv3+YnP5/SwB27BXtmpV4souq34BFoJ/l5TddY8PH3Hrme67zTOgPN31+6k8Yld4yVFJW7ZutV07dxItur2LGM1jFCF7LC29MYETSh1rHBZe33IqRQD+WCnf/6Xkb2D+VfPYAO32gBGjbO4PH9q49l1ZMDFwTYQ9O6rhGlvGwyEg2DcETtCPuMMpsCjmKoSn6kX+7xbvtIvPz7QVn1xtbYZdZ6l1k+y0009zcwPvBzIEUZHv6PHxYDOaWHdjR4+xU087w3r07EZoe2gb5F869FvWdZpbvSvuw5tLzFlADiUoCMqxl3Pnb4hRBSBklSBFgbS+LsoNBsVCRDK9kRG7vHa80vjN+9IQ6z3dFgz4f0e3k1eqJQ9edFONU+soA0QnpY4MCcuxy1PZ8gQpIMDH8uUrbDqBdhcuXWi9m6XakHGDLDlWlDrRtsF8Dtcx8NEQFKXOMBuRE/I/BUZQ+S5gK3CucG4O4YPcFYk7DDOXb9Zsttfn/Wh/fuARq0Mgj5KCQ5TLiXfYq6anN3BtW7t2rUOAPpy7m7/Cx/8CAvQGAHhy5HUkO8ue7dvtrbfetK/nL7B0sLwW6LLX3rPTrrnQLr7yCmuS2ZTdAsEupDVMELsehqBOtqClroK8CaxpPDRoSgIKKWfWrFlt7Tv1dIAjoJT8YNPmjV4e9+kBU3XkV33wA38LsFW+kxuqPUy8EKzQlfO+oO1CcO4+eVW2QFmCcW9hCCFyX6ulMul1+Zkqb2BSUFYv1LorzlEWGgONj0wdFHlELO+sud94cf94+UK8J7L693PF7N29jfcLcFOSfAgtXufWtm7NGhsM26Y+/NTiCWzL/9R3f6z9cVa9YYQWG96/r9047UO7OT/I0mpztoRikrnx98eRsePoRmbZSpAR7kFG+Pz8QsuIzLORzYjdmFbLzhzfwB5/b7cN6pVurdsG2/Q3LrZBZ02ydZ2WWJPWHZlTWGapWzUXEoIxZlWwUTlf/fr2574RAfkHNtfOVqxAoWLtwCc+lSc5nBCZ+iJzHVcOLVNLBRtwme6bm0v3tVKWCdJQ7krBHl88ePCDrqoAwYrS8eZO9XrtBs4q80pJUc5aEuyVADvyFZc5kVjTPIiCH9essA8//8jWrljhyr92eG/rlEIILOCnEASvlrtuqmy+C55cR/jtTkIEJoORz+q+Aq+GQqpyCKjjvmTMPXPtVnt55nL7071/ZUPLxESHQ8VA+orWJAoxkkANffqfbIuXLLZhuB6qf4Ew4PdXffsl6X8cATL1zCODQueCMWzNzj1o19x5t4099zz74yVXYtCJczQI4tBfHrQthG06uftwe/b1x2wQATwr2C09BCNtILuoQwzH77gGyB8kDZDkX2vXbUCeAnABWEqqb9u2bdaDwKdSmOj3sZJflj8BAiaffdcE6LvuBUlwTrO0Kyp5hhved32Gg/SdthNElpu9xzNGJb8s/cugNhyiJJ/K1C4ZF1OLnRgAoi8KLhmPxlwHalelynpUabDECYR/n/Hl1zZ09HC7466/WEpKkt1z1x+xA3za+vY9yV556Xm0wett9Jix1rxVG6tbNwHqG08B2uCoDCiB/4tJ4+HPpSjASJRPfdAOKuURlCCsnHBkfK9CDHwXtVEBtVdamoeipJZt3lNkfU9B41s30tpc8IlNvywGNriepWMY3eT8mTbn7o7E+0uxO6/vYg89+izxFK+01KRGyOiIaIwpiqh1UW9Vc606uFeXw9Efe/RRu/Kqq+z5l990lJOiwehAcB2dqXeqFinzGJiOhuCjfwXm+8/v5D26qKOyaKyq6tSTyt+MovteyhoEz3EcK9GseRZCfEyN64Ed+zgHeaE99+4zyDbNmqbVt7smDCZ8/35sJolNWV6Lw48EmyqUMmgGK9r7XtUgoBh41t1SEJq+qamKEBOLJ4sOkZ+2cru9N3+V3X7HPXj3NCfwghezk0cO1hVVR6KBzh072Zdffsk511c5aw71KRAWKPYXp/8VBKiFFgxwhMH6Tpv2hg0eNdZ6DTnZue3gKsjBypw4QBj5rgOz7OVlc+y1R/5pu/fusokTznaDXiphNzMniseRcSfQfX/ARFYrSd7o3xPCWbXiB7vtttucyYaE1DUlvaOkBSAAE/WmsvzvAvZiXJAUxDSHcFPl1JeHav8HFA05udlWgqzkMCy+TAfyEBi//OwLNVV1zPttiSJ9ar+BTjyAGNwq8A1rhsasaUZzFJHR3K9j76CZO3fCRLsAqq9Zi9b0LtiuIrjrH269wUaMHIWW7TT79xtT7Nabfm8333aHZe/ZKzB2/xw7Rh99RHPMRvwv3fTapGXkzYFYtvqwTEoL1u2z9nXj6Ku3IN1NPhToAFsLqA8CeRaGWO3ECHtv0V77x92dbe0XYy3r4vdt+6YS27noLNux/bB1GfOWzX3vLLt6YpbV6/KQPfHEYzZm4tl2zeUX2UA2D5nPKAWOj9s02OZGjxnjEODyZYtdoNUSuZc5GAlAfu7t/50PtVxwq8ClglMZxIcCPzp3uHB/AWH+N9oC4kd+MuNT18CRnVtY94wkYhziU0wEndqJsfbYnDXWLlM2hhpnhwErO3NkXvzeSfQgGlG0YYRYfWxzKuA6DpVH22fLttvHi1fZTbf9Ge12Byi/PLdZaH1pPEU2KIkWyczMtH898qDt2LGjKtJ84Pi7jL/w438cAbqgjwCFiGLs8dlZI+1bIrCcPO40Z2+kgI1RAnBwUDaIMKpeol165132xRuv251332NXXXWl1VckY5Cgk1doLXgw6Q0BZQf+9BHWkQHzFo8Ql6oRxebHR2vXrr1DZvmYYgDhjpLQLgTRrawuSXUvmaWS1O4CqPXr1xM5OB9B7VZbg4nNgewD9uqUNzln5EeXL/CjDT8ENrX4a9TU7MYRuCVhTa9QQIogfLSVvIBBNlzIgQQJUKyHMQ5d9M/7bSVNFA24lr/AdNakSfbGlCnuVsNGGe6dQwiU6yelADz1bO36DUQxGWWXX3mdffHpx/bX++5xeSdNGu+u/qbgftTw4Y9pDY9/49tQHlq4/JWyUWWgCb7yqkvti0efsond+zJv2FjyT0jfQ5QSTSBTq8CQFtVjagyG8PN/tMUb4qxHy2b21n1jrPfEpy1nVzCURpJ9/uIo633qc7Z96fU25ZFx9s6s/XbPBbF24dgsu+mpj+3U04c5EYygTHuhE1lwFVeQlJTEsaUv2OTzz3cHbEVGxRFVhbpBOCcyrv/NwAk+A+FeZfnz5F2Vg7BdmJvIDlKRyRX8Y8fmHUR+XmuvzZhthvdGozpm5w/sZM3q18ZThiCpmNUQ4xwby3yCpsZYawL87uH84BYEkziM+UolE0XZ3hrx2qBPYJdPHZ+pYLwh+P5W4H65E7h98rN5tjP3gF174+3WsWNbF5FIVLvWszYszZ1rLWtNxxLI80dp79697ur3Sz+OrGv36Gd/VJ0KdyJvBlZcU341qCYWUu+IEFbnJAMsQQg9cuRw27pnq731r8dt0IRJWJCiUcpjwKDuIslYUsAewqCMPPccW03w0ptvu8vGDhlgHTBclTGs2A/Vp6uGXYcru4FnZiRzcQhOyKMy+ZpmbTJaKHrJR4D79u4mFywOt8thsYPZtSpEQcBShjJB4RjG5hNufv2WLbYVlnn2nDm2B7e65595xi/eXRXVtn37ZBs9srfVwcA4BkF7CCG4dDwm4ADi96iUEvLpuwt176TFWrSo+vmn/wwlVzEPeBzw1fHUBJdsc/YQGshD2lZOuwrp82EUGvu4PleJ/Fx22HxpCRkIAD7Oho8Yae+++7b1JKR7KNGsh40aw4LgjIyXX7a2rdo5YCqHBdLxitq1/VQdyPT7p2DhWHn0TvWyAuvwy/zPdzUQLDbEAfkg8y9nfG0zZi9kPoItFg3sh18utS3k2IVLWwbmMPmHqYcwSxUV4VYIu1UvlDMoGIdcjjGoWxFhTTBTWb6m2Lq1rLBuHc3uurqbPfrOevtb45Z2SvcUu+GiHvbR5+tt5MCG9ren1lh63Vb26cybLaHtcPtqxgwbkJXlmq32eqwwMM93weGQIUPd6XFffz3DRhFtRX6yNAREofFkUh2U6qsm2BVzzA89Vv6q8Qp81c2NbghSvLIBLYqUJI+14OBG77PaBLcgPYmLFMC0EHOv/bl7iBa+wr78Zq7tWr9SGW1clwxr23WAJcaEcRocbQbgioguXgF8lGIHI7mcGNkWiWG2Zu9ha5wEonI2LXTDwQNwDUUoEY7apWgwoZQhRVAoGnGchG39oQr7y5szrC3umJMuHYGxeYhzG5RJEhIeFcQ61rB48KUoN1rVUWiFx50xyUUxz2Ls/THx4UXtV/pPuPHuB35Wz/OzKED/Zb9ivyF+Bbqve4H5queRAaqSDvhxWizYx6uuuNJ2bdhhLzz8sPWEEmnYvI2VF3g7SjiAI7lLdl65ZXbuahdnNrfPiG835Y7brVlaunPn6sl5AzIM1ZkN9TnXV/HeZC6guhVhQknAKWToe6hE4rQN5nGUVxCTpaRQSxp+hRLS6WEhkdiCwc7mEe5Hp4At+HaRbVm/zh55+gmXXx/t+Tu3X0vLSGvAATdYj9E/yesomIgdCIW5lsr3EWpR48Oy5A0PQgXkJW7MtDgAGyFxN37KQR6+S6nhtGa0PQgEh2GEe93lqyxJMsfa4cWWwo7dpUtzq0hrZxu/I6ouMh6YdMYZcpoxicUYen/2XistZhtmd6WndvLgETZrznwOf1pgAwcPot/UTJ0aG28+aZsaqiaTjre5eTmO/vRhxb/rw4j/W1c/jw8rx8pDxcRrNJs+fbpF7P2r3TqmMxTMASjvcqszKtJuX5Vsj3y01W4e14QT40ADnKkczbhHMqfz1uZZJgs2EcpdZ/A2aV5sKzZhcM35thFloXbOuC7WpP9rdtGoVOvcoZZdPKGjtRj0geUsa2GjRyRS52Y7dUK6LZp2kXU94wpbP/NDi0vmpD9cISLD8TACuSqKiuRnOpvi5htvJuBAVxdTMim1IeYhUFFsnjojRmyh4FBzq/k+KjHGru8MtihcjbVgX1SmPC78ZxiYuLnT3AaJDeV5KXERhWRDOEzeeX6Awdgf2AwQxQh+N22ybRxB+e67H1Il3i6kUztzYuFpaHOBm3gi2VSgrCiBTS1AQ6uWSQEnHtYZKxO8tAQuJa12rM1blws462xpNljJ6ahLbakAIYYh1y8B1hQjWrAey6ZVCPzM3XaIUFwLbMCAATZu4vlYH9S2OTO+sAzOowljLZai5HSny9FP12fg0J1xLINBJl6h9dUHF6MwgKL217XGxv/uOneCH1WG0CqgpiTADHzuA+qx8ivfsfIG3vPz6CpzgVCAsEW7VnbN1RfYZZdfbdf/8ymLx22rGABlLtxkxIhlZe1Hcw7hxAvPsx2Dh9gPHGqzZeNGm/n087YL1rNgxzabDIs8CllMbRQFDsiYCAGcX788DZScUFuTJ8CFylKK4DwHJdln5R/Ks2Xzl3Ks32L74y3X4YvpHlkLLleO6m9J0QAN8pNowvcEOb9PAbh2ToCRnVbkfBH9cyySGz9eFE7Th76KImCy/d9CmO5gHofwhfz1jD8hPvK7T8rhFeRZ3lwpj6atGOARQisuDLI9q7Lt0qvPtFksxEKASwuO18iHlrTSpEBmMU1b1qGPsDVxtezSSy6z2275vY0/E6VI81ba/OkHgw0g0wP3voBLYgN/LN0iJps/rnolMOl+dTip/rt6/sDffl6vfDY0qIWyMM6kgCVLD29ozdsR1utgbYvAY6N77zQ7a0g7u/uZ763z3bNt/h86ceRipL290DisPtc+Kc+w3cv22b2NiSSTXGoZCYX2yaZ83N7yhCYsIz3K7ri2m02bvd06toiBQoy123/XyZYt32lD+zS0hUs2WUV2uHVpHmqPX9UUzTCyC9IVV15tt97yB0w1ktzca3wOIddtCWs95RUCJJxzrjtsSwo32f1pIkUpyj9YyfMRdl+9D+bS6zd52DAZXf5AL7q4eQSWqUO0mDgD3RNHI0IiHIRfTHj/ksIDbPIHbcvmTQTB3WnfYb6yYdUyV37TOLOzsjpbY2wfUziVToe7Q2kgpz7kBFLiNQBY2iBF3JEk05VQDi4Rsovjnbq1o2zb/sPWpL7HmuLU4WBSsF4M/IdAeesg0hiQeDbzNW3ZFpvxwyY7c9J5NmToKOFUZ99YFw+abTu2WFO4jzIiTwvxKQnmnb2j2qLNAjisj0nc2n1YL5BFsCd41Fj58OfDi3/177sCAz58uPSfh+oULr3kF+Zf/Xf0W5UJceiq5Bfi5/Hv6eqXJdcrvyxd/aR39edkcJX5GVfbm3/AUjloujV/i2ZjvjHhdE0H0gedLsWRhEQs+fLNV20112Ejx1jnnl0sscEYdoRyDqspxvCy1Pbv22Mvo+n86spr7BWE/BEcrqodUmYuBzjWMxyDS0XGVVJ/Igi+6LdH97Zs3mJr1q6zhRyI/jps4bTpX+i2je3ayhojE0mpl8DBLlB4IIdQ7J3K8HcUcQWfwNh4ShHHPmqWSJoof/z4Qj7gzT0BOSgLi1pkvssjRKNb2k2ZeCW1Ddmzo8iKJTWtvB/Erhg4puEgtyiE8znA8/6IbOR8q23n7l1OjtyuXSfawYE+jE99NJVKWzZvRHbW3DVJATCFGE8ZPNTOHHqm3Xj3H6x1xzYoVRoxR7DXhIBXXSIGFLBSPpTyq1bblALb4W5U+5DWXX9KeudY+f37uvpuhf49vad3ykVlIU/q272bdWg32TZ2zLBGCUQ2AenrPNrM5GJ77Nae2O+FW48/T7eJ/VvbxGsG2cgrXrcnrmpjowa2tPShb9iiKzItKRbtO8bK+YfLCWwaacEcsDTulAbW4ZLZdhmH3NevU2jD+qfashW5Nm50PXvoReRioznpEJpm5KBMu+LOOjbj5YvsszlfWOawM2zjJ1OcPaVguja+w5LXnjVxgu3ak21TpkyxyZMnY/YBLLK5+TAhZCJ4qJ7UbzIy3gy4AwNtc6KwgB1gTNYDOlVN1GNRQaHlYje7Pyfbdu/ba+s2rrel62Bp8SFW6oCHXjtMyEaP7GH1kLXXh2LVgeOiELVZFxZyGiBmPRWi5GiOZIMyEFf9gUkwywyQgSZzjGrr5Nq2lugzzZMRLXCoURTsbRHUnzaTcIzIZclRwGitIFDs3z+YpTfthtvvwG+5B8QBkWFcgYS/b4Av95Il1qCxt6GoThEB6pv6L4qyDKo6NArKHfhtgrxXXjP7Udh4MCHtOhsAfZJnkA+Xx4Ixla2kZxpjrX/9cQwtlFVl8l/0r/595dGAefIOoSUPmP3ngVc1RpPsT/SRZ5rtIwOriRTLoLrkwrRl0xZ7e9ZMe/mZZ+3pq25EdkB2yL9gqCzhrOXLl9sd1//OPvvkY1vw+Qc2b8aXNuHcC612Uj0rR7haxC5Yu3Fju/Xe+2zikBH27bx5sHeD3cKpnD5ne6Xw90qHkSXVoTnadNSGWgmR9sc7HuTvj+55/2ZRdgvGngqZFF0BpcRupsnjwFbHYhcDzEGcsaDRCGUBCpo1Ttrp3WE93BLDKm23noUJ+MnncJjeVS3s4oCcE+eJKNCYMTtqkHsuqk3HYYIReZ+NwC0EoIf7WmRO7kR+LY9CAESHnY3t18ceuucOlc75Ck2sT58si4sn0i/C+FqICc4773x78XnMYfoPIlpwpDPGDcfEo3HLlvbF55/apHPPcu/efsedNm78OGsMIpQNJhs5baU/dMC10+X66Q8Bm5RFPpy5BX6M13Rffz7sVM/imZ4UWru2Le22u+62x5/60P50WxZjDyvG6s1jsUSGHrY/XNaWMQ+z+5/81O6/d4Btev10yxj4mo3of44temWk/W7sVHv49pa2+rPtdvD6IEsFiR6AnWvSKNoy6wURVqrI4hODrG1ynH38zX4IJKJnd6rN+bj7KT/WGtQrs+su64oCYJ3dffMoWzzuT/bGv9+Gir7AjY+QNKbA9CPUrvvd79wC/t3vrsEeczAbcDRjrnNM4DbYSMLF0x8jifrRmAkW3D8oIC38fZxXswVN6OG8HFuXc9AKNm1wb6O3sBQKTWvd0C7o1NgScKdrULc2wQqAORlxM2diMcXCBjkKTZHB8UQhXH0FNpI6ZzoUsY1AVUdzElDnqMTQYrYiAGUuYXGTsPT+7tAuyycEWRD9KWcjkRQpFDY8CHldbmGwfbVui308f7l169HbJkw6x5ISGzp3UwLUEDjWM3XTEQ2KS1mIxUR0bJwLTOzk8cC6NNMlbHoyGSsCV7z4/Av2/vv/du2SnF+h/bWufbykBycKl1rvwlPCP04J4iO8moBTz/WnCgLz+t9dq/jw3z/SEAG1FqpkEzQWjc761Wtt/sIF9h2Yfx/O1DoGct36NZbUoo11GTTQXt20zdIbcrC4BpVDrN2xgwhha6EJGjN0hJ3Uq69lZWWZIhnffOn5NoFDXLphnhCDNX4oCPMgiowlCxdrOLxmgTxC2dXC+V1M+8M0myRuM3wy+JQBKIeD7y+0dNzDTu0CtVcnwaJRWoQQ+jw4mN0GlkNmLeUgQe3i8skMBmBciB9YA8k6dNaEyKQwJkw1V5QUYCog2z52KKSBRfAJB5FrHiIKS0kxLAfAzWnBULAllofxZx7CadceyvfIf9UT6s4ujgSgo0Cu0fxBDDuqUXaAcSyieClYuFeBS1s4LFCnurXskXNH2RoivDz+2QKA5gM7bew4eP5Q27ZpvW2FylXas3WTJTdKg/opt3U/brbP0QJeeHIX69Oqia3ZuM7eueduu5e/Jx79l408bRxRTTgZDJYaFEX7PGrXE+wfmXdXcMCHDzdCfj6sCEaqf9e96vdVTFU+anXUMnMFFNrVF1xiKQ3vtIFD29rgLnGwfrj0sWCKSqDS0PjeenlzW71xj73+5Q67+cJm9urfs2zojd/ad68OsYsf7Gdd7xZ87GPsI6iEgKJlhzAKj7XbTm9nsxZss05tWkHpl1r7pklOkJ/ZOArlF68k4C9eGmtDeibZC2/9aMMHtbSH7r3U2o34vY0aMcyaZjZmXhWIQiwvXAnC/8suu9imfjjVtn39nDWqH2lrskNs7fZ82ya2UZ38GakNu2ZjWP94EEevRlEW27orB6zHWDzimFARHoI/4LICZCflVylEgRSBFWwQUkSEseEKkRVpLRNP0DPp4ZwQkJ4CHcgwWauDoTwqSTandSIpkAy445nPWBQl+w4UWHJ9pH2IXyKQQeeVRdtaIk8//tFs9/748RMtFRhLTm6E4gVoj6FtwFsQ78qYX9GHGuHNs5c1m4E7nYxmhLA1x6Vg4WCQM92x9z98nxieOwiXn+CUkfIdFj4uR1spDq8KvgQm/KuCG/p+rKTn+hOegrw4kvwXj9w5+puANDBV/61nKkPYVVc1xnkmMMC7du62W667hpBB4Tb0jPE24prfO7ZAUSHEhhdAju/ZucPWr1ll6xB2tkcWFVe3Dkf2hWDyUmGtenSyuGZNbOrHU230+DPspKwB9jq+rc89/oSNvOd269elF0LpGHvxHw/akFMGY4fVxTVVu1opZPLhIHY5BjicTidEytgZ+ozZFq0lq3MZlYxqnWo90xOtJI9jDpEPlREWafPe/U6Tl56aDhLNs3DYiAM5+Za9ZZn9uPSAfQsLLGmNFPRp/CF5tPb9OlrL9FRYrEO2nCMpcw+V2B6c5Tdsy+GpoEt886+QQqOtRVq81a8VRZSOcmubiftgRLFwnfXIqG8tJg+zqXM+s4f+vNRC6tex1Yu18EmET7rt9pstsw3HfmJKsnHdWoNbssatB1r5ge3WLDHZrr9gqG3AjvHyq662x9/+0J69/25r3aoNiwZqlo3JCfArF8rx4EYwUh1OAn9X/66yHOwEwJrmyEEecydKJrlBsr388vM2bNIFtu+72ywunI2qDPEG81vInEbGldqfb+xhp5+7yM4bmWpnEPn5hTd2INJYbRPHtrQp01bb519vsugYkIQkUmVQZVBDzRuF2GufHbRg/MIxHbQ+7WtZbNQhNsZ4tPlEJi6FtYwosdR6CqmGjDWnyJo1jbEJI6Pt61lz8VhqRNvZVBnLcvpRRlsVtebMCRNt2Quz7eQep9gBxDUFzJULZ68VUrWkfKyjG3x3/+k3Kx2ijXwiIugba0lISojNucqxGZUjihGcg3vhpvhC3UpuHPUdaqlEgjqNLZu+eCA5FXiJOnhFQWKVvDe9J1Wf1FHK+ARDiARDZBTEmKU2rIs46qC1TEq0nOgg2wlif2fFGlu5ZoN1I+L0qWecRbDYTHcu9bZtG7DYaMgaPwSCphOu0/Jz9sKlHdi+xYIzaRt9hOn1kBJ4IiYyyGYSFHXdvFdtIOLGWbM1xhkOeYayGstR2qgs4RuX1Hg3OJU/K8fB+/Wfn4K9oxDgf2b5hXdUsTrJVTx6EdTN9VdfblHN2tqDf/gDWh88IZgsyQKwErGcrTttUt/RduNFp1t3fCm/nzfH3vhoqtVr1sLad+9p9VNTGJYIC4US+uMDf7LBI0cTyy7E0jIz7Q8P/MXO2bjFZs2caatWr7B/v/mGnTxokNWGinOIWEPKpEURfyyaSfx+yx7YGVgAWEYtKx3IIroinlPvkcQiZD+M/ALKKxwWaNkmW1MrwxKxRZw5400bibvTlq8+tXV1+1jGkGut/YUNbAheFuVE2S1CzrV35x7bhonMS//6OwO31Bs8IU5civqn17OW7dOoDxYLb45aUL7gEbd7ii+WHaAmxBeMO+DVBsI9ms2OB2URQWw5qItcqMsDGFdrY4BJsDwE0iWHi+yxFZsqJyzEOjXNREtuNqBzX9u486DtKthpZ08cYPXYzUvhZ1dgzpN7qMjCCQgwKqOXNaifAGWBETfeDhVluCMRvTe9Ti17/KLRNmfON9ajVy+b9smXNvjkkxwb9R9kQmXN/+0lECEGliXYdolFIuHCsFHDrHnn8fbBVyts8qnNrIQxCIdCL4QNLj4YzQYUbnfd0cqmzztkk0Yn2APXdbMul35s+bMb2S2XtrI1B9Aew/IKsyrSigzW42uHWzHf88uEJEqhetmwWPSNEvB8gJIX1VLM2KTV44DvWBQBuYetae1ImzTmNLvy6WeR+42w+IhEEJ+0wl5+tblF8+Z29SKzrp1gN0sO4i/LvFG1i/KsBji0I+SkXlb11H0PwVQMIHEcRjly3mI8MGAOtbRc2/W2S7qh/+6Bf7Pyqj4GlOu9E1iPe7XaSwE/ySqRS5jGHlitQC7cOj7aVm49aOty823+1v320Xxkj6Srrr3Z2nfuDIsaCiwhd2Yz1QmMSUnpUFyMs8ML4s+0QSCDja3FxlKKQu6gRUZjrE4fJfuLYn3MnPmlPf3Ck/b42aMst2i7vTl1qo0ZMwJRTi0CM9ALh+yO7kdANwM6UPPX3wYBqpMksXKyn1u2bJm99s57Nn/Xc5w+H4GPYSEjTtUi1/OIAXbXHXbf/dfaGRMmWDksYcd+/S0vZ7f98MMKm/32FNu8Yw/O+rmWj1br4fv+jiyGPYyJkNwrnJ2rceNM9xfYTbcr0AyZD0RyMtZBEMbrX35i593/L5dNFIxI4DIAX3ZSHdISbdW2ldazMX60uE4t3Zljn++Osr9fMtESate3BSkN7Y6H77Pb/vagTRh/Dk7bsEfAg7T06i0bs9Yl5ZXbmddcYSsXzrG3//q4zft+gWMd6tRrbikY4TLFICD8Ul12yWSgREHQglvnKO5mVQ13zQQVs0tq9+fwmV07FxgeXZbWrKuli4qVQBlQqkgS211mvQpq055IDsMpRUu3z37YEGXTv59nuYzR8BYsTOSehqwlJj7CutWpB7mArEoLm+EsKj/kTksLY+EigQHg1Tk9P2Qn9e1lIak7bcSwk23Rd99ZFwBcCNtpWbxm/tefPtXnI0D/6hes8SkV1aKQ6vzVI3jp3+451353Fp4tA6+0uogCZHMWEkxIKVZBCSKL4X3jbfHqaCvJLrA27UH0J6XY92tyrXf7RnbJCDSnZTFsjuJWpIEss8y0NMZgn+3JLrWMZIIA4F9cgWtSCM9LUXrJlzUIyi+aEFoxEXuJ+pJEo8I4zDveIgrx/oG6i8e6SotQ/VE7lNoTeCOjaRPbti/XGihKMuuDagFhmbMwziR9uj5Xrh13k49S5sOtczZA0UchIOEgzZEDEWAY5ChQqQQX/7Vf9wr8hRexyUdJARhCnMQo2wE1tyv7kN2xZLWr69Txk+2k/hAr9RhDZPrIeYDIEKsHAly96kdCl7WkrSBA7nqjwlSCH3T8RAQb+8GD2U4OKNFnJJYVi76dbk8/8bDdOKafxYYXECasns1+a5qtW72GYLA90bazSTMinomQX+LP7/ZvgwBph/YpTZjS9p277Ka//AOWtrbtzC9BTqMJRONHUMU1CxZZEWHaR+OdUEKMtBIMR4NAklExCYQWGmDdkPlJy5V74AAHLKcyEJJBeQCmxoshKGEgHbCB1JzaXEOsXVOCXmRzUrXf9pd77ZWnnrE/jR1os4sjKW8/uQBwAEonwKXWr2dvfLHZBnVqbSm4/nwxb6lNOONyXPLqIrwtsnad2ru+1AXxJTVMQqss5MAuBuTKTkqRPRRei2otqUETS8vI5KjBIfbtrBk2e+q7uJ697d5vzWf7Xp3RyomlKEZ2AgAD3EIookSFCV27+OXsuXQGREQ9+/LjWZZ69lnIfGrbA08+YYNgQ1r17o6LUQFBMEGoAFdoGOdisGCF8DPjkpA7Am4VdS2XABDr9hywf6wsROhcZKObN7IODYiijJFrMQs7uIjRAxEKmHTUYigbS2kwkAhWL+WsR7H+/Zqk2OaNEXbjLXfbe28+j90lh8mz6njl5yWtVF6qkAKoctXKlENyHC0I31D9iBz5SPElyFCnTHkI7WVdGz9+pA3s3802RGdZDvLbuslRVgy7GF2EIXQklEUJsj0oi+5NS1BYILcCDi4b29ymfr3FeqAeHdQD5RmiidCEIpQDyOygfGM5WCgIhVe+wsczkZpXaTb1Pzb8ADJhCTjwhYU9k1ImFOoZ8pE4dXRkzXe2ZsVWS+6f7jZdLXRH8QCTicDWJRdfYe/dfD1BU4fYYajIONaAoj07WqhS6Kah9IfTDQ0fEiMxWg6hqiHanBRkgIYKY7rn7qXqgrsjw/bT31SZq9j7ovoqa6UKyQZJGEiHUd/+wyE2bfMu+3DBElfukFGTrH+fLpbSqAnrAa08hyM5+0HeUij/SDxO4uNr2T68OJJS0glWocgxlK728l8cmBwa9mXvtNT0DDaXcFu0YKb985//sGuQSbesHw31t9/qRCbacIZ7PVYaXRGHuRiCtFIwU32zdA07wY/fDAGqfr9hAqO3P5pm52GjJ0tzNkrU1sSl27rdTu/d03QmRXR4lBUgZNfh5CKJdf6q1NQyCUkgRl+devUca1HK4qyQpTqDx7Jx8+Zs7VQheSUMKWZn1c4sRcHiH5dZlzanWydbZ0+cPxIWJZTzFBJt7UbJKnoCWAATk5GUlKwS7PaPvrHeGWm2PNdsQiasFVtSEZRQDALby6+81q4/71zrhQY1BNcysAf1AywgD5RhnhyG+ksPybQgyGJqJ9mw8RNsALZP51xzo303Z7bNnTbN3pg5w9VFCZYUSczb1AgMuptaAoa6LpgBZZbBQmzZvB6q94DNI07BiLPOtpv/8pBFYWbRd9goe/WZZ+zRj96jX2Zt+3TAXxNrf1j4UlyWlKRVFishjV5KrVAcbGpbH+zb9pfutq/WrLdvCTM+vnsLS41CcoZcTwPp9YXNO5TFC5XlwhmBgBViPTIoz0b0HWw3vzLVvl+83E4amOUW4M/DgNTglFCIJBhPyaPY0Wz/ob321VfTbdVybDmxK6qABc3HRKhOVG0be9rJ1qRxmtXBELaQuWgejTA8fbrddN1LViv1ZLPd2VBTrWi8EDeaeoygaT7oR/MB+go6jGJKXhCF1qt9HXtu6iY7lFNsbTNYhAxSCSyXbCXV/2ACItSSWyJiCmhuNgFaCJKS+Xoo4wGpzrxKwYDIJCIZVozNt7DU6mBIfOYEQuJXnjEsTaZ8bL0k5BVsp3E+820gwNEHkCFG46ECzKG2EfQwJlSupItwkJ80Jy6fd1tISeMtdtS94vLzEfiO/+4xrorKrDkuo38y13LIF9jG4h9EhK0fhUo2DvqH6+A2FC9CO9oabtsZs7Wy1FiyzpU8bNhoEFEvy2zaSvop1gl9AdaCnXZWY08B/Kn99ROTCECy0lIwt3LibzfeEAAiZBir2LhY27ebDZG1/dXXn9tLzz5u5w3uae2I1CPf+RAOVCpjo247sIs9+dwUgnwMJpgsIi7mTnPsCAbV9wvSb4YANdAeNQbF066DbZjzpb3z3LPWb8RQJi/U1q5eaxOHD8IA9ybr3bsXOwehzJkU9UPeIqIepb5Xki8hjCPviapkEr3bPFEt+qy6wVsALXKbEARnS5eA/Dg9bFLXZOvRaaSFIbs4BLWTXwEi2J+DjAJSHWmPKLeIysCg3Vs0t7kLEdiQZN0vk5YgooyEYeuUltrI3d+9a5s1SEvV0aZezcAJMOmSULLTXosaBQ8XIG8Mi4i1Zt26W+uu3W3shPNsz7attmH1Spv1+Wd26GC+zUNrVrBhhVdAwGfdFhnW+pQB9uiEc617/4EWHh/v2NQepwyzDn372I9zL7OP33jNXpnykntrTEuo0/opICsAEeraKXegLr1jPxHOR5dDCdWzcW2KYfEL7d9zFtvlAzsgY+QFBlKbjTatYLG/JM9Ih+2LQAJFmErEQ1kpfb9lnZ1kWW4elP+YcieXs/qHFgRzDEs5b/5qW7/hRxRFhfbW1G9s7vTX7YHbu9qYfk0sNhLPgPy9tnbXIXsYA+39JUPt7EsuQoaaYR/M225/v7qTPXLTIZu17Es7++2uloEmsqysCLhRxGBYVtok0YcQD9IFt7CLYHFrxVTYwB4NbcX6g9arS5SjDEVVMXuu37o2hDKuKGdXYkyEmISg3OCwyclTIgSErUXZGFu42FjVK/ObIMOu2Fv05HZwy8IXcPiUbNMWTQg+cbu98sC9WC+MseA8dlg0s05LSx1k9arRtVpSC7zktaUK/o888DMc91ouhQp1hYRJoeOZmpTj5inkElKpJAkFBiKCsLtjLR4kz/o9ebZiy0b7etkGV/aIUWdY7779OFcmlc0F+1xYXeFQ5APAD6MlLEhSf0RbKPhISkoqZwevgaAhViOUvjqqlSzCSKfzxXAguwISP/av21GarLKJ/dpbzzSstjlAXbxdBWNeDBWfWCfeNk3/AbOZQuZSxXg4wjeg5s7PTr8JAvQXhBCT5IBNmjWzeXPmWK++fc2uPdLGJ9HgngdFJVZBgOTZeol18BZgVU4HSw5Eqm7V9KWMiQxj11q/bo11AvmN64p/Z5cm2HPl2WEhRhbGwtVbLaphFGyr8uLwzYLx7IrMLrj4MsvK2mAPPHgfvodTCcl9PtRKDBFciiwxJc1Vu2H9Bsvs0h3jYpAFgkBwnYcMeEpR7GSV2BAIkN1hKYLrcNT/2NxbOOxQI2wXG3fpZH1PPdMt1lzi8+3atQX5JxEzoC4EpKHI+JIbZloCChgZcANHUMiw3YwNWwXnmMRZtyGDreOA/jbp2utsydxv7R2Mvz+YPcu1sSmfDfnDBdjCMRQr2izDD6gU/ppg2J1POSu377eDyFzroT0PgsoTwqN6QM5L3ogLCcAqsnhy8cxRWrV4oR2G5InG5st12Mt+Qp/yMiguCbapn75ojzz4tP37pQl23+WxtnbwaTZjbo4tWLzZRg9OsVH9mlqbTqE2OqsDSpx8mzP3b3bHY4dt2pxlNuakc61/pwQb1Rf6VxwB7SoupbUhUDYgnkC84ECJ+dDB5yEhBda1WYztztXOpc1WsKbcvAtbWw7sdG+XCiuMJ0PZPuYUllOIzJWoRUvZLLoy7AwTojAADkcYzwbjOA+yOblz5ShUlup+OUqUTXby5Ml2PwhwI4EzmoZVWD4bjWZb5fu1VL7+m1xCBZeIAkq0KzDJZYrOggA4UpyCOBjWYRGwmgsntmr3QZsy64fKdkTYpPMuhJDpbCnJaZifBGNXCqvLXLIrwPqyfiQuEfAHJP2UyVQMcBJNyP/DKDpq1yPgKcjM2RRTjlz+DuXstYWLFzjkdyX+810TozyffggliZgcqmStRsLRZRfv4QiHVZY+KBMDBsQRlQi3qtrAga+6WfOXX4wAa9r5dV9/wghuUqHeNL09+/RBW7nfcgi9lMO1PrZ+KbCRWmwlyLC0a+o9LXD3PveFSPXdR6jqhv9M35UCn+l3GO8fJLL0tZdcZRDcNqhjaztckGtwLkSfDbGNOQdwqUGTsO87wlVlczJcLeaQwAsYZCoVwtZ27dmXCCO/J77bP6whso0Bg0QW6RsHAABAAElEQVS10hsmTElKDClAFO1XigvtgHqk/iqp+5KDOJ9GAENvCWgcrYEwWXHYhHTDMFQVTNZrmMH5rRnee5XlaF4hZtwxiGKz5J8bShkChjBECEVY8BcfVsloKbu2twwi2QybNMG27thsW9eutwLMdzZsxP9z60YW/H5rPqaFDezc1pJTm1nuhg324MN32egeHKEIVVlOOHxRTNqIKgDYIz2hDfQqRBFWIuravB/28oyn2TSMhULYnhNKR+YQ9gvkEYH6++br/2B9OmbZyiXvW+ZJBTb59OZ27vBC+35tjn02o8R+x/kQA/pG2+kDk3C5IoDBGX1sxPA8giBk2HtzD1u7jBQ09SiBoFZ0JCarFwE95iButDUulYlxpmfMF/OEPLNxAxQo2EqWYubkw46P48pZmK0bhlu08sJeaXFr2r3xECzylflEuGGt20QDa/ksZO6xyYeGpbtDtqpy+8DADaeoECGAjeUDDz7E0avX2zOiAvF+8kZbG4+3XvT+L0lqWkCV/1GE2l6Ibas8PcL5I0YEmzO/ESflQ71lIztds3O/zUfJsG7Hfvd+n6xB1gFzstacLVyLg9TFIZQUEJYfOFRrXYup1BEwYoGqNUBjJ+5DIoP0tEb244/fW/+swYg5ZCNJiC2Jo6Au//XYo0SlWW9/OK23Na8Tg3cP8whGLivH8gOqWxYQIVhQS1uvtB15ooK4llJOEMSOwwl+3f7V5fzpD+wnJU/46beUR8jpp/KqMcrndkO+OzJYGAJgkjN+LAuuYWW7nL+jEB6//fL16MiCqcwYcPHrVx7/z3+s3yKxv/riM5s24wu787RTLLI02w6xLKSNRqpju3Gj6TNkjO1YMd8djpSajnwI2VKdOkwwBRVDauXjl9iley8bOWadvfTC05z72gHE0RBE6snX6iEv1PovByAc8Atx8Ft/uiVK1gXDVIGMr6MaADK1gEeGzB2Ah5KoJB7klF9cwLiRH9zo5IfCQ0J4sioDN2n46IV3mI/Yex1VKBtLyWnK90tqg7ICci+zRVtrhX0fXA0vUR6sheRRQrqF+Xm2FNOCp55+xrokFNpwfDDL0eZp8bmVzmpWeRJeS6AtIKtAwBMZFm9Lt+23RftybEKPTNuP2ED9Vzoe/Giu9FdlqEr+ELwFRDFHR6TbiFPHWIsO7e2xZ9+wsGfetKt/18e6dU0kSkup5RyMJ9R9kc1cEGT9MB8JLjoACx5h409pDOLk2NWKQovHVEca4dLgbOTBiVDPGNMGHXTw5BrHh5sCwS6DXVIWYbExxVYb+79iIke7Tdfl1iYrJQwG9+F5jDt9R9vpFnhlaequfqu/hWxIyUniwYiW4h4wR6EolZDhKumeg3/3Sx96T+xzkJ157iR7/sVX7McdO61tIiIN7DA1Hm781VhNtC78uWnRF1eJ7npJKF0P9VyP9Vyt1XfPzERl+A8lFGLTZF7DySd5eTnUaAFs8Ca0uNv287d1j81dt1UlucCkZ08ebS3bdEZ5kwqVHwGBwpqg/QoNJ1Mjse2yFpBrpFKpKEgBb7V2qoEOOUIEJRK3ceniuQRUzUP5WcvB18Jv59gjD//NalPGbeOzrAkihSLWXinITs50ocHIFdiUpH1XAFXhv5bkLWK9alHIGUGn4jm5Mp3X+j8ePKqtSj5cCk+5c4H14k+leBCXT535SOhY7/hISSevVS/3ULV6VI7yKAqz78enMmsqX3n1TL69si3y21PVDnUDlnPB957BbxoGwoeDEYKjETzM6NVijopyt9mQcy+z3U0b2aoffrSWbTthToIGMKaOxaG02J+9y2mlyhD8Dhk6xj764ENb8t08G5rSyMkiVVcS1uuS77ljAqnPA2ChIIEaO5PaAaJjT2RWqBQE6T9zed1jEL8IFR7zn4VMboBIgmEHpCAfb3y4ocl1WdH0+sQNJKJgThSPZKfKK/ZeTuvCfRIFlGObmIuVffauvfbNl5/b4/fexROz60Z0tlZpaLUJQuks7vGGUPVqpw7PlmV/BciioiKPaCfhuF0V2iOfzLO7cQ0Mrdhv34N8DiuCNLXL5s21k0Ug+V5g0nzJyF3+vf7ceT1RLiGFCGvfpo099fc/2ZdfDbCWl11jz12UZhOGd0CxUGwjesdYDq6HpbDM0WFxyN4OOveoxHghaRGhfCBPDCmLY3xwrWQM6ERgE7gvRCGhAmMUhoUBPFVpBewbwK+kefGQDkgPJEjJtEuIX+9p1L3k0AtzIs+emCgokmAM5UGSISBhsY3r93xvtbftZfPEnIhFLBMNadS9RHmaGxqdEBtv9/3lPjt71HD766XjLKZkD66VmPAwn6Ei+1n67hAkmqWo55pjwYO66icXal494547NEnsPD/URs8tUzkFD/SRuSyTrI5+Zx8kUALAsXbXAfti4QpQjJfC6ta3s8+ZbBmZTTktsZFFYipFzRAGxU5Gy0C4Nmh0BN8OR7NB8pN8XtL8Kh2ZZ54pD6mI8YzF9lXOBIeR9UciYnj/7bfsgw/etq6sw/HdmrAphUNgaF5oN3OgKNIhIF21Q1RUOTLoMBR9MtiPZ+6iEQtFYBcsOb1aobO2pUfwcILGseakNipWgSwOflY8wJqLPLEnWijHSv7gHetZ4L2a3lcelaEO5WbvsJVzPSVGUViBhRdjmkIMuGAoD9H9kXGEMGLMUqHoXn7sny6Kr9yC5FAfj1uR/C1bMsBiy+PrptjlV/zennj8HzZsxFjbTdzCNj06co5EfSkvMa72NKhaSpJFeAtHcEE/+e+BhBaSANibFKdoACC0JP2kYREwS3uoEZKNomQk8oesvOPK02KVK5J8ij0BPYDBd52PYgUHcOcj1tv2bPDaQVu8dJHNmjXbFnz0rqumN593nt7HkmrXwZwBGy3kPJ7NvSgE/tE+IVDJbINgOwqhUpPKEmz97hz780fz7ayTOlkG7MkPGzZaRJLeEHAf6YOHzV1VP/GhHnrJsdxaONR98qAsy54z06678yH79Lr77W9332wNEwoIL4YLYixmQqLYQD5AvIgCjYQbk8qPyhID2uNXEnCtgjMhhuMm+nekmVU59b6oH2mwVb02rRDY7zwatGpXS9v93ic2YGQ3vEYQKziPjKpXmW8oFbTJQZiB9OjW1cZOPNu+Wfhv69M5i36xoFn4cjeDTgSYpIWGMOB1sZDqYyDss3U4WbEgLSIUxQ8IPpRwVkXAY4loJ+TNhciTD1Pajv15tmVXjv2wYqPtOdIc/ORHW9PmTS0ZOz0djRrltN+UTH1yLfSTYztdG/w7R66BQ+TGhkETta9NUUhLYhu1tZwQXcVQdWmNGttXWC7s2befMz6+tdMGdLDBDRKd+55sB/0xd8EQhPS0qcpWGEQYyneJsb6h+vNZ5y5pDvjz36u66X05oU8XDqsKME7olZ+XyS87cAKrl3C8Z9Xz1vRbZUjgGgWl2qxTN8Oz33bklFqL+DA7wP0IWNuw0nBLjm9g6zbvJG5eZ6zbCKK5eyc2e42dQbTOBNmPfNK1h9ktYAds26GDq3InxtgHcP1ZsXEXRs7fY9NY32oRzicOG6dwdqFQKEGR48x4JRXnUYCOsqIEoRl912Q5wTpX/msduaTvSlqaIk7CxTJzU2jGbXKabH4oUKuomUIWHcSZVSBHeuelF2zDgnm2au0qW4bm208jWpjdOrqXJeH6FqNFgg1CcWGuY3kdFeHwAO0SsgWxSP4HOkWWJkRTaEt35dsDX8y3M3q2s6yMWlaI0DuKuHCSAVWZbviVBXbGv3cCV421KB2ZR9RJrGuP/PU6e+Qf8dao9y323UdXWOdm8VaYhzkUbHx5JUWhUdO4CQ3/TybVpzGqgPx30UrYsELZCfM4JyMxcZi9/uo/7L7COywI39zyINFXR9qnORflRC9AMEWw+Afs08UltiZnjcXlbUZ7H+eokjjMnOIw8RC7LFgJgTp0G2pAR4VYFLOyAISZh2jgAG6WxcVEA8L4exuc17INgaiOFznEaAhmODouoVXLNhZfq77FyK8Wn3YpghQUoAAlmDemgj9vg9Da9UxujvQjoBlHffXXsBcwBfih3BKoOJWVEFqEuOmAzcd//9MvPqVTZneN7MmGWgshAliNtanNXjDpFJJchQQV9k2ugKKAJUflhAGX2rbDa4vk5oM2+nW7mz/zoxKV/sy3asjukJ0GjT/x4uq8EIMAVmFt1Ln/prE1VFt1WzKKaFiMjIYN3b05a/dZSpd0KD7pX1ljCLbTEhPth9n/tifROimtWLbcGjVp7gIbCBB8Vl9yQRg3ZCCxNubU0fbm++/hY1thmXt22eTBA9279VJbWLeBPS0MO6ZGaQ0I79PI2RNG1op1rkAK9CBXQB2ME80OK+0088q4AFCMhdaHLppsmQzJIFqspRA5N4ELEBPIQXZ4h3EnPIgx+Lq1a6zfSQMts1NH2AQQ+MaN9pcbfu/ao49bTh9qdYNhC1Gw1MZdUGwMHrKURTAIdlKFPtIy1M7ssSgAO/+1OJ1vJhFLDhFZee76nfbmwnV2flYn69+4Pu5hRRaH+1Ehi1uBWUUxHpVUgFtCR909oR9iUYNpcwl2fMFQ67+/9mLOOUm1LiPPtZmvnWMndU7BfOkAVICjiShToK+NQEv2fy452EX2pe1MMj2t5Armp6QkxM47GbhaEGx7t2bjKpfotYwhEWWvH6HINrSZhaC0gfxGHnuAaDE3EZexNRHRd9pOzrzZvGOXFeMMsHrpWiI1ryH3z0uZHdpaA/yRz+yaZLEguzrIqmujvNDJavEgVRf9xa0EZN14EZXh9lnBhiZzIXEmzrtHSIh15CPBn9MCwbW4IG1UgoZojH0VfHfDquX25ltv2fKVq214t3bWv2kdfNYxWGfcQoJRwAFLFcytlJGCS70suZ/klmUoOkph4XVW8Dccxn72pHOso4gS3nFryQ2w10qP2FLNJ57+CwQoREdFqo+rc3tjV9K2xScasWhcifItB42NntVGCCreX65nTlB74m088ZzaKUBykSx6pQ2wA9NmQcHgeVGKxvIgSC26PN+GtG9o2/bkWMdB3eyF1160zJZtiQ9Q295HJnE6TtzaYWsTKNTJbZjEdIw9P3jvr/bHM7Ms+YLhmIOU4ItbaAdBSltnvkRYb7PndxNE9TgtVR2KFqwdtYr1I7+3WYgCQzDNeK0DWH4qPUSGN+Z8ax369LBFCxci/jfcg3rY1Nnzna1V8wbxyM3y2TFl0IqQnXEQkpPdlmzknD0YwBYsExJYuLJS7uMVUo5J0AZi2D326bd2kD7dNKq3NatDrDfYtiKLdi5JQVAapVBj1cR9Dgx+qt3Vn6vvQipS9ZTDyumkQBg5kH6+jT11nH34QT3rP2Y458FcbCejCCkk+rNnwCtjXdrMPNEtwA8g/HlwX70pJ/Rb8K5FJ+rMIV/Gs5jxjY9Mti7tcmz3eqL95OxkbJodVZ76KEpLh4DTRSi8GCwgkiw7P9e6YBYVDXJq16wNVBgkDiIN2cQpfqVkYs6kincCk0P/IAeJUcpE2bPuEqITrBbHA7htSciDOnVfaNoRHyCSYrTgjqXmmc7ddSPHuGkv0/6lUF6eEkHLWG0+umJ3jwHQP33ngzzibTQeCszLfb5EcgSBRAPbtqy2b6Z/btM+/cQ1//ohaHkx+pepUikKp3IprEB8ggC/DCf/hgCQSY7cRIPgRqKR/e0iIvEbS7fY7IcvQvaHnSJrPAgCIwD/uTp+7scvRoDeQHmkqr5L/lbOIG/essM+JvzPlk3rXRDOw7gVLft6trXAT/CCcyZbp/Yd3GD9JvCqnQQ5Xx5UYBdGYnz7FLvl1e9sS/FKm9CtsaXHQZKzA+czUU0JbtociiwOhcBzf74BP12zoe2b2+Lpn9h+TFVaN29ppRz4vPvgYZv5zTd2M/ZJTXg/r+CgJaPtS67H0MGydc0Y5Nih08EIB5H7FDG5xXiwCHgOgOz3wB7tBNkGFW2zsl0HAUZBG/iIXYJhEwx5rDMLJD6CyC4t0vBvNlj1eDw4FCKIeCX4RtZj0UTzUj7C7Jdf/9TO6tvTnvn4E/vmq89tQr8ka9ciwRKCO9s/P5ljPVs3tQmdMcsIZzyK0KBSniLihGPQXQrgF/On805CMO7WYo4kwkkO8/TF6o32+eKN1r99MxvVsbnVCy6iPgxdyR8jlzsWfIlkcAR/kMb06DmkM9Xu1ASM/sLyr45iANCDceGIop1y3ysmRl/WoN72wftT7ZSxo23x1CuIEEQQDfgghXYKpm0yqJcNmmo+eqnWVPN/ex+ES6c9vCDUq7BOKPBi9tvWvBjbA4nXoGEzb2ykJUNJoiHx2gY3xLdyKOkItNl9+2VxzOOVdlKfQSh2EGmQT2IIZ6DMeS31k2MdFeYQTbVma9y9sXPfHNISUSEKTtyEMJpiBRShvfWRlV+OEIyXvFbpO3uhSw6ZucdHkJ//nn+VTaOoQyEeyb2FhvVfFGQM/ZJiKp9NYNbXX0P1veHKndCrvXXNxHMGSzNRdM51FdY2RBR9VTOEPN224thd9SEPbqB+FMd0Iv+9/fUv7Xe//z22xH28vrMWHNrUhFSmqq75N07gGoQ2lbGsakWNr/isofL6g+ECCoolYADnzfnGpjz9hH2AHdGVRGSOT0nBVCCVCLn1MFQ9ZA/dcqtNe+/ftmHTRquLH6lc2lSOypMWWNpCvx1++TU1RlrgQ9j6aSL8d5RX1FUMmqFFazdZXw5NevmC3rajIMKenrnaNuzYYRN7tbFO6SCKSLRcjH1xca4FoU0SkmBvxA0PeQ7U3e4921loBZaPrZ1C8afUzwABRXvyMxfZ11vqMqSWLZUCOwgQZGumiRNWk1JFh2DTQBd9xbNX0s8jY+31k0l3t/gAiUoGVwa7JKDUr3AQI1aSUNEYU2tHxNVoHyYzny9Zb1+s3GS9mtezyRgEV0Ddyg1pe06B3ff+XAurk2R/zGqGeQ+noWEwqnBK+ezMUYx7lFhh2YBRRy79XIRt5msz11pL3hnbuwleDlhPM8ZqGESkE2OUYSsYTVTh77fstnW12tjTDz9EfDaiorDo/H54o3Jk1tRXzavmV9/9eZWcSN/1pwXrogBD0+VDdu7O3cdMgJTxPqmFP1YUUVeen/KqXXHJpbZh7vWWWZtwUhiER/BuMatQXkOiHo4sgyP1/098E/KKw8Nm5qLDlnX2FNu4ZQPhyRJQSCDUDCCT1X8pCLRmFFV79erV1rVrV7vnvr9ZZtMWyBHzPSJCuyLJhxN/zAL7Uv2ZP7bHyhv43i/97pevq/oggkcUrTfmeI5AEMi+MCcH+fiPK+2pJx51VWW1a2IDmjewRrFoeFFmlaDECoEkZdkeNwkRygMd4TrKrxJ7ed4q22jxtmjG5xw/kO4UhH5ffVZdWmCZFFXHCceqSP2o0gIfK8OJ3NOOIxK7iIXyMIcZfciBQZfddINdiEFuWHgkA6UdiM4KP+TH2g78JB/91z8tASGvH1DzROr5OXkkRxNpnZlcCxlSK9tbEgm1FmLXDWxrS7fG2bMzV9hrFDi2O+HPMxpYaiTCZnbJMoIIaFGKcqsFEoxuSNAFWKwIRctFECtgPlycB+mNVFCyH+px/WdhS2ZRgVmF0JU8PpS0oEXpaTOWfEWqeoVbCkwOqLjhkICu/ImWUQAH7aihvCM3oXLqFgVWxkBKihSMDVo8NlxCVJB49sWSjTa0XbElxsuKP8iaxkfaP84eZB9/v8FueXeO3T++vzXA//ow/aujsulfUUm47USeuByN8bsL11omkVVuGN7NMnE8j2BnEOXrsCNtDxeSYddWwAhZm+3fu8XiGxAdhgn2ehvYqxP/rv4raax+XPm9PfXoi1DAW2xgVqLt25Fvs5fm2vfrQuziCydBjWdaz66d7cEn59nfbuyGLAtkSSSXcNg/MDV//1voj/ajMcf30dZt2mfNeg8kfiVsqZCDNr+Adum3EIcWqK5pRJ45Y/x4m4ctXHqjDHffH1Dl9d6niGOk4z07Rvb/+pbq03xJESEEzg/gSEF6MZXi3w4Ihh8WLeIclBepq8zaN0ix4Zwp3BjFTnDZAcvBKDWGACThcGdFOipOSq/jJKCc4HcYXGP+9O6GfPt23Q5bMO+tSuSndXT88TlO0f/x6BezwKJO9P8hwkPtgQJ68tXXkCPpSELI78Py/sBHEkf74pz9dtPFl9ilk8fbOWdNZPJlDCAMfyT5i+HInV/2TeU4Fg02ce3alXawTQpO9AAjg96ncbq1Sqpvyzid6qW56+z9BVuJl1fPerVMsbaww/FQhWGEcC4vwXgAZCVEV4TywB9sp94HcEXqq+0VUL4KJS4E6OytRE6hspegW2yMotA4PoFFyoiwUjTUgawFI+gjAbdYeMy4SLuo++54SuoqQ95RIRkdlUrpIdlXOLZ60mqPat8AW79iu/OdmfaXCb0JCRUN0lfw52I7DduqJom17dZP5tsdw7tbOlFkthBVZ1POHnvhhwMEYdxnfVon2Y3Du1pDPCMiQHAlzE0hfzLACJFBIvVgKeUWo3M5AikX4lWShvY7HER6ohpC9Sww+bu0P+9FIPpm7VvgI11gPy75ynq2zbS+FzWiLcEcT3mv3T9xma3AfP7bRVvspJ6pHOyTAeUpLwqE42hJBW1HAMrNDncEYyI1jo2mK4fereUjeZRX73vJRzRV0+M/qLwK/UaCwPcdCreL/jjNnuVs4vox9WDZML+iKD13ramsTOXpTwq2eOzQzp400UaPPdWyBpzMuc3phH5H/ilDT8Qoar+Sxshvh7tReS9wDKs/9/P94uvRw1BVjNhfHaouGVwh62TDqh9sydKl2PP92+U5g+AaCnG/ak8ua8pc8Fdt2lGIThS9WRHZZQspeWNgFd53b74cYQHsFSMK+XBFrn25YAkRtT+0bj17w5SwEbPxqr/OTExU1n+ZquwAf2oQNRE+wPo72ayZX9ui9dvsz088juM0jZJHAkZ2YVBdimG3fe0K+/NNNxLo9CI7ZcgQMDr2XCAReuCa7depq1+2HgR+r94/P69/9Z/7ZQWjOY0FEVx22WWWvXIu4e0TnY/tYTBIDKe+DRDCa1Ab+7a9NnN9rr083VM6nNIqkdPAEi0Z1byOCQxjpwsOxvgWY+AKwgNXgNBKw2FTiqLsmw1Y8TeFlYalDFK/mQcJmp04X/6LsM6yjocQAJGAQMTWouFVt93GUdVopl5955/GRTIRsYQSK4vKlCGtovBKoF0GJRgGNerkJDwrBilGgqBO79IUZBRt93z4oz14ahei6oAOMPcpRmbXIbOOXR/V2e55cz7iyhDi3FVYI1z/Lmxbm4OtG1kdjhHVuBVBGUrjLEZeJ+TpPAsa5J5JaK48NMuhE7zNLAEWTzJfyXPUfm/svTn1u6Zr9XkM/O3Pl+51whhdfwcxZF20aLm9/ilBZ//1rJ19eiIucsPsgnFDbfbc9fbCm1E24YrZ1vPrdM68UCgvKRY8JOdHTSnhPOBQ7OHCMOQuDakMXaURJq9WnsJqSGQh9zkJ6h0lqz7IPYdx9ZamRCvMIXPmzF5YtFJgCAZ4xVu8zG8wkWtCYPEXLM1x2K5f1klQf3pXwyLPHw+unda/8rv6qzoLYdf6nZRlo0eN4vyab4h6cybUMMjcjbs2WihF6j0iszsysho7fyz9cTzytOZvgJNLQKamzZWh7VxleMoHKVX4AwmHEdfS2aEKjtkcw4B1Hah0iEC6K5Hvz5s/z+Z+/aUrb3SXxtauUQYbKeODxcKqHYdtS16+1YtiPEuJuYiGv1QG68yLXEKdW6cbJLWENcLYFwNPRYQii2JjLYZz+2zlDvt0wTKbShDUUSNHOo5KMKck5O+PQfX+++OifNWf6V5gUl5Xor4EvhiYyf+uwvwC/bzbNm/GDi4OjQ07Idr9UEJZlUL95KHW/2LqNLv9yovtk48/xv8vy1lpe+8LghyEuDr9uv2y/fqOdfXrVV7P3ujIQPjPSlihociNevTsY28/+aR1a4XPJRojgbbY0nzMQ+OI99exSaJ1yki3PQdb2opd2TZ9FW1e+YOrtn2TNOvREO1crSJLiw0mLBOmHwjpQ8rj7d+E/W6ODWC9EOLJYfVfDNWLwZNzywF0YMGxNxTVBMYQZaIQP5pkt2Nz1XcloQu5mflAqcNq3OpyN7z5EFAGsRvq2ADJFHFYYQ7gfMuweg+LYoEgK0NxML5nptVBwvzWsk02oWuGkxeWF4NMqaMe7WvdINYGtGiKPWCc1cbdCB0adWPsDECzyt1C1ZiqpVq4rHrXQk2TWuyb6WhBH9iFJT4acjmxl7BIhPT9sVduP6k8f25171jzG3hP81kLy4HePdpbZ5QwWzaeZs+887GNa32LffziaTZkYHsb0KeNtXl2pv39tZX20PXN8aUFMbO4tIjVVA1dGItHZ+SWReEZgjV3JJ4a5aWiYmHBONHPBRNlM0E3BYUSjD94HtGIFRZLOEwIFc055FudhBi0qlGIG8I5H4SeE7ygvJjoJGx4JbB0oRyXgO7VDpWF2x//8bDdj49vk+bNLDs3G+6Q8zYkKnHj6bG+3oxrJDTvwAZUbz1CvN12+x+sZ4+elpV1MlFROG0PpKc5FuLVRuTGXsDyM1PgnFSNsytHhbvR0sw6uNTgldIesKA7YCqKNVwMIiqD2hNHUpSXZ7t2bLRVPy62KS+9TEsqLJnPi0/pbq3+H/bOA8DK4mrDZ3sv7C596b0JSAcVEEREFHvsvddfTURjYoslxtij0SRq7NFEjV00ggoKFkCkS5He2y5sb//zzndn93LZhVXBGJOBu1+bPmfOnDlzCgeBmSz4Wu2LMLibgD2/9g3R8d3MNr9NvBXgV0cLT6yE2BmgKvoOdEpf82NLLPnIcmA7mnmbw+HP+oIqe2nhapv65Wx7HeQ3lgUiILhITl18UJv0Uzt9+zxh5t+H94FPp6t/r3gg1QCrhkeo7V4A6qi3UAaaJ/0YuFPOOJOBb2udUGkqwbTRfKw//3b8z61dS8P50TTr1aM32+ICGKWytsGqwjZKWyc3yqG8tCI63kJtBYe9841VPbQFUwN8Y320ChBVAsb5MkDMb/JyLIhIE9enFZiXSzsEKiABt0S56Mc14VDmwDapllfa3pZgKWQBZvP/PEnIkJNZ9GAHdc20bEzsL1u73VZzinfywFQ8YW236DSAFPIvCeQlc0s6EwOLMUFAJQx2gPSEzHgrEios0AOsjEKAAcBHQ1poUmgyi2JwwtI8VzDoqmkciEoaLNI/lmR9EgNQxVa/mG1BAeIUyWwZJs1aYh2YuAPb53CAA5uGFfeThYvsiG6tbX9M8m/Bp0UVW5cqufEE8Og+13++WrubZ0IM6kPOQXBy09AdbJRjc1EIMxygfF7+qm/1GVsPX6VQEBLH6NC5s91xQ2s747gxdvgx421wj5vssRsvs1+e09sGn/qhzV3Y0nq1T8BKjyZG0A4xG9htMQIAObzShEqQlIA8GTeI8HqXrSmzxSsq8ZW72iZ/td7+NRXWROESYpPJLoHe6NnJjqXBffs3xvZdI+vUJBMLx4mWGldg5UgHVDWIsVfeXYSDL7OXXj+efqDPmcwSaRLPogbm/NjX9HAsMqXF8Fp7IhVx2WWX29/+9qxdfPEVwCYsG9KqFforJEXrdqndnl5oXoQHP0+Un5tvzGcwnkN8glOH/LjGheapzN5vWbvEVqzEYZbcQMxd5LIb17u9dW7T0FqzWMWzuBYjTbCdvNTvUWJHsCvo2qKhTZy13vbLRX0W9Q0n4yccSVOkw+5gnwTSN2d7xfjhsAlkuJIF66YXJjgW2YR33rVRhx7i2AWixh0VLUorFKrbQ6a6168unODThF+FQwSXzi9w+Ie67vNgmgtIFTSwUpRv3LSxvfTCC3bsz35WnezQg4fx7nkbMWokVoNzXAN08KHGU0sXL5KsV0WkO+wbVZ1Z2I3K1ECp4jp10ilObcGtoABQmxYt3ecKJoJs+unE1g00gBWDM232PHjH0ikm6JGBiIWiagivsHk6DrBzEX5GbGYzkvVrNlTapMXrbNWsWdXF/Qkk1CwzzZplYcg0sQF8Gyw7AzhJAEQMFF+sPG5RX8Gx87Oq/ooESACROeDeu76hjvynn9wGRSQjbaXNMNn1fSsHLMVQH4UgiI35m2wLclErN5XZZ2zjYehZp5Yt7DT8qfxz+iLspqXjq6LKNuLvdhlUxJjsLMwIoU0h3h5Ar62jEK7g3hF71S2r48YBGHH5jCk2AI01HBGimBiY3MpkN0Gn9f7EXmOnMa4raIwzcGivw7UKhIXxQIQfjR7wBR/HAfl91rHvnfbN9KvtT7cOtFcnLLOeHTu5/lF+SktvwyNlvBHdSYQKs7RKW78xzr6Yt9GeeneuvfjiWuKssxH0w4GjOtjow0HmOd1w9h1ljWAjxCDTV8ox5VaokPwduNfNr7BF27fZky8stuvu44W1sisvamtHD+9m/btk2zc4Vz/9on/YW29NsAbYtNMpZGYD+aKRBgTR9xDUL6xEdvGFF2PAorONOHgE5vN7O80SubAMtqo1C4zPbnd9qDjqCx8n6JdQ/9D1AYtFZxjMA1XSvRMlhvjK9jxbjSrogqUL7O3Jk23bogDpHdShjY0eO8xapFVZjlZhEF0B4j9VLCrxIDy5jJVF9ErU3Rxli7hYEXCbX4SxWCyfO/NbARkCMoQAYoGIhu0SjXSIfP5KrnTuhkS779V3cHY21B687wEMkHTHE5xMkgUEgeZCeNB71VnwJ7wkaYNkWDr1DduhasVThABxkBMCoNqTqzCPAD3ASxNAp11xaHqsW7sWlRxmBqRz02ZInwPE8RgYddZK6GABvk/nS/ADpKsaoPx179/7eJFXxZUTZDVWVffx3T2jKRPb0ioozq+0404+0waVzcPxN6a4ETLV1FN80Vla/WQBl8K5BwCgCDU4WpWEEGPYzotyZM9jpZjo/2D2UpuDXbrBbZvYOoyYTl69xrau1+kvpw4KnHx3b5xhjdgu53Dsn4bGRDqCz+AKtlS49xNX3KEQF9vdyvafhDllgUbUQCV8liLeyfagfB5sRStgM86OFm8vsnyoz6AsgDa2gQ3tlmnN0hKsWUYmvLwoy8YSTCJbvAWYMnppwTa7cVhb+2L5altKnY8c2NKdyMPOcXydcski0LbAeMHOgBWq3U4X9VYZ/ZqK74b3Z0y1ZTkH2lt/xxk4aoABUtNs3zUf9bUmuQAtcvx9ASHwc+Oi+1QU8WPhQUUxQcqRo4OpwAENEwvK6Pb7H7Znbvs/mzL5F/b2B+vswK5Yv8EhrhYFZ3KKKnAOhYHjRFuLGuSEaRvsnIfR+Fn8jZ09JNdGiveZlYh5dahCTULqLK6gk0N0/DcZMVC/UJ6rIJQ9VqITYDVs2BJrXyJq8fKHs+2NBfF2+AkH2psvvm/3Pfg7O//sCxGbgkJPZ8sN3EvkSKAVBN2I5ql+4V6rbyS6Ucy2UbA8deo0G3HICLvhlls57cQ0PBZT6DRHpaiW6l/1jyhx+S8RInPUYXW3B/mrnx2BwaOoc9VDZbkf+QQqfGgWsYPZvj2f7T+6witX2HzE1z75cDLl6FQdGdp+nThEy8YxFGwANIqcnjLN8LuUGMZHYo46JpMuM2iI+oCy2eomJFQAe2juQLH1bZuJ7KwE/SU1oLklKOYwhZtkpBk2wkKY+OVCe/vLr+3XN99o/3fhJahENsRVLHDDPNC4Kgj/abzUF75Neq8+CeAm1e1K1E4f9L62oDhCgDqMqvYLHJ6wtkR65+MEGasi8n9aiamoXGsCVaQJJkXoIioeByBIsFNbvJ222aqwBiSsEOXn89bruiqub4oX/t2n9elFVZXJnyi+MZIRb9m8udSawxOC7Q2DV+Q49dHIMQjakot0l7yWOGA69ZTlDDHJdRKsOMoXhR7LYlJ2bZFjB3VoZEW08ZAezeEjYSYIg6IFCECv2bwFnhKTlvK/wOTG0gXYzisM7KqREQGKxIXwQRH7QRXQFiwAPBeFP91ym1hzhKFzOUlvg6pd0x4d2I5zwkt08V1l5l0AASFOfVGbAwmUlOZZc1w4NoFinbq8wL5YVWCHdcoC7JBFgvcJqSqWH+kAKqXjx45klxCYiOfkTmOlicbYOvEYGGUJjbrYxxPesOXfLLIevfowFkqu8auBD/cm+LDTuIa/13148GOqrY4mbiVCsMwZJoooYaoOdT3+iotsw/JNdvc9t9oZ570Aa+VrAG4JpSdjHxHxIEYKzp+9+sFaO/5mDGKsxQ3Cz3pYH07IG+HxLg6qX4c9TD+oaWQRRAlTCYk8FTOOMkSgEREdqQYhvcaWFEoHMZfszCg7NCfFhp7Sz36VX2WvzVzmWCzrV663rVAqDaD83BDCLxRacIKmvg/UyepndRLBt9XBLO0tRvOj/8D+mIOfaOPH/wq3q9usx/69OahJtNaoV+Y0aexUK+XiIYbFNI5T2FhkPl12lKF83OEYPVFCXp5akvc/mZzTnCxmQV27epVt3LzR+aD+asFCWzE/0BvPoE4DOufaOcM6W25OujVIRPuCMuSASWll1VmIR8DC7KM83jNPBB9RzC19E4Jiz4M2EZQlu4127KTeX1lgXTUhoAbLhYzpVu1ykiAoSmHjzFifZw+8PsH1yTtoi4wYeYgT/5IT+Fjy9zNC3eh3jUEP1vShSxz6o37w+ECvXP+GR4i41/f6MQBDCf3A6dHBPVdlIkwaHrQKaVor+HjBE381WNUPwU14xcPLiIhW70fGzfEnBg8/yGY++Arb9dZURDM/VDJIUNPADWeoguLfUTU3wAEA60H8SuaY+BYMoEMynPzpKjm9Bpidb5QOpUjeXZu3BChYB5mtAoQSHJ/Az2YFhzkMdSyrGkImomBZM9wqrpNjiRQ4/iiApm2JfnKInsBPgtXOC5iqTUVk9UNIm9MPJ1fozEHpExWXIYboYlZr9t0HdWplt736kfVt1wbeJSp91DeO0+QS+CyxonT9oGg5ri3QjmrkQB0k3ognFytkazd1BSI0hA0ITyuozxju0KRWRYPggc+Pp7/677Vdw+NokrkDGcYlyJ+6Q07fdgtC9tkfWf+D5mEuC3U/EJT0hzPQ3F6Ba8xfPb7Ann7ifbtnbGcbhb50GlswSGzsHhZbAfKU1BZ+LBZmYHnEMG5uK0g/JNAvcZxUVsKPkniNDqXEa01mUdSiUcq2GmUZqFH0yVMr7cLhzWzEfk3slw/cbXc8+5598M/ncLvQlTWUQxfgJlbHxeocF7j6W5536huNHQ2UPO0QbFA+8cifrHvv7la+cp5lZ5tNXE1fh3Kx5Czr0LqVtWjVmnHNAnY4buPnkBKHCGXA0AockG3I34KsZ7ktXbzCiTv55KnctNEDdRncr7WdOHYQ1B3mvdBTTwLetC3VjkfjUIGPj2JgVRSYW5RcugBe4OTz5FAi8YXwlSXtgJ8nbZRCPPMlZsTbhgVr0KpqYm2TMxANKrY4Dic1DVfiV/vGV95z6W695U47/cyT0aPPZU7o5B1ihV2lkK5DHq7f6oBTl0MAe+GwE3q9x4vSfCsEWFeOfkCrv1NpVdnVvfrlD3MDenDWZkXeL1q02L5hnh7EWs4BvlBcaNg0eDvXbtcnKB8oQ2ELra5KWwqi1ymi+IfKqYpJUcIhiJBnjN4BQHJQLl8c8fyclgiraTn28MQPDDqFPAH6gL/HS/dek10nh+SpSSMkighGNJhHDrqlJB7NN1Fuqjd0anVnqu8D4ANg43Dog1OkVvKMDo+0C/rN8SBMuf+OZnvtTExT290G8tMBjKrlqkbbY6hLDEzv6Us2Oc0MpV+/Xgb2a0KgRhVMhpq33+dOpdcEMdjLYcynZzW0N96+2q4/bJyN/vwqi0bkKgn7fNPnbbG+J0y2NhVLbeJ4TidRZpFNxx2I1kBvcDKM3CJoPIpJGg3yS4dyiWHrKT7VRk6Dt0HNr8PHbTHiQ2WMe6rENjgBboAxjEaYVcvOxHQ89gTFMy6CIkwpwmwa5sGe/tVwe+L9STasX3eb/uXntn/PvoyBtoNBH6oFga2+nduj98HYBf0sCmcHuuUtEdO69Z477cOfj7dTxx3FwWI5mj9oJUGxFuDGdLtYJRvnWOHyfNsGcl4wf7MtJq8G/Jrw67J/W2uJT5UYpAPG9GErndAZNgw8W3ZjmRwOpqTQG3ili2GrJAmFOLas5RzsVZCv7Clq3moUHXIgneq4K3KJREjiccNSArTYpCDOEmOp9O8Q9Es/XgicNGeHgn730jVb7ZmZW0Qf23Xjr7WTTv6ZybKLjNHqHMARTiIUgLlvE3bBP98i8V5BgLuU9+3qv0vy7/tCCOv555+3R/7wkA3BQkYx9tHi2cJqhRE5EZxC77kUL/QcBaAlwncs2Yr1ZPLWTwYNBOKiDp1jbY73BT7R8PSkDwx4QQ0EW2jWRuKFCHoAhYhAmICMK/caQNVNAtVuK6Nn3kO0sUIBUqyqwAhoj1NLl4RhC+tj8V7Ez6uA6pAOaDqWNo7v0QbdyyQOgMiLU+8SeJByduTa7irgsuKPMtJmWlOWe1G7vJElDgFiGWRsEvbtFiAu8sTEmfbQAw/aIDQepnwy2U4+7RSXiWvHTnm613v1j6hf8ZG0UR2KO9CKzofZ7MUbbMBgEOI7q+2IC160sw9pZFcfMARB8EIc3DPN6FNRbCVVaCaxWGXE5FslDu83Y9Bh1pYSm7km3/42axsOirZb346NrXkSoilo0qTSf2KaFNP2jXnlthCk1wWLKm1AGnGIfEgQvTgmFSoF3yBVG+yiUSNAwrOtTy/cdK5cY21ymwb9TN/XJ2jXoCAkUlWZjLjPKPtV5Xg7IK/UcjG73xh3ndGcRkdXQEFFZVpFC4lFARwa8wGimkC4QkCUp9cqVTxbSQxogjvLPZQhb9TajTh3rhhGkFXnYreYshgoHQk1lqKLPFJxc0Z56GOdIUhYxSIr8aoYRF/i2fbm4g3ulUXrrUGfIcBgtGXkJtlT1/azAVC6uS3auhPzCogGzRPJy6riKk+I8IcK+wYB/lC1jyhHg6fOy8/Lx9vceLvyqvE26e13cWBdYI3S2NpwiioaCghwA60JHwjQBghn5+wCJKbtpkYmHTdUZashJ0kueSSH+HhwYyXkJxjguSK4cauo4D8oQytkqKt5pzyEXIItHs/cx0hzRA0IbUsFAjzxo3zq6/g6woJKHBb0pE13OatuElSC7MeVcqrZIhP/J+s3WM+oXCcOI16OmwhKoIwVXD1A5KJcNfk8IuTeIXraHp+Ubsu3FtkdL0+2q35+jWU272RRS1ZzuJHHqScaD7gzVbWFRPdpoK7FYg+wjqSmN7HL/+94+/CrP4OIcKAE8vvl0a3t4j7NQHQcGBUjJM4CVQFVXl4KWyFB/pdjbX1hnL02Z5tNmVNgzXIqbGCbDHviiMa4vEXmE0pPi2QSlJ+2gpUYipA2kNgMUWyFi+Fry0q2WPlRmHASjy+W/ilg851YstlOw3nT+kXv2UPowf725pvYAksOk45R3+wWeegzUOLKIe+CYuvWrpVd/Iur7Y3777bzzhyHu4St2EgVJ5eRhrLSeLvFlau2qBKb0SGFEKl48irP4V7tIASB1EGLuWBLo+QOQogjJOnihcbbyaUKDmi3r5PGtFbkF9olKE9QrWIBo1By7FSkVVSOD5nXnnvN7nriOTv4kIMdfGY3SANe2IkQ5IlOp8OxsCW0uPm+0vyttTyXau//+UkhwBC0aSyYJBnWrWsndDSX2qL1861VRisk3NHqwMeoUIYQk3NcJAaPoMBN4ogOFs8OgBCrPB5l7jyASyxfbabYNTEBNPAaugDIdO9M+ISyUZYuWyaiq5TeV7/UQ7Dyu9dijhCqqxG6Ue6uBDeJqGdEUDQ5UBfQSMcaLg5HAZyUc2JbWg51AxlZBNaIkdiCkCmAWx1ILPCNZ3tYCUBWIXSt7a8U0GMQ/E1JSrZPl2+0P777mV1x5c+t1/4DQSg4SJK4CJMrOO2EAlb3BViwOuu9fgP1KmMQlXHqsxi0KPpZl65nc7/Wbjuio528f0MY/rAkaEc0viRKaEsMMqHxCSgicn3tq0L754IddmLHBLv+8GbWooEE1pmM5KkNaxk8qlIoqRLx/jSecOzFuI8Xj8uJDNFPGJGX5EMxVFoc8m2VUDk6py4ifjp9fekpI6zLTbfbOaefZl06dWbtDFEzGqSwbo/sm2Ax5C2Llw7qZCzhyFFH2MMIV4/dXmJN2ebLFL9OWYQqpP4nsSrBHX8ohzuH5SiGOiu4R+LoycGPexd80+Lm3rrFVnmoctrPCCDYbWhA9SZ0dQ+hP0FdialFWWmgQsXxV85lwF8ahyAVSRn2/uwFZgccYoPRjkkEjkS9C/mJWhXSFsUnS0QKQS3dbfWf8LLD76sj7KWboEf2UmY/pmyWL1+BalWxjcSH7pO495u1aj1GQhkEtpFVaDBo0MpBChWs7pWsOgKSyKDVnuXQvZZx063b8uDTyG+B+BQMPwO5c6oAkDSku/4ic6//s0Bzt0HVDCA+WOlVKX4+lbvuXFGXnQA+QYYeeCpDCl+HNNGsxin4ZyiFz/X3T+Y65HftL39t/fsPYfukrVIs1LAcSPncg5rtSyBVCcwz2gZC0oIF0olle6/wf6Mz7OQBuCio2M7pPALKCK5LeFyUfRJWrVftiLbbJ+HGif75C9Te2B5Z1iYjCUpSPIV82ltsCcBAImJMomajEzGywIRNTUmyLKjGDH4JTGyR+tvor2K21AjqcLCp02FU7kB/20vTbcoy5GKT8W07xmz6rHmubv6PLKfUFaoXDo0h8CTh+BIO2nr07GbHn3y6vfvJe1B4bH+dBw+QBndlQj4OiQVj7CbxzsPhiguBAZEcauNd9Rv3PXIMd1dHX89gnJUSWACBCnnpYK4c82SJyZiiw+f2G1Pm2zNTFuD75BbLRFRKYlA6/FMe4nUGeLWWCtdVgX34/ltRgGq8fr4z6qrXt50M4fntLv/dfQuviy9/3cbN1qt3f7vxtrvt5uuvtRN6t7Eh3VtxMsgqVQo4o4MoUR6WflZ53pFJ+LBooLRlluyT5B2zEUvRCa+CgM4hQSHPEDC6D//uP8C4ANKdFkfWRfAf1kAtA8WxGZaO+Ix4jmXIZWmlXrgh3+58bZpldGxnd9x7Py4NW6PmpFNt9YVkP2XZWtSiMlQ/MBECqHbPtb3z332a6ohhNz5O2KvqW/Fh9R3ukhMoeuf9j9y3CwalQokXsh2HCgNpFVC/WKjARLap3xTl2JQVm+yMvunWLQfWABMRIQ22aaW2GQS2ZmuSrYU/upwDnR04XyooKJUFJvTYt8G/amotMeLbMHGTtchpYEkgpjQOSEoh/XfE5jvxHFkLYjm0pqjdLcRd5IOfYvhXbFr6xge1NxDhqOn88P6qbrNgjSjRwKNU7hpgsv7i88+z4c89ZYdiCLYZEgflMHjFeXb80BDc+eEMRsKXunev1XVUtrRH/8SP5LgDiQW5K0gBaSfb15uL7Xcvv+5gbPIUpBDQDpOpfYWAr+eRn3tV65/IvvFl1wU3+h7+zd/7dLUWEvbS+QQJe67zVhmLQbmnjCMrsKf4KjC8ET59bRXxdfBxIvP2zyKthaYkklCOTa72rTrZ3fffby+9+De74un37NLRA609llI4lYe3gkAkONABUjUU6YlhhiKSWpYoDp3iJiJ7Vw4ScFsPF8Whwdqq+gO/CyaXawN/PB8lAFX9DQW/5eExeMd2ihPBEhj88XDBN6MB8a/ZX9ukucvsxNPPtFEHHGDx6Y0sr5AtMRseyZ8pBCbPg8VQz77fde9D+DvPSPff6rpqXH06f/Vxg5NCjIlitXr10oV22Xm/sNcu6IH9R2iwAo4sGjS0lWs32hoQW6dGyTgqwhE3jTyqS6qlwgLgHAzfMIn25YZom7Boqz39Lls0R/sKUlIwbptubWWtmHgry5pZ6RSJ+axycTLYKt92Qhvr3z7ZsrEL6YSlQUqSBdACWgovdHS3ZHt5SZKd/5LZ53e0JN3O40BM905/wtsWMP21o6DtIHmJTjmeI6bZunbpaKeefpY9/9ITdvnPRnPIgU09KHM07hmDYARrUG119vW4Udqa+kQm0Jcg9+CLW9t4obqpWM2FKORs4zCHVYJHvq83ltpzr77pxHWuuOxSO++y861FkxZWtAN2BC5cZUkoGNtQxhGLZXj5kX0T/q22e+UbjpcCRFtbzJ3fKZ1+sVLfCS9052g1T878kVpfj6D8tNoqBA2vu7N9HPE96hvUSJ9/eBpXFt8EbBdffBGHIfDAwFF5xXmW0zjbLjrvSk4ND7B7n3oSOY5pdsZBnbGY0sLSxUQG2ZVyUCBrIQIu8SnkP1W+H+LgAyEmRUD+D9BwNCCWCWLxu1Eq/V23FRYAawUXL27PC0V4vb/vvfpbjGgNj6uLtrLiPwlYoVBi2MJViicm3hH1p3auHdJHLUMkJA9Ri88WrraXp8617n0G2g2/udDat+0AkzwGjYHikA4yQu8I2SZKjIKrKEHJrknIXfcKtcGRKEpp7tQ3SKZU+UTCjSBP4xID723BgsV2iG1Gy2B/K81fhXpWir0xZ6XN3hJnp/dGaBv7jHHw0jpxmhsHglxfkGIfLt1mdz0326G01i2y7Y7/62s9e7a1hlmIbCALl56KPBwiS5yUQQ3HW9525AZ3lNhadH4/m1dgl/5thdmLM+yyYbl28uC2iBpJlTLP4uU+lMO1LfTTkW3j7WPqOW/R19azVy/6C8RM+2NBFBXIYLoAqe0szNAW8VsDuBe8gASBG0kRaJGtZBvcAIH3Ky670Po99YSt3VrKwRa8WuTjHL9Pa6/2IRp0FnyNv6M0Xb7qLX7AZTDz1J8BTAbRJQuqhVGQEKBQbWXFNhBvUdIH0l13zpcUQ5MI4E6IFtKDicSBXR4LwJKthfboO4E839jDx9rVV19hvTnhlTGIUjSv4tMgRGgvs8jx26XtpHlFY11X7OnPt4EbIUHhhEi4qasMwa1gM1bqOD6RvyqRv/dX+e2tDcBrK0Bp5LfXB5+Hrgrh+ajiQq516fb6PMKvmiRSnfPY3uerOBrWJLRQ+mOo4a4HH7HBBw21eCZ5IfJesUyePn362cPokM6eu8D+/t6/4A++Z2N7trSe7VpgyRfej+PrBKIFZTD6oxFBKClFaBZASCNfmdsqhw9WCTVUDOM8Hv5RGaeGgjTxqAI+mqjEfRc8gvAlOErVASk8JE0iVt0q6WVy4FOBJRSZI9K2SeSN1P+SmZCFHOgsY4WesWKlvfXZXLJKtCuu/qV169bDElJS3alnJYcdcugkxFOFFkico4bxrZK/nbZyigjA5UNlK/gx8HXTVWMrS+KyCF3foHGVTFh1PnQs85IFCtRN+fL9+tbHn9uYo2hOxQ6rSM6xP3+2zuahnXHzwRlsfznMwGI2bEzbAf/uo8Uldupj01RDu+qCoTZuaCvr0KoR9cJRFZpCxqmxQwMsGDrcYbZy5Wy3EYLBOD/vhrjLgb3K7Nxj2tuc2X3tb69/bYNun2LnD25q/bDG3bdpLD6MOU6hP6Pxlzz67J520in4MDngYDQ4cCaFNpAWkQrgRouTYCQlRep+bKllWQJ4A2ioA8uqvgM4AbKCGqUPBa83//pmu/E3N9r9546z9IqN7hCGgUZ1Evhk8Y5xfFEdhCmt8CFIhvYKoXGq4ogAyZdqMdeccU63eHYGBioon2iKj84I8IHMIDxyHfywijJ3dHimtNG2HUH45ZsK7KtF82zS/LVKZGefe66dfOKJOLTv51Qj3Uv+JLGrCA/BnIXnGjr48DghPI5/J7iRL2n96hsKoZildimkpvTqu8gQnr/wmUOA6pBw4PWJfAb+6hP7q48XedV3/Txy0nc9K/i8b5AvOgAAQABJREFU3EPYs39fVzwfX98V1+fv6+7TK56j3oClVHxofPXFVHg2RRx+YJNM4gFQNEUgQp2QDukPBdCjG9Lyi+yDjz6y216e5Io5YQCetRqkWAu2yAmIkyQCaNh4hF/IFpDtjjslRaZMesMCZs1OrXB6lqiKThSdPB+5+fa4jPfiH5+vJhSlqGMDYNctdXE8e5UHwiuHFxYdj3FabATK5t0mkPVK/AfPQF1uytzFltW8nZ1/0ZXWk4kmPVxpuRQBTOpbN5E0dvyXRWinRgYhI+2WlnjeE/WiHYQAydfJFav6RACg/x75XvEV9F3fXLmUrfvqNHzXiX0pfQxtYZtWL7BhjbqxtY2xhz7Pt0lfF9iTp7SFEpXqVwmUR4qtLUq0v742zx74dLPdcOVhduJhnTCAy+FHzFb4dlswZQb/TtQUo6dzffUhpBwXHWxBxaPeKBk1nYomsODlIFh88MA0G4AZ+5OP72Dj7/vM/vTXWRbfLMvO75xiB3duaANwt9CycT61LbYnH/+ztWjXwdZu2mwfT59lr340Hz806IljfWjcmJFQd+nWFR6ZeNQpySyswlVQn+XsM6VCqk7XQqDF46yLLrCb2LX8a+oEO2zwoWyFC+SzleFlEUKURB7UdCgB+Lk+8063XB+CUYX8RO2xmlGGQ2fEJ66mBCfZgaYHqn8gY7oYJqjgCb104GUDyHvd5q32zbrN9jb68Aqde+1nt916oY05/AiMOHSHLREscA7xuLq4aO6PH1c9qJza5qyPrfrqFx6qYSDivY+zc/4B3KiM2kJ4/j5ft7GrLfLeeOcLqSuvPX2PTBfZOZHf3bNWPYJMjisUIgOYhqUPhyzUifzXAlcMIowHMXbFXFdXqJ5jjz/G5i2Ya2++/Z5t+nSOS9unRbL1xVtXFlZfZAggj60O0Ap1wbThRLQ8Rk5xACjELAI5PWUfHBQEp11BXVxm++CPKAUFbWvLENDVtlcURBynmWUc8qwvRfA3tiOkUJl9gRvGJWu32oQv5odqEmuHjDnCjjvhZJoExQciK6GvBDy+n3V1Y0QzpOIkhK93UsbvgesDBR9H22AhQh9qG1ufr49T21Xp9AuPqylLr3LyCrLl9HfU6KPs83tesKjUEXbr36fYxF/0h3Ivgxpli56Uxdas3IbfNdmsY0ub+uoxCDlzUFIF/6wQNCGxFnYEcHNBOgJ/OgyEJX/LwYLCRHVIEEQEwqAqUEawR0CGVcVQSNwP651kbzxwqD318n72898+Z0t697GXp3xj/VdjoWY4+mtJDeyXN9/kmnf8kDgbcuBAO+ayHNTTzL7BWOht469DLykUUrvZ724+ycYg9tK1c3d0fTnlZYejcZDpuDIW2xZYXZr+8mu2f5/uWEl+1S4d0d3aN2+DM3OOdKBaJXol4WbBnDuc0kLs+lBtEX2rZZGGMI5CqNrNaKEW9ERD5Yk/KpexFfy2bS60fF58PvN9+1hs0FBo2bWL/RYqtHffAdapS0/mRDoWWKRcx4LI2EsYRvlHQnz4OIay2u3lu8LNbjMN+xiZ/z5FgGHl/mC32pqVoPfZtm1bG3HwMFuzeo11zW4KUIknqa0Gow1gaTXUwGkFFYA0hGk7rEkulOGBuCTc7OygzZk5wx6dOGmnujdNIO9mCNzCJ8pBPUr8HJm/T0KIFgVKAIoLk0wnxwqaXj7UdL7AJPyLYkSCTm3fg3fVQBUC8kR4lW4gsZoCXsc7HT4/MOU0fTHmhmIW2KdzvnZV0EJ9/Emn2P79BnFimm4Lvp7LFoeVXmaZCDFEkCBrdU0oziEj3sjGnVNXYhu6bv1659vFJeKP4oQjP/9+b11VH3HQYpjMpai3HYi2x2uvHGu/ueclu2pMH+uUXmFbccWQgSWe6Zuq7Ih7p8HvHW43X9jJGuJSsQSzS5WVmLSI24ZsXyr6vSAZKDttOx3fDBSncQsCiIEtlIL0TqwqmXggZYlPcQChawmHRSmphXbFGQ2tQ+5p9vgby2zKfUeyuCyxi6551aX951Pn2CgoxvL8MkvCwrhMwYvkKiLvq8/pDjwm2ryFa2zC58vtmqt/ZdfYr+ymG2+wc8873y3e0olViAK2KqFqe+/fzb5evMze+PsLduejj9v69zltJYxrbdauHepk5C0KS3rkaVjsjkcdrkKCxmIEOziJsq0s2sVoghShKxwFrORhgm7T+q2oNW626RCV4WHsUUfZL/fbz3p172atWrXBwVkTS4VSTcQRVgU8YIGvdlUynOEWDJ6dea3wTP4D7r+XV7ja2qfJoJ/s++m6p6A4WpX2hlc4lSVH1UJ0spbywIN/wFjlPDvm+BORYas57HEOwamaVnxtmYUQnbNvruVsq8T3ioOaKS3AXBDbwfxNG2358m/gAX3FYcE3YJj11c1qyl031KhyWaWz2HqlwXSXcG0mmiOJAJ9DTBQhiXwp3mu1dnsPngHZAIBcbkQKUXRCQTJYqlnpjCEwWYO6EoV3QR9DmbBfKQchbcZ4XQE6yXkgsq9XrbCP5q2rrl9iZgM7auyR1r5DJ2vUsLGlY/xVk6qseDuOZj62AUOGkJ/oGv2jTFE/1FHlBeOn5SFgaUhXmXMFu/CCS23ie684PU7xAasRcnWpwY3S1+YVLiKae1Rc5SO48Ycq/h0Vcds4oaRKKL3kBKwGP/+KnX7O+fbPKwfagAyQE7yw2SCbQ++abeMvPtR+eUF3DrdKoGaoO0PimPq0Q+pm4ie6EyIQoRt+mu3aKtAh+HKj2BJKzFmLUxR9FC1kwqIpa0FyXhWDVkZccmN756OVaKWstzsu62dzFqy0i26fCatlC+5Wj7RebRMdSwHalcWG+BhviBF/No7FCREkXLrYyo3FNmHKZrvihhdc+W9hCXn0mLG0SeXx07BQ92L4sFUs7mvQ8Plm+TKHwD6YPNk++uRTKN9E5F6xFM73VXNnu3zq+tOm237wHyXriJMv/mVhS2/c2MOtaZMmtDHamrPAZzXMAZHCv+WwAGzt+l+UIlQDMAjfHgSaTjp6y80hwQvDV2cQVSsTVMrf9XWdMYP+1wGIfn4s6oruv4sHKIs39c3/e3uFq6tC/+730opwXudAZgMHDrJf/erXNu6oYxxfTv5INUZSUyoPrfJu8jILZMaHsYRPhq4Hv2LeRXNYkIlZ9KwGWdYWBDJwyEF2GivounVrAGqcvm/ZYtM//wzREYylfl2DFH0f9GiZzcSD5gRwWmZlsJVpCJ+FQkB9CjJ5Hk+5Zazwqpck5v3hggy4iqcja7lCQe6Ei6sotMUbNjuzRqUgvAXbXVZhf6LsmONOwqZcC5y951hjVu4UeHuS4ROCK3JAIhcBohagGIsKEGDFFLuQnnJRX6hcHgRM2tqrxqpPCgC5ecMaiynabo3Q83TxQyWrH/cE2KGoe7x4oPbALF8ZAVXGNhu93OIdsfbJ54GwcQ/GJi4ePxU7qkB+M+yac0bb+HNBfgmbOcVFBxgdaKVxfU4d5UfYNU6NUo/QCHcb1hiP0KtkOEG62C7QDxq20LbZIUeo0codm+2wIY04aIq1CR+us9HDm9jzDx5k9z4w03qPedQmPHmqHTKIw5ntqNVVwGqIL7RY+tKgAEsQ3E5ABqtTk3hrfWJDO6jvxXbPXz6zMUceaffcdYddeNEl6GEzdkK6wHMJLhi0a2nSrKk1bd7M9ffBw4YBP/ADGSvxY4UEdLC5YsUKZ2fQqb2p2jRU+ciKTIsWLegzzGpBeIgClpC/KHghETfmggEQneIXk7c6Sd3jiGTJwDg2k/rRgYtL47roP/CPI1D+A+tdZ5WddFZo4nbktFdIcBVUUes27fERIYY9k1q8sjCAV2YOL7mb4IPbGgHxpQhJBxMbQIF3lIzIRceshqSHz8IWYPABw91BSwmCsJs2bLT1a9aCbORbeK19wmmlHP0ofIUlDJuzzN1//z8CxTjrP3iYndqxI20CQPWDounXfzB6rSnUjzhsiQTERWI+OWAVZSx+pahRnERBBehUNSUtm3gBYlBb/aGGkI5z8UkfJOP0pQz/wvfe/4Ddfu9t+LPIDiYeHbm3EF9kv6juqq8OXRSiOM1OTGhgeVvzbOXKGe7dk5/PtjxU3+5/daFdfvkwu/7sLqhjbbai7RjMleECEIdrm9K7FMHi45/cqzr/KIVDj2Ex/LOu2kSzleQkedTAOA4Kkmw7ZvdbZEbbHdcMtMbNy+3QM56x1/96mo0dhKoeiMkh81jsZVZgvZhhkcGCopItLCrpUIvxdvf1A61bzzZ21S+us0VLV9otv7nFcrCL5U/GJVajxbB6waH/EziM0hh4S+kav47AxZ4sJLtDCyE3fuprIVDlIyToQ/i9f+d7oPqZNMrjPzH85BCgxkHyVHLOJCOVZ511pk38YAo8wXZuuRJfBJirVwgGVYOrU06AlT9aGSXfpYMATU5DMDcef8GJmKBvkNvMOvfdj0mB4hLAfuQJJ7JiY1Fa8mAg3R078qG4ArebeRhsKMDMkQxWeuB2dWfCqf5a0ZW/8k6A8Z/NlkTiQmmIqMRDuWkrHM/7BEyKR7Pdz9+2zRbMncfpdzpUgPid4nFp68wtWzchdLVBAO50R1n9RQG4skLfFFeMdzn00Zqv9ssum7ZLxbADHv7D/dajSxs77rjj3DdNIE22fRVUP1E0urq+tnRbvXGt/fO15+zNNz+yE4/tb4mdOtgttz9rF5w+ym69oCMaG0UccGEWDFGgGDzEVRSmIspULFpvV1z2HSqu/vOTXX2kPnYEfWGmtc8u5XCjHB/MjBtb5ytPOQBZ0XQ74syn7b0Xz7KRvfAlsxlqHgvLFXiswxoxseFNVmLMCmHtQjRRshnzn/+sqTXPOslOu+xhEH6s3X7Hne7EXchIyM8vUKqH6uODxkPPgieFcB8ZevZxq+tPeo8E9V3vVYbP18fXt59q+MkhQOeMCJJep5aSjxt+0DD7/V33Ym12FDwXMbQVBDQhwBEQEY/LrsGR+loNoYo0fwAuZ3yUV84XsChJJkCM5Ow4IdYWuwzKS0xz8Xm0MieTMSWQOMoy2TY67OuKF59PCCq8YId2iA2SEhJ0aeFZgbEh7hyykvnxKihMbVMF+0K0ogCjaW95cSEycOyJxfSiVJnc11XFKb5IDonxOG9jPKsMUSSKEKz0ULXkI4TPEuJOFpPZ9ublbbFnnv6rzZyJvbveZ4GMQMBQx6qf8lDYuR3u1Xf649rs2i3kizl6KM9Ytpfr1q+xl199hQOD8Rgb6GqT/nai9e/V3KZ8tgl7LC3txkvbcRIMFYOArhBEdFQBbYmnfaLgGR+4saBE6hSMp2v0d6pheCLBA9aTWVRL47ezOGKSCtiLlsocO4Q4TEBdfkJX25ZXYYec8KwtmHi+dWzOoleYAvIDCeKTJhojG0kY0nMyp6hbFrHAxmPT8dRRTSz2wdPtpMsesPbtO9jFl1xKH6ufVf8g+D4XXEbe61ljGoxrEN/Br2A4FN9fFUdIU/2m8dzbY+rr+2O8xowfP/4mVawa8AR8tfz8arKnRiitOlArVXioLU+9U9AAiHyvT1AaDZYmR3jw+eudm/bE02TOgueR2xy7ZG9MtD77d6VepJcslMwegShkv0wIwZnFAmgFFHwMZe2whrvXnS9D9+6ZejsBUyYWNJbLx1lmcTUQvxwqDPwg4WTxboKfqFPeU45OihVHWzR/r74L0gWAKH6mTIRLA0DCtKIoJeLgjKdSD/WdjLSKwlu1erk7rZODJ7k5jNV2X+0hqO4ururKKzng3rRpg0Mw6eidyjq1mPyOqiGC+EKJnKCuXLLQfn7l5davaZod1a+rbYC6OXzMKMqMZxEIYEX5K28F30e6qi8FN/WhEhVXP8frFHIFQcQkghDwDfH2q2/bsBEH2+ZN/7IXHjrPLji8s7Vrn2jLVlXYgHEv2Rdvn2jtmqHutgM5NuT1WD7Ii/FQ/UCAQnq03tXv+yI+tasmBPeVwE20YMfR/iASLVI66IAKlAGtbr0a28y55RhJWIWtv3YgPPjMksVyVaK2ai8IOjicwa0n1GAUenxd8GGSld3CrrjmLiez2hlRlJJiLTw1CM/3t/rO3/s6ak75+5o6h2AmNHY+jdKL2ta1trH08Xx+iqOxrU9QGsF0XXM2PG/dqw4B1e/5r7svRWkENyrDh8g8w58Vx/dNtWN0j/V9BuHXyMTh32q7V+coP9+ZutaWv/LVN/0U9Ly7EB5P+akcn4fPPzyO8ipCEHrUYYfbzFlz7ZWXnrFRI8bhsCiNDqNceFvyjlbOyaKTAKUegXBzUJ/Iuri3YXUVyAeBK1XXs3sXaocoNz/hXB8qkoK7CGCrH9zr0Idd+yGUXyi1ixveVaLGBDAKQfuFmALkF96nqnrQX6JRQ0HvEHGpwECADmDoVgTFEb3ASswHEz+yx/94v515YBc7sEsr50DnC8SKyui7WCzrlKMJozorT7/t8tnqncbEl++v/nv4NTyeEyYnXXwKWgcLltjtN/wG7YsX7Yn7zrCjh2WzHYdPlYc4yY4Yu+WeWfbiIyOsezuoxTwOJEB+jsoTgnCdzGg4Kj68tL19r7IIID2Z8XcBuAwMaDApUcNrnF5u9/26j/U45BE79rDOdsyBDawoJol0QjhKQQ5Ki1ByANepLACwWHYU2dnHtLU5cwdwsHWCzf7yK2vRviVUPlSuQ57hEBFQ9Orn8DkQ2e+Rz0GFaxYwpfVzyV99HFfTUP5+7oV/q+0+vC5CUCpfacPLCU+n7yo3PF3498h7H0/vlb9fbGuru+Iof59G97F7YpQqkYIYpHsKyliZKng9Pv+8p7RF8MO+TZCaTF15+wb6/ORQ58wzTrXOXbrahPc/Rfj3WMwN9XEWNwLqCmqVugthKU+l56Q/2Hb6TOp93Rko653se0T07U3FC514k6qBO8SpQXM1uVdjPgGDtsYyognVxIySeaMkvICtX7MS16bP2bSPJ9tNRx9kLXOgzgvRniDvPLyISeQiURS7XN4F078m/4g7IcZI5BgRxfW3b4P8oyRGJ9rbb75pxxx7rA0alGIL3r/MOuWClDlPKthWgpxxhn3w3hpr3DbOjj8ol20vSBGz9RWVGAoI2yL6PEMgGVnsPnoGfpjASDzRbk6hk/BQiDGJ7q1S7dHbj7VjL3zRVn18oTVGdKesTPBWU40qBOpjAbwSeLCxOHEqLoxHZW+jXXdRf3v875/aI398yG79/T2IFiGPGPA0XOK65sG3nbNeH7+u/GpqGtztzTkbmbfGbk/5K46vq5Cq4N8/R+YX+SxqVL9Y3+jICJHP22CyewzuC/YA5uOqcL1TZaRrV9+gfKU77FcGpfN5hd/ru+KqzmpsfcMOrHVk4z7y8osuskUTX7KV7zxuf37tVTt33LHWpWMHa4igtLanYv7rUMKJDoAMg7U9DGPUt8AfOh79rq2uFh2nkUL5enZbpUgEBZIDtbufGyueRagmg3iKdmyxyZ98YH9+5CE7sn2m/enU4TjdicVUFOqAHPa4PrLtbJvxjwGVqJPoPQXpZ4bLZ3kAdTCkxCGYUZUcGwTK5onH/uJ4XtdedJhdc3o3a5C9w4rytiPIDIJj8q/Bm9h7s8vsxvM6YjoKP3BCIlHpyG4i0M3WObBsHMBQ0NYwLLOnCu+F75XwIlGzh4pDCDtqm3PAFIdF52MObmIXWFt77t2FdtVJXUzOx9X5QZ9ArcJ7FW8ZVI/ANIYDOOQqLkiw9i0q7LF7TrBzrnrUTjzlJDvgwKFunvm+rKvK+3rOKv/a5qnq499/1zkrmJHcoCg6ByseTiIaq/y1wErnvL5bcmUheVNtm6sPQVTI7oIapMJ21+k+D89MVVwf33+LLEPf9U1pfNzIOP5Z8VQH/Xx+Po1/9nH9VaBfwcqahGiI/BX3aByL0+UDrNeKjTb5gbvsL3w/8dRz8FMw0DLhF8oYZgkUkfQyddgg/7E6wAjMycN7c1sdXtFdorICRBOUpjr4+vjyf5BrqC4qS/UOeEoAIfV2o6pOUIW5SmVK/9SmaJju8v2BaL/NnDHT7rv7dxiGKrfrxg21jtmYeWJRKET4VupSYtrLSVMMZs8DYfOaLZcfw8i2+r7wY+a/+3H0/SWKKZrTzk045X78L3+x39x6mz37wKl29Eg8n7EVL8bQQVUC6myIk8fjh2PajEI7d2wTy4CPuSVqh6WVshhG4Q4StMPmmTHwvCC1Xj91wA8XKiXwThViUL+LrsgEQec73m1OZoa9+vAAG3fxs3bKId2tCW4LSjhAcXWkfytlZZpxiSvnZB/PajACOS2Ox6RbmR0xJNc6dDG7/64HrF+/gY53q22e7+PI1um9nyd1xVEaP2++7ZxVWuW/u7x9/n78fVk+jX9WvPDgv6tOSru7oDwUT0H3Pm1deSue4uindN/KL7AvRNe6gjL2FYm831Oaur5Hvvf56+rvVVbtAYYq27wKeCyFOGtZswLVuC551jY70VqdOcZG5W21SVNftCufecyGjhhlBw460Fq1bYt4SSKdJJ1aIRSA1J0UB3JxEiNx5vS5OsCBX6PgAa72euzLt/BtaH9gchwmPDJ+Feit6kBGltR1kKIzUPmAEIUhKf74ZITBEcB+7/VJ9vWmrbZ09nS79FDMzDfNYguKrwsZsiSuHFjrwKVCByR0QzEWQQKH6jWApJbV3f9BuzVOCrqqn/zkFUWveq9ev8puvOHX9uJz/7CX/3SGHXWgDmbYDvKdSmCSTInjnSe0/Ts1sZa5+c6BfDLb5QrMNGn4xe/TqXxNEEzUBRc1sfb2Hd0H2oW37CamhNzlRCGOtmy3of2yrFeTzvbu1G/szCNzWXuwKMQiW8bCIjitZIGtABEGgcUepI+7aWuIX+IbLzreTr3873bVzC9hDQxwfen6k3FmqGoNvt9r/Rh6qbHz41HfOevzq0/+iuvzV3z9VM6eYEbp6pu/z9Nf65t3NQWowvYUfOb1ibenOD/cdxj0IADx99q17WTTloEQkL4vE2UTixtCbK6dNHSoDe620eYvm2W33vqu5bRqYaOHjbS+XQZZWqM0i0mX8DCm88uYuGWFjqZwZqe0d+R/te8HGlXfPtqb7VeZMnOlihVisr+hVLhACpzFG5bKne9heRFLhDIpLS+yFesW29LJS+zJPz/iqnHW0K528UlDLAPWAufrTDidSorCdXifOz3Rb0xQqQlK6FrBA9m3abPSCOkJCeqqw5vtOFj6/a/vsBef/4dNeAJDrIM5JCjKI39WdpCabFTJB4U8niSxKDVvup4x5IQTxCIlQ2pKvyuesECAaLn5kQShQigZqlXGopSKkPT4a7rbH56dYcePzHX+nysQl4mOh4cpnuAuFI+2xfj8xdn4sH7NXJs++nCSDRo4wPW/cncrU2iMwhutcdmX4duM+3epx77OX3X6VgjwuzTix5Amlq1SBdu37j172iYqVAEpE4vsnrazeZy8JbMKd85Js/ZZPW1ol0r7avVae+fJJ+wZe8KGDBtuXdt0tNat2lpO86aIhqDWhIiIVMMkkiL+g6f8NKElPsJs/EGb7RCRmOKUux3eBntUZ7svUTJpiTEowBfZmrVrbM2GFTZ/5iz74F8TTDrMl43sZp0a4xENSzflpClhuykjmKKjmF2uDc7iDPdqkRx+F2Ma//tOLFHNylf9qLzEc3zs+afs9YdPRKMiC31WTJjhW8KVqtNR4paLymZfGQfDT+JLsNToa+QaRfAFVQ3d/LB97zppt3/UlyJfy6D0MDeGrOawARl20l2FtmhVCXYoWVhFYfM9lpPhSNAR8ozBrSeRrBmqt78df4hde911duaZZ1pjdHcFf+yea7pgt3X538fIHvivQIBgPwdYSRgrmE8PFGP2R5MwBuBJQqG9ECsf0hCxqCJ3GHBQh5bWr0MujFIsJc+ZYX/+YJLrt6zu/WxU757WuX17p3GRnZNjcRweiHmtlVi8LBkHcFMQpBHMSz87a/gTkYNQ/eznrk9S/UE35O+/u/f+QSwHaB+MMKSnpyEnVuhM9+dv22KbMC66Hr7aJzOm22cT33apRnbLtt8cO5TJRJvZakVBDe+AupA5hETqDJErfBOUQMZuu6oC2GPJc1rTFk3RIdbhR1AfIV9PCQapav+rOA5x0ja3UCC7qDyee/Zpu+mWO+zpu39mh6NHW1SST34ywinERxTtu0mjLokGEVSyv6xAtzcaazycE/IhyCcodSdsGLz6MfylQyUvGuBBnCc1TLDbT97PPp6x3np1bE8fF9BaIblQQ8PqLO90ztwa1KGllLMraeS+rkUfXQhQ4xMsT+qh/4Vv2wP/FQhQNA2bEIwCJFvjVp1t+5ZN1jQnA5+yMJuRaUuM16Si6zgsqQIpFMuxMzyb7NQqGz2ktw0fFGWbCsps41osDz/9F/tbqJd7DBhkndhWd0W8RsgnKTkZMRJOvyWErC0eW2+nMaGsmcda6R0/kWchg+oNG7Drd3BCpGJjCZwdW8cxdwIkE8vWU/5OoDlB4NqWC1ORAkp0B4YZJNj84QcTOSSYaZNmg+pXL+Y7JrC6ptmwo4dak9QEy0DMApFXjDohw1cM9RGDhRLtNIlXghmvGBYH7ddUP/9j7rINq0TFbLsV7ki2bflbrVFOY1IoFX+JWxcS9IjP84AcQUeGCclR9t47H9gll15uN192sB0zKhdqBl1ZrFjDCaN9Ae9VBy7O0ght1ihW6b0oVZE9WLuu4X0FdXEV+pH9Ub1d+2WcVuw92jR8cK7d+tBUO+PoVpaA9zqrTAUG6EdIWtcSAQABpgvsDLEkODDBpFv7loF0xVtvvWO9eu2vznexHMC4FP/782164L8DAYKMyjFWkMMp7yXnnmmb377WmjccBaLR9gFjouCRODCQABASA2oINSXAEI4XxCMHCMyyXHxGdOjaxtKy023DomIb0S7HVqxdYV++M9X+8XzQ5U07drWu+IPNTMbBTuvWlt2oIQLDnLQiMZ+EjmdcLHwrdG3Fj3SiyLoHcp0PEbJwvh64VoGRHEIBKwfe0GS/DgTFoYV8LUh8Qry+5StX2vI1a5x9tvfemkzKLa4iQ1sn2Gkdulpa3xH2+eLlNrxfJ7ZfGP1ExKeEdlaC6HWCXYWWhKyjBPwzEI3arD7QnCIESI3tGe5DRRZGJ6XaZ++9Z7+++Vq78/Z7LTMtM5jYyovv4WpaQQ5C+sFJpba7ovx0Qh0Pr3HtsuV21qlX2dADutlFpw7gYAaXlmg5uLKh8mSwQcuAkK+vkHiaQh5wM/kUWiV8QT9qDIDALTAkNoMWxjLGsG27GGxVxtlSLPvs1zwdY60CwqBZvsUaBmmVlEu1EXW+hJIY+J/JdsoRzW3VyjWuX7T4uKbvZhGq7qL/3ezSA/8VCFBziANbbNnJ+U2qzfjErHt36bNCSbAVE7oD04R1jmhGAhSbQ1AArniIkg8sZovZLSfVBuZmWK+G7e3QXj1se2GeFWBscsmaDbZw2su2LDrVXlpfbfPX5du+ew/rjoWOtDRcawpZgARLKTtBYjaCfSFDkJOqUQrlWCqZRJBHjFPZM9u8Od/e/eIrq1y7NKyeZj2QlkhnflxyCFZ6M/azVPJrwAl2ApZQStnez/kmxjagy9syIx7KlvwxEkDRBNpDPURbRIZwis6hRUhRSaeVlmE1s1mu/ekPf7UTjz7Dhh88LDLpLs9CovoJ+TnhXR2ilJfaC0//07bsWGp3jD8bw7LlWG+RLTcQMoSdFh7XKbvUzdc1fKx2KfJH+EKaDaqW7AAWgctirWFqov3s8Kb24YxtIMDGLLoydUUkYIIlycGBWqvT/WAR1MYAHiguP/v27mtX3vKQXfOLq6112zakQ3bVsQJ+hE3/kVfJIcBwgN9dfT0wu1VnNxEDyoEpFox6nTF9ueHxlDb8OTzx7r6Fx/P3Pv+AMgGc0PvtDIX2VyKw83XbOm1NRVOIDR0spT41V62qID2HBEgr8Y/EqjRbCcLbXJ5hMSV55IEQJtRdNlZZWma2seHd2zLZ45wTpq2YjyoAkRVDaW7EWOeamRNt3jdb2X6ye+MH3jK5jsrjFxm0wQQXuDNOnXW2axNvpyAIltFrIAhPupJQlTgrSU+KR5ZJKkZMG8oSpVaBw+8CrFfGMclSseSynS19NOIi0vUV8vXBzUn/EHb146f2S69Efndj4zNt6Yqv7biR43CaU2FTPp5cjQDVz3QTQTkGSMr3va76iedaTv1kvXr6Zx/bL2+73m6/+Tjr1x6ryQUSeKWfyUSGRkWR+oOlsGr9h99qE8u40xfaBUThQGkwDt0P+tk0O/7AXGuYxjLM8IltEmjxBP0YwCdMD/jMFVC/8SwS3bvmuL7YsGmTQ4Aac7INUcs13aRx9D+Nwd4OHk503Vv5K5/wfP19XflHlu3j+XSRbQ7PX9/2uV/gyAqEP/tKhjfCNyA8nr/XN22pfByf3n+PvPrvOinTllMbzyjSf0lEmcJPYMLJ8JPMlXs+S3geovhkeMAF7rUprkLftHQbEuRlgDM8RGcMAciVQKuMD0QjVBxtBfgNwesXbha13RTJFdOYrU+bbIsaDj9IiEh5KWtW/BIQSplkClnFSeHUqLQdFZUoPx9CKoHjGyhC0saA8BRfeciJtoRnVU9RdoHLROotsRemXIqY6ELwOLLGIJfb8jqxEnIlGfFd63b648ZD/EpSqh4JbH2X5BXZC3O32h8vHWrz535lq9esdmkUN+hnZRRMWn3wfR8sHoGJJplOLywusKeefdGlPePQxhbDwQpnvrSLlrPICOW67Xkt9XKJ/kP/wFlxPpirEHIux0xXeWmMdQYezDbYR19stOMPb2lV+eI9Ax+MpcZGC4r6TDAr7ZZysSJ417RJ0M/Okg9vJC4U6J0rTU3QXPEsCL2NnPw+pt67wxTGUsGPnf9e21VxfN5+PipekJcQYk15iufDnvIO/x6ezqePvKo8Xw99Uzt2F3z+Sqefc4u5uwT+W7jOsM/Efwu/KlMFWaWtT1B8VVqm032oK3+ft65Ssaorns9HV8UV3yUBeb8KkNPED951nwuRyUpMxLcDAr/FTD7vntCn0VU0IfDGN/1ny4qGSEYaFBeIrUpbOu3XNNCsyhJiFdSWa0vLOwEtBSoHriAz5DZi0LyQa0INWACwotzwm0veiZQhFOsQmHRz1Y8CSH6qAoQB7yTM7DJ32aogN95gMcdVVESwoPhsQoZxUKzRMM5jESmRX+AozUKXmyLSImWlCcY7Rz265Bgg1STECXcCByLxuJ1cuCXf7nhpkl3zy5sxld7YdmBYFg05Z0HYA2ltYyEUHQ0/Mx2beKqTVN0+nPyBPfb44/bSH061pgj3lhSyCMGHTeQEvhh7f4lQm8W0M4rtuxB3UE8aQ3A8Wnf3H/iHcdN2VshdPS5qLjOl0o4ek2vPvzrLDhnSGDuPjITginELVqdgvOhFtV6puIDsxLQmzJ+zwLru19WqsKsvNwfO9SljGiBP1BtRjXQwRNzaxsdlwh8/r77PnKVy2mC5xbuIsY4H3oS8ZZdTXuoEJ9u3swADmCpvT/XRjkHqbb5udcX338Vi+TY4QbAolVrnGN13xO6uQoB1VSIynSqlytQnKK6EYcMR7J7SSYlZuoLqJN8BdaXR9yQamsBBxJp1W+zO3//eRS0r2IJ6VaIVcOoof8FRIKkgaLL5e26DuUc5ApQoSycvKasXQLlk4CSmHN4f5AvUEtsa/kU7wTSlJ2EorbuRGX5loi9MAm1ttMl1+Trg1rbYawAobVjBSuPSgcTc++qMXVWDXDVJVEfFBZUCcGpSFUAoSzdVIBTXrFC+AVpVpjoEIl9UrgScpViFiWGhyIjCXwUntZOXLrXH3p9tN9z4G2vXsTvUKgZcAbYEtDO8CSKNgwBQJftqB3XZ2SeI4nz44RR9sqF9slh4KAsNiEREeBZtyUQ0B71enmPcSXVNGzWha55c8v+4P6p/hcaCUZZNx4BiL7cTRraxk656z25bUWqdO+N0XHza6jHWokoS0gmOZayCzQL+q9Nt1JAYe+eN92zk2JGWgmmycsSvcFqtyPyUrAoE2IAxDeDCvdzNH+UvBLin+eTyJu6uc5aKxrEzQU70d3c/Yll4Gjz3/NMQFwMJEwQrfs4GjeIv+dSGU4QsZexEZdQ3eJ8g9cUJDvlJTlYdpErs6ec7xl/rqpi+6+fz9dc95a/8fNrd5e3j1SdfH4fGuSxLELBVOGzcMTZ75TYsdSRDyeAYG0pHqyvVJgSDEjkwyiFYYQUouFiESnMnqCKjCAFCCe49ELoP1X+0gtf+Jaidvvl/QTwf33+vPXV1AWE31KM6keqkX/WLsHhQlgBbDMhZzdApcQrIMxm3mqvZej/+0lv2+sJ8+91d91vbzj2tEPNh6qQEaZQA0Opf/YI8QFzi43lYCsGVML0Qn7bzGzdttBt/jbrbw8dagyy2dfkcrXC6PumzfFu0Mh8ZzO1WyiSSf41oWAfOsbervaingOrZqfL/YQ/BeAbjoPbE0c6OTbLtiisOsk9nbxCA0aLgu3i6jIibE2qm71f1ZQP4hT06dkZ+k52HJDhJ51wigCh9PFFffq76q/KpLfh5V512DzhBY67g0+kKGBFi7aXXX7GYHTdbcsU9djPynbK2pCAYkTN4TUU/L1Wegi9XVw9T7gN/fBn+OfKq7z6E5xueZ2331Wn8zb66hlfw+5ahhnyb4Mv26SpCg3HU4WOsonUP+zuyVKWxqagjkSt51yd7GUUoY8tRQeQqkGAAqN+mVj9Q3Bq4AIrqLtN5O9M2mdPFBCjiQiyYTFu6wa57boI1HDDGbv7lTda4ZXsrxBya27YzMbUVbdhQflFqAHgXQBVg6j9RpIanMG3aNHft37UFhhbw7pYcbx/DV3z8rbkImGNeqwT9a9JUSH8PMReHD0ihUjwydBn8BP44dgTILAY7Dsfu38i2wFYowNisM6zqJrVr9S4wKUQTDV+3UXaKbdtRzLhpPHi5mzGuT3f5sVRcP2/qSue/B2kCGBBsFOB7+qI7nrMxQ06wU44cZ7///W9s7lzkUQmKKyTo0/q8w8vVu8jvehcZR+/2Vqgffby3Svs35CO48LCRhHiIgnRdxx5/qjUcdrbd+fibNn8tpD8HGs4FIOMpWs2dloLsgqDVlPd8C5wESY2Oe30UQgjF+lFdqKtAUwilZhax1eVZyESMdrcNQ1xG5q0q4xNtzqYiu+KpCfbIJ6vs2utvteNOP89iMuQ0HbP7UCOgJu3GnHhOp86dHWA66g7grg1IVbT6TCebClM/+cLGHmL4N2HbWxZvq7eU2YGXTLZfndkDOUVkHJkgcRhbiOWAQMYX/Mjp9P2nhgBFoRVZvuVmZ1jnjk2sVYtMW7ZcLgBCewmNE6yLEJSFEIP6REAonnmmfTRxPoLp2xmTnWHQjbmi/iCBsh3CNlvy9RLbNusta9YgxVJgpZxwtNmixctcLRRH66V+Pn5tMPODVDmsEFjZP+3gwCmEyOQEXKGQlTMjO96GjRxjLVq3s9vvv9f2Q0v4iBHDrGVKIryHMuyxaWCBNanIsbpxXs49WhPct2vVyHYUManTmZilIAaY+AFo8teTLf8utCgg41+FRGNAWnFsb0upk/RHZBFGWFsHKdr26uBlB21atrHE3n1jis2gKWeff4n17T8E3k0aOrnSDdapNNtcIDfoRjmelyCPtjUBqyMAbjKLCMFpLvwiDoYK0FH+bP5qu3zcaHx3lOFDo8ru+ssCu+7EDtalfYpz1xlFH+uQSGIirvND+TkKJyLv//RHLZ6VsgCTjJOsyixLxtzNyjWbrUeHNLb/UlPUaT9IA+Sv9ocjC7GZO3eSNve7eKJbZvs17iOxBsnwO/wYgKAgct8FsSg4sQGWODILqSNOeHuiHX+UYQuUcpGWaNWoq23aKO37ILg5osWSCu6Msn2MH/76k0eA6lIhMgUdtJx22sl4UNtqLXJLcUBTaR07d7M//u739uX0GXbbow9Yd+IdcmhPa4MWR5J6B9UrbMcCrPBa+CXD8G+Vk4VFlU3WsXEWlAo4hdWOUvgp7IoIgvc/0F+qwUYDik3trrDNGNZsXoqLSCpayNsUDEMkYTuvCIfhMzbssLc++tSWoIl13Onn2ykD+iGUjBkqdHFljVd5iOQDz5FXcJonU++FhUXO+KR4OmVsnzU5/aq+UytJo/VDYeWa5fbB63+1By8/D3lg+fVdY/f/9X1b9O7F9B/9y6SWjxFmB7HFzP+JB5BabDnGD2hrCb6mO7VJRaAeSQgQmcYNMAt1w87wJJZLdCWaJC0zXAflbYE652QYVETf7Rx3X/ZgJWyKKtglsSBr7Qo2b1tt11x/Lz5bjmN48V3DXMlISQ+IByoiGNE/TRPxgwPXtD9cfevqi588AhTyc0fv7FlT0ALp07ePzf3qG2fSPb8InVjs/cVh4eXAg4dbl/32s89nzrR730Tbd+0sO6Z3M9uvfVdrjO28BE6KK2LZNjKQGQzwJoSLsUkC2IEAWHoFBH7DXFdn78v34RQCkIl3Ovn5iEMzBKvC1DUjLgfd5yLLQyB71uxv7JkFK1x1evbva/edcg6OznM5YWQyoqYlKtEjNYeTBLxChGBCuc3UCX86FKI//BDy24UKVEImckKIIi5lsVFolBqHrUGz4SdNtVuvOsza5+Jyky22toQ1a4eQ4L9/crgK75M/ooDQ9KGjkUJFNrLYmjVE0HyVLPewEOA3WI6yoh1lFVkBDhyQXJAEq0KJM4sTwN4P2WOyMCkqrqoC3jAqnp9+Jj/Na2zIfq1x+IRBCw7SCgo2WstUVJUIDj6BkwCOAorWw5iL8G/685NHgL5fAwTBlhC5tK+XrrADS0rR0UU0BdJDhxnboJDSsxrZoSNH24CBg2z511/atA8/tJv+/i/rRiY9+ze39k3bWOusNMtim7wkOoP06RZdtNH5oHWuNQHn4FBk34KicheKCA+AlhN90VTQ1jMJSiERc1FtW7W1D8WI5oR1xqeTbTLqwtl9utulV19jXTp0RpZsuiVnpdsWaXuArGJBcBLu9hSfShKiC4A1KHHd+vXODJOe9E0h6F93GzxTwXIWjEImMw5CcR5faeccZpaZ3dAeeVGHIcvtqENHYi15K9trmP/CuoSgXUIMkS10n38if2gbfSzXo/IgnJJBn0P5FeQXWElOnKPSJScoarhKpDFxa4LYGxiGQY4NFWJbv24DB0osWi6K+gwY5BKsO+HpanLY051SCbmFMg2ii5IIq4dM/subIiIBGMkospvveNIevOUwa9pAgty4e2KH8dX8Jda8p3SdgBP4Jw5GyDxwQxvUze8cwuHHvwsK3rd/q73C7a1igomizvv3B9VFY+brJPm4tm3a20ezH7XThPj4p2HQIaWcVFcUFsogDPJTidarz0HWuXsfO/L4TTZ//jz74PPP7LlXp1gr4qc1j7fkjDa2lq1LKrJK0TJ0QEZyLVmBSI2jhhwAOXB0d4JnISY3xQVMClA9vqfcF62QvPbvgtT6GwJIbrS9dYKyvHPnBMqGd1FVbEvj2OpC8ZVAbRUhIrFwW75NXfqNTZuzxKbNXmSHHHG0XbNfb2sH4hM7IBrEn5SagyWZAsvGuksJq3mIYAtRdIEIi2SrHLUHciwpLbG5s76wHE6Bfaju37AJolkoPlFCOYp8OmUHGT/2NvOl0VR7+Ml37drzDsKySSr28RCfQdG/EpnAQFlMrf0vCPD24hhv6GoMZbAYswC3bZGGeTUO2KQTB2KMQs0yEhGJgqooz0MFMtMuPr+PTfzwMzv2+MORZEgGMQaLUZBGfVgDSbvrUY2fX8iCFMwbMUCYI8E5jKh/4Iz89d1JA7ANf+RPz2ABia242jHrJTvp3iu4xz8yqQsLiq2gvJddevGF6C73sL6Y8ceWkKMbJa6iOaKf7hU80gvm7M4w4OZTOGy5FMEfxf8+YScKsFZAJndfWV+Qr6x/Dr/6PHwaX0F/ra0xkflFPteWv96Fx/Pl+nJq0gghEZcJqWHt2LGLtU4v4yBkm6VlZLktsAZX5xxV8LQ0xFGcEhewFZR0fePmLa1ps+Y2ZPBgGLobbNGSRfbp9C/ssxn85i20jlD4/fr1sUaZqegDY0IrAWsyCPLKTp3Kk0Cy+B0oMjng4Za6SGgaKoctohs+d5igemqV5LtLp7/cg3ScADWNkEYLh7YgW/g9HCxUaQWGpycioaIqxdaj1fLN5jybsWi5zViy2nXBgcMPsWuuO9GatWyLI6lMJ1xaJmoDoVeZok9GqLawYCsIsBn1onz4SfQG9aAnHC9KQs68p4Nkar9oRwEnkA3YMgd26Xw/q98j+16I2akTIu7RqW1ne+nlV+y22+90SUYc3MISxGusyoOPqv7g52Yff1yn+Jx/ulcdLMkcRQziL7JClpaCtXEHH6KAOdRQf+wywRkXrGLHon3UokFju37iQruPLXRyAppOEssCRrTYOmAL67rwuRL22t3uPG4gJsZecCZDIcVYICosYsxxKhbLAl/Gu5j4cnv2iddt7erz7es5Zi+8wQnwpIusQQKiPCWcUEMVFuIkfhve8Dq1aW2/vfV6e+Kp17BZic4TBjUqgTE1S8jPSxHo3iPhyPrpeU/199/9VWnULj3v3L6d86pGgIroE/urMlEIzyD8Pvha89en83H07LF9eON8PKXUvUwlKfh07qGWPz6d4vn7yPz9e59cz/r5vHMwYpqM2Efe1i2WDgJUkACvt+zsxC00OPyEkNxhABRQbHyyQyLNWraxgQcMs/Ub1tnGdets+fLl9qy3h0VeUlPfr30jzB11QmIfcQBO81IxmpoYiwgOCEUev8BZ3Gv7KEil3Xp00B76AILkgJAonKAKmeo7yILNLYgOTWN4jjuQy9sO1bAO9aI5X31iXwT4jpiZNvLgvnZg8/ZQBqdaZiZHcspDVpTRWqnQdiksxMMGKMNUmMbH95XozGrT7JoIolT5noIWzPRP59qxRx9hWQ2ynFtB37c1aWsyd+9og8YoDUR71Lgj7NOpn9jsFdOsV7uGjpKJjhMCF9KV3F/QHzU5/JfchZoNmNYrVHHyqm1mz14tzK5/FLNsa61bZg7vQodfLFjqcx887Pvn8Gv4uDnRLwBP6piVSRVY69lmzz/+hMUUzLY2vU4CrkZCrcbY/MVL7eLLzrKNs66xh55cYn/+faW1hZebvx1H8CBmgMcZ/ejYIt1uvOtoO+Gip7EkVIzGCnKmwLkstEvlTzsm8X7VbiFChfD66Hl3dffxdVUQDGu3ohCej+4jg38X63Vwd1eQIgsRhGcamWH4s/Ly+er97vLWdzVeLvAUfMXcQx1/hDAzOa30cXeXv77JP6rc4HmgOOaYY2316lU4P+ogWschv+qOE2YKBcXXT+XIQxoMQ+TVZMMvnlPktpjJ72D79epvh4wabfkYWV2xYgXGQvNs4fy59pcJk302JqmAdrkNOB3DQAL8w2z82malpwJMknyPcZSYxkhIRuXJ1L62M/qhnIEOcqVtzt+IKt8StrdlNm9HPG56A60Ww+TBqLHH2AWjG1mT1tggbNLYWYmZDpWakp4BWoGClP4UxzVCZOoP32+qYCz6ybIcLflGkQ6i2IRyRfW5gwnIS7kb1HSa+cWn9tADdzsPcjkNcxwC9P1T1xgIbqQCpRPjVatWoVVyl93728MtB22GEviC5E45KjHAAr5udeVHxP/qoH4RlV6KheiWeDhUWLFqnQ3sP7h6PPRO/ai+Dyc89L6uoHzT0mAqEjQf4+AxTv7kS0vf9ldU2k61/gNPsF4fL7dOHZvZw48+bn/87RDLwRLR5GnL7M6reyBcK0F5ygVuIBFtxoJCO/SARtauKR7xkPauQrA9BYO8ybG4BdAWn3jaIQmu3MLH8IePudRdNWf1zsOEq1zYH70X/KmN0u2VK14fNzyvsCTVt4JJ9Q+4pJoIrP5Y240oJHVMOMD7wnx8X1nFqW++Pq1Uq3yl/bW2/FUHISuPsHz63V2VnzpUQR3Vs1dPu+DSn9vAgQc4Kkt5qSytoBoUorvgy9dVA6U9mjNeCnlYClKthKKTjFxicpolpDa0Fh0ABEy6Dx8x0vLhJ5YU7rBN69eCbFez1S6zrXn5GMDcZBPmr0VuYG5QSD3/ZrXpZF3aDMVsUpodig+Ptm3bWaMmDdmWYmyAgU+AOsNQMupsHPSwwmo7mY8dwAZskwAxxo02svoKvftxclfeOOqPb4Ba0A/EEQVMCksEYW/bstHe+2CS/fOFZ23ChHetV+9eTh+4PrqadJ0T84hF42Yq1J/CsB7NOPxgvHEczj5YswbFD0qjk1Un9bVQ4v9CLT1Af8reRin62mkYO7j+/J726bTpdsIxx1TrzmqMFTRX/M5GMOzHPfga/PXvdJWVcWngkNKNwd9eeM3GHzGCBbXSfnZ2F5v0/kdWUNjBHv/jH+w3n56P/nGBrS2ssMbI1JYgLhYvmScE6rduL7PJM9fZn67v6qQOFi6cG8AWh2KLEZbObdHMOeoSkPpFObxOutd8FU7wc9PXMzKeYFftlE/sb4MTlEZlVLvF9JM9sgD/rAoIsemqn4K/+jg+D48kw+P4bz6uvyoPffN5+/fhaf27yHh69nXYU/6Kpw4SA1d8wDQUyHeApBLhi8k1oYRv5VnNCRBDKTm+F88uHe32p3GOR6gKQWmLcNFErVJHwsQvxKpvlQ5FOIjIzkyxqAaNrFmztta7H4BFXaWKV4kx0FJ0kosQwZGFYCElHSw4z3KYvBfwSnBY/aGg01FRYInJKc5Vp05sYnX6wRZIohLinckPSUkpFLr6kvziWdSaQqFtZZvfMKcpK67yAtHxE2L0fe62PKzEcjIvr2vOdJcoP/oo2hlHqLC5c2ba7Q//2c7u2dRRso0by4RTEHyf+/z8e3/Ve02oeJyqF8NHmvzRDOsIkdGyIRZnoEqjZFIGOUW1gT8uKE+H/JT2f2GXHhB8qIOicREak1BuIwf2suFn/8OuvfIya9i0CWMt82eCdf2t/5wNkAcLILuA2NhkWwVr593Jb9gdpx8AdVdpvXv1s7Enn+bq89e7TrZmaLCsX70NrR4WycRkV6lKZELj2TovXrnDBrVpbq2aZtiUmcvswkt+zuFHrN394B/tF5dfaX957FE752wQqGAAAw5yciX49HDk4IY3Hif4Z391lVAKYEXBzxXd652P57/rfXjQd/2UDggMgk/kn2u71pVheFzl48lu3fsK7S5/Hy88n93dh9ejPvn7+KqXtsMtW7W0ccePsc8+/dgOG3OU03jQcRcoynVKJdu+YOUJaiHKUHwNBTct9axXmrjuk1jWPHMa7FAY6ct0MhEKEhZWUDvFWI5PzbBEfIcIQP374EZ5Bn3m6ywkpfL1XA5yqmI7LFpWtFyQWpoT3OlgRfHAyuUMbAw2/ErxE6JBlvweRbtVXe3wfaZvZMm2XsYQOImFSk1kWyOnRBu3rrO3Jrxpk157za45fJD1bt/Cpk78Av3Tba6qrt3KNBTqGl+JP8QgJ7Zy9QJ7/ImH7M7xY9AyiTO5wSjERwbGxZjQtDl0pF1XPr6c//arTumjODGRRlIRfLVuXXQa/7XNmb/AhoMAmdYBDAlW9C8EO+q32vrWv/MwAQZ0XTx91gI7uPVSy84YB+unmEULWCeceUJvO/JgcborcBi2yQ7rkQ07CpuWRVs55IDAQDh67uJ0O/ZwTv9jK23Vig04dB9tTz//siXlPWKLp15r7QddYIceerjlNm/uFn6V7U6WydXXR2WF113PdYW64oXnVVtaV25tH77vuz0V/H3z/7bpfX2ClQLkwKHAaWedZf944zWb+tF7Fs22WIxZmaAP1Lsk9iEEBnarZwhQQSi+EF0tP2WlThfz11FtiNFIC6AYinUtl38AAEAASURBVND94EkUFbF15l0pVKF+ItMVvxKSXUE4Rz+J1Pgy3IcQoDvETTskylPIFrwMqtO9I5Jqp7QeYNz2AUAmKUgwwbkLKCnZYR9+8o794pILrMmKOXbv6SOtc+NkqyzYZhtJXyohNOXj/gZ5hW5rvcQgA1jG9mj+wsXu+7BBzcX1o6+RgXOZsPCILP1fqFcPOHuR8VD/UTussjgRzZ0Su/2qPogWPQ/1J10fnbCyoCCFEIy4xtyPVu1FCB4UdNVOqIzt9SsYChnY/wBLSuEbW9vKHYK/NLvhgsH4vAGx4hpi9spSGz2gA2JMheyG4rABGGMbt8VaH0QjOraFb15calsK0u3qS8+1ZfMesvNPPMKyEYdSKAUugyCE/f3Cntq3u9wDsmZ3Mb7DN9+h3yHpPk2iQwatihUMTLfOPeyhex60Bx641x575jHbsHG1pWGdJI5tsiPNRVGFqL69USmPdJSXR1xCyLX96vpe33qIYnSnu4gwyByRg/8QkLutNuU6YFeGUBHpGelQEyU27dPP7MJrfm1P/OGPdsORQ2zo/jiB4hS7EG2DkrgkEBdHLqGT62qktSfoJb4Wk1f/+Y61RtMrF+KhAorYUTIw/cSLEVX6v1DfHgjYM1WVKaA6DhIg2EYP62P/eOpPtmDh105urxJq3n1QJ3+HsHHDBnvy0ftsYM/ODBbOtBifFWtj7NDRHaxFwxTmRzFeEuOtXeN069otHkQnn8aMM9uguJQKa9dsO/KfKajHxdolNyyyX5w1ym67fDTL3haEvQuDGnmkq6fvVs3v0LJdk1RvgXf99NN645AyE199LSsolWWF1qplrmtk2poZds2Vr9npZ51vgwYORiOkIbqyQh6yShwanz2sonvqre+ySvmFxF93l4eqJ5gS71I8RAnTSrxHp69eDU3POsQR0tEBhsxfaQJN+mCiLVu/0WZPm2xnD+1h/dp0dciuFB5NGbzNGPwmR5NunfoitEWqhloRD3UAsLh50ZB5Wzdssw8+mmpHn7Q/7kjT2S4hCyYeFYLS0aIIdPrxv1CvHtAYxzJolSzUcmJfCP+3Y4cGdtYJWfbE08/bXbd2h6kgkSkxYzQ4ewoBe0TUolsUAaSZ0+U0wqxTqxRHqWm7PW/lKhZDqHn+ySUDbqVt7KFNzBjLiqh03smiEmdauIOoLIOn/P/sXQeAVcXV/t7u297pvTcLKhZQEVDsEsVektiwxW6MLRpjrNFETCSW/AZjjSUqttgpBmwoqGChKSAgnV2W7e3t/33z9izD873dt7AYwB14O/feOXNm7ty55545c0oy49CQAL/y4B4YMaItUrkaqeVG2NqKMOcnTlVJu8NOoK4b28J3zCFs4p+fzMwLL3/D72qARCFEzqhHj244cH8Kd/caiFtPOhBzJjyECy+7hgRhEko35NNTcbr7STWET5VTKqyZ7wgRH5hbuTHXi+6IlB6i++qGl6dNfBYxwY0jjAagdtWs2hXHqvsUgUsJ0l6Uy6AqeVcRr0AipU26FG5IpKem0PtKEd3TM8Tl3x/Dqy9PwB4Jq3DP6YdgaC/uLHMJpF1GSkyJiy8al8nE7nQcK+sJYLg3DRFlvVRpVLZeNGcRli/5FkN68oNDRwzak5EzWb2ieiIav5YU3wjI1NLpJCh+CgmG5L/ptAK67KzTqV70J3w661PSE3L4tOrxVzCaF3pWvnxez0e4wvqo5O9JWPXhfP+9Wfj5KTS3S+e8ZgS7BBKplSvW4PD9GMM4lZt31emOCCYoSiA3EwMMZ+BeET7GBMYtCVDAW8EPcU56Io4+oCOtXLhByBUVv3kMp6BnDlqQhD0zOfmvJvD/gPipHz8JDlAP3r5ueuHCLx25Iyp1KoZHadEGdGqVjnNPOByHr83H5Ffvwfj7Azj7tNPRb4/dkNexM5JSaffLN7e2Wt6gyetLaE9CKnogvSxNJac7wswtteu4Imu3YUKhRxF/8ieyJrab1GxP7Woeqc0guTvpS4QoR0yjInaAO9MyoZIC94zZX+C5f/4f+P3GBSN3Q7cRh6OtXN9UllLuSAVVInEyRk4PTexqRmoTR0GjJ95l+MbURmMp/HJVIX+dpIfAkEHdHLHWC6cdd21nSqT4k/kKu1HY8j8a+fDwk8Pmvwpa9uzWKxl/umY4Lvv1XXjntce4s5/BeUHtBo6zflrNSBasn+aLzRtxYnqWNTVFfOatGexqFf541+/w5LgT+bGU+lgtoxom4atF5Tj5MPoL5wZMiGIRPTuFM6ihuZ7WVVLM1iMVrvDz5Lwkw1DB5XICiaQsXsgW0ifkB7jksuuQlysjBM5TbsDFM5e2fNSiY9gqBLA5X/bo3d6Sq3p44mlkpkUvtiRqRdxZzWaczGp+Nftyidb70CNxIFVIPvzwcTzyNGMI77kvRg4/AD179kBmTjvaz2Y6PTiptTj+hQ9dD9EMvrXU1LnGYWs8XE1gN7F5J9ZODZdDUoHRBBdnqgBM8twic7bS8mIsXPAtJr4/CZ+9/x76st5vjtgbfTrkcbFE7o7xkau4lNFUdvVI1cP95j2JwBNenorzmduzdXbNvEc363n9B4llgiks3YBJUya74uycIqoMZfLbIU8iYWLNR6D3oCVt5gg43VVyaDVUjD7rhH1x39//hGuvb48/3/VnijioME9VLxE7iTx87k+rBGkkuBUCVVFSgrTp5Afvs1kzXU8G79TZBVtKomnngqWFeGdVKp7sRg9AFB05Vk4rWa44SPbc43M6nN5ztMdaQ7aPrg6JvxIFlUm4+u4ZuP/+c8KOSGi6p3fFTbDNvP8trdbsBFAvjr2UW9q5rVVfXJtMsEimaIFC9jyHmwYUSYTI/ZRTBaSKBKF3bjp6H3QYjqYcbB49bky5726MI+O3D6/tNmBnyg97oH3bDlToJKFJDcszqkmAKrVjK6rBpEmmyccBcYRCk2VLHnaYKAlVmLBKnid8CimZHEymvh1lf9QzrCVbVUAF7C8+m4F5c7/Ah5/PQOGy5Th6YDc6RBiKXq3TqUtIRVNysRQhhekPuUVHzCXJdt3VFK5riyXiNbQJIoVql/S51y9G0rug0ip63Rn/yJu489ej0DpTkflkrsQyAjj1Io6NEdUYqFouRxmBTcaM1KeyKgmtWuVj4kuXoN/I+7hETcd1v/01I6tRz5XzUD/NH+UyUnDvKedrCrUFJEtct7oEL73yGv1lnonLxwxB904UgNDYKJWu4OYvWYMbj6D1UhsGX6IvSNmJV/M9aSgchFZE+tBpjkiMkkrz0yXfcW4x7bILN1eYtER3H0I3U2LPJQfcwB97LxoAiVnkCKAQbDKgMcAFY3DRGrUy5dHKI9Fau5arvKG6DZVF4tZ5JF5XXw+GA89CLsMSsUuffijLn0nuiWY0IiBi1QlTQquKIP3+ZfE3pEtrDDrjCKwuKsHcxYsw/cG38TCh+u2yO7r36EFD/97oyGVyBk2JcsjaB8SFaQwII7UanhA/MZNg6afJoXJtQOjMlSuvW0bYVLCcaJik3EpDdFEPl8IclnBJpWbD+jW0/CjA98uWYsYXszHzvXdpJBd2xCIjwyuPORgD29NahNr7WtBWVJHb45H0xqT375xFqCvEt4lNruYGv9LlVMeh/Qr1BMUpMPG6e8a8Dx3rfsKXNx7rPH9tAf8uQft2e9FCJZ1CewajoqV9eFhUZ9O7VJ2W1PAIuHHnuNUNOZ+YlJfpWXpDIvq2T8aiqZfizFvGovfBH+OJ267AvnR0m01bbCfW4AdZc1B1JcqZ88UcBqqfg2/mvkYC9axr+ISDezo9w5Igl7ChZEyeVojzTtmDe1VFXPLSuRnnSSXZOrfi4LGm7g+TZhbbYQB4apdyTqXQjvw7gnXFwIFyOyy5MucdJ4LkmPauWi587h3x5pWuRSYfRmXhsQnXjYS1cpuryoOyt7NK0SrYNdn2ipvxO2hlfq5y/QyvGmkMv17qeGx7rW3pxhUUUPGyrj9++/6x2tbDlpnMRjtBTRzZ9tYiJzuRhv1H4Kkz/oV9+/ZBsXzicVEY5LI4xN3LWm58VPHrWqV7oOyvUwZtgHfpgRG798X3a9fhlpemo0PZ97jvtZfDzWa2wpA990L7PIZ4pFJqtx69+IWlRQj7GSSHpn5oU0BES//UPxczVbU5Zo4WaEK4A04RUUgmqZFo00AWJMWlnFDi8igEX/rdYspsluI7ym1mvv8RIUm1mY7cqQ2uOHww+rTJQzldz9/w75nIqC4md0jnDpThuMi/nHTaFNE3mHfnJrFIUYgfhSTCVSaTU2OIRm7d0SokREsO7daRoFKmWMx+yIGCvp4CUarL3LPWfSnJgmX+osXuuGdPeptJKBYLSRgJzoVP96cfP0gtKe4RsPG1CpovtdTDEyEpo0VQ19ZJeOlP1+LNSfNxzYknIGXIcTjuiMGMO9IGWXnZWLOiiB/KZVhE2/W0mg9x9OCuOOX83TDusVGY138uBvZv5zYx5Ez3yyXrseuAtvTqQhlyyQZyllmoZuS+AOefc9evTtjDtw5ZzqVWDWNSJ9PLzYr1tTjnphdx7W+vcfO+oMDsfEWMNz5/e2cVtnLjO9s4QZNdbzw0QV0THZH1kzjhoAhEPEkds19D8HYDhrcx4qdyEbJ4cVvbht/Oo+XCbXDCr2S5dqT04Lp06YwPeXg6rRVSK8gBkuBpI1esuTtw1USsyAVpGUFvF2QKkUMzruG79cG5Q3bCqfuVoJiE87v1JTT4X4gF3PV8Zc2m49q6Zz/069wFHdu1p1cVuULnv1SauCVxYUlCq/44QsRqqlnJjYISEjqKSZwuXwXleHIrP/cT7vJ5aWced+6VgHMP25k6Wp2QybmUqbWq7pf6ewkUWPdto1jGqsSx1l+WiQOTBCd81WUs5TXJ/8hk1uoPU1kyg8ETcsmy+bSe+Tm93PRjPJRi58rKcbcOSvC6o/D4asz1kwxqPe2Rlbp0ouv0Gq6p2GQtN1UEHH4k4T44oJY/mz0Cmp/8bLnnQDVXegKvxWmj+2HkwZcy2NJiuq16B7PfWIxVRQvRa8/2GDFgKC4cymVtp0ORwyBGs+csxc33vIZXHh3NgO2JjHtNdRbO6eoN7fCzg/ms+AGl0Rs/kLThpjeieFSXwoSZEQfpDOGLL8Lz4KjDDkcyDQ8q1EkmThv91R+XGnpnDcbP7X3WNc05nQtHQ0nlgtOPm3GNV/CRWWX/mn9sjTcFr2CV/Lo+Tjv221adeNow3IZDuWtHFIDvuLygFPLaslL6PaOSZptW9Jknv2qUVZlLqPp+ES5Zg0diQkUTbp7QcoNCYe61ohWpYpt2adirXW9U7tGT/tAoX6T7qepEEgE6h1y8cjkKV3+K+V+tx0quCjeGilGPYicRrPb8dc6lDW2rPAwiwe3Wjs4PSFySadGSS9O1ZHK02tCRTbK4gIpyRYioQarkdRS7tEsPopgcbyL1uGrJIehliUV2RIQlG0xUvAd+5YO0LU1OaYUpUxfi0Gt+zmBSrbGO4Q+1dHEfCptseh51t6Ex13Jd1i1fzPrMXc3NlvqEXpwqfvAEKU5Qz8IIYV3llmwzRoDPgptdTrTDySlxSwV/VdyAakPP5e379cEQfikrjh9Et27ceEilRnr5BpTz+SZS0bm4IITr75iDU4/cC4cM6oFq2qkH6LqqmgrVffpQ9pu4kisgqq0EKC+kikuA4hMnJuG8ayhJ4pdKe+X1G4K49PcTMOqIo+lNZmeKbMg91r2/yo1wGa7Id9Z/7w3Gcr8ssp7BNJQ3+yZIQ41tadnm3KARL79tpz9VdyGPfu2UvuImQS1jU6RlZKIVN0gobeMOmVj4MNekt1thCiv4AqdwSVhKIllYyF00LpcDSXRmoB0t6kCJD0okccghe53LnTdxkgnUedqnXV+af/XRN9pJP6pIYEWDQ+R+i0rK6GWDQYgYtS4kV0GEyMpMQxaXyyJs2oJIIjGT5r+0b6Q8Kq65UhHpSMyqSNyquAso2FA1vTqzLInEroYTLZmEsnvbdliyOh+D6JJr0zEUIQpPYhEzHYbPKMMkTr5GyKYwfVlZIsR3/vnQQ1BBPT7un9CRgQhYeFnMIkcMRVY1mZW0u1eQvx7vffwVbr/qYN4LVSbIhsqNunM0QVqoJjftj6va8mczRkAyXH2UlDRPk0QIa7hcJRdXwXkdoqqMPIavWFmO559fgP0PysWe3bNIEDMw4aPleO39mfj6jfPooq2Mqw8iCVGCnKAd32LOhTTOU4pr9Gy5MqlhW2HxkwD1FDdN6oXmgsQ0gaRsvPnmYixYUYw/3/8LOlbNo+OOcpaF54lqNjYHGivftPWmnW1XBLBptxYbWiRFhIoOvtC9ew+MGnUSskrmcnnaDYVFa8KOUvlyh0jo5MWiirCUsJCAcbeYMjFNrkouEQJcKvDp0Q1QMjExkViSZDpnBGEFZOuDCAMtKkiYxP0kciJp4MOcWC3y6BsvkJ1FmYrqk2gSpyZRLYmhmyf6SnL5XcF6CZSX1pC7pDWtu4sAVXFcNDXH7JFrpW5jLcsDlOM5e1DOsxRO2CLqc4VVJoTZJi3xsiXVD5IgS2XBCbZ5M3IMW51KLg5peOyfL+Gkk07DEAZQclyzq08crGeY1GGpAZlcNpwzGPrUL/GbE/d3HmrK6DNOO44ulkT4DnWX/LWkLRkBzWV96DXmkl/oOERRjuO0aylro9gmPSGdcutqHH3BVAzaPw0PjPkG703+GZYvX49TL34WLz5wEnbqTQ/dJZJRV/BjvJ67JNT8JBEMcc6H+OEKUl1GIR/4zeU1bWLUP33XffVDl4JcIlfwIy3Hw9+tqMbPr34exx1/Iobud6Czb2f36uaJEVDh+d/MA97KTyvpIYlzdw+L45+ZlUH/gP3w3rtfoIhqMHncqFDIy1IqaMruVUtJOrhyO2jaDKjmg9XXMImbGoXcFSZI3TVxi3qg3F0lfk5DHulMP+44c3K6x0wCQDJFrMLN5SZhFZ6zlOo2JeS6yvmr4LHM0ERwpa/o3GZpgpPgcAaSYLGSkPFc7vVJdpzsUmoHUlR1u32C4YRNZJ+CiZVYLvtb/kskYdvIFbNfdec1dAYhzjUk3UHiVGSylGAeXpjyKvJ79ca9946lg9UMVNEoXjFrSZ35FjDTbFZXXHfCHKC4QL2M+fnaAV6OjtwZr+FyqoZLdsW6FYfrKrM/LWnLR0CESEpd4eGUHIxznCcKXQmGrdQqozqQiWv/NhXnnTAAj9x0JA46ohXu/ddXGH7tRIy/82QceWAnVJRw9cHnKsuP2hrK+1g/SGVnpyFAObVWQyx0HY4kfm66sU2Fb9CHXM+/lPPxoRcWOPjrrr0GKRl0hsqluTbAwvMmjKuu41s+EJuB4SdHAF1QZo57gNyOuJFEfs5OPeUUzOLgJeVQcB9Kw6L1VZjx7VKsrqZOIImBXDVJv04PmRUdB5aWkoa16wqcnptkhZobxqqHX2t7uLGfiiD0E8Fyv7pz4XG46lBsgklEsA7OHdT9MZhw2+GLmmSO0PMey2mD6dxi+ZUIrHurJk6SVRJkElPWSeYuXxWX44+/+DLeWdsN77/5FtV8OrrlbRJVhcTJuglsjRKna4t1TU1H50u506jUqbs2QEicxVWH31J3veXP1hoBzh964EmoYcB5frTTKCf+8JMlePr5FThhdF969lmLVH4Y7/jz+3jh0hE447iOjG4oIipiRz9lm5EkQgnW8Dnz+W6gjCSdy+2JH63HH+97A38aezv23HOQC88qbzPetNmMlpq3yg5PAI0o1Q9b3eiLMDiCRiI4cOBA/PPxf+G2p97F5xvoaZfbqAn0jbd+fTHjHyzD6lJG7aIXZrnLUh3302OsIwTua8gXfptL7l7DxHRFXdwPUtb6bmpstFTS8j6Ny1NZhWRT/SCfL8NFD7+D9sNPxPdf/Rd9+nJjh7qGtqwVcdNxPaGuw6hzKdoql4xn/XptL9HuM5ncJ/VlJBeUmV1L2tojwA825dEJVGrPJDEqKQngmrGf4OE790Ob1lUoLkrCjFkVmPbSz3HoQa1RVUSiF6QLNr0TVH3ZPI6MXtLp2CKBS6JW/HjOX1aF4855DPsPHYgLzrmQxJirCop0FDBM/Om2knZ4Aug4lYjR1vBL7lVLLrBaXAmXnb847Tj8Zdxf8cjb75Gw0f8dOZ02ivCWlo1Cqq99u2YDOSN6RXELWD5ELQG5Q6oX3d5pvfjbWtKXWULx6mK6NVJfvQ6K2+N85I+bOFSFCKW3xcRZC3Dt05Nx2+9ucJ57O3XtgXLq/ong2dLWuD2d2/javYeJpD4GAQrcV+PwkTmMg8IdREoTKxkCk5IhlmyDHwtvXLbfQ45r3dCKz6oKrKdNeBomv7+azkvLcMzQnhR5UByyDvjg63WgdhT9/HFOUFOhinqE0hKgk0H37H44BuFnqus2hzZ9imHVp3Ru3K0rS8BVf53uUNz7l4cpU8+jqIfLcSpRKxhSY8nmVGNwzVG+SVxgm8TRENukV1msDqq+frYT6ONpCLcPZ8fR2jAcVma56liZ1bc88rrq2AOsp1y8lkzvKZf+6kLuUmXh3DPPwTkH7IrynDZYv3IellUUoS/Nxxbm16AD5RjpFCrLizKquJXBuvqmaQmr+3aqM7z2v066RylPyxaEG9Wc+FQ+1rF2B7k7JxlcEu8hyB3vioTWmEeHpfdM+RrDDj+S3nDGY/iwA0j0glQYJxx1Fd1Gi2p5RG8jsWMBU62k4/wwKE5JGW2L35g0HScOOYAeROgAlmpGoaT27JNUdbi5Qw6lJTXXCHDe8UsXktMBcnx8xPyo0fyMoTJXFyTh2ItexUsPHcOAWZTDVmZixtdLKPfOo2uyTHKKNISgQ4NER5xkDskU8SHXG+OWx7TqIJWkbqpENmQOyO1V811IqNMZTZWohQaT9z87F6++Nh2TJ7/OeMD7uA2TEEO5KgxokO7vTXsg8t200Yi87t5Z1vXf90hYK7M8fBv1b7qBb5ILdhMOUBdsguvYfu7FZgfUsYZ+greGDZe16OOyYz+3enbz0doRvMFZXZ3bsfoZrf9+PYff3QcfqR4iZRLSBdSXKUCCMOaMMZj23jQ8smgdfv/4qy62Rhp3aCVDkwJnBsc0gctmSgcZZL0biUeq4wrDO6wacPax4XFXl7d6Cis7c9nJf5Xcgd6/Y7YLup1IwXYmudv09DxuugQwfc4cXPjQK/g42AMvTXgZ/3nmGRw44kA+57C9qAK+O265buzVcY2hjbmNrUQKTsao2+e4FhUXY9rX+ejcln0gfCX1yGZ/k8+NFXIC4cdYNwY62eRC3fWWLN4RcO8K57Hj6Plhk79AcfUJVIG658kZuOxXQ3Dofu0p+0uhc9sQ/vr0PJx0WAd6AOfmCNWnEsiV11IU4ignVws88JoWHaBskKsHqnZy0UPVrERukJBTrEyiag0/rPL6giBj0jA2yN+eXcpQmP/B4488hoMOOtLVSyDOJOkQ8sMbkq05k82fuN9Z1mmIJvg4/eNo9EDXlIQv6Aut3dUofwRolaIU11+qfxn4sog70E91laysHtg7EJxkR0oG7xXXH6oPxnWo34I1+Fj4rQ82EAZfj7T+gPcoBWj+G7zPYCz4cBqeef4/uOHKKxzE3ccfQvlJOvXtqFZAHakkOn7syVgIhbS26EJXUpVOtsFJwrEVF6jdW38a1TfzIx04XUea82lnejnN9nLbtyWhz6J3lvVYuGYVnpw0CRvYlxNPOxMvXX0iDhg2lMb0eTT9k04ht0M4vhpTjZdUcHRsY6kyHVuysU8i4VNYRHf/Wnsvr0KbruQ8+QKtL0jA3G8ZvnGnXI6zdMw0CY3w/S9Hyu5i+80lx5XyczI/brUkUPqQ19Jv35+fmI3y4lTceUlfzleqtpAjnPlVAWbOnI37frcrnwtlO87FhZ6Dnqf/TGw8SCTIISqKX3VVK+qjctZQQyGNZo5l1BEt5HuQy7lQFczFfU8twDW3TsCdDH160skn8J2Wvh9lkeQSXWyYZn5nNe9izUubk3YXfq65K3qgX4B/3CyMTRjCxEu2vaqgJOSR8P41leXk5DgY/7rfCf9YvsoUF9i9bHVtxMIvfAptmZGR4fohPJGwPm4dy05QP7vxH8ATZ4hqIBqKZL64qfSll5JOA26OzLxvFuCVCbRh/N3N2J/2wIcMYWSEDiOR0yUTL3+6EN2ykrFnJ9rcUsDrOEqSvWpORnYqshs/8rmeEdVpQpl45LOF+Pyr+XXtJ+OXp5yAw392OPbae0907daDk4CEn78K6g/WkFjJ9ZU9CxE7PXeNnT5SGnfZaeo5KPlj6ZRbhYsv4OyvvsTuuw7EgqnnoU+nZLw1uQipedUYMbANyhk9b1MCWNe1lmzzRoCPIpHupmoqsinXppoSrS9mzZLe33IcObIdsirIgZPYVXNz79d3fI7PP1mDiU+PIt9XRGtJyrWdaks04hfujj7n5PWciCMlQNcaCWlYujYLqTzu2I7u1mpy8Kd/fI6b730bT/3zIRxx9GiK+yjvpTPVsOMLvlckmC76YJR5E+2mFbfXXHk5QhXlfdIctHkp3KIL0eZlNPyiN/rQO1O4aAD+NU1yIbYGVWbXDE7n+tnLouuC17VYycr9erohSyq35OM3nJYLxnAZfOQ1w2V47NzBs4+JkmvwpJovv9y1y3GA7B779+2Lq6++gh5yf45pM6bhk9kLsJCBz6fe/2/6OAFO2KsHEvp0RiIDopNVcg5Wedc/4P7C1zbej2u32f6E8aoNPhmOBTlR5lLVqSJHIOJ3042/xwHDh6J3n95o376Tc8wg6LLyEu4SMnA51zda2vM7wN1AYtFLVUf8/HG2Y8vdLRBWumL8nvJACjVJjIW82BXlptHLDrmEt2eswHnHdQkzGXyhwgfqgUZ9a42L68KO/4fDWF2dyWUogxORs04qz8Tu/Yqx/16taI5YRDKXyqDnSZj5RRUeeux9PHvfKcigS5eKMmo71HPiehZKykUMw+fuyTB0QYA7vHK4INlfQUUWJkz6Buef1Asb6OH5mntexf89MQevTngOww450inh13A1xUnlcDmPL3V0YJN5w9LI5L/H9o6qjn52bnX8a4bXcoOJzH38KquPCxwJGHkuxD7yyM4IPvKa4CMbjMRr9UT4DH8kHqsTicvgIvtm8Jb75dYnK9uYh8mHlnDa0EgUJZA5kJbmlK906dEFp/U4DccfG0JJ4XoU3/1nfDXnaxx7JOUcGxbjkIOOp2vycn71yhgkWmNBAsJ5pAkkxWl5NZCbLM0rTS8BiMiE71kStPCU07Rzy2f2QWZvUrQWj66vtDgwTUD1z3FbdZNU3ngZXIOOJ8mFspLi7VZSneGb/FLc9vzLOP7UX+LKa66izIc2oExqSwqrUkoWsUrQkok45ek3xE4laknFf/bltTGz3OFQ/1nHJWa6F/W31sV4lRusYgzrRm6ZS7HvlhMbNz46t8+ikrfs6FRLf/RTb1rSloxA+D2g3h/nqZt4dFGfxsMKqr/UUhk5lQ9mXXkG/vLQRGr+t8HwIW2QQNGNNiYSatP4FCiX086JS9wk46ohIbiO8y7PfRQDiUWEpWFANXVFuUl4zz+n47D9dsFy6ssOGP5X7DP0SHzy8T8wcJdBNCbgJoybV3q6WgKHRVV2f/47a9di5Ta//HkXDdbgVBZJI6LB2zXVo1gz/mSdb6iGddbvVCx4g7E6seAir8fTD9Ux/DpWnYbqhUkQX2IOins/WSeBk0KKm1Zf7ypNYJHSuhVa8af4wl99swi33f4nXPTQg/jFQGAAvWy0Tk+jO3HxYCQqbFdyOJIY5mHrDl13rDs5Rk0T9TPcJokcjzUVw33ndcIofGSIchfnRp64RAQdUeTmjeRtCfT+rL4FOdmKKdReuLIY97w2yfV7/COP48QTjyfxyyBXywlPTk9tSfePiPjT2JA4E6euC4+NhY2fjZvlDrFANyGCqk0ht/pHslpUVIYRx/RGSnY6Pn9/ETrmJSKL7sdKCmnzTILbkppvBMLPiSsWqnPxs8tnKcdoXHaSwJVydzeNYSzfmrAAz731GV5+7DS05XOooCPgRIbX1ETSs7dnG96pLWPwLHos4kaHYtd/W9QebVMLkZgZwLPvrMAdT9M6KrgcN989Hffe/yhOPukQZKTl0E1aOVtnYgfUkwCJr5tQ+su5Ek+yOSdY1YmnXrxwwhmJv0kEUAjiSWpEnfIbi6fetgfjSILrlrsXLfNInbQp4LgjEp8+vXvgL2NvwdHHHoU33pyIGx+818EP5d99D98HXHnQky5NwZIYTpDWJGmUjQSJp5pR5xKoM2Ut2INOUhm5UM0gOTcQcdNyPJEyFF0UQU3hrJT7+nLK7KrJrZXmr0VByQq8+tZSzHWtA3/52ziMOupo9O3Vw12RD0HFX+AUcP/rwJqURT7PyHNSY/fMKxnX+L0Pp2P0frtrbYbXpyzB6cf2c9x0WJ0mvpehSZ1rAXYjIOKnFKDuUyWJm3Qwp81djV9c9xIuOP0ghjoV98eVioJmUSWplqsWcYlh7878GuqLqHleQQKam4Rvvg5gDj0ZHXNYKzz+2nKceemz2HdQH+R0/QWDsf8BPbv1oLMFzkWakSbxfbB57DrhyGD4aGv+/cE8bEJjbhMkHgSFhYVuCRYPrGCasgmipZ3wx4NbA9zUTRAJVEtKShzX1djYCL8EqhLa6jiyT7pWf53HihNXVFJM/2vpdAJbhFVrVuDLBXPw0YyP8eGUaZj+0Qw2Gd76V9uH9QqgW9cBaE0b5Ex+frRhINdRLhFfKn0FSu9OBLayotR5YtZXWW1qPlXQImMFgwx9+cF3+CRcy/099JDDceDwYRg8eB8MotlR67ZteJ3cJJW1HbfouEQKjd0LYmSX8j7ik0DYl716aDc5VJ/8TZAfjA1vk+IilBYUMbRoNj5+/mx07paNzoOfwJrpp9J1F+2lHW3f2P4mDbSc1I2APhD+GOlY1yxFlrOU80MrA1WTzE2y7Eyug7/8pgYDR01gxSAWTjsB3VtzXpVxmUuRCc1znMxQivKaXq4NilO0Akkid7ihOIjf/WMhrh3TB+++uxBnX/8yprzzJnYbtBta5XZEaXEhv2/08SjXWWzXBW0Xmoikvm3OJkhT3lnNS9sEiZyXEd1xp/WbINEKW67FHgFHiKyYhElyviROmg1lBUhi/IRuPbqiB7nCww8+GAXn/QpFJIrfLfmOgdfX0OVVMab8dxpe+ewrfi3pSIEEZcW3S4iNk6ix1KYT1Ri4nKGbrd3olPT4uy7FBW1aO0/a4kJzc+nckstyWaso3q/jUjmppczsiB4nYXhhwpdkqyUupYk7xEhg9GCIvt1b4bX/zsPvr94XrRhItoru+GtpnO/8aW3ygm+1Dm13iPWUOKtITEik5DSAc0wkKUBhsP65T6k4Nm5euOfpaCEnIUU1El2EuFqQPC8zIwefL16P0y96h2PAmDYvjEHPtiEUl1BGxzmSXEsvL3wWkgI7N1qcJGEiqg+8lOTT8fTL3+CbNQHcMu4/eOSZNZg3Zzb6DaCch6mSz7KkiisTwsnBbaL0Rvlve0t1rMf21u3/XX8jOSU5VSCbSCJDP3ecezX0mpHA5YD2JVq3aYs29MXXu08fxz3rC3sad5MVoaua8jh9dSu4FC5cT5MlEjZxY4sWLeLObLnjtDp1Cu/WVhImkyEJlLQzKzUU/dQX/fTFqyYXXUll7Vo2nKylNa87KaQoUn3a5KT+anMd6OWldBL5jKl8+H6rqItWjRffXIWLz9yV7yflnVyy8+Y5CnTK2mIJ8oNhF6evuZTI6IQJdEhLzT6eVpG7ouu1hFzniDQ1g0tW2tVWkbOvorWHbGxlfxGgGVuAmx1pqfTawtjPUz5eh5G/fMS18cajZ2LErlnOOidIJwW18hnJlYc2z6T/nCQOsIYhURMy6NCgEhk52Xhx0ipcesd/XP3zzrsUS5deTe/pXTnP5IZNH1apz7BdHodliGF5tquwHf1pIYDN8LDsy8dVB2cD9eZ4oIlRrR3PumRsuZYCEg/4qUuXLvWnA3fbrf7YDqSDack4UOUSHUiHUik8EcNcgtoKw7kOWdWtnotj4euEDz78FD13GYpZS9Lx0sTFuPe2oY77UxhGbQTZWGz1Dm1HDehJJVEVK0Rb3SqaJ9LgjN9L2akzCFEoG9VplZi1NIiVi6owbEg18rITuLtLR6UkXAkkXCF+eEurUhnCsgqvTv2KHp4nE2MKJv3rlzhw7wwue6lnG+JH1G2qUbOPPvsC1B0M0Zyzgjt7gdps7vLS+WluFl7572ocf8ETOHXMGJxNP5AjDzrAxZnWfNMM026zAqjrWElzbXt9pk0igPHeZLxwbvS24h//wViflIeJww8bjlbm4/hhjfAVmwjhszDx07G1acda8urn1FnqBMbWF2vbzlVH3J3B61zHxvEZnN8Gb0xgcSX/vgyH9SEaApUZnJX7OJz8qE6VYvGSJejZ+zC8MrGcZlid0DqPVgQltnst5dPwznoYj16i+Pttbe9ouUbB+UsMUgmfnF2N4u+StwskktOjP8cgXbMN6FaFhYs34NgrllOhPJXBirLJraVQ8T4Rq1cX4YPPVuCx58JOCM4+eW9cMWYf7NaDFhsMpBUi8UsixxaSF0raDMs6R9OFEmgXACmBnp+DyZl464MCjD7ncbzz9iQMHTaMHCU5dsoKq6nGlcjVjowFtET2n33kvNiSZxMNb2PzUnX85OPwr0c7jksRWhWFVC9gPDfrdyAeeMMfrYOR1/ybbQy3X27Ew68fDbfgLPn17Vq0XDgbwqs6Vq5ceJWLmEUma9/KBGcEU9d0bj/V9fsY63pkG5H1mmdsyBnULWvTuYT69N2JePKLFZh2697OsWyR9MjYf8UZkd0pR6CuW9sbAfxhf8M+JnU//En2IRBLusxzl7k/+lzamkFAdfhUFqQPxdUkLsEQenXIZYiGUoZgqKB3ZcbjpXwlk9zaaQe2xeBdsvHC28tw6p2fAauWsGJ4FUDPi7jqwgNwxAH9CZNODQQSt+JShnvlMjmRziicrC6Ty2fKhRNou0sZID2fgVeo95qBtz/ZgCPOeIQbHlMwYsSBRMsIhFwSgzYjSdTF0jyRCEZJc1LnNpf9eegAvD+al5YaghOMX96UeWn1LLf2GsrVr6DMQeKpJBmV3bzg/ZuKbETltjQTXGP4daNaGjaWrF3Db31oDL/wCr/gDEdDbQmv+h9P34VH42J9aAy/yjV5IvtubVl9tW84zfTM6qg9jZmfrJ6IpPVd5YbDh/WPhVO7Z0qGwy/3j639aPj5eaQcNODiAH866z0898L7OObwjtipG5dnitZJ+VStXhgeikZsTBtfjo3XttEj9ZtLzQCJuXN+obuhuKOaXJpCqZLScBecCkp8NBpLaQhIq4QPm0rqfGaUg8rLN0t5WctOHVJsIEEcYSWzzWL8mGffWU21poU4YWQHdG+Xg9xUtkMusCJFct5k9O4SxDVn9cc15+yMj78pw5AjHsC4O3+G04/oS5uPasqHGcqgcgOVkoP0WqSPjkI2lCI1OQ1zFwa4YcZ2cunfsTQdOZkbqCvYGS+8tgBn/OY5vDjhBYwYNpSy6TKKcmQWyVugHJLiaXdPmrtKTX1nmzov1Ua876zmpX42Lxub88Kt/gh/sJheO2xiqyBWUozOeBCrvvCtp2A/HryC0aaAAjfHm0S0hV+EoLE29KCk1tIU/Np+17gYx9VQv9S+Yhrbx6EhWCtT3yMJmJX5uXALr/DHmzRBhT+evqsPilWsmM/xJtlURxt7p59Iz8MrlizDk+OfIrqeOH5ET7TOJZdRwaCh8gQil0vazXRvvrW4/RBAPQ9airPjuhdyQyRyiSJ8tKelcl14eclNCefuifpAit9C1xKk/YotTc6OXFxyIgOUU7ZXyeWoNjNkjRGQs1hSmgoGsWqVUoZzj6Hs9F1Qd+9TDN89G0P37EyvOsno2Skbmal0X8bd13JGHPye7tlenCQ1qzTCtUNuCgPecyOsspSqT1RNSSWhrmEsZhp9UBE+D29MXUCi2BondZdPxhxadVRi3vetcdcD4/HIczWM/DcF++y9F/LXKwTmRlGOPSnLNQ5iiCJl2VYeLdc7K1W3eOel1Fqa8s7KbliqLfHQBPVffXdxgZtC1ATrJgHzWEnlPozViQWv69YH1fPP3Yn3x/Aq10BaPQ9kk8PI8njxC4nqRtbfBLl3Yngtj1XPLzf8ds1DV38YDaYh3CoTPqtXjyjGgT+G1o/G8AuVX89Qa1GbSAF7FT0Lh1MI+w/uQE5JSy0+LzrdlCsk53ps+6F5dnvMOa5UP9EGl7i8WsrEkuhyvjaxlCojOSQsGbS9Zrxk2l7XptH/oRTPaVWtPSpK3ajQTtWV8lysrSlmkPoyJKeH9da4u8DdXXHzDLRKTq2aHFsKZW2nHp2DXfvmYuBREzDukWl1/cjjUrUNOmdXYkn+Kl4Lj/Vj9/0CA7ozmHlpIeWF1ALgT4rQVVzmBrhszm6VQvf03+MpymQfuj4PxWTn5n9biddnfYDf3/Ypzjr7PCz//jbk5GbSdpgbIfxQxX7Dw12xedKUeRPPvDS8dTfs5rOOI6/75VYWbV4aXKy8SZsgsZA0dN0GqCGYeMvsRuOFV9uR7TcVR7xtGVxT8Ef2zXBYbuXCacdWFpk3pd3Iuna+ZTg41pQpffKRAmgCf7h6AHp0TOHSSZy0r1Tu1n3W5HaU64NIl+4kZtVVUj8hh0eHAoWV2Zg7N4gB/WqQX7QO5WVtuONdjswc6mySmGXQLDIjmdHwqBpVnVaDOQsq8dobi/HNomXo3q0bFYpTHPHKIvcmx7HL1yVh/tICrF2xGHPmaCNkOYbt0hkXn3ckdfsq8O3CBYzxTE2C9F3QpW17DNqjFdq3ZryairVcSTOWMwm0Fqny0ae40UmM6fz2tAIcfsa/SFR3w61jS/Dy9Kn4+mtg5N6nYeI7d2LEgfuTG0pjECvqptam0mJIIprmeTRNnVPR5nlTcTSl51udADalMy2w2/EIkEhXUp614Ks57iaOoSeERMm8KiVflKG+5FyRplI/5v1uGeEVQUigHLOGxEe6lqlpFMPQr+GN4z7Ag48tx8jhmZg8dTlvSPe7hr9u6NSpFkcNycWw/bthQJ/W6NUxhKEDOmDKJ8twWLe90a19Gt6bMQcffFrEAFJF+PK7KnRoXc140slo2yEDRx3eATdfsyd27tMO7bK49CYBxgH7kRmVIwPKEblZUktZX00JeU1a+gTozCJIW/BKBkTKoBywpKotHnzhC1z+u1cxdPAeePujVcjolYNfnfInHDZqOPr12xXZWbL51fKbuoeUZSZw6d5cxO/HfLqb21YLAdzckWup50ZAX2y39OBZRUUllq0qwpiTB6Jfd7riokI3zQ7CUeT4kkn+z/f2R09smSRYtq90CkrnnDXsC7cGuBljG0mxiCM3dii/c6WU51XT2UBAnlYYLjS/vBK/uX0KHp0wE3ddM5IxcBNw3bmDyNFxk4uuCCqrg/hu6QYsW1mNrxZX4I/PfUR/fcXolZSLt2Z8h8UfjUH39iEcuf8QuiDj0pryvxIqiofoZTm1lgrJCdyk4K5wgOFXqVmPotK1XCVTaE95Yw2XtwkMYF7NPInemR2/x49PKnXzkmhuCC6T531XQguOO/DUq63oAHWmi+pXVVXBkA85yGOMjnCSaaJ2bCSi4AYOWccEBr6XcoyW/D+F1EIAfwpPeSveo4TO2n1L5U7y7Flf4N/PPoqPXhpDDom6Y+szUZtcwheWL7KzahDBiUVstk4nuXCl7JFUl1xRoryTUA4ZUOwMme0whQmyvewedZaVBGVhjuuiVUY1vRsnkANL4TJzA9VJzrnjEzp7aI+x11/GeMncbaVunExx2Vj4FkntDxjQ2kUSFPN7Q9kuWLS8CCvXL8P9nfZDh/QQyvJpSRGU01p5TQ4iLYmEWu6kUjfQz14aVq+sRX4JTSwp6+tPGV9GMgkV96vkaTmxmn3j5kmt/PSxn2mM+1vCndq5jHj00pszcNPdk3D2WRdj1cob0a59e3ev9kcbHPqvz4BCnOpY3K0+ZM41vntGBr1j5y0EcMd+vlv97rSTrBdHS6i///NxXHbmIOzetx2qytZSlsTdSO6aJpCDCct2jNBs9W7VN+BapLOJYG0JiRO5JTqJrahIZ9/ofIP/9PZLj+8HrCm5Q9GzauYBej6WKkkiLTKqkirx279Nx69G7YzDj+iK8gJuiDCkqNxPubgy1JcT15jkdoDpwozG4jUkulnk7Ab2TMCuwW4kYPxoVNEtGAlXDQOQ13KzQkHsA4kZFBtw84L4kmkr3r1dFmq+A96dWYYJ0+ajYkM1urZLwd67ZSGZ8sLUpBwSWNDGvJxc3gx8MjcfT708F7vsOhgvPP8sjjrqWH6IkimHpZ4fCZyLyUvVFidT82m9+l6/7v3xn1H9w/ofHLQQwP/BoO8ITdoLY6o0n3w8HU8/Ph4fvfJLKjyXopREpYJOWd0qk7uSWqb9L5Je51ruxiaEMhFKZqziEm5gkCClZopZE/EjQNR1OQkZy2tJ3OU0tpabHrWZIYx/ain+/lIQP9snD/fd/zbmLaMpWjnN0CoWIDO7KzcdUilXy0aPbt3RszutYLKocJKRzKUxNxeoCpTApXEiOclk7YqncqdYnGggD6vzK1CwXi7m21D9gzvDVI8RN7nzLq2w88BcFK0PYd7idVi0shJPvh3CtC+L8fX0iey8zCSrsN8xJ9M6ZCTeuuBA7Lz7TsjOo1chLstDbDMo575yTBrrVt11jyLy/KeSfhIE0F7Wbe2h6ku8rfatsbFSv9V/LYFL6A5s3IOP48pz98CevVqhkmZcNdVpWFKQgV5tpGfK96+OsYhKaxprbIvKReTo+j2lChvoyn3x0ioM7M8QAJR3yTeDulWnh+wIhBaG4atUZCb1DnFpytB/SMkrx38+rMQlN7/sevOzc77E8SeegXbtyFVmVlKBOZ0qJOVsowzvf/AtPrhHLqg2pouO74nuPVujX9+u6MYYKdkZ6SwUp0fV5bRSTPysAhdc+iStMhio6thuaN92CXbqOQId23cnQU2gPmUN2nfIQm6nKjqZ7cS2lpAArnMNPPrEQ1Q6P5oywwwSPbq6J8dXVUYdQ4rypF+oVCsO8McffNf2tvynPi5wY4OjCW8vq+XRbkx4oin5xoPf8DWGX+XC58PFwq/rfplfx9qLlRusX78xWKsTDU54/LEx2MbwG5xw+seRbQiPyi238lj4I3FFnlt9Pzf8XKnR/T9VNxgn+cOpH2PCvx/FwsmXcRlXTPlaG8yctx5tabGQxCVgNWVw7Fpdqj+wC82Yc5eW1hjJXE6CEcuqSGDEecpLSjnH5YMPaugrEUilHl8pObGgc95AEHKrXLRy7KicrJ1Wyiup7MJ74U3S8iM1KxGfzQ/hmNMfxUP/9yD23GtvdOnchYSJLuPJXWmTRUrPFXQ0WlpaggouYcuomLs2P58xWaqxdt06fDJzFj5csBjvLiT2FNagHK+8tAzZiWXkEHORQR2866+/DkeNOpqK6TlYuzYfE+lP8pnX56OwiOopNIub/t+Z7Fsh1VZ2R5s2abjpppswatQoDNipv+O0q8kxsgvOekNcoe4q/Owdmed9uv/EsWkSjFl4+CWx5o3B+PPFP7Zyy4XH5o0PFwu/DyMckeeG188NxvJYuCPr1BNAq6g8WmVdi3bdR2jHkgnpRReusGA1TDytDR+PrsVq0/BZ7tfXAxNu4TIcfrnVsX7o3G/Xyv3c6isXfnE3Sv71SBxWFnndx+sf29hY/xvCL9zCGw9uvx/CrTqNjf3mjI3Il/BrJzI9JYHOWZfhqNGX4K83j0LPDmQ56Er99TkFqNlQgQPbptMKRCZ9kqb9GIm7owxkpV1eBArc7moNTcfSuHR9azp3WZO/R+d2uSgr1vhop1N3wzHm3xrGTk50sQfo5ombHlLdqaKCsnZ3vy8IYM+jHsCkt1/EyEOPZVn0VFOjoEFBLoUZhCgnFx1JJO1ZHHrwweSMKQvle6GldwkJpKxwZI6VlEyizTyFbsz8NPKgYSSU1c49mixw1q5d6zac8vJaubkpE0mlSo6xQhHoPpSczxae8PWomzucQ+GimH9tXmoeNTZvDGZbmpfuvjfjnQ3Q7Eufh0aTEYOGAO2FVW4vVzyDZC9rQ7hVZviFW794cKueHmhjsIZb8FsTf1PHRv1p6tjbS6e6jaXNHXup26YnZePvjz2Ii8+/CCs/vJQKuSHM/74a4yesxh/O7YVkGgFXkctq7OVrrI9xl3MmJ3LZSm02txEhKwiGwsW8JRV4+PU1uOX8VjSe0CKTnJ6sJdwCuA67lrsknEmBFEd0aimDCzIqfAU3Ps647q/oOuAa3PXHazgHqS4isV3dTdm8Um7zxvprc0rXNQeVNsJT2YQmoPo4qFywem4Go3ObtzoWbA1326W2IniDs7pGtFzHCK92mjpvDJf10TUS449g7J5igLjL/hgIfzy4VTFe/Na2cDcFv40tvVhztOJIsv+0F8tuKrKqOq1rQi7b4XiT8Aq/1Vc9/9jwCK9g9eWTfW+8SV9Pc68dq+/CZfhlg6gvcrxJ/vrkzFT1Y+HX/Vhqil21HqrGRsnGO9rY6Jpg1e+m2FDKIau5xI/Vd7VtPuAyGVkuNTXMqcxdMBc79dsJb/3rdBy2fyusWluGS2+ajzsv3gm9egDrKeBPpd+5rZ80tiIK/NCRMw1RwCdHoYkppVRWTsYlv1+Amy/aA/07l5Fb4q4s5WIilNrnFf8nQijbXpm6BTi/Eli/gkvftMz2uH/8TFxx13wsWzgVKYztEpBXACYbb+V67lIF0pwUV6fnoGtK/pj6z03HsjfX3NSxkuroWHNcxEu5zSkdm825j9vKDYdDVPdHtrfql8r8fvgwVk/5tvTOSp7q2+PbPfh917HuX/co297NeWfrN0Hs5YpsQOdqXD81pjxWMhx6eEaNDd7KIuuq3B6OPdhIGDsXnGD0M3zx4BeM1TNc0XLD7/dJcNZWtDqCjQe/4Whq39Wm8CupbqxkfdfYW1tWz84j61p5Y2MT/kRSpcMJ1MPcRznlXOdffQOuvHAPHLJPO5RzyXveTbNx2cn90Wsn7k0WUvWYumny/sw7iGy6Gc81f0gonEt2clDOjEs7t0z0fnzX/YtwymEd0L9HBYrLSO7oGiqRxE3EUl9+jY2GN9xDHnOMK7mczMzojNcmLiXxm4gP/zsVOR3aorKcxvxUVQmHJNVcDNfSOBq3FcYX/hipC0qR428ETmWqp/E3oqlcRvqqI67P3iPDr7aMMOrYyiPbUJl+wq08VlI9/awPVk/wkTgNh+FrDLfhEJx+hs/q27nhtdzKrZ5dj5YLh42Njq1uLNzCIRj9VK8+LrBVjNaIf60hxIbcOiKc/rGPxz+2thvDbXUMp3I7NhwGEy3fHPzCGw9utRcPfoMRzqb0PV78/gsRD367N+tX5Lixl9wNpQUCiUxKFYkDZWOSPL328puY9vIEPPPJFSQlSbhp7IfYa5c8jNy/C8rWr3AWCgHK1RTTYmslifpEupK4jCTNYk/pcTqRuwC0SQ5S/+3JN7+h8nMNjj2IBLqslMcEd/SblheEVnXrXoKUpFXOTZC0tEzMnr8OR1/4OB5++GHstteebmMjQDM35+mFFTkrWDf8sbGxs/G23MaWWDdJInBKesEt2XzwCYWepSV7rlbH8lhtWD3l1j//mn9sbSuPPPbhoh03htvqWP8F39j4WB3l8eI3nJbHMy7CX88B6qSxZMjjgWsM5scuj7fvm9Mv4Y43NQU2XpxbCtfQ2IhMJNFNu2Rf67jj2TElA4u+XYATTz1DNAP2AAAm2UlEQVQW0184C51yk3H13VOxsCwfT543DFWlKxjCk3uQFUnkZMhpubHZOkRQBE1u+KtCGWSlViIYaocAubyUXO74flWA0y/7DKtmnMwlEj3SUNfFiB1r/WDIuI1AilSLdLqHX7O2Cnv87O+45LIrMPqYY+h+iq7j6zgzR5TUbhQcP0AaxwWNvT8n/ONY1eN9uQ13vPCx2ot1Xfi3Fm61+WPg3/gJinWXLdcbHYGtOQkabfxHAJCycGKgnFxdFUrWrcJZV16FcbccgsF7dMRt97yJye/X4J/XH4qUitXkBunSqSqJ7tfTaXUmlZKtQ/x023Kz5TgxqpIkyL8d5XMprWrw7apyDD3+bXz2+lFolyFVFHpL2Uj9oo6YlJSTaQ+7oSIDv75tAvr32wvXXf0bboRwt5j3L65NsqZ4CFTUBloubpMj0CQOcJu8g5ZObeUR0AYBrQhoiJ+dXI0/3P4gWpW/gjGn/A5/fOBj3PhMAZa9fACyacdaXpbOYO+EJWdQHaL5WBJlWSRSWy3JTTstNBIDVLymb77kjFqs3pCEPsNfx9P3DMLAvkmoKuTGBr0pa9eaErG6rqhPEkHoVFyMuA3qBdJG+Mm3vsBTb3yPBV+8i86du2Bd/joqE4ftnSUzchxgHZaWbPsfAZsR2/+dtNxBs43AplyOIohVIMhl5cuvvYMnXrgFf739jxj/xH9x/diP8fXzR9FbMb0+VzPOLDeHq6hiIr95CZTFyU61eZPHTZJ4iX5pKyNAG9/UZLZHSnbj7z/DTb9uTd93vVFVTB26JPVHXplJCB0XSJ6RAsMayg0TqKSdxBCSVYzGlkq3UK9N34BLrn0Lb058B3127UOLCsoTmbTcNeLXGCfpKrT82W5GoIUAbjeP6sfrqAniRS/kylxOPb9gMPfTfn4hxv3xETz0wnJccds0fPbGmdipSwLKSujwk+6mEiUno81pOFGhmFebLwmXputGnG75S9ldjVxc0aTsbw8vwPSClTTJG4FahoGsIWGUg9DEQBrrUcpHTi5E9RjtGKcm0E6Edd77lPa4qemYMb8IR5/9D4y773YGtT8kzOlxAIzgmRC/+e6nBdO2MAItBHBbeArbSB/0spvgWcc11ONLJ/FbtnwNzjjmTIw5+0g88/xU3HHn3/DV25djj95lXGJK9YSbq04lxJeoiP1rbhZQ+Ny61dHBAPtXQS4uNT0JT7y+HNeOfQPP33EgMlBEvUxujJD7SwhRJkm3UVU0a6uljXKILvuD5ABDta1w59+/QUZWBVYXhLDPqH/goksvwDlnX0g4NsMNEddWXXM8aUk74Ai4GWuTvrH700thX8RNl0mb1owHRjWsXR+X6vrnPuaGynw4Ozb8OlfdxvoVid/6YfUMb6zc4KxeNDiDiVbW2LXI/vnwDZX5cHYca2yEx7gdCf6lTHvT727Euor5eOr1WnRvXYx5b15BRecalBenOWuJ2kAhuT9yWdxI2MigGeXYyLFZ25uXh3dLTf1DQZhkuZGVnoi331uHMVe/jo9fOxd9OlLuRwcptUmMbUGTthAVsZNpHheQLJKORlPIEZZSTnn17e9j2AH0tNK/C0YeeDNGn3oKbrn9VqSn5dHGWSZrddymuq+500Cnm2vsY82bSPwGp+sNJZXbT3BWr6E6TSmz9pXHwt1QWbS2hMfHa8fx4jc4qxfZRiT+zY4LbA35Dfg3G6sDkfA6F6xeOlt6RcPtw1l5Y2345b7w2ur7ffHx23V72ew8Vi588eC3NqxflsfCa/CG2+Cj9V9lgrMyg42F2y83/Aar8yAF/jK7euihf+DpZ55yRScc3Qb3/PYodMutQGlR+OWrDNLPHuNIyEFnTbCYx1oC66dyI4Ku+hb/ERlyfRVqEqhUuo1675N8HD7mKbz2yCnYuzvtfEvo5y+ZoTjpASbI/otnrOZSt4b9S0kvYIS0XJxx83Ts1SOA447sj99c+zhWtx6OV8b+la6r2tKZAi16ksJ+DNVhtSdS6I+rHatcycY+fBY+t+Noeayxj8RrdQ2/1Yt3Xqrv+lm9aPhVpuvCaXCWW/vRcsEYbsNruQ9vcHatMdx+ufBbioZbZYbf6jU2NgYnfK7/sgW2i9ZYtNyH8Y8jYa2jft4QvNX3YfxjK7fcx6trOm8M3h+UxmANp99eQ3UE55f7x4bD8qb23er5OP1jK7e8qfht7Px6TthPApPIGBPPPP4EzhpztkN/2dmjcP0FA9E+Ox+ljDsRlP906gbKrIysFleMCgEpMzGRC/0sbR4RTOBz5VtJIqznq60OvgxsR5xcLZe1aSR+H85chv1PfQb33XUMLjq6Px2JlqCau9E1NfTgwh2ZGgYFqqVpXGVKIUN7Uxma3bvgzpfRKTmAm244A7f/6QXcNTUX8964Hz169KMaDWWZUk6WhrX0Brm89h05NPfYb+68tOdmIxwr9/vrH0fC+89fZdsCfr+//nFjfY+3/4YzKJs7u2HLfSR2rSn2q6pTUFBQ31fDoVzJGtexqHBT7Vf9GKPC4eM33HZN+GWjaZ4zBN9Ykt2wbDT9Cao60fqva2aj2RheK49mV234rf92LoK0OXGBDY+Ng+GzPqhcY2N21XZvzqyMPvDefPPtOuKXhrt+OxznnzwAucEilFZqa4OqLvSmokgbJBUcGNEMHtdvgGwe0bO+uVw4tWwl0eMR/9E0jEvaWpqxJZMDffe9tTjozGdw+29GM45uX+7YrmNv6PxAhCuUxSDkUt6mH0AqSmcl5CB/QypjeLyGsqyDcd4lQ/Dbm87F4uCZ+GbSdejYqiOKCzeQ5onwadnIjPQ2OSWJrqnkty++JLtq2a/qmWk8/bEXBhtjG3vFY7ag9PG0IJtttaF5GYnb8PvPXe9s5BxuqB29s9bHhvCrTOKRzY0LHNlX65O1qXmpuMCyq443KS6wfjb20er5+DU2gg3aYKqCDZ5/bNessuXRGtA1levnD7zOlQyXO/HO7XosOINXuWANv/Xd6lsueB3757oWL37BCrfh17lSJD67ZngtjwYnWL9cuAVn1wyXcksqbwzGYA2Pch+335dox3aPUvLVhFBsivfefQ9HHnk4UXfCY2NH4rQjFKSnBOsrk5CuWLgBclqJ2vkV+dOPY0WuSZ7lmydxA0ZEiFTI7a2QCCbxQhVlecmZiXjvw2ISv6fx6/MOxOVjeiGlsgDFtQx0TV+EVVX0UkOF5kK64Q+G8pGW3AoruFFz6djJeOHV2bjxpoNxxU2vYOdBN+GhK89neEmG7ixmHFx5TSYHG2CDHHZHcG28bGztPPIeVW5lNvaCsWsGb+eW2/V48aue8Nszs/qW+3h1bHgt98utjnKV62flfl0fzi+343hwG6zfd7sWiT/yelPxR9aPhd/w+tt2PmyzHFsjsZA1Vh5Zr6Gbi4TVufBHttFUHNHw+te2BH9kXR+vjq1cfbbjSBg79+9LsP65wfi54VMu43oZ3ut46n/fxYiDDkL7nh0x4e6DsP/ubRkQCPhuZTH9/5HIpK13Hl5qqU4S3ikNU73mI37hXtZy2Stvxs72lkS2mnp86XQv//6sIoz45VP41S8OwC0X7spNjTKwZ/T+In9+iQw4HsTi1euQm9YWGTkhfLE0EUeMGofl9I36y1OPx6p1Idx4/fXYfdBO3B1m/A2q8IRS6Vk5RBz1tr3+SDX/scbZxt+wN/a8DG5z88bw++WRfYtss7FyH97H61+PdRwNd1NxxMKt65H4tyoBbKgjLWX/ixEQJ27yuY0Cci1n1qxZjfHjx9Mr8Q04/6zBuOn8QejUMRtrCorwwsS16N+pI4buSm/P9LNeSY8r/O+WvE5cpjUw6aBbrcZxW4qrKy5LfJaryL/qGcmdu84VEHupTRh5dCanTK40OTMdkz9ajUPolfmCUwbjz78ehDQGJi+jzbEkhKlplVR1ycGTLy1C6zY5OGZkAj76rAT7Hf93nHTambjzD9ehS7eeyC8rZ9xdoIzerGtoQZKYmMWXQv3hDYW4dJVbLPWEebhXzFrSDjsCLQRwh320G792+oK6ryg3BGq5McDFLmVcDMOYxsDl1VVc8k7BsOPG0IHfYjx/3+k4emQuo44xGtn7+Xj2nXk4/7h+NCujyyYqRYekHyc8jkho8HTkMv2NI7EvVFxWnI6QdovFrfJHzRTq9JHclXGDhSZ3VVyO1tKtvaKj1SZn47m3F+Pnl/4bF587DHdcPpixMxgrl5saCYkbkB5sxXCTCbjrwYnYe0hPHDkyDy++vgzHX/wErvr1Vbjh9zc4OapkSynlpSihT0C5sHeen+k+38X+5W0Z8dNNuEhx4TuL455aQLbXEWghgNvrk2uk38bq18uL+IJXuNi4NcjU0pIyr7kLFuGp8Y9g7P3jcNKIQfjjTaPQe6d0rP6uAveM+xrJVHG5ncvMVrlZDMJDQsQNBv3byOqJb2ti4mYJw3dzecuwkIxKVJNE9ZRgCPnF6Sgpr0bbTG50UN5XnbQaWXQxXxnIxgPPfIYr//AOLj1zBG65bE9kJGzA+vJEpNLtTGpyO8z4Zh0uv+t7XHHucOy9VxquuPodPPD8pwzUNA7nUbE5NUU2yWGOl/SabddtIriui9PjLYVZUncc/hNe2nsXWg53wBFoIYA74EO1WxKXE6J8zyW+z2n065eQlMIQjKvx6ov/xqWX/8YV/fsfv8SJB9CVFJeBE15egRMum41/390bxx8yiPa81PnjslHLwjCN2AyiZx1iLs6qlsHFK+muPielQswkl6pcyianYXCvtfTtp1BGScgIpKOAAdXvGT8Xt9/3Dm64fB9cdc4eyKaydXlpImN1CGkygwYtxmlXPodfnXMJ1q1ti1673oh9Dh+Fjz/9BPvssTd3p4spw6RJHGMDa/HvPgw/IHbC1ZJ+iiPQQgC3k6cuDsW4usa7LO6FXBSXrNrZdT9ZdazKx/szpuCUP/4ZtZ99gZuvOALnntQXnbrmYvbc73Hb2E9oJZGJL94YjF27M4A4o5FV0342EOTuKL3BuOVuk/phPRXR1LKZiQxXRmADEtJzsJQbEs+/tR7Z7Wpw4rCAtjK4oUEOldoPS5Zl4aYHP8LjL0zDndccg0tP78s7qkB5FX32kfotZpzc3939Ev71Yh6uvvp2lBYvx9ICqsdMmYRBg/dCNvHXcmc7kKBARVxbM9n4NW0sXdWWPzvoCNRHhWuu+9vWJlfk0kYvQeS15rr35sbj9zOy39xHJEFQCnNkEt1LQ08KbLXkkqTOomhjQaqErF27Bl/MnuWWuu++NQljRu+PS2+7GLvvlIb8/CrcOW4mfnv3Z3h03IE4cWR7pJMTrCykIwRnN0sVkUrqwtG5qTYnjHniMDoVlXDr6gv1Cl0PCOTGOLzMDFeg/JDKydykRXJaFooZKvI/U77D2H8uxl+v3g1Dd8ulflsViV8Q2Yz3MuPrapzx2/9gzrw5+NsfT8A5R3flcrmYPgapCsN4JFPmF2PkceMxdMRIfLfkEXRo343mayWM4UH737ohqaETVATpnoYMcIKz6w2PVr1IIAzZLH/95xQvwsg6kc83XjzNBWcfh+bC15x4tubY1HOAkQ/EvwF1wH663tBgCU+08obw+201Bb/fTiz8fr+tHb+eXfNzK7dcZbHw+/V07NeJLPPLBWc4LY+EN1wS3ispD9cTRyMSGDbTkpJwDZeLQXo5SaT6iMIrijCtWbOGsXBn4rY/P4h5s9/FwUMG4INnL8XgfZLJMdXiX9wxPf3KF3DYQYPx5ZTjsTPj+NZSLaRchI4ywiBN3NQOqQ/bDRM/65PsPkQMq1XgyF8djydlaG5yVNNSJEATtFB1OV3okxAzcHgp+zh9dhEuvG0i9tk1F6/etz860M1WRWG5I4zaoHlu8jqcfNETxAluyJyF0Qdms23KC1MyUVBajX8+vxRX3zIBF5x3AX534w3o0rUru0gqJ/M3thnuDSWViey7Tpy8b2Pfrf+uAf6JNfZW7ueRdf0y4bFyy1UeC78PY3iiXbMy5VZueSzc0WCtjo/PjiP7bngtN7iG8qbgNzyx8EfDFe2a4VFu5ZbrWjz46wmgvWRC4Fe0c13zr6uBWElw9rLaF1d4rHM+nljXo+H26wu/cAuX4fDLrb7dl879dq3cz62+rqmelo5KPv5IHH6dyDJXOcofw+f33drxcQjOzhNI2GQP4e6ZcrQQne0FGNpRTF+Qy1RxVzpZv74EX876Gu9MnoRbxj3Hnd1ZuPyU/TH+xsuwz8B01FaWYuLk1TjiutnA2ll4atzPccTwzshLLUYF69bSBM6tGGUKUZdEan+YqKxMHRiR4FpxWDQbSySXpw0O9ppXa+iEoByJaSnYQHnfjM/X4a9PfYVXX52B5+8/E6NHZFJ5mt2rqERO60wsW1KJe5+ajbHjp2Kntn3x8D8OxJD+WUig/VpFagY+XlCAP9zzOSa++ynue/BenDr6VKRlJFG+x2BNdc9Jket+mEx2GX7+Nh/+1/OyKfPGh/Xnpd2LP08i79/mT+T1yHN/LvrH1raPR9fs3fCvR+LUuV9f/d3m3lmGc4w2uze5F92EzM/sZjYp9E5UbgMi8zY79kCiHmpghF/wsdrwcWsQTXE3njZk5SBl33jwqx3h1gP224za8bqL6ntjk9HHFe/YqL/qt/ov/NLXE5eTQO5Mci31sbKqDEUMm5m/uhCTP/gvJrw9CZNffNH17PZfH4rRh/fHgN7ZNBWrwAefLMPv//YhbWiX4LcXHYDzThqInvSeUl1J4lctbi+FOOkEtNEZQcKXQEsQ7gonsg49CLCf5BKT6KCAzyaB8XQVp2NVQQlli/m4m4RvyuRZuPGiI3D2Kf3QtRMVkIvLkcb+l9Ce+L+z1uGG2z/A5/O+xXnHjsQVl/TFzl3TUEW/VIVFKXhx4vc4/7fPunua+O4bOGjEoSTk9DpN5edqjoszmWOpghSFO7/pDdjYazzjDZ1oYx/vvFQbej76WXuuww388eelnq/ajEyGS7n6boQ7GmxkXYVqVT0lwxMJ41+Pd14Kh72zkfj8cx+3vbN+eUPHNu91n3YPkfCGX/nmvrMMeUCt0DiS7OyMiMQCV2c1MLrZpsTtFV6LTRsLt64Lv2BlIxhvv1VPE022vWECEn2i+fhln9kU/MJtcYENT6yHpnLZOcYzgQUrIifbYeGT3bacFBRTJWXBvIWU7ZVg8eLluPUP9xLyW4HjxvOG4tanzkGffplo1z4RGwrK8daH3+D2cVPxwYyVOO3IvXHvi0dgz/45JFzUiSuhMl0wjU5c+PKGShCgo4MQfefpVdyUjDj0dZORi+6aLPaNz5oEMzGFHyQGFK+h2soabpx8+00h3p25iERtOiutxZXnDMa4N87FgO6tECQHWsWQmskZQXy1tBKPPL2AXN/bDvmDfzgWvzimJ7LyyunVhYrMXxbjT4+8hzffmoWLLj4b117ze3Tr1oObG1z8U06ZSGXFksLw2IRflOixP2xe6vlr7JuSmjIv9ayaMm9EXN0zZb1YBFB99ee97iHeJJv2eN5ZIyRNfWdl99zYPLa+y+a8KWOjfv8Y72x9WMyGXlh7CCJsDd2wDaQmgj1Qg4+FX+UqUx2DjfWABef3wdoTfEP4hdfwK4+WrL6V27n1L1odXVO5j9/gdM1Pwqef8Pu4BWPnPvzGY/qvY0jHq666BlMnvolfnDyKDgxKsfybz9Gxw/fo2q0fnvnb7ujd4zAM6BlAZip3a6kft3rtekz4TyF+f9t0fPX9Yhw6vAPeeeo8HLRnGxIsBswu4uTlpkNyKpes1Ru4jObHhTFvq5O4pFSf9CHjLfDpuCWtiGECl9yJ3FyhcE+dJtGvwsoNARRVpOKbRdxombcGNzzyMYWP4UDud191GA45qBP6dsvhxgq5vpp8WpylYf2qLLwybQXOvX4Ssa7E8JE7465Lh2Lf3RnWkpshc+Yl4el3vsKtY99yw/DMM0/jZ8cchwwup7WzncC1M6ci5xg5UG+uyXY4PO4/HHtd19yxsbbnY+euIe+PylUm/AbrFW9yKLjNmZdCYvVs3m2CmCfWP5s3Orf+WFlkHZ0Lxu7X4KPBGQ5/HA3eyiLrWbn1PbLcPxcOwVlfVGb1G8NvY9/Y2Ai3JeEU/li4rX3BCG/958Q6ZYii5Q0hNXjh8Ymfdagh/CqLB7e14cPGg1/1BOfXM1yRucH4fWqo71bf6tl5tFx4NDZKhl95Q/hruGPapm0bXP3bG6jQm4tVC8fjrLNGY/czjkGHNlnERA5OLowpF6wsCeLrBWV4c+oC/OaPk1lWht33ycPjjG629+5VJKRFmP7F92hNhwJtaC4WosOARKqVBOQynpyf9AC5sCTx44eOAYdc4vI24GRrQW6chLBo2XqsXFPOpWkFFn63Gnf8YyGbXxWGRStccNquOPTQTtijfzo6tc6iHJCqOM7srBJrSoOYOmkl/vLYl/j4szmsk4CHx56AUw7tTs/Myfjmu1I898Z8XH/Xmw7ftddchQsuuBA9e/Xi/YUVmV2UtrrWNO9t3MO55pEKo/Gu4WWbjbXgGxt7lRv+uiYbzHzYePAbMr+eXYvMDcb61FjfVV91rF4kPv9cuJr6zhp+H09Dx4bf+tSc/TecfvvC31hSvXoC2BhwU8rjabwp+H7KsCI+FRQ/7DKgD+697y+Y/eUZmP7eh/jdrdfinF8OxcCd+6FNbg6+4ZL49U+K8MCjU9xw/eK4wfgl3UTtu2d7zPysCDsf9DqOPKwd3vhgIVAsDq0nf3QCALpPjpq0VKSz0wZTZ5x6RGcce8gwdGkfRLvOeejalvE2kkmAKNuTfK62Mhnz8wN4+6PvcOmtE0mT5W++FpedczQuoCeXjq2Ar7/cgNf+Ox833zfVtXbWmT/H+RdcjL322tvJvbRUNC7C7048L7cP33K8Y47AltCbrUIAWyZm8000ucQLUp+ttoKG+2Qe9997H4wYMgxZ2Vm48JKL2NBy/qi0h3z+2uJ2KjePHNYGO3PjI4ssPlIL0aULi0h0zh49EONvHcrdW8oV6f9uyfINKFjHnV/u3iosZIjeUQLBZBSW12DNyuVo37YVWmWno5LWIBQLsl4VYwMH0L1zO8oXU+mOPoV6e8kMmsR2EkvJ0LGD5BiLCmuxdE0Zvl60Hq9N+QiPPrdYHcA+Q/rik+kL3PG/Hn4Va9ll7ddQYw/IG4ixY/+M4fuNxB577kwVPmfq4WRYIn5bMsldg9voH70rO+q9baNDvkm3tgoB3KSFlpMtGgE5AHWrZgYaT+fSdUPhOnw6ezbenfpuHd5FuOj03XHUIUdirz3yqFcnB55k/6voSYV2ZqHqHPTvXYWv//MzDLpwIo7qm4SzT90Lew9sg/7duURN4fI3SD065+GZVC7EuB4kkDXlPd2GC63NmHhNy0rtskoIz5eWPuZ5mbYZNJNbTSZyfUUtvluUgKWr1mDap4vx6FPTVRGdeg7FAw9djmHDDkZGaitu6KzFypVLSHS581+ejDMuCKB9m7bIa5WHVm3a0QFppnNL73aV1SSXMr5XZoe05U/LCDTTCLQQwGYayOZCY9yzcnE+suZIYHSzNWSXpk+ktcZTL2HSS4/i+IN7UFF4NPrt0gFtM3KwIr8U8+asxcyiKpRRYThIlZSK1CDyuGvct2Mu+vTKxJIXR+ONyctw/vUfYeVaeewuxN1XDUX//p3p64+705mpyOCObl4Gc66Aa7hpUcnlZ00oh3F/U2iGVkECnO82IsrpqWUVVW++nrcCs+YU481pn3tDkILfXPlrjPoZnSv07ofcvFzq6smbTAI6tM5Gvz69SNm460z5ZSVVabjVQJUaqrRwA6QqTTvdpLqk4Qly2uA4pDAh9BpoOWwZgWYZgRYC2CzD2HxIbJNEO1RyVTX767n4z4tT8fcn/g+nDq3FxSR8/3fx5ejWMYj1pTV49PWFuOaWN8i1LQG6D8CY4bnoTJMwKSlnkaAtS67kcjSAYSQ0vbpW4Izj2mH0oUfi64VlmP1VAd75uBBX3f0sb4DLV6XW/SganI9fnnwMDjv0EOJJwMxZc3DfA0+yUETzh2nwvkNxzdVXYcCAAW7DYpdd6EGmVSu3yyY1jxKqS1Rrx4JETUYbzJgqFUKd/3iBzq/CFI9EXw4bpFxN54Jh7i8MrRotqWUEmnsEWghgc4/oFuATtyPCJx0oKTw/8MDfcfnll2HsHwZj6qND0bNdNhK5LJS8bV1JNXY7/RUM79IZ7z97MJe+KYzRkEIOrobupajbSVlcKIkKygQPcqlaWc2YuOW0z6CdMG09sO+uKRiyc2ecdnx33HddH5an4q3pxbjg2odx8FEjcMNtd6FDu06sXYvDRpXhisuuQlV1JXW5EpFG1ZkaqZzIAoXurFIpB0wnC5mcrOV3OOleRMyrnEOCjTK8jZtzIn6bJlZhcn82LWg5axmBrTQCW4UAtgh1N/9pSaFaGvlLly5xxO/cs47CWSftiVbcDSkp3YCSSrqD4q7BxwuKsHLeUpx4Fd3WDyGhojMAyf3KK6pRxLgYcgAQKGEQ8IQ0bjLQRphL0EqSwwDte9NIGBNkvcFuZtE/X3lCa8xdvA5vvvMwdtpvEP56x1j06NjN4dKzzMzhhkdeGmqoMiNlWbmWikwhyg4jlW7d5kUk4P/w3OaliHM8KV64eHDFgrE+xSrfkutbE/eW9GtbqvvDmRxn76JNDg24ruvX1MH360TDrW4Zfr+LTW1HdZuC328r1rHwxdN/1Vd/Y/VZ103pU3ato0cfi/GPvsTf6xjYnYZqFI3NCG+i1nfl5PMex5gT8nD28Sege/c2aJdbg+yg3CSwLe6kytuynCSEuLTUcSXlg9+vCKCARHTOgrWY9kkB7n+aS2guO//vvodx79EHIie3Db0mc4HKZavIZID7KTK9q+bGRTX1EuVQtIoET6bI0j9UXbmt59N3cjte2GSM4xmbWGMiXLGS1WkMvw8nXDpXncaSwUXWj6xncP51q+Nfi3bcWN9VJxr+aLgir0XeY+S5jzve/lob0eAbwt9QPStrKI+GW/CbOzbWVqCwkDoLcaR0evOIdtPRqqqzMp2LJwlWL73wx7rJSDwybZNsKZ7+CKdM28RVxYNfOIXbbEAj2448F071XfcQD37V19jEgtV19UHcE58N1uXnk+Mrd0b/IjXhew7DCHYDYV5/ezLuvedPrmsnHQYqPe9L1RRaddDdu4hSoDwJpRU1KKAC8xczP8LrXztQ/knDeb8ag8MOPhDDhw9Du3bt3dK5isvvIL0rmDxSbYbq+lXJsIwVleXsX939cvawWK1EWdLWOrNFs4+1VhvKZf4UyUk2BL8156X6LbPLWM8qsl9bc16qLc1LtRGeA5Gt//B8a46N5ntTwlaaaVs8fdd4633VexvP2Avn5ryzeseC6lg8jQhYv3iS8Bleda4h/CoTTLy41b5wayKoTkO4hVcvsQayKfiFMx786otghbsp+NV/9Sva2Ng14dVyU/FLNZEbSgcfMhKXXjQGy5evxnJGRJu/eBE2lFchhSZxshKpCdK6I5vL3Q7JGLrz8RidR9WY/v3QvkN72tZ2r8cfolxPDF1SUnhhYNyo2vaffA090XBPhfeukujfTxt75U0ZG42LxsfGQS3ESoJRH5XHk4TbJyCx5o7wCbapz1W445k3hl/2sU0Zm6bMS41HU/ov3E19Z/350dj429irT7HGXThsbEQAmzI2qhvP2AtO7asd9T+og6YkqxyrjsojYRprw8pVT8nOI9vw8QomFpxfz4fZGvjVluH1++f3wY4Nzvru983HY5NED1R19KIot+s6Nly6Ji8ovXv3d7+aEIMFlRRyuUsffHRQEKIeIYG5hNUGCFuRgS9TItevNuGN4IiH03QQ7mh9s2vWf0ISU+z5Y/BqLxpOXbfkl2/Eb6XRc8FZPcujQ4bbj4Tx+xdZz++D6inFgvfx+vUicfrnPq6tgV9tGV6/f34f7FjlPozfN4OJlVsbsepE4o0F5+P3YbYGfr+tzZYB+kh+rGN/YOJt0wbQ4DcHh9Vt7jxa39Q/fS2VRNz0U9LXStet/0YMVablqYiYIGtodxsoo4MDbheXM1C4dOwcVSP1S9RHg+oxNbT8qKGStJSNbfku3HwNoixk1cKPm/yXprGWbTyaC64xPM1VHu3ZNxfuzcHjj2Nk3zYHn9Xx8dq1pubNgSNWm9sVAYx1EzvSdSNyeuhGCHWsSekTPd2zTQyLcpZA3blqboBUc8mbSM/MSaFUGmtor5eEjXZ0tdxJ5iqXeMS1icuTHE+Ywsnw2XlL3jICO/oItBDAbewJGxEyLshyddOODcbvevga5RoBafmRFySRUwRg0kTV5H+3wI1YsdYpPxPCcFru4245bhmBHXUEwuurHfXutvP7EsGz5B/btQZzyfIE4ER00eV0InY+wfOPG8TdUtgyAjvICOzwBLDlpY5/pjaZyMaPugUyYgS2tXn5U332CTv6je8o9/djvDA/RhsRdGC7OW1obBoqi3WD29q83Jx7iHVvW3r9xxybhHhv3O+UjiPPY910NPx2Tbkd+/Vj4ffbNPho9f0ylfv1InEbbFNya9Nyv268+AUXrb7hUpmV+zh17Cc7t9zKdG717ZqfW5lfT8f+ueAjz30cDR3Hi78hHPGU+X32j1XX+h55XWXWPx1Hpmhl0XD4+OPBYTDCH9lGNPwG39Q8ErfqN4bf7kWw0errulK0vut6JH7DFy1vKv5I3Gpvc5O1bf36QWB0K4jWgJUpFyI7F2y0Y/+a4dM1u27Hdu7DRMMvOP+67ZhaPeU+LsPvl9uxD+df8wcoEr/VicytvnKV+X30y6zcrhmeaOcq08/UYKLBRF6zOrHa99szWMNhuQ/jX9N1H2+ssYmsY+fKI3FEltm54OLBH21s/D5amz7eyGuRZVauPujnp8hzH9ZvN7LvBme4DHcsfAZnuQ+n40j8fnm0tuya30fDbbnhsNyuW107V7l+UsvyYXUcid9gdd2SXyfaNZXr/iJTtHqC0XVr1+o21J5gDJdypwhtF4TQr6xzS7puDdmxlVnu47FrVkfnOv7/xsogBYIYhmH///Wgg0CYDLQQ4rq2Zm97nfpqd/PLJ8PUv3r1qstGl8OdLKP/161vh36PjHqry/FNHn3f3WbYffdb5NRkrl49tZuOpxw8Moz+1TEnw453twzvV275Ztww1HLcy8cnW5/7nnpqd7Pl4JNh6l+9etVlo+H4znbk+7a95tSbkVGfrP4rm/6VlSPf31F/e/vHZkdGdzn4ZvWXbdfcB9r9sU/jUIAbAAAAAElFTkSuQmCC
`

const Sha256_deploy_obc_objectbucket_io_objectbucketclaims_crd_yaml = "e0822d03a3670e12981ba36697b0e8746368e66f561005f66134a481429966fb"

const File_deploy_obc_objectbucket_io_objectbucketclaims_crd_yaml = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.objectbucket.io
spec:
  conversion:
    strategy: None
  group: objectbucket.io
  names:
    kind: ObjectBucketClaim
    listKind: ObjectBucketClaimList
    plural: objectbucketclaims
    shortNames:
    - obc
    - obcs
    singular: objectbucketclaim
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - description: StorageClass
      jsonPath: .spec.storageClassName
      name: Storage-Class
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
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
          spec:
            description: Specification of the desired behavior of the claim.
            properties:
              additionalConfig:
                additionalProperties:
                  type: string
                description: AdditionalConfig gives providers a location to set proprietary
                  config values (tenant, namespace, etc)
                type: object
              bucketName:
                description: BucketName (not recommended) the name of the bucket.
                  Caution! In-store bucket names may collide across namespaces.  If
                  you define the name yourself, try to make it as unique as possible.
                type: string
              objectBucketName:
                description: ObjectBucketName is the name of the object bucket resource. 
                  This is the authoritative determination for binding.
                type: string
              generateBucketName:
                description: GenerateBucketName (recommended) a prefix for a bucket
                  name to be followed by a hyphen and 5 random characters. Protects
                  against in-store name collisions.
                type: string
              storageClassName:
                description: StorageClass names the StorageClass object representing
                  the desired provisioner and parameters
                type: string
            required:
            - storageClassName
            type: object
          status:
            description: Most recently observed status of the claim.
            properties:
              phase:
                description: ObjectBucketClaimStatusPhase is set by the controller
                  to save the state of the provisioning process
                enum:
                - Pending
                - Bound
                - Released
                - Failed
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
`

const Sha256_deploy_obc_objectbucket_io_objectbuckets_crd_yaml = "a1da53a81af9a94b7cc6ac677d0f5bb181b8b34dad92338a94228722067b4361"

const File_deploy_obc_objectbucket_io_objectbuckets_crd_yaml = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.objectbucket.io
spec:
  conversion:
    strategy: None
  group: objectbucket.io
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    shortNames:
    - ob
    - obs
    singular: objectbucket
  scope: Cluster
  versions:
  - additionalPrinterColumns:
    - description: StorageClass
      jsonPath: .spec.storageClassName
      name: Storage-Class
      type: string
    - description: ClaimNamespace
      jsonPath: .spec.claimRef.namespace
      name: Claim-Namespace
      type: string
    - description: ClaimName
      jsonPath: .spec.claimRef.name
      name: Claim-Name
      type: string
    - description: ReclaimPolicy
      jsonPath: .spec.reclaimPolicy
      name: Reclaim-Policy
      type: string
    - description: Phase
      jsonPath: .status.phase
      name: Phase
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: Age
      type: date
    name: v1alpha1
    schema:
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
          spec:
            description: Specification of the desired behavior of the bucket.
            properties:
              additionalState:
                additionalProperties:
                  type: string
                description: additionalState gives providers a location to set proprietary
                  config values (tenant, namespace, etc)
                type: object
              claimRef:
                description: ObjectReference to ObjectBucketClaim
                type: object
              endpoint:
                description: Endpoint contains all connection relevant data that an
                  app may require for accessing the bucket
                properties:
                  additionalConfig:
                    additionalProperties:
                      type: string
                    description: AdditionalConfig gives providers a location to set
                      proprietary config values (tenant, namespace, etc)
                    type: object
                  bucketHost:
                    description: Bucket address hostname
                    type: string
                  bucketName:
                    description: Bucket name
                    type: string
                  bucketPort:
                    description: Bucket address port
                    type: integer
                  region:
                    description: Bucket region
                    type: string
                  subRegion:
                    description: Bucket sub-region
                    type: string
                type: object
              reclaimPolicy:
                description: Describes a policy for end-of-life maintenance of ObjectBucket.
                enum:
                - Delete
                - Retain
                - Recycle
                type: string
              storageClassName:
                description: StorageClass names the StorageClass object representing
                  the desired provisioner and parameters
                type: string
            required:
            - storageClassName
            type: object
          status:
            description: Most recently observed status of the bucket.
            properties:
              phase:
                description: ObjectBucketStatusPhase is set by the controller to save
                  the state of the provisioning process
                enum:
                - Bound
                - Released
                - Failed
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
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

const Sha256_deploy_obc_storage_class_yaml = "d84f84e94b6802c1ae96a9ed473d04ac1fb968f41d368c4cb7d438b75a8ceeb4"

const File_deploy_obc_storage_class_yaml = `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    description: Provides Object Bucket Claims (OBCs)
  name: noobaa.noobaa.io
provisioner: noobaa.noobaa.io/obc
reclaimPolicy: Delete
`

const Sha256_deploy_olm_catalog_source_yaml = "7e8580ab7a46eac1f975cc8b77010e065a7b9e516fba18eb1486d790de7aa6d5"

const File_deploy_olm_catalog_source_yaml = `apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: noobaa-operator-catalog
  namespace: default
spec:
  sourceType: grpc
  image: noobaa/noobaa-operator-catalog
`

const Sha256_deploy_olm_csv_config_yaml = "7902c00f83ed852ecb10b9ba2602e5c0271fc4f94afdc81dc757198942c63217"

const File_deploy_olm_csv_config_yaml = `role-paths:
- deploy/role.yaml
- deploy/cluster_role.yaml
`

const Sha256_deploy_olm_description_md = "e821fdae5dc993ff1bf79e6393aa965ba9d2d9488462e47ebe2daff6bb83bf2e"

const File_deploy_olm_description_md = `The noobaa operator creates and reconciles a NooBaa system in a Kubernetes/Openshift cluster.

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
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.12.0/install.sh | bash -s 0.12.0
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

const Sha256_deploy_olm_noobaa_operator_clusterserviceversion_yaml = "3b11ab7cce6a4dfc36ad13f75b37821c8e200aec4cf21007208948e74ce9cc44"

const File_deploy_olm_noobaa_operator_clusterserviceversion_yaml = `apiVersion: operators.coreos.com/v1alpha1
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
    alm-examples: placeholder
    operators.openshift.io/infrastructure-features: '["disconnected"]'
  name: placeholder
  namespace: placeholder
spec:
  displayName: NooBaa Operator
  version: "999.999.999-placeholder"
  minKubeVersion: 1.16.0
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
  - supported: true
    type: AllNamespaces
  install:
    strategy: deployment
    spec: {}
  description: placeholder
  icon:
  - mediatype: image/png
    base64data: placeholder`

const Sha256_deploy_olm_noobaa_icon_base64 = "4684eb3f4be354c728e210364a7e5e6806b68acb945b6e129ebc4d75fd97073c"

const File_deploy_olm_noobaa_icon_base64 = `iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAIAAAD2HxkiAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAA25pVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDIxIDc5LjE1NTc3MiwgMjAxNC8wMS8xMy0xOTo0NDowMCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtcE1NOk9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDpBNjQyRDdGQUIxMDkxMUU0QURFMENEMjA1QUJCMENEMyIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDoxOTA3OEQwNDAyRjAxMUU1QjdFQkI4RTFBMzY3NkQxRiIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDoxOTA3OEQwMzAyRjAxMUU1QjdFQkI4RTFBMzY3NkQxRiIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNCAoV2luZG93cykiPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDo5NWU4ZDg3YS1mNGU4LTRlMTYtOGIwYi1hZGIzYzY2OThkOGUiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6QTY0MkQ3RkFCMTA5MTFFNEFERTBDRDIwNUFCQjBDRDMiLz4gPC9yZGY6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`

const Sha256_deploy_olm_operator_group_yaml = "6a81a348f305328e33cea2dd4fa6b16581995c323b8cd2e8c698fdecabe750bb"

const File_deploy_olm_operator_group_yaml = `apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: noobaa-operator-group
  namespace: default
spec:
  targetNamespaces:
  - default
`

const Sha256_deploy_olm_operator_source_yaml = "2f5cc3b1bec5332087fd6f3b80f0769c404a513d061ef822604fb87b6f301f30"

const File_deploy_olm_operator_source_yaml = `apiVersion: operators.coreos.com/v1
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

const Sha256_deploy_olm_operator_subscription_yaml = "77611fd0b8ec510d277f3f9cb7eb7f8845f2b0fda04732bf9111887d2855d7d3"

const File_deploy_olm_operator_subscription_yaml = `apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: noobaa-operator-subscription
  namespace: default
spec:
  channel: alpha
  name: noobaa-operator
  source: noobaa-operator-catalog
  sourceNamespace: default
`

const Sha256_deploy_operator_yaml = "5399fbfcd1c421acd978f2762d6f8c5048d68fb3c5acc79a595d62cd035a3bc0"

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
        # Notice that changing the serviceAccountName would need to update existing AWS STS role trust policy for customers
      serviceAccountName: noobaa
      securityContext:
        seccompProfile:
          type: RuntimeDefault
      volumes:
      # This service account token can be used to provide identity outside the cluster.
      # For example, this token can be used with AssumeRoleWithWebIdentity to authenticate with AWS using IAM OIDC provider and STS.
      - name: bound-sa-token
        projected:
          sources:
          - serviceAccountToken:
              path: token
              # For testing purposes change the audience to api
              audience: openshift
      # SHOULD BE RETURNED ONCE COSI IS BACK
      # - name: socket
      #   emptyDir: {}
      - name: ocp-injected-ca-bundle
        configMap:
          name: ocp-injected-ca-bundle
          items:
          - key: ca-bundle.crt
            path: ca-bundle.crt
          optional: true
      containers:
        - name: noobaa-operator
          image: NOOBAA_OPERATOR_IMAGE
          volumeMounts:
          - name: bound-sa-token
            mountPath: /var/run/secrets/openshift/serviceaccount
            readOnly: true
          - name: ocp-injected-ca-bundle
            mountPath: /etc/ocp-injected-ca-bundle
          # SHOULD BE RETURNED ONCE COSI IS BACK
          # - name: socket
          #   mountPath: /var/lib/cosi
          resources:
            limits:
              cpu: "250m"
              memory: "512Mi"
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
        # SHOULD BE RETURNED ONCE COSI IS BACK
        # - name: objectstorage-provisioner-sidecar
        #   image: COSI_SIDECAR_IMAGE
        #   args:
        #   - "--v=5"
        #   resources:
        #     limits:
        #       cpu: "100m"
        #       memory: "512Mi"
        #   imagePullPolicy: Always
        #   env:
        #   - name: POD_NAMESPACE
        #     valueFrom:
        #       fieldRef:
        #         fieldPath: metadata.namespace
        #   volumeMounts:
        #   - mountPath: /var/lib/cosi
        #     name: socket
`

const Sha256_deploy_role_yaml = "e145ce24b4267e2e0e63ab56442295bcc605bdc4f6ef723ad6cc15fd38973101"

const File_deploy_role_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: noobaa
rules:
- apiGroups:
  - security.openshift.io
  resourceNames:
  - noobaa
  resources:
  - securitycontextconstraints
  verbs:
  - use
- apiGroups:
  - noobaa.io
  resources:
  - '*'
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
  - serviceaccounts
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
- apiGroups:
  - ceph.rook.io
  resources:
  - cephclusters
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - route.openshift.io
  resources:
  - routes
  verbs:
  - get
  - create
  - update
  - list
  - watch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - get
  - create
  - update
  - patch
  - list
  - watch
  - delete
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - create
  - update
  - patch
  - list
  - watch
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - '*'
  verbs: 
  - '*'
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

const Sha256_deploy_role_binding_auth_delegator_hpav2_yaml = "124f08dd9f953821b4318a33fda3ffbccd477d604dedefc918e2600e5fe51174"

const File_deploy_role_binding_auth_delegator_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-system-auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: custom-metrics-prometheus-adapter
`

const Sha256_deploy_role_binding_auth_reader_hpav2_yaml = "342faf51f7e0f2e718a1479ed4a74cca0775a4f8328215a552e0f5ecde31d092"

const File_deploy_role_binding_auth_reader_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-auth-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: custom-metrics-prometheus-adapter
`

const Sha256_deploy_role_binding_core_yaml = "23dd0d60002ea999cc9f7e10fb3a8000e2c19f8a3ee27971f443acd06f698729"

const File_deploy_role_binding_core_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: noobaa-core
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: noobaa-core
subjects:
- kind: ServiceAccount
  name: noobaa-core
`

const Sha256_deploy_role_binding_db_yaml = "3a4872fcde50e692ae52bbd208a8e1d115c574431c25a9644a7c820ae13c7748"

const File_deploy_role_binding_db_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: noobaa-db
subjects:
  - kind: ServiceAccount
    name: noobaa-db
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: noobaa-db
`

const Sha256_deploy_role_binding_endpoint_yaml = "ab85a33434b0a5fb685fb4983a9d97313e277a5ec8a2142e49c276d51b8ba0e9"

const File_deploy_role_binding_endpoint_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: noobaa-endpoint
subjects:
  - kind: ServiceAccount
    name: noobaa-endpoint
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: noobaa-endpoint
`

const Sha256_deploy_role_binding_resource_reader_hpav2_yaml = "d49b425d57233840163f3b19551eb638bf45b4ae2a45e2674d8ad72d2d7766b6"

const File_deploy_role_binding_resource_reader_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-resource-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-adapter-resource-reader
subjects:
- kind: ServiceAccount
  name: custom-metrics-prometheus-adapter
`

const Sha256_deploy_role_binding_server_resources_hpav2_yaml = "ca281230973b36e8eebcfa851e1522a4f229e44dfd253fadb170c049ea2885f5"

const File_deploy_role_binding_server_resources_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-hpa-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prometheus-adapter-server-resources
subjects:
- kind: ServiceAccount
  name: custom-metrics-prometheus-adapter
`

const Sha256_deploy_role_core_yaml = "c3cfb5b87298224fd6e4e4bff32d3948ad168a0110b8569118a260739ef5d5e7"

const File_deploy_role_core_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: noobaa-core
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
  - serviceaccounts
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
  - security.openshift.io
  resourceNames:
  - noobaa-core
  resources:
  - securitycontextconstraints
  verbs:
  - use
`

const Sha256_deploy_role_db_yaml = "bc7eeca1125dfcdb491ab8eb69e3dcbce9f004a467b88489f85678b3c6872cce"

const File_deploy_role_db_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: noobaa-db
rules:
  - apiGroups:
      - security.openshift.io
    resourceNames:
      - noobaa-db
    resources:
      - securitycontextconstraints
    verbs:
      - use
`

const Sha256_deploy_role_endpoint_yaml = "27ace6cdcae4d87add5ae79265c4eee9d247e5910fc8a74368139d31add6dac2"

const File_deploy_role_endpoint_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: noobaa-endpoint
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
  - serviceaccounts
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
  - security.openshift.io 
  resourceNames:
  - noobaa-endpoint 
  resources:
  - securitycontextconstraints 
  verbs: 
  - use
`

const Sha256_deploy_role_resource_reader_hpav2_yaml = "e5a0adb9c762cdf7ce08da61663e00d1f039013ecb3e921adb19331ecc78b57c"

const File_deploy_role_resource_reader_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-resource-reader
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - pods
  - services
  - configmaps
  verbs:
  - get
  - list
  - watch
`

const Sha256_deploy_role_server_resources_hpav2_yaml = "580f0e861bf4571d81351fb00401b8c77aa9c4443edf662ecacd8329b0a9f4de"

const File_deploy_role_server_resources_hpav2_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app: prometheus-adapter
  name: prometheus-adapter-server-resources
rules:
- apiGroups:
  - custom.metrics.k8s.io
  resources: ["*"]
  verbs: 
    - get
    - list
    - watch
    - create
    - update
    - delete
`

const Sha256_deploy_role_ui_yaml = "d1929d57d9d9bf021e83be275ea45ec17f83a700d9a74936742ba50c41d2c0bb"

const File_deploy_role_ui_yaml = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: noobaa-odf-ui
rules:
  - verbs:
      - get
      - watch
      - list
    apiGroups:
      - noobaa.io
    resources:
      - noobaas
      - bucketclasses
`

const Sha256_deploy_scc_yaml = "baa4d3a3def2d63a5d9e53bc4fc1ac961f9b4fe5172db7118d1529caa14e2191"

const File_deploy_scc_yaml = `apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: noobaa
requiredDropCapabilities:
  - ALL
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
readOnlyRootFilesystem: true
`

const Sha256_deploy_scc_core_yaml = "dd3fb26a323dddbbb9f399b8ff86c41dbbfe63b3bbb0cfe79b785c68948063a8"

const File_deploy_scc_core_yaml = `apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: noobaa-core
allowPrivilegeEscalation: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: false
requiredDropCapabilities:
  - ALL
fsGroup:
  type: MustRunAs
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
`

const Sha256_deploy_scc_db_yaml = "cea49b11eead99f2704b3f36349473fe2961be6312dbcf5ea56a13ebe3075ee2"

const File_deploy_scc_db_yaml = `apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: noobaa-db
allowPrivilegeEscalation: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: false
requiredDropCapabilities:
  - ALL
fsGroup:
  type: RunAsAny
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
`

const Sha256_deploy_scc_endpoint_yaml = "b540b01b4e31dde0c5ff93c116f7873350a186c8985a35e21388091b63b221c7"

const File_deploy_scc_endpoint_yaml = `apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: noobaa-endpoint
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
- SETUID
- SETGID
defaultAddCapabilities: []
fsGroup:
  type: MustRunAs
groups: []
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities:
  - ALL
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`

const Sha256_deploy_service_account_yaml = "7c68e5bd65c614787d7d4cdf80db8c14d9159ce8e940c5134d33d21dbe66893f"

const File_deploy_service_account_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa
  annotations:
    serviceaccounts.openshift.io/oauth-redirectreference.noobaa-mgmt: '{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"noobaa-mgmt"}}'
`

const Sha256_deploy_service_account_core_yaml = "7e8f1d49bdba0969a33e8acc676cc5e2d50af9f4c94112b6de07548f3f704c24"

const File_deploy_service_account_core_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa-core

`

const Sha256_deploy_service_account_db_yaml = "fcbccd7518ee5a426b071a3acc85d22142e27c5628b61ce4292cc393d2ecac31"

const File_deploy_service_account_db_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa-db
`

const Sha256_deploy_service_account_endpoint_yaml = "c2331e027114658e48a2bd1139b00cce06dfd834aa682eae923de54874a6baed"

const File_deploy_service_account_endpoint_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa-endpoint
`

const Sha256_deploy_service_account_ui_yaml = "d6cb0e92fdb350148399e1ac42bfa640e254bdbb295c9a15dc9edfd4335e73f6"

const File_deploy_service_account_ui_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: noobaa-odf-ui
`

const Sha256_deploy_service_acount_hpav2_yaml = "c7c9ec142994295c1fa4315ca6cd4340a195aa0b02fa2b4fc0b9f03ae4589c94"

const File_deploy_service_acount_hpav2_yaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: prometheus-adapter
  name: custom-metrics-prometheus-adapter
`

