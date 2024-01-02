[NooBaa Operator](../README.md) /
# Postgresql 15 automatic upgrade

As part of ODF ver. 15, NooBaa operator will run an automatic upgrade in order to update the DB data to work with this new version.

## The auto-upgrade state machine

![](https://github.com/noobaa/noobaa-operator/assets/20266280/dd12ab01-0a2d-42c7-8048-a3a0c5e5a80e)

#
### Those are the different steps running during the upgrade: 

## Prepare
Will start the upgrade procedure and will add a 2nd init container running [dumpdb.sh](../deploy/internal/configmap-postgres-initdb.yaml#L42) that will dump the old data.

in case of the DB pod using more than 33% of the attached PV space - this step will intentionelly fail and move to revert. 

## Upgrade
Will change the running version from 12 to 15 and add a 2nd init container running [upgradedb.sh](../deploy/internal/configmap-postgres-initdb.yaml#L58) - this one will save the 12 data to a backup and will run the db_dump on the new 15 DB.

in case of the DB pod using more than 33% of the attached PV space - this step will intentionelly fail and move to revert.

## Clean
Will remove the init container and will start a new db pod running version 15 with the new data.

## Revert
In case of a failure in any of the above steps, will start the revert procedure and will add a 2nd init container running [revertdb.sh](../deploy/internal/configmap-postgres-initdb.yaml#L80) that will copy the old data over the new and will start the OLD DB.

## Failed
After revert has finished - will stop reconciling the new DB and will keep running the old pgsql version 12. 

It will wait until new annotation will be added to NooBaa CR:
### retry_upgrade
Will be used in case of expending the attached PV so that it will be using less than 33% -  will restart the upgrade procedure from Prepare.
```bash
kubectl annotate noobaa noobaa retry_upgrade=true
```
### manual_upgrade_completed
Will be used after running a manual upgrade and will set the upgrade to finished.
```bash
kubectl annotate noobaa noobaa manual_upgrade_completed=true
```

## Finished
DB upgrade has successfully finished and will never run again.

