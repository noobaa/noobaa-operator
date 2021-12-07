package nb

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

// Websocket timeouts
const (
	pingInterval   = 5 * time.Second
	reconnectDelay = 3 * time.Second
	connectTimeout = time.Minute
	pongTimeout    = time.Minute
)

// We're reusing existing connections, stored in connMap
// Upon disconnect store reconnect deadline in reconnMap
var (
	connMap   = make(map[string]*RPCConnWS)
	reconnMap = make(map[string]*time.Time)
	connMapLock = sync.Mutex{}
)


// RPCConnWS is an websocket connection which is shared and multiplexed
// for all concurrent requests to the same address
type RPCConnWS struct {
	RPC             *RPC
	Address         string
	connected       bool
	WS              *websocket.Conn
	PendingRequests map[string]*RPCPendingRequest
	NextRequestID   uint64
	Lock            sync.Mutex
	cancelPings     context.CancelFunc
	cancelSend      context.CancelFunc
	CallChan        chan *RPCPendingRequest
}

// RPCPendingRequest is a struct that describes the fields related to an rpc pending requests
type RPCPendingRequest struct {
	Conn      *RPCConnWS
	Req       *RPCMessage
	Res       RPCResponse
	ReplyChan chan error
}

// NewRPCConnWS returns a connected websocket connection
// or nil if connection refused or within reconnect delay
func NewRPCConnWS(r *RPC, address string) RPCConn {
	connMapLock.Lock()
	defer connMapLock.Unlock()

	// If there is an existing connection to the address, then use it
	conn := connMap[address]
	if conn != nil {
		return conn
	}

	// Handle reconnect delay
	reconn := reconnMap[address]
	if reconn != nil && reconn.After(time.Now()) {
		logrus.Infof("RPC: no connection to %v, reconnect in %v at %v", address, time.Until(*reconn), reconn)
		return nil
	}

	// Create a new connection
	c := &RPCConnWS{
		RPC:             r,
		Address:         address,
		connected:       false,
		PendingRequests: map[string]*RPCPendingRequest{},
		Lock:            sync.Mutex{},
		CallChan:        make(chan *RPCPendingRequest, 1),
	}

	// Establish a connection to the server
	c.WS = c.connect()
	if c.WS == nil {
		reconn := time.Now().Add(reconnectDelay)
		reconnMap[c.Address] = &reconn
		return nil // connection refused
	}

	// reuse this connection for next calls
	connMap[address] = c

	return c
}

// GetAddress returns the connection address
func (c *RPCConnWS) GetAddress() string {
	return c.Address
}

// Call calls an API method to noobaa over wss
func (c *RPCConnWS) Call(req *RPCMessage, res RPCResponse) error {
	replyChan, err := c.NewRequest(req, res)
	if err != nil {
		return err
	}
	return <-replyChan
}

// connect establishes connection to the remote server and
// starts read/write/ping goroutines
func (c *RPCConnWS) connect() *websocket.Conn {
	logrus.Infof("RPC: Connecting websocket to %v", c.GetAddress())
	dialCtx, dialCancel := context.WithTimeout(context.Background(), connectTimeout) 
	defer dialCancel()
	ws, _, err := websocket.Dial(dialCtx, c.Address, &websocket.DialOptions{HTTPClient: &c.RPC.HTTPClient})
	if err != nil {
		logrus.Errorf("RPC: websocket dial error: %v", err)
		return nil
	}

	ws.SetReadLimit(RPCMaxMessageSize)
	c.WS = ws
	c.connected = true

	// Start read/write/ping goroutines
	go c.ReadMessages()

	pingCtx, pingCancel := context.WithCancel(context.Background())
	go c.SendPings(pingCtx)
	c.cancelPings = pingCancel

	sendCtx, sendCancel := context.WithCancel(context.Background())
	go c.SendMessages(sendCtx)
	c.cancelSend = sendCancel

	logrus.Infof("RPC: Connected websocket (%p) %+v", c, c)
	return ws
}

