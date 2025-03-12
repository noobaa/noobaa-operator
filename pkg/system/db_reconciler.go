package system

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

func (r *Reconciler) ReconcileCNPGCluster() error {

	// there are several cases to handle:
	// 1. Reconciling a fresh install - No CNPG cluster and no previous DB to import from
	//    - In this case we need to create a new empty CNPG cluster
	// 2. Reconciling an upgrade from a version with a standalone DB - No CNPG cluster and DB statefulset exists
	//    - In this case we need to create a new CNPG cluster and import the DB from the previous statefulset
	//    - Import is done by providing externalCluster details in the CNPG cluster spec (https://cloudnative-pg.io/documentation/1.25/database_import/#the-microservice-type)
	//    - After Import is completed, cleanup the old DB resources.
	// 3. Reconciling an existing CNPG cluster with no standalone DB - CNPG cluster exists and DB statefulset does not exist
	//    - If major version is the same, check if the DB image is changed and update the ImageCatalog
	//    - If major version is different, handle Major version upgrade (future feature)

	// check for existing CNPG cluster
	util.Logger().Infof("DZDZ: Reconciling CNPG cluster")
	cnpgCluster := cnpg.GetCnpgClusterObj(r.NooBaa.Namespace, r.NooBaa.Name)
	util.Logger().Infof("DZDZ: CNPG cluster %v", cnpgCluster)
	gvk := cnpgCluster.GetObjectKind().GroupVersionKind()
	util.Logger().Infof("DZDZ: Looking for CNPG cluster %q with gvk %q", cnpgCluster.Name, gvk.String())
	if util.KubeCheck(cnpgCluster) {
		r.Logger.Infof("CNPG cluster %q found", cnpgCluster.Name)
	} else {
		r.Logger.Infof("CNPG cluster %q not found", cnpgCluster.Name)
	}

	return nil
}
