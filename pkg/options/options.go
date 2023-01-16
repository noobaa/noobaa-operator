package options

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Print the list of global flags",
		Run:   RunOptions,
		Args:  cobra.NoArgs,
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
	ContainerImageTag = "master-20220913"
	// ContainerImageSemverLowerBound is the lower bound for supported image versions
	ContainerImageSemverLowerBound = "5.0.0"
	// ContainerImageSemverUpperBound is the upper bound for supported image versions
	ContainerImageSemverUpperBound = "6.0.0"
	// ContainerImageName is the default image name without the tag/version
	ContainerImageName = ContainerImageOrg + "/" + ContainerImageRepo
	// ContainerImage is the full default image url
	ContainerImage = ContainerImageName + ":" + ContainerImageTag

	// AdminAccountEmail is the default email used by the admin account
	AdminAccountEmail = "admin@noobaa.io"

	// OperatorAccountEmail is the default email used by the operator account
	OperatorAccountEmail = "operator@noobaa.io"

	// SystemName is a constant as we want just a single system per namespace
	SystemName = "noobaa"

	// ServiceServingCertCAFile points to OCP root CA to be added to the default root CA list
	ServiceServingCertCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
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

// DBPostgresImage is the default postgres db image url
// currently it can not be overridden.
var DBPostgresImage = "centos/postgresql-12-centos7"

// DBMongoImage is the default mongo db image url
// this is used during migration to solve issues where mongo STS referencing to postgres image
var DBMongoImage = "centos/mongodb-36-centos7"

// DBType is the default db image type
// it can be overridden for testing or different types.
var DBType = "postgres"

// DBVolumeSizeGB can be used to override the default database volume size
var DBVolumeSizeGB = 0

// DBStorageClass is used for PVC's allocation for the noobaa server data
// it can be overridden for testing or different PV providers.
var DBStorageClass = ""

// MongoDbURL is used for providing mongodb url
// it can be overridden for testing or different url.
var MongoDbURL = ""

// DebugLevel can be used to override the default debug level
var DebugLevel = "default_level"

// PVPoolDefaultStorageClass is used for PVC's allocation for the noobaa server data
// it can be overridden for testing or different PV providers.
var PVPoolDefaultStorageClass = ""

// ImagePullSecret is optionally used to authenticate when pulling the container images
// which is needed when using a private container registry.
var ImagePullSecret = ""

// MiniEnv setting this option indicates to the operator that it is deployed on low resource environment
// This info is used by the operator for environment based decisions (e.g. number of resources to request per
// pod)
var MiniEnv = false

// DisableLoadBalancerService is used for setting the service type to ClusterIP instead of LoadBalancer
var DisableLoadBalancerService = false

// AdmissionWebhook is used for deploying the system with admission validation webhook
var AdmissionWebhook = false

// S3LoadBalancerSourceSubnets is used for setting the source subnets for the load balancer
// created for noobaa S3 service
var S3LoadBalancerSourceSubnets = []string{}

// STSLoadBalancerSourceSubnets is used for setting the source subnets for the load balancer
// created for noobaa STS service
var STSLoadBalancerSourceSubnets = []string{}

// ShowSecrets is used to show the secrets in the status output
var ShowSecrets = false

// SubDomainNS returns a unique subdomain for the namespace
func SubDomainNS() string {
	return Namespace + ".noobaa.io"
}

// ObjectBucketProvisionerName returns the provisioner name to be used in storage classes for OB/OBC
func ObjectBucketProvisionerName() string {
	return SubDomainNS() + "/obc"
}

// WatchNamespace returns the namespace which NooBaa operator will watch for changes
func WatchNamespace() string {
	ns, err := util.GetWatchNamespace()
	if err == nil {
		return ns
	}

	return Namespace
}

// FlagSet defines the
var FlagSet = pflag.NewFlagSet("noobaa", pflag.ContinueOnError)

func init() {
	ns, _ := util.GetWatchNamespace()
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
	FlagSet.StringVar(
		&DBType, "db-type",
		DBType, "The type of database container image (mongodb, postgres)",
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
		&MongoDbURL, "mongodb-url",
		MongoDbURL, "url for mongodb",
	)
	FlagSet.StringVar(
		&DebugLevel, "debug-level",
		DebugLevel, "The type of debug sets that the system prints (all, nsfs, warn, default_level)",
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
	FlagSet.BoolVar(
		&DisableLoadBalancerService, "disable-load-balancer",
		false, "Set the service type to ClusterIP instead of LoadBalancer",
	)
	FlagSet.BoolVar(
		&AdmissionWebhook, "admission",
		false, "Install the system with admission validation webhook",
	)
	FlagSet.BoolVar(
		&ShowSecrets, "show-secrets",
		false, "Show the secrets in the status output",
	)
	FlagSet.StringArrayVar(
		&S3LoadBalancerSourceSubnets, "s3-load-balancer-source-subnets",
		[]string{}, "The source subnets for the S3 service load balancer",
	)
	FlagSet.StringArrayVar(
		&STSLoadBalancerSourceSubnets, "sts-load-balancer-source-subnets",
		[]string{}, "The source subnets for the STS service load balancer",
	)
}
