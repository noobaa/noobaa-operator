package cli

import (
	"io"
	"net/http"

	"github.com/noobaa/noobaa-operator/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (cli *CLI) HubInstall() {
	hub := cli.loadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCreateSkipExisting(cli.Client, obj)
	}
}

func (cli *CLI) HubUninstall() {
	hub := cli.loadHubConf()
	for _, obj := range hub.Objects {
		util.KubeDelete(cli.Client, obj)
	}
}

func (cli *CLI) HubStatus() {
	hub := cli.loadHubConf()
	for _, obj := range hub.Objects {
		util.KubeCheck(cli.Client, obj)
	}
}

type HubConf struct {
	Objects []*unstructured.Unstructured
}

func (cli *CLI) loadHubConf() *HubConf {
	hub := &HubConf{}
	req, err := http.Get("https://operatorhub.io/install/noobaa-operator.yaml")
	util.Fatal(err)
	decoder := yaml.NewYAMLOrJSONDecoder(req.Body, 16*1024)
	for {
		obj := &unstructured.Unstructured{}
		err = decoder.Decode(obj)
		if err == io.EOF {
			break
		}
		util.Fatal(err)
		hub.Objects = append(hub.Objects, obj)
		// p := printers.YAMLPrinter{}
		// p.PrintObj(obj, os.Stdout)
	}
	return hub
}
