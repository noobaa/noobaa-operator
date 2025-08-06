package system

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage noobaa systems",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdDelete(),
		CmdStatus(),
		CmdSetDebugLevel(),
		CmdList(),
		CmdReconcile(),
		CmdYaml(),
	)
	return cmd
}

var ctx = context.TODO()

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a noobaa system",
		Run:   RunCreate,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("core-resources", "", "Core resources JSON")
	cmd.Flags().String("db-resources", "", "DB resources JSON")
	cmd.Flags().String("endpoint-resources", "", "Endpoint resources JSON")
	cmd.Flags().Bool("use-standalone-db", false, "Create NooBaa system with standalone DB (Legacy)")
	cmd.Flags().Bool("use-obc-cleanup-policy", false, "Create NooBaa system with obc cleanup policy")
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a noobaa system",
		Run:   RunDelete,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().Bool("cleanup_data", false, "clean object buckets")
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List noobaa systems",
		Run:   RunList,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa system",
		Run:   RunStatus,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdSetDebugLevel returns a CLI command
func CmdSetDebugLevel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-debug-level <level>",
		Short: "Sets core and endpoints debug level. level can be 'warn' or 0-5",
		Run:   RunSetDebugLevel,
		Args:  cobra.ExactArgs(1),
	}
	return cmd
}

// CmdReconcile returns a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "reconcile",
		Short:  "Runs a reconcile attempt like noobaa-operator",
		Run:    RunReconcile,
		Args:   cobra.NoArgs,
	}
	return cmd
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled noobaa yaml",
		Run:   RunYaml,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// LoadSystemDefaults loads a noobaa system CR from bundled yamls
// and apply's changes from CLI flags to the defaults.
func LoadSystemDefaults() *nbv1.NooBaa {
	sys := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	sys.Namespace = options.Namespace
	sys.Name = options.SystemName
	sys.Finalizers = []string{nbv1.GracefulFinalizer}
	dbType := options.DBType
	sys.Spec.DBType = nbv1.DBTypes(dbType)
	sys.Spec.DisableLoadBalancerService = options.DisableLoadBalancerService
	sys.Spec.DisableRoutes = options.DisableRoutes
	sys.Spec.ManualDefaultBackingStore = options.ManualDefaultBackingStore
	sys.Spec.LoadBalancerSourceSubnets.S3 = options.S3LoadBalancerSourceSubnets
	sys.Spec.LoadBalancerSourceSubnets.STS = options.STSLoadBalancerSourceSubnets

	LoadConfigMapFromFlags()

	if options.AutoscalerType != "" {
		sys.Spec.Autoscaler.AutoscalerType = nbv1.AutoscalerTypes(options.AutoscalerType)
	}
	if options.PrometheusNamespace != "" {
		sys.Spec.Autoscaler.PrometheusNamespace = options.PrometheusNamespace
	}
	if options.NooBaaImage != "" {
		image := options.NooBaaImage
		sys.Spec.Image = &image
	}
	if options.ImagePullSecret != "" {
		sys.Spec.ImagePullSecret = &corev1.LocalObjectReference{Name: options.ImagePullSecret}
	}
	if options.DBImage != "" {
		dbImage := options.DBImage
		sys.Spec.DBImage = &dbImage
	}
	if options.DBVolumeSizeGB != 0 {
		sys.Spec.DBVolumeResources = &corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: *resource.NewScaledQuantity(int64(options.DBVolumeSizeGB), resource.Giga),
			},
		}
	}
	if options.DBStorageClass != "" {
		sc := options.DBStorageClass
		sys.Spec.DBStorageClass = &sc
	}
	if options.PostgresDbURL != "" {
		sys.Spec.ExternalPgSecret = &corev1.SecretReference{
			Name:      "noobaa-external-pg-db",
			Namespace: sys.Namespace,
		}

		sys.Spec.ExternalPgSSLRequired = options.PostgresSSLRequired
		sys.Spec.ExternalPgSSLUnauthorized = options.PostgresSSLSelfSigned

		if options.PostgresSSLCert != "" && options.PostgresSSLKey != "" {
			sys.Spec.ExternalPgSSLSecret = &corev1.SecretReference{
				Name:      "noobaa-external-db-cert",
				Namespace: sys.Namespace,
			}
		}
	}

	if options.PVPoolDefaultStorageClass != "" {
		sc := options.PVPoolDefaultStorageClass
		sys.Spec.PVPoolDefaultStorageClass = &sc
	}
	if options.MiniEnv {
		coreResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(1), resource.Giga),
		}
		logResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(50), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(200), resource.Mega),
		}
		dbResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(500), resource.Mega),
		}
		endpointResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(100), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(500), resource.Mega),
		}
		sys.Spec.CoreResources = &corev1.ResourceRequirements{
			Requests: coreResourceList,
			Limits:   coreResourceList,
		}
		sys.Spec.LogResources = &corev1.ResourceRequirements{
			Requests: logResourceList,
			Limits:   logResourceList,
		}
		sys.Spec.DBResources = &corev1.ResourceRequirements{
			Requests: dbResourceList,
			Limits:   dbResourceList,
		}
		sys.Spec.Endpoints = &nbv1.EndpointsSpec{
			MinCount: 1,
			MaxCount: 1,
			Resources: &corev1.ResourceRequirements{
				Requests: endpointResourceList,
				Limits:   endpointResourceList,
			}}
	}
	if options.DevEnv {
		coreResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(500), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(1), resource.Giga),
		}
		dbResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(1000), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(2), resource.Giga),
		}
		endpointResourceList := corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(int64(500), resource.Milli),
			corev1.ResourceMemory: *resource.NewScaledQuantity(int64(500), resource.Mega),
		}
		sys.Spec.CoreResources = &corev1.ResourceRequirements{
			Requests: coreResourceList,
			Limits:   coreResourceList,
		}
		sys.Spec.DBResources = &corev1.ResourceRequirements{
			Requests: dbResourceList,
			Limits:   dbResourceList,
		}
		sys.Spec.Endpoints = &nbv1.EndpointsSpec{
			MinCount: 1,
			MaxCount: 1,
			Resources: &corev1.ResourceRequirements{
				Requests: endpointResourceList,
				Limits:   endpointResourceList,
			},
		}
	}
	for _, componentName := range []string{"core", "db", "endpoints"} {
		if viper.IsSet(fmt.Sprintf("resources.%s", componentName)) {
			var component *corev1.ResourceRequirements

			switch componentName {
			case "core":
				if sys.Spec.CoreResources == nil {
					sys.Spec.CoreResources = &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{},
						Limits:   corev1.ResourceList{},
					}
				}
				component = sys.Spec.CoreResources
			case "db":
				if sys.Spec.DBResources == nil {
					sys.Spec.DBResources = &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{},
						Limits:   corev1.ResourceList{},
					}
				}
				component = sys.Spec.DBResources
			case "endpoints":
				if sys.Spec.Endpoints == nil {
					sys.Spec.Endpoints = &nbv1.EndpointsSpec{
						MinCount: viper.GetInt32("resources.endpoints.minCount"),
						MaxCount: viper.GetInt32("resources.endpoints.maxCount"),
						Resources: &corev1.ResourceRequirements{
							Requests: corev1.ResourceList{},
							Limits:   corev1.ResourceList{},
						},
					}
				} else {
					if viper.IsSet("resources.endpoints.minCount") {
						sys.Spec.Endpoints.MinCount = viper.GetInt32("resources.endpoints.minCount")
					}

					if viper.IsSet("resources.endpoints.maxCount") {
						sys.Spec.Endpoints.MaxCount = viper.GetInt32("resources.endpoints.maxCount")
					}
				}

				component = sys.Spec.Endpoints.Resources
			}

			if viper.IsSet(fmt.Sprintf("resources.%s.cpuMilli", componentName)) {
				component.Requests[corev1.ResourceCPU] = *resource.NewScaledQuantity(
					int64(viper.GetInt(fmt.Sprintf("resources.%s.cpuMilli", componentName))),
					resource.Milli,
				)

				component.Limits[corev1.ResourceCPU] = *resource.NewScaledQuantity(
					int64(viper.GetInt(fmt.Sprintf("resources.%s.cpuMilli", componentName))),
					resource.Milli,
				)
			}
			if viper.IsSet(fmt.Sprintf("resources.%s.memoryMB", componentName)) {
				component.Requests[corev1.ResourceMemory] = *resource.NewScaledQuantity(
					int64(viper.GetInt(fmt.Sprintf("resources.%s.memoryMB", componentName))),
					resource.Mega,
				)
				component.Limits[corev1.ResourceMemory] = *resource.NewScaledQuantity(
					int64(viper.GetInt(fmt.Sprintf("resources.%s.memoryMB", componentName))),
					resource.Mega,
				)
			}
		}
	}

	return sys
}

