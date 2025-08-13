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
	ContainerImageTag = "master-20250812"
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
)

// Namespace is the target namespace for locating the noobaa system
// default is "noobaa" but in shared clusters (mainly for developers?)
// this can be very confusing and cause unintentional overrides
// so we may consider to use current namespace.
var Namespace = "noobaa"

// OperatorImage is the container image url built from https://github.com/noobaa/noobaa-operator
// it can be overridden for testing or different registry locations.
var OperatorImage = "noobaa/noobaa-operator:" + version.Version

// CosiSideCarImage is the container image url built from https://github.com/kubernetes-sigs/container-object-storage-interface-provisioner-sidecar
var CosiSideCarImage = "gcr.io/k8s-staging-sig-storage/objectstorage-sidecar/objectstorage-sidecar:v20221117-v0.1.0-22-g0e67387"

// NooBaaImage is the container image url built from https://github.com/noobaa/noobaa-core
// it can be overridden for testing or different registry locations.
var NooBaaImage = ContainerImage

// DBImage is the default db image url
// it can be overridden for testing or different registry locations.
var DBImage = "quay.io/sclorg/postgresql-16-c9s"

// PostgresMajorVersion is the default postgres major version
// it must match the postgres image version
var PostgresMajorVersion = 16

// PostgresInstances is the default number of postgres instances in a managed postgres cluster
var PostgresInstances = 2

// Psql12Image is the default postgres12 db image url
// currently it can not be overridden.
var Psql12Image = "centos/postgresql-12-centos7"

// DefaultDBVolumeSize is the default volume size
var DefaultDBVolumeSize = "50Gi"

// DBType is the default db image type
// it can be overridden for testing or different types.
var DBType = "postgres"

// DBVolumeSizeGB can be used to override the default database volume size
var DBVolumeSizeGB = 0

// DBStorageClass is used for PVC's allocation for the noobaa server data
// it can be overridden for testing or different PV providers.
var DBStorageClass = ""

// PostgresDbURL is used for providing postgres url
// it can be overridden for testing or different url.
var PostgresDbURL = ""

// PostgresSSLRequired is used to force noobaa to work with SSL with external pgsql
// when using an external postgres DB.
var PostgresSSLRequired = false

// PostgresSSLSelfSigned is used to allow noobaa to work with self-signed SSL with external pgsql
// when using an external postgres DB.
var PostgresSSLSelfSigned = false

// PostgresSSLKey is used for providing the path to the client SSL key file when working with external pgsql
var PostgresSSLKey = ""

// PostgresSSLCert is used for providing the path to the client SSL cert file when working with external pgsql
var PostgresSSLCert = ""

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

// DevEnv setting this option indicates to the operator that it is deployed on development environment
// This info is used by the operator for environment based decisions (e.g. number of resources to request per
// pod)
var DevEnv = false

// DisableLoadBalancerService is used for setting the service type to ClusterIP instead of LoadBalancer
var DisableLoadBalancerService = false

// DisableRoutes is used to disable the reconciliation of openshift route resources
var DisableRoutes = false

// CosiDriverPath is the cosi socket fs path
var CosiDriverPath = "/var/lib/cosi/cosi.sock"

// AdmissionWebhook is used for deploying the system with admission validation webhook
var AdmissionWebhook = false

// TestEnv is used for deploying the system with test env minimal resources
var TestEnv = false

// S3LoadBalancerSourceSubnets is used for setting the source subnets for the load balancer
// created for noobaa S3 service
var S3LoadBalancerSourceSubnets = []string{}

// STSLoadBalancerSourceSubnets is used for setting the source subnets for the load balancer
// created for noobaa STS service
var STSLoadBalancerSourceSubnets = []string{}

// ShowSecrets is used to show the secrets in the status output
var ShowSecrets = false

// ManualDefaultBackingStore is used for disabling and allow deletion of default backingstore
var ManualDefaultBackingStore = false

// AutoscalerType is the default noobaa-endpoint autoscaler type
// it can be overridden for testing or different types. there is no default autoscaler for endpoint
var AutoscalerType = ""

// PrometheusNamespace is prometheus installed namespace
// it can be overridden for testing or different namespace.
var PrometheusNamespace = ""

// AWSSTSARN is used in an AWS STS cluster to assume role ARN
// it can be overridden for testing.
var AWSSTSARN = ""

