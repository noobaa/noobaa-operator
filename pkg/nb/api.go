// Package nb makes client API calls to noobaa servers.
package nb

// Client is the interface providing typed noobaa API calls
type Client interface {
	SetAuthToken(token string)
	GetAuthToken() string

	ReadAuthAPI() (ReadAuthReply, error)

	ListAccountsAPI() (ListAccountsReply, error)
	ListBucketsAPI() (ListBucketsReply, error)

	CreateAuthAPI(CreateAuthParams) (CreateAuthReply, error)
	CreateSystemAPI(CreateSystemParams) (CreateSystemReply, error)
	CreateBucketAPI(CreateBucketParams) (CreateBucketReply, error)

	DeleteBucketAPI(DeleteBucketParams) (DeleteBucketReply, error)
}

//////////
// READ //
//////////

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

// CreateAuthAPI calls auth_api.read_auth()
func (c *RPCClient) ReadAuthAPI() (ReadAuthReply, error) {
	req := RPCRequest{API: "auth_api", Method: "read_auth"}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       ReadAuthReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

//////////
// LIST //
//////////

// ListAccountsReply is the reply to account_api.list_accounts()
type ListAccountsReply struct {
	Accounts []struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		AccessKeys []struct {
			AccessKey string `json:"access_key"`
			SecretKey string `json:"secret_key"`
		} `json:"access_keys"`
	} `json:"accounts"`
}

// ListAccountsAPI calls account_api.list_accounts()
func (c *RPCClient) ListAccountsAPI() (ListAccountsReply, error) {
	req := RPCRequest{API: "account_api", Method: "list_accounts"}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       ListAccountsReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

type ListBucketsReply struct {
	Buckets []struct {
		Name string `json:"name"`
	} `json:"buckets"`
}

// ListBucketsAPI calls bucket_api.list_buckets()
func (c *RPCClient) ListBucketsAPI() (ListBucketsReply, error) {
	req := RPCRequest{API: "bucket_api", Method: "list_buckets"}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       ListBucketsReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

////////////
// CREATE //
////////////

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

// CreateAuthAPI calls auth_api.create_auth()
func (c *RPCClient) CreateAuthAPI(params CreateAuthParams) (CreateAuthReply, error) {
	req := RPCRequest{API: "auth_api", Method: "create_auth", Params: params}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       CreateAuthReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

// CreateSystemParams is the params of system_api.create_system()
type CreateSystemParams struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateSystemReply is the reply of system_api.create_system()
type CreateSystemReply struct {
	Token string `json:"token"`
}

// CreateSystemAPI calls system_api.create_system()
func (c *RPCClient) CreateSystemAPI(params CreateSystemParams) (CreateSystemReply, error) {
	req := RPCRequest{API: "system_api", Method: "create_system", Params: params}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       CreateSystemReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

// CreateBucketParams is the params of bucket_api.create_bucket()
type CreateBucketParams struct {
	Name string `json:"name"`
}

// CreateBucketReply is the reply of bucket_api.create_bucket()
type CreateBucketReply struct {
}

// CreateBucketAPI calls bucket_api.create_bucket()
func (c *RPCClient) CreateBucketAPI(params CreateBucketParams) (CreateBucketReply, error) {
	req := RPCRequest{API: "bucket_api", Method: "create_bucket", Params: params}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       CreateBucketReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}

////////////
// DELETE //
////////////

// DeleteBucketParams is the params of bucket_api.delete_bucket()
type DeleteBucketParams struct {
	Name string `json:"name"`
}

// DeleteBucketReply is the reply of bucket_api.delete_bucket()
type DeleteBucketReply struct {
}

// DeleteBucketAPI calls bucket_api.delete_bucket()
func (c *RPCClient) DeleteBucketAPI(params DeleteBucketParams) (DeleteBucketReply, error) {
	req := RPCRequest{API: "bucket_api", Method: "delete_bucket", Params: params}
	res := struct {
		RPCResponse `json:",inline"`
		Reply       DeleteBucketReply `json:"reply"`
	}{}
	err := c.Call(req, &res)
	return res.Reply, err
}
