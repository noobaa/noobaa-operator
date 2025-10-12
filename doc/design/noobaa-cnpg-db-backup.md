# NooBaa CloudNative-PG Database Backup Design

[NooBaa Operator](../README.md) / [Design](../design/) /

## Overview

This document outlines the design for implementing automated database backup and restore functionality for NooBaa using CloudNative-PG's volume snapshot capabilities. The solution will provide customers with an easy-to-use backup and restore mechanism to minimize the risk of data loss.

## Goals

- Provide an automated backup solution for NooBaa database using CloudNative-PG volume snapshots
- Enable scheduled backups with configurable retention policies
- Support on-demand backup operations via NooBaa CLI
- Implement recovery functionality from volume snapshots
- Minimize service interruption during backup operations

## Current State

Currently, NooBaa does not provide an official backup and restore solution for the database. Customers either:
- Perform no backups (high risk of data loss)
- Use manual `pg_dump` operations (inconvenient and error-prone)

With the integration of CloudNative-PG in NooBaa 4.19+, we can leverage CNPG's built-in backup capabilities to provide a robust, automated solution.

## CloudNative-PG Backup Capabilities

CloudNative-PG provides several backup methods:

### Volume Snapshots
- **Physical backups** using Kubernetes volume snapshots
- **Offline backups** for consistency (PostgreSQL server stopped during snapshot)
- **Online backups** for minimal downtime (PostgreSQL server running)
- **Scheduled backups** using cron expressions
- **Retention policies** for automatic cleanup

### Backup Resources
- **ScheduledBackup**: For recurring backup schedules
- **Backup**: For one-time backup operations
- **VolumeSnapshot**: Kubernetes native volume snapshots

## Design Specification

### NooBaa CR Schema Changes

#### Backup Configuration
Add a new `dbBackup` property under `dbSpec`:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
spec:
  dbSpec:
    dbBackup:
      schedule: "0 0 1 * *"                   # Required: Cron format
      volumeSnapshot:
        volumeSnapshotClass: "csi-snapshotter"  # Required
        maxSnapshots: 12                        # Required: Retention count
```

#### Recovery Configuration
Add a new `dbRecovery` property under `dbSpec`:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
spec:
  dbSpec:
    dbRecovery:
      volumeSnapshotName: "snapshot-jan-2025"  # Required: Snapshot name
```

### API Type Definitions

#### NooBaaDBSpec Extensions
```go
type NooBaaDBSpec struct {
    // ... existing fields ...
    
    // DBBackup (optional) configure automatic scheduled backups
    // +optional
    DBBackup *DBBackupSpec `json:"dbBackup,omitempty"`
    
    // DBRecovery (optional) configure database recovery from snapshot
    // +optional
    DBRecovery *DBRecoverySpec `json:"dbRecovery,omitempty"`
}

type DBBackupSpec struct {
    // VolumeSnapshotClass the volume snapshot class for the database volume
    VolumeSnapshotClass string `json:"volumeSnapshotClass"`
    
    // Schedule the schedule for the database backup in cron format
    Schedule string `json:"schedule"`
    
    // MaxSnapshots the maximum number of snapshots to keep
    MaxSnapshots int `json:"maxSnapshots"`
}

type DBRecoverySpec struct {
    // VolumeSnapshotName specifies the name of the volume snapshot to recover from
    VolumeSnapshotName string `json:"volumeSnapshotName"`
}
```

