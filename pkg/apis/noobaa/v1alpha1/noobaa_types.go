package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
)

// Note 1: Run "make gen-api" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&NooBaa{}, &NooBaaList{})
}

// Labels are label for a given daemon
type Labels map[string]string

// LabelsSpec is the main spec label for all daemons
type LabelsSpec map[string]Labels

// Annotations are annotation for a given daemon
type Annotations map[string]string

// AnnotationsSpec is the main spec annotation for all daemons
type AnnotationsSpec map[string]Annotations

// NooBaa is the Schema for the NooBaas API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=nb
// +kubebuilder:printcolumn:name="S3-Endpoints",type="string",JSONPath=".status.services.serviceS3.nodePorts",description="S3 Endpoints"
// +kubebuilder:printcolumn:name="Sts-Endpoints",type="string",JSONPath=".status.services.serviceSts.nodePorts",description="STS Endpoints"
// +kubebuilder:printcolumn:name="Iam-Endpoints",type="string",JSONPath=".status.services.serviceIam.nodePorts",description="IAM Endpoints"
// +kubebuilder:printcolumn:name="Syslog-Endpoints",type="string",JSONPath=".status.services.serviceSyslog.nodePorts",description="Syslog Endpoints"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.actualImage",description="Actual Image"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NooBaa struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the noobaa system.
	// +optional
	Spec NooBaaSpec `json:"spec,omitempty"`

	// Most recently observed status of the noobaa system.
	// +optional
	Status NooBaaStatus `json:"status,omitempty"`
}

// NooBaaList contains a list of noobaa systems
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NooBaaList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Systems.
	Items []NooBaa `json:"items"`
}