// RunOperatorCreate creates the default system when the operator starts.
func RunOperatorCreate(cmd *cobra.Command, args []string) {
	// The concern with this behavior was that it might be unexpected to the user that this happens.
	// For now we disable it but might reconsider later.
	//
	// sys := LoadSystemDefaults()
	// util.KubeCreateSkipExisting(sys)
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()
	sys := LoadSystemDefaults()
	ns := util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	ns.Name = sys.Namespace

	coreResourcesJSON, _ := cmd.Flags().GetString("core-resources")
	dbResourcesJSON, _ := cmd.Flags().GetString("db-resources")
	endpointResourcesJSON, _ := cmd.Flags().GetString("endpoint-resources")
	useOBCCleanupPolicy, _ := cmd.Flags().GetBool("use-obc-cleanup-policy")
	useStandaloneDB, _ := cmd.Flags().GetBool("use-standalone-db")
	useCNPG := !useStandaloneDB

	if useOBCCleanupPolicy {
		sys.Spec.CleanupPolicy.Confirmation = nbv1.DeleteOBCConfirmation
	}
	if coreResourcesJSON != "" {
		util.Panic(json.Unmarshal([]byte(coreResourcesJSON), &sys.Spec.CoreResources))
	}
	if dbResourcesJSON != "" {
		util.Panic(json.Unmarshal([]byte(dbResourcesJSON), &sys.Spec.DBResources))
	}
	if endpointResourcesJSON != "" {
		if sys.Spec.Endpoints == nil {
			sys.Spec.Endpoints = &nbv1.EndpointsSpec{
				MinCount: 1,
				MaxCount: 1,
			}
		}
		util.Panic(json.Unmarshal([]byte(endpointResourcesJSON), &sys.Spec.Endpoints.Resources))
	}

	if options.PostgresDbURL != "" {
		if (options.PostgresSSLCert != "" && options.PostgresSSLKey == "") ||
			(options.PostgresSSLCert == "" && options.PostgresSSLKey != "") {
			log.Fatalf("❌ Can't provide only ssl-cert or only ssl-key - please provide both!")
		}
		err := CheckPostgresURL(options.PostgresDbURL)
		if err != nil {
			log.Fatalf(`❌ %s`, err)
		}
		o := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
		secret := o.(*corev1.Secret)
		secret.Namespace = sys.Spec.ExternalPgSecret.Namespace
		secret.Name = sys.Spec.ExternalPgSecret.Name
		secret.StringData = map[string]string{
			"db_url": options.PostgresDbURL,
		}
		secret.Data = nil
		util.KubeCreateSkipExisting(secret)
		if sys.Spec.ExternalPgSSLSecret != nil {
			secretData := make(map[string][]byte)
			data, err := os.ReadFile(options.PostgresSSLKey)
			if err != nil {
				log.Fatalf("❌ Can't open key file %q please try again, error: %s", options.PostgresSSLKey, err)
			}
			secretData["tls.key"] = data
			data, err = os.ReadFile(options.PostgresSSLCert)
			if err != nil {
				log.Fatalf("❌ Can't open cert file %q please try again, error: %s", options.PostgresSSLKey, err)
			}
			secretData["tls.crt"] = data
			o := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
			secret := o.(*corev1.Secret)
			secret.Namespace = sys.Spec.ExternalPgSSLSecret.Namespace
			secret.Name = sys.Spec.ExternalPgSSLSecret.Name
			secret.Data = secretData
			util.KubeCreateSkipExisting(secret)
		}
	}

	if useCNPG {
		dbVolumeSize := ""
		if sys.Spec.DBVolumeResources != nil {
			dbVolumeSize = sys.Spec.DBVolumeResources.Requests.Storage().String()
		}
		sys.Spec.DBSpec = &nbv1.NooBaaDBSpec{
			DBImage:              sys.Spec.DBImage,
			PostgresMajorVersion: &options.PostgresMajorVersion,
			Instances:            &options.PostgresInstances,
			DBResources:          sys.Spec.DBResources,
			DBMinVolumeSize:      dbVolumeSize,
			DBStorageClass:       sys.Spec.DBStorageClass,
		}
	}

	// TODO check PVC if exist and the system does not exist -
	// fail and suggest to delete them first with cli system delete.
	util.KubeCreateSkipExisting(ns)
	util.KubeCreateSkipExisting(sys)
}

