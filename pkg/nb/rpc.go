package nb

import (
	"net/http"
	"strings"
	"sync"
	"time"

	util "github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"
)

const (
	// RPCVersionNumber specifies the RPC version number
	RPCVersionNumber uint32 = 0xba000000

	// RPCMaxMessageSize is a limit to protect the process from allocating too much memory
	// for a single incoming message for example in case the connection is out of sync or other bugs.
	RPCMaxMessageSize = 64 * 1024 * 1024

	// RPCSendTimeout is a limit the time we wait for getting reply from the server
	RPCSendTimeout = 120 * time.Second
)

// GlobalRPC is the global rpc
var GlobalRPC *RPC

func init() {
	// GlobalRPC initialization
	GlobalRPC = NewRPC()
}

// RPC is a struct that describes the relevant fields upon handeling rpc protocol
type RPC struct {
	HTTPClient  http.Client
	ConnMap     map[string]RPCConn
	ConnMapLock sync.Mutex
	Handler     RPCHandler
}

// RPCClient makes API calls to noobaa.
// Requests to noobaa are plain http requests with json request and json response.
type RPCClient struct {
	RPC       *RPC
	Router    APIRouter
	AuthToken string
}

// RPCConn is a common connection interface implemented by http and ws
type RPCConn interface {
	// GetAddress returns the connection address
	GetAddress() string
	// Reonnect should make sure the connection is ready to be used
	Reconnect()
	// Call sends request and receives the response
	Call(req *RPCMessage, res RPCResponse) error
}

// RPCMessage structure encoded in every RPC message
type RPCMessage struct {
	Op        string      `json:"op"`
	API       string      `json:"api,omitempty"`
	Method    string      `json:"method,omitempty"`
	RequestID string      `json:"reqid,omitempty"`
	AuthToken string      `json:"auth_token,omitempty"`
	Took      float64     `json:"took,omitempty"`
	Error     *RPCError   `json:"error,omitempty"`
	Params    interface{} `json:"params,omitempty"`
	Buffers   []RPCBuffer `json:"buffers,omitempty"`
	RawBytes  []byte      `json:"-"`
}

// RPCMessageReply structure encoded in every RPC message that contains reply
type RPCMessageReply struct {
	RPCMessage `json:",inline"`
	Reply      interface{} `json:"reply,omitempty"`
}

// RPCBuffer is a struct that describes the fields related to an rpc buffer
type RPCBuffer struct {
	Name   string `json:"name,omitempty"`
	Length int32  `json:"len,omitempty"`
	Buffer []byte `json:"-"`
}

// RPCError is a struct sent by noobaa servers to denote an error response.
type RPCError struct {
	RPCCode string `json:"rpc_code,omitempty"`
	Message string `json:"message"`
}

// RPCHandler is the interface for RPCHandler struct
type RPCHandler func(req *RPCMessage) (interface{}, error)

// RPCResponse is the interface for response structs.
// RPCMessage is the only real implementor of it.
type RPCResponse interface {
	Response() *RPCMessage
}

// SetAuthToken is setting the client token for next calls
func (c *RPCClient) SetAuthToken(token string) { c.AuthToken = token }

// GetAuthToken is getting the client token for next calls
func (c *RPCClient) GetAuthToken() string { return c.AuthToken }

// Error is implementing the standard error type interface
func (e *RPCError) Error() string { return e.Message }

// Response is implementing the RPCResponse interface
func (msg *RPCMessage) Response() *RPCMessage { return msg }

// SetBuffers assigns the buffers from the message and slices them to the message buffers
func (msg *RPCMessage) SetBuffers(buffers []byte) {
	pos := 0
	for i := range msg.Buffers {
		b := &msg.Buffers[i]
		end := pos + int(b.Length)
		b.Buffer = buffers[pos:end]
		pos = end
	}
}

var _ Client = &RPCClient{}
var _ RPCResponse = &RPCMessage{}
var _ error = &RPCError{}

// NewRPC initializes an RPC with defaults
func NewRPC() *RPC {
	return &RPC{
		HTTPClient: http.Client{
			Transport: util.InsecureHTTPTransport,
		},
		ConnMap:     make(map[string]RPCConn),
		ConnMapLock: sync.Mutex{},
	}
}

// NewClient initializes an RPCClient with defaults
func NewClient(router APIRouter) Client {
	return &RPCClient{
		Router: router,
		RPC:    GlobalRPC,
	}
}

// Call an API method to noobaa over wss or https protocol
// The response type should be defined to include RPCResponse inline.
// This is needed in order for json.Unmarshal() to decode into the reply structure.
func (c *RPCClient) Call(req *RPCMessage, res RPCResponse) error {
	if res == nil {
		res = &RPCMessage{}
	}
	api := req.API
	method := req.Method
	if req.AuthToken == "" {
		req.AuthToken = c.AuthToken
	}

	address := c.Router.GetAddress(api)

	// u := address + strings.TrimSuffix(api, "_api") + "/" + method
	u := strings.TrimSuffix(api, "_api") + "." + method + "()"
	logrus.Infof("✈️  RPC: %s Request: %+v", u, req.Params)

	conn := c.RPC.GetConnection(address)
	err := conn.Call(req, res)
	if err != nil {
		logrus.Errorf("⚠️  RPC: %s Call failed: %s", u, err)
		return err
	}

	r := res.Response()
	if r.Error != nil {
		logrus.Errorf("⚠️  RPC: %s Response Error: Code=%s Message=%s", u, r.Error.RPCCode, r.Error.Message)
		return r.Error
	}

	logrus.Infof("✅ RPC: %s Response OK: took %.1fms", u, r.Took)
	return nil
}

// GetConnection finds the connection related to the pending request or creates a new one
func (r *RPC) GetConnection(address string) RPCConn {
	var conn RPCConn
	if strings.HasPrefix(address, "wss:") || strings.HasPrefix(address, "ws:") {
		r.ConnMapLock.Lock()
		conn = r.ConnMap[address]
		if conn == nil {
			conn = NewRPCConnWS(r, address)
			logrus.Warnf("RPC: GetConnection creating connection to %s %p", address, conn)
			r.ConnMap[address] = conn
		}
		r.ConnMapLock.Unlock()
	} else {
		// http connections are transient and not inserted to ConnMap!
		conn = NewRPCConnHTTP(r, address)
	}
	return conn
}

// RemoveConnection removes the connection from the RPC connections map and start reconnecting
func (r *RPC) RemoveConnection(conn RPCConn) {
	r.ConnMapLock.Lock()
	address := conn.GetAddress()
	current := r.ConnMap[address]
	if current == conn {
		logrus.Warnf("RPC: RemoveConnection %s current=%p conn=%p", address, current, conn)
		delete(r.ConnMap, address)
	}
	if current == conn || current == nil {
		go func() {
			r.GetConnection(address).Reconnect()
		}()
	}
	r.ConnMapLock.Unlock()
}