// NooBaaSpec defines the desired state of System
// +k8s:openapi-gen=true
type NooBaaSpec struct {

	// Image (optional) overrides the default image for the server container
	// +optional
	Image *string `json:"image,omitempty"`

	// DBSpec (optional) DB spec for a managed postgres cluster
	// +optional
	DBSpec *NooBaaDBSpec `json:"dbSpec,omitempty"`

	// DBImage (optional) overrides the default image for the db container
	// +optional
	DBImage *string `json:"dbImage,omitempty"`

	// DBConf (optional) overrides the default postgresql db config
	// +optional
	DBConf *string `json:"dbConf,omitempty"`

	// DBType (optional) overrides the default type image for the db container.
	// The only possible value is postgres
	// +optional
	// +kubebuilder:validation:Enum=postgres
	DBType DBTypes `json:"dbType,omitempty"`

	// CoreResources (optional) overrides the default resource requirements for the server container
	// +optional
	CoreResources *corev1.ResourceRequirements `json:"coreResources,omitempty"`

	// LogResources (optional) overrides the default resource requirements for the noobaa-log-processor container
	// +optional
	LogResources *corev1.ResourceRequirements `json:"logResources,omitempty"`

	// DBResources (optional) overrides the default resource requirements for the db container
	// +optional
	DBResources *corev1.ResourceRequirements `json:"dbResources,omitempty"`

	// DBVolumeResources (optional) overrides the default PVC resource requirements for the database volume.
	// For the time being this field is immutable and can only be set on system creation.
	// This is because volume size updates are only supported for increasing the size,
	// and only if the storage class specifies `allowVolumeExpansion: true`,
	// +immutable
	// +optional
	DBVolumeResources *corev1.VolumeResourceRequirements `json:"dbVolumeResources,omitempty"`

	// DBStorageClass (optional) overrides the default cluster StorageClass for the database volume.
	// For the time being this field is immutable and can only be set on system creation.
	// This affects where the system stores its database which contains system config,
	// buckets, objects meta-data and mapping file parts to storage locations.
	// +immutable
	// +optional
	DBStorageClass *string `json:"dbStorageClass,omitempty"`

	// ExternalPgSecret (optional) holds an optional secret with a url to an extrenal Postgres DB to be used
	// +optional
	ExternalPgSecret *corev1.SecretReference `json:"externalPgSecret,omitempty"`

	// ExternalPgSSLRequired (optional) holds an optional boolean to force ssl connections to the external Postgres DB
	// +optional
	ExternalPgSSLRequired bool `json:"externalPgSSLRequired,omitempty"`

	// ExternalPgSSLUnauthorized (optional) holds an optional boolean to allow unauthorized connections to external Postgres DB
	// +optional
	ExternalPgSSLUnauthorized bool `json:"externalPgSSLUnauthorized,omitempty"`

	// ExternalPgSSLSecret (optional) holds an optional secret with client key and cert used for connecting to external Postgres DB
	// +optional
	ExternalPgSSLSecret *corev1.SecretReference `json:"externalPgSSLSecret,omitempty"`

	// DebugLevel (optional) sets the debug level
	// +optional
	// +kubebuilder:validation:Enum=all;nsfs;warn;default_level
	DebugLevel int `json:"debugLevel,omitempty"`

	// PVPoolDefaultStorageClass (optional) overrides the default cluster StorageClass for the pv-pool volumes.
	// This affects where the system stores data chunks (encrypted).
	// Updates to this field will only affect new pv-pools,
	// but updates to existing pools are not supported by the operator.
	// +optional
	PVPoolDefaultStorageClass *string `json:"pvPoolDefaultStorageClass,omitempty"`

	// Tolerations (optional) passed through to noobaa's pods
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity (optional) passed through to noobaa's pods
	// +optional
	Affinity *AffinitySpec `json:"affinity,omitempty"`

	// ImagePullSecret (optional) sets a pull secret for the system image
	// +optional
	ImagePullSecret *corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// Region (optional) provide a region for the location info
	// of the endpoints in the endpoint deployment
	// +optional
	Region *string `json:"region,omitempty"`

	// Endpoints (optional) sets configuration info for the noobaa endpoint
	// deployment.
	// +optional
	Endpoints *EndpointsSpec `json:"endpoints,omitempty"`

	// JoinSecret (optional) instructs the operator to join another cluster
	// and point to a secret that holds the join information
	// +optional
	JoinSecret *corev1.SecretReference `json:"joinSecret,omitempty"`

	// CleanupPolicy (optional) Indicates user's policy for deletion
	// +optional
	CleanupPolicy CleanupPolicySpec `json:"cleanupPolicy,omitempty"`

	// Security represents security settings
	Security SecuritySpec `json:"security,omitempty"`

	// The labels-related configuration to add/set on each Pod related object.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +nullable
	// +optional
	Labels LabelsSpec `json:"labels,omitempty"`

	// The annotations-related configuration to add/set on each Pod related object.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +nullable
	// +optional
	Annotations AnnotationsSpec `json:"annotations,omitempty"`

	// DisableLoadBalancerService (optional) sets the service type to ClusterIP instead of LoadBalancer
	// +nullable
	// +optional
	DisableLoadBalancerService bool `json:"disableLoadBalancerService,omitempty"`

	// DisableRoutes (optional) disables the reconciliation of openshift route resources in the cluster
	// +nullable
	// +optional
	DisableRoutes bool `json:"disableRoutes,omitempty"`

	// Deprecated: DefaultBackingStoreSpec is not supported anymore, use ManualDefaultBackingStore instead.
	// +optional
	DefaultBackingStoreSpec *BackingStoreSpec `json:"defaultBackingStoreSpec,omitempty"`

	// ManualDefaultBackingStore (optional - default value is false) if true the default backingstore/namespacestore
	// will not be reconciled by the operator and it should be manually handled by the user. It will allow the
	// user to  delete DefaultBackingStore/DefaultNamespaceStore, user needs to delete associated buckets and
	// update the admin account with new BackingStore/NamespaceStore in order to delete the DefaultBackingStore/DefaultNamespaceStore
	// +nullable
	// +optional
	ManualDefaultBackingStore bool `json:"manualDefaultBackingStore,omitempty"`

	// LoadBalancerSourceSubnets (optional) if given will allow access to the NooBaa services
	// only from the listed subnets. This field will have no effect if DisableLoadBalancerService is set
	// to true
	// +optional
	LoadBalancerSourceSubnets LoadBalancerSourceSubnetSpec `json:"loadBalancerSourceSubnets,omitempty"`

	// Configuration related to autoscaling
	// +optional
	Autoscaler AutoscalerSpec `json:"autoscaler,omitempty"`

	// DenyHTTP (optional) if given will deny access to the NooBaa S3 service using HTTP (only HTTPS)
	// +optional
	DenyHTTP bool `json:"denyHTTP,omitempty"`

	// BucketLogging sets the configuration for bucket logging
	// +optional
	BucketLogging BucketLoggingSpec `json:"bucketLogging,omitempty"`

	// BucketNotifications (optional) controls bucket notification options
	// +optional
	BucketNotifications BucketNotificationsSpec `json:"bucketNotifications,omitempty"`
}