// RunUpgrade runs a CLI command
func RunUpgrade(cmd *cobra.Command, args []string) {

	sys := &nbv1.NooBaa{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		},
	}

	if !util.KubeCheck(sys) {
		log.Fatalf("NooBaa system %q not found", options.SystemName)
	}

	if options.NooBaaImage != "" {
		image := options.NooBaaImage
		sys.Spec.Image = &image
		sys.Namespace = options.Namespace
	}

	dbVolumeSize := ""
	if sys.Spec.DBVolumeResources != nil {
		dbVolumeSize = sys.Spec.DBVolumeResources.Requests.Storage().String()
	}

	// set dbSpec
	sys.Spec.DBSpec = &nbv1.NooBaaDBSpec{
		DBImage:              &options.DBImage,
		PostgresMajorVersion: &options.PostgresMajorVersion,
		Instances:            &options.PostgresInstances,
		DBResources:          sys.Spec.DBResources,
		DBMinVolumeSize:      dbVolumeSize,
		DBStorageClass:       sys.Spec.DBStorageClass,
	}
	util.KubeApply(sys)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()

	sys := &nbv1.NooBaa{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		},
	}

	util.KubeCheck(sys)

	log.Infof("Notice: Deletion of External secrets should be preformed manually")

	if err := SetAllowNoobaaDeletion(sys); err != nil {
		log.Errorf("NooBaa %q failed to update cleanup policy - deletion may fail", options.SystemName)
	}

	cleanupData, _ := cmd.Flags().GetBool("cleanup_data")
	objectBuckets := &nbv1.ObjectBucketList{}
	obcSelector, _ := labels.Parse("noobaa-domain=" + options.SubDomainNS())
	util.KubeList(objectBuckets, &client.ListOptions{LabelSelector: obcSelector})
	finalizersArray := sys.GetFinalizers()

	if util.Contains(finalizersArray, nbv1.GracefulFinalizer) {

		confirm := sys.Spec.CleanupPolicy.Confirmation
		if !cleanupData && len(objectBuckets.Items) != 0 && confirm != nbv1.DeleteOBCConfirmation {
			log.Fatalf(`❌ %s`, fmt.Sprintf("Failed to delete NooBaa. object buckets in namespace %q are not cleaned up.", options.Namespace))

		} else {
			log.Infof("Deleting All object buckets in namespace %q", options.Namespace)
			// deletion of OBCSC
			sc := &storagev1.StorageClass{}
			sc.Name = options.SubDomainNS()
			if err := util.DeleteStorageClass(sc); err != nil {
				log.Errorf("failed to delete storageclass %q", sc.Name)
			}
			util.RemoveFinalizer(sys, nbv1.GracefulFinalizer)
			if !util.KubeUpdate(sys) {
				log.Errorf("NooBaa %q failed to remove finalizer %q", options.SystemName, nbv1.GracefulFinalizer)
			}
		}
	}

	util.KubeDelete(sys)

	if sys.Spec.ExternalPgSecret != nil {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sys.Spec.ExternalPgSecret.Name,
				Namespace: sys.Spec.ExternalPgSecret.Namespace,
			},
		}
		util.KubeDelete(secret)
	}
	backingStores := &nbv1.BackingStoreList{}
	util.KubeList(backingStores, &client.ListOptions{Namespace: options.Namespace})
	for i := range backingStores.Items {
		bstore := &backingStores.Items[i]
		if bstore.Spec.Type == nbv1.StoreTypePVPool {
			pods := &corev1.PodList{}
			util.KubeList(pods, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name})
			for i := range pods.Items {
				pod := &pods.Items[i]
				util.RemoveFinalizer(pod, nbv1.Finalizer)
				if !util.KubeUpdate(pod) {
					log.Errorf("Pod %q failed to remove finalizer %q",
						pod.Name, nbv1.Finalizer)
				}
			}
			util.KubeDeleteAllOf(&corev1.Pod{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name},
				client.GracePeriodSeconds(0))
			util.KubeDeleteAllOf(&corev1.PersistentVolumeClaim{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": bstore.Name},
				client.GracePeriodSeconds(0))
		}
		util.RemoveFinalizer(bstore, nbv1.Finalizer)
		if !util.KubeUpdate(bstore) {
			log.Errorf("BackingStore %q failed to remove finalizer %q",
				bstore.Name, nbv1.Finalizer)
		}
		util.KubeDelete(bstore, client.GracePeriodSeconds(0))
	}
	objectBucketClaims := &nbv1.ObjectBucketClaimList{}
	configMaps := &corev1.ConfigMapList{}
	secrets := &corev1.SecretList{}

	util.KubeList(objectBucketClaims, &client.ListOptions{LabelSelector: obcSelector})
	util.KubeList(configMaps, &client.ListOptions{LabelSelector: obcSelector})
	util.KubeList(secrets, &client.ListOptions{LabelSelector: obcSelector})

	for i := range objectBucketClaims.Items {
		obj := &objectBucketClaims.Items[i]
		log.Warnf("ObjectBucketClaim %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucketClaim %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range objectBuckets.Items {
		obj := &objectBuckets.Items[i]
		log.Warnf("ObjectBucket %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range configMaps.Items {
		obj := &configMaps.Items[i]
		log.Warnf("ConfigMap %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	for i := range secrets.Items {
		obj := &secrets.Items[i]
		log.Warnf("Secret %q removing without grace", obj.Name)
		util.RemoveFinalizer(obj, nbv1.ObjectBucketFinalizer)
		if !util.KubeUpdate(obj) {
			log.Errorf("ObjectBucket %q failed to remove finalizer %q",
				obj.Name, nbv1.ObjectBucketFinalizer)
		}
		util.KubeDelete(obj, client.GracePeriodSeconds(0))
	}
	//delete keda resources
	deleteKedaResources(options.Namespace)
	//delete HPAV2 resources
	if err := deleteHPAV2Resources(options.Namespace); err != nil {
		log.Errorf("falied to delete HPAV2 Resources : %s", err)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaa"},
	}
	if !util.KubeList(list) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No systems found.\n")
		return
	}

	table := (&util.PrintTable{}).AddRow(
		"NAMESPACE",
		"NAME",
		"S3-ENDPOINTS",
		"STS-ENDPOINTS",
		"IMAGE",
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		s := &list.Items[i]
		table.AddRow(
			s.Namespace,
			s.Name,
			fmt.Sprint(s.Status.Services.ServiceS3.NodePorts),
			fmt.Sprint(s.Status.Services.ServiceSts.NodePorts),
			s.Status.ActualImage,
			string(s.Status.Phase),
			util.HumanizeDuration(time.Since(s.CreationTimestamp.Time).Round(time.Second)),
		)
	}
	fmt.Print(table.String())
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	sys := LoadSystemDefaults()
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(sys, os.Stdout))
}

