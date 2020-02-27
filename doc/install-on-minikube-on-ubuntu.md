# Install on minikube on ubuntu
The following script was tested with the following versions:
- Ubuntu 18.04 LTS server edition
- minikube v1.7.2
- kubectl v1.17.3
- noobaa v2.0.10

### Cleanup previous versions 
This is just in case, because the tools installed from snap do not work well:
```bash
sudo snap remove kubectl minikube docker
sudo apt-get remove docker docker-engine docker.io
```

### Install docker
```bash
sudo apt-get install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

### Install kubectl
```bash
KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mkdir -p /usr/local/bin/
sudo install kubectl /usr/local/bin/
```

### Install minikube
```bash
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
chmod +x minikube
sudo mkdir -p /usr/local/bin/
sudo install minikube /usr/local/bin/
```

### Install noobaa CLI
```bash
NOOBAA_VERSION=v2.0.10
curl -Lo noobaa https://github.com/noobaa/noobaa-operator/releases/download/$NOOBAA_VERSION/noobaa-linux-$NOOBAA_VERSION
chmod +x noobaa
sudo install noobaa /usr/local/bin/
```

### Start minikube
```bash
minikube config set vm-driver none
minikube config set memory 4000
minikube config set cpus 4
sudo minikube start
```

### Check kubectl is working
Might need to fix permissions to allow non-root to use the tools
```bash
sudo chown -R $(id -un):$(id -gn) .minikube/ .kube/
kubectl get node
NAME      STATUS   ROLES    AGE     VERSION
gubuntu   Ready    master   5m48s   v1.17.2
```

### Set current namespace to "noobaa"
```bash
kubectl config set-context --current --namespace noobaa
kubectl config get-contexts
CURRENT   NAME       CLUSTER    AUTHINFO   NAMESPACE
*         minikube   minikube   minikube   noobaa
```

### Install noobaa to minikube
```bash
noobaa install
```

### Example output:
```
INFO[0000] CLI version: 2.0.10
INFO[0000] noobaa-image: noobaa/noobaa-core:5.2.13
INFO[0000] operator-image: noobaa/noobaa-operator:2.0.10
INFO[0000] Namespace: noobaa
INFO[0000]
INFO[0000] CRD Create:
INFO[0000] ✅ Created: CustomResourceDefinition "noobaas.noobaa.io"
INFO[0000] ✅ Created: CustomResourceDefinition "backingstores.noobaa.io"
INFO[0000] ✅ Created: CustomResourceDefinition "bucketclasses.noobaa.io"
INFO[0000] ✅ Created: CustomResourceDefinition "objectbucketclaims.objectbucket.io"
INFO[0000] ✅ Created: CustomResourceDefinition "objectbuckets.objectbucket.io"
INFO[0000]
INFO[0000] Operator Install:
INFO[0000] ✅ Created: Namespace "noobaa"
INFO[0000] ✅ Created: ServiceAccount "noobaa"
INFO[0000] ✅ Created: Role "noobaa"
INFO[0000] ✅ Created: RoleBinding "noobaa"
INFO[0000] ✅ Created: ClusterRole "noobaa.noobaa.io"
INFO[0000] ✅ Created: ClusterRoleBinding "noobaa.noobaa.io"
INFO[0000] ✅ Created: Deployment "noobaa-operator"
INFO[0000]
INFO[0000] System Create:
INFO[0000] ✅ Already Exists: Namespace "noobaa"
INFO[0000] ✅ Created: NooBaa "noobaa"
INFO[0000]
INFO[0000] NOTE:
INFO[0000]   - This command has finished applying changes to the cluster.
INFO[0000]   - From now on, it only loops and reads the status, to monitor the operator work.
INFO[0000]   - You may Ctrl-C at any time to stop the loop and watch it manually.
INFO[0000]
INFO[0000] System Wait Ready:
INFO[0000] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0003] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0006] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0009] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0012] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0015] ⏳ System Phase is "". Pod "noobaa-operator-7d479b7f7b-xvf26" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [noobaa-operator]). ContainersNotReady (containers with unready status: [noobaa-operator]).
INFO[0018] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0021] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0024] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0027] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0030] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0033] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0036] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0039] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0042] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0045] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0048] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotInitialized (containers with incomplete status: [init]). ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0051] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0054] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0057] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0060] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0063] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0066] ⏳ System Phase is "Connecting". Pod "noobaa-core-0" is not yet ready: Phase="Pending". ContainersNotReady (containers with unready status: [core db]). ContainersNotReady (containers with unready status: [core db]).
INFO[0069] ⏳ System Phase is "Connecting". Container "core" is not yet ready: starting...
INFO[0072] ⏳ System Phase is "Connecting". Container "core" is not yet ready: starting...
INFO[0075] ⏳ System Phase is "Connecting". Container "core" is not yet ready: starting...
INFO[0078] ⏳ System Phase is "Connecting". Container "core" is not yet ready: starting...
INFO[0081] ⏳ System Phase is "Connecting". Container "core" is not yet ready: starting...
INFO[0084] ⏳ System Phase is "Connecting". Waiting for phase ready ...
INFO[0087] ✅ System Phase is "Ready".
INFO[0087]
INFO[0087]
INFO[0087] CLI version: 2.0.10
INFO[0087] noobaa-image: noobaa/noobaa-core:5.2.13
INFO[0087] operator-image: noobaa/noobaa-operator:2.0.10
INFO[0087] Namespace: noobaa
INFO[0087]
INFO[0087] CRD Status:
INFO[0087] ✅ Exists: CustomResourceDefinition "noobaas.noobaa.io"
INFO[0087] ✅ Exists: CustomResourceDefinition "backingstores.noobaa.io"
INFO[0087] ✅ Exists: CustomResourceDefinition "bucketclasses.noobaa.io"
INFO[0087] ✅ Exists: CustomResourceDefinition "objectbucketclaims.objectbucket.io"
INFO[0087] ✅ Exists: CustomResourceDefinition "objectbuckets.objectbucket.io"
INFO[0087]
INFO[0087] Operator Status:
INFO[0087] ✅ Exists: Namespace "noobaa"
INFO[0087] ✅ Exists: ServiceAccount "noobaa"
INFO[0087] ✅ Exists: Role "noobaa"
INFO[0087] ✅ Exists: RoleBinding "noobaa"
INFO[0087] ✅ Exists: ClusterRole "noobaa.noobaa.io"
INFO[0087] ✅ Exists: ClusterRoleBinding "noobaa.noobaa.io"
INFO[0087] ✅ Exists: Deployment "noobaa-operator"
INFO[0087]
INFO[0087] System Status:
INFO[0087] ✅ Exists: NooBaa "noobaa"
INFO[0087] ✅ Exists: StatefulSet "noobaa-core"
INFO[0087] ✅ Exists: Service "noobaa-mgmt"
INFO[0087] ✅ Exists: Service "s3"
INFO[0087] ✅ Exists: Secret "noobaa-server"
INFO[0087] ✅ Exists: Secret "noobaa-operator"
INFO[0087] ✅ Exists: Secret "noobaa-admin"
INFO[0087] ✅ Exists: StorageClass "noobaa.noobaa.io"
INFO[0087] ✅ Exists: BucketClass "noobaa-default-bucket-class"
INFO[0087] ⬛ (Optional) Not Found: BackingStore "noobaa-default-backing-store"
INFO[0087] ⬛ (Optional) CRD Unavailable: CredentialsRequest "noobaa-cloud-creds"
INFO[0087] ⬛ (Optional) CRD Unavailable: PrometheusRule "noobaa-prometheus-rules"
INFO[0087] ⬛ (Optional) CRD Unavailable: ServiceMonitor "noobaa-service-monitor"
INFO[0087] ⬛ (Optional) CRD Unavailable: Route "noobaa-mgmt"
INFO[0087] ⬛ (Optional) CRD Unavailable: Route "s3"
INFO[0087] ✅ Exists: PersistentVolumeClaim "db-noobaa-core-0"
INFO[0087] ✅ System Phase is "Ready"
INFO[0087] ✅ Exists:  "noobaa-admin"