// Affinity is a group of affinity scheduling rules.
type AffinitySpec struct {
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty" protobuf:"bytes,1,opt,name=nodeAffinity"`
	// Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
	// +optional
	PodAffinity *corev1.PodAffinity `json:"podAffinity,omitempty" protobuf:"bytes,2,opt,name=podAffinity"`
	// Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).
	// +optional
	PodAntiAffinity *corev1.PodAntiAffinity `json:"podAntiAffinity,omitempty" protobuf:"bytes,3,opt,name=podAntiAffinity"`

	// TopologyKey (optional) the TopologyKey to pass as the domain for TopologySpreadConstraint and Affinity of noobaa components
	// It is used by the endpoints and the DB pods to control pods distribution between topology domains (host/zone)
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`
}

// AutoscalerSpec defines different actoscaling spec such as autoscaler type and prometheus namespace
type AutoscalerSpec struct {
	// Type of autoscaling (optional) for noobaa-endpoint, hpav2(default) and keda - Prometheus metrics based
	// +kubebuilder:validation:Enum=hpav2;keda
	// +optional
	AutoscalerType AutoscalerTypes `json:"autoscalerType,omitempty"`

	// Prometheus namespace that scrap metrics from noobaa
	// +optional
	PrometheusNamespace string `json:"prometheusNamespace,omitempty"`
}

// BucketLoggingSpec defines the bucket logging configuration
type BucketLoggingSpec struct {
	// LoggingType specifies the type of logging for the bucket
	// There are two types available: best-effort and guaranteed logging
	// - best-effort(default) - less immune to failures but with better performance
	// - guaranteed - much more reliable but need to provide a storage class that supports RWX PVs
	// +optional
	LoggingType BucketLoggingTypes `json:"loggingType,omitempty"`

	// BucketLoggingPVC (optional) specifies the name of the Persistent Volume Claim (PVC) to be used
	// for guaranteed logging when the logging type is set to 'guaranteed'. The PVC must support
	// ReadWriteMany (RWX) access mode to ensure reliable logging.
	// For ODF: If not provided, the default CephFS storage class will be used to create the PVC.
	// +optional
	BucketLoggingPVC *string `json:"bucketLoggingPVC,omitempty"`
}

// BucketNotificationsSpec controls bucket notification configuration
type BucketNotificationsSpec struct {
	// Enabled - whether bucket notifications is enabled
	Enabled bool `json:"enabled"`

	//PVC (optional) specifies the name of the Persistent Volume Claim (PVC) to be used
	//for holding pending notifications files.
	//For ODF - If not provided, the default CepthFS storage class will be used to create the PVC.
	// +optional
	PVC *string `json:"pvc,omitempty"`

	//Connections - A list of secrets' names that are used by the notifications configrations
	//(in the TopicArn field).
	Connections []corev1.SecretReference `json:"connections,omitempty"`
}

// LoadBalancerSourceSubnetSpec defines the subnets that will be allowed to access the NooBaa services
type LoadBalancerSourceSubnetSpec struct {
	// S3 is a list of subnets that will be allowed to access the Noobaa S3 service
	// +optional
	S3 []string `json:"s3,omitempty"`

	// STS is a list of subnets that will be allowed to access the Noobaa STS service
	// +optional
	STS []string `json:"sts,omitempty"`

	// IAM is a list of subnets that will be allowed to access the Noobaa IAM service
	// +optional
	IAM []string `json:"iam,omitempty"`
}

// SecuritySpec is security spec to include various security items such as kms
type SecuritySpec struct {
	KeyManagementService KeyManagementServiceSpec `json:"kms,omitempty"`
}

// KeyManagementServiceSpec represent various details of the KMS server
type KeyManagementServiceSpec struct {
	// +optional
	EnableKeyRotation bool `json:"enableKeyRotation,omitempty"`
	// +optional
	Schedule          string            `json:"schedule,omitempty"`
	ConnectionDetails map[string]string `json:"connectionDetails,omitempty"`
	TokenSecretName   string            `json:"tokenSecretName,omitempty"`
}

// NooBaaDBSpec defines the desired state of a managed postgres cluster
// +k8s:openapi-gen=true
type NooBaaDBSpec struct {
	// DBImage (optional) overrides the default image for the db instances
	// +optional
	DBImage *string `json:"image,omitempty"`

	// PostgresMajorVersion (optional) overrides the default postgres major version
	// It is the user's responsibility to ensure that the postgres image matches the major version.
	PostgresMajorVersion *int `json:"postgresMajorVersion,omitempty"`

	// Instances (optional) overrides the default number of db instances
	// +optional
	Instances *int `json:"instances,omitempty"`

	// DBResources (optional) overrides the default resource requirements for the db container
	// +optional
	DBResources *corev1.ResourceRequirements `json:"dbResources,omitempty"`

	// DBMinVolumeSize (optional) The initial size of the database volume.The actual size might be larger.
	// Increasing the size of the volume is supported if the underlying storage class supports volume expansion.
	// The new size should be larger than actualVolumeSize in dbStatus for the volume to be resized.
	// +optional
	DBMinVolumeSize string `json:"dbMinVolumeSize,omitempty"`

	// DBStorageClass (optional) overrides the default cluster StorageClass for the database volume.
	// +optional
	DBStorageClass *string `json:"dbStorageClass,omitempty"`

	// DBConf (optional) overrides the default postgresql db config
	// +optional
	DBConf map[string]string `json:"dbConf,omitempty"`

	// DBBackup (optional) configure automatic scheduled backups of the database volume.
	// Currently, only volume snapshots are supported.
	// +optional
	DBBackup *DBBackupSpec `json:"dbBackup,omitempty"`

	// DBRecovery (optional) configure database recovery from snapshot
	// +optional
	DBRecovery *DBRecoverySpec `json:"dbRecovery,omitempty"`
}

// DBBackupSpec defines the desired parameters for the database automatic backup
type DBBackupSpec struct {
	// Schedule the schedule for the database backup in cron format.
	// +kubebuilder:validation:Required
	Schedule string `json:"schedule"`

	// VolumeSnapshot the volume snapshot backup configuration.
	// Currently this is the only supported backup method and hence it is required.
	// +kubebuilder:validation:Required
	VolumeSnapshot *VolumeSnapshotBackupSpec `json:"volumeSnapshot,omitempty"`
}

type VolumeSnapshotBackupSpec struct {
	// VolumeSnapshotClass the volume snapshot class for the database volume.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolumeSnapshotClass string `json:"volumeSnapshotClass"`

	// MaxSnapshots the maximum number of snapshots to keep.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxSnapshots int `json:"maxSnapshots"`
}

// DBRecoverySpec defines the desired parameters for database recovery from snapshot
type DBRecoverySpec struct {
	// VolumeSnapshotName specifies the name of the volume snapshot to recover from
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolumeSnapshotName string `json:"volumeSnapshotName"`
}

// EndpointsSpec defines the desired state of noobaa endpoint deployment
// +k8s:openapi-gen=true
type EndpointsSpec struct {
	// MinCount, the number of endpoint instances (pods)
	// to be used as the lower bound when autoscaling
	MinCount int32 `json:"minCount,omitempty"`

	// MaxCount, the number of endpoint instances (pods)
	// to be used as the upper bound when autoscaling
	MaxCount int32 `json:"maxCount,omitempty"`

	// AdditionalVirtualHosts (optional) provide a list of additional hostnames
	// (on top of the builtin names defined by the cluster: service name, elb name, route name)
	// to be used as virtual hosts by the the endpoints in the endpoint deployment
	// +optional
	AdditionalVirtualHosts []string `json:"additionalVirtualHosts,omitempty"`

	// Resources (optional) overrides the default resource requirements for every endpoint pod
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// NooBaaStatus defines the observed state of System
// +k8s:openapi-gen=true
type NooBaaStatus struct {

	// ObservedGeneration is the most recent generation observed for this noobaa system.
	// It corresponds to the CR generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase is a simple, high-level summary of where the System is in its lifecycle
	// +optional
	Phase SystemPhase `json:"phase,omitempty"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects related to this operator.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`

	// ActualImage is set to report which image the operator is using
	// +optional
	ActualImage string `json:"actualImage,omitempty"`

	// Accounts reports accounts info for the admin account
	// +optional
	Accounts *AccountsStatus `json:"accounts,omitempty"`

	// Services reports addresses for the services
	// +optional
	Services *ServicesStatus `json:"services,omitempty"`

	// Endpoints reports the actual number of endpoints in the endpoint deployment
	// and the virtual hosts list used recognized by the endpoints
	// +optional
	Endpoints *EndpointsStatus `json:"endpoints,omitempty"`

	// Upgrade reports the status of the ongoing upgrade process
	// +optional
	UpgradePhase UpgradePhase `json:"upgradePhase,omitempty"`

	// Upgrade reports the status of the ongoing postgres upgrade process
	// +optional
	PostgresUpdatePhase UpgradePhase `json:"postgresUpdatePhase,omitempty"`

	// Readme is a user readable string with explanations on the system
	// +optional
	Readme string `json:"readme,omitempty"`

	// LastKeyRotateTime is the time system ran an encryption key rotate
	// +optional
	LastKeyRotateTime metav1.Time `json:"lastKeyRotateTime,omitempty"`

	// BeforeUpgradeDbImage is the db image used before last db upgrade
	// +optional
	BeforeUpgradeDbImage *string `json:"beforeUpgradeDbImage,omitempty"`

	// DBStatus is the status of the postgres cluster
	// +optional
	DBStatus *NooBaaDBStatus `json:"dbStatus,omitempty"`
}

// SystemPhase is a string enum type for system phases
type SystemPhase string

// These are the valid phases:
const (

	// SystemPhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Describe the noobaa system to see events.
	SystemPhaseRejected SystemPhase = "Rejected"

	// SystemPhaseVerifying means the operator is verifying the spec
	SystemPhaseVerifying SystemPhase = "Verifying"

	// SystemPhaseCreating means the operator is creating the resources on the cluster
	SystemPhaseCreating SystemPhase = "Creating"

	// SystemPhaseConnecting means the operator is trying to connect to the pods and services it created
	SystemPhaseConnecting SystemPhase = "Connecting"

	// SystemPhaseConfiguring means the operator is configuring the as requested
	SystemPhaseConfiguring SystemPhase = "Configuring"

	// SystemPhaseReady means the noobaa system has been created and ready to serve.
	SystemPhaseReady SystemPhase = "Ready"
)

type NooBaaDBStatus struct {
	// DBClusterStatus is the status of the postgres cluster
	DBClusterStatus DBClusterStatus `json:"dbClusterStatus,omitempty"`

	// DBCurrentImage is the image of the postgres cluster
	DBCurrentImage string `json:"dbCurrentImage,omitempty"`

	// CurrentPgMajorVersion is the major version of the postgres cluster
	CurrentPgMajorVersion int `json:"currentPgMajorVersion,omitempty"`

	// ActualVolumeSize is the actual size of the postgres cluster volume. This can be different than the requested size
	ActualVolumeSize string `json:"actualVolumeSize,omitempty"`

	// BackupStatus reports the status of database backups
	// +optional
	BackupStatus *DBBackupStatus `json:"backupStatus,omitempty"`

	// RecoveryStatus reports the status of database recovery
	// +optional
	RecoveryStatus *DBRecoveryStatus `json:"recoveryStatus,omitempty"`
}

type DBClusterStatus string

const (
	// DBClusterStatusNone means there is no DB cluster configured
	DBClusterStatusNone DBClusterStatus = "None"

	// DBClusterStatusCreating means a new DB cluster is being created
	DBClusterStatusCreating DBClusterStatus = "Creating"

	// DBClusterStatusUpdating means the DB cluster is being updated
	DBClusterStatusUpdating DBClusterStatus = "Updating"

	// DBClusterStatusImporting means a new DB cluster is being created and data is being imported from the previous DB
	DBClusterStatusImporting DBClusterStatus = "Importing"

	// DBClusterStatusRecovering means a new DB cluster is being created and data is being recovered from a volume snapshot
	DBClusterStatusRecovering DBClusterStatus = "Recovering"

	// DBClusterStatusReady means the DB cluster is ready
	DBClusterStatusReady DBClusterStatus = "Ready"

	// DBClusterStatusFailed means the DB cluster reconciliation encountered an error
	DBClusterStatusFailed DBClusterStatus = "Failed"
)

// DBBackupStatus reports the status of database backups
type DBBackupStatus struct {
	// LastBackupTime timestamp of the last successful backup
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`

	// NextBackupTime timestamp of the next scheduled backup
	NextBackupTime *metav1.Time `json:"nextBackupTime,omitempty"`

	// TotalSnapshots current number of snapshots
	TotalSnapshots int `json:"totalSnapshots,omitempty"`

	// AvailableSnapshots list of available snapshot names
	AvailableSnapshots []string `json:"availableSnapshots,omitempty"`
}