// CheckNooBaaImages runs a CLI command
func CheckNooBaaImages(cmd *cobra.Command, sys *nbv1.NooBaa, args []string) string {
	log := util.Logger()

	desiredImage := ""
	runningImage := ""
	if sys.Status.ActualImage != "" {
		desiredImage = sys.Status.ActualImage
	} else if sys.Spec.Image != nil {
		desiredImage = *sys.Spec.DBImage
	}
	sts := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	sts.Namespace = options.Namespace
	if util.KubeCheckQuiet(sts) {
		runningImage = sts.Spec.Template.Spec.Containers[0].Image
	}
	if desiredImage != "" && runningImage != "" && desiredImage != runningImage {
		log.Warnf("⚠️  The desired noobaa image and the running image are not the same")
	}
	return runningImage
}

// CheckNooBaaDBImages runs a CLI command
func CheckNooBaaDBImages(cmd *cobra.Command, sys *nbv1.NooBaa, args []string) string {
	log := util.Logger()

	desiredImage := ""
	runningImage := ""
	if sys.Spec.DBImage != nil {
		desiredImage = *sys.Spec.DBImage
	}
	sts := util.KubeObject(bundle.File_deploy_internal_statefulset_postgres_db_yaml).(*appsv1.StatefulSet)
	sts.Name = "noobaa-db-pg"
	sts.Namespace = options.Namespace
	if util.KubeCheckQuiet(sts) {
		runningImage = sts.Spec.Template.Spec.Containers[0].Image
	}
	if desiredImage != "" && runningImage != "" && desiredImage != runningImage {
		log.Warnf("⚠️  The desired db image and the running db image are not the same")
	}
	return runningImage
}

// CheckOperatorImage runs a CLI command
func CheckOperatorImage(cmd *cobra.Command, args []string) string {
	runningImage := ""
	deployment := util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)
	deployment.Namespace = options.Namespace
	if util.KubeCheckQuiet(deployment) {
		runningImage = deployment.Spec.Template.Spec.Containers[0].Image
	}
	return runningImage
}

