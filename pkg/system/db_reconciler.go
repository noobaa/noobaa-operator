package system

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	"github.com/go-test/deep"
	storagesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	pgClusterSuffix      = "-db-pg-cluster"
	pgImageCatalogSuffix = "-db-pg-image-catalog"

	noobaaDBUser = "noobaa"
	noobaaDBName = "nbcore"

	defaultCalcMemoryKB     = int64(16 * 1024 * 1024) // 16GB in KB — used when DBResources not specified
	defaultCalcCPU          = int64(4)                // used when DBResources not specified
	defaultFailoverDelaySec = int32(30)
)

// ReconcileCNPGCluster reconciles the CNPG cluster
// There are several cases to handle:
// 1. Reconciling a fresh install - No CNPG cluster and no DBRecovery configuration
//   - In this case we need to create a new empty CNPG cluster
//   - Create a new CNPG image catalog
//   - Create a new CNPG cluster
//
// 2. Reconciling recovery from a snapshot - CNPG cluster was deleted by the user to initiate a recovery from a snapshot
//   - In this case we need to create a new CNPG cluster and set the bootstrap configuration to recover from the snapshot
//   - All other pods (core, endpoints) should be stopped before starting the recovery
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

		if r.NooBaa.Status.DBStatus.RecoveryStatus != nil {
			r.NooBaa.Status.DBStatus.RecoveryStatus.Status = nbv1.DBRecoveryStatusCompleted
		}

	} else {
		return fmt.Errorf("cnpg cluster is not ready")
	}

	// Reconcile backup configuration
	if err := r.reconcileDBBackup(); err != nil {
		r.cnpgLogError("got error reconciling backup. error: %v", err)
		return err
	}

	// Reconcile PDB alert silencer to disable PodDisruptionBudgetAtLimit alert
	if err := r.ReconcilePDBAlertSilencer(); err != nil {
		r.cnpgLogError("got error reconciling PDB alert silencer. error: %v", err)
		// log and ignore error on PDB alert silencer. Don't fail the reconciliation if it fails.
		return nil
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
		// If the cluster CR has not been created yet (UID = ""), it means one of two options.:
		// 1. The cluster is being created for the first time
		// 2. The cluster is being recovered from a snapshot

		// Check if noobaa CR has a recovery configuration.
		if r.NooBaa.Spec.DBSpec.DBRecovery == nil {
			// No recovery configuration found, set bootstrap configuration to init a new DB
			if r.CNPGCluster.Spec.Bootstrap == nil {
				r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
					InitDB: &cnpgv1.BootstrapInitDB{
						Database:      noobaaDBName,
						Owner:         noobaaDBUser,
						LocaleCollate: "C",
					},
				}
			}
		} else {
			// Recovery configuration found, set bootstrap configuration to recover from the snapshot
			if err := r.reconcileClusterRecovery(); err != nil {
				r.cnpgLogError("got error setting up cluster recovery. error: %v", err)
				return err
			}
		}

		// create the cluster
		r.cnpgLog("creating new cluster")
		r.Own(r.CNPGCluster)
		if err := r.Client.Create(r.Ctx, r.CNPGCluster); err != nil {
			r.cnpgLogError("got error creating the cluster resources in kubernetes api server. error: %v", err)
			return err
		}

		// update the DB status
		if r.CNPGCluster.Spec.Bootstrap.Recovery == nil {
			r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusCreating
		} else {
			r.NooBaa.Status.DBStatus.DBClusterStatus = nbv1.DBClusterStatusRecovering
			snapshotName := ""
			if r.NooBaa.Spec.DBSpec.DBRecovery != nil {
				snapshotName = r.NooBaa.Spec.DBSpec.DBRecovery.VolumeSnapshotName
			} else if r.NooBaa.Status.DBStatus.RecoveryStatus != nil {
				// fall back to previously recorded snapshot name, if any
				snapshotName = r.NooBaa.Status.DBStatus.RecoveryStatus.SnapshotName
			}
			r.NooBaa.Status.DBStatus.RecoveryStatus = &nbv1.DBRecoveryStatus{
				Status:       nbv1.DBRecoveryStatusRunning,
				SnapshotName: snapshotName,
				RecoveryTime: &metav1.Time{Time: time.Now()},
			}
		}
	} else {
		// Handle Cluster CRD changes

		// check if the cluster is changed
		if r.wasClusterSpecChanged(existingClusterSpec) {
			diff := deep.Equal(*existingClusterSpec, r.CNPGCluster.Spec)
			r.cnpgLog("cluster spec is changed, updating cluster. diff: %v", diff)

			currentDBClusterStatus := r.NooBaa.Status.DBStatus.DBClusterStatus
			// avoid updating a cluster that is being created.
			// We might want to consider allowing this somehow for supportability (through annotation or something)
			if currentDBClusterStatus == nbv1.DBClusterStatusCreating || currentDBClusterStatus == nbv1.DBClusterStatusRecovering {
				r.cnpgLog("the cluster spec was changed but the cluster creation is still in progress, skipping update")
				return fmt.Errorf("cluster creation is still in progress, skipping update")
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
	r.CNPGCluster.Spec.Instances = getDBInstances(r.NooBaa)

	// update db resources
	r.CNPGCluster.Spec.Resources = getDBResources(r.NooBaa)

	// update db volume resources
	// update the storage configuration
	err := setDesiredStorageConf(&r.CNPGCluster.Spec.StorageConfiguration, dbSpec, r.NooBaa.Annotations)
	if err != nil {
		r.cnpgLogError("got error getting desired storage configuration for cnpg cluster. error: %v", err)
		return err
	}

	// set tolerations and node affinity to the cluster spec
	if r.NooBaa.Spec.Affinity != nil {
		r.CNPGCluster.Spec.Affinity.NodeAffinity = r.NooBaa.Spec.Affinity.NodeAffinity
		if r.NooBaa.Spec.Affinity.TopologyKey != "" {
			r.CNPGCluster.Spec.Affinity.TopologyKey = r.NooBaa.Spec.Affinity.TopologyKey
		}
	} else {
		r.CNPGCluster.Spec.Affinity.NodeAffinity = nil
		r.CNPGCluster.Spec.Affinity.TopologyKey = ""
	}
	if r.NooBaa.Spec.Tolerations != nil {
		r.CNPGCluster.Spec.Affinity.Tolerations = r.NooBaa.Spec.Tolerations
	}
	r.CNPGCluster.Spec.PriorityClassName = r.NooBaa.Spec.DBPriorityClassName

	// by default enable monitoring of the DB instances. if the annotation is "true", disable monitoring
	disableMonStr := r.NooBaa.Annotations[nbv1.DisableDBDefaultMonitoring]
	if r.CNPGCluster.Spec.Monitoring == nil {
		r.CNPGCluster.Spec.Monitoring = &cnpgv1.MonitoringConfiguration{}
	}

	// TODO: This is deprecated and will be ignored in future versions. migrate to manually managed PodMonitor
	r.CNPGCluster.Spec.Monitoring.EnablePodMonitor = disableMonStr != "true"             //nolint:staticcheck // SA1019 ignore deprecation
	r.CNPGCluster.Spec.Monitoring.TLSConfig = &cnpgv1.ClusterMonitoringTLSConfiguration{ //nolint:staticcheck // SA1019 ignore deprecation
		Enabled: true,
	}

	r.CNPGCluster.Spec.FailoverDelay = defaultFailoverDelaySec

	r.setPostgresConfig()

	// Configure backup settings if specified
	if backupSpec := r.NooBaa.Spec.DBSpec.DBBackup; backupSpec != nil {
		if backupSpec.VolumeSnapshot == nil {
			r.cnpgLogError("volume snapshot backup configuration is not specified")
			return fmt.Errorf("volume snapshot backup configuration is not specified")
		}
		offlineBackup := false
		if r.CNPGCluster.Spec.Backup == nil {
			r.CNPGCluster.Spec.Backup = &cnpgv1.BackupConfiguration{
				VolumeSnapshot: &cnpgv1.VolumeSnapshotConfiguration{
					ClassName: backupSpec.VolumeSnapshot.VolumeSnapshotClass,
					Online:    &offlineBackup,
				},
			}
		} else {
			if r.CNPGCluster.Spec.Backup.VolumeSnapshot == nil {
				r.CNPGCluster.Spec.Backup.VolumeSnapshot = &cnpgv1.VolumeSnapshotConfiguration{}
			}
			r.CNPGCluster.Spec.Backup.VolumeSnapshot.ClassName = backupSpec.VolumeSnapshot.VolumeSnapshotClass
			r.CNPGCluster.Spec.Backup.VolumeSnapshot.Online = &offlineBackup
		}
	} else {
		// Remove backup configuration if not specified
		r.CNPGCluster.Spec.Backup = nil
	}

	// TODO: consider specifying a separate WAL storage configuration in Spec.WalStorage
	// currently, the same storage will be used for both DB and WAL

	return nil

}

// cleanupDBBackup removes the scheduled backup configuration if it exists
func (r *Reconciler) cleanupDBBackup() error {
	scheduledBackup := cnpg.GetCnpgScheduledBackupObj(r.CNPGCluster.Namespace, r.getBackupResourceName())

	if util.KubeCheckQuiet(scheduledBackup) {
		r.cnpgLog("removing ScheduledBackup %s", scheduledBackup.Name)
		if err := r.Client.Delete(r.Ctx, scheduledBackup); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (r *Reconciler) reconcileClusterRecovery() error {

	r.cnpgLog("recovery configuration found, setting up cluster to recover from snapshot")

	// set recovery status to pending
	r.NooBaa.Status.DBStatus.RecoveryStatus = &nbv1.DBRecoveryStatus{
		Status:       nbv1.DBRecoveryStatusPending,
		SnapshotName: r.NooBaa.Spec.DBSpec.DBRecovery.VolumeSnapshotName,
	}

	// stop core and endpoints pods and wait for them to be terminated
	numRunningPods, err := r.stopNoobaaPodsAndGetNumRunningPods()
	if err != nil {
		r.cnpgLogError("got error stopping noobaa-core and noobaa-endpoint pods. error: %v", err)
		return err
	}
	if numRunningPods != 0 {
		r.cnpgLog("waiting for noobaa-core and noobaa-endpoint pods to be terminated")
		return fmt.Errorf("waiting for noobaa-core and noobaa-endpoint pods to be terminated")
	}

	r.cnpgLog("setting up cluster to recover from snapshot %q", r.NooBaa.Spec.DBSpec.DBRecovery.VolumeSnapshotName)

	VolSnapshotAPIGroup := storagesnapshotv1.GroupName
	//setup once the pods are terminated, set recovery in the bootstrap configuration and continue to create the cluster
	r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
		Recovery: &cnpgv1.BootstrapRecovery{
			Database: noobaaDBName,
			Owner:    noobaaDBUser,
			VolumeSnapshots: &cnpgv1.DataSource{
				Storage: corev1.TypedLocalObjectReference{
					Kind:     "VolumeSnapshot",
					Name:     r.NooBaa.Spec.DBSpec.DBRecovery.VolumeSnapshotName,
					APIGroup: &VolSnapshotAPIGroup,
				},
			},
		},
	}

	// delete the existing scheduled backup configuration to avoid a creation of a new snapshot while recovering
	if err := r.cleanupDBBackup(); err != nil {
		r.cnpgLogError("got error cleaning up existing backup resources. error: %v", err)
		return err
	}

	r.cnpgLog("recovery configuration set, continuing to create the cluster")

	return nil

}

// reconcileDBBackup reconciles the backup configuration for the CNPG cluster
func (r *Reconciler) reconcileDBBackup() error {
	if r.NooBaa.Spec.DBSpec.DBBackup == nil {
		// Clean up existing backup resources if any
		if err := r.cleanupDBBackup(); err != nil {
			r.cnpgLogError("got error cleaning up existing backup resources. error: %v", err)
			return err
		}
		r.NooBaa.Status.DBStatus.BackupStatus = nil
		return nil
	}

	if r.NooBaa.Status.DBStatus.BackupStatus == nil {
		r.NooBaa.Status.DBStatus.BackupStatus = &nbv1.DBBackupStatus{}
	}

	// currently only volume snapshot backup is supported. VolumeSnapshot is required.
	if r.NooBaa.Spec.DBSpec.DBBackup.VolumeSnapshot == nil {
		r.cnpgLogError("volume snapshot backup configuration is not specified")
		return fmt.Errorf("volume snapshot backup configuration is not specified")
	}

	// Create or update ScheduledBackup
	if err := r.reconcileScheduledBackup(); err != nil {
		r.cnpgLogError("got error reconciling scheduled backup. error: %v", err)
		return err
	}

	// reconcile the backup retention
	if err := r.reconcileBackupRetention(); err != nil {
		r.cnpgLogError("got error reconciling backup retention. error: %v", err)
		return err
	}

	return nil
}

func (r *Reconciler) getBackupResourceName() string {
	return r.CNPGCluster.Name + "-scheduled-backup"
}

func (r *Reconciler) reconcileScheduledBackup() error {
	backupSpec := r.NooBaa.Spec.DBSpec.DBBackup
	offlineBackup := false
	// convert the standard cron schedule to the cnpg cron schedule
	cnpgSchedule, err := convertToSixFieldCron(backupSpec.Schedule)
	if err != nil {
		r.cnpgLogError("failed to convert cron schedule to six-field format. error: %v", err)
		return err
	}
	scheduledBackup := cnpg.GetCnpgScheduledBackupObj(r.CNPGCluster.Namespace, r.getBackupResourceName())

	return r.ReconcileObject(scheduledBackup, func() error {
		// Update spec if needed
		scheduledBackup.Spec.Schedule = cnpgSchedule
		scheduledBackup.Spec.Cluster.Name = r.CNPGCluster.Name
		scheduledBackup.Spec.Method = cnpgv1.BackupMethodVolumeSnapshot
		scheduledBackup.Spec.Online = &offlineBackup
		scheduledBackup.Spec.Target = cnpgv1.BackupTargetStandby
		if scheduledBackup.Status.LastScheduleTime != nil {
			r.NooBaa.Status.DBStatus.BackupStatus.LastBackupTime = scheduledBackup.Status.LastScheduleTime
		}
		if scheduledBackup.Status.NextScheduleTime != nil {
			r.NooBaa.Status.DBStatus.BackupStatus.NextBackupTime = scheduledBackup.Status.NextScheduleTime
		}
		return nil
	})
}

func (r *Reconciler) reconcileBackupRetention() error {
	maxSnapshots := r.NooBaa.Spec.DBSpec.DBBackup.VolumeSnapshot.MaxSnapshots
	if maxSnapshots == 0 {
		r.cnpgLog("backup retention is not specified, skipping")
		return nil
	}

	// list all snapshots created as part of the scheduled backup
	snapshots, err := r.listVolumeSnapshotsOrderByCreate()
	if err != nil {
		r.cnpgLogError("got error listing cluster volume snapshots. error: %v", err)
		return err
	}
	totalSnapshots := len(snapshots)
	availableItems := snapshots
	if totalSnapshots > maxSnapshots {
		numToDelete := totalSnapshots - maxSnapshots
		availableItems = availableItems[numToDelete:]
		r.cnpgLog("found %d snapshots, maxSnapshots is %d, deleting the oldest %d snapshots", totalSnapshots, maxSnapshots, numToDelete)
		// delete the oldest backups
		for _, snapshot := range snapshots[:numToDelete] {
			if r.NooBaa.Spec.DBSpec.DBRecovery != nil && r.NooBaa.Spec.DBSpec.DBRecovery.VolumeSnapshotName == snapshot.Name {
				r.cnpgLog("skipping deletion of recovery snapshot %s", snapshot.Name)
				availableItems = append(availableItems, snapshot)
				continue
			}
			// delete the snapshot
			r.cnpgLog("deleting snapshot %s", snapshot.Name)
			// if encountered an error we only report it and continue with the reconciliation
			if err := r.Client.Delete(r.Ctx, &snapshot); err != nil && !errors.IsNotFound(err) {
				r.cnpgLogError("got error deleting snapshot %s. error: %v", snapshot.Name, err)
				// skipping the deletion of the backup so we can try to delete again next time
				continue
			}
			totalSnapshots--
			// delete the coresponding backup resource
			backup := cnpg.GetCnpgBackupObj(r.CNPGCluster.Namespace, snapshot.Name)
			if err := r.Client.Delete(r.Ctx, backup); err != nil && !errors.IsNotFound(err) {
				r.cnpgLogError("got error deleting backup %s. error: %v", backup.Name, err)
			}
		}
	}

	// update the backup status
	r.NooBaa.Status.DBStatus.BackupStatus.TotalSnapshots = totalSnapshots
	avaiallableSnapshots := []string{}
	for _, item := range availableItems {
		avaiallableSnapshots = append(avaiallableSnapshots, item.Name)
	}
	r.NooBaa.Status.DBStatus.BackupStatus.AvailableSnapshots = avaiallableSnapshots

	return nil
}

// listVolumeSnapshotsOrderByCreate lists all volume snapshots of the scheduled backup, ordered by creation timestamp
func (r *Reconciler) listVolumeSnapshotsOrderByCreate() ([]storagesnapshotv1.VolumeSnapshot, error) {
	volumeSnapshots := storagesnapshotv1.VolumeSnapshotList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: storagesnapshotv1.GroupName,
			Kind:       "VolumeSnapshotList",
		},
	}

	// list all volume snapshots having the label cnpg.io/cluster: <cluster-name>
	if err := r.Client.List(r.Ctx, &volumeSnapshots, &client.ListOptions{
		Namespace: r.CNPGCluster.Namespace,
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
			"cnpg.io/cluster": r.CNPGCluster.Name,
		}),
	}); err != nil {
		r.cnpgLogError("got error listing volume snapshots. error: %v", err)
		return nil, err
	}

	// filter the list by name. include only snapshots starting with the scheduled backup name
	filteredItems := []storagesnapshotv1.VolumeSnapshot{}
	for _, snapshot := range volumeSnapshots.Items {
		if strings.HasPrefix(snapshot.Name, r.getBackupResourceName()) {
			filteredItems = append(filteredItems, snapshot)
		}
	}
	// sort the list by creation timestamp
	slices.SortFunc(filteredItems, func(a, b storagesnapshotv1.VolumeSnapshot) int {
		return a.CreationTimestamp.Compare(b.CreationTimestamp.Time)
	})
	return filteredItems, nil
}

