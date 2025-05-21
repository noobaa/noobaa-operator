package system

import (
	"fmt"
	"reflect"
	"slices"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	"github.com/go-test/deep"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	pgClusterSuffix      = "-db-pg-cluster"
	pgImageCatalogSuffix = "-db-pg-image-catalog"

	noobaaDBUser = "noobaa"
	noobaaDBName = "nbcore"
)

// ReconcileCNPGCluster reconciles the CNPG cluster
// There are several cases to handle:
// 1. Reconciling a fresh install - No CNPG cluster and no previous DB to import from
//   - In this case we need to create a new empty CNPG cluster
//   - Create a new CNPG image catalog
//   - Create a new CNPG cluster
//
// 2. Reconciling an upgrade from a version with a standalone DB - No CNPG cluster and DB statefulset exists
//   - In this case we need to create a new CNPG cluster and import the DB from the previous statefulset
//   - Import is done by providing externalCluster details in the CNPG cluster spec (https://cloudnative-pg.io/documentation/1.25/database_import/#the-microservice-type)
//   - After Import is completed, cleanup the old DB resources. For now we only scale down the standalone DB pod to 0 replicas
//   - All other pods (core, endpoints) should be stopped before starting the import
//
// 3. Reconciling an existing CNPG cluster with no standalone DB - CNPG cluster exists and DB statefulset does not exist
//   - If the major version is the same, check if the DB image is changed and update the ImageCatalog
//   - If the major version is different, handle Major version upgrade (future feature. Not implemented yet)
func (r *Reconciler) ReconcileCNPGCluster() error {

	r.cnpgLog("reconciling CNPG cluster")

	// init the DB status if not set
	if r.NooBaa.Status.DBStatus == nil {
		r.NooBaa.Status.DBStatus = &nbv1.NooBaaDBStatus{
			DBClusterStatus: nbv1.DBClusterStatusNone,
		}
	}

	// reconcile the DB image in the image catalog
	if err := r.reconcileCNPGImageCatalog(); err != nil {
		r.cnpgLogError("got error reconciling image catalog. error: %v", err)
		return err
	}

	// reconcile the DB cluster CR and apply changes
	err := r.reconcileDBCluster()
	if err != nil {
		r.cnpgLogError("got error reconciling cluster. error: %v", err)
		return err
	}

	if isClusterReady(r.CNPGCluster) {
		r.cnpgLog("cnpg cluster is ready")
		// update the DB status
		r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusReady
		r.NooBaa.Status.DBStatus.ActualVolumeSize = r.CNPGCluster.Spec.StorageConfiguration.Size

		standaloneDBPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      r.NooBaaPostgresDB.Name + "-0",
				Namespace: r.NooBaaPostgresDB.Namespace,
			},
		}

		if util.KubeCheck(standaloneDBPod) {
			// stop the standalone DB pod. For now it is only scaled down to 0 replicas, to keep it around as backup
			r.cnpgLog("Scaling down the standalone DB pod")
			if err := r.ReconcileObject(r.NooBaaPostgresDB, func() error {
				zeroReplicas := int32(0)
				r.NooBaaPostgresDB.Spec.Replicas = &zeroReplicas
				return nil
			}); err != nil {
				r.cnpgLogError("got error scaling down the standalone DB pod. error: %v", err)
				return err
			}
		}

	} else {
		return fmt.Errorf("cnpg cluster is not ready")
	}

	return nil
}