// RunSystemVersionsStatus runs a CLI command
func RunSystemVersionsStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace
	isSystemExists := util.KubeCheckQuiet(sys)

	noobaaImage := ""
	noobaaDbImage := ""
	noobaaOperatorImage := ""

	if isSystemExists {
		noobaaImage = CheckNooBaaImages(cmd, sys, args)
		if sys.Spec.ExternalPgSecret == nil {
			noobaaDbImage = CheckNooBaaDBImages(cmd, sys, args)
		}
		noobaaOperatorImage = CheckOperatorImage(cmd, args)
	} else {
		noobaaImage = options.NooBaaImage
		noobaaDbImage = options.DBImage
		noobaaOperatorImage = options.OperatorImage
	}

	log.Printf("CLI version: %s\n", version.Version)
	log.Printf("noobaa-image: %s\n", noobaaImage)
	log.Printf("operator-image: %s\n", noobaaOperatorImage)
	if options.PostgresDbURL == "" && sys.Spec.ExternalPgSecret == nil {
		log.Printf("noobaa-db-image: %s\n", noobaaDbImage)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysKey, klient, scheme.Scheme, nil)
	r.CheckAll()
	var NooBaaDB = r.NooBaaPostgresDB

	if r.shouldReconcileStandaloneDB() {
		// NoobaaDB
		for i := range NooBaaDB.Spec.VolumeClaimTemplates {
			t := &NooBaaDB.Spec.VolumeClaimTemplates[i]
			pvc := &corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      t.Name + "-" + NooBaaDB.Name + "-0",
					Namespace: options.Namespace,
				},
			}
			util.KubeCheck(pvc)
		}
	}

	// sys := cli.LoadSystemDefaults()
	// util.KubeCheck(cli.Client, sys)

	if r.NooBaa.Status.Phase != nbv1.SystemPhaseReady {
		log.Printf("❌ System Phase is %q\n", r.NooBaa.Status.Phase)
		util.IgnoreError(CheckWaitingFor(r.NooBaa))
		return
	}

	log.Printf("✅ System Phase is %q\n", r.NooBaa.Status.Phase)
	secret := r.SecretAdmin.DeepCopy()
	util.KubeCheck(secret)

	fmt.Println("")
	if options.ShowSecrets {
		fmt.Printf("%s password : %s\n", secret.StringData["email"], secret.StringData["password"])
	} else {
		fmt.Printf("%s password : %s\n", secret.StringData["email"], nb.MaskedString(secret.StringData["password"]))
	}

	fmt.Println("")
	fmt.Println("#-----------------#")
	fmt.Println("#- STS Addresses -#")
	fmt.Println("#-----------------#")
	fmt.Println("")

	util.PrettyPrint("ExternalDNS", r.NooBaa.Status.Services.ServiceSts.ExternalDNS)
	fmt.Println("ExternalIP  :", r.NooBaa.Status.Services.ServiceSts.ExternalIP)
	fmt.Println("NodePorts   :", r.NooBaa.Status.Services.ServiceSts.NodePorts)
	fmt.Println("InternalDNS :", r.NooBaa.Status.Services.ServiceSts.InternalDNS)
	fmt.Println("InternalIP  :", r.NooBaa.Status.Services.ServiceSts.InternalIP)
	fmt.Println("PodPorts    :", r.NooBaa.Status.Services.ServiceSts.PodPorts)

	fmt.Println("")
	fmt.Println("#----------------#")
	fmt.Println("#- S3 Addresses -#")
	fmt.Println("#----------------#")
	fmt.Println("")

	util.PrettyPrint("ExternalDNS", r.NooBaa.Status.Services.ServiceS3.ExternalDNS)
	fmt.Println("ExternalIP  :", r.NooBaa.Status.Services.ServiceS3.ExternalIP)
	fmt.Println("NodePorts   :", r.NooBaa.Status.Services.ServiceS3.NodePorts)
	fmt.Println("InternalDNS :", r.NooBaa.Status.Services.ServiceS3.InternalDNS)
	fmt.Println("InternalIP  :", r.NooBaa.Status.Services.ServiceS3.InternalIP)
	fmt.Println("PodPorts    :", r.NooBaa.Status.Services.ServiceS3.PodPorts)

	fmt.Println("")
	fmt.Println("#------------------#")
	fmt.Println("#- S3 Credentials -#")
	fmt.Println("#------------------#")
	fmt.Println("")
	if options.ShowSecrets {
		fmt.Printf("AWS_ACCESS_KEY_ID     : %s\n", secret.StringData["AWS_ACCESS_KEY_ID"])
		fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n", secret.StringData["AWS_SECRET_ACCESS_KEY"])
	} else {
		fmt.Printf("AWS_ACCESS_KEY_ID     : %s\n", nb.MaskedString(secret.StringData["AWS_ACCESS_KEY_ID"]))
		fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n", nb.MaskedString(secret.StringData["AWS_SECRET_ACCESS_KEY"]))
	}
	fmt.Println("")

}

