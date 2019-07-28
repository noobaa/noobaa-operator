package cli

import (
	"fmt"
	"strconv"

	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/sirupsen/logrus"
)

// BucketCreate runs a CLI command
func (cli *CLI) BucketCreate(args []string) {
	if len(args) < 1 || args[0] == "" {
		logrus.Fatalf("Expected 1st argument: bucket-name")
	}
	bucketName := args[0]
	nbClient := cli.GetNBClient()
	_, err := nbClient.CreateBucketAPI(nb.CreateBucketParams{Name: bucketName})
	util.Panic(err)
}

// BucketDelete runs a CLI command
func (cli *CLI) BucketDelete(args []string) {
	if len(args) < 1 || args[0] == "" {
		logrus.Fatalf("Expected 1st argument: bucket-name")
	}
	bucketName := args[0]
	nbClient := cli.GetNBClient()
	_, err := nbClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: bucketName})
	util.Panic(err)
}

// BucketList runs a CLI command
func (cli *CLI) BucketList() {
	nbClient := cli.GetNBClient()
	list, err := nbClient.ListBucketsAPI()
	util.Panic(err)
	if len(list.Buckets) == 0 {
		fmt.Printf("No buckets found.\n")
	}
	table := (&util.PrintTable{}).AddRow("#", "BUCKET-NAME")
	for i := range list.Buckets {
		b := &list.Buckets[i]
		table.AddRow(strconv.Itoa(i+1), b.Name)
	}
	fmt.Print(table.String())
}