// DBRecoveryStatus reports the status of database recovery
type DBRecoveryStatus struct {
	// Status current recovery status
	Status DBRecoveryStatusType `json:"status,omitempty"`

	// SnapshotName name of the snapshot being recovered from
	SnapshotName string `json:"snapshotName,omitempty"`

	// RecoveryTime timestamp when recovery was initiated
	RecoveryTime *metav1.Time `json:"recoveryTime,omitempty"`
}

// DBRecoveryStatusType represents the status of database recovery
type DBRecoveryStatusType string

const (
	// DBRecoveryStatusNone means no recovery is configured or in progress
	DBRecoveryStatusNone DBRecoveryStatusType = "None"

	// DBRecoveryStatusPending means recovery is pending cluster deletion
	DBRecoveryStatusPending DBRecoveryStatusType = "Pending"

	// DBRecoveryStatusRunning means recovery is currently in progress
	DBRecoveryStatusRunning DBRecoveryStatusType = "Running"

	// DBRecoveryStatusCompleted means recovery has completed successfully
	DBRecoveryStatusCompleted DBRecoveryStatusType = "Completed"

	// DBRecoveryStatusFailed means recovery has failed
	DBRecoveryStatusFailed DBRecoveryStatusType = "Failed"
)

// These are the valid conditions types and statuses:
const (
	ConditionTypeKMSStatus conditionsv1.ConditionType = "KMS-Status"
	ConditionTypeKMSType   conditionsv1.ConditionType = "KMS-Type"
)

