[![slack](https://img.shields.io/badge/slack-noobaa-brightgreen.svg?logo=slack)](https://www.noobaa.io/community)
[![noobaa-core](https://img.shields.io/github/v/release/noobaa/noobaa-core?label=noobaa-core)](https://github.com/noobaa/noobaa-core/releases/latest)
[![noobaa-operator](https://img.shields.io/github/v/release/noobaa/noobaa-operator?label=noobaa-operator)](https://github.com/noobaa/noobaa-operator/releases/latest)
<div id="top"></div>

# NooBaa Operator

NooBaa is an object data service for hybrid and multi cloud environments. NooBaa runs on kubernetes, provides an S3 object store service (and Lambda with bucket triggers) to clients both inside and outside the cluster, and uses storage resources from within or outside the cluster, with flexible placement policies to automate data use cases.

# Usage

Install latest operator CLI
(or pick from the [releases page](https://github.com/noobaa/noobaa-operator/releases)):

```bash
OS="linux"
# or
OS="mac"

VERSION=$(curl -s https://api.github.com/repos/noobaa/noobaa-operator/releases/latest | jq -r '.name')
curl -LO https://github.com/noobaa/noobaa-operator/releases/download/$VERSION/noobaa-$OS-$VERSION
chmod +x noobaa-$OS-$VERSION
mv noobaa-$OS-$VERSION /usr/local/bin/noobaa
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

Notes:
- Help: `noobaa --help`
- kubeconfig: - same as kubectl - The CLI operates on the current context from kubeconfig which can be changed with `export KUBECONFIG=/path/to/custom/kubeconfig` or use the --kubeconfig and --namespace flags.
- minikube: use `noobaa install --mini` in order to allocate less resources.
- Uninstalling: `noobaa uninstall`

The CLI helps with most management tasks and focuses on ease of use for manual operations or scripts.

Here is the top level usage:
```bash
$ noobaa --help

._   _            ______
| \ | |           | ___ \
|  \| | ___   ___ | |_/ / __ _  __ _
| . \ |/ _ \ / _ \| ___ \/ _\ |/ _\ |
| |\  | (_) | (_) | |_/ / (_| | (_| |
\_| \_/\___/ \___/\____/ \__,_|\__,_|

Install:
  install      Install the operator and create the noobaa system
  uninstall    Uninstall the operator and delete the system
  status       Status of the operator and the system

Manage:
  backingstore Manage backing stores
  bucketclass  Manage bucket classes
  obc          Manage object bucket claims
  diagnose     Collect diagnostics
  ui           Open the NooBaa UI

Advanced:
  operator     Deployment using operator
  system       Manage noobaa systems
  api          Make api call
  bucket       Manage noobaa buckets
  pvstore      Manage noobaa pv store
  crd          Deployment of CRDs
  olm          OLM related commands

Other Commands:
  completion   Generates bash completion scripts
  options      Print the list of global flags
  version      Show version

Use "noobaa <command> --help" for more information about a given command.

```

```
$ noobaa options

The following options can be passed to any command:

      --db-image='centos/mongodb-36-centos7': The database container image
      --db-storage-class='': The database volume storage class name
      --db-volume-size-gb=0: The database volume size in GB
      --image-pull-secret='': Image pull secret (must be in same namespace)
      --kubeconfig='': Paths to a kubeconfig. Only required if out-of-cluster.
      --master='': The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if
out-of-cluster.
      --mini=false: Signal the operator that it is running in a low resource environment
  -n, --namespace='noobaa': Target namespace
      --noobaa-image='noobaa/noobaa-core:5.6.0': NooBaa image
      --operator-image='noobaa/noobaa-operator:5.6.0': Operator image
      --pv-pool-default-storage-class='': The default storage class name for BackingStores of type pv-pool

```

```shell
$ noobaa version

INFO[0000] CLI version: 5.13.0
INFO[0000] noobaa-image: noobaa/noobaa-core:5.13.0
INFO[0000] operator-image: noobaa/noobaa-operator:5.13.0

```

# Troubleshooting

- Verify that there are enough resources for noobaa pods:
    - `kubectl describe pod | less`
    - `kubectl get events --sort-by .metadata.creationTimestamp`
- Make sure that there is a single **default** storage class:
    - `kubectl get sc`
    - or specify which storage class to use with `noobaa install --db-storage-class XXX --pv-pool-default-storage-class YYY`

# Documentation
- [About NooBaa](doc/about-noobaa.md)
- CRDs
    - [NooBaa](doc/noobaa-crd.md) - The basic CRD to deploy a NooBaa system.
    - [BackingStore](doc/backing-store-crd.md) - Storage resources.
    - [NamespaceStore](doc/namespace-store-crd.md) - Data resources.
    - [BucketClass](doc/bucket-class-crd.md) - Policies applied to a class of buckets.
- [OBC Provisioner](doc/obc-provisioner.md) - Method to claim a new/existing bucket.

# Developing

- Fork and clone the repo: `git clone https://github.com/<username>/noobaa-operator`
- Use minikube: `minikube start`
- Use your package manager to install `go` and `python3`.
- Source the devenv into your shell: `. devenv.sh`
- Build the project: `make`
- Test with the alias `nb` that runs the local operator from `build/_output/bin` (alias created by devenv)
- Install the operator and create the system with: `nb install`