// RunSetDebugLevel sets the system debug level
func RunSetDebugLevel(cmd *cobra.Command, args []string) {
	log := util.Logger()
	level := 0
	if args[0] == "warn" {
		level = -1
	} else {
		var err error
		level, err = strconv.Atoi(args[0])
		if err != nil || level < 0 || level > 5 {
			log.Fatalf(`invalid debug-level argument. must be 'warn' or 0 to 5. %s`, cmd.UsageString())
		}
	}
	nbClient := GetNBClient()
	err := nbClient.PublishToCluster(nb.PublishToClusterParams{
		Target:     "",
		MethodAPI:  "debug_api",
		MethodName: "set_debug_level",
		RequestParams: nb.SetDebugLevelParams{
			Module: "core",
			Level:  level,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	fmt.Printf("Debug level was set to %s successfully\n", args[0])
	fmt.Println("Debug level is not persistent and is only effective for the currently running core and endpoints pods")
}

// RunReconcile runs a CLI command
func RunReconcile(cmd *cobra.Command, args []string) {
	log := util.Logger()
	klient := util.KubeClient()
	interval := time.Duration(3)
	util.Panic(wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      options.SystemName,
			},
		}
		res, err := NewReconciler(req.NamespacedName, klient, scheme.Scheme, nil).Reconcile()
		if err != nil {
			return false, err
		}
		if res.RequeueAfter != 0 {
			log.Printf("\nRetrying in %d seconds\n", interval)
			return false, nil
		}
		return true, nil
	}))
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady() bool {
	log := util.Logger()
	klient := util.KubeClient()

	sysKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	interval := time.Duration(3)

	err := wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		sys := &nbv1.NooBaa{}
		err := klient.Get(util.Context(), sysKey, sys)
		if err != nil {
			log.Printf("⏳ Failed to get system: %s", err)
			return false, nil
		}
		if sys.Status.Phase == nbv1.SystemPhaseRejected {
			log.Errorf("❌ System Phase is %q. describe noobaa for more information", sys.Status.Phase)
			return false, fmt.Errorf("SystemPhaseRejected")
		}
		if sys.Status.Phase != nbv1.SystemPhaseReady {
			return false, CheckWaitingFor(sys)
		}
		log.Printf("✅ System Phase is %q.\n", sys.Status.Phase)
		return true, nil
	})
	return (err == nil)
}

// CheckWaitingFor checks what the system deployment is waiting for in order to become ready
// in order to help the user troubleshoot common deployment issues.
func CheckWaitingFor(sys *nbv1.NooBaa) error {
	log := util.Logger()
	klient := util.KubeClient()

	operAppReady, operAppErr := CheckDeploymentReady(sys, "noobaa-operator", "noobaa-operator=deployment")
	if !operAppReady {
		return operAppErr
	}

	if sys.Spec.DBSpec != nil && sys.Spec.ExternalPgSecret == nil {

		cnpgOperReady, cnpgOperErr := CheckDeploymentReady(sys, cnpg.CnpgDeploymentName, "app.kubernetes.io/name=cloudnative-pg")
		if !cnpgOperReady {
			return cnpgOperErr
		}

		cnpgCluster := cnpg.GetCnpgClusterObj(sys.Namespace, sys.Name+pgClusterSuffix)
		cnpgClusterErr := klient.Get(util.Context(),
			client.ObjectKey{Namespace: sys.Namespace, Name: cnpgCluster.Name},
			cnpgCluster)
		if errors.IsNotFound(cnpgClusterErr) {
			log.Printf(`⏳ System Phase is %q. CNPG Cluster %q is not found yet`,
				sys.Status.Phase, cnpgCluster.Name)
			return nil
		}
		if cnpgClusterErr != nil {
			log.Printf(`⏳ System Phase is %q. CNPG Cluster %q is not found yet (error): %s`,
				sys.Status.Phase, cnpgCluster.Name, cnpgClusterErr)
			return nil
		}
		if !isClusterReady(cnpgCluster) {
			log.Printf(`⏳ System Phase is %q. CNPG Cluster %q is not ready: %s`,
				sys.Status.Phase, cnpgCluster.Name, cnpgCluster.Status.Phase)
			return nil
		}
	}

	coreApp := &appsv1.StatefulSet{}
	coreAppName := sys.Name + "-core"
	coreAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: coreAppName},
		coreApp)

	if errors.IsNotFound(coreAppErr) {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not found yet`,
			sys.Status.Phase, coreAppName)
		return nil
	}
	if coreAppErr != nil {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not found yet (error): %s`,
			sys.Status.Phase, coreAppName, coreAppErr)
		return nil
	}
	desiredReplicas := int32(1)
	if coreApp.Spec.Replicas != nil {
		desiredReplicas = *coreApp.Spec.Replicas
	}
	if coreApp.Status.Replicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. StatefulSet %q is not yet ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			coreAppName,
			coreApp.Status.ReadyReplicas,
			desiredReplicas)
		return nil
	}

	corePodList := &corev1.PodList{}
	corePodSelector, _ := labels.Parse("noobaa-core=" + sys.Name)
	corePodErr := klient.List(util.Context(),
		corePodList,
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: corePodSelector})

	if corePodErr != nil {
		return corePodErr
	}
	if len(corePodList.Items) != int(desiredReplicas) {
		return fmt.Errorf("Can't find the core pods")
	}
	corePod := &corePodList.Items[0]
	if corePod.Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not yet ready: %s`,
			sys.Status.Phase, corePod.Name, util.GetPodStatusLine(corePod))
		return nil
	}
	for i := range corePod.Status.ContainerStatuses {
		c := &corePod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q is not yet ready: %s`,
				sys.Status.Phase, c.Name, util.GetContainerStatusLine(c))
			return nil
		}
	}

	log.Printf(`⏳ System Phase is %q. Waiting for phase ready ...`, sys.Status.Phase)
	return nil
}

