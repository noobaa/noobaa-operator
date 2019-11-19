package olm

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/crd"
	"github.com/noobaa/noobaa-operator/v2/pkg/operator"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/version"

	"github.com/blang/semver"
	operv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/printers"
	sigyaml "sigs.k8s.io/yaml"
)

type unObj = map[string]interface{}
type unArr = []interface{}

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
	}
	cmd.Flags().String("dir", "./build/_output/olm", "The output dir for the OLM package")
	return cmd
}

// CmdCSV returns a CLI command
func CmdCSV() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "Print CSV yaml",
		Run:   RunCSV,
	}
	return cmd
}

// CmdHubInstall returns a CLI command
func CmdHubInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator from operatorhub.io",
		Run:   RunHubInstall,
	}
	return cmd
}

// CmdHubUninstall returns a CLI command
func CmdHubUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator from operatorhub.io",
		Run:   RunHubUninstall,
	}
	return cmd
}

// CmdHubStatus returns a CLI command
func CmdHubStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of noobaa-operator from operatorhub.io",
		Run:   RunHubStatus,
	}
	return cmd
}

// CmdLocalInstall returns a CLI command
func CmdLocalInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-install",
		Short: "Install noobaa-operator from local OLM catalog",
		Run:   RunLocalInstall,
	}
	return cmd
}

// CmdLocalUninstall returns a CLI command
func CmdLocalUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-uninstall",
		Short: "Uninstall noobaa-operator from local OLM catalog",
		Run:   RunLocalUninstall,
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
	versionDir := dir + version.Version + "/"

	opConf := operator.LoadOperatorConf(cmd)

	pkgBytes, err := sigyaml.Marshal(unObj{
		"packageName":    "noobaa-operator",
		"defaultChannel": "alpha",
		"channels": unArr{unObj{
			"name":       "alpha",
			"currentCSV": "noobaa-operator.v" + version.Version,
		}},
	})
	util.Panic(err)

	util.Panic(os.MkdirAll(versionDir, 0755))
	util.Panic(ioutil.WriteFile(dir+"noobaa-operator.package.yaml", pkgBytes, 0644))
	util.Panic(util.WriteYamlFile(versionDir+"noobaa-operator.v"+version.Version+".clusterserviceversion.yaml", GenerateCSV(opConf)))
	crd.ForEachCRD(func(c *crd.CRD) {
		if c.Spec.Group == nbv1.SchemeGroupVersion.Group {
			util.Panic(util.WriteYamlFile(versionDir+c.Name+".crd.yaml", c))
		}
	})
}

// RunCSV runs a CLI command
func RunCSV(cmd *cobra.Command, args []string) {
	opConf := operator.LoadOperatorConf(cmd)
	csv := GenerateCSV(opConf)
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(csv, os.Stdout))
}

