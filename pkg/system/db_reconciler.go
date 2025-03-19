package system

import (
	"fmt"
	"reflect"
	"slices"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pgClusterSuffix      = "-db-pg-cluster"
	pgImageCatalogSuffix = "-db-pg-image-catalog"
)

func (r *Reconciler) ReconcileCNPGCluster() error {

	// there are several cases to handle:
	// 1. Reconciling a fresh install - No CNPG cluster and no previous DB to import from
	//    - In this case we need to create a new empty CNPG cluster
	// 	    - Create a new CNPG image catalog
	// 	    - Create a new CNPG cluster
	// 2. Reconciling an upgrade from a version with a standalone DB - No CNPG cluster and DB statefulset exists
	//    - In this case we need to create a new CNPG cluster and import the DB from the previous statefulset
	//    - Import is done by providing externalCluster details in the CNPG cluster spec (https://cloudnative-pg.io/documentation/1.25/database_import/#the-microservice-type)
	//    - After Import is completed, cleanup the old DB resources.
	//    - All other pods (core, endpoints) should be stopped before starting the import
	// 3. Reconciling an existing CNPG cluster with no standalone DB - CNPG cluster exists and DB statefulset does not exist
	//    - If major version is the same, check if the DB image is changed and update the ImageCatalog
	//    - If major version is different, handle Major version upgrade (future feature)

	// init the DB status if not set
	if r.NooBaa.Status.DBStatus == nil {
		r.NooBaa.Status.DBStatus = &nbv1.NooBaaDBStatus{
			DBClusterStatus: nbv1.DBClusterStatusNone,
		}
	}

	if err := r.reconcileCNPGImageCatalog(); err != nil {
		r.Logger.Errorf("got error reconciling image catalog. error: %v", err)
		return err
	}

	if err := r.reconcileCluster(); err != nil {
		r.Logger.Errorf("got error reconciling cluster. error: %v", err)
		return err
	}

	// check the cluster status. return temporary error if the cluster is not ready
	clusterReady := false
	for _, cond := range r.CNPGCluster.Status.Conditions {
		if cond.Type == string(cnpgv1.ConditionClusterReady) && cond.Status == metav1.ConditionTrue {
			clusterReady = true
			break
		}
	}

	// TODO: check and handle different cluster failures

	if !clusterReady {
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
			r.Logger.Infof("no image entry found for major version %q. updating image %q", desiredMajorVersion, desiredDbImage)
			r.CNPGImageCatalog.Spec.Images = append(r.CNPGImageCatalog.Spec.Images, cnpgv1.CatalogImage{
				Major: desiredMajorVersion,
				Image: desiredDbImage,
			})
		} else if r.CNPGImageCatalog.Spec.Images[imageIndex].Image != desiredDbImage {
			// handle minor version change
			r.Logger.Infof("image entry for major version %q found and is different from the desired image. updating image from %q to %q",
				desiredMajorVersion, r.CNPGImageCatalog.Spec.Images[imageIndex].Image, desiredDbImage)
			r.CNPGImageCatalog.Spec.Images[imageIndex].Image = desiredDbImage
		}

		return nil
	}); err != nil {
		r.Logger.Errorf("got error reconciling image catalog. error: %v", err)
		return err
	}

	// update the DB status
	r.NooBaa.Status.DBStatus.CurrentPgMajorVersion = desiredMajorVersion
	r.NooBaa.Status.DBStatus.DBCurrentImage = desiredDbImage

	return nil
}

func (r *Reconciler) reconcileCluster() error {

	// get the existing cluster
	util.KubeCheck(r.CNPGCluster)

	existingClusterSpec := r.CNPGCluster.Spec.DeepCopy()

	dbSpec := r.NooBaa.Spec.DBSpec

	desiredMajorVersion := getDesiredMajorVersion(dbSpec)

	// set app=noobaa label on the cluster to be propagated to the DB pods
	r.CNPGCluster.Spec.InheritedMetadata = &cnpgv1.EmbeddedObjectMetadata{
		Labels: map[string]string{
			"app": "noobaa",
		},
	}

	// update the image catalog ref
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
	desiredStorageConfiguration, err := getDesiredStorageConfiguration(dbSpec)
	if err != nil {
		r.Logger.Errorf("got error getting desired storage configuration for cnpg cluster. error: %v", err)
		return err
	}
	r.CNPGCluster.Spec.StorageConfiguration = desiredStorageConfiguration

	// TODO: consider specifying a separate WAL storage configuration in Spec.WalStorage
	// currently, the same storage will be used for both DB and WAL

	// TODO: handle import of existing DB in bootstrap section
	// define bootstrap section
	r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
		InitDB: &cnpgv1.BootstrapInitDB{
			Database: "nbcore",
			Owner:    "noobaa",
		},
	}

	// check if the cluster is changed
	if reflect.DeepEqual(*existingClusterSpec, r.CNPGCluster.Spec) {
		// The cluster spec is unchanged, no need to update
		return nil
	}

	// create or update the modified cluster
	if r.CNPGCluster.UID == "" {
		r.Logger.Infof("creating new cluster")
		r.Own(r.CNPGCluster)
		if err := r.Client.Create(r.Ctx, r.CNPGCluster); err != nil {
			r.Logger.Errorf("got error creating cluster. error: %v", err)
			return err
		}
		// update the DB status
		r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusCreating
	} else {
		r.Logger.Infof("cluster spec is changed, updating cluster")
		if err := r.Client.Update(r.Ctx, r.CNPGCluster); err != nil {
			r.Logger.Errorf("got error updating cluster. error: %v", err)
			return err
		}
		// update the DB status
		r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusUpdating
	}

	return nil
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

func getDesiredStorageConfiguration(dbSpec *nbv1.NooBaaDBSpec) (cnpgv1.StorageConfiguration, error) {
	desiredStorageConfiguration := cnpgv1.StorageConfiguration{}
	if dbSpec.DBStorageClass != nil {
		desiredStorageConfiguration.StorageClass = dbSpec.DBStorageClass
	} else {
		storageClassName, err := findLocalStorageClass()
		if err != nil {
			return cnpgv1.StorageConfiguration{}, err
		}
		desiredStorageConfiguration.StorageClass = &storageClassName
	}

	if dbSpec.DBMinVolumeSize != "" {
		desiredStorageConfiguration.Size = dbSpec.DBMinVolumeSize
	} else {
		desiredStorageConfiguration.Size = options.DefaultDBVolumeSize
	}

	return desiredStorageConfiguration, nil
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