// reconcileCNPGImageCatalog creates or updates the CNPG image catalog, mapping a desired major version to a desired db image
func (r *Reconciler) reconcileCNPGImageCatalog() error {

	desiredMajorVersion := getDesiredMajorVersion(r.NooBaa.Spec.DBSpec)
	desiredDbImage := getDesiredDbImage(r.NooBaa.Spec.DBSpec)
	// reconcile CNPG image catalog
	if err := r.ReconcileObject(r.CNPGImageCatalog, func() error {

		// find the image entry for the desired major version
		imageIndex := slices.IndexFunc(r.CNPGImageCatalog.Spec.Images, func(image cnpgv1.CatalogImage) bool {
			return image.Major == desiredMajorVersion
		})

		if imageIndex == -1 {
			r.cnpgLog("no image entry found for major version %q. updating image %q", desiredMajorVersion, desiredDbImage)
			r.CNPGImageCatalog.Spec.Images = append(r.CNPGImageCatalog.Spec.Images, cnpgv1.CatalogImage{
				Major: desiredMajorVersion,
				Image: desiredDbImage,
			})
		} else if r.CNPGImageCatalog.Spec.Images[imageIndex].Image != desiredDbImage {
			// handle minor version change
			r.cnpgLog("image entry for major version %q found and is different from the desired image. updating image from %q to %q",
				desiredMajorVersion, r.CNPGImageCatalog.Spec.Images[imageIndex].Image, desiredDbImage)
			r.CNPGImageCatalog.Spec.Images[imageIndex].Image = desiredDbImage
		}

		return nil
	}); err != nil {
		r.cnpgLogError("got error reconciling image catalog. error: %v", err)
		return err
	}

	// update the DB status
	r.NooBaa.Status.DBStatus.CurrentPgMajorVersion = desiredMajorVersion
	r.NooBaa.Status.DBStatus.DBCurrentImage = desiredDbImage

	return nil
}

func (r *Reconciler) reconcileDBCluster() error {

	// get the existing cluster
	util.KubeCheck(r.CNPGCluster)

	dbSpec := r.NooBaa.Spec.DBSpec

	existingClusterSpec := r.CNPGCluster.Spec.DeepCopy()

	// set the desired cluster spec except for the bootstrap configuration which should only be set for the CR creation
	if err := r.reconcileClusterSpec(dbSpec); err != nil {
		r.cnpgLogError("got error setting desired cluster spec. error: %v", err)
		return err
	}

	// Apply changes to the cluster resources. create or update the modified cluster
	if r.CNPGCluster.UID == "" {
		// Cluster resource was not created yet. Create it and handle import if needed

		if err := r.reconcileClusterCreateOrImport(); err != nil {
			r.cnpgLogError("got error setting up cluster import. error: %v", err)
			return err
		}

		// create the cluster
		r.cnpgLog("creating new cluster")
		r.Own(r.CNPGCluster)
		if err := r.Client.Create(r.Ctx, r.CNPGCluster); err != nil {
			r.cnpgLogError("got error creating the cluster resources in kubernetes api server. error: %v", err)
			return err
		}

		// update the DB status
		if r.CNPGCluster.Spec.Bootstrap.InitDB.Import == nil {
			r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusCreating
		} else {
			r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusImporting
		}

	} else {
		// Handle Cluster CRD changes

		// check if the cluster is changed
		if r.wasClusterSpecChanged(existingClusterSpec) {
			diff := deep.Equal(*existingClusterSpec, r.CNPGCluster.Spec)
			r.cnpgLog("cluster spec is changed, updating cluster. diff: %v", diff)

			currentDBClusterStatus := r.NooBaa.Status.DBStatus.DBClusterStatus
			// avoid updating a cluster that is being created or imported.
			// We might want to consider allowing this somehow for supportability (through annotation or something)
			if currentDBClusterStatus == nbv1.DBClusterStatusCreating || currentDBClusterStatus == nbv1.DBClusterStatusImporting {
				r.cnpgLog("the cluster spec was changed but the cluster creation or import is still in progress, skipping update")
				return fmt.Errorf("cluster creation or import is still in progress, skipping update")
			}

			r.cnpgLog("cluster spec is changed, updating cluster")
			if err := r.Client.Update(r.Ctx, r.CNPGCluster); err != nil {
				r.cnpgLogError("got error updating cluster. error: %v", err)
				return err
			}
			// update the DB status
			r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusUpdating
		}

		// The cluster spec is unchanged, no need to update
		return nil
	}

	return nil
}

