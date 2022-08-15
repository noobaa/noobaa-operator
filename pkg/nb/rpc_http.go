package nb

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// RPCConnHTTP is an http connection which is created per request
// since the actual http connection pooling is handled in the standard http library
type RPCConnHTTP struct {
	RPC     *RPC
	Address string
}

// NewRPCConnHTTP returns a new http connection
func NewRPCConnHTTP(r *RPC, address string) *RPCConnHTTP {
	return &RPCConnHTTP{
		RPC:     r,
		Address: address,
	}
}

// GetAddress returns the connection address
func (c *RPCConnHTTP) GetAddress() string {
	return c.Address
}

// Reconnect is doing nothing for http connection
func (c *RPCConnHTTP) Reconnect() {
}

// Call calls an API method to noobaa over https
func (c *RPCConnHTTP) Call(req *RPCMessage, res RPCResponse) error {
	reqBytes, err := json.Marshal(req)
	util.Panic(err)

	httpRequest, err := http.NewRequest("PUT", c.Address, bytes.NewReader(reqBytes))
	util.Panic(err)

	httpResponse, err := c.RPC.HTTPClient.Do(httpRequest)
	defer func() {
		if httpResponse != nil && httpResponse.Body != nil {
			httpResponse.Body.Close()
		}
	}()
	if err != nil {
		return err
	}

	bodyLen, err := strconv.ParseInt(httpResponse.Header.Get("X-Noobaa-Rpc-Body-Len"), 10, 64)
	if err != nil {
		return err
	}

	resBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	jsonBytes := resBytes[:bodyLen]
	err = json.Unmarshal(jsonBytes, res)
	if err != nil {
		return err
	}

	buffers := resBytes[bodyLen:]
	if len(buffers) > 0 {
		res.Response().SetBuffers(buffers)
	}

	return nil
}
