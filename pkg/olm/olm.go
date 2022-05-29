package olm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/crd"
	"github.com/noobaa/noobaa-operator/v5/pkg/operator"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"

	"github.com/blang/semver/v4"
	operv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/printers"
	sigyaml "sigs.k8s.io/yaml"
)

type unObj = map[string]interface{}
type unArr = []interface{}

type generateCSVParams struct {
	IsForODF  bool
	SkipRange string
	Replaces  string
}

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "olm",
		Short: "OLM related commands",
	}
	cmd.AddCommand(
		CmdCatalog(),
		CmdCSV(),
		CmdHubInstall(),
		CmdHubUninstall(),
		CmdHubStatus(),
		// CmdLocalInstall(),
		// CmdLocalUninstall(),
	)
	return cmd
}

// CmdCatalog returns a CLI command
func CmdCatalog() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Create OLM catalog dir",
		Run:   RunCatalog,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("dir", "./build/_output/olm", "The output dir for the OLM package")
	cmd.Flags().Bool("odf", false, "Build package according to ODF requirements")
	cmd.Flags().String("csv-name", "", "File name for the CSV YAML")
	cmd.Flags().String("skip-range", "", "set the olm.skipRange annotation in the CSV")
	cmd.Flags().String("replaces", "", "set the replaces property in the CSV")
	return cmd
}

// CmdCSV returns a CLI command
func CmdCSV() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "Print CSV yaml",
		Run:   RunCSV,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdHubInstall returns a CLI command
func CmdHubInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator from operatorhub.io",
		Run:   RunHubInstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdHubUninstall returns a CLI command
func CmdHubUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator from operatorhub.io",
		Run:   RunHubUninstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdHubStatus returns a CLI command
func CmdHubStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa-operator from operatorhub.io",
		Run:   RunHubStatus,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdLocalInstall returns a CLI command
func CmdLocalInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-install",
		Short: "Install noobaa-operator from local OLM catalog",
		Run:   RunLocalInstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdLocalUninstall returns a CLI command
func CmdLocalUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-uninstall",
		Short: "Uninstall noobaa-operator from local OLM catalog",
		Run:   RunLocalUninstall,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCatalog runs a CLI command
func RunCatalog(cmd *cobra.Command, args []string) {
	log := util.Logger()

	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		log.Fatalf(`Missing required flag: --dir: %s`, cmd.UsageString())
	}

	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	opConf := operator.LoadOperatorConf(cmd)

	var versionDir string

	forODF, _ := cmd.Flags().GetBool("odf")
	skipRange, _ := cmd.Flags().GetString("skip-range")
	replaces, _ := cmd.Flags().GetString("replaces")
	csvParams := &generateCSVParams{
		IsForODF:  forODF,
		SkipRange: skipRange,
		Replaces:  replaces,
	}
	if forODF {
		versionDir = dir
	} else {
		versionDir = dir + version.Version + "/"
	}

	pkgBytes, err := sigyaml.Marshal(unObj{
		"packageName":    "noobaa-operator",
		"defaultChannel": "alpha",
		"channels": unArr{unObj{
			"name":       "alpha",
			"currentCSV": "noobaa-operator.v" + version.Version,
		}},
	})
	util.Panic(err)

	csvFileName, _ := cmd.Flags().GetString("csv-name")
	if csvFileName == "" {
		csvFileName = versionDir + "noobaa-operator.v" + version.Version + ".clusterserviceversion.yaml"
	} else {
		csvFileName = versionDir + csvFileName
	}

	util.Panic(os.MkdirAll(versionDir, 0755))
	if !forODF {
		util.Panic(ioutil.WriteFile(dir+"noobaa-operator.package.yaml", pkgBytes, 0644))
	}
	util.Panic(util.WriteYamlFile(csvFileName, GenerateCSV(opConf, csvParams)))
	crd.ForEachCRD(func(c *crd.CRD) {
		if c.Spec.Group == nbv1.SchemeGroupVersion.Group {
			util.Panic(util.WriteYamlFile(versionDir+c.Name+".crd.yaml", c))
		}
	})
}

// RunCSV runs a CLI command
func RunCSV(cmd *cobra.Command, args []string) {
	opConf := operator.LoadOperatorConf(cmd)
	csv := GenerateCSV(opConf, nil)
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(csv, os.Stdout))
}

