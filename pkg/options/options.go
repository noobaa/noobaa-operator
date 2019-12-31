package options

import (
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/version"
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
	ContainerImageTag = "5.3.0-master20191223"
	// ContainerImageSemverLowerBound is the lower bound for supported image versions
	ContainerImageSemverLowerBound = "5.0.0"
	// ContainerImageSemverUpperBound is the upper bound for supported image versions
	ContainerImageSemverUpperBound = "6.0.0"
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

// OperatorImage is the container image url built from https://github.com/noobaa/noobaa-operator
// it can be overridden for testing or different registry locations.
var OperatorImage = "noobaa/noobaa-operator:" + version.Version

// NooBaaImage is the container image url built from https://github.com/noobaa/noobaa-core
// it can be overridden for testing or different registry locations.
var NooBaaImage = ContainerImage

// DBImage is the default db image url
// it can be overridden for testing or different registry locations.
var DBImage = "centos/mongodb-36-centos7"

// DBVolumeSizeGB can be used to override the default database volume size
var DBVolumeSizeGB = 0

// DBStorageClass is used for PVC's allocation for the noobaa server data
// it can be overridden for testing or different PV providers.
var DBStorageClass = ""

// PVPoolDefaultStorageClass is used for PVC's allocation for the noobaa server data
// it can be overridden for testing or different PV providers.
var PVPoolDefaultStorageClass = ""

// ImagePullSecret is optionally used to authenticate when pulling the container images
// which is needed when using a private container registry.
var ImagePullSecret = ""

// MiniEnv setting this option indicates to the operator that it is deployed on low reosurce environment
// This info is used by the operator for environment based decisions (e.g. number of resources to request per
// pod)
var MiniEnv = false

// SubDomainNS returns a unique subdomain for the namespace
func SubDomainNS() string {
	return Namespace + ".noobaa.io"
}

// ObjectBucketProvisionerName returns the provisioner name to be used in storage classes for OB/OBC
func ObjectBucketProvisionerName() string {
	return SubDomainNS() + "/obc"
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
		&OperatorImage, "operator-image",
		OperatorImage, "Operator image",
	)
	FlagSet.StringVar(
		&NooBaaImage, "noobaa-image",
		NooBaaImage, "NooBaa image",
	)
	FlagSet.StringVar(
		&DBImage, "db-image",
		DBImage, "The database container image",
	)
	FlagSet.IntVar(
		&DBVolumeSizeGB, "db-volume-size-gb",
		DBVolumeSizeGB, "The database volume size in GB",
	)
	FlagSet.StringVar(
		&DBStorageClass, "db-storage-class",
		DBStorageClass, "The database volume storage class name",
	)
	FlagSet.StringVar(
		&PVPoolDefaultStorageClass, "pv-pool-default-storage-class",
		PVPoolDefaultStorageClass, "The default storage class name for BackingStores of type pv-pool",
	)
	FlagSet.StringVar(
		&ImagePullSecret, "image-pull-secret",
		ImagePullSecret, "Image pull secret (must be in same namespace)",
	)
	FlagSet.BoolVar(
		&MiniEnv, "mini",
		false, "Signal the operator that it is running in a low resource environment",
	)
}