// ping sends a single ping request to the remote server
func (c *RPCConnWS) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pongTimeout)
	defer cancel()
	return c.WS.Ping(ctx)
}

// SendMessages is a go routine running while a connection is open
// it pumps request messages from the Call channel into the WE connection
func (c *RPCConnWS) SendMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case pending := <-c.CallChan:
			if err := c.SendMessage(pending.Req); err != nil {
				logrus.Errorf("RPC: SendMessage error: %v", err)
				c.Close()
			}
		}
	}
}

// SendPings sends pings to improve detection of server disconnect
// https://github.com/nhooyr/websocket/issues/265
func (c *RPCConnWS) SendPings(ctx context.Context) {
	t := time.NewTimer(pingInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		if err := c.ping(); err != nil {
			c.Close()
			return
		}
		t.Reset(pingInterval)
	}
}

// Close releases the connection resources
// terminates goroutines: ping and send
// gracefully closes the underlying WS connection
// once WS connection is closed, read messages goroutine exits
// pending requests are notified with error
func (c *RPCConnWS) Close() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if !c.connected {
		return
	}
	c.connected = false

	logrus.Errorf("RPC: closing connection (%p) %+v", c, c)

	// stop ping go routine
	if c.cancelPings != nil {
		c.cancelPings()
		c.cancelPings = nil
	}

	// stop send go routine
	if c.cancelSend != nil {
		c.cancelSend()
		c.cancelSend = nil
	}

	// close the websocket, will close also the read goroutine
	if c.WS != nil {
		err := c.WS.Close(
			websocket.StatusNormalClosure,
			fmt.Sprintf("RPC: close conn %s", c.Address),
		)

		// Close errors could be ignored
		if err != nil {
			logrus.Warnf("RPC: could not close web socket %s - %v", c.Address, err)
		}
		c.WS = nil
	}

	// wakeup pending waiters with error
	for reqid := range c.PendingRequests {
		pending := c.PendingRequests[reqid]
		pending.ReplyChan <- fmt.Errorf("RPC: connection closed while request is pending %s %s", c.Address, reqid)
	}

	// delete the connection from the connection map
	// and set reconnect time
	reconn := time.Now().Add(reconnectDelay)
	connMapLock.Lock()
	defer connMapLock.Unlock()

	reconnMap[c.Address] = &reconn
	delete(connMap, c.Address)
}

// getPending returns a corresponding RPCPendingRequest
// and removes this request from pending requests map
func (c *RPCConnWS) getPending(msg *RPCMessage) *RPCPendingRequest {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	pending := c.PendingRequests[msg.RequestID]
	delete(c.PendingRequests, msg.RequestID)
	return pending
}

// setPending inserts the request into pending requests map
// assigns requests ID and increase the counter
// returns false if connection is closed, true otherwise
func (c *RPCConnWS) setPending(pending *RPCPendingRequest) bool {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if !c.connected {
		return false
	}

	pending.Req.RequestID = fmt.Sprintf("%s-%d", c.Address, c.NextRequestID)
	c.NextRequestID++
	c.PendingRequests[pending.Req.RequestID] = pending
	return true
}

// NewRequest initializes the request id and register it on the connection pending requests
// and pushes new pending request into call channel to SendMessages goroutine
func (c *RPCConnWS) NewRequest(req *RPCMessage, res RPCResponse) (chan error, error) {
	pending := &RPCPendingRequest{
		Req:       req,
		Res:       res,
		Conn:      c,
		ReplyChan: make(chan error, 1),
	}

	if !c.setPending(pending) {
		return nil, fmt.Errorf("RPC: setPending failed")
	}

	c.CallChan <- pending
	return pending.ReplyChan, nil
}