func (r *Reconciler) setPostgresConfig() {
	desiredParameters := r.CNPGCluster.Spec.PostgresConfiguration.Parameters
	if desiredParameters == nil {
		desiredParameters = map[string]string{}
	}

	// default parameters
	overrideParameters := map[string]string{
		"huge_pages":                   "off",
		"checkpoint_completion_target": "0.9",
		"default_statistics_target":    "100",
		"random_page_cost":             "1.1",
		"effective_io_concurrency":     "300",
		"min_wal_size":                 "2GB",
		"max_wal_size":                 "8GB",
		"jit":                          "off",
		"wal_compression":              "lz4",
		"wal_level":                    "replica",
		"wal_sender_timeout":           "60s",
		"wal_receiver_timeout":         "60s",
		"pg_stat_statements.track":     "all",
	}

	_, endpointMaxCount := r.getEndpointMinMaxCount()

	totalMemoryKB := defaultCalcMemoryKB
	cpuNum := defaultCalcCPU
	dbResources := getDBResources(r.NooBaa)
	if !dbResources.Limits.Memory().IsZero() {
		totalMemoryKB = dbResources.Limits.Memory().Value() / 1024
	} else if !dbResources.Requests.Memory().IsZero() {
		totalMemoryKB = dbResources.Requests.Memory().Value() / 1024
	}
	if !dbResources.Limits.Cpu().IsZero() {
		cpuNum = dbResources.Limits.Cpu().MilliValue() / 1000
	} else if !dbResources.Requests.Cpu().IsZero() {
		cpuNum = dbResources.Requests.Cpu().MilliValue() / 1000
	}
	if cpuNum < 1 {
		cpuNum = 1
	}

	for k, v := range calculatePGConfig(totalMemoryKB, cpuNum, int(endpointMaxCount), r.CNPGCluster.Spec.StorageConfiguration.Size) {
		overrideParameters[k] = v
	}
	r.cnpgLog("PGTune config: memory=%dKB, cpu=%d, endpoints=%d", totalMemoryKB, cpuNum, endpointMaxCount)

	// apply any user-specified DBConf overrides on top of the calculated values
	if r.NooBaa.Spec.DBSpec.DBConf != nil {
		for k, v := range r.NooBaa.Spec.DBSpec.DBConf {
			overrideParameters[k] = v
		}
	}

	for param, overrideVal := range overrideParameters {
		if desiredParameters[param] != overrideVal {
			r.cnpgLog("overriding postgres config parameter %q with value %q", param, overrideVal)
			desiredParameters[param] = overrideVal
		}
	}

	r.CNPGCluster.Spec.PostgresConfiguration.Parameters = desiredParameters
}

