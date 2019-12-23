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

// RPCConnWS is an websocket connection which is shared and multiplexed
// for all concurrent requests to the same address
type RPCConnWS struct {
	RPC             *RPC
	Address         string
	State           string
	WS              *websocket.Conn
	PendingRequests map[string]*RPCPendingRequest
	NextRequestID   uint64
	Lock            sync.Mutex
	ReconnectDelay  time.Duration
}

// RPCPendingRequest is a struct that describes the fields related to an rpc pending requests
type RPCPendingRequest struct {
	Conn      *RPCConnWS
	Req       *RPCMessage
	Res       RPCResponse
	ReplyChan chan error
}

// NewRPCConnWS returns a new websocket connection
func NewRPCConnWS(r *RPC, address string) *RPCConnWS {
	return &RPCConnWS{
		RPC:             r,
		Address:         address,
		State:           "init",
		PendingRequests: map[string]*RPCPendingRequest{},
		Lock:            sync.Mutex{},
	}
}

// GetAddress returns the connection address
func (c *RPCConnWS) GetAddress() string {
	return c.Address
}

// Reconnect connects after setting a delay
func (c *RPCConnWS) Reconnect() {
	c.Lock.Lock()
	c.ReconnectDelay = 3 * time.Second
	c.ConnectUnderLock()
	c.Lock.Unlock()

}

// Call calls an API method to noobaa over wss
func (c *RPCConnWS) Call(req *RPCMessage, res RPCResponse) error {

	c.Lock.Lock()

	err := c.ConnectUnderLock()
	if err != nil {
		c.Lock.Unlock()
		return err
	}

	replyChan := c.NewRequest(req, res)

	c.Lock.Unlock()

	err = c.SendMessage(req)
	if err != nil {
		return err
	}

	return <-replyChan
}

// ConnectUnderLock is opening a ws connection for new connection or after the previous one closed
// it can delay the reconnect attempts in case of repeated failures
// such as when the host is unreachable, etc.
func (c *RPCConnWS) ConnectUnderLock() error {

	if c.State == "connected" {
		return nil
	}
	if c.State != "init" {
		return fmt.Errorf("RPC: connection already closed %s %p", c.Address, c)
	}

	if c.ReconnectDelay != 0 {
		logrus.Infof("RPC: Reconnect delay %s %s %p", c.ReconnectDelay, c.Address, c)
		time.Sleep(c.ReconnectDelay)
	}

	logrus.Infof("RPC: Connecting websocket %s %p", c.Address, c)
	ws, _, err := websocket.Dial(context.TODO(), c.Address, &websocket.DialOptions{HTTPClient: &c.RPC.HTTPClient})
	if err != nil {
		c.CloseUnderLock()
		return err
	}

	logrus.Infof("RPC: Connected websocket %s %p", c.Address, c)
	ws.SetReadLimit(RPCMaxMessageSize)
	c.WS = ws
	c.State = "connected"
	go c.ReadMessages()

	return nil
}

// Close locks the connection and call close
func (c *RPCConnWS) Close() {
	c.Lock.Lock()
	c.CloseUnderLock()
	c.Lock.Unlock()
}

// CloseUnderLock closes the connection
func (c *RPCConnWS) CloseUnderLock() {
	if c.State == "closed" {
		return
	}
	logrus.Errorf("RPC: closing connection %s %p", c.Address, c)
	c.State = "closed"

	// close the websocket
	if c.WS != nil {
		err := c.WS.Close(
			websocket.StatusInternalError,
			fmt.Sprintf("RPC: close conn %s", c.Address),
		)
		if err != nil {
			logrus.Errorf("RPC: could not close web socket %s", c.Address)
		}
	}

	// wakeup pending waiters with error
	for reqid := range c.PendingRequests {
		pending := c.PendingRequests[reqid]
		pending.ReplyChan <- fmt.Errorf("RPC: connection closed while request is pending %s %s", c.Address, reqid)
	}

	// tell the RPC to remove this connection which will reconnect if desired
	c.RPC.RemoveConnection(c)
}

// NewRequest initializes the request id and register it on the connection pending requests
func (c *RPCConnWS) NewRequest(req *RPCMessage, res RPCResponse) chan error {
	pending := &RPCPendingRequest{
		Req:       req,
		Res:       res,
		Conn:      c,
		ReplyChan: make(chan error, 1),
	}
	req.RequestID = fmt.Sprintf("%s-%d", c.Address, c.NextRequestID)
	c.NextRequestID++
	c.PendingRequests[req.RequestID] = pending
	return pending.ReplyChan
}

// SendMessage sends the pending request
func (c *RPCConnWS) SendMessage(msg interface{}) error {
	writer, err := c.WS.Writer(context.TODO(), websocket.MessageBinary)
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
		// if err != nil && err.Error() != "failed to read: cannot use EOFed reader" {
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
	c.Lock.Lock()
	pending := c.PendingRequests[msg.RequestID]
	delete(c.PendingRequests, msg.RequestID)
	c.Lock.Unlock()

	if pending == nil {
		logrus.Errorf("RPC: no pending request for %s %s", c.Address, msg.RequestID)
	} else {
		err := json.Unmarshal(msg.RawBytes, pending.Res)
		pending.ReplyChan <- err
	}
}

// HandlePing handles an incoming message of type ping
func (c *RPCConnWS) HandlePing(msg *RPCMessage) {
	c.SendMessage(&RPCMessage{
		Op:        "pong",
		RequestID: msg.RequestID,
		Took:      0,
	})
}