func (r *Reconciler) reconcileClusterSpec(dbSpec *nbv1.NooBaaDBSpec) error {

	// set app=noobaa label on the cluster to be propagated to the DB pods
	if r.CNPGCluster.Spec.InheritedMetadata == nil {
		r.CNPGCluster.Spec.InheritedMetadata = &cnpgv1.EmbeddedObjectMetadata{}
	}
	if r.CNPGCluster.Spec.InheritedMetadata.Labels == nil {
		r.CNPGCluster.Spec.InheritedMetadata.Labels = map[string]string{}
	}
	r.CNPGCluster.Spec.InheritedMetadata.Labels["app"] = "noobaa"

	// update the image catalog ref
	desiredMajorVersion := getDesiredMajorVersion(dbSpec)
	r.CNPGCluster.Spec.ImageCatalogRef = &cnpgv1.ImageCatalogRef{
		TypedLocalObjectReference: corev1.TypedLocalObjectReference{
			Kind:     "ImageCatalog",
			APIGroup: &cnpg.CnpgAPIGroup,
			Name:     r.CNPGImageCatalog.Name,
		},
		Major: desiredMajorVersion,
	}

	// update number of instances
	r.CNPGCluster.Spec.Instances = getDesiredInstances(dbSpec)

	// update db resources
	if dbSpec.DBResources != nil {
		r.CNPGCluster.Spec.Resources = *dbSpec.DBResources
	}

	// update db volume resources
	// update the storage configuration
	err := setDesiredStorageConf(&r.CNPGCluster.Spec.StorageConfiguration, dbSpec)
	if err != nil {
		r.cnpgLogError("got error getting desired storage configuration for cnpg cluster. error: %v", err)
		return err
	}

	// set tolerations and node affinity to the cluster spec
	if r.NooBaa.Spec.Affinity != nil {
		r.CNPGCluster.Spec.Affinity.NodeAffinity = r.NooBaa.Spec.Affinity.NodeAffinity
	}
	if r.NooBaa.Spec.Tolerations != nil {
		r.CNPGCluster.Spec.Affinity.Tolerations = r.NooBaa.Spec.Tolerations
	}

	// by default enable monitoring of the DB instances. if the annotation is "true", disable monitoring
	disableMonStr := r.NooBaa.ObjectMeta.Annotations[nbv1.DisableDBDefaultMonitoring]
	if r.CNPGCluster.Spec.Monitoring == nil {
		r.CNPGCluster.Spec.Monitoring = &cnpgv1.MonitoringConfiguration{}
	}
	r.CNPGCluster.Spec.Monitoring.EnablePodMonitor = disableMonStr != "true"

	r.setPostgresConfig()

	// TODO: consider specifying a separate WAL storage configuration in Spec.WalStorage
	// currently, the same storage will be used for both DB and WAL

	return nil

}