// These are NooBaa condition statuses
const (
	// External KMS initialized
	ConditionKMSInit corev1.ConditionStatus = "Init"

	// The root key was synchronized from external KMS
	ConditionKMSSync corev1.ConditionStatus = "Sync"

	// The root key was rotated
	ConditionKMSKeyRotate corev1.ConditionStatus = "KeyRotate"

	// Invalid external KMS definition
	ConditionKMSInvalid corev1.ConditionStatus = "Invalid"

	// Error reading secret from external KMS
	ConditionKMSErrorRead corev1.ConditionStatus = "ErrorRead"

	// Error writing initial root key to external KMS
	ConditionKMSErrorWrite corev1.ConditionStatus = "ErrorWrite"

	// Error in data format, internal error
	ConditionKMSErrorData corev1.ConditionStatus = "ErrorData"

	// Error in data format, internal error
	ConditionKMSErrorSecretReconcile corev1.ConditionStatus = "ErrorSecretReconcile"
)

// AccountsStatus is the status info of admin account
type AccountsStatus struct {
	Admin UserStatus `json:"admin"`
}

// ServicesStatus is the status info of the system's services
type ServicesStatus struct {
	ServiceMgmt ServiceStatus `json:"serviceMgmt"`
	ServiceS3   ServiceStatus `json:"serviceS3"`
	// +optional
	ServiceSts    ServiceStatus `json:"serviceSts,omitempty"`
	ServiceIam    ServiceStatus `json:"serviceIam,omitempty"`
	ServiceSyslog ServiceStatus `json:"serviceSyslog,omitempty"`
}

