
# Postgresql 15 automatic upgrade Phases
As part of ODF ver. 15, NooBaa operator will run an automatic upgrade in order to update the DB data to work with this new version.

![image](https://github.com/user-attachments/assets/9286f25d-0ee7-423c-93a4-fe0986f767b5)



## Phase: None
When we have phase = None, means we have not upgraded to PG16 and we check the phase everytime whether upgrade is required. The upgrade required is decided based on the desired image set on the noobaa CR. If it is PG16 then we consider the upgrade and set the Phase to PG16

## Phase: Prepare
In this phase we prepare for the upgrade. Below steps prepare for the upgrade
```
1. Bring down the endpoint pod
2. Set the postgres16 image for upgrade in the sts
3. Set "POSTGRESQL_UPGRADE = copy" env in the sts
4. Restart the DB pods to update the setting done in the db pod
5. Set the DB upgrade phase to UPGRADE
```

## Phase: Upgrade
Once the pods are restarted, we shoud be waiting in this stage till db pod status to be Running. Once the db pod is Running that indicates the Pg upgrade is successful. Set the DB upgrade phase to CLEANING. If the status is found to be CrashLoopBackOff or error then we mark the Phase as prepare and set the env and restart the pods to make the pod ready for upgrade. Since the upgrade is completely handled by postgres, if the status remains in error state then we need to check the db logs to fix the issue. We can't really revert the upgrade at this stage

## Phase: Cleaning
Phase cleaning is reverse of Prepare phase. 
```
1. Remove the "POSTGRESQL_UPGRADE = copy" env from db sts
2. Restore the endpoint pod
```
After this set the Phase to Finished

## Phase: Finished
This phase marks upgrade is finished 

## Phase: Failed
If the upgrade is failed, we need to verify the db pod logs manually. 
