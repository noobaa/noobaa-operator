package system

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/spf13/cobra"
	sigyaml "sigs.k8s.io/yaml"
)

// CmdAPICall returns a CLI command
func CmdAPICall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <api-name> <method-name> [<params-json>]",
		Short: "Make api call",
		Run:   RunAPICall,
		Args:  cobra.RangeArgs(2, 3),
	}
	cmd.Flags().StringP("output", "o", "yaml", "yaml|json|golang")
	return cmd
}

// RunAPICall runs a CLI command
func RunAPICall(cmd *cobra.Command, args []string) {
	log := util.Logger()

	args = append(args, "", "", "")
	apiName := args[0]
	methodName := args[1]
	paramsStr := args[2]
	if apiName != "" && !strings.HasSuffix(apiName, "_api") {
		apiName += "_api"
	}

	sysClient, err := Connect(true)
	if err != nil {
		log.Fatalf("❌ %s", err)
	}

	req := &nb.RPCMessage{API: apiName, Method: methodName}
	res := &struct {
		nb.RPCMessage `json:",inline"`
		Reply         map[string]interface{} `json:"reply"`
	}{}

	if paramsStr != "" {
		err := json.Unmarshal([]byte(paramsStr), &req.Params)
		if err != nil {
			log.Fatalf("❌ %s", err)
		}
	}

	err = sysClient.NBClient.Call(req, res)
	if err != nil {
		log.Fatalf("❌ %s", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "golang":
		fmt.Printf("%#v\n", res.Reply)

	case "json":
		bytes, err := json.Marshal(res.Reply)
		util.Panic(err)
		fmt.Println(string(bytes))

	default:
		fallthrough
	case "yaml":
		bytes, err := sigyaml.Marshal(res.Reply)
		util.Panic(err)
		fmt.Println(string(bytes))
	}
}
