package olm

import (
	"fmt"
	"io"
	"net/http"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Cmd returns a CLI command
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
		CmdCSV(),
	)
	return cmd
}

// CmdHubInstall returns a CLI command
func CmdHubInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator from operatorhub.io",
		Run:   RunHubInstall,
	}
	return cmd
}

// CmdHubUninstall returns a CLI command
func CmdHubUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator from operatorhub.io",
		Run:   RunHubUninstall,
	}
	return cmd
}

// CmdHubStatus returns a CLI command
func CmdHubStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa-operator from operatorhub.io",
		Run:   RunHubStatus,
	}
	return cmd
}

// CmdLocalInstall returns a CLI command
func CmdLocalInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-install",
		Short: "Install noobaa-operator from local OLM catalog",
		Run:   RunLocalInstall,
	}
	return cmd
}

// CmdLocalUninstall returns a CLI command
func CmdLocalUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-uninstall",
		Short: "Uninstall noobaa-operator from local OLM catalog",
		Run:   RunLocalUninstall,
	}
	return cmd
}

// CmdCSV returns a CLI command
func CmdCSV() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "CSV commands",
	}
	cmd.AddCommand(
		CmdCSVYaml(),
	)
	return cmd
}

// CmdCSVYaml returns a CLI command
func CmdCSVYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled CSV",
		Run:   RunCSVYaml,
	}
	return cmd
}

// RunHubInstall runs a CLI command
func RunHubInstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCreateSkipExisting(obj)
	}
}

// RunHubUninstall runs a CLI command
func RunHubUninstall(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for i := len(hub.Objects) - 1; i >= 0; i-- {
		obj := hub.Objects[i]
		util.KubeDelete(obj)
	}
}

// RunHubStatus runs a CLI command
func RunHubStatus(cmd *cobra.Command, args []string) {
	hub := LoadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCheck(obj)
	}
}

// RunLocalInstall runs a CLI command
func RunLocalInstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalInstall()")
}

// RunLocalUninstall runs a CLI command
func RunLocalUninstall(cmd *cobra.Command, args []string) {
	panic("TODO implement olm.RunLocalUninstall()")
}

// RunCSVYaml runs a CLI command
func RunCSVYaml(cmd *cobra.Command, args []string) {
	fmt.Print(bundle.File_deploy_olm_catalog_package_noobaa_operator_v1_1_1_clusterserviceversion_yaml)
}

// HubConf keeps the operatorhub yaml objects
type HubConf struct {
	Objects []*unstructured.Unstructured
}

// LoadHubConf loads the operatorhub yaml objects
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
