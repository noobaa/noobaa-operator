package main

import (
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/noobaa/noobaa-operator/pkg/apis"
	"github.com/noobaa/noobaa-operator/pkg/cli"
	"github.com/noobaa/noobaa-operator/pkg/util"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	util.Fatal(apiextv1beta1.AddToScheme(scheme.Scheme))
	util.Fatal(apis.AddToScheme(scheme.Scheme))
	cli.New().Run()
}
