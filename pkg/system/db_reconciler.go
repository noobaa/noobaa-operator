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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	pgClusterSuffix      = "-db-pg-cluster"
	pgImageCatalogSuffix = "-db-pg-image-catalog"
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
		r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusReady

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

		if err := r.reconcileClusterImport(); err != nil {
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

	// TODO: consider specifying a separate WAL storage configuration in Spec.WalStorage
	// currently, the same storage will be used for both DB and WAL

	return nil

}

func (r *Reconciler) reconcileClusterImport() error {

	// The bootstrap configuration should only be set for the CR creation.

	// set default bootstrap configuration
	if r.CNPGCluster.Spec.Bootstrap == nil {
		r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
			InitDB: &cnpgv1.BootstrapInitDB{
				Database: "nbcore",
				Owner:    "noobaa",
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
				"nbcore",
			},
		}

		// provide the external cluster connection parameters
		r.CNPGCluster.Spec.ExternalClusters = []cnpgv1.ExternalCluster{
			{
				Name: externalClusterName,
				ConnectionParameters: map[string]string{
					"host":   r.NooBaaPostgresDB.Name + "-0." + r.NooBaaPostgresDB.Spec.ServiceName + "." + r.NooBaaPostgresDB.Namespace + ".svc",
					"user":   "noobaa",
					"dbname": "nbcore",
				},
				Password: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.SecretDB.Name,
					},
					Key: "password",
				},
			},
		}
		return nil
	} else {
		// No existing DB statefulset, set bootstrap to init with a new DB
		r.cnpgLog("no existing DB statefulset found, configuring the cluster to bootstrap a new nbcore DB")
		return nil
	}

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

	if dbSpec.DBMinVolumeSize != "" {
		storageConfiguration.Size = dbSpec.DBMinVolumeSize
	} else {
		storageConfiguration.Size = options.DefaultDBVolumeSize
	}

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

func (r *Reconciler) getEnvFromClusterSecretKey(key string) *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: r.CNPGCluster.Name + "-app",
			},
			Key: key,
		},
	}
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
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.Size, r.CNPGCluster.Spec.StorageConfiguration.Size)
}
