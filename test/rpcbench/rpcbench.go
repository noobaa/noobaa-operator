package main

import (
	"flag"
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/sirupsen/logrus"
)

var client = flag.Bool("client", false, "client option")
var server = flag.Bool("server", false, "server option")
var help = flag.Bool("help", false, "help option")
var port = flag.String("port", "5656", "port option")
var hostname = flag.String("hostname", "127.0.0.1", "hostname option")
var proto = flag.String("proto", "wss", "protocol (ws/wss/http/https)")
var wsize = flag.Int64("wsize", 1024*1024, "wsize option")
var rsize = flag.Int64("rsize", 1024*1024, "rsize option")
var concur = flag.Int("concur", 5, "concur option")

// noobaa-core/     $ node src/rpc/rpc_benchmark.js --server
// noobaa-operator/ $ go run test/rpcbench/rpcbench.go --client -v
func main() {

	flag.Parse()

	// help print
	if *help || (!*server && !*client) {
		fmt.Println("Usage:")
		fmt.Println("  rpcbench --client")
		fmt.Println("  rpcbench --server")
		return
	}

	// client side
	if *client {
		ClientMain()
	}

	if *server {
		fmt.Println("Server is not yet implemented. Sorry...")
		return
	}
}

// ClientMain handles client option
func ClientMain() {

	addr := fmt.Sprintf("%s://%s:%s/rpc/", *proto, *hostname, *port)
	fmt.Println(addr)

	c := nb.NewClient(&nb.SimpleRouter{Address: addr})
	type BenchParams struct {
		Rsize int64 `json:"rsize"`
		Wsize int64 `json:"wsize"`
	}
	for i := 0; i < *concur; i++ {
		go func() {
			for {
				req := &nb.RPCMessage{
					API:    "rpcbench",
					Method: "io",
					Params: &BenchParams{
						Rsize: *rsize,
						Wsize: *wsize,
					},
				}
				res := &struct {
					nb.RPCMessage `json:",inline"`
					Reply         BenchParams `json:"reply"`
				}{}
				err := c.Call(req, res)
				if err != nil {
					logrus.Errorf("RPCBenchmark error: %v", err)
				}

				//fmt.logf("Made it %#v\n", res.Reply)
			}
		}()
	}
	neverStop := make(chan int)
	<-neverStop
}
