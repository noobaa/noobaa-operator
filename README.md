[![slack](https://img.shields.io/badge/slack-noobaa-brightgreen.svg?logo=slack)](https://noobaa.slack.com/channels/general)
[![noobaa-core](https://img.shields.io/github/v/release/noobaa/noobaa-core?label=noobaa-core)](https://github.com/noobaa/noobaa-core/releases/latest)
[![noobaa-operator](https://img.shields.io/github/v/release/noobaa/noobaa-operator?label=noobaa-operator)](https://github.com/noobaa/noobaa-operator/releases/latest)
<div id="top"></div>

# NooBaa
- NooBaa is an object data service for hybrid and multi cloud environments.
- NooBaa can be run on kubernetes (k8s).\
`*` we have more ways to deploy NooBaa, based on the needs, please refer to [noobaa-core](https://github.com/noobaa/noobaa-core/) repository for more information).
- NooBaa is a highly customizable and dynamic data gateway for objects, providing data services such as caching, tiering, mirroring, dedup, encryption, compression, over any storage resource including: Amazon (S3), Google (GCS), Azure (Blob), IBM (COS), other S3-compatble, Filesystems (NSFS), PV-pool etc.
- NooBaa provides an S3 object store service (and Lambda with bucket triggers) to clients both inside and outside the cluster, and uses storage resources from within or outside the cluster, with flexible placement policies to automate data use cases.
- NooBaa goal is to simplify data flows for system administrators by connecting to any of the storage silos from private or public clouds, and providing a single scalable data services, using the same S3 API and management tools. NooBaa allows full control over data placement with dynamic policies per bucket or account.

# NooBaa Operator
NooBaa operator (the Operator) watches for NooBaa changes and reconciles them to apply the desired state.
- The NooBaa operator is following the [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) - you can find extensive documentation in the Kubernetes docs.
  - Operators are software extensions to Kubernetes that make use of custom resources ([CRs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)) to manage applications and their components.
  - NooBaa operator is a [controller](https://kubernetes.io/docs/concepts/architecture/controller/). In Kubernetes, controllers are control loops that watch the state of your cluster, then make or request changes where needed. Each controller tries to move the **current state** of the cluster closer to the **desired state**.

#### Reconcile loops
System Reconciler example - the system reconciliation is divided to the following phases:
1. [Verifying](pkg/system/phase1_verifying.go) - checks the validity of the CR (noobaa image, address, etc.).
2. [Creating](pkg/system/phase2_creating.go) - creates the needed resources (secrets, services, etc.).
3. [Connecting](pkg/system/phase3_connecting.go) - populate the statuses of mgmt and s3 services, initialize client for making calls to the server, etc.
4. [Configuring](pkg/system/phase4_configuring.go) - creates default backingstore, default bucketclass, etc.

## Getting Started

### Installation
Install the latest NooBaa CLI
(or pick from the [releases page](https://github.com/noobaa/noobaa-operator/releases)):

```bash
OS="linux"
# or
OS="darwin"

ARCH=amd64
# or
ARCH=arm64

VERSION=$(curl -s https://api.github.com/repos/noobaa/noobaa-operator/releases/latest | jq -r '.name')
curl -LO https://github.com/noobaa/noobaa-operator/releases/download/$VERSION/noobaa-operator-$VERSION-$OS-$ARCH.tar.gz
tar -xvzf noobaa-operator-$VERSION-$OS-$ARCH.tar.gz
chmod +x noobaa-operator
mv noobaa-operator /usr/local/bin/noobaa
```

Install with Mac Homebrew:
```bash
brew install noobaa/noobaa/noobaa
```

Install NooBaa to Kubernetes:
```bash
# Prepare namespace and set as current (optional)
kubectl create ns noobaa
kubectl config set-context --current --namespace noobaa

# Install the operator and system on your cluster:
noobaa install

# You can always get system status and information with:
noobaa status
```

### Usage
#### NooBaa CLI
The CLI helps with most management tasks and focuses on ease of use for manual operations or scripts.

Notes:
- Help: `noobaa --help`\
use `noobaa <command> --help` for more information about any command.
- kubeconfig - same as kubectl - the CLI operates on the current context from kubeconfig which can be changed with `export KUBECONFIG=/path/to/custom/kubeconfig` or use the `--kubeconfig` and `--namespace` flags.
- local clusters (minikube, rancher desktop): use `noobaa install --mini` or `--dev` to allocate less resources.
- Uninstalling: `noobaa uninstall`

**Examples:**\
Here is the top level usage:

**1) Help menu**
```bash
$ noobaa help
```

```
#                       #
#    /~~\___~___/~~\    #
#   |               |   #
#    \~~|\     /|~~/    #
#        \|   |/        #
#         |   |         #
#         \~~~/         #
#                       #
#      N O O B A A      #

Install:
  install          Install the operator and create the noobaa system
  upgrade          Upgrade the system, its components and CRDS
  uninstall        Uninstall the operator and delete the system
  status           Status of the operator and the system

Manage:
  backingstore     Manage backing stores
  namespacestore   Manage namespace stores
  bucketclass      Manage bucket classes
  account          Manage noobaa accounts
  obc              Manage object bucket claims
  cosi             Manage cosi resources
  diagnostics      diagnostics of items in noobaa system
  sts              Manage the NooBaa Security Token Service

Advanced:
  operator         Deployment using operator
  system           Manage noobaa systems
  api              Make api call
  bucket           Manage noobaa buckets
  pvstore          Manage noobaa pv store
  crd              Deployment of CRDs
  olm              OLM related commands

Other Commands:
  completion       Generates bash completion scripts
  options          Print the list of global flags
  version          Show version

Use "noobaa <command> --help" for more information about a given command.
```
(taken from branch 5.16)

**2) Option menu**
In case you would like to add flags that are not specific for a certain command.

```bash
$ noobaa options
```

```
The following options can be passed to any command:

    --admission=false: Install the system with admission validation webhook
    --autoscaler-type='': The type of autoscaler (hpav2, keda)
    --aws-sts-arn='': The AWS STS Role ARN which will assume role
    --cosi-driver-path='/var/lib/cosi/cosi.sock': unix socket path for COSI
    --cosi-sidecar-image='gcr.io/k8s-staging-sig-storage/objectstorage-sidecar/objectstorage-sidecar:v20221117-v0.1.0-22-g0e67387': The cosi side car container image
    --db-image='quay.io/sclorg/postgresql-15-c9s': The database container image
    --db-storage-class='': The database volume storage class name
    --db-volume-size-gb=0: The database volume size in GB
    --debug-level='default_level': The type of debug sets that the system prints (all, nsfs, warn, default_level)
    --dev=false: Set sufficient resources for dev env
    --disable-load-balancer=false: Set the service type to ClusterIP instead of LoadBalancer
    --image-pull-secret='': Image pull secret (must be in same namespace)
    --kubeconfig='': Paths to a kubeconfig. Only required if out-of-cluster.
    --manual-default-backingstore=false: allow to delete the default backingstore
    --mini=false: Signal the operator that it is running in a low resource environment
    -n, --namespace='default': Target namespace
    --noobaa-image='noobaa/noobaa-core:5.18.0': NooBaa image
    --operator-image='noobaa/noobaa-operator:5.18.0': Operator image
    --pg-ssl-cert='': ssl cert for postgres (client-side cert - need to be signed by external pg accepted CA)
    --pg-ssl-key='': ssl key for postgres (client-side cert - need to be signed by external pg accepted CA)
    --pg-ssl-required=false: Force noobaa to work with ssl (external postgres - server-side) [if server cert is self-signed, needs to add --ssl-unauthorized]
    --pg-ssl-unauthorized=false: Allow the client to work with self-signed ssl (external postgres - server-side)
    --postgres-url='': url for postgresql
    --prometheus-namespace='': namespace with installed prometheus for autoscaler
    --pv-pool-default-storage-class='': The default storage class name for BackingStores of type pv-pool
    --s3-load-balancer-source-subnets=[]: The source subnets for the S3 service load balancer
    --show-secrets=false: Show the secrets in the status output
    --sts-load-balancer-source-subnets=[]: The source subnets for the STS service load balancer
    --test-env=false: Install the system with test env minimal resources
```
(taken from branch 5.16)

**3) Current version**\
When you want to print the current CLI version and images:
```bash
$ noobaa version
```

```
INFO[0000] CLI version: 5.18.0
INFO[0000] noobaa-image: noobaa/noobaa-core:5.18.0
INFO[0000] operator-image: noobaa/noobaa-operator:5.18.0
```

## Troubleshooting
- Verify that there are enough resources for noobaa pods:
    - `kubectl describe pod | less`
    - `kubectl get events --sort-by .metadata.creationTimestamp`
- Make sure that there is a single **default** storage class:
    - `kubectl get sc`
    - or specify which storage class to use with `noobaa install --db-storage-class XXX --pv-pool-default-storage-class YYY`

## Documentation
You can find documentation related to noobaa operator and noobaa components in kubernetes in [doc](doc) directory.
For example:
- [NooBaa](doc/about-noobaa.md) - Basic terminology and links to videos.
- [S3 API Compatibility](doc/s3-compatibility.md) - Overview of S3 API compatibility in NooBaa
- [AWS API Compatibility](https://github.com/noobaa/noobaa-core/blob/master/docs/design/AWS_API_Compatibility.md) - Overview of AWS API calls support in NooBaa
- Custom Resource Definitions ([CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#adding-custom-resources)):
A custom resource is an extension of the Kubernetes API that is not necessarily available in a default Kubernetes installation. CRDs allow users to create new types of resources without adding another API server.
Once a custom resource is installed, users can create, update and access its objects using [`kubectl`](https://kubernetes.io/docs/reference/kubectl/), just as they do for built-in resources.
  - [NooBaaSystem](doc/noobaa-crd.md) - The basic CRD to deploy a NooBaa system. Represents a single installation of NooBaa that includes a set of sub-resources (backing-stores, bucket-classes, and buckets).
  - [BackingStore](doc/backing-store-crd.md) - Storage resources. These storage targets are used to store deduplicated, compressed and encrypted chunks of data.
  - [NamespaceStore](doc/namespace-store-crd.md) - Data resources. These storage targets are used to store and read plain data.
  - [BucketClass](doc/bucket-class-crd.md) - Policies applied to a class of buckets, defines bucket policies relating to data placement.
  - [Bucket Types](doc/bucket-types.md) - Overview of data and namespace buckets, and supported services
  - [Bucket Replication](doc/bucket-replication.md) - Overview of bucket replication rules in NooBaa, including log-based optimizations, inner workings, and example rules
  - [Account](doc/noobaa-account-crd.md) - We use the account to receive new credentials set for accessing different noobaa services.
- Bucket Claim:
  - [OBC Provisioner](doc/obc-provisioner.md) - OBC (Object Bucket Claim) is currently the main CR to provision buckets, however it is being deprecated in favor of COSI.
  - [COSI Provisioner](doc/cosi-provisioner.md) - COSI (Container Object Storage Interface) is a new kubernetes storage standard (like CSI, Container Storage Interface) to provision object storage buckets.
- DB:
  - The default DB is postgres, internal in the cluster.
  - [External Postgresql DB support](doc/external-postgres.md)
- Other:
  - [HA controller](doc/high-availability-controller.md) - High Availability controller improves NooBaa pods recovery in the case of a node failure.
  - [Admission Controller](doc/noobaa-admission.md) - The utilize k8s admission webhook feature to validate various NooBaa custom resource definitions.

Additional information can be found in:
- [noobaa/noobaa-core](https://github.com/noobaa/noobaa-core) repository.
- [noobaa wiki](https://github.com/noobaa/noobaa-core/wiki).
- [noobaa.io](https://www.noobaa.io/) - it is work in progress.

## Contributing
- Fork and clone the repo: `git clone https://github.com/<username>/noobaa-operator` (see [here](https://github.com/noobaa/noobaa-core/wiki/Git-Pull-Request-Guide) the full needed procedure when contributing code in Github).
- Use a [local cluster](doc/deply_noobaa_on_minikube_or_rancher_desktop.md):
  - Minikube (`minikube start`).
  - Rancher Desktop.
- Use your package manager to install `go` and `python3`.
- Source the devenv into your shell: `. devenv.sh`.
- Build the project: `make`.
- Test with the alias `nb` that runs the local operator from `build/_output/bin` (alias created by devenv).
- Install the operator and create the system with: `nb install`.
- Other:
  - [Deploy NSFS on kubernetes](https://github.com/noobaa/noobaa-core/wiki/NSFS-on-Kubernetes).