#------------------#
#- Mgmt Addresses -#
#------------------#

ExternalDNS : []
ExternalIP  : []
NodePorts   : [https://172.20.82.62:32271]
InternalDNS : [https://noobaa-mgmt.noobaa.svc:443]
InternalIP  : [https://10.99.127.220:443]
PodPorts    : [https://172.17.0.5:8443]

#--------------------#
#- Mgmt Credentials -#
#--------------------#

email    : admin@noobaa.io
password : 4QSSrGFbtP5g5mGhZixnyA==

#----------------#
#- S3 Addresses -#
#----------------#

ExternalDNS : []
ExternalIP  : []
NodePorts   : [https://172.20.82.62:30540]
InternalDNS : [https://s3.noobaa.svc:443]
InternalIP  : [https://10.105.251.227:443]
PodPorts    : [https://172.17.0.5:6443]

#------------------#
#- S3 Credentials -#
#------------------#

AWS_ACCESS_KEY_ID     : ku2pBEPGqsTvGd47qfL9
AWS_SECRET_ACCESS_KEY : KQTQOIK1BGZcYYPqaYw8RFfpt0RmsVsyfdn/BR4J

#------------------#
#- Backing Stores -#
#------------------#

No backing stores found.

#------------------#
#- Bucket Classes -#
#------------------#

NAME                          PLACEMENT                                                             PHASE      AGE
noobaa-default-bucket-class   {Tiers:[{Placement: BackingStores:[noobaa-default-backing-store]}]}   Rejected   3s
 
#-----------------#
#- Bucket Claims -#
#-----------------#

No OBC's found.

```
 