// calculatePGConfig computes dynamically-calculated PostgreSQL configuration parameters.
// Calculations are based on PGTune logic: https://github.com/le0pard/pgtune/blob/master/src/features/configuration/configurationSlice.js
// Fixed assumptions: DB Version: 16, OS: Linux, DB Type: OLTP, Storage: SAN.
// totalMemoryKB is the total DB memory in kilobytes, cpuNum is the number of CPU
// cores (0 means unknown), numEndpoints is the number of endpoint pods used to
// derive max_connections, and dataVolumeSize is the PVC size string used to derive wal_keep_size.
func calculatePGConfig(totalMemoryKB int64, cpuNum int64, numEndpoints int, dataVolumeSize string) map[string]string {
	params := map[string]string{}

	maxConnections := calculateMaxConnections(numEndpoints)

	// shared_buffers = 25% of total memory
	sharedBuffersKB := totalMemoryKB / 4
	params["shared_buffers"] = formatBytesKB(sharedBuffersKB)

	// effective_cache_size = 75% of total memory
	params["effective_cache_size"] = formatBytesKB((totalMemoryKB * 3) / 4)

	// maintenance_work_mem = totalMem / 16, capped at 8GB
	maintenanceWorkMemKB := totalMemoryKB / 16
	if maintenanceWorkMemKB > 8*1024*1024 {
		maintenanceWorkMemKB = 8 * 1024 * 1024
	}
	params["maintenance_work_mem"] = formatBytesKB(maintenanceWorkMemKB)

	// wal_buffers = 3% of shared_buffers, clamped to [32kB, 16MB], rounded up near 14MB
	walBuffersKB := (3 * sharedBuffersKB) / 100
	const maxWalBuf = 16 * 1024
	if walBuffersKB > maxWalBuf {
		walBuffersKB = maxWalBuf
	} else if walBuffersKB > 14*1024 {
		walBuffersKB = maxWalBuf
	}
	if walBuffersKB < 32 {
		walBuffersKB = 32
	}
	params["wal_buffers"] = formatBytesKB(walBuffersKB)

	params["max_connections"] = fmt.Sprintf("%d", maxConnections)

	// parallel query settings (only beneficial with >= 4 cores)
	parallelForWorkMem := int64(8) // PG default max_worker_processes
	if cpuNum >= 4 {
		params["max_worker_processes"] = fmt.Sprintf("%d", cpuNum)
		params["max_parallel_workers"] = fmt.Sprintf("%d", cpuNum)

		workersPerGather := (cpuNum + 1) / 2
		if workersPerGather > 4 {
			workersPerGather = 4
		}
		params["max_parallel_workers_per_gather"] = fmt.Sprintf("%d", workersPerGather)

		parallelMaintenanceWorkers := (cpuNum + 1) / 2
		if parallelMaintenanceWorkers > 4 {
			parallelMaintenanceWorkers = 4
		}
		params["max_parallel_maintenance_workers"] = fmt.Sprintf("%d", parallelMaintenanceWorkers)

		parallelForWorkMem = cpuNum
	}

	// work_mem = (totalMem - shared_buffers) / ((max_connections + max_worker_processes) * 3)
	// minimum 4MB to prevent disk-spill for simple queries
	workMemKB := (totalMemoryKB - sharedBuffersKB) / ((int64(maxConnections) + parallelForWorkMem) * 3)
	if workMemKB < 4*1024 {
		workMemKB = 4 * 1024
	}
	params["work_mem"] = formatBytesKB(workMemKB)

	// wal_keep_size = 12% of data volume, capped at 16GB
	params["wal_keep_size"] = calculateWalKeepSize(dataVolumeSize)

	return params
}

