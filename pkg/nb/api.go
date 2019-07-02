// Package nb makes client API calls to noobaa servers.
package nb

// Client is the interface providing typed noobaa API calls
type Client interface {
	SetAuthToken(token string)
	CreateSystemAPI(CreateSystemParams) (CreateSystemReply, error)
	ListAccountsAPI() (ListAccountsReply, error)
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