// CnpgVersion is the version of cloudnative-pg operator to use
var CnpgVersion = "1.25.0"

// CnpgImage is the container image url of cloudnative-pg operator
var CnpgImage = "quay.io/noobaa/cloudnative-pg-noobaa:v1.25.0"

// UseCnpgApiGroup indicates if the original CloudNativePG API group should be used for the installation manifests
// Relevant for the CLI commands. during reconciliation we look at the env variable USE_CNPG_API_GROUP to determine
// if we should use the original API group
var UseCnpgApiGroup = false

// CnpgApiGroup is the API group used for cloudnative-pg CRDs
var CnpgApiGroup = "postgresql.cnpg.noobaa.io"

// SubDomainNS returns a unique subdomain for the namespace
func SubDomainNS() string {
	return Namespace + ".noobaa.io"
}

// ObjectBucketProvisionerName returns the provisioner name to be used in storage classes for OB/OBC
func ObjectBucketProvisionerName() string {
	return SubDomainNS() + "/obc"
}

// COSIDriverName returns the driver name to be used in for COSI
func COSIDriverName() string {
	return "noobaa.objectstorage.k8s.io"
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
		&Psql12Image, "psql-12-image",
		Psql12Image, "The database old container image",
	)
	FlagSet.StringVar(
		&CosiSideCarImage, "cosi-sidecar-image",
		CosiSideCarImage, "The cosi side car container image",
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
		&PostgresDbURL, "postgres-url",
		PostgresDbURL, "url for postgresql",
	)
	FlagSet.BoolVar(
		&PostgresSSLRequired, "pg-ssl-required",
		false, "Force noobaa to work with ssl (external postgres - server-side) [if server cert is self-signed, needs to add --ssl-unauthorized]",
	)
	FlagSet.BoolVar(
		&PostgresSSLSelfSigned, "pg-ssl-unauthorized",
		false, "Allow the client to work with self-signed ssl (external postgres - server-side)",
	)
	FlagSet.StringVar(
		&PostgresSSLKey, "pg-ssl-key",
		PostgresSSLKey, "ssl key for postgres (client-side cert - need to be signed by external pg accepted CA)",
	)
	FlagSet.StringVar(
		&PostgresSSLCert, "pg-ssl-cert",
		PostgresSSLCert, "ssl cert for postgres (client-side cert - need to be signed by external pg accepted CA)",
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
		&DevEnv, "dev",
		false, "Set sufficient resources for dev env",
	)
	FlagSet.BoolVar(
		&TestEnv, "test-env",
		false, "Install the system with test env minimal resources",
	)
	FlagSet.BoolVar(
		&DisableLoadBalancerService, "disable-load-balancer",
		false, "Set the service type to ClusterIP instead of LoadBalancer",
	)
	FlagSet.BoolVar(
		&DisableRoutes, "disable-routes",
		false, "disable the reconciliation of openshift route resources",
	)
	FlagSet.StringVar(
		&CosiDriverPath, "cosi-driver-path",
		CosiDriverPath, "unix socket path for COSI",
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
	FlagSet.BoolVar(
		&ManualDefaultBackingStore, "manual-default-backingstore",
		false, "allow to delete the default backingstore",
	)
	FlagSet.StringVar(
		&AutoscalerType, "autoscaler-type",
		AutoscalerType, "The type of autoscaler (hpav2, keda)",
	)
	FlagSet.StringVar(
		&PrometheusNamespace, "prometheus-namespace",
		PrometheusNamespace, "namespace with installed prometheus for autoscaler",
	)
	FlagSet.StringVar(
		&AWSSTSARN, "aws-sts-arn",
		AWSSTSARN, "The AWS STS Role ARN which will assume role",
	)
	FlagSet.StringVar(
		&CnpgVersion, "cnpg-version",
		CnpgVersion, "Version of CloudNativePG operator",
	)
	FlagSet.StringVar(
		&CnpgImage, "cnpg-image",
		CnpgImage, "CloudNativePG operator image",
	)
	FlagSet.BoolVar(
		&UseCnpgApiGroup, "use-cnpg-api-group",
		UseCnpgApiGroup, "Use the original CloudNativePG API group for the installation manifests. Should be used when using an original image of CloudNativePG.",
	)
}
