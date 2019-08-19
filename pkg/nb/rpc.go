package nb

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RPCClient makes API calls to noobaa.
// Requests to noobaa are plain http requests with json request and json response.
type RPCClient struct {
	Router     APIRouter
	HTTPClient http.Client
	AuthToken  string
}

// RPCRequest is the structure encoded in every request
type RPCRequest struct {
	API       string      `json:"api"`
	Method    string      `json:"method"`
	AuthToken string      `json:"auth_token,omitempty"`
	Params    interface{} `json:"params,omitempty"`
}

// RPCResponse is the structure encoded in every response
// Specific API response structures should include this inline,
// and add the standard Reply field with the specific fields.
// Refer to examples.
type RPCResponse struct {
	Op        string    `json:"op"`
	RequestID string    `json:"reqid"`
	Took      float64   `json:"took"`
	Error     *RPCError `json:"error,omitempty"`
}

// RPCError is a struct sent by noobaa servers to denote an error response.
type RPCError struct {
	RPCCode string `json:"rpc_code,omitempty"`
	Message string `json:"message"`
}

// RPCResponseIfc is the interface for response structs.
// RPCResponse is the only real implementor of it.
type RPCResponseIfc interface {
	Response() *RPCResponse
}

// SetAuthToken is setting the client token for next calls
func (c *RPCClient) SetAuthToken(token string) { c.AuthToken = token }

// GetAuthToken is getting the client token for next calls
func (c *RPCClient) GetAuthToken() string { return c.AuthToken }

// Response is implementing the RPCResponseIfc interface
func (r *RPCResponse) Response() *RPCResponse { return r }

// Error is implementing the standard error type interface
func (e *RPCError) Error() string { return e.Message }

var _ Client = &RPCClient{}
var _ RPCResponseIfc = &RPCResponse{}
var _ error = &RPCError{}

// NewClient initializes an RPCClient with defaults
func NewClient(router APIRouter) Client {
	return &RPCClient{
		Router: router,
		HTTPClient: http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// Call an API method to noobaa.
// The response type should be defined to include RPCResponseIfc inline.
// This is needed in order for json.Unmarshal() to decode into the reply structure.
func (c *RPCClient) Call(req *RPCRequest, res RPCResponseIfc) error {
	api := req.API
	method := req.Method
	if req.AuthToken == "" {
		req.AuthToken = c.AuthToken
	}
	address := c.Router.GetAddress(api)
	// u := address + strings.TrimSuffix(api, "_api") + "/" + method
	u := strings.TrimSuffix(api, "_api") + "." + method + "()"
	logrus.Infof("✈️  RPC: %s Request: %+v", u, req.Params)

	reqBytes, err := json.Marshal(req)
	fatal(err)

	httpRequest, err := http.NewRequest("PUT", address, bytes.NewReader(reqBytes))
	fatal(err)

	httpResponse, err := c.HTTPClient.Do(httpRequest)
	defer func() {
		if httpResponse != nil && httpResponse.Body != nil {
			httpResponse.Body.Close()
		}
	}()
	if err != nil {
		logrus.Errorf("⚠️  RPC: %s Sending http request failed: %s", u, err)
		return err
	}

	resBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		logrus.Errorf("⚠️  RPC: %s Reading http response failed: %s", u, err)
		return err
	}

	err = json.Unmarshal(resBytes, res)
	if err != nil {
		logrus.Errorf("⚠️  RPC: %s Decoding response failed: %s", u, err)
		return err
	}

	r := res.Response()
	if r.Error != nil {
		logrus.Errorf("⚠️  RPC: %s Response Error: Code=%s Message=%s", u, r.Error.RPCCode, r.Error.Message)
		return r.Error
	}

	logrus.Infof("✅ RPC: %s Response OK: %+v", u, r)
	return nil
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
