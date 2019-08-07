package options

import (
	"github.com/noobaa/noobaa-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Print the list of global flags",
		Run:   RunOptions,
	}
	return cmd
}

// RunOptions runs a CLI command
func RunOptions(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

const (
	// ContainerImageOrg is the org of the default image url
	ContainerImageOrg = "noobaa"
	// ContainerImageRepo is the repo of the default image url
	ContainerImageRepo = "noobaa-core"
	// ContainerImageTag is the tag of the default image url
	ContainerImageTag = "5"
	// ContainerImageConstraintSemver is the contraints of supported image versions
	ContainerImageConstraintSemver = ">=5, <6"
	// ContainerImageName is the default image name without the tag/version
	ContainerImageName = ContainerImageOrg + "/" + ContainerImageRepo
	// ContainerImage is the full default image url
	ContainerImage = ContainerImageName + ":" + ContainerImageTag

	// MongoImage is the default mongodb image url
	MongoImage = "centos/mongodb-36-centos7"

	// AdminAccountEmail is the default email used for admin account
	AdminAccountEmail = "admin@noobaa.io"

	// SystemName is a constant as we want just a single system per namespace
	SystemName = "noobaa"
)

var Namespace = "noobaa" //util.CurrentNamespace()
var NooBaaImage = ContainerImage
var OperatorImage = "noobaa/noobaa-operator:" + version.Version
var StorageClassName = ""
var ImagePullSecret = ""

// FlagSet defines the
var FlagSet = pflag.NewFlagSet("noobaa", pflag.ContinueOnError)

func init() {
	ns, _ := k8sutil.GetWatchNamespace()
	if ns != "" {
		Namespace = ns
	}
	FlagSet.StringVarP(
		&Namespace, "namespace", "n",
		Namespace, "Target namespace",
	)
	FlagSet.StringVar(
		&StorageClassName, "storage-class",
		StorageClassName, "Storage class name",
	)
	FlagSet.StringVar(
		&NooBaaImage, "noobaa-image",
		NooBaaImage, "NooBaa image",
	)
	FlagSet.StringVar(
		&OperatorImage, "operator-image",
		OperatorImage, "Operator image",
	)
	FlagSet.StringVar(
		&ImagePullSecret, "image-pull-secret",
		ImagePullSecret, "Image pull secret (must be in same namespace)",
	)
}
