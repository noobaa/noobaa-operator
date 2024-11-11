// Package nb makes client API calls to noobaa servers.
package nb

// Client is the interface providing typed noobaa API calls
type Client interface {
	Call(req *RPCMessage, res RPCResponse) error

	SetAuthToken(token string)
	GetAuthToken() string

	ReadAuthAPI() (ReadAuthReply, error)
	ReadAccountAPI(ReadAccountParams) (AccountInfo, error)
	ReadSystemStatusAPI() (ReadySystemStatusReply, error)
	ReadSystemAPI() (SystemInfo, error)
	ReadBucketAPI(ReadBucketParams) (BucketInfo, error)
	ReadPoolAPI(ReadPoolParams) (PoolInfo, error)
	ReadNamespaceResourceAPI(ReadNamespaceResourceParams) (NamespaceResourceInfo, error)
	ReadNamespaceResourceOperatorInfoAPI(ReadNamespaceResourceParams) (NamespaceResourceOperatorInfo, error)
	SetNamespaceStoreInfo(NamespaceStoreInfo) error

	ListAccountsAPI(ListAccountsParams) (ListAccountsReply, error)
	ListBucketsAPI(ListBucketsParams) (ListBucketsReply, error)
	ListHostsAPI(ListHostsParams) (ListHostsReply, error)

	CreateAuthAPI(CreateAuthParams) (CreateAuthReply, error)
	CreateSystemAPI(CreateSystemParams) (CreateSystemReply, error)
	CreateAccountAPI(CreateAccountParams) (CreateAccountReply, error)
	CreateBucketAPI(CreateBucketParams) error
	UpdateBucketAPI(CreateBucketParams) error

	CreateHostsPoolAPI(CreateHostsPoolParams) (string, error)
	GetHostsPoolAgentConfigAPI(GetHostsPoolAgentConfigParams) (string, error)
	UpdateHostsPoolAPI(UpdateHostsPoolParams) error
	CreateCloudPoolAPI(CreateCloudPoolParams) error
	UpdateCloudPoolAPI(UpdateCloudPoolParams) error
	CreateTierAPI(CreateTierParams) error
	CreateNamespaceResourceAPI(CreateNamespaceResourceParams) error
	CreateTieringPolicyAPI(TieringPolicyInfo) error

	DeleteBucketAPI(DeleteBucketParams) error
	DeleteBucketAndObjectsAPI(DeleteBucketParams) error
	DeleteAccountAPI(DeleteAccountParams) error
	DeletePoolAPI(DeletePoolParams) error
	DeleteNamespaceResourceAPI(DeleteNamespaceResourceParams) error

	UpdateAccount(UpdateAccountParams) error
	UpdateAccountS3Access(UpdateAccountS3AccessParams) error
	UpdateAllBucketsDefaultPool(UpdateDefaultResourceParams) error
	UpdateBucketClass(UpdateBucketClassParams) (BucketClassInfo, error)

	AddExternalConnectionAPI(AddExternalConnectionParams) error
	CheckExternalConnectionAPI(CheckExternalConnectionParams) (CheckExternalConnectionReply, error)
	UpdateExternalConnectionAPI(UpdateExternalConnectionParams) error
	DeleteExternalConnectionAPI(DeleteExternalConnectionParams) error

	UpdateEndpointGroupAPI(UpdateEndpointGroupParams) error

	RegisterToCluster() error
	PublishToCluster(PublishToClusterParams) error

	PutBucketReplicationAPI(BucketReplicationParams) error
	GetBucketReplicationAPI(ReadBucketParams) (ReplicationPolicy, error)
	ValidateReplicationAPI(BucketReplicationParams) error
	DeleteBucketReplicationAPI(DeleteBucketReplicationParams) error

	GenerateAccountKeysAPI(GenerateAccountKeysParams) error
	UpdateAccountKeysAPI(UpdateAccountKeysParams) error
	ResetPasswordAPI(ResetPasswordParams) error
}