// UserStatus is the status info of a user secret
type UserStatus struct {
	SecretRef corev1.SecretReference `json:"secretRef"`
}

// ServiceStatus is the status info and network addresses of a service
type ServiceStatus struct {

	// NodePorts are the most basic network available.
	// NodePorts use the networks available on the hosts of kubernetes nodes.
	// This generally works from within a pod, and from the internal
	// network of the nodes, but may fail from public network.
	// https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
	// +optional
	NodePorts []string `json:"nodePorts,omitempty"`

	// PodPorts are the second most basic network address.
	// Every pod has an IP in the cluster and the pods network is a mesh
	// so the operator running inside a pod in the cluster can use this address.
	// Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
	// Note2: when running the operator outside of the cluster, pod IP is not accessible.
	// +optional
	PodPorts []string `json:"podPorts,omitempty"`

	// InternalIP are internal addresses of the service inside the cluster
	// https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	InternalIP []string `json:"internalIP,omitempty"`

	// InternalDNS are internal addresses of the service inside the cluster
	// +optional
	InternalDNS []string `json:"internalDNS,omitempty"`

	// ExternalIP are external public addresses for the service
	// LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
	// IngressPorts are manually created public addresses for the service
	// https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
	// https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
	// https://kubernetes.io/docs/concepts/services-networking/ingress/
	// +optional
	ExternalIP []string `json:"externalIP,omitempty"`

	// ExternalDNS are external public addresses for the service
	// +optional
	ExternalDNS []string `json:"externalDNS,omitempty"`
}

