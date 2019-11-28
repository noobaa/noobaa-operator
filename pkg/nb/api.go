// Package nb makes client API calls to noobaa servers.
package nb

// Client is the interface providing typed noobaa API calls
type Client interface {
	Call(req *RPCRequest, res RPCResponseIfc) error

	SetAuthToken(token string)
	GetAuthToken() string

	ReadAuthAPI() (ReadAuthReply, error)
	ReadAccountAPI(ReadAccountParams) (AccountInfo, error)
	ReadSystemAPI() (SystemInfo, error)
	ReadBucketAPI(ReadBucketParams) (BucketInfo, error)

	ListAccountsAPI() (ListAccountsReply, error)
	ListBucketsAPI() (ListBucketsReply, error)

	CreateAuthAPI(CreateAuthParams) (CreateAuthReply, error)
	CreateSystemAPI(CreateSystemParams) (CreateSystemReply, error)
	CreateAccountAPI(CreateAccountParams) (CreateAccountReply, error)
	CreateBucketAPI(CreateBucketParams) error
	CreateHostsPoolAPI(CreateHostsPoolParams) error
	CreateCloudPoolAPI(CreateCloudPoolParams) error
	CreateTierAPI(CreateTierParams) error
	CreateTieringPolicyAPI(TieringPolicyInfo) error

	DeleteBucketAPI(DeleteBucketParams) error
	DeleteBucketAndObjectsAPI(DeleteBucketParams) error
	DeleteAccountAPI(DeleteAccountParams) error
	DeletePoolAPI(DeletePoolParams) error

	UpdateAccountS3Access(UpdateAccountS3AccessParams) error
	UpdateAllBucketsDefaultPool(UpdateDefaultPoolParams) error

	AddExternalConnectionAPI(AddExternalConnectionParams) error
	CheckExternalConnectionAPI(AddExternalConnectionParams) (CheckExternalConnectionReply, error)
	EditExternalConnectionCredentialsAPI(EditExternalConnectionCredentialsParams) error
	DeleteExternalConnectionAPI(DeleteExternalConnectionParams) error
}

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

// ReadBucketAPI calls bucket_api.read_bucket()
func (c *RPCClient) ReadBucketAPI(params ReadBucketParams) (BucketInfo, error) {
	req := &RPCRequest{API: "bucket_api", Method: "read_bucket", Params: params}
	res := &struct {
		RPCResponse `json:",inline"`
		Reply       BucketInfo `json:"reply"`
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

// CreateBucketAPI calls bucket_api.create_bucket()
func (c *RPCClient) CreateBucketAPI(params CreateBucketParams) error {
	req := &RPCRequest{API: "bucket_api", Method: "create_bucket", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
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

// CreateTierAPI calls tier_api.create_tier()
func (c *RPCClient) CreateTierAPI(params CreateTierParams) error {
	req := &RPCRequest{API: "tier_api", Method: "create_tier", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// CreateTieringPolicyAPI calls tiering_policy_api.create_policy()
func (c *RPCClient) CreateTieringPolicyAPI(params TieringPolicyInfo) error {
	req := &RPCRequest{API: "tiering_policy_api", Method: "create_policy", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeleteBucketAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAPI(params DeleteBucketParams) error {
	req := &RPCRequest{API: "bucket_api", Method: "delete_bucket", Params: params}
	res := &RPCResponse{}
	return c.Call(req, res)
}

// DeleteBucketAndObjectsAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAndObjectsAPI(params DeleteBucketParams) error {
	req := &RPCRequest{API: "bucket_api", Method: "delete_bucket_and_objects", Params: params}
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

// UpdateAllBucketsDefaultPool calls bucket_api.update_all_buckets_default_pool()
func (c *RPCClient) UpdateAllBucketsDefaultPool(params UpdateDefaultPoolParams) error {
	req := &RPCRequest{API: "bucket_api", Method: "update_all_buckets_default_pool", Params: params}
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
