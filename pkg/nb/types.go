package nb

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

const (
	petaInBytes = 1024 * 1024 * 1024 * 1024 * 1024
	maskString  = "****"
)

// MaskedString is a string type for sensitive string, masked when formatted
type MaskedString string

func (MaskedString) String() string {
	return maskString
}

// SystemInfo is a struct of system info returned by the API
type SystemInfo struct {
	Accounts           []AccountInfo           `json:"accounts"`
	Buckets            []BucketInfo            `json:"buckets"`
	Pools              []PoolInfo              `json:"pools"`
	Tiers              []TierInfo              `json:"tiers"`
	Version            string                  `json:"version"`
	NamespaceResources []NamespaceResourceInfo `json:"namespace_resources"`
	// TODO SystemInfo struct is partial ...
}

// AccountInfo is a struct of account info returned by the API
type AccountInfo struct {
	Name               string         `json:"name"`
	Email              string         `json:"email"`
	IsSupport          bool           `json:"is_support"`
	HasLogin           bool           `json:"has_login"`
	HasS3Access        bool           `json:"has_s3_access"`
	CanCreateBuckets   bool           `json:"can_create_buckets"`
	NextPasswordChange int64          `json:"next_password_change"`
	DefaultResource    string         `json:"default_resource"`
	AccessKeys         []S3AccessKeys `json:"access_keys"`
	AllowedIPs         []struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"allowed_ips"`
	ExternalConnections struct {
		Count       int                      `json:"count"`
		Connections []ExternalConnectionInfo `json:"connections"`
	} `json:"external_connections"`
	Systems []struct {
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	} `json:"systems"`
	Preferences struct {
		UITheme string `json:"ui_theme"`
	} `json:"preferences"`
	// NsfsAccountConfig specifies the configurations on Namespace FS
	// +nullable
	// +optional
	NsfsAccountConfig *nbv1.AccountNsfsConfig `json:"nsfs_account_config,omitempty"`
}

// BucketInfo is a struct of bucket info returned by the API
type BucketInfo struct {
	Name         string `json:"name"`
	BucketType   string `json:"bucket_type"`
	Mode         string `json:"mode"`
	Undeletable  string `json:"undeletable"`
	ForceMd5Etag *bool  `json:"force_md5_etag,omitempty"`

	BucketClaim  *BucketClaimInfo   `json:"bucket_claim,omitempty"`
	Tiering      *TieringPolicyInfo `json:"tiering,omitempty"`
	DataCapacity *struct {
		Size                      *BigInt `json:"size,omitempty"`
		SizeReduced               *BigInt `json:"size_reduced,omitempty"`
		Free                      *BigInt `json:"free,omitempty"`
		AvailableSizeToUpload     *BigInt `json:"available_size_for_upload,omitempty"`
		AvailableQuantityToUpload *BigInt `json:"available_quantity_for_upload,omitempty"`
		LastUpdate                int64   `json:"last_update"`
	} `json:"data,omitempty"`
	StorageCapacity *struct {
		Values     *StorageInfo `json:"values,omitempty"`
		LastUpdate int64        `json:"last_update"`
	} `json:"storage,omitempty"`
	NumObjects *struct {
		Value      int64 `json:"value"`
		LastUpdate int64 `json:"last_update"`
	} `json:"num_objects,omitempty"`
	Quota       *QuotaConfig `json:"quota,omitempty"`
	PolicyModes *struct {
		ResiliencyStatus string `json:"resiliency_status"`
		QuotaStatus      string `json:"quota_status"`
	} `json:"policy_modes,omitempty"`
	Namespace *NamespaceBucketInfo `json:"namespace,omitempty"`
	// TODO BucketInfo struct is partial ...
}

// TieringPolicyInfo is the information of a tiering policy
type TieringPolicyInfo struct {
	Name             string            `json:"name"`
	Tiers            []TierItem        `json:"tiers"`
	ChunkSplitConfig *ChunkSplitConfig `json:"chunk_split_config,omitempty"`
	DataCapacity     *StorageInfo      `json:"data,omitempty"`
	StorageCapacity  *StorageInfo      `json:"storage,omitempty"`
	Mode             string            `json:"mode,omitempty"`
}

// TierInfo is the information of a tier
type TierInfo struct {
	Name             string            `json:"name"`
	DataPlacement    string            `json:"data_placement,omitempty"`
	AttachedPools    []string          `json:"attached_pools,omitempty"`
	ChunkCoderConfig *ChunkCoderConfig `json:"chunk_coder_config,omitempty"`
	DataCapacity     *StorageInfo      `json:"data,omitempty"`
	StorageCapacity  *StorageInfo      `json:"storage,omitempty"`
}

// StorageInfo contains storage capacity information with specific break down
type StorageInfo struct {
	Total           *BigInt `json:"total,omitempty"`
	Free            *BigInt `json:"free,omitempty"`
	UnavailableFree *BigInt `json:"unavailable_free,omitempty"`
	UnavailableUsed *BigInt `json:"unavailable_used,omitempty"`
	Used            *BigInt `json:"used,omitempty"`
	UsedOther       *BigInt `json:"used_other,omitempty"`
	UsedReduced     *BigInt `json:"used_reduced,omitempty"`
	Alloc           *BigInt `json:"alloc,omitempty"`
	Limit           *BigInt `json:"limit,omitempty"`
	Reserved        *BigInt `json:"reserved,omitempty"`
	Real            *BigInt `json:"real,omitempty"`
}

// BigInt is an api type to handle large integers that cannot be represented by JSON which is limited to 53 bits (less than 8 PB)
type BigInt struct {
	N    int64 `json:"n"`
	Peta int64 `json:"peta"`
}

// MarshalJSON is custom marshalling because the json schema is oneOf integer or {n,peta}
func (n BigInt) MarshalJSON() ([]byte, error) {
	if n.Peta == 0 {
		return json.Marshal(n.N)
	}
	type bigint BigInt
	return json.Marshal(bigint(n))
}

// UnmarshalJSON is custom unmarshalling because the json schema is oneOf integer or {n,peta}
func (n *BigInt) UnmarshalJSON(data []byte) error {
	var i int64
	err := json.Unmarshal(data, &i)
	if err == nil {
		n.N = i
		n.Peta = 0
		return nil
	}
	type bigint BigInt
	return json.Unmarshal(data, (*bigint)(n))
}

// ToString convert bigInt to string
func (n *BigInt) ToString() string {
	return strconv.FormatInt((n.N + (n.Peta * petaInBytes)), 10)
}

// PoolInfo is a struct of pool info returned by the API
type PoolInfo struct {
	Name         string `json:"name"`
	ResourceType string `json:"resource_type"`
	Mode         string `json:"mode,omitempty"`
	Region       string `json:"region,omitempty"`
	PoolNodeType string `json:"pool_node_type,omitempty"`
	Undeletable  string `json:"undeletable,omitempty"`
	CloudInfo    *struct {
		EndpointType EndpointType    `json:"endpoint_type,omitempty"`
		Endpoint     string          `json:"endpoint,omitempty"`
		TargetBucket string          `json:"target_bucket,omitempty"`
		Identity     string          `json:"identity,omitempty"`
		NodeName     string          `json:"node_name,omitempty"`
		CreatedBy    string          `json:"created_by,omitempty"`
		Host         string          `json:"host,omitempty"`
		AuthMethod   CloudAuthMethod `json:"auth_method,omitempty"`
	} `json:"cloud_info,omitempty"`
	MongoInfo *map[string]interface{} `json:"mongo_info,omitempty"`
	HostInfo  *PoolHostsInfo          `json:"host_info,omitempty"`
	Hosts     *struct {
		ConfiguredCount int64 `json:"configured_count"`
		Count           int64 `json:"count"`
	} `json:"hosts,omitempty"`
	// TODO PoolInfo struct is partial ...
}

// NamespaceResourceInfo is a struct of namespace resource info returned by the API
type NamespaceResourceInfo struct {
	Name         string              `json:"name"`
	Mode         string              `json:"mode,omitempty"`
	Undeletable  string              `json:"undeletable,omitempty"`
	EndpointType EndpointType        `json:"endpoint_type,omitempty"`
	Endpoint     string              `json:"endpoint,omitempty"`
	TargetBucket string              `json:"target_bucket,omitempty"`
	AccessMode   nbv1.AccessModeType `json:"access_mode"`
	Identity     string              `json:"identity,omitempty"`
	AuthMethod   CloudAuthMethod     `json:"auth_method,omitempty"`
	CpCode       string              `json:"cp_code,omitempty"`
}

// NamespaceResourceOperatorInfo is a struct of namespace resource secrets returned by the API
type NamespaceResourceOperatorInfo struct {
	AccessKey   string `json:"access_key,omitempty"`
	SecretKey   string `json:"secret_key,omitempty"`
	NeedK8sSync bool   `json:"need_k8s_sync,omitempty"`
}

// PoolHostsInfo is the config/info of a hosts pool
type PoolHostsInfo struct {
	// TODO encode/decode BigInt
	VolumeSize int64 `json:"volume_size"`
}

// S3AccessKeys is a struct holding S3 access and secret keys
type S3AccessKeys struct {
	AccessKey MaskedString `json:"access_key"`
	SecretKey MaskedString `json:"secret_key"`
}

// ReadAuthReply is the reply of auth_api.read_auth()
type ReadAuthReply struct {
	Account struct {
		Name               string `json:"name"`
		Email              string `json:"email"`
		IsSupport          bool   `json:"is_support"`
		MustChangePassword bool   `json:"must_change_password"`
	} `json:"account"`
	System struct {
		Name string `json:"name"`
	} `json:"system"`
	AuthorizedBy string                 `json:"authorized_by"`
	Role         string                 `json:"role"`
	Extra        map[string]interface{} `json:"extra"`
}

// ReadAccountParams is the params to account_api.read_account()
type ReadAccountParams struct {
	Email string `json:"email"`
}

// ReadBucketParams is the params to bucket_api.read_bucket()
type ReadBucketParams struct {
	Name string `json:"name"`
}

// ReadPoolParams is the params to pool_api.read_pool()
type ReadPoolParams struct {
	Name string `json:"name"`
}

// ListAccountsParams is the params to account_api.list_accounts()
type ListAccountsParams struct {
	Filter struct {
		FsIdentity struct {
			UID int `json:"uid"`
			GID int `json:"gid"`
		} `json:"fs_identity"`
	} `json:"filter"`
}

// ReadNamespaceResourceParams is the params to pool_api.read_namespace_resource()
type ReadNamespaceResourceParams struct {
	Name string `json:"name"`
}

// ReadySystemStatusReply is the reply to system_pai.get_system_status()
type ReadySystemStatusReply struct {
	State string `json:"state,omitempty"`
}

// ListAccountsReply is the reply to account_api.list_accounts()
type ListAccountsReply struct {
	Accounts []*AccountInfo `json:"accounts"`
}

// ListBcuketsParams is the params to account_api.list_buckets()
type ListBucketsParams struct {
	ContinuationToken *string `json:"continuation_token,omitempty"`
	MaxBuckets *int `json:"max_buckets,omitempty"`
}

// ListBucketsReply is the reply of bucket_api.list_buckets()
type ListBucketsReply struct {
	Buckets []struct {
		Name string `json:"name"`
	} `json:"buckets"`
}

// ListHostsParams is the params to host_api.list_hosts()
type ListHostsParams struct {
	Query ListHostsQuery `json:"query"`
}

// ListHostsQuery is the query params to host_api.list_hosts()
type ListHostsQuery struct {
	Pools []string `json:"pools"`
}

// ListHostsReply is the reply of host_api.list_hosts()
type ListHostsReply struct {
	Hosts []HostInfo `json:"hosts"`
}

// HostInfo is the information of a host(partial)
type HostInfo struct {
	Name string `json:"name"`
}

// CreateAuthParams is the params of auth_api.create_auth()
type CreateAuthParams struct {
	System   string `json:"system"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// CreateAuthReply is the reply of auth_api.create_auth()
type CreateAuthReply struct {
	Token string `json:"token"`
}

// CreateSystemParams is the params of system_api.create_system()
type CreateSystemParams struct {
	Name     string       `json:"name"`
	Email    string       `json:"email"`
	Password MaskedString `json:"password"`
}

// CreateSystemReply is the reply of system_api.create_system()
type CreateSystemReply struct {
	Token         string `json:"token"`
	OperatorToken string `json:"operator_token"`
}

// CreateBucketParams is the params of bucket_api.create_bucket()
type CreateBucketParams struct {
	Name         string               `json:"name"`
	Tiering      string               `json:"tiering,omitempty"`
	ForceMd5Etag *bool                `json:"force_md5_etag,omitempty"`
	BucketClaim  *BucketClaimInfo     `json:"bucket_claim,omitempty"`
	Namespace    *NamespaceBucketInfo `json:"namespace,omitempty"`
	Quota        *QuotaConfig         `json:"quota,omitempty"`
}

// QuotaConfig quota configuration
type QuotaConfig struct {
	Size     *SizeQuotaConfig     `json:"size,omitempty"`
	Quantity *QuantityQuotaConfig `json:"quantity,omitempty"`
}

// SizeQuotaConfig size quota configuration
type SizeQuotaConfig struct {
	//limits the max total size value
	Value float64 `json:"value,omitempty"`
	//Units of max total size per bucket
	Unit string `json:"unit,omitempty"`
}

// QuantityQuotaConfig quantity quota configuration
type QuantityQuotaConfig struct {
	//limits the max total quantity value
	Value int `json:"value,omitempty"`
}

// IsEqual return true if quotas are equal
func (q *QuotaConfig) IsEqual(q2 *QuotaConfig) bool {
	if q == nil && q2 == nil {
		return true
	}
	if (q == nil && q2.Size == nil && q2.Quantity == nil) ||
		(q2 == nil && q.Size == nil && q.Quantity == nil) {
		return true
	}
	return reflect.DeepEqual(q, q2)
}

// NamespaceBucketInfo is the information needed for creating namespace bucket
type NamespaceBucketInfo struct {
	WriteResource NamespaceResourceFullConfig   `json:"write_resource,omitempty"`
	ReadResources []NamespaceResourceFullConfig `json:"read_resources,omitempty"`
	Caching       *CacheSpec                    `json:"caching,omitempty"`
}

// NamespaceResourceFullConfig is the resource configuration for creating namespace bucket
type NamespaceResourceFullConfig struct {
	Resource string `json:"resource"`
	Path     string `json:"path,omitempty"`
}

// CacheSpec specifies the cache specifications for the bucket class
type CacheSpec struct {
	// TTL specifies the cache ttl
	TTLMs int `json:"ttl_ms,omitempty"`
}

// BucketClaimInfo is the params of bucket_api.create_bucket()
type BucketClaimInfo struct {
	BucketClass string `json:"bucket_class,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// CreateAccountParams is the params of account_api.create_account()
type CreateAccountParams struct {
	Name              string                  `json:"name"`
	Email             string                  `json:"email"`
	HasLogin          bool                    `json:"has_login"`
	S3Access          bool                    `json:"s3_access"`
	AllowBucketCreate bool                    `json:"allow_bucket_creation"`
	DefaultResource   string                  `json:"default_resource,omitempty"`
	ForceMd5Etag      *bool                   `json:"force_md5_etag,omitempty"`
	BucketClaimOwner  string                  `json:"bucket_claim_owner,omitempty"`
	NsfsAccountConfig *nbv1.AccountNsfsConfig `json:"nsfs_account_config,omitempty"`
}

// CreateAccountReply is the reply of account_api.create_account()
type CreateAccountReply struct {
	Token      string         `json:"token"`
	AccessKeys []S3AccessKeys `json:"access_keys"`
}

// GenerateAccountKeysParams is the params of account_api.generate_account_keys()
type GenerateAccountKeysParams struct {
	Email string `json:"email"`
}

// UpdateAccountKeysParams is the params of account_api.update_account_keys()
type UpdateAccountKeysParams struct {
	Email      string       `json:"email"`
	AccessKeys S3AccessKeys `json:"access_keys"`
}

// ResetPasswordParams is the params of account_api.reset_password()
type ResetPasswordParams struct {
	Email                string       `json:"email"`
	VerificationPassword MaskedString `json:"verification_password"`
	Password             MaskedString `json:"password"`
}

// BackingStoreInfo describes backingstore info
type BackingStoreInfo struct {
	// Name describes backingstore name
	Name string `json:"name"`
	// Namespace describes backingstore namespace
	Namespace string `json:"namespace"`
}

// NamespaceStoreInfo describes namespacestore info
type NamespaceStoreInfo struct {
	// Name describes backingstore name
	Name string `json:"name"`
	// Namespace describes backingstore namespace
	Namespace string `json:"namespace"`
}

// CreateHostsPoolParams is the params of pool_api.create_hosts_pool()
type CreateHostsPoolParams struct {
	Name         string            `json:"name"`
	IsManaged    bool              `json:"is_managed"`
	HostCount    int               `json:"host_count"`
	HostConfig   PoolHostsInfo     `json:"host_config"`
	Backingstore *BackingStoreInfo `json:"backingstore,omitempty"`
}

// GetHostsPoolAgentConfigParams is the params of pool_api.get_hosts_pool_agent_config()
type GetHostsPoolAgentConfigParams struct {
	Name         string            `json:"name"`
	Backingstore *BackingStoreInfo `json:"backingstore,omitempty"`
}

// UpdateHostsPoolParams is the params of pool_api.update_hosts_pool()
type UpdateHostsPoolParams struct {
	Name         string            `json:"name"`
	Backingstore *BackingStoreInfo `json:"backingstore,omitempty"`
}

// CreateCloudPoolParams is the params of pool_api.create_cloud_pool()
type CreateCloudPoolParams struct {
	Name              string            `json:"name"`
	Connection        string            `json:"connection"`
	TargetBucket      string            `json:"target_bucket"`
	Backingstore      *BackingStoreInfo `json:"backingstore,omitempty"`
	AvailableCapacity *BigInt           `json:"available_capacity,omitempty"`
}

// UpdateCloudPoolParams is the params of pool_api.create_cloud_pool()
type UpdateCloudPoolParams struct {
	Name              string  `json:"name"`
	AvailableCapacity *BigInt `json:"available_capacity,omitempty"`
}

// CreateNamespaceResourceParams is the params of pool_api.create_cloud_pool()
type CreateNamespaceResourceParams struct {
	Name           string              `json:"name"`
	Connection     string              `json:"connection"`
	TargetBucket   string              `json:"target_bucket"`
	AccessMode     APIAccessModeType   `json:"access_mode"`
	NSFSConfig     *NSFSConfig         `json:"nsfs_config,omitempty"`
	NamespaceStore *NamespaceStoreInfo `json:"namespace_store,omitempty"`
}

// APIAccessModeType is the type of all the optional access modes
type APIAccessModeType = string

const (
	// APIAccessModeReadWrite is the read-write access mode
	APIAccessModeReadWrite APIAccessModeType = "READ_WRITE"
	// APIAccessModeReadOnly is the read-only access mode
	APIAccessModeReadOnly APIAccessModeType = "READ_ONLY"
)

// NSFSConfig is the namespace fs config needed for creating namespace resource of type fs()
type NSFSConfig struct {
	FsBackend  string `json:"fs_backend,omitempty"`
	FsRootPath string `json:"fs_root_path,omitempty"`
}

// CreateTierParams is the reply of tier_api.create_tier()
type CreateTierParams struct {
	Name             string            `json:"name"`
	DataPlacement    string            `json:"data_placement,omitempty"`
	AttachedPools    []string          `json:"attached_pools,omitempty"`
	ChunkCoderConfig *ChunkCoderConfig `json:"chunk_coder_config,omitempty"`
}

// ChunkCoderConfig defines a storage coding configuration
type ChunkCoderConfig struct {
	DigestType     *string `json:"digest_type,omitempty"`
	FragDigestType *string `json:"frag_digest_type,omitempty"`
	CompressType   *string `json:"compress_type,omitempty"`
	CipherType     *string `json:"cipher_type,omitempty"`
	// Data Copies:
	Replicas *int64 `json:"replicas,omitempty"`
	// Erasure Coding:
	DataFrags   *int64  `json:"data_frags,omitempty"`
	ParityFrags *int64  `json:"parity_frags,omitempty"`
	ParityType  *string `json:"parity_type,omitempty"`
	// LRC:
	LrcGroup *int64  `json:"lrc_group,omitempty"`
	LrcFrags *int64  `json:"lrc_frags,omitempty"`
	LrcType  *string `json:"lrc_type,omitempty"`
}

// ChunkSplitConfig defines a storage chunking (splitting objects) configuration
type ChunkSplitConfig struct {
	AvgChunk   int64 `json:"avg_chunk"`
	DeltaChunk int64 `json:"delta_chunk"`
}

// TierItem is an item in a tiering policy
type TierItem struct {
	Order int64  `json:"order"`
	Tier  string `json:"tier"`
	Mode  string `json:"mode,omitempty"`
}

// DeleteBucketParams is the params of bucket_api.delete_bucket()
type DeleteBucketParams struct {
	Name string `json:"name"`
}

// DeleteAccountParams is the params of account_api.delete_account()
type DeleteAccountParams struct {
	Email string `json:"email"`
}

// DeletePoolParams is the params of pool_api.delete_pool()
type DeletePoolParams struct {
	Name string `json:"name"`
}

// DeleteNamespaceResourceParams is the params of pool_api.delete_namespace_resource()
type DeleteNamespaceResourceParams struct {
	Name string `json:"name"`
}

// UpdateAccountParams is the params of account_api.update_account_s3_access()
type UpdateAccountParams struct {
	Name                *string                  `json:"username,omitempty"`
	Email               string                   `json:"email"`
	NewEmail            *string                  `json:"new_email,omitempty"`
	AllowedIPs          *[]struct {
		Start   string `json:"start"`
		End     string `json:"end"`
	} `json:"ips,omitempty"`
	RoleConfig          interface{}              `json:"role_config,omitempty"`
	RemoveRoleConfig    bool                     `json:"remove_role_config,omitempty"`
}

// UpdateAccountS3AccessParams is the params of account_api.update_account_s3_access()
type UpdateAccountS3AccessParams struct {
	Email               string                  `json:"email"`
	S3Access            bool                    `json:"s3_access"`
	DefaultResource     *string                 `json:"default_resource,omitempty"`
	ForceMd5Etag        *bool                   `json:"force_md5_etag,omitempty"`
	AllowBucketCreation *bool                   `json:"allow_bucket_creation,omitempty"`
	NsfsAccountConfig   *nbv1.AccountNsfsConfig `json:"nsfs_account_config,omitempty"`
}

// UpdateDefaultResourceParams is the params of bucket_api.update_all_buckets_default_pool()
type UpdateDefaultResourceParams struct {
	PoolName string `json:"pool_name"`
}

// UpdateBucketClassParams is the params of tiering_policy_api.update_bucket_class()
type UpdateBucketClassParams struct {
	Name   string            `json:"name"`
	Policy TieringPolicyInfo `json:"policy"`
	Tiers  []TierInfo        `json:"tiers"`
}

// BucketReplicationParams is the params of bucket_api.put_bucket_replication()
type BucketReplicationParams struct {
	Name              string            `json:"name"`
	ReplicationPolicy ReplicationPolicy `json:"replication_policy"`
}

// ReplicationPolicy is the struct representing replication configuration
type ReplicationPolicy struct {
	Rules              []interface{} `json:"rules,omitempty"`
	LogReplicationInfo interface{}   `json:"log_replication_info,omitempty"`
}

// DeleteBucketReplicationParams is the params of bucket_api.delete_bucket_replication()
type DeleteBucketReplicationParams struct {
	Name string `json:"name"`
}

// BucketClassInfo is the is the reply of tiering_policy_api.update_bucket_class()
type BucketClassInfo struct {
	ErrorMessage   string                  `json:"error_message"`
	ShouldRevert   bool                    `json:"should_revert"`
	RevertToPolicy UpdateBucketClassParams `json:"revert_to_policy"`
}

// CloudAuthMethod is an enum
type CloudAuthMethod string

// EndpointType is an enum
type EndpointType string

// ExternalConnectionStatus is an enum
type ExternalConnectionStatus string

const (
	// CloudAuthMethodAwsV2 enum
	CloudAuthMethodAwsV2 CloudAuthMethod = "AWS_V2"
	// CloudAuthMethodAwsV4 enum
	CloudAuthMethodAwsV4 CloudAuthMethod = "AWS_V4"

	// EndpointTypeAws enum
	EndpointTypeAws EndpointType = "AWS"
	// EndpointTypeAwsSTS enum
	EndpointTypeAwsSTS EndpointType = "AWSSTS"
	// EndpointTypeAzure enum
	EndpointTypeAzure EndpointType = "AZURE"
	// EndpointTypeGoogle enum
	EndpointTypeGoogle EndpointType = "GOOGLE"
	// EndpointTypeS3Compat enum
	EndpointTypeS3Compat EndpointType = "S3_COMPATIBLE"
	// EndpointTypeIBMCos enum
	EndpointTypeIBMCos EndpointType = "IBM_COS"

	// ExternalConnectionSuccess enum
	ExternalConnectionSuccess ExternalConnectionStatus = "SUCCESS"
	// ExternalConnectionTimeout enum
	ExternalConnectionTimeout ExternalConnectionStatus = "TIMEOUT"
	// ExternalConnectionInvalidEndpoint enum
	ExternalConnectionInvalidEndpoint ExternalConnectionStatus = "INVALID_ENDPOINT"
	// ExternalConnectionInvalidCredentials enum
	ExternalConnectionInvalidCredentials ExternalConnectionStatus = "INVALID_CREDENTIALS"
	// ExternalConnectionNotSupported enum
	ExternalConnectionNotSupported ExternalConnectionStatus = "NOT_SUPPORTED"
	// ExternalConnectionTimeSkew enum
	ExternalConnectionTimeSkew ExternalConnectionStatus = "TIME_SKEW"
	// ExternalConnectionUnknownFailure enum
	ExternalConnectionUnknownFailure ExternalConnectionStatus = "UNKNOWN_FAILURE"
)

// ExternalConnectionInfo is a struct for reply with connection info
type ExternalConnectionInfo struct {
	Name         string          `json:"name"`
	EndpointType EndpointType    `json:"endpoint_type"`
	Endpoint     string          `json:"endpoint"`
	Identity     string          `json:"identity"`
	AuthMethod   CloudAuthMethod `json:"auth_method,omitempty"`
	Usage        []struct {
		UsageType      string `json:"usage_type"`
		Entity         string `json:"entity"`
		ExternalEntity string `json:"external_entity"`
	} `json:"usage"`
}

// AzureLogAccessKeysParams is used to map between the Azure secret CR fields and the API ones
type AzureLogAccessKeysParams struct {
	AzureTenantID                 string `json:"azure_tenant_id"`
	AzureClientID                 string `json:"azure_client_id"`
	AzureClientSecret             string `json:"azure_client_secret"`
	AzureLogsAnalyticsWorkspaceID string `json:"azure_logs_analytics_workspace_id"`
}

// AddExternalConnectionParams is the params of account_api.add_external_connection()
type AddExternalConnectionParams struct {
	Name               string                    `json:"name"`
	EndpointType       EndpointType              `json:"endpoint_type"`
	Endpoint           string                    `json:"endpoint"`
	Identity           MaskedString              `json:"identity"`
	Secret             MaskedString              `json:"secret"`
	AuthMethod         CloudAuthMethod           `json:"auth_method,omitempty"`
	AWSSTSARN          string                    `json:"aws_sts_arn,omitempty"`
	Region             string                    `json:"region,omitempty"`
	AzureLogAccessKeys *AzureLogAccessKeysParams `json:"azure_log_access_keys,omitempty"`
}

// CheckExternalConnectionParams is the params of account_api.check_external_connection()
type CheckExternalConnectionParams struct {
	Name                   string                    `json:"name"`
	EndpointType           EndpointType              `json:"endpoint_type"`
	Endpoint               string                    `json:"endpoint"`
	Identity               MaskedString              `json:"identity"`
	Secret                 MaskedString              `json:"secret"`
	AuthMethod             CloudAuthMethod           `json:"auth_method,omitempty"`
	AWSSTSARN              string                    `json:"aws_sts_arn,omitempty"`
	IgnoreNameAlreadyExist bool                      `json:"ignore_name_already_exist,omitempty"`
	AzureLogAccessKeys     *AzureLogAccessKeysParams `json:"azure_log_access_keys,omitempty"`
	Region                 string                    `json:"region,omitempty"`
}

// CheckExternalConnectionReply is the reply of account_api.check_external_connection()
type CheckExternalConnectionReply struct {
	Status ExternalConnectionStatus `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// UpdateExternalConnectionParams is the params of account_api.update_external_connection()
type UpdateExternalConnectionParams struct {
	Name               string                    `json:"name"`
	Identity           MaskedString              `json:"identity"`
	Secret             MaskedString              `json:"secret"`
	AzureLogAccessKeys *AzureLogAccessKeysParams `json:"azure_log_access_keys,omitempty"`
	Region             string                    `json:"region,omitempty"`
}

// DeleteExternalConnectionParams is the params of account_api.delete_external_connection()
type DeleteExternalConnectionParams struct {
	Name string `json:"connection_name"`
}

// IntRange Hold a min/max integer range
type IntRange struct {
	Min int32 `json:"min"`
	Max int32 `json:"max"`
}

// UpdateEndpointGroupParams is the params of system_api.update_endpoint_group()
type UpdateEndpointGroupParams struct {
	GroupName     string   `json:"group_name"`
	IsRemote      bool     `json:"is_remote"`
	Region        string   `json:"region"`
	EndpointRange IntRange `json:"endpoint_range"`
}

// SetDebugLevelParams - params for debug_api.set_debug_level
type SetDebugLevelParams struct {
	Module string `json:"module"`
	Level  int    `json:"level"`
}

// PublishToClusterParams are the parmas for redirector_api.publish_to_cluster
type PublishToClusterParams struct {
	Target        string      `json:"target"`
	MethodAPI     string      `json:"method_api"`
	MethodName    string      `json:"method_name"`
	RequestParams interface{} `json:"request_params"`
}

// BigIntToHumanBytes returns a human readable bytes string
func BigIntToHumanBytes(bi *BigInt) string {
	return IntToHumanBytes(bi.N + (bi.Peta * petaInBytes))
}

// IntToHumanBytes returns a human readable bytes string
func IntToHumanBytes(bi int64) string {
	f, u := GetBytesAndUnits(bi, -1)
	s := ""
	if f < 0 {
		s = "-"
		f = -f
	}
	return s + strconv.FormatFloat(f, 'f', 3, 64) + " " + u + "B"
}

// GetBytesAndUnits returns bytes and unit
// The precision prec controls the number of digits after the decimal point
func GetBytesAndUnits(bi int64, prec int) (float64, string) {
	if bi == 0 {
		return 0, ""
	}
	units := []string{"", "K", "M", "G", "T", "P", "E", "Z", "Y"}
	f := float64(bi)
	u := 0
	for f >= 1024 || f <= -1024 {
		f /= 1024
		u++
	}

	if prec >= 0 {
		numDigitsPow := math.Pow(10, float64(prec))
		f = math.Floor(f*numDigitsPow) / numDigitsPow
	}

	return f, units[u]
}

// UInt64ToBigInt convert uint64 based value to BigInt value
func UInt64ToBigInt(value uint64) BigInt {
	return BigInt{
		Peta: int64(value / petaInBytes),
		N:    int64(value % petaInBytes),
	}
}