// GenerateCSV creates the CSV
func GenerateCSV(opConf *operator.Conf, csvParams *generateCSVParams) *operv1.ClusterServiceVersion {
	almExamples, err := json.Marshal([]runtime.Object{
		util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml),
	})
	util.Panic(err)

	o := util.KubeObject(bundle.File_deploy_olm_noobaa_operator_clusterserviceversion_yaml)
	csv := o.(*operv1.ClusterServiceVersion)
	csv.Name = "noobaa-operator.v" + version.Version
	csv.Namespace = options.Namespace
	csv.Annotations["containerImage"] = options.OperatorImage
	// this annotation hides the operator in OCP console
	// csv.Annotations["createdAt"] = ???
	csv.Annotations["alm-examples"] = string(almExamples)
	csv.Spec.Version.Version = semver.MustParse(version.Version)
	csv.Spec.Description = bundle.File_deploy_olm_description_md
	csv.Spec.Icon[0].Data = bundle.File_deploy_olm_noobaa_icon_base64
	csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions = []operv1.StrategyDeploymentPermissions{}
	csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions = append(csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions,
		operv1.StrategyDeploymentPermissions{
			ServiceAccountName: opConf.SA.Name,
			Rules:              opConf.ClusterRole.Rules,
		})
	csv.Spec.InstallStrategy.StrategySpec.Permissions = []operv1.StrategyDeploymentPermissions{}
	csv.Spec.InstallStrategy.StrategySpec.Permissions = append(csv.Spec.InstallStrategy.StrategySpec.Permissions,
		operv1.StrategyDeploymentPermissions{
			ServiceAccountName: opConf.SA.Name,
			Rules:              opConf.Role.Rules,
		})
	csv.Spec.InstallStrategy.StrategySpec.Permissions = append(csv.Spec.InstallStrategy.StrategySpec.Permissions,
		operv1.StrategyDeploymentPermissions{
			ServiceAccountName: opConf.SAUI.Name,
			Rules:              opConf.RoleUI.Rules,
		})
	csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs = []operv1.StrategyDeploymentSpec{}
	csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs = append(csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs,
		operv1.StrategyDeploymentSpec{
			Name: opConf.Deployment.Name,
			Spec: opConf.Deployment.Spec,
		})

	if csvParams != nil {
		if csvParams.IsForODF {
			// add anotation to hide the operator in OCP console
			csv.Annotations["operators.operatorframework.io/operator-type"] = "non-standalone"

			// add env vars for noobaa-core and noobaa-db images
			operatorContainer := &csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Template.Spec.Containers[0]
			operatorContainer.Env = append(operatorContainer.Env,
				corev1.EnvVar{
					Name:  "NOOBAA_CORE_IMAGE",
					Value: options.NooBaaImage,
				},
				corev1.EnvVar{
					Name:  "NOOBAA_DB_IMAGE",
					Value: options.DBImage,
				},
				corev1.EnvVar{
					Name:  "ENABLE_NOOBAA_ADMISSION",
					Value: "true",
				})

			csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Template.Spec.Tolerations = []corev1.Toleration{
				{
					Key:      "node.ocs.openshift.io/storage",
					Effect:   corev1.TaintEffectNoSchedule,
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
				},
			}
		}

		if csvParams.Replaces != "" {
			csv.Spec.Replaces = csvParams.Replaces
		}

		if csvParams.SkipRange != "" {
			csv.Annotations[operv1.SkipRangeAnnotationKey] = csvParams.SkipRange
		}
	}

	csv.Spec.CustomResourceDefinitions.Owned = []operv1.CRDDescription{}
	csv.Spec.CustomResourceDefinitions.Required = []operv1.CRDDescription{}
	crdDescriptions := map[string]string{
		"NooBaa": `A NooBaa system - Create this to start`,
		"BackingStore": `Storage target spec such as aws-s3, s3-compatible, ibm-cos, PV's and more. ` +
			`Used in BucketClass to construct data placement policies.`,
		"NamespaceStore": `Storage target spec such as aws-s3, s3-compatible, ibm-cos and more. ` +
			`Used in BucketClass to construct namespace policies.`,
		"BucketClass": `Storage policy spec  tiering, mirroring, spreading, namespace policy. ` +
			`Combines BackingStores Or NamespaceStores. Referenced by ObjectBucketClaims.`,
		"ObjectBucketClaim": `Claim a bucket just like claiming a PV. ` +
			`Automate you app bucket provisioning by creating OBC with your app deployment. ` +
			`A secret and configmap (name=claim) will be created with access details for the app pods.`,
		"ObjectBucket": `Used under-the-hood. Created per ObjectBucketClaim and keeps provisioning information.`,
	}
	crdDisplayNames := map[string]string{
		"NooBaa":            "NooBaa",
		"BackingStore":      "Backing Store",
		"NamespaceStore":    "Namespace Store",
		"BucketClass":       "Bucket Class",
		"ObjectBucketClaim": "Object Bucket Claim",
		"ObjectBucket":      "Object Bucket",
	}
	const (
		uiTectonic                     = "urn:alm:descriptor:com.tectonic.ui:"
		uiText                         = uiTectonic + "text"
		uiResources                    = uiTectonic + "resourceRequirements"
		uiFieldGroup                   = uiTectonic + "fieldGroup:"
		uiKubernetes                   = "urn:alm:descriptor:io.kubernetes:"
		uiK8sSecret                    = uiKubernetes + "Secret"
		uiK8sTolerations               = uiKubernetes + "Tolerations"
		uiBooleanSwitch                = uiTectonic + "booleanSwitch"
		uiNumber                       = uiTectonic + "number"
		uiFieldGroupAwsS3              = uiFieldGroup + "awsS3"
		uiFieldGroupAzureBlob          = uiFieldGroup + "azureBlob"
		uiFieldGroupGoogleCloudStorage = uiFieldGroup + "googleCloudStorage"
		uiFieldGroupPvPool             = uiFieldGroup + "pvPool"
		uiFieldGroupS3Compatible       = uiFieldGroup + "s3Compatible"
		uiFieldGroupIBMCos             = uiFieldGroup + "ibmCos"
		uiFieldGroupPlacementPolicy    = uiFieldGroup + "placementPolicy"
		uiFieldGroupNamespacePolicy    = uiFieldGroup + "namespacePolicy"
	)

	crdSpecDescriptors := map[string][]operv1.SpecDescriptor{
		"NooBaa": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{
				Path:         "image",
				XDescriptors: []string{uiText},
				Description:  "DBImage (optional) overrides the default image for the db container.",
				DisplayName:  "DB Image",
			},
			operv1.SpecDescriptor{
				Path:         "dbImage",
				XDescriptors: []string{uiText},
				Description:  "Image (optional) overrides the default image for the server container.",
				DisplayName:  "Image",
			},
			operv1.SpecDescriptor{
				Path:         "coreResources",
				XDescriptors: []string{uiResources},
				Description:  "CoreResources (optional) overrides the default resource requirements for the server container.",
				DisplayName:  "Core Resources",
			},
			operv1.SpecDescriptor{
				Path:         "dbResources",
				XDescriptors: []string{uiResources},
				Description:  "DBResources (optional) overrides the default resource requirements for the db container.",
				DisplayName:  "DB Resources",
			},
			operv1.SpecDescriptor{
				Path:         "dbVolumeResources",
				XDescriptors: []string{uiResources},
				Description:  "DBVolumeResources (optional) overrides the default PVC resource requirements for the database volume. For the time being this field is immutable and can only be set on system creation. This is because volume size updates are only supported for increasing the size, and only if the storage class specifies `allowVolumeExpansion: true`, +immutable.",
				DisplayName:  "Image",
			},
			operv1.SpecDescriptor{
				Path:         "dbStorageClass",
				XDescriptors: []string{uiText},
				Description:  "DBStorageClass (optional) overrides the default cluster StorageClass for the database volume. For the time being this field is immutable and can only be set on system creation. This affects where the system stores its database which contains system config, buckets, objects meta-data and mapping file parts to storage locations. +immutable.",
				DisplayName:  "DB StorageClass",
			},
			operv1.SpecDescriptor{
				Path:         "pvPoolDefaultStorageClass",
				XDescriptors: []string{uiText},
				Description:  "PVPoolDefaultStorageClass (optional) overrides the default cluster StorageClass for the pv-pool volumes. This affects where the system stores data chunks (encrypted). Updates to this field will only affect new pv-pools, but updates to existing pools are not supported by the operator.",
				DisplayName:  "PV Pool DefaultStorageClass",
			},
			// this descriptor caused the OCP console to crash on noobaa CRD page, when trying to display tolerations.
			// removing for now
			// operv1.SpecDescriptor{
			// 	Path:         "tolerations",
			// 	XDescriptors: []string{uiK8sTolerations},
			// 	Description:  "Tolerations.",
			// 	DisplayName:  "Tolerations",
			// },
			operv1.SpecDescriptor{
				Path:         "imagePullSecret",
				XDescriptors: []string{uiK8sSecret},
				Description:  "ImagePullSecret (optional) sets a pull secret for the system image.",
				DisplayName:  "Image Pull Secret",
			},
		},
		"BackingStore": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{
				Description:  "Region is the AWS region.",
				Path:         "awsS3.region",
				XDescriptors: []string{uiFieldGroupAwsS3, uiText},
				DisplayName:  "Region",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "awsS3.secret.name",
				XDescriptors: []string{uiFieldGroupAwsS3, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SSLDisabled allows to disable SSL and use plain http.",
				Path:         "awsS3.sslDisabled",
				XDescriptors: []string{uiFieldGroupAwsS3, uiBooleanSwitch},
				DisplayName:  "SSL Disabled",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target S3 bucket.",
				Path:         "awsS3.targetBucket",
				XDescriptors: []string{uiFieldGroupAwsS3, uiText},
				DisplayName:  "Target Bucket",
			},
			operv1.SpecDescriptor{
				Description:  " Secret refers to a secret that provides the credentials. The secret should define AccountName and AccountKey as provided\nby Azure Blob.",
				Path:         "azureBlob.secret.name",
				XDescriptors: []string{uiFieldGroupAzureBlob, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBlobContainer is the name of the target Azure Blob container.",
				Path:         "azureBlob.targetBlobContainer",
				XDescriptors: []string{uiFieldGroupAzureBlob, uiText},
				DisplayName:  "Target Blob Container",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define GoogleServiceAccountPrivateKeyJson containing\nthe entire json string as provided by Google.",
				Path:         "googleCloudStorage.secret.name",
				XDescriptors: []string{uiFieldGroupGoogleCloudStorage, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target S3 bucket.",
				Path:         "googleCloudStorage.targetBucket",
				XDescriptors: []string{uiFieldGroupGoogleCloudStorage, uiText},
				DisplayName:  "Target Bucket",
			},
			operv1.SpecDescriptor{
				Description:  "NumVolumes is the number of volumes to allocate.",
				Path:         "pvPool.numVolumes",
				XDescriptors: []string{uiFieldGroupPvPool, uiNumber},
				DisplayName:  "Num Volumes",
			},
			operv1.SpecDescriptor{
				Description:  "VolumeResources represents the minimum resources each volume should have.",
				Path:         "pvPool.resources",
				XDescriptors: []string{uiFieldGroupPvPool, uiResources},
				DisplayName:  "Resources",
			},
			operv1.SpecDescriptor{
				Description:  "StorageClass is the name of the storage class to use for the PV's.",
				Path:         "pvPool.storageClass",
				XDescriptors: []string{uiFieldGroupPvPool, uiText},
				DisplayName:  "Storage Class",
			},
			operv1.SpecDescriptor{
				Description:  "Endpoint is the S3 compatible endpoint: http(s)://host:port.",
				Path:         "s3Compatible.endpoint",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiText},
				DisplayName:  "End Point",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "s3Compatible.secret.name",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SignatureVersion specifies the client signature version to use when signing requests.",
				Path:         "s3Compatible.signatureVersion",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiNumber},
				DisplayName:  "Signature Version",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target S3 bucket.",
				Path:         "s3Compatible.targetBucket",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiText},
				DisplayName:  "Target Bucket",
			},
			operv1.SpecDescriptor{
				Description:  "Endpoint is the IBM COS endpoint: http(s)://host:port.",
				Path:         "IBMCos.endpoint",
				XDescriptors: []string{uiFieldGroupIBMCos, uiText},
				DisplayName:  "End Point",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY.",
				Path:         "IBMCos.secret.name",
				XDescriptors: []string{uiFieldGroupIBMCos, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SignatureVersion specifies the client signature version to use when signing requests.",
				Path:         "IBMCos.signatureVersion",
				XDescriptors: []string{uiFieldGroupIBMCos, uiNumber},
				DisplayName:  "Signature Version",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target IBM COS bucket.",
				Path:         "IBMCos.targetBucket",
				XDescriptors: []string{uiFieldGroupIBMCos, uiText},
				DisplayName:  "Target Bucket",
			},
		},

		"NamespaceStore": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{
				Description:  "Region is the AWS region.",
				Path:         "awsS3.region",
				XDescriptors: []string{uiFieldGroupAwsS3, uiText},
				DisplayName:  "Region",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "awsS3.secret.name",
				XDescriptors: []string{uiFieldGroupAwsS3, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SSLDisabled allows to disable SSL and use plain http.",
				Path:         "awsS3.sslDisabled",
				XDescriptors: []string{uiFieldGroupAwsS3, uiBooleanSwitch},
				DisplayName:  "SSL Disabled",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target S3 bucket.",
				Path:         "awsS3.targetBucket",
				XDescriptors: []string{uiFieldGroupAwsS3, uiText},
				DisplayName:  "Target Bucket",
			},
			operv1.SpecDescriptor{
				Description:  " Secret refers to a secret that provides the credentials. The secret should define AccountName and AccountKey as provided\nby Azure Blob.",
				Path:         "azureBlob.secret.name",
				XDescriptors: []string{uiFieldGroupAzureBlob, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBlobContainer is the name of the target Azure Blob container.",
				Path:         "azureBlob.targetBlobContainer",
				XDescriptors: []string{uiFieldGroupAzureBlob, uiText},
				DisplayName:  "Target Blob Container",
			},
			operv1.SpecDescriptor{
				Description:  "Endpoint is the S3 compatible endpoint: http(s)://host:port.",
				Path:         "s3Compatible.endpoint",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiText},
				DisplayName:  "End Point",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "s3Compatible.secret.name",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SignatureVersion specifies the client signature version to use when signing requests.",
				Path:         "s3Compatible.signatureVersion",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiNumber},
				DisplayName:  "Signature Version",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target S3 bucket.",
				Path:         "s3Compatible.targetBucket",
				XDescriptors: []string{uiFieldGroupS3Compatible, uiText},
				DisplayName:  "Target Bucket",
			},
			operv1.SpecDescriptor{
				Description:  "Endpoint is the IBM COS endpoint: http(s)://host:port.",
				Path:         "IBMCos.endpoint",
				XDescriptors: []string{uiFieldGroupIBMCos, uiText},
				DisplayName:  "End Point",
			},
			operv1.SpecDescriptor{
				Description:  "Secret refers to a secret that provides the credentials. The secret should define IBM_COS_ACCESS_KEY_ID and IBM_COS_SECRET_ACCESS_KEY.",
				Path:         "IBMCos.secret.name",
				XDescriptors: []string{uiFieldGroupIBMCos, uiK8sSecret},
				DisplayName:  "Secret",
			},
			operv1.SpecDescriptor{
				Description:  "SignatureVersion specifies the client signature version to use when signing requests.",
				Path:         "IBMCos.signatureVersion",
				XDescriptors: []string{uiFieldGroupIBMCos, uiNumber},
				DisplayName:  "Signature Version",
			},
			operv1.SpecDescriptor{
				Description:  "TargetBucket is the name of the target IBM COS bucket.",
				Path:         "IBMCos.targetBucket",
				XDescriptors: []string{uiFieldGroupIBMCos, uiText},
				DisplayName:  "Target Bucket",
			},
		},

		"BucketClass": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{
				Description:  "BackingStores is an unordered list of backing store names. The meaning of the list depends on the placement.",
				Path:         "placementPolicy.tiers[0].backingStores[0]",
				XDescriptors: []string{uiFieldGroupPlacementPolicy, uiText},
				DisplayName:  "Backing Stores",
			},
			operv1.SpecDescriptor{
				Description:  "Placement specifies the type of placement for the tier If empty it should have a single backing store.",
				Path:         "placementPolicy.tiers[0].placement",
				XDescriptors: []string{uiFieldGroupPlacementPolicy, uiText},
				DisplayName:  "Placement",
			},
			operv1.SpecDescriptor{
				Description:  "Namespace Policy type specifies the type of the namespace policy configuration.",
				Path:         "namespacePolicy.type",
				XDescriptors: []string{uiFieldGroupNamespacePolicy, uiText},
				DisplayName:  "Namespace Policy",
			},
			operv1.SpecDescriptor{
				Description:  "Resource specifies the namespace store configured by the bucket class to be read and write targets.",
				Path:         "namespacePolicy.single.resource",
				XDescriptors: []string{uiFieldGroupNamespacePolicy, uiText},
				DisplayName:  "Resource",
			},
			operv1.SpecDescriptor{
				Description:  "Read Resources specifies the namespace stores configured by the bucket class to be read targets.",
				Path:         "namespacePolicy.Multi.readResources",
				XDescriptors: []string{uiFieldGroupNamespacePolicy, uiText},
				DisplayName:  "Read Resources",
			},
			operv1.SpecDescriptor{
				Description:  "Write Resource specifies the namespace store configured by the bucket class to be write target.",
				Path:         "namespacePolicy.Multi.writeResource",
				XDescriptors: []string{uiFieldGroupNamespacePolicy, uiText},
				DisplayName:  "Write Resource",
			},
			operv1.SpecDescriptor{
				Description:  "Hub Resource specifies the target namespace store configured by the bucket class to be read and write targets.",
				Path:         "namespacePolicy.Cache.hubResource",
				XDescriptors: []string{uiFieldGroupNamespacePolicy, uiText},
				DisplayName:  "Hub Resource",
			},
		},

		"ObjectBucketClaim": []operv1.SpecDescriptor{},
		"ObjectBucket":      []operv1.SpecDescriptor{},
	}

	crd.ForEachCRD(func(c *crd.CRD) {
		crdDesc := operv1.CRDDescription{
			Name:            c.Name,
			Kind:            c.Spec.Names.Kind,
			Version:         c.Spec.Versions[0].Name,
			DisplayName:     crdDisplayNames[c.Spec.Names.Kind],
			Description:     crdDescriptions[c.Spec.Names.Kind],
			SpecDescriptors: crdSpecDescriptors[c.Spec.Names.Kind],
			Resources: []operv1.APIResourceReference{
				operv1.APIResourceReference{Name: "services", Kind: "Service", Version: "v1"},
				operv1.APIResourceReference{Name: "secrets", Kind: "Secret", Version: "v1"},
				operv1.APIResourceReference{Name: "configmaps", Kind: "ConfigMap", Version: "v1"},
				operv1.APIResourceReference{Name: "statefulsets.apps", Kind: "StatefulSet", Version: "v1"},
			},
		}
		if c.Spec.Group == nbv1.SchemeGroupVersion.Group {
			csv.Spec.CustomResourceDefinitions.Owned = append(csv.Spec.CustomResourceDefinitions.Owned, crdDesc)
		} else {
			csv.Spec.CustomResourceDefinitions.Required = append(csv.Spec.CustomResourceDefinitions.Required, crdDesc)
		}
	})

	aw := util.KubeObject(bundle.File_deploy_internal_admission_webhook_yaml).(*admissionv1.ValidatingWebhookConfiguration)
	vaw := aw.Webhooks[0]

	webhookDefinition := operv1.WebhookDescription{
		Type:                    operv1.ValidatingAdmissionWebhook,
		AdmissionReviewVersions: vaw.AdmissionReviewVersions,
		ContainerPort:           443,
		TargetPort: &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 8080,
			StrVal: "8080",
		},
		DeploymentName: "noobaa-operator",
		FailurePolicy:  vaw.FailurePolicy,
		MatchPolicy:    vaw.MatchPolicy,
		GenerateName:   vaw.Name,
		Rules:          vaw.Rules,
		SideEffects:    vaw.SideEffects,
		WebhookPath:    vaw.ClientConfig.Service.Path,
	}
	csv.Spec.WebhookDefinitions = append(csv.Spec.WebhookDefinitions, webhookDefinition)
	return csv
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