func (r *Reconciler) reconcileClusterCreateOrImport() error {

	// The bootstrap configuration should only be set for the CR creation.

	// set default bootstrap configuration
	if r.CNPGCluster.Spec.Bootstrap == nil {
		r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
			InitDB: &cnpgv1.BootstrapInitDB{
				Database:      noobaaDBName,
				Owner:         noobaaDBUser,
				LocaleCollate: "C",
			},
		}
	}

	// We first want to check if a standalone DB statefulset exists, and trigger import if so
	util.KubeCheck(r.NooBaaPostgresDB)
	if r.NooBaaPostgresDB.UID != "" {

		r.cnpgLog("standalone DB statefulset found, setting up import")

		// stop core and endpoints pods and wait for them to be terminated
		numRunningPods, err := r.stopNoobaaPodsAndGetNumRunningPods()
		if err != nil {
			r.cnpgLogError("got error stopping noobaa-core and noobaa-endpoint pods. error: %v", err)
			return err
		}
		if numRunningPods > 0 {
			r.cnpgLog("waiting for noobaa-core and noobaa-endpoint pods to be terminated")
			return fmt.Errorf("waiting for noobaa-core and noobaa-endpoint pods to be terminated")
		}

		externalClusterName := r.NooBaaPostgresDB.Name
		//setup once the pods are terminated, set import in the bootstrap configuration and continue to create the cluster
		r.CNPGCluster.Spec.Bootstrap.InitDB.Import = &cnpgv1.Import{
			Source: cnpgv1.ImportSource{
				ExternalCluster: externalClusterName,
			},
			// microservice type - import only the nbcore database
			Type: cnpgv1.MicroserviceSnapshotType,
			Databases: []string{
				noobaaDBName,
			},
		}

		// provide the external cluster connection parameters
		r.CNPGCluster.Spec.ExternalClusters = []cnpgv1.ExternalCluster{
			{
				Name: externalClusterName,
				ConnectionParameters: map[string]string{
					"host":   r.NooBaaPostgresDB.Name + "-0." + r.NooBaaPostgresDB.Spec.ServiceName + "." + r.NooBaaPostgresDB.Namespace + ".svc",
					"user":   noobaaDBUser,
					"dbname": noobaaDBName,
				},
				Password: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.SecretDB.Name,
					},
					Key: "password",
				},
			},
		}

		// cnpg will perform the import using the PV of the new DB instance. We need to make sure the PV is large enough.
		// If the used percentage is over the threshold (currently 33%), double the size of the PV.
		usedSpaceThreshold := 33
		oldDbPodName := r.NooBaaPostgresDB.Name + "-0"
		oldDbContainerName := r.NooBaaPostgresDB.Spec.Template.Spec.Containers[0].Name
		oldDbMountPath := r.NooBaaPostgresDB.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath
		percent, err := util.GetVolumeUsedPercent(options.Namespace, oldDbPodName, oldDbContainerName, oldDbMountPath)
		if err != nil {
			r.cnpgLogError("failed to get the existing DB volume used percentage. error: %v", err)
			return err
		}
		if percent > usedSpaceThreshold {
			// get the actual size of the PV from the PVC.
			// Not using the requested size from noobaa CR, since it can be different from the actual due to manual resize that could have been done.
			pvc := &corev1.PersistentVolumeClaim{}
			pvc.Name = r.NooBaaPostgresDB.Spec.VolumeClaimTemplates[0].Name + "-" + oldDbPodName
			pvc.Namespace = options.Namespace
			if !util.KubeCheck(pvc) {
				return fmt.Errorf("failed to get the existing DB PVC %s in namespace %s", pvc.Name, options.Namespace)
			}
			pvcSize := pvc.Status.Capacity[corev1.ResourceStorage]
			existingPvcSize := pvcSize.String()
			// set the cluster sorage size to double the size of the existing PVC
			pvcSize.Mul(2)
			r.cnpgLog("the existing DB PVC has %d%% used space, out of %s. setting the cluster storage size to %s to avoid import failure", percent, existingPvcSize, pvcSize.String())
			r.CNPGCluster.Spec.StorageConfiguration.Size = pvcSize.String()
		}
		return nil
	} else {
		// No existing DB statefulset, set bootstrap to init with a new DB
		r.cnpgLog("no existing DB statefulset found, configuring the cluster to bootstrap a new nbcore DB")
		return nil
	}

}