// CheckDeployment checks if the deployment is ready
func CheckDeploymentReady(sys *nbv1.NooBaa, operAppName string, listLabel string) (bool, error) {

	log := util.Logger()
	klient := util.KubeClient()

	operApp := &appsv1.Deployment{}
	operAppErr := klient.Get(util.Context(),
		client.ObjectKey{Namespace: sys.Namespace, Name: operAppName},
		operApp)

	if errors.IsNotFound(operAppErr) {
		log.Printf(`❌ Deployment %q is missing.`, operAppName)
		return false, operAppErr
	}
	if operAppErr != nil {
		log.Printf(`❌ Deployment %q unknown error in Get(): %s`, operAppName, operAppErr)
		return false, operAppErr
	}
	desiredReplicas := int32(1)
	if operApp.Spec.Replicas != nil {
		desiredReplicas = *operApp.Spec.Replicas
	}
	if operApp.Status.ReadyReplicas != desiredReplicas {
		log.Printf(`⏳ System Phase is %q. Deployment %q is not ready:`+
			` ReadyReplicas %d/%d`,
			sys.Status.Phase,
			operAppName,
			operApp.Status.ReadyReplicas,
			desiredReplicas)
		return false, nil
	}

	operPodList := &corev1.PodList{}
	operPodSelector, _ := labels.Parse(listLabel)
	operPodErr := klient.List(util.Context(),
		operPodList,
		&client.ListOptions{Namespace: sys.Namespace, LabelSelector: operPodSelector})

	if operPodErr != nil {
		return false, operPodErr
	}
	if len(operPodList.Items) != int(desiredReplicas) {
		return false, fmt.Errorf("Can't find the operator pods")
	}
	operPod := &operPodList.Items[0]
	if operPod.Status.Phase != corev1.PodRunning {
		log.Printf(`⏳ System Phase is %q. Pod %q is not yet ready: %s`,
			sys.Status.Phase, operPod.Name, util.GetPodStatusLine(operPod))
		return false, nil
	}
	for i := range operPod.Status.ContainerStatuses {
		c := &operPod.Status.ContainerStatuses[i]
		if !c.Ready {
			log.Printf(`⏳ System Phase is %q. Container %q in pod %q is not yet ready: %s`,
				sys.Status.Phase, c.Name, operPod.Name, util.GetContainerStatusLine(c))
			return false, nil
		}
	}
	return true, nil
}

// Client is the system client for making mgmt or s3 calls (with operator/admin credentials)
type Client struct {
	NooBaa      *nbv1.NooBaa
	ServiceMgmt *corev1.Service
	SecretOp    *corev1.Secret
	SecretAdmin *corev1.Secret
	NBClient    nb.Client
	MgmtURL     *url.URL
	S3URL       *url.URL
}

// GetNBClient returns an api client
func GetNBClient() nb.Client {
	c, err := Connect(true)
	if err != nil {
		util.Logger().Fatalf("❌ %s", err)
	}
	return c.NBClient
}

// Connect loads the mgmt and S3 api details from the system.
// It gets the endpoint address and token from the system status and secret that the
// operator creates for the system.
// When isExternal is true we return  : s3 => external DNS, mgmt => port-forwarding (router)
// When isExternal is false we return : s3 => internal DNS, mgmt => node-port
func Connect(isExternal bool) (*Client, error) {

	klient := util.KubeClient()
	sysObjKey := client.ObjectKey{Namespace: options.Namespace, Name: options.SystemName}
	r := NewReconciler(sysObjKey, klient, scheme.Scheme, nil)

	if !CheckSystem(r.NooBaa) {
		return nil, fmt.Errorf("Connect(): System not found")
	}
	if !util.KubeCheck(r.ServiceMgmt) {
		return nil, fmt.Errorf("Connect(): ServiceMgmt not found")
	}
	if !util.KubeCheck(r.SecretOp) {
		return nil, fmt.Errorf("Connect(): SecretOp not found")
	}
	if !util.KubeCheck(r.SecretAdmin) {
		return nil, fmt.Errorf("Connect(): SecretAdmin not found")
	}

	authToken := r.SecretOp.StringData["auth_token"]
	mgmtStatus := &r.NooBaa.Status.Services.ServiceMgmt
	s3Status := &r.NooBaa.Status.Services.ServiceS3

	if authToken == "" {
		return nil, fmt.Errorf("Connect(): Operator secret with auth token is not ready")
	}
	if len(mgmtStatus.NodePorts) == 0 {
		return nil, fmt.Errorf("Connect(): System mgmt service (NodePorts) is not ready")
	}
	if len(s3Status.InternalDNS) == 0 {
		return nil, fmt.Errorf("Connect(): System s3 service (InternalDNS) is not ready")
	}

	mgmtEndpoint := mgmtStatus.NodePorts[0]
	s3Endpoint := s3Status.InternalDNS[0]
	var nbClient nb.Client

	if isExternal {

		// setup port forwarding
		router := &nb.APIRouterPortForward{
			ServiceMgmt:  r.ServiceMgmt,
			PodNamespace: r.NooBaa.Namespace,
			PodName:      r.NooBaa.Name + "-core-0",
		}
		err := router.Start()
		if err != nil {
			return nil, err
		}
		nbClient = nb.NewClient(router)
		mgmtEndpoint = strings.TrimSuffix(router.GetAddress(""), "/rpc/")

		// set s3 to external if possible, otherwise fallback to node-port which is like external for minikube
		if len(s3Status.ExternalIP) > 0 {
			s3Endpoint = s3Status.ExternalIP[0]
		} else if len(s3Status.NodePorts) > 0 {
			s3Endpoint = s3Status.NodePorts[0]
		}

	} else {
		nbClient = nb.NewClient(&nb.APIRouterServicePort{
			ServiceMgmt: r.ServiceMgmt,
		})
	}

	nbClient.SetAuthToken(authToken)

	mgmtURL, err := url.Parse(mgmtEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Connect(): Failed to parse mgmt url %q. got error: %v", mgmtEndpoint, err)
	}
	s3URL, err := url.Parse(s3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Connect(): Failed to parse s3 url %q. got error: %v", s3Endpoint, err)
	}

	return &Client{
		NooBaa:      r.NooBaa,
		ServiceMgmt: r.ServiceMgmt,
		SecretOp:    r.SecretOp,
		SecretAdmin: r.SecretAdmin,
		NBClient:    nbClient,
		MgmtURL:     mgmtURL,
		S3URL:       s3URL,
	}, nil
}