#### NooBaaDBStatus Extensions
```go
type NooBaaDBStatus struct {
    // ... existing fields ...
    
    // BackupStatus reports the status of database backups
    // +optional
    BackupStatus *DBBackupStatus `json:"backupStatus,omitempty"`
    
    // RecoveryStatus reports the status of database recovery
    // +optional
    RecoveryStatus *DBRecoveryStatus `json:"recoveryStatus,omitempty"`
}

type DBBackupStatus struct {
    // LastBackupTime timestamp of the last successful backup
    LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
    
    // NextBackupTime timestamp of the next scheduled backup
    NextBackupTime *metav1.Time `json:"nextBackupTime,omitempty"`
    
    // TotalSnapshots current number of snapshots
    TotalSnapshots int `json:"totalSnapshots,omitempty"`
    
    // AvailableSnapshots list of available snapshot names
    AvailableSnapshots []string `json:"availableSnapshots,omitempty"`
}

type DBRecoveryStatus struct {
    // Status current recovery status
    Status DBRecoveryStatusType `json:"status,omitempty"`
    
    // SnapshotName name of the snapshot being recovered from
    SnapshotName string `json:"snapshotName,omitempty"`
    
    // RecoveryTime timestamp when recovery was initiated
    RecoveryTime *metav1.Time `json:"recoveryTime,omitempty"`
}

type DBRecoveryStatusType string

const (
    DBRecoveryStatusNone     DBRecoveryStatusType = "None"
    DBRecoveryStatusPending  DBRecoveryStatusType = "Pending"
    DBRecoveryStatusRunning  DBRecoveryStatusType = "Running"
    DBRecoveryStatusCompleted DBRecoveryStatusType = "Completed"
    DBRecoveryStatusFailed   DBRecoveryStatusType = "Failed"
)
```

### Backup Reconciliation Logic

#### Phase 2: Creating - Database Reconciliation
The backup configuration is handled during the database cluster reconciliation:

```go
func (r *Reconciler) ReconcileCNPGCluster() error {
    // ... existing cluster reconciliation ...
    
    // Reconcile backup configuration
    if err := r.reconcileDBBackup(); err != nil {
        return err
    }
    
    // Recovery is handled within reconcileClusterCreateOrImport
    // when the cluster is being created or recreated
    
    return nil
}
```

#### Backup Reconciliation
```go
func (r *Reconciler) reconcileDBBackup() error {
    if r.NooBaa.Spec.DBSpec.DBBackup == nil {
        // Clean up existing backup resources if any
        return r.cleanupDBBackup()
    }
    
    backupSpec := r.NooBaa.Spec.DBSpec.DBBackup
    
    // Configure cluster backup settings
    r.CNPGCluster.Spec.Backup = &cnpgv1.BackupConfiguration{
        VolumeSnapshot: &cnpgv1.VolumeSnapshotConfiguration{
            ClassName: backupSpec.VolumeSnapshotClass,
        },
    }
    
    // Create or update ScheduledBackup
    return r.reconcileScheduledBackup(backupSpec)
}

func (r *Reconciler) reconcileScheduledBackup(backupSpec *nbv1.DBBackupSpec) error {
    scheduledBackup := &cnpgv1.ScheduledBackup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      r.CNPGCluster.Name + "-backup",
            Namespace: r.CNPGCluster.Namespace,
        },
        Spec: cnpgv1.ScheduledBackupSpec{
            Schedule: backupSpec.Schedule,
            Cluster: cnpgv1.LocalObjectReference{
                Name: r.CNPGCluster.Name,
            },
            Method: cnpgv1.BackupMethodVolumeSnapshot,
            Online: false, // Offline backup for consistency
        },
    }
    
    r.Own(scheduledBackup)
    return r.ReconcileObject(scheduledBackup, func() error {
        // Update spec if needed
        return nil
    })
}
```

#### Snapshot Retention Management
```go
func (r *Reconciler) reconcileSnapshotRetention() error {
    if r.NooBaa.Spec.DBSpec.DBBackup == nil {
        return nil
    }
    
    maxSnapshots := r.NooBaa.Spec.DBSpec.DBBackup.MaxSnapshots
    
    // List all snapshots for this cluster
    snapshots, err := r.listClusterSnapshots()
    if err != nil {
        return err
    }
    
    if len(snapshots) > maxSnapshots {
        // Sort by creation time (oldest first)
        sort.Slice(snapshots, func(i, j int) bool {
            return snapshots[i].CreationTimestamp.Before(&snapshots[j].CreationTimestamp)
        })
        
        // Delete oldest snapshots
        toDelete := len(snapshots) - maxSnapshots
        for i := 0; i < toDelete; i++ {
            if err := r.Client.Delete(r.Ctx, &snapshots[i]); err != nil {
                r.Logger.Errorf("Failed to delete snapshot %s: %v", snapshots[i].Name, err)
            }
        }
    }
    
    return nil
}
```

### Recovery Implementation