func (r *Reconciler) setPostgresConfig() {

	// set postgresql configuration
	// initially using the same as in configmap-postgres-db.yaml. we should reavaluate these values
	desiredParameters := r.CNPGCluster.Spec.PostgresConfiguration.Parameters
	if desiredParameters == nil {
		desiredParameters = map[string]string{}
	}
	overrideParameters := map[string]string{
		"huge_pages":                   "off",
		"max_connections":              "600",
		"effective_cache_size":         "3GB",
		"maintenance_work_mem":         "256MB",
		"checkpoint_completion_target": "0.9",
		"wal_buffers":                  "16MB",
		"default_statistics_target":    "100",
		"random_page_cost":             "1.1",
		"effective_io_concurrency":     "300",
		"work_mem":                     "1747kB",
		"min_wal_size":                 "2GB",
		"max_wal_size":                 "8GB",
		// setting pg_stat_statements config
		// cnpg operator will automatically add the extension to the DB (https://cloudnative-pg.io/documentation/1.25/postgresql_conf/#enabling-pg_stat_statements)
		"pg_stat_statements.track": "all",
	}
	// a reasonable value for shared_buffers when mem>1GB is 25% of the total memory (https://www.postgresql.org/docs/9.1/runtime-config-resource.html)
	// if resources are not specified, we will use the default value
	if r.NooBaa.Spec.DBResources != nil {
		requiredDBMemMB := r.NooBaa.Spec.DBSpec.DBResources.Requests.Memory().ScaledValue(resource.Mega)
		sharedBuffersMB := requiredDBMemMB / 4
		desiredParameters["shared_buffers"] = fmt.Sprintf("%dMB", sharedBuffersMB)
	}

	// set any parameters from DBSpec.DBConf in overrideParameters
	if r.NooBaa.Spec.DBSpec.DBConf != nil {
		for k, v := range r.NooBaa.Spec.DBSpec.DBConf {
			overrideParameters[k] = v
		}
	}

	// override the desired parameters with the override parameters
	for param, overrideVal := range overrideParameters {
		if desiredParameters[param] != overrideVal {
			r.cnpgLog("overriding postgres config parameter %q with value %q", param, overrideVal)
			desiredParameters[param] = overrideVal
		}
	}

	r.CNPGCluster.Spec.PostgresConfiguration.Parameters = desiredParameters
}

func (r *Reconciler) stopNoobaaPodsAndGetNumRunningPods() (int, error) {
	// stop core\endpoints pods
	zeroReplicas := int32(0)
	if err := r.ReconcileObject(r.CoreApp, func() error {
		r.CoreApp.Spec.Replicas = &zeroReplicas
		return nil
	}); err != nil {
		r.cnpgLogError("got error stopping noobaa-core pods. error: %v", err)
		return -1, err
	}
	if err := r.ReconcileObject(r.DeploymentEndpoint, func() error {
		r.DeploymentEndpoint.Spec.Replicas = &zeroReplicas
		return nil
	}); err != nil {
		r.cnpgLogError("got error stopping noobaa-endpoints pods. error: %v", err)
		return -1, err
	}
	corePodsList := &corev1.PodList{}
	if !util.KubeList(corePodsList, client.InNamespace(options.Namespace), client.MatchingLabels{"noobaa-core": "noobaa"}) {
		return -1, fmt.Errorf("got error listing noobaa-core pods")
	}
	endpointPodsList := &corev1.PodList{}
	if !util.KubeList(endpointPodsList, client.InNamespace(options.Namespace), client.MatchingLabels{"noobaa-s3": "noobaa"}) {
		return -1, fmt.Errorf("got error listing noobaa-endpoints pods")
	}
	return len(corePodsList.Items) + len(endpointPodsList.Items), nil
}

func getDesiredMajorVersion(dbSpec *nbv1.NooBaaDBSpec) int {
	desiredMajorVersion := options.PostgresMajorVersion
	if dbSpec.PostgresMajorVersion != nil {
		desiredMajorVersion = *dbSpec.PostgresMajorVersion
	}
	return desiredMajorVersion
}

func getDesiredDbImage(dbSpec *nbv1.NooBaaDBSpec) string {
	desiredDbImage := options.DBImage
	if dbSpec.DBImage != nil {
		desiredDbImage = *dbSpec.DBImage
	}
	return desiredDbImage
}

