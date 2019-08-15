// Package nb makes client API calls to noobaa servers.
package nb

// Client is the interface providing typed noobaa API calls
type Client interface {
	SetAuthToken(token string)
	GetAuthToken() string

	ReadAuthAPI() (ReadAuthReply, error)
	ReadAccountAPI(ReadAccountParams) (AccountInfo, error)
	ReadSystemAPI() (SystemInfo, error)

	ListAccountsAPI() (ListAccountsReply, error)
	ListBucketsAPI() (ListBucketsReply, error)

	CreateAuthAPI(CreateAuthParams) (CreateAuthReply, error)
	CreateSystemAPI(CreateSystemParams) (CreateSystemReply, error)
	CreateBucketAPI(CreateBucketParams) (CreateBucketReply, error)
	CreateAccountAPI(CreateAccountParams) (CreateAccountReply, error)
	CreateHostsPoolAPI(CreateHostsPoolParams) error
	CreateCloudPoolAPI(CreateCloudPoolParams) error

	DeleteBucketAPI(DeleteBucketParams) error
	DeleteAccountAPI(DeleteAccountParams) error
	DeletePoolAPI(DeletePoolParams) error

	UpdateAccountS3Access(UpdateAccountS3AccessParams) error

	AddExternalConnectionAPI(AddExternalConnectionParams) error
	CheckExternalConnectionAPI(AddExternalConnectionParams) (CheckExternalConnectionReply, error)
	EditExternalConnectionCredentialsAPI(EditExternalConnectionCredentialsParams) error
	DeleteExternalConnectionAPI(DeleteExternalConnectionParams) error
}

///////////
// TYPES //
///////////