#### Recovery Process
Recovery is integrated into the existing `reconcileDBCluster` flow within `reconcileClusterCreateOrImport`. The recovery process is initiated when:
1. `dbSpec.dbRecovery` is specified in NooBaa CR
2. Existing Cluster CR is manually deleted by user
3. Operator detects missing cluster and recovery configuration

```go
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

    // Check for recovery configuration first
    if r.NooBaa.Spec.DBSpec.DBRecovery != nil {
        return r.setupRecoveryBootstrap()
    }

    // We first want to check if a standalone DB statefulset exists, and trigger import if so
    util.KubeCheck(r.NooBaaPostgresDB)
    if r.NooBaaPostgresDB.UID != "" {
        return r.setupImportBootstrap()
    } else {
        // No existing DB statefulset, set bootstrap to init with a new DB
        r.cnpgLog("no existing DB statefulset found, configuring the cluster to bootstrap a new nbcore DB")
        return nil
    }
}

func (r *Reconciler) setupRecoveryBootstrap() error {
    recoverySpec := r.NooBaa.Spec.DBSpec.DBRecovery
    snapshotName := recoverySpec.VolumeSnapshotName
    
    r.cnpgLog("recovery configuration found, setting up cluster to recover from snapshot %q", snapshotName)
    
    // Configure cluster bootstrap for recovery
    r.CNPGCluster.Spec.Bootstrap = &cnpgv1.BootstrapConfiguration{
        Recovery: &cnpgv1.BootstrapRecovery{
            VolumeSnapshot: &cnpgv1.VolumeSnapshotSource{
                Name: snapshotName,
            },
        },
    }
    
    // Update status
    r.NooBaa.Status.DBStatus.RecoveryStatus = &nbv1.DBRecoveryStatus{
        Status:       nbv1.DBRecoveryStatusRunning,
        SnapshotName: snapshotName,
        RecoveryTime: &metav1.Time{Time: time.Now()},
    }
    
    return nil
}
```

### CLI Integration

#### On-Demand Backup Command
```go
// CmdDBBackup returns a CLI command for database backup
func CmdDBBackup() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "db-backup",
        Short: "Create an on-demand database backup",
        Run:   RunDBBackup,
        Args:  cobra.NoArgs,
    }
    
    cmd.Flags().String("name", "", "Name for the backup")
    cmd.Flags().Bool("online", false, "Perform online backup (default: offline)")
    
    return cmd
}

func RunDBBackup(cmd *cobra.Command, args []string) {
    backupName, _ := cmd.Flags().GetString("name")
    online, _ := cmd.Flags().GetBool("online")
    
    if backupName == "" {
        backupName = "backup-" + time.Now().Format("20060102-150405")
    }
    
    // Create Backup resource
    backup := &cnpgv1.Backup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      backupName,
            Namespace: options.Namespace,
        },
        Spec: cnpgv1.BackupSpec{
            Cluster: cnpgv1.LocalObjectReference{
                Name: options.SystemName + "-db-pg-cluster",
            },
            Method: cnpgv1.BackupMethodVolumeSnapshot,
            Online: online,
        },
    }
    
    util.KubeCreate(backup)
    util.Logger().Printf("Backup %s created successfully", backupName)
}
```

#### Backup Status Command
```go
// CmdDBBackupStatus returns a CLI command for backup status
func CmdDBBackupStatus() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "db-backup-status",
        Short: "Show database backup status",
        Run:   RunDBBackupStatus,
        Args:  cobra.NoArgs,
    }
    
    return cmd
}

func RunDBBackupStatus(cmd *cobra.Command, args []string) {
    // Get NooBaa system status
    system := &nbv1.NooBaa{}
    util.KubeCheck(system)
    
    if system.Status.DBStatus.BackupStatus != nil {
        backupStatus := system.Status.DBStatus.BackupStatus
        fmt.Printf("Last Backup: %s\n", backupStatus.LastBackupTime)
        fmt.Printf("Next Backup: %s\n", backupStatus.NextBackupTime)
        fmt.Printf("Total Snapshots: %d\n", backupStatus.TotalSnapshots)
        fmt.Printf("Available Snapshots: %v\n", backupStatus.AvailableSnapshots)
    }
}
```
