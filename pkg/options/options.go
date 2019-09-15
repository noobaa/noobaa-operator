package options

import (
	"github.com/noobaa/noobaa-operator/pkg/util"
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
	util.IgnoreError(cmd.Usage())
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

	// AdminAccountEmail is the default email used for admin account
	AdminAccountEmail = "admin@noobaa.io"

	// SystemName is a constant as we want just a single system per namespace
	SystemName = "noobaa"
)

// Namespace is the target namespace for locating the noobaa system
// default is "noobaa" but in shared clusters (mainly for developers?)
// this can be very confusing and cause unintentional overrides
// so we may consider to use current namespace.
var Namespace = "noobaa"

// NooBaaImage is the container image url built from https://github.com/noobaa/noobaa-core
// it can be overriden for testing or for different registry locations.
var NooBaaImage = ContainerImage

// MongoImage is the default mongodb image url
// it can be overriden for testing or for different registry locations.
var MongoImage = "centos/mongodb-36-centos7"

// OperatorImage is the container image url built from https://github.com/noobaa/noobaa-operator
// it can be overriden for testing or for different registry locations.
var OperatorImage = "noobaa/noobaa-operator:" + version.Version

// StorageClassName is used for PVC's allocation for the noobaa server data
// it can be overriden for testing or for different PV providers.
var StorageClassName = ""

// ImagePullSecret is optionally used to authenticate when pulling the container images
// which is needed when using a private container registry.
var ImagePullSecret = ""

// ObjectBucketProvisionerName returns the provisioner name to be used in storage classes for OB/OBC
func ObjectBucketProvisionerName() string {
	return "noobaa.io/" + Namespace + ".bucket"
}

// FlagSet defines the
var FlagSet = pflag.NewFlagSet("noobaa", pflag.ContinueOnError)

func init() {
	ns, _ := k8sutil.GetWatchNamespace()
	if ns == "" {
		ns = util.CurrentNamespace()
	}
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
		&MongoImage, "mongo-image",
		MongoImage, "MongoDB image",
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