// formatBytesKB formats a value in kilobytes to a PostgreSQL memory parameter string,
// choosing the largest whole unit (GB > MB > kB).
func formatBytesKB(kb int64) string {
	const (
		gbInKB = 1024 * 1024
		mbInKB = 1024
	)
	if kb >= gbInKB && kb%gbInKB == 0 {
		return fmt.Sprintf("%dGB", kb/gbInKB)
	}
	if kb >= mbInKB && kb%mbInKB == 0 {
		return fmt.Sprintf("%dMB", kb/mbInKB)
	}
	return fmt.Sprintf("%dkB", kb)
}

// calculateWalKeepSize computes wal_keep_size as 12% of the data PVC size,
// capped at 16GB. Falls back to "6GB" if the volume size cannot be parsed.
func calculateWalKeepSize(dataVolumeSize string) string {
	const (
		walKeepFractionPercent = 12
		maxWalKeepSizeMB       = int64(16 * 1024) // 16GB in MB
		minWalKeepSizeMB       = int64(1024)      // 1GB minimum
	)
	qty, err := resource.ParseQuantity(dataVolumeSize)
	if err != nil {
		return "6GB"
	}
	volMB := qty.Value() / (1024 * 1024)
	walKeepMB := volMB * int64(walKeepFractionPercent) / 100
	if walKeepMB > maxWalKeepSizeMB {
		walKeepMB = maxWalKeepSizeMB
	}
	if walKeepMB < minWalKeepSizeMB {
		walKeepMB = minWalKeepSizeMB
	}
	return formatBytesKB(walKeepMB * 1024)
}

