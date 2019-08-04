package olm

import (
	"io"
	"net/http"

	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "olm",
		Short: "Deployment using OLM",
	}
	cmd.AddCommand(
		CmdHubInstall(),
		CmdHubUninstall(),
		CmdHubStatus(),
		// CmdLocalInstall(),
		// CmdLocalUninstall(),
	)
	return cmd
}

func CmdHubInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator from operatorhub.io",
		Run:   RunHubInstall,
	}
	return cmd
}

func CmdHubUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator from operatorhub.io",
		Run:   RunHubUninstall,
	}
	return cmd
}

func CmdHubStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa-operator from operatorhub.io",
		Run:   RunHubStatus,
	}
	return cmd
}

func CmdLocalInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-install",
		Short: "Install noobaa-operator from local OLM catalog",
		Run:   RunLocalInstall,
	}
	return cmd
}

func CmdLocalUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-uninstall",
		Short: "Uninstall noobaa-operator from local OLM catalog",
		Run:   RunLocalUninstall,
	}
	return cmd
}

func RunHubInstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCreateSkipExisting(obj)
	}
}

func RunHubUninstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeDelete(obj)
	}
}

func RunHubStatus(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCheck(obj)
	}
}

func RunLocalInstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalInstall()")
}

func RunLocalUninstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalUninstall()")
}

type HubConf struct {
	Objects []*unstructured.Unstructured
}

func LoadHubConf() *HubConf {
	hub := &HubConf{}
	req, err := http.Get("https://operatorhub.io/install/noobaa-operator.yaml")
	util.Panic(err)
	decoder := yaml.NewYAMLOrJSONDecoder(req.Body, 16*1024)
	for {
		obj := &unstructured.Unstructured{}
		err = decoder.Decode(obj)
		if err == io.EOF {
			break
		}
		util.Panic(err)
		hub.Objects = append(hub.Objects, obj)
		// p := printers.YAMLPrinter{}
		// p.PrintObj(obj, os.Stdout)
	}
	return hub
}