// GetDesiredDBImage returns the desired DB image according to spec or env or default (in options)
func GetDesiredDBImage(sys *nbv1.NooBaa, currentImage string) string {
	// Postgres upgrade failure workaround
	// if the current Postgres image is a postgresql-12 image, use NOOBAA_PSQL12_IMAGE. otherwise use GetdesiredDBImage
	if IsPostgresql12Image(currentImage) {
		psql12Image, ok := os.LookupEnv("NOOBAA_PSQL_12_IMAGE")
		util.Logger().Warnf("The current Postgres image is a postgresql-12 image. using (%s)", psql12Image)
		if !ok {
			psql12Image = currentImage
			util.Logger().Warnf("NOOBAA_PSQL_12_IMAGE is not set. using the current image %s", currentImage)
		}
		return psql12Image
	}

	if sys.Spec.DBImage != nil {
		return *sys.Spec.DBImage
	}

	if os.Getenv("NOOBAA_DB_IMAGE") != "" {
		return os.Getenv("NOOBAA_DB_IMAGE")
	}

	return options.DBImage
}

// IsPostgresql12Image checks if the image is a postgresql-12 image
func IsPostgresql12Image(image string) bool {
	return strings.Contains(image, "postgresql-12")
}

// CheckSystem checks the state of the system and initializes its status fields
func CheckSystem(sys *nbv1.NooBaa) bool {
	log := util.Logger()
	found := util.KubeCheck(sys)

	if ts := sys.GetDeletionTimestamp(); ts != nil {
		log.Printf("❌ NooBaa system deleted at %v", ts)
		return false // deleted
	}

	if util.EnsureCommonMetaFields(sys, nbv1.GracefulFinalizer) {
		if !util.KubeUpdate(sys) {
			log.Errorf("❌ NooBaa %q failed to add mandatory meta fields", sys.Name)
			return false
		}
	}

	if sys.Status.Accounts == nil {
		sys.Status.Accounts = &nbv1.AccountsStatus{}
	}
	if sys.Status.Services == nil {
		sys.Status.Services = &nbv1.ServicesStatus{}
	}

	return found && sys.UID != ""
}

// CheckPostgresURL checks if the postgresurl structure is valid and if we use postgres as db
func CheckPostgresURL(postgresDbURL string) error {
	// This is temporary checks - In next PRs we will change to psql client checks instead
	u, err := url.Parse(postgresDbURL)
	if err != nil {
		return fmt.Errorf("failed parsing external DB url: %q, error: %s", postgresDbURL, err)
	}
	_, _, err = net.SplitHostPort(u.Host)
	if err != nil {
		return fmt.Errorf("failed splitting host and port from external DB url: %q", postgresDbURL)
	}
	if !strings.Contains(postgresDbURL, "postgres://") &&
		!strings.Contains(postgresDbURL, "postgresql://") {
		return fmt.Errorf("invalid postgres db url %s, expecting the url to start with postgres:// or postgresql://", postgresDbURL)
	}
	return nil
}

// LoadConfigMapFromFlags loads a config-map with values from the cli flags, if provided.
func LoadConfigMapFromFlags() {
	if options.DebugLevel != "default_level" {
		cm := util.KubeObject(bundle.File_deploy_internal_configmap_empty_yaml).(*corev1.ConfigMap)
		cm.Namespace = options.Namespace
		cm.Name = "noobaa-config"

		DefaultConfigMapData := map[string]string{
			"NOOBAA_LOG_LEVEL": options.DebugLevel,
		}

		cm.Data = DefaultConfigMapData

		util.KubeCreateSkipExisting(cm)
	}
}

// MapSecretToBackingStores returns a list of backingstores that uses the secret in their secretReference
// used by backingstore_controller to watch secrets changes
func MapSecretToNooBaa(secret types.NamespacedName) []reconcile.Request {
	log := util.Logger()
	log.Infof("checking which nooBaas to reconcile. mapping secret %v to nooBaas external postgres secret", secret)
	nbList := &nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaList"},
	}
	if !util.KubeList(nbList, &client.ListOptions{Namespace: secret.Namespace}) {
		log.Infof("Could not found NooBaa in namespace %q, while trying to find NooBaa that uses %s secret", secret.Namespace, secret.Name)
		return nil
	}

	reqs := []reconcile.Request{}

	for _, nb := range nbList.Items {
		nbSecret := util.GetNooBaaExternalPgSecret(&nb)
		if nbSecret != nil && nbSecret.Name == secret.Name {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      nb.Name,
					Namespace: nb.Namespace,
				},
			})
		}
	}

	return reqs
}

// SetAllowNoobaaDeletion sets AllowNoobaaDeletion Noobaa CR field to true so the webhook won't block the deletion
func SetAllowNoobaaDeletion(noobaa *nbv1.NooBaa) error {
	// Explicitly allow deletion of NooBaa CR
	if !noobaa.Spec.CleanupPolicy.AllowNoobaaDeletion {
		noobaa.Spec.CleanupPolicy.AllowNoobaaDeletion = true
		if !util.KubeUpdate(noobaa) {
			return fmt.Errorf("failed to update AllowNoobaaDeletion")
		}
	}
	return nil
}