// calculateMaxConnections computes the PostgreSQL max_connections based on the
// number of endpoint pods. Each noobaa-core process (web_server, bg_worker,
// hosted_agents) uses up to 20 DB connections (default + md pools). Each
// endpoint process uses up to 80 connections (default + md pools, excluding
// the read_only pool which targets the standby replica).
// 3 additional connections are reserved for superuser access.
func calculateMaxConnections(numEndpoints int) int {
	const (
		coreProcesses          = 3
		connectionsPerCore     = 20
		connectionsPerEndpoint = 80
		superuserReserved      = 3
	)
	return coreProcesses*connectionsPerCore + numEndpoints*connectionsPerEndpoint + superuserReserved
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

func setDesiredStorageConf(storageConfiguration *cnpgv1.StorageConfiguration, dbSpec *nbv1.NooBaaDBSpec, annotations map[string]string) error {
	if storageConfiguration == nil {
		return fmt.Errorf("storage configuration is nil")
	}
	if storageConfiguration.PersistentVolumeClaimTemplate == nil {
		storageConfiguration.PersistentVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{}
	}
	accessMode := getDesiredDBPVCAccessMode(annotations)
	storageConfiguration.PersistentVolumeClaimTemplate.AccessModes = []corev1.PersistentVolumeAccessMode{accessMode}
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

// getDesiredDBPVCAccessMode returns the DB PVC access mode from the NooBaa CR annotation.
// Default is ReadWriteOncePod. Set noobaa.io/pvc_access_mode_rwo=true to use ReadWriteOnce
// when the CSI driver lacks RWOP support.
func getDesiredDBPVCAccessMode(annotations map[string]string) corev1.PersistentVolumeAccessMode {
	if annotations != nil && annotations[nbv1.PVCAccessModeRWO] == "true" {
		return corev1.ReadWriteOnce
	}
	return corev1.ReadWriteOncePod
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

// convertToSixFieldCron converts a cron schedule to six-field format.
// cnpg requires the schedule to be in six-field format, so we need to convert the standard cron schedule to six-field format.
func convertToSixFieldCron(schedule string) (string, error) {
	// If it starts with @, it's a descriptor or interval - return as-is
	if len(schedule) > 0 && schedule[0] == '@' {
		return schedule, nil
	}

	// Count the number of space-separated fields
	fields := strings.Fields(schedule)

	// If already 6 fields, return as-is
	if len(fields) == 6 {
		return schedule, nil
	}

	// If 5 fields (standard cron), prepend "0" for seconds
	if len(fields) == 5 {
		return "0 " + schedule, nil
	}

	// If not 5 or 6 fields, return an error
	return "", fmt.Errorf("invalid cron schedule %q", schedule)
}

// wasClusterSpecChanged checks if any of the cluster spec fields that matter for us were changed.
// This need to be updated if we change more fields in the cluster spec.
// For some reason reflect.DeepEqual always returns false when comparing the entire spec.
func (r *Reconciler) wasClusterSpecChanged(existingClusterSpec *cnpgv1.ClusterSpec) bool {
	return !reflect.DeepEqual(existingClusterSpec.InheritedMetadata, r.CNPGCluster.Spec.InheritedMetadata) ||
		!reflect.DeepEqual(existingClusterSpec.ImageCatalogRef, r.CNPGCluster.Spec.ImageCatalogRef) ||
		existingClusterSpec.Instances != r.CNPGCluster.Spec.Instances ||
		!reflect.DeepEqual(existingClusterSpec.Affinity, r.CNPGCluster.Spec.Affinity) ||
		!reflect.DeepEqual(existingClusterSpec.Resources, r.CNPGCluster.Spec.Resources) ||
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.StorageClass, r.CNPGCluster.Spec.StorageConfiguration.StorageClass) ||
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.Size, r.CNPGCluster.Spec.StorageConfiguration.Size) ||
		!reflect.DeepEqual(existingClusterSpec.StorageConfiguration.PersistentVolumeClaimTemplate, r.CNPGCluster.Spec.StorageConfiguration.PersistentVolumeClaimTemplate) ||
		!reflect.DeepEqual(existingClusterSpec.Monitoring, r.CNPGCluster.Spec.Monitoring) ||
		!reflect.DeepEqual(existingClusterSpec.PostgresConfiguration.Parameters, r.CNPGCluster.Spec.PostgresConfiguration.Parameters) ||
		!reflect.DeepEqual(existingClusterSpec.Backup, r.CNPGCluster.Spec.Backup) ||
		existingClusterSpec.PriorityClassName != r.CNPGCluster.Spec.PriorityClassName ||
		existingClusterSpec.FailoverDelay != r.CNPGCluster.Spec.FailoverDelay
}
