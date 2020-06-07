# NooBaa Operator

NooBaa is an object data service for hybrid and multi cloud environments. NooBaa runs on kubernetes, provides an S3 object store service (and Lambda with bucket triggers) to clients both inside and outside the cluster, and uses storage resources from within or outside the cluster, with flexible placement policies to automate data use cases.

[About NooBaa](doc/about-noobaa.md)

# Using the operator as CLI

- Download the compiled operator binary from the [releases page](https://github.com/noobaa/noobaa-operator/releases)

For Mac
```
brew install noobaa/noobaa/noobaa
# or
wget https://github.com/noobaa/noobaa-operator/releases/download/v2.2.0/noobaa-mac-v2.2.0; mv noobaa-mac-* noobaa; chmod +x noobaa
```

For Linux
```
wget https://github.com/noobaa/noobaa-operator/releases/download/v2.2.0/noobaa-linux-v2.2.0; mv noobaa-linux-* noobaa; chmod +x noobaa
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
      --noobaa-image='noobaa/noobaa-core:5.4.0': NooBaa image
      --operator-image='noobaa/noobaa-operator:2.2.0': Operator image
      --pv-pool-default-storage-class='': The default storage class name for BackingStores of type pv-pool

# Troubleshooting

- The operator is running, but there is no noobaa-core-0 pod 

INFO[0000] CLI version: 2.2.0
INFO[0000] noobaa-image: noobaa/noobaa-core:5.4.0
INFO[0000] operator-image: noobaa/noobaa-operator:2.2.0

    Verify that there are enough resources. run `oc describe pod noobaa-core-0` for more information

# Operator Design

CRDs
- [NooBaa](doc/noobaa-crd.md) - The basic CRD to deploy a NooBaa system.
- [BackingStore](doc/backing-store-crd.md) - Connection to cloud or local storage to use in policies.
- [BucketClass](doc/bucket-class-crd.md) - Policies applied to a class of buckets.

Applications
- [OBC Provisioner](doc/obc-provisioner.md) - Method to claim a new/existing bucket.

# Developing

- Fork and clone the repo: `git clone https://github.com/<username>/noobaa-operator`
- Use minikube: `minikube start`
- Use your package manager to install `go` and `python3`.
- Install operator-sdk 
  For Mac:
   ```
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.13.0/operator-sdk-v0.13.0-x86_64-apple-darwin
    mv operator-sdk-v0.13.0-x86_64-apple-darwin /usr/local/bin/operator-sdk
    chmod +x /usr/local/bin/operator-sdk
    operator-sdk version

    ```

  For Linux:
```
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.13.0/operator-sdk-v0.13.0-x86_64-linux-gnu
    chmod +x operator-sdk-v0.13.0-x86_64-linux-gnu
    sudo mv operator-sdk-v0.13.0-x86_64-linux-gnu /usr/local/bin/operator-sdk
    operator-sdk version    

```
- Source the devenv into your shell: `. devenv.sh`
- Build the project: `make`
- Test with the alias `nb` that runs the local operator from `build/_output/bin` (alias created by devenv)
- Install the operator and create the system with: `nb install`