// SendMessage sends a single pending request
func (c *RPCConnWS) SendMessage(msg interface{}) error {
	ctx, cancel := context.WithTimeout(context.TODO(), RPCSendTimeout)
	defer cancel()
	writer, err := c.WS.Writer(ctx, websocket.MessageBinary)
	if err != nil {
		return err
	}

	reqBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.BigEndian, RPCVersionNumber)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.BigEndian, uint32(len(reqBytes)))
	if err != nil {
		return err
	}

	_, err = writer.Write(reqBytes)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}


// ReadMessages handles incoming messages
func (c *RPCConnWS) ReadMessages() {
	for {
		msg, err := c.ReadMessage()
		if err != nil {
			logrus.Errorf("RPC: ReadMessages error: %v", err)
			c.Close()
			return
		}

		switch msg.Op {
		case "req":
			c.HandleRequest(msg)
		case "res":
			c.HandleResponse(msg)
		case "ping":
			c.HandlePing(msg)
		case "pong":
			logrus.Infof("RPC pong %#v", msg)
		case "routing_req":
			fallthrough
		case "routing_res":
			fallthrough
		default:
			logrus.Errorf("RPC should handle message op %#v", msg)
		}
	}
}

// ReadMessage handles a message
func (c *RPCConnWS) ReadMessage() (*RPCMessage, error) {
	_, reader, err := c.WS.Reader(context.TODO())
	if err != nil {
		return nil, err
	}

	var rpcVersionNumber uint32
	var bodySize uint32

	if err := binary.Read(reader, binary.BigEndian, &rpcVersionNumber); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &bodySize); err != nil {
		return nil, err
	}
	if rpcVersionNumber != RPCVersionNumber {
		return nil, fmt.Errorf("RPC: mismatch RPC version number expected %d received %d", RPCVersionNumber, rpcVersionNumber)
	}
	if bodySize > RPCMaxMessageSize {
		return nil, fmt.Errorf("RPC: message body too big %d", bodySize)
	}

	msgBytes := make([]byte, bodySize)
	n, err := io.ReadFull(reader, msgBytes)
	if err != nil {
		return nil, err
	}
	if n < int(bodySize) {
		return nil, fmt.Errorf("RPC Error: short read")
	}

	msg := &RPCMessage{}
	if err = json.Unmarshal(msgBytes, &msg); err != nil {
		return nil, err
	}
	msg.RawBytes = msgBytes

	// Read message buffers if any
	if msg.Buffers != nil && len(msg.Buffers) > 0 {
		buffers, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		if buffers != nil {
			msg.SetBuffers(buffers)
		}
	}

	return msg, nil
}

// HandleRequest handles an incoming message of type request
func (c *RPCConnWS) HandleRequest(req *RPCMessage) {

	if c.RPC.Handler == nil {
		logrus.Errorf("RPC request but not handler %#v", req)
		return
	}

	go func() {
		reply, err := c.RPC.Handler(req)
		res := &RPCMessageReply{
			RPCMessage: RPCMessage{
				Op:        "res",
				RequestID: req.RequestID,
				Took:      0,
			},
		}
		if err != nil {
			res.Error = &RPCError{RPCCode: "INTERNAL_ERROR", Message: err.Error()}
		} else {
			res.Reply = reply
		}

		err = c.SendMessage(res)
		if err != nil {
			c.Close()
		}
	}()
}

// HandleResponse handles an incoming message of type response
func (c *RPCConnWS) HandleResponse(msg *RPCMessage) {
	pending := c.getPending(msg)

	if pending == nil {
		logrus.Errorf("RPC: no pending request for %s RequestID %v", c.Address, msg.RequestID)
	} else {
		err := json.Unmarshal(msg.RawBytes, pending.Res)
		pending.ReplyChan <- err
	}
}

// HandlePing handles an incoming message of type ping
func (c *RPCConnWS) HandlePing(msg *RPCMessage) {
	err := c.SendMessage(&RPCMessage{
		Op:        "pong",
		RequestID: msg.RequestID,
		Took:      0,
	})
	if err != nil {
		logrus.Errorf("RPC: got error in HandlePing: %v", err)
	}

}
