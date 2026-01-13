# NooBaa DB Backup


## Overview

Starting with version 5.21, NooBaa supports automated, scheduled database backups using volume snapshots. This feature is built on the CNPG volume snapshot backup functionality.
NooBaa provides the following functionality to configure automated backups:
- Configure the schedule in cron format
- Configure the volume snapshot class
- Configure the maximum number of snapshots to retain
- Automated recovery from a volume snapshot
- Manual on-demand backup via the NooBaa CLI.


## Backup and Recovery Configuration 

### Backup Configuration
The backup configuration is specified in the NooBaa CR `dbSpec.dbBackup` field. The configuration is as follows:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
spec:
  dbSpec:
    dbBackup:
      schedule: "0 0 1 * *" # 1. The schedule for the backup in cron format
      volumeSnapshot:
        volumeSnapshotClass: "csi-snapshotter" # 2. The volume snapshot class to use for the backup. 
        maxSnapshots: 10 # 3. The maximum number of snapshots to retain.
```
1. The backup schedule in cron format. Note that scheduling backups too frequently can strain the system. This field is required.
2. The volume snapshot class to use for the backup. The referenced volumeSnapshotClass must use the same csi-driver as the `dbStorageClass`. It is the user's responsibility to ensure the compatibility of the volume snapshot class, and it is not validated by the operator. This field is required.
3. The maximum number of snapshots to retain. When this limit is exceeded, the oldest snapshot will be deleted. This field is required with a minimum value of 1.

### Recovery Configuration
The DB recovery configuration is specified in the NooBaa CR `dbSpec.dbRecovery` field. The configuration is as follows:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
spec:
  dbSpec:
    dbRecovery:
      volumeSnapshotName: "snapshot-jan-2025"
```
After setting the configuration, the referenced volume snapshot is skipped during the automatic cleanup of old snapshots. To initiate the recovery process, the user should explicitly delete the `Cluster` resource `noobaa-db-pg-cluster`. **This operation is destructive and should be performed only for recovery purposes, as a last resort, with caution**.  

### On-demand Backup
On-demand backup can be taken with the NooBaa CLI
```bash
noobaa system db-backup
```
By default the `Backup` resource and volume snapshot names are `noobaa-db-pg-cluster-backup-<timestamp>`. The user can specify a custom name for the backup using the `--name` flag. The recovery from an on-demand backup is the same as the recovery from a scheduled backup.


## Technical Details
### Backup Process
When a backup configuration is specified, noobaa-operator will create a `ScheduledBackup` resource in the operator's namespace. This resource is managed by the cnpg controller, which creates a volume-snapshot `Backup` resource based on the configured schedule.
The backup is an offline backup taken from the secondary DB instance. During the backup process, the secondary DB instance is fenced, and the volume snapshot is taken after the PostgreSQL server on the secondary instance is stopped to ensure data consistency. After the backup is completed, the secondary DB instance is unfenced. Since the backup is taken from the secondary, the data is not as up to date as the primary and is missing any changes that have not yet been replicated to the secondary instance.
The backup details and status are shown in the NooBaa CR's `status.dbStatus.backupStatus` field. e.g.:
```yaml
status:
  dbStatus:
    backupStatus:
      availableSnapshots:
      - noobaa-db-pg-cluster-scheduled-backup-20260113062400
      - noobaa-db-pg-cluster-scheduled-backup-20260113085000
      lastBackupTime: "2026-01-13T08:50:00Z"
      nextBackupTime: "2026-01-13T08:55:00Z"
      totalSnapshots: 2
```

### Recovery Process
The process of recovering from a backup is automated and performed by noobaa-operator. It involves creating a new CNPG cluster based on a desired volume-snapshot.
The DB recovery configuration is specified in the NooBaa CR's `dbSpec.dbRecovery` field. The recovery process does not start automatically by the noobaa-operator to avoid unintentional deletions of the CNPG cluster. To initiate the recovery process, the user should explicitly delete the `Cluster` resource `noobaa-db-pg-cluster`. **This operation is destructive and should be performed only for recovery purposes, as a last resort, with caution**.  
After the `Cluster` resource is deleted, the noobaa-operator will start the recovery process by creating a new `Cluster` resource from the desired volume snapshot. During the creation of the new cluster, the noobaa-core and noobaa-endpoint pods will be terminated. Once the new cluster reaches a healthy state, the noobaa-operator will restart the noobaa-core and noobaa-endpoint pods. The `ScheduledBackup` resource is deleted when starting the recovery process to prevent the creation of a new snapshot during recovery. It will be created again after the recovery process is completed.
The recovery details and status are shown in the NooBaa CR's `status.dbStatus.recoveryStatus` field. e.g.:
```yaml
status:
  dbStatus:
    recoveryStatus:
      recoveryTime: "2026-01-13T06:19:59Z"
      snapshotName: noobaa-db-pg-cluster-scheduled-backup-20260113061600
      status: Completed
```

### Automatic Cleanup of Old Snapshots
When the maximum number of snapshots is exceeded, the noobaa-operator will delete the oldest snapshots to reach the desired number. The user can configure the maximum number of snapshots to retain in the NooBaa CR `dbSpec.dbBackup.volumeSnapshot.maxSnapshots` field. If a volume snapshot is specified for recovery in `dbSpec.dbRecovery`, it will be skipped during deletion.

### On-demand Backup
On-demand backups complement the scheduled backup feature. Using on-demand backups requires a `dbBackup` configuration in the NooBaa CR.
The NooBaa CLI command to take an on-demand backup creates a `Backup` resource. The backup settings are identical to those of the scheduled backup (offline backup from the secondary).  
On-demand backups are not counted toward the maximum number of snapshots to retain and are not deleted automatically as part of the automatic cleanup of old snapshots.