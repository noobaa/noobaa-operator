package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

func (cli *CLI) BucketCreate(args []string) {
	if len(args) < 1 || args[0] == "" {
		util.Fatal(fmt.Errorf("Expected 1st argument: bucket-name"))
	}
	bucketName := args[0]
	nbClient := cli.GetNBClient()
	_, err := nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName})
	util.Fatal(err)
}

func (cli *CLI) BucketDelete(args []string) {
	if len(args) < 1 || args[0] == "" {
		util.Fatal(fmt.Errorf("Expected 1st argument: bucket-name"))
	}
	bucketName := args[0]
	nbClient := cli.GetNBClient()
	_, err := nbClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: bucketName})
	util.Fatal(err)
}

func (cli *CLI) BucketList() {
	nbClient := cli.GetNBClient()
	list, err := nbClient.ListBucketsAPI()
	util.Fatal(err)
	for _, bucket := range list.Buckets {
		cli.Log.Println(bucket.Name)
	}
}

func (cli *CLI) GetNBClient() nb.Client {
	s := system.New(types.NamespacedName{Namespace: cli.Namespace, Name: cli.SystemName}, cli.Client, scheme.Scheme, nil)
	s.Load()

	mgmtStatus := s.NooBaa.Status.Services.ServiceMgmt
	if len(mgmtStatus.NodePorts) == 0 {
		fmt.Println("❌ Mgmt service not ready")
		os.Exit(1)
	}
	if s.SecretOp.StringData["auth_token"] == "" {
		fmt.Println("❌ Auth token not ready")
		os.Exit(1)
	}

	nodePort := mgmtStatus.NodePorts[0]
	nodeIP := nodePort[strings.Index(nodePort, "://")+3 : strings.LastIndex(nodePort, ":")]
	nbClient := nb.NewClient(&nb.APIRouterNodePort{
		ServiceMgmt: s.ServiceMgmt,
		NodeIP:      nodeIP,
	})
	nbClient.SetAuthToken(s.SecretOp.StringData["auth_token"])
	return nbClient
}
