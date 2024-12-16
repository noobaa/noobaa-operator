
# Postgresql 15 automatic upgrade Phases
As part of ODF ver. 15, NooBaa operator will run an automatic upgrade in order to update the DB data to work with this new version.

![image](https://github.com/user-attachments/assets/8561cf15-5eb3-4a1b-a9f6-6c6373543b94)


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
Once the pods are restarted, we shoud be waiting in this stage till db pod status to be Running. Once the db pod is Running that indicates the Pg upgrade is successful. Set the DB upgrade phase to CLEANING

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
