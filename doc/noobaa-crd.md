[NooBaa Operator](../README.md) /
# NooBaa CRD

NooBaa CRD represents a single installation of NooBaa that includes a set of sub-resources (backing-stores, bucket-classes, and buckets) and has a lifecycle as a single integrated system.


# Definitions

- Kubebuilder types: [noobaa_types.go](../pkg/apis/noobaa/v1alpha1/noobaa_types.go)
- CRD: [noobaa.io_noobaas_crd.yaml](../deploy/crds/noobaa.io_noobaas_crd.yaml)
- CR: [noobaa.io_v1alpha1_noobaa_cr.yaml](../deploy/crds/noobaa.io_v1alpha1_noobaa_cr.yaml)


# Reconcile

The operator watches for NooBaaSystem changes and reconcile them to apply the following deployment:

- Kubernetes Resources
  - Core server:
    - StatefulSet
    - Service (mgmt)
    - Service (s3)
    - Secrets
    - Route/Ingress (?)
  - Endpoints:
    - Deployment
    - HorizontalPodAutoscaler
- NooBaa Setup
  - Admin Account
    - Once the server is up and running the operator will call an API to setup a new system in the server which returns a secret token:
      - `noobaa-admin-mgmt-secret` will be created to contain the admin account token. It will be used for next API calls on this system.
      - `noobaa-admin-s3-secret` will be created with the admin S3 credentials to allow easy onboarding for the admin.
  - Internal backing-store
    - In order to reduce the friction of setting up backing-stores right after deployment, we create an internal backing-store with the following characteristics:
      - The internal store has limited size
      - It is using the core server PV for storage
      - It should not be used for production workloads.
  - Default bucket-class
    - This is a simple class that uses just the internal backing-store.
    - Once backing-stores are added the default class should be updated and existing data will automatically move from internal store to the new stores.
  - first.bucket
    - The operator will create a `first.bucket` using the default bucket-class.


# Status

The operator will set the status of the NooBaaSystem to represent the current state of reconciling to the desired state.\
Here is the example status structure as would be returned by a `kubectl get noobaa -n noobaa -o yaml`:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: noobaa
spec:
    # ...
status:
  accounts:
    admin:
      secretRef:
        name: noobaa-admin
        namespace: noobaa
  actualImage: noobaa/noobaa-core:X.Y.Z
  conditions:
  - lastHeartbeatTime: "2019-11-05T13:50:20Z"
    lastTransitionTime: "2019-11-06T07:03:48Z"
    message: noobaa operator completed reconcile - system is ready
    reason: SystemPhaseReady
    status: "True"
    type: Available
  - lastHeartbeatTime: "2019-11-05T13:50:20Z"
    lastTransitionTime: "2019-11-06T07:03:48Z"
    message: noobaa operator completed reconcile - system is ready
    reason: SystemPhaseReady
    status: "False"
    type: Progressing
  - lastHeartbeatTime: "2019-11-05T13:50:20Z"
    lastTransitionTime: "2019-11-05T13:50:20Z"
    message: noobaa operator completed reconcile - system is ready
    reason: SystemPhaseReady
    status: "False"
    type: Degraded
  - lastHeartbeatTime: "2019-11-05T13:50:20Z"
    lastTransitionTime: "2019-11-06T07:03:48Z"
    message: noobaa operator completed reconcile - system is ready
    reason: SystemPhaseReady
    status: "True"
    type: Upgradeable
  observedGeneration: 1
  phase: Ready
  readme: |

Welcome to NooBaa