// EndpointsStatus is the status info for the endpoints deployment
type EndpointsStatus struct {
	ReadyCount   int32    `json:"readyCount"`
	VirtualHosts []string `json:"virtualHosts"`
}

// UpgradePhase is a string enum type for upgrade phases
type UpgradePhase string

// These are the valid phases:
const (
	UpgradePhaseNone UpgradePhase = "NoUpgrade"

	UpgradePhasePrepare UpgradePhase = "Preparing"

	UpgradePhaseMigrate UpgradePhase = "Migrating"

	UpgradePhaseClean UpgradePhase = "Cleanning"

	UpgradePhaseFinished UpgradePhase = "DoneUpgrade"

	UpgradePhaseReverting UpgradePhase = "Reverting"

	UpgradePhaseFailed UpgradePhase = "Failed"

	UpgradePhaseUpgrade UpgradePhase = "Upgrading"
)

// CleanupPolicySpec specifies the cleanup policy
type CleanupPolicySpec struct {
	Confirmation CleanupConfirmationProperty `json:"confirmation,omitempty"`

	// +optional
	AllowNoobaaDeletion bool `json:"allowNoobaaDeletion,omitempty"`
}

// CleanupConfirmationProperty is a string that specifies cleanup confirmation
type CleanupConfirmationProperty string

const (
	// Finalizer is the name of the noobaa finalizer
	Finalizer = "noobaa.io/finalizer"

	// GracefulFinalizer is the name of the noobaa graceful finalizer
	GracefulFinalizer = "noobaa.io/graceful_finalizer"

	// DeleteOBCConfirmation represents the validation to destry obc
	DeleteOBCConfirmation CleanupConfirmationProperty = "yes-really-destroy-obc"

	// SkipTopologyConstraints is Annotation name for skipping the reconciliation of the default topology Constraints
	SkipTopologyConstraints = "noobaa.io/skip_topology_spread_constraints"

	// DisableDBDefaultMonitoring is Annotation name for disabling default db monitoring
	DisableDBDefaultMonitoring = "noobaa.io/disable_db_default_monitoring"
)

// DBTypes is a string enum type for specify the types of DB that are supported.
type DBTypes string

// These are the valid DB types:
const (
	// DBTypePostgres is postgres
	DBTypePostgres DBTypes = "postgres"
)

// AutoscalerTypes is a string enum type for specifying the types of autoscaling supported.
type AutoscalerTypes string

// These are the valid AutoscalerTypes types:
const (
	// AutoscalerTypeKeda is keda
	AutoscalerTypeKeda AutoscalerTypes = "keda"
	// AutoscalerTypeHPAV2 is hpav2
	AutoscalerTypeHPAV2 AutoscalerTypes = "hpav2"
)

// BucketLoggingTypes is a string enum type for specifying the types of bucketlogging supported.
type BucketLoggingTypes string

// These are the valid BucketLoggingTypes types:
const (
	// BucketLoggingTypeBestEffort is best-effort
	BucketLoggingTypeBestEffort BucketLoggingTypes = "best-effort"

	// BucketLoggingTypeGuaranteed is guaranteed
	BucketLoggingTypeGuaranteed BucketLoggingTypes = "guaranteed"
)