// SystemInfo is a struct of system info returned by the API
type SystemInfo struct {
	Accounts []AccountInfo `json:"accounts"`
	Buckets  []BucketInfo  `json:"buckets"`
	Pools    []PoolInfo    `json:"pools"`
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
	DefaultPool        string         `json:"default_pool"`
	AccessKeys         []S3AccessKeys `json:"access_keys"`
	AllowedIPs         []struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"allowed_ips"`
	ExternalConnections struct {
		Count       int                      `json:"count"`
		Connections []ExternalConnectionInfo `json:"connections"`
	} `json:"external_connections"`
	AllowedBuckets AllowedBuckets `json:"allowed_buckets"`
	Systems        []struct {
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	} `json:"systems"`
	Preferences struct {
		UITheme string `json:"ui_theme"`
	} `json:"preferences"`
}

// BucketInfo is a struct of bucket info returned by the API
type BucketInfo struct {
	Name string `json:"name"`
	// TODO BucketInfo struct is partial ...
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
	// TODO PoolInfo struct is partial ...
}

// PoolHostsInfo is the config/info of a hosts pool
type PoolHostsInfo struct {
	// TODO encode/decode BigInt
	VolumeSize int64 `json:"volume_size"`
}

// S3AccessKeys is a struct holding S3 access and secret keys
type S3AccessKeys struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
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

// ListAccountsReply is the reply to account_api.list_accounts()
type ListAccountsReply struct {
	Accounts []AccountInfo `json:"accounts"`
}

// ListBucketsReply is the reply of bucket_api.list_buckets()
type ListBucketsReply struct {
	Buckets []struct {
		Name string `json:"name"`
	} `json:"buckets"`
}

// CreateAuthParams is the params of auth_api.create_auth()
type CreateAuthParams struct {
	System   string `json:"system"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateAuthReply is the reply of auth_api.create_auth()
type CreateAuthReply struct {
	Token string `json:"token"`
}

// CreateSystemParams is the params of system_api.create_system()
type CreateSystemParams struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateSystemReply is the reply of system_api.create_system()
type CreateSystemReply struct {
	Token         string `json:"token"`
	OperatorToken string `json:"operator_token"`
}

// CreateBucketParams is the params of bucket_api.create_bucket()
type CreateBucketParams struct {
	Name string `json:"name"`
}

// CreateBucketReply is the reply of bucket_api.create_bucket()
type CreateBucketReply struct {
}

// AccountAllowedBuckets is part of CreateAccountParams
type AccountAllowedBuckets struct {
	FullPermission bool     `json:"full_permission"`
	PermissionList []string `json:"permission_list"`
}

// CreateAccountParams is the params of account_api.create_account()
type CreateAccountParams struct {
	Name              string                `json:"name"`
	Email             string                `json:"email"`
	HasLogin          bool                  `json:"has_login"`
	S3Access          bool                  `json:"s3_access"`
	AllowBucketCreate bool                  `json:"allow_bucket_creation"`
	AllowedBuckets    AccountAllowedBuckets `json:"allowed_buckets"`
}

// CreateAccountReply is the reply of account_api.create_account()
type CreateAccountReply struct {
	Token      string         `json:"token"`
	AccessKeys []S3AccessKeys `json:"access_keys"`
}

// CreateHostsPoolParams is the params of pool_api.create_hosts_pool()
type CreateHostsPoolParams struct {
	Name       string        `json:"name"`
	IsManaged  bool          `json:"is_managed"`
	HostCount  int           `json:"host_count"`
	HostConfig PoolHostsInfo `json:"host_config"`
}

// CreateCloudPoolParams is the reply of pool_api.create_cloud_pool()
type CreateCloudPoolParams struct {
	Name         string `json:"name"`
	Connection   string `json:"connection"`
	TargetBucket string `json:"target_bucket"`
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

// UpdateAccountS3AccessParams is the params of account_api.update_account_s3_access()
type UpdateAccountS3AccessParams struct {
	Email               string          `json:"email"`
	S3Access            bool            `json:"s3_access"`
	DefaultPool         *string         `json:"default_pool,omitempty"`
	AllowBucketCreation *bool           `json:"allow_bucket_creation,omitempty"`
	AllowBuckets        *AllowedBuckets `json:"allowed_buckets,omitempty"`
}

// AllowedBuckets is a struct for setting which buckets an account can access
type AllowedBuckets struct {
	FullPermission bool     `json:"full_permission"`
	PermissionList []string `json:"permission_list"`
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
	// EndpointTypeAzure enum
	EndpointTypeAzure EndpointType = "AZURE"
	// EndpointTypeGoogle enum
	EndpointTypeGoogle EndpointType = "GOOGLE"
	// EndpointTypeS3Compat enum
	EndpointTypeS3Compat EndpointType = "S3_COMPATIBLE"

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

// AddExternalConnectionParams is the params of account_api.add_external_connection()
type AddExternalConnectionParams struct {
	Name         string          `json:"name"`
	EndpointType EndpointType    `json:"endpoint_type"`
	Endpoint     string          `json:"endpoint"`
	Identity     string          `json:"identity"`
	Secret       string          `json:"secret"`
	AuthMethod   CloudAuthMethod `json:"auth_method,omitempty"`
}

// CheckExternalConnectionReply is the reply of account_api.check_external_connection()
type CheckExternalConnectionReply struct {
	Status ExternalConnectionStatus `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// EditExternalConnectionCredentialsParams is the params of account_api.edit_external_connection_credentials()
type EditExternalConnectionCredentialsParams struct {
	Name     string `json:"name"`
	Identity string `json:"identity"`
	Secret   string `json:"secret"`
}

// DeleteExternalConnectionParams is the params of account_api.delete_external_connection()
type DeleteExternalConnectionParams struct {
	Name string `json:"connection_name"`
}

//////////
// APIS //
//////////

// ReadAuthAPI calls auth_api.read_auth()
func (c *RPCClient) ReadAuthAPI() (ReadAuthReply, error) {
	req := &RPCRequest{API: "auth_api", Method: "read_auth"}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       ReadAuthReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadAccountAPI calls account_api.read_account()
func (c *RPCClient) ReadAccountAPI(params ReadAccountParams) (AccountInfo, error) {
	req := &RPCRequest{API: "account_api", Method: "read_account", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       AccountInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadSystemAPI calls system_api.read_system()
func (c *RPCClient) ReadSystemAPI() (SystemInfo, error) {
	req := &RPCRequest{API: "system_api", Method: "read_system"}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       SystemInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ListAccountsAPI calls account_api.list_accounts()
func (c *RPCClient) ListAccountsAPI() (ListAccountsReply, error) {
	req := &RPCRequest{API: "account_api", Method: "list_accounts"}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       ListAccountsReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ListBucketsAPI calls bucket_api.list_buckets()
func (c *RPCClient) ListBucketsAPI() (ListBucketsReply, error) {
	req := &RPCRequest{API: "bucket_api", Method: "list_buckets"}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       ListBucketsReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateAuthAPI calls auth_api.create_auth()
func (c *RPCClient) CreateAuthAPI(params CreateAuthParams) (CreateAuthReply, error) {
	req := &RPCRequest{API: "auth_api", Method: "create_auth", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       CreateAuthReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateSystemAPI calls system_api.create_system()
func (c *RPCClient) CreateSystemAPI(params CreateSystemParams) (CreateSystemReply, error) {
	req := &RPCRequest{API: "system_api", Method: "create_system", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       CreateSystemReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateBucketAPI calls bucket_api.create_bucket()
func (c *RPCClient) CreateBucketAPI(params CreateBucketParams) (CreateBucketReply, error) {
	req := &RPCRequest{API: "bucket_api", Method: "create_bucket", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       CreateBucketReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateAccountAPI calls account_api.create_account()
func (c *RPCClient) CreateAccountAPI(params CreateAccountParams) (CreateAccountReply, error) {
	req := &RPCRequest{API: "account_api", Method: "create_account", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       CreateAccountReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateHostsPoolAPI calls pool_api.create_hosts_pool()
func (c *RPCClient) CreateHostsPoolAPI(params CreateHostsPoolParams) error {
	req := &RPCRequest{API: "pool_api", Method: "create_hosts_pool", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// CreateCloudPoolAPI calls pool_api.create_cloud_pool()
func (c *RPCClient) CreateCloudPoolAPI(params CreateCloudPoolParams) error {
	req := &RPCRequest{API: "pool_api", Method: "create_cloud_pool", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeleteBucketAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAPI(params DeleteBucketParams) error {
	req := &RPCRequest{API: "bucket_api", Method: "delete_bucket", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeleteAccountAPI calls account_api.delete_account()
func (c *RPCClient) DeleteAccountAPI(params DeleteAccountParams) error {
	req := &RPCRequest{API: "account_api", Method: "delete_account", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeletePoolAPI calls pool_api.delete_pool()
func (c *RPCClient) DeletePoolAPI(params DeletePoolParams) error {
	req := &RPCRequest{API: "pool_api", Method: "delete_pool", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// UpdateAccountS3Access calls account_api.update_account_s3_access()
func (c *RPCClient) UpdateAccountS3Access(params UpdateAccountS3AccessParams) error {
	req := &RPCRequest{API: "account_api", Method: "update_account_s3_access", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// AddExternalConnectionAPI calls account_api.add_external_connection()
func (c *RPCClient) AddExternalConnectionAPI(params AddExternalConnectionParams) error {
	req := &RPCRequest{API: "account_api", Method: "add_external_connection", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// CheckExternalConnectionAPI calls account_api.check_external_connection()
func (c *RPCClient) CheckExternalConnectionAPI(params AddExternalConnectionParams) (CheckExternalConnectionReply, error) {
	req := &RPCRequest{API: "account_api", Method: "check_external_connection", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       CheckExternalConnectionReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// EditExternalConnectionCredentialsAPI calls account_api.edit_external_connection_credentials()
func (c *RPCClient) EditExternalConnectionCredentialsAPI(params EditExternalConnectionCredentialsParams) error {
	req := &RPCRequest{API: "account_api", Method: "edit_external_connection_credentials", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeleteExternalConnectionAPI calls account_api.delete_external_connection()
func (c *RPCClient) DeleteExternalConnectionAPI(params DeleteExternalConnectionParams) error {
	req := &RPCRequest{API: "account_api", Method: "delete_external_connection", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}