func getDesiredInstances(dbSpec *nbv1.NooBaaDBSpec) int {
	desiredInstances := options.PostgresInstances
	if dbSpec.Instances != nil {
		desiredInstances = *dbSpec.Instances
	}
	return desiredInstances
}

func setDesiredStorageConf(storageConfiguration *cnpgv1.StorageConfiguration, dbSpec *nbv1.NooBaaDBSpec) error {
	if storageConfiguration == nil {
		return fmt.Errorf("storage configuration is nil")
	}
	if dbSpec.DBStorageClass != nil {
		storageConfiguration.StorageClass = dbSpec.DBStorageClass
	} else {
		storageClassName, err := findLocalStorageClass()
		if err != nil {
			return err
		}
		storageConfiguration.StorageClass = &storageClassName
	}

	// set storageConfiguration.Size only if it is not set or if it is less than dbSpec.DBMinVolumeSize
	desiredSize := options.DefaultDBVolumeSize
	if dbSpec.DBMinVolumeSize != "" {
		desiredSize = dbSpec.DBMinVolumeSize
	}
	if storageConfiguration.Size != "" {
		currentQuantity, err := resource.ParseQuantity(storageConfiguration.Size)
		if err != nil {
			return err
		}
		desiredQuantity, err := resource.ParseQuantity(desiredSize)
		if err != nil {
			return err
		}
		if currentQuantity.Cmp(desiredQuantity) > 0 {
			desiredSize = storageConfiguration.Size
		}
	}
	storageConfiguration.Size = desiredSize

	return nil
}

func isClusterReady(cluster *cnpgv1.Cluster) bool {
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == string(cnpgv1.ConditionClusterReady) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *Reconciler) shouldReconcileCNPGCluster() bool {
	return r.NooBaa.Spec.DBSpec != nil && r.NooBaa.Spec.ExternalPgSecret == nil
}

func (r *Reconciler) shouldReconcileStandaloneDB() bool {
	return r.NooBaa.Spec.DBSpec == nil && r.NooBaa.Spec.ExternalPgSecret == nil
}

func (r *Reconciler) getClusterSecretName() string {
	return r.CNPGCluster.Name + "-app"
}

func (r *Reconciler) cnpgLog(format string, args ...interface{}) {
	r.Logger.Infof("cnpg:: "+format, args...)
}

func (r *Reconciler) cnpgLogError(format string, args ...interface{}) {
	r.Logger.Errorf("cnpg:: "+format, args...)
}

// wasClusterSpecChanged checks if any of the cluster spec fields that matter for us were changed.
// This need to be updated if we change more fields in the cluster spec.
// For some reason reflect.DeepEqual always returns false when comparing the entire spec.
func (r *Reconciler) wasClusterSpecChanged(existingClusterSpec *cnpgv1.ClusterSpec) bool {

	return !reflect.DeepEqual(existingClusterSpec.InheritedMetadata, r.CNPGCluster.Spec.InheritedMetadata) ||
		!reflect.DeepEqual(existingClusterSpec.ImageCatalogRef, r.CNPGCluster.Spec.ImageCatalogRef) ||
		existingClusterSpec.Instances != r.CNPGCluster.Spec.Instances ||
		!reflect.DeepEqual(existingClusterSpec.Resources, r.CNPGCluster.Spec.Resources) ||
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.StorageClass, r.CNPGCluster.Spec.StorageConfiguration.StorageClass) ||
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.Size, r.CNPGCluster.Spec.StorageConfiguration.Size) ||
		!reflect.DeepEqual(existingClusterSpec.Monitoring, r.CNPGCluster.Spec.Monitoring) ||
		!reflect.DeepEqual(existingClusterSpec.PostgresConfiguration.Parameters, r.CNPGCluster.Spec.PostgresConfiguration.Parameters)
}