// GenerateCSV creates the CSV
func GenerateCSV(opConf *operator.Conf) *operv1.ClusterServiceVersion {

	install := &unstructured.Unstructured{Object: unObj{
		"deployments":        unArr{unObj{"name": opConf.Deployment.Name, "spec": opConf.Deployment.Spec}},
		"permissions":        unArr{unObj{"serviceAccountName": opConf.SA.Name, "rules": opConf.Role.Rules}},
		"clusterPermissions": unArr{unObj{"serviceAccountName": opConf.SA.Name, "rules": opConf.ClusterRole.Rules}},
	}}
	installRaw, err := install.MarshalJSON()
	util.Panic(err)

	almExamples, err := json.Marshal([]runtime.Object{
		util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml),
		util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml),
	})
	util.Panic(err)

	o := util.KubeObject(bundle.File_deploy_olm_noobaa_operator_clusterserviceversion_yaml)
	csv := o.(*operv1.ClusterServiceVersion)
	csv.Name = "noobaa-operator.v" + version.Version
	csv.Namespace = options.Namespace
	csv.Annotations["containerImage"] = options.OperatorImage
	// csv.Annotations["createdAt"] = ???
	csv.Annotations["alm-examples"] = string(almExamples)
	csv.Spec.Version.Version = semver.MustParse(version.Version)
	csv.Spec.Description = bundle.File_deploy_olm_description_md
	csv.Spec.Icon[0].Data = bundle.File_deploy_olm_noobaa_icon_base64
	csv.Spec.InstallStrategy.StrategySpecRaw = installRaw
	csv.Spec.CustomResourceDefinitions.Owned = []operv1.CRDDescription{}
	csv.Spec.CustomResourceDefinitions.Required = []operv1.CRDDescription{}
	crdDescriptions := map[string]string{
		"NooBaa": `A NooBaa system - Create this to start`,
		"BackingStore": `Storage target spec such as aws-s3, s3-compatible, PV's and more. ` +
			`Used in BacketClass to construct data placement policies.`,
		"BucketClass": `Storage policy spec  tiering, mirroring, spreading. ` +
			`Combines BackingStores. Referenced by ObjectBucketClaims.`,
		"ObjectBucketClaim": `Claim a bucket just like claiming a PV. ` +
			`Automate you app bucket provisioning by creating OBC with your app deployment. ` +
			`A secret and configmap (name=claim) will be created with access details for the app pods.`,
		"ObjectBucket": `Used under-the-hood. Created per ObjectBucketClaim and keeps provisioning information.`,
	}
	uiText := []string{"urn:alm:descriptor:com.tectonic.ui:text"}
	uiResources := []string{"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	crdSpecDescriptors := map[string][]operv1.SpecDescriptor{
		"NooBaa": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{Path: "image", XDescriptors: uiText},
			operv1.SpecDescriptor{Path: "dbImage", XDescriptors: uiText},
			operv1.SpecDescriptor{Path: "coreResources", XDescriptors: uiResources},
			operv1.SpecDescriptor{Path: "dbResources", XDescriptors: uiResources},
			operv1.SpecDescriptor{Path: "dbVolumeResources", XDescriptors: uiResources},
			operv1.SpecDescriptor{Path: "dbStorageClass", XDescriptors: uiText},
			operv1.SpecDescriptor{Path: "pvPoolDefaultStorageClass", XDescriptors: uiText},
			operv1.SpecDescriptor{Path: "tolerations", XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Tolerations"}},
			operv1.SpecDescriptor{Path: "imagePullSecret", XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret"}},
		},
		"BackingStore": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{Description: "Region is the AWS region.",
				Path:         "awsS3.region",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:awsS3", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Region"},
			operv1.SpecDescriptor{Description: "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "awsS3.secret",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:awsS3", "urn:alm:descriptor:com.tectonic.ui:k8s:secret"},
				DisplayName:  "Secret"},
			operv1.SpecDescriptor{Description: "SSLDisabled allows to disable SSL and use plain http.",
				Path:         "awsS3.sslDisabled",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:awsS3", "urn:alm:descriptor:com.tectonic.ui:k8s:booleanSwitch"},
				DisplayName:  "SSL Disabled"},
			operv1.SpecDescriptor{Description: "TargetBucket is the name of the target S3 bucket.",
				Path:         "awsS3.targetBucket",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:awsS3", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Target Bucket"},
			operv1.SpecDescriptor{Description: " Secret refers to a secret that provides the credentials. The secret should define AccountName and AccountKey as provided\nby Azure Blob.",
				Path:         "azureBlob.secret",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:azureBlob", "urn:alm:descriptor:com.tectonic.ui:k8s:secret"},
				DisplayName:  "Secret"},
			operv1.SpecDescriptor{Description: "TargetBlobContainer is the name of the target Azure Blob container.",
				Path:         "azureBlob.targetBlobContainer",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:azureBlob", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Target Blob Container"},
			operv1.SpecDescriptor{Description: "Secret refers to a secret that provides the credentials. The secret should define GoogleServiceAccountPrivateKeyJson containing\nthe entire json string as provided by Google.",
				Path:         "googleCloudStorage.secret",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:googleCloudStorage", "urn:alm:descriptor:com.tectonic.ui:k8s:secret"},
				DisplayName:  "Secret"},
			operv1.SpecDescriptor{Description: "TargetBucket is the name of the target S3 bucket.",
				Path:         "googleCloudStorage.targetBucket",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:googleCloudStorage", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Target Bucket"},
			operv1.SpecDescriptor{Description: "NumVolumes is the number of volumes to allocate.",
				Path:         "pvPool.numVolumes",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:pvPool", "urn:alm:descriptor:com.tectonic.ui:number"},
				DisplayName:  "Num Volumes"},
			operv1.SpecDescriptor{Description: "VolumeResources represents the minimum resources each volume should have.",
				Path:         "pvPool.resources",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:pvPool", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Resources"},
			operv1.SpecDescriptor{Description: "StorageClass is the name of the storage class to use for the PV's.",
				Path:         "pvPool.storageClass",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:pvPool", "urn:alm:descriptor:com.tectonic.ui:number"},
				DisplayName:  "Storage Class"},
			operv1.SpecDescriptor{Description: "Endpoint is the S3 compatible endpoint: http(s)://host:port.",
				Path:         "s3Compatible.endpoint",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:s3Compatible", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "End Point"},
			operv1.SpecDescriptor{Description: "Secret refers to a secret that provides the credentials. The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.",
				Path:         "s3Compatible.secret",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:s3Compatible", "urn:alm:descriptor:com.tectonic.ui:k8s:secret"},
				DisplayName:  "Secret"},
			operv1.SpecDescriptor{Description: "SignatureVersion specifies the client signature version to use when signing requests.",
				Path:         "s3Compatible.signatureVersion",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:s3Compatible", "urn:alm:descriptor:com.tectonic.ui:k8s:number"},
				DisplayName:  "Signature Version"},
			operv1.SpecDescriptor{Description: "TargetBucket is the name of the target S3 bucket.",
				Path:         "s3Compatible.targetBucket",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:s3Compatible", "urn:alm:descriptor:com.tectonic.ui:k8s:text"},
				DisplayName:  "Target Bucket"},
		},
		"BucketClass": []operv1.SpecDescriptor{
			operv1.SpecDescriptor{Description: "BackingStores is an unordered list of backing store names. The meaning of the list depends on the placement.",
				Path:         "placementPolicy.tiers[0].backingStores[0]",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:placementPolicy", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Backing Stores"},
			operv1.SpecDescriptor{Description: "Placement specifies the type of placement for the tier If empty it should have a single backing store.",
				Path:         "placementPolicy.tiers[0].placement",
				XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:awsS3", "urn:alm:descriptor:com.tectonic.ui:text"},
				DisplayName:  "Placement"},
		},
		"ObjectBucketClaim": []operv1.SpecDescriptor{},
		"ObjectBucket":      []operv1.SpecDescriptor{},
	}
	for kind := range crdSpecDescriptors {
		for i := range crdSpecDescriptors[kind] {
			d := &crdSpecDescriptors[kind][i]
			d.DisplayName = d.Path
			d.Description = d.Path
		}
	}
	crd.ForEachCRD(func(c *crd.CRD) {
		crdDesc := operv1.CRDDescription{
			Name:            c.Name,
			Kind:            c.Spec.Names.Kind,
			Version:         c.Spec.Version,
			DisplayName:     c.Spec.Names.Kind,
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