// ReadAuthAPI calls auth_api.read_auth()
func (c *RPCClient) ReadAuthAPI() (ReadAuthReply, error) {
	req := &RPCMessage{API: "auth_api", Method: "read_auth"}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ReadAuthReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadAccountAPI calls account_api.read_account()
func (c *RPCClient) ReadAccountAPI(params ReadAccountParams) (AccountInfo, error) {
	req := &RPCMessage{API: "account_api", Method: "read_account", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      AccountInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadSystemStatusAPI calls system_api.get_system_status()
func (c *RPCClient) ReadSystemStatusAPI() (ReadySystemStatusReply, error) {
	req := &RPCMessage{API: "system_api", Method: "get_system_status"}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ReadySystemStatusReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadSystemAPI calls system_api.read_system()
func (c *RPCClient) ReadSystemAPI() (SystemInfo, error) {
	req := &RPCMessage{API: "system_api", Method: "read_system"}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      SystemInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadBucketAPI calls bucket_api.read_bucket()
func (c *RPCClient) ReadBucketAPI(params ReadBucketParams) (BucketInfo, error) {
	req := &RPCMessage{API: "bucket_api", Method: "read_bucket", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      BucketInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadPoolAPI calls pool_api.read_pool()
func (c *RPCClient) ReadPoolAPI(params ReadPoolParams) (PoolInfo, error) {
	req := &RPCMessage{API: "pool_api", Method: "read_pool", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      PoolInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ListAccountsAPI calls account_api.list_accounts()
func (c *RPCClient) ListAccountsAPI(params ListAccountsParams) (ListAccountsReply, error) {
	req := &RPCMessage{API: "account_api", Method: "list_accounts", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ListAccountsReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ListBucketsAPI calls bucket_api.list_buckets()
func (c *RPCClient) ListBucketsAPI(params ListBucketsParams) (ListBucketsReply, error) {
	req := &RPCMessage{API: "bucket_api", Method: "list_buckets", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ListBucketsReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ListHostsAPI calls host_api.list_hosts()
func (c *RPCClient) ListHostsAPI(params ListHostsParams) (ListHostsReply, error) {
	req := &RPCMessage{API: "host_api", Method: "list_hosts", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ListHostsReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateAuthAPI calls auth_api.create_auth()
func (c *RPCClient) CreateAuthAPI(params CreateAuthParams) (CreateAuthReply, error) {
	req := &RPCMessage{API: "auth_api", Method: "create_auth", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      CreateAuthReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateSystemAPI calls system_api.create_system()
func (c *RPCClient) CreateSystemAPI(params CreateSystemParams) (CreateSystemReply, error) {
	req := &RPCMessage{API: "system_api", Method: "create_system", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      CreateSystemReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateAccountAPI calls account_api.create_account()
func (c *RPCClient) CreateAccountAPI(params CreateAccountParams) (CreateAccountReply, error) {
	req := &RPCMessage{API: "account_api", Method: "create_account", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      CreateAccountReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// CreateBucketAPI calls bucket_api.create_bucket()
func (c *RPCClient) CreateBucketAPI(params CreateBucketParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "create_bucket", Params: params}
	return c.Call(req, nil)
}

// UpdateBucketAPI calls bucket_api.update_bucket()
func (c *RPCClient) UpdateBucketAPI(params CreateBucketParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "update_bucket", Params: params}
	return c.Call(req, nil)
}

// CreateHostsPoolAPI calls pool_api.create_hosts_pool()
func (c *RPCClient) CreateHostsPoolAPI(params CreateHostsPoolParams) (string, error) {
	req := &RPCMessage{API: "pool_api", Method: "create_hosts_pool", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      string `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// GetHostsPoolAgentConfigAPI calls pool_api.get_hosts_pool_agent_config()
func (c *RPCClient) GetHostsPoolAgentConfigAPI(params GetHostsPoolAgentConfigParams) (string, error) {
	req := &RPCMessage{API: "pool_api", Method: "get_hosts_pool_agent_config", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      string `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// UpdateHostsPoolAPI calls pool_api.scale_hosts_pool()
func (c *RPCClient) UpdateHostsPoolAPI(params UpdateHostsPoolParams) error {
	req := &RPCMessage{API: "pool_api", Method: "update_hosts_pool", Params: params}
	return c.Call(req, nil)
}

// CreateCloudPoolAPI calls pool_api.create_cloud_pool()
func (c *RPCClient) CreateCloudPoolAPI(params CreateCloudPoolParams) error {
	req := &RPCMessage{API: "pool_api", Method: "create_cloud_pool", Params: params}
	return c.Call(req, nil)
}

// UpdateCloudPoolAPI calls pool_api.update_cloud_pool()
func (c *RPCClient) UpdateCloudPoolAPI(params UpdateCloudPoolParams) error {
	req := &RPCMessage{API: "pool_api", Method: "update_cloud_pool", Params: params}
	return c.Call(req, nil)
}

// CreateNamespaceResourceAPI calls pool_api.create_namespace_resource()
func (c *RPCClient) CreateNamespaceResourceAPI(params CreateNamespaceResourceParams) error {
	req := &RPCMessage{API: "pool_api", Method: "create_namespace_resource", Params: params}
	return c.Call(req, nil)
}

// ReadNamespaceResourceAPI calls pool_api.read_namespace_resource()
func (c *RPCClient) ReadNamespaceResourceAPI(params ReadNamespaceResourceParams) (NamespaceResourceInfo, error) {
	req := &RPCMessage{API: "pool_api", Method: "read_namespace_resource", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      NamespaceResourceInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ReadNamespaceResourceOperatorInfoAPI calls pool_api.get_namespace_resource_operator_info()
func (c *RPCClient) ReadNamespaceResourceOperatorInfoAPI(params ReadNamespaceResourceParams) (NamespaceResourceOperatorInfo, error) {
	req := &RPCMessage{API: "pool_api", Method: "get_namespace_resource_operator_info", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      NamespaceResourceOperatorInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// SetNamespaceStoreInfo calls pool_api.set_namespace_store_info()
func (c *RPCClient) SetNamespaceStoreInfo(info NamespaceStoreInfo) error {
	req := &RPCMessage{API: "pool_api", Method: "set_namespace_store_info", Params: info}
	return c.Call(req, nil)
}

// DeleteNamespaceResourceAPI calls pool_api.delete_namespace_resource()
func (c *RPCClient) DeleteNamespaceResourceAPI(params DeleteNamespaceResourceParams) error {
	req := &RPCMessage{API: "pool_api", Method: "delete_namespace_resource", Params: params}
	return c.Call(req, nil)
}

// CreateTierAPI calls tier_api.create_tier()
func (c *RPCClient) CreateTierAPI(params CreateTierParams) error {
	req := &RPCMessage{API: "tier_api", Method: "create_tier", Params: params}
	return c.Call(req, nil)
}

// CreateTieringPolicyAPI calls tiering_policy_api.create_policy()
func (c *RPCClient) CreateTieringPolicyAPI(params TieringPolicyInfo) error {
	req := &RPCMessage{API: "tiering_policy_api", Method: "create_policy", Params: params}
	return c.Call(req, nil)
}

// DeleteBucketAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAPI(params DeleteBucketParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "delete_bucket", Params: params}
	return c.Call(req, nil)
}

// DeleteBucketAndObjectsAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAndObjectsAPI(params DeleteBucketParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "delete_bucket_and_objects", Params: params}
	return c.Call(req, nil)
}

// DeleteAccountAPI calls account_api.delete_account()
func (c *RPCClient) DeleteAccountAPI(params DeleteAccountParams) error {
	req := &RPCMessage{API: "account_api", Method: "delete_account", Params: params}
	return c.Call(req, nil)
}

// DeletePoolAPI calls pool_api.delete_pool()
func (c *RPCClient) DeletePoolAPI(params DeletePoolParams) error {
	req := &RPCMessage{API: "pool_api", Method: "delete_pool", Params: params}
	return c.Call(req, nil)
}

// UpdateAccount calls account_api.update_account()
func (c *RPCClient) UpdateAccount(params UpdateAccountParams) error {
	req := &RPCMessage{API: "account_api", Method: "update_account", Params: params}
	return c.Call(req, nil)
}

// UpdateAccountS3Access calls account_api.update_account_s3_access()
func (c *RPCClient) UpdateAccountS3Access(params UpdateAccountS3AccessParams) error {
	req := &RPCMessage{API: "account_api", Method: "update_account_s3_access", Params: params}
	return c.Call(req, nil)
}

// UpdateBucketClass calls bucket_api.update_bucket_class()
func (c *RPCClient) UpdateBucketClass(params UpdateBucketClassParams) (BucketClassInfo, error) {
	req := &RPCMessage{API: "tiering_policy_api", Method: "update_bucket_class", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      BucketClassInfo `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// UpdateAllBucketsDefaultPool calls bucket_api.update_all_buckets_default_pool()
func (c *RPCClient) UpdateAllBucketsDefaultPool(params UpdateDefaultResourceParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "update_all_buckets_default_pool", Params: params}
	return c.Call(req, nil)
}

// AddExternalConnectionAPI calls account_api.add_external_connection()
func (c *RPCClient) AddExternalConnectionAPI(params AddExternalConnectionParams) error {
	req := &RPCMessage{API: "account_api", Method: "add_external_connection", Params: params}
	return c.Call(req, nil)
}

// CheckExternalConnectionAPI calls account_api.check_external_connection()
func (c *RPCClient) CheckExternalConnectionAPI(params CheckExternalConnectionParams) (CheckExternalConnectionReply, error) {
	req := &RPCMessage{API: "account_api", Method: "check_external_connection", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      CheckExternalConnectionReply `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// UpdateExternalConnectionAPI calls account_api.update_external_connection()
func (c *RPCClient) UpdateExternalConnectionAPI(params UpdateExternalConnectionParams) error {
	req := &RPCMessage{API: "account_api", Method: "update_external_connection", Params: params}
	return c.Call(req, nil)
}

// DeleteExternalConnectionAPI calls account_api.delete_external_connection()
func (c *RPCClient) DeleteExternalConnectionAPI(params DeleteExternalConnectionParams) error {
	req := &RPCMessage{API: "account_api", Method: "delete_external_connection", Params: params}
	return c.Call(req, nil)
}

// UpdateEndpointGroupAPI updates the noobaa core about endpoint configuration changes
func (c *RPCClient) UpdateEndpointGroupAPI(params UpdateEndpointGroupParams) error {
	req := &RPCMessage{API: "system_api", Method: "update_endpoint_group", Params: params}
	return c.Call(req, nil)
}

// RegisterToCluster calls redirector_api.register_to_cluster()
func (c *RPCClient) RegisterToCluster() error {
	req := &RPCMessage{API: "redirector_api", Method: "register_to_cluster"}
	return c.Call(req, nil)
}

// PublishToCluster calls redirector_api.publish_to_cluster()
func (c *RPCClient) PublishToCluster(params PublishToClusterParams) error {
	req := &RPCMessage{API: "redirector_api", Method: "publish_to_cluster", Params: params}
	return c.Call(req, nil)
}

// PutBucketReplicationAPI calls bucket_api.put_bucket_replication()
func (c *RPCClient) PutBucketReplicationAPI(params BucketReplicationParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "put_bucket_replication", Params: params}
	return c.Call(req, nil)
}

// GetBucketReplicationAPI calls bucket_api.get_bucket_replication()
func (c *RPCClient) GetBucketReplicationAPI(params ReadBucketParams) (ReplicationPolicy, error) {
	req := &RPCMessage{API: "bucket_api", Method: "get_bucket_replication", Params: params}
	res := &struct {
		RPCMessage `json:",inline"`
		Reply      ReplicationPolicy `json:"reply"`
	}{}
	err := c.Call(req, res)
	return res.Reply, err
}

// ValidateReplicationAPI calls bucket_api.validate_replication()
func (c *RPCClient) ValidateReplicationAPI(params BucketReplicationParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "validate_replication", Params: params}
	return c.Call(req, nil)
}

// DeleteBucketReplicationAPI calls bucket_api.delete_bucket_replication()
func (c *RPCClient) DeleteBucketReplicationAPI(params DeleteBucketReplicationParams) error {
	req := &RPCMessage{API: "bucket_api", Method: "delete_bucket_replication", Params: params}
	return c.Call(req, nil)
}

// GenerateAccountKeysAPI calls account_api.generate_account_keys()
func (c *RPCClient) GenerateAccountKeysAPI(params GenerateAccountKeysParams) error {
	req := &RPCMessage{API: "account_api", Method: "generate_account_keys", Params: params}
	return c.Call(req, nil)
}

// UpdateAccountKeysAPI calls account_api.update_account_keys()
func (c *RPCClient) UpdateAccountKeysAPI(params UpdateAccountKeysParams) error {
	req := &RPCMessage{API: "account_api", Method: "update_account_keys", Params: params}
	return c.Call(req, nil)
}

// ResetPasswordAPI calls account_api.reset_password()
func (c *RPCClient) ResetPasswordAPI(params ResetPasswordParams) error {
	req := &RPCMessage{API: "account_api", Method: "reset_password", Params: params}
	return c.Call(req, nil)
}