S3 Endpoint
-----------
- Access key            : export AWS_ACCESS_KEY_ID=$(kubectl get secret noobaa-admin-s3-secret -n noobaa -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
- Secret key            : export AWS_SECRET_ACCESS_KEY=$(kubectl get secret noobaa-admin-s3-secret -n noobaa -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
- External address      : https://222.222.222.222:8443
- ClusterIP address     : https://s3.noobaa
- NodePort address      : http://192.168.99.100:30361
- Port forwarding       : kubectl port-forward -n noobaa service/s3 10443:443 # then open https://localhost:10443
- aws-cli               : alias s3="aws --endpoint https://localhost:10443 s3"

Management
-------------
- Username/password     : kubectl get secret noobaa-admin-mgmt-secret -n noobaa -o json | jq '.data|map_values(@base64d)'
- External address      : https://111.111.111.111:8443
- ClusterIP address     : https://noobaa-mgmt.noobaa:8443
- Node port address     : http://192.168.99.100:30785
- Port forwarding       : kubectl port-forward -n noobaa service/noobaa-mgmt 11443:8443 # then open https://localhost:11443


  services:
    serviceMgmt:
      externalDNS:
      - https://noobaa-mgmt-noobaa.apps.noobaa.noobaa.org
      - https://noobaa.us-east-2.elb.amazonaws.com:443
      internalDNS:
      - https://noobaa-mgmt.noobaa:443
      internalIP:
      - https://1.1.1.1:443
      nodePorts:
      - https://1.1.1.1:31433
      podPorts:
      - https://1.1.1.1:8443
    serviceS3:
      externalDNS:
      - https://s3-noobaa.apps.noobaa.noobaa.org
      - https://noobaa.us-east-2.elb.amazonaws.com:443
      internalDNS:
      - https://s3.noobaa:443
      internalIP:
      - https://1.1.1.1:443
      nodePorts:
      - https://1.1.1.1:32367
      podPorts:
      - https://1.1.1.1:6443
  ```

# Custom Images

The NooBaa spec below shows how to override the noobaa-core image used for the system deployment. Another way to change the default image is to set the env `NOOBAA_CORE_IMAGE` on the operator pod (on its deployment) which makes the operator assume a different default core image even when the NooBaa spec is not specifying it. In any case when using custom images, you will have to make sure the operator and core images are compatible with eachother.

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: noobaa
spec:
  image: noobaa/noobaa-core:v9999.9.9
```

# Private Image Registry

See below how to set `spec.imagePullSecret` in order to [pull from a private image repository](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: noobaa
spec:
  image: noobaa/noobaa-core:v9999.9.9
  dbImage: centos/mongodb-36-centos7
  imagePullSecret:
    name: <SECRET-NAME>
```

# Custom CPU and Memory Resources

The NooBaa spec can be used to control the resources of each component. Below is an example of how to set the spec to use custom compute resources. 

Keep the following in mind when choosing your custom resources values:
- Endpoints are deployed as a deployment with autoscaling, so the `minCount`/`maxCount` values should be used to set a range for the autoscaler to use, and this is typically how you can increase the system's S3 throughput. It should be preferred to increase the number of endpoints rather than increasing the resources for each endpoint.
- Setting `requests` and `limits` to the same exact value will make the pods get a "Guaranteed" QoS class - see https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/.
- When running on minikube or other small environments, the noobaa cli provides a `--mini` flag that sets a spec of minimal resources.

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: noobaa
spec:
  coreResources:
    requests:
      cpu: "8"
      memory: "16Gi"
    limits:
      cpu: "8"
      memory: "16Gi"
  dbResources:
    requests:
      cpu: "8"
      memory: "16Gi"
    limits:
      cpu: "8"
      memory: "16Gi"
  endpoints:
    minCount: 4
    maxCount: 4
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
      limits:
        cpu: "2"
        memory: "4Gi"
```

# Delete

The operator will detect deletion of a system CR, and will followup by deleting all the owned resources.

This is done by connecting owner references and letting Garbage Collection do the rest as described here:

https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/

# Database Configuration

NooBaa allows database configuration via `dbConf` field under `spec` in the NooBaa CR. This field accepts string which can contain custom database configuration.

## Example
Following example will change PostgreSQL database `max_connections` to 1000 (default being 600).

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: noobaa
spec:
  image: noobaa/noobaa-core:5.9.0
  dbImage: centos/postgresql-12-centos7
  dbType: postgres
  dbConf: |+
    max_connections = 1000
```

## Notes
1. `dbConf` field will have no effect if `dbType` is not "postgres".
2. `dbConf` configuration is not validated.
3. NooBaa uses `ConfigMap` to pass database configuration to the databases. Althought the ConfigMap is editable, it should not and cannot be used to pass custom database overrides. The reason being that NooBaa operator, as part of its reconcile process will overwrite the ConfigMap to the default values.